package engine

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyaccount"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
)

const (
	strategyAccountKRManifestDigestEnv = "TOSSOS_STRATEGY_ACCOUNT_KR_MANIFEST_SHA256"
	strategyAccountUSManifestDigestEnv = "TOSSOS_STRATEGY_ACCOUNT_US_MANIFEST_SHA256"
	strategyAccountKeyIDEnv            = "TOSSOS_STRATEGY_ACCOUNT_KEY_ID"
	strategyAccountPublicKeyEnv        = "TOSSOS_STRATEGY_ACCOUNT_PUBLIC_KEY_BASE64"
)

type StrategyAccountReason string

const (
	StrategyAccountReady                StrategyAccountReason = "READY"
	StrategyAccountProposalNotReady     StrategyAccountReason = "PROPOSAL_NOT_READY"
	StrategyAccountAuthorityUnavailable StrategyAccountReason = "ACCOUNT_AUTHORITY_UNAVAILABLE"
	StrategyAccountInternalFailure      StrategyAccountReason = "INTERNAL_FAILURE"
)

type StrategyAccountMarketSnapshot struct {
	Market                   StrategyMarket
	Ready                    bool
	Reason                   StrategyAccountReason
	Generation               uint64
	QuoteCurrency            string
	ManifestDigest, Identity string
}

type PairedStrategyAccountSnapshot struct {
	ObservedAt time.Time
	KR, US     StrategyAccountMarketSnapshot
}

func (snapshot PairedStrategyAccountSnapshot) For(market StrategyMarket) StrategyAccountMarketSnapshot {
	if market == StrategyMarketKR {
		return snapshot.KR
	}
	if market == StrategyMarketUS {
		return snapshot.US
	}
	return StrategyAccountMarketSnapshot{Market: market, Reason: StrategyAccountInternalFailure}
}

type strategyAccountMarketAuthority struct {
	market    StrategyMarket
	authority strategyaccount.Authority
	snapshot  StrategyAccountMarketSnapshot
}

type strategyAccountAuthorityPair struct {
	observedAt time.Time
	kr, us     strategyAccountMarketAuthority
}

func (pair strategyAccountAuthorityPair) forMarket(market StrategyMarket) strategyAccountMarketAuthority {
	if market == StrategyMarketKR {
		return pair.kr
	}
	if market == StrategyMarketUS {
		return pair.us
	}
	return strategyAccountMarketAuthority{market: market}
}

func (pair strategyAccountAuthorityPair) Snapshot() PairedStrategyAccountSnapshot {
	return PairedStrategyAccountSnapshot{ObservedAt: pair.observedAt, KR: pair.kr.snapshot, US: pair.us.snapshot}
}

type loadProductionStrategyAccount func(context.Context, strategyaccount.ProductionConfig) (strategyaccount.Authority, error)

type strategyAccountAuthorityLoader struct {
	configDir, accountRef, accountCurrency string
	observedAt                             time.Time
	digests                                map[StrategyMarket]string
	keyID                                  string
	key                                    ed25519.PublicKey
	load                                   loadProductionStrategyAccount
}

func newStrategyAccountAuthorityLoader(configDir, accountRef, accountCurrency string, observedAt time.Time, getenv func(string) string) *strategyAccountAuthorityLoader {
	if getenv == nil {
		getenv = os.Getenv
	}
	encoded := strings.TrimSpace(getenv(strategyAccountPublicKeyEnv))
	key, err := base64.StdEncoding.Strict().DecodeString(encoded)
	if err != nil || base64.StdEncoding.EncodeToString(key) != encoded || len(key) != ed25519.PublicKeySize {
		key = nil
	}
	return &strategyAccountAuthorityLoader{configDir: filepath.Clean(strings.TrimSpace(configDir)), accountRef: strings.TrimSpace(accountRef),
		accountCurrency: strings.ToUpper(strings.TrimSpace(accountCurrency)), observedAt: observedAt.UTC(),
		digests: map[StrategyMarket]string{StrategyMarketKR: strings.TrimSpace(getenv(strategyAccountKRManifestDigestEnv)),
			StrategyMarketUS: strings.TrimSpace(getenv(strategyAccountUSManifestDigestEnv))},
		keyID: strings.TrimSpace(getenv(strategyAccountKeyIDEnv)), key: ed25519.PublicKey(key), load: strategyaccount.LoadProductionAuthority}
}

func (loader *strategyAccountAuthorityLoader) collect(ctx context.Context, proposals strategyProposalAuthorityPair) strategyAccountAuthorityPair {
	if loader == nil || ctx == nil || loader.observedAt.IsZero() || !loader.observedAt.Equal(proposals.observedAt) {
		return failedStrategyAccountPair(accountLoaderTime(loader), StrategyAccountInternalFailure)
	}
	type outcome struct {
		market StrategyMarket
		value  strategyAccountMarketAuthority
	}
	outcomes := make(chan outcome, 2)
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		market := market
		go func() {
			value := strategyAccountMarketAuthority{market: market, snapshot: StrategyAccountMarketSnapshot{Market: market, Reason: StrategyAccountInternalFailure}}
			func() {
				defer func() {
					if recover() != nil {
						value = strategyAccountMarketAuthority{market: market, snapshot: StrategyAccountMarketSnapshot{Market: market, Reason: StrategyAccountInternalFailure}}
					}
				}()
				value = loader.collectMarket(ctx, market, proposals.forMarket(market))
			}()
			outcomes <- outcome{market: market, value: value}
		}()
	}
	pair := strategyAccountAuthorityPair{observedAt: loader.observedAt}
	for range 2 {
		result := <-outcomes
		if result.market == StrategyMarketKR {
			pair.kr = result.value
		} else {
			pair.us = result.value
		}
	}
	return pair
}

func (loader *strategyAccountAuthorityLoader) collectMarket(ctx context.Context, market StrategyMarket, proposal strategyProposalMarketAuthority) strategyAccountMarketAuthority {
	fail := func(reason StrategyAccountReason) strategyAccountMarketAuthority {
		return strategyAccountMarketAuthority{market: market, snapshot: StrategyAccountMarketSnapshot{Market: market, Reason: reason}}
	}
	if len(proposal.entries) != 1 || !proposal.entries[0].authority.Proposal().ValidProposal() {
		return fail(StrategyAccountProposalNotReady)
	}
	if loader.load == nil || len(loader.key) != ed25519.PublicKeySize || loader.configDir == "." || loader.accountRef == "" {
		return fail(StrategyAccountInternalFailure)
	}
	result := proposal.entries[0].authority.Proposal()
	accountMarket := strategyaccount.MarketKR
	if market == StrategyMarketUS {
		accountMarket = strategyaccount.MarketUS
	}
	authority, err := loader.load(ctx, strategyaccount.ProductionConfig{ConfigDir: loader.configDir, AccountRef: loader.accountRef,
		AccountCurrency: loader.accountCurrency, Symbol: result.Lineage.Symbol, Market: accountMarket, ManifestDigest: loader.digests[market],
		TrustedKeyID: loader.keyID, TrustedKey: loader.key, ObservedAt: loader.observedAt})
	if err != nil || authority.Market() != accountMarket || authority.ManifestDigest() != loader.digests[market] {
		return fail(StrategyAccountAuthorityUnavailable)
	}
	return strategyAccountMarketAuthority{market: market, authority: authority, snapshot: StrategyAccountMarketSnapshot{Market: market,
		Ready: true, Reason: StrategyAccountReady, Generation: authority.Generation(), QuoteCurrency: authority.QuoteCurrency(),
		ManifestDigest: authority.ManifestDigest(), Identity: authority.Identity()}}
}

func failedStrategyAccountPair(observedAt time.Time, reason StrategyAccountReason) strategyAccountAuthorityPair {
	market := func(value StrategyMarket) strategyAccountMarketAuthority {
		return strategyAccountMarketAuthority{market: value, snapshot: StrategyAccountMarketSnapshot{Market: value, Reason: reason}}
	}
	return strategyAccountAuthorityPair{observedAt: observedAt, kr: market(StrategyMarketKR), us: market(StrategyMarketUS)}
}

func accountLoaderTime(loader *strategyAccountAuthorityLoader) time.Time {
	if loader == nil {
		return time.Time{}
	}
	return loader.observedAt
}

type productionStrategyFirstLegAuthorityLoader struct {
	clk       clock.Clock
	journal   *journal.Journal
	guardian  *execgw.RiskGuardian
	schedule  strategyScheduleAuthorityPair
	proposals strategyProposalAuthorityPair
	risk      strategyRiskAuthorityPair
	fx        strategyFXAuthorityPair
	accounts  strategyAccountAuthorityPair
}

func newProductionStrategyFirstLegAuthorityLoader(clk clock.Clock, jrn *journal.Journal, guardian *execgw.RiskGuardian,
	schedule strategyScheduleAuthorityPair, proposals strategyProposalAuthorityPair, riskAuthority strategyRiskAuthorityPair,
	fx strategyFXAuthorityPair, accounts strategyAccountAuthorityPair,
) *productionStrategyFirstLegAuthorityLoader {
	return &productionStrategyFirstLegAuthorityLoader{clk: clk, journal: jrn, guardian: guardian, schedule: schedule,
		proposals: proposals, risk: riskAuthority, fx: fx, accounts: accounts}
}

func (loader *productionStrategyFirstLegAuthorityLoader) collectStrategyFirstLegAuthority(ctx context.Context, accepted strategyFirstLegAccepted) (execgw.QFinalCampaignFirstLegIssuance, error) {
	if loader == nil || ctx == nil || loader.clk == nil || loader.journal == nil || loader.guardian == nil {
		return execgw.QFinalCampaignFirstLegIssuance{}, errors.New("production first-leg authority is unavailable")
	}
	market := StrategyMarket(accepted.market)
	proposal, riskAuthority, fx, account, schedule := loader.proposals.forMarket(market), loader.risk.forMarket(market),
		loader.fx.forMarket(market), loader.accounts.forMarket(market), loader.schedule.forMarket(market)
	if len(proposal.entries) != 1 || !riskAuthority.snapshot.Ready || !fx.snapshot.Ready || !account.snapshot.Ready ||
		!schedule.snapshot.Ready || schedule.restore.Activation == nil {
		return execgw.QFinalCampaignFirstLegIssuance{}, errors.New("paired production authority is incomplete for market")
	}
	proposalAuthority := proposal.entries[0].authority
	result := proposalAuthority.Proposal()
	if result.Lineage.Identity != accepted.result.Lineage.Identity || result.ExecutionTerms.Identity() != accepted.result.ExecutionTerms.Identity() {
		return execgw.QFinalCampaignFirstLegIssuance{}, errors.New("production proposal identity changed")
	}
	scope := riskAuthority.bundle.Scope()
	if err := riskAuthority.bundle.Validate(scope); err != nil || scope.AccountID != result.Lineage.AccountRef ||
		string(scope.Market) != string(result.Lineage.Market) || scope.Symbol != result.Lineage.Symbol || !scope.AsOf.Equal(loader.accounts.observedAt) {
		return execgw.QFinalCampaignFirstLegIssuance{}, errors.New("production risk authority scope changed")
	}
	cas, err := loader.journal.CurrentPositionCampaignCAS(ctx, result.Lineage.AccountRef, string(result.Lineage.Market), result.Lineage.Symbol)
	if err != nil || cas.Claimed || cas.State != "FLAT" && cas.State != "CLOSED" || cas.Generation < 0 ||
		result.Lineage.PositionGeneration != uint64(cas.Generation)+1 {
		return execgw.QFinalCampaignFirstLegIssuance{}, errors.New("production position campaign CAS changed")
	}
	entries := riskAuthority.bundle.Entries()
	buckets := make([]riskbucket.BucketSnapshot, 0, len(entries))
	references := make([]journal.RiskBucketSnapshotReference, 0, len(entries))
	for _, entry := range entries {
		buckets = append(buckets, entry.Bucket)
		reference := entry.Reference
		references = append(references, journal.RiskBucketSnapshotReference{Key: reference.Key, SnapshotID: reference.SnapshotID,
			SnapshotDigest: reference.SnapshotDigest, SnapshotVersion: reference.SnapshotVersion, PolicyDigest: reference.PolicyDigest,
			PolicyObservedAt: reference.PolicyObservedAt, PolicyFreshUntil: reference.PolicyFreshUntil,
			SnapshotObservedAt: reference.SnapshotObservedAt, SnapshotFreshUntil: reference.SnapshotFreshUntil})
	}
	entryPrice, entryOK := result.ExecutionTerms.Entry().MajorDecimal()
	stopPrice, stopOK := result.ExecutionTerms.EffectiveStop().MajorDecimal()
	targetPrice, targetOK := result.ExecutionTerms.Target().MajorDecimal()
	if !entryOK || !stopOK || !targetOK {
		return execgw.QFinalCampaignFirstLegIssuance{}, errors.New("production strategy price unit invalid")
	}
	bindingDigest := strategyFirstLegBindingDigest(result, account.authority, riskAuthority.bundle.Digest(), schedule.snapshot.ActivationManifestDigest)
	transactionID := "strategy-risk:" + strings.TrimPrefix(bindingDigest, "sha256:")[:32]
	attemptID := "strategy-attempt:" + strings.TrimPrefix(bindingDigest, "sha256:")[:32]
	observedAt := loader.accounts.observedAt
	collect := func(readCtx context.Context, _ int) (execgw.ExposureSnapshot, error) {
		now := loader.clk.Now().UTC()
		if readCtx == nil || readCtx.Err() != nil || now.IsZero() || now.After(account.authority.FreshUntil()) {
			return execgw.ExposureSnapshot{}, errors.New("production account exposure snapshot expired")
		}
		version, versionErr := loader.journal.ReservationVersion(readCtx, result.Lineage.AccountRef)
		if versionErr != nil {
			return execgw.ExposureSnapshot{}, versionErr
		}
		return execgw.ExposureSnapshot{AsOf: account.authority.ObservedAt(), Version: version, OpenExposure: account.authority.OpenExposure()}, nil
	}
	owner := riskbucket.OwnerClaim{Key: riskbucket.OwnerKey{AccountID: result.Lineage.AccountRef, Market: scope.Market, Symbol: result.Lineage.Symbol},
		LaneID: result.Lineage.LaneID, CampaignID: result.Lineage.CampaignID}
	return execgw.QFinalCampaignFirstLegIssuance{Entry: execgw.QFinalEntryIssuance{Market: string(result.Lineage.Market), Currency: accepted.currency,
		Symbol: result.Lineage.Symbol, QCandidate: result.Quantity, EntryPrice: entryPrice, StopPrice: stopPrice, TargetPrice: targetPrice,
		Account: account.authority.AccountState(), Collect: collect, Admission: journal.RiskBucketAdmissionPlan{TransactionID: transactionID,
			Admission: riskbucket.AdmissionRequest{QCandidate: result.Quantity, Policy: riskAuthority.bundle.Policy(), Buckets: buckets},
			Owner:     owner, Snapshots: references, CreatedAt: observedAt}, FXAuthority: fx.read.evidence,
		ExpectedPolicyVersion: loader.guardian.PolicyVersion(), ExpectedLimitsDigest: loader.guardian.LimitsDigest()},
		Result: result, ActivationManifestDigest: schedule.snapshot.ActivationManifestDigest, AttemptID: attemptID, Revision: 1,
		Campaign: journal.FirstLegCampaignRequest{CampaignID: result.Lineage.CampaignID, ExpectedPositionGeneration: cas.Generation,
			ExpectedPositionVersion: cas.Version, CreateCommandKey: "campaign-create:" + strings.TrimPrefix(bindingDigest, "sha256:")[:32],
			FirstLegCommandKey: "campaign-leg:" + strings.TrimPrefix(bindingDigest, "sha256:")[:32],
			FirstLegPlanID:     "first-leg:" + strings.TrimPrefix(bindingDigest, "sha256:")[:32]}, Weekly: proposalAuthority.WeeklyBinding()}, nil
}

func strategyFirstLegBindingDigest(result strategyflow.Result, account strategyaccount.Authority, riskDigest, activationDigest string) string {
	value := strings.Join([]string{"TossOS/production-first-leg/v1", result.Lineage.Identity, result.ExecutionTerms.Identity(),
		account.Identity(), riskDigest, activationDigest}, "\x00")
	sum := sha256.Sum256([]byte(value))
	return "sha256:" + hex.EncodeToString(sum[:])
}

var _ strategyFirstLegAuthorityLoader = (*productionStrategyFirstLegAuthorityLoader)(nil)
