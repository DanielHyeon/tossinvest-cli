package engine

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
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
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
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
	// scopes 는 소유자 범위별 계좌 권한이다(a112 5.2.2.2). 활성화 없는 시장은 원소 하나.
	scopes []strategyAccountScopeAuthority
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
	// a112 5.2.2.2: 활성화 없는 시장은 오늘처럼 항목이 정확히 하나여야 한다(토글 OFF = upstream). 서명 활성화된 시장은 조립이 중재한
	// 항목(소유자 범위)마다 계좌 권한 하나 — 보이스 A #5: 계좌 권한은 `entries[0]` 이 아니라 **그 범위의 종목**으로 적재되어야 한다.
	activated := proposal.familyActivation().Verified()
	if len(proposal.entries) == 0 || !activated && (len(proposal.entries) != 1 || !proposal.entries[0].authority.Proposal().ValidProposal()) {
		return fail(StrategyAccountProposalNotReady)
	}
	if loader.load == nil || len(loader.key) != ed25519.PublicKeySize || loader.configDir == "." || loader.accountRef == "" {
		return fail(StrategyAccountInternalFailure)
	}
	accountMarket := strategyaccount.MarketKR
	if market == StrategyMarketUS {
		accountMarket = strategyaccount.MarketUS
	}
	scopes := make([]strategyAccountScopeAuthority, 0, len(proposal.entries))
	for _, entry := range proposal.entries {
		result := entry.authority.Proposal()
		key, keyed := strategyOwnerKeyOf(result.Lineage)
		scoped := strategyAccountScopeAuthority{key: key, reason: StrategyAccountProposalNotReady,
			cause: errors.New("production account owner scope key invalid or proposal invalid")}
		if keyed && result.ValidProposal() {
			scoped.reason = StrategyAccountAuthorityUnavailable
			authority, err := loader.load(ctx, strategyaccount.ProductionConfig{ConfigDir: loader.configDir, AccountRef: loader.accountRef,
				AccountCurrency: loader.accountCurrency, Symbol: result.Lineage.Symbol, Market: accountMarket, ManifestDigest: loader.digests[market],
				TrustedKeyID: loader.keyID, TrustedKey: loader.key, ObservedAt: loader.observedAt})
			switch {
			case err != nil:
				// 적재기는 ctx 종료를 자기 오류로 접으므로 ctx 를 직접 봐 원인을 보존함. 계좌 적재 실패는 원인과 무관하게 1차 레그에서 결함(판정 (A)).
				scoped.cause = err
				if ctxErr := ctx.Err(); ctxErr != nil {
					scoped.cause = ctxErr
				}
			case authority.Market() == accountMarket && authority.ManifestDigest() == loader.digests[market]:
				scoped.authority, scoped.ready, scoped.reason, scoped.cause = authority, true, StrategyAccountReady, nil
			default:
				scoped.cause = errors.New("production account authority does not match the loader's market or manifest")
			}
		}
		scopes = append(scopes, scoped)
	}
	return strategyAccountMarketFromScopes(market, scopes)
}

// strategyAccountMarketFromScopes 는 범위별 계좌 권한에서 시장 권한을 만든다. 시장 칸은 첫 준비된 범위의 것(범위 하나면 오늘과 같은 값).
// 준비된 범위가 없으면 첫 범위의 사유로 시장 전체가 준비 안 됨 — 범위 하나일 때 오늘의 사유 그대로.
func strategyAccountMarketFromScopes(market StrategyMarket, scopes []strategyAccountScopeAuthority) strategyAccountMarketAuthority {
	for _, scope := range scopes {
		if !scope.ready {
			continue
		}
		authority := scope.authority
		// 범위가 둘 이상이면 식별은 준비된 범위들의 묶음(위험 BundleDigest 와 대칭 — 5.2.2.2 리뷰 A #3); 범위 하나면 오늘과 같은 값.
		identity := authority.Identity()
		if len(scopes) > 1 {
			ready := make([]string, 0, len(scopes))
			for _, each := range scopes {
				if each.ready {
					ready = append(ready, each.authority.Identity())
				}
			}
			identity = strategyWorkerEvidenceDigest(ready...)
		}
		return strategyAccountMarketAuthority{market: market, authority: authority, scopes: scopes, snapshot: StrategyAccountMarketSnapshot{
			Market: market, Ready: true, Reason: StrategyAccountReady, Generation: authority.Generation(), QuoteCurrency: authority.QuoteCurrency(),
			ManifestDigest: authority.ManifestDigest(), Identity: identity}}
	}
	reason := StrategyAccountProposalNotReady
	if len(scopes) != 0 {
		reason = scopes[0].reason
	}
	return strategyAccountMarketAuthority{market: market, scopes: scopes, snapshot: StrategyAccountMarketSnapshot{Market: market, Reason: reason}}
}

// earliestFreshUntil 은 준비된 계좌 범위들 중 가장 이른 FreshUntil 이다(worker 권한 만료 — 5.2.2.2 리뷰 A #3). 준비된 범위가 없으면 0.
func (authority strategyAccountMarketAuthority) earliestFreshUntil() time.Time {
	var earliest time.Time
	for _, scope := range authority.scopes {
		if fresh := scope.authority.FreshUntil(); scope.ready && (earliest.IsZero() || fresh.Before(earliest)) {
			earliest = fresh
		}
	}
	return earliest
}

// forScope 는 그 소유자 범위의 준비된 계좌 권한이다. 없으면 false — 봉투 값으로 폴백하지 않는다.
func (authority strategyAccountMarketAuthority) forScope(key strategyrouter.OwnerKey) (strategyaccount.Authority, bool) {
	for _, scope := range authority.scopes {
		if scope.key == key && scope.ready {
			return scope.authority, true
		}
	}
	return strategyaccount.Authority{}, false
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
	// 제안 쌍을 **떼어 내어** 든다(a112 6.2 봉인 리뷰 codex #1): entries 는 slice 라 그대로 들면 dispatch · 조립이 쥔 같은 배열의
	// 원소 교체가 이 권한의 대조 원본까지 바꾼다 — 그러면 봉인의 재유도가 교체된 값을 교체된 값과 비교한다. 여기서 복사해
	// 공유 자체를 끊는다(detachedStrategyProposalPair 머리말).
	return &productionStrategyFirstLegAuthorityLoader{clk: clk, journal: jrn, guardian: guardian, schedule: schedule,
		proposals: detachedStrategyProposalPair(proposals), risk: riskAuthority, fx: fx, accounts: accounts}
}

func (loader *productionStrategyFirstLegAuthorityLoader) collectStrategyFirstLegAuthority(ctx context.Context, accepted strategyFirstLegAccepted) (execgw.QFinalCampaignFirstLegIssuance, error) {
	if loader == nil || ctx == nil || loader.clk == nil || loader.journal == nil || loader.guardian == nil {
		return execgw.QFinalCampaignFirstLegIssuance{}, errors.New("production first-leg authority is unavailable")
	}
	market := StrategyMarket(accepted.market)
	proposal, riskAuthority, fx, account, schedule := loader.proposals.forMarket(market), loader.risk.forMarket(market),
		loader.fx.forMarket(market), loader.accounts.forMarket(market), loader.schedule.forMarket(market)
	if !riskAuthority.snapshot.Ready || !fx.snapshot.Ready || !account.snapshot.Ready ||
		!schedule.snapshot.Ready || schedule.restore.Activation == nil {
		return execgw.QFinalCampaignFirstLegIssuance{}, errors.New("paired production authority is incomplete for market")
	}
	// a112 6.2 봉인(의미 봉인): 조립이 중재해 둔 권한 쌍에서 **소유자 범위로** 항목 하나를 다시 꺼냄 — 0 또는 복수면 거절.
	// 범위는 선택 기준이고 대조는 아래 identity 가드가 함(자기 참조 함정 회피 — authorityForOwnerScope 머리말).
	proposalAuthority, scoped := proposal.authorityForOwnerScope(accepted.result.Lineage)
	if !scoped {
		return execgw.QFinalCampaignFirstLegIssuance{}, errors.New(strategyFirstLegOwnerScopeRefusal)
	}
	result := proposalAuthority.Proposal()
	if result.Lineage.Identity != accepted.result.Lineage.Identity || result.ExecutionTerms.Identity() != accepted.result.ExecutionTerms.Identity() {
		return execgw.QFinalCampaignFirstLegIssuance{}, errors.New("production proposal identity changed")
	}
	// 5.2.2.2 리뷰 codex #1: 마지막 권한 경계의 수용 집합은 서명 활성화 밖에서 넓어지지 않는다 — 활성화 없는 시장은 6.2 위치에서 편집 전과
	// 같은 시장 단위 개수 관문으로 거절한다(상류 handoff · 계좌 적재기도 막지만 마지막 권한이 스스로 막는다).
	if !proposal.familyActivation().Verified() && len(proposal.entries) != 1 {
		return execgw.QFinalCampaignFirstLegIssuance{}, errors.New("paired production authority is incomplete for market")
	}
	// a112 5.2.2.2 — 시장 단위 개수 관문을 걷었다. 그 관문이 지키던 「하류 권한이 시장당 하나」를 이제 **범위별 재유도**가 진다: 위험 ·
	// 계좌 권한도 봉인과 같은 소유자 범위 키로 조립 권한에서 다시 고른다. 그 범위의 권한이 없으면 **그 범위만** 거절한다(Manager 판정 J3 —
	// 범위 거절 타입 · 봉투 값 폴백 없음). 원장 · Gateway · 중앙 오류와 identity 불일치(위조 의심)는 범위 거절이 아니다.
	key, keyed := strategyOwnerKeyOf(result.Lineage)
	if !keyed { // 봉인이 정규화를 보장하므로 도달 불가 — 바뀌면 범위 거절이 아니라 결함(리뷰 A #4)
		return execgw.QFinalCampaignFirstLegIssuance{}, errors.New("production first-leg owner scope key invalid")
	}
	// 범위 권한이 없을 때 그것이 **범위 국소 원인**이면 그 범위만 타입 거절(J3), 원장 결함 · ctx · 항목 부재면 타입 없는 오류로 주기를 멈춤
	// (J4 — 5.2.2.2 리뷰 A #1 · codex #2: 편집 전에는 원장 결함도 범위 거절로 접혀 같은 주기의 다음 범위가 발급됐다).
	riskBundle, riskScoped := riskAuthority.forScope(key)
	if !riskScoped {
		scopeLocal, cause := riskAuthority.riskScopeCause(key)
		if !scopeLocal {
			return execgw.QFinalCampaignFirstLegIssuance{}, fmt.Errorf("production risk authority fault for owner scope %s: %w", key.Symbol, cause)
		}
		return execgw.QFinalCampaignFirstLegIssuance{}, &strategyScopeRefusal{scope: key, detail: "no ready risk authority for this owner scope", cause: cause}
	}
	accountAuthority, accountScoped := account.forScope(key)
	if !accountScoped {
		// 계좌 매니페스트는 시장 단위 파일 — 범위 국소 원인이 없으므로 계좌 실패는 언제나 결함(codex 재확인 P1 → Manager 판정 (A)).
		return execgw.QFinalCampaignFirstLegIssuance{}, fmt.Errorf("production account authority fault for owner scope %s: %w", key.Symbol,
			account.accountScopeCause(key))
	}
	// 발급 통화는 봉투(accepted.currency)가 아니라 조립 권한의 계보 시장에서 다시 유도한다(6.2 리뷰 보이스 A #3).
	currency, currencyKnown := map[strategyrouter.Market]string{strategyrouter.MarketKR: "KRW", strategyrouter.MarketUS: "USD"}[result.Lineage.Market]
	if !currencyKnown {
		return execgw.QFinalCampaignFirstLegIssuance{}, errors.New("production first-leg market currency unknown")
	}
	scope := riskBundle.Scope()
	if err := riskBundle.Validate(scope); err != nil || scope.AccountID != result.Lineage.AccountRef ||
		string(scope.Market) != string(result.Lineage.Market) || scope.Symbol != result.Lineage.Symbol || !scope.AsOf.Equal(loader.accounts.observedAt) {
		return execgw.QFinalCampaignFirstLegIssuance{}, errors.New("production risk authority scope changed")
	}
	cas, err := loader.journal.CurrentPositionCampaignCAS(ctx, result.Lineage.AccountRef, string(result.Lineage.Market), result.Lineage.Symbol)
	if err != nil || cas.Claimed || cas.State != "FLAT" && cas.State != "CLOSED" || cas.Generation < 0 ||
		result.Lineage.PositionGeneration != uint64(cas.Generation)+1 {
		return execgw.QFinalCampaignFirstLegIssuance{}, errors.New("production position campaign CAS changed")
	}
	entries := riskBundle.Entries()
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
	bindingDigest := strategyFirstLegBindingDigest(result, accountAuthority, riskBundle.Digest(), schedule.snapshot.ActivationManifestDigest)
	transactionID := "strategy-risk:" + strings.TrimPrefix(bindingDigest, "sha256:")[:32]
	attemptID := "strategy-attempt:" + strings.TrimPrefix(bindingDigest, "sha256:")[:32]
	observedAt := loader.accounts.observedAt
	collect := func(readCtx context.Context, _ int) (execgw.ExposureSnapshot, error) {
		now := loader.clk.Now().UTC()
		if readCtx == nil || readCtx.Err() != nil || now.IsZero() || now.After(accountAuthority.FreshUntil()) {
			return execgw.ExposureSnapshot{}, errors.New("production account exposure snapshot expired")
		}
		version, versionErr := loader.journal.ReservationVersion(readCtx, result.Lineage.AccountRef)
		if versionErr != nil {
			return execgw.ExposureSnapshot{}, versionErr
		}
		return execgw.ExposureSnapshot{AsOf: accountAuthority.ObservedAt(), Version: version, OpenExposure: accountAuthority.OpenExposure()}, nil
	}
	owner := riskbucket.OwnerClaim{Key: riskbucket.OwnerKey{AccountID: result.Lineage.AccountRef, Market: scope.Market, Symbol: result.Lineage.Symbol},
		LaneID: result.Lineage.LaneID, CampaignID: result.Lineage.CampaignID}
	return execgw.QFinalCampaignFirstLegIssuance{Entry: execgw.QFinalEntryIssuance{Market: string(result.Lineage.Market), Currency: currency,
		Symbol: result.Lineage.Symbol, QCandidate: result.Quantity, EntryPrice: entryPrice, StopPrice: stopPrice, TargetPrice: targetPrice,
		Account: accountAuthority.AccountState(), Collect: collect, Admission: journal.RiskBucketAdmissionPlan{TransactionID: transactionID,
			Admission: riskbucket.AdmissionRequest{QCandidate: result.Quantity, Policy: riskBundle.Policy(), Buckets: buckets},
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
