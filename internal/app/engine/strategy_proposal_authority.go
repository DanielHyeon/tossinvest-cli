package engine

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/officialfx"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyproposal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

const (
	strategyProposalKRManifestDigestEnv = "TOSSOS_STRATEGY_PROPOSAL_KR_MANIFEST_SHA256"
	strategyProposalUSManifestDigestEnv = "TOSSOS_STRATEGY_PROPOSAL_US_MANIFEST_SHA256"
	strategyProposalKeyIDEnv            = "TOSSOS_STRATEGY_PROPOSAL_KEY_ID"
	strategyProposalPublicKeyEnv        = "TOSSOS_STRATEGY_PROPOSAL_PUBLIC_KEY_BASE64"
	strategyProposalKREvidenceIDEnv     = "TOSSOS_STRATEGY_EVIDENCE_KR_ID"
	strategyProposalUSEvidenceIDEnv     = "TOSSOS_STRATEGY_EVIDENCE_US_ID"
)

type StrategyProposalReason string

const (
	StrategyProposalReady            StrategyProposalReason = "READY"
	StrategyProposalRouteNotReady    StrategyProposalReason = "ROUTE_NOT_READY"
	StrategyProposalFXNotReady       StrategyProposalReason = "FX_NOT_READY"
	StrategyProposalAuthorityInvalid StrategyProposalReason = "PROPOSAL_AUTHORITY_INVALID"
	StrategyProposalNoAcceptedScope  StrategyProposalReason = "NO_ACCEPTED_PROPOSAL"
	StrategyProposalInternalFailure  StrategyProposalReason = "INTERNAL_FAILURE"
	// 한 종목이 두 가족 이상을 제안했다. 고르는 일은 평가 뒤 조정자의 몫이고
	// 아직 그 조정자가 없으므로 그 시장 전체를 닫는다(리뷰 지적 C2).
	StrategyProposalAmbiguousFamily StrategyProposalReason = "AMBIGUOUS_FAMILY_PROPOSAL"
)

type StrategyProposalMarketSnapshot struct {
	Market                                   StrategyMarket
	Ready                                    bool
	Reason                                   StrategyProposalReason
	RoutedCount, ProposedCount, RefusedCount int
	ManifestDigest, ProposalSetDigest        string
}

type PairedStrategyProposalSnapshot struct {
	ObservedAt time.Time
	KR, US     StrategyProposalMarketSnapshot
}

func (snapshot PairedStrategyProposalSnapshot) For(market StrategyMarket) StrategyProposalMarketSnapshot {
	if market == StrategyMarketKR {
		return snapshot.KR
	}
	if market == StrategyMarketUS {
		return snapshot.US
	}
	return StrategyProposalMarketSnapshot{Market: market, Reason: StrategyProposalInternalFailure}
}

type strategyProposalEntryAuthority struct {
	route     strategyRouteEntryAuthority
	authority strategyproposal.ProductionAuthority
}

type strategyProposalMarketAuthority struct {
	market   StrategyMarket
	entries  []strategyProposalEntryAuthority
	snapshot StrategyProposalMarketSnapshot
}

type strategyProposalAuthorityPair struct {
	observedAt time.Time
	kr, us     strategyProposalMarketAuthority
}

func (pair strategyProposalAuthorityPair) forMarket(market StrategyMarket) strategyProposalMarketAuthority {
	if market == StrategyMarketKR {
		return pair.kr
	}
	if market == StrategyMarketUS {
		return pair.us
	}
	return strategyProposalMarketAuthority{market: market}
}

func (pair strategyProposalAuthorityPair) Snapshot() PairedStrategyProposalSnapshot {
	return PairedStrategyProposalSnapshot{ObservedAt: pair.observedAt, KR: pair.kr.snapshot, US: pair.us.snapshot}
}

func (pair strategyProposalAuthorityPair) ResultAuthority() strategyResultAuthorityPair {
	convert := func(market StrategyMarket, value strategyProposalMarketAuthority) strategyResultMarketAuthority {
		if len(value.entries) != 1 || !value.entries[0].authority.Proposal().ValidProposal() {
			return strategyResultMarketAuthority{market: market}
		}
		return strategyResultMarketAuthority{market: market, ready: true, result: value.entries[0].authority.Proposal()}
	}
	return strategyResultAuthorityPair{observedAt: pair.observedAt, kr: convert(StrategyMarketKR, pair.kr), us: convert(StrategyMarketUS, pair.us)}
}

type loadProductionProposalBatch func(context.Context, strategyproposal.ProductionConfig, []strategyproposal.ProductionTarget, officialfx.Evidence) (strategyproposal.ProductionBatchAuthority, error)

type strategyProposalAuthorityLoader struct {
	configDir, evidencePath, journalPath, accountRef string
	getenv                                           func(string) string
	load                                             loadProductionProposalBatch
}

func newStrategyProposalAuthorityLoader(configDir, evidencePath, journalPath, accountRef string, getenv func(string) string) *strategyProposalAuthorityLoader {
	if getenv == nil {
		getenv = os.Getenv
	}
	return &strategyProposalAuthorityLoader{configDir: filepath.Clean(strings.TrimSpace(configDir)), evidencePath: filepath.Clean(strings.TrimSpace(evidencePath)),
		journalPath: filepath.Clean(strings.TrimSpace(journalPath)), accountRef: strings.TrimSpace(accountRef), getenv: getenv,
		load: strategyproposal.LoadProductionAuthorityBatch}
}

func (loader *strategyProposalAuthorityLoader) collect(ctx context.Context, schedule strategyScheduleAuthorityPair, routes strategyRouteAuthorityPair, fx strategyFXAuthorityPair) strategyProposalAuthorityPair {
	if loader == nil || ctx == nil || schedule.observedAt.IsZero() || !schedule.observedAt.Equal(routes.observedAt) || !schedule.observedAt.Equal(fx.observedAt) {
		return failedStrategyProposalPair(schedule.observedAt, StrategyProposalInternalFailure)
	}
	type outcome struct {
		market StrategyMarket
		value  strategyProposalMarketAuthority
	}
	outcomes := make(chan outcome, 2)
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		market := market
		go func() {
			value := strategyProposalMarketAuthority{market: market, snapshot: StrategyProposalMarketSnapshot{Market: market, Reason: StrategyProposalInternalFailure}}
			func() {
				defer func() {
					if recover() != nil {
						value = strategyProposalMarketAuthority{market: market, snapshot: StrategyProposalMarketSnapshot{Market: market, Reason: StrategyProposalInternalFailure}}
					}
				}()
				value = loader.collectMarket(ctx, schedule.forMarket(market), routes.forMarket(market), fx.forMarket(market), schedule.observedAt)
			}()
			outcomes <- outcome{market: market, value: value}
		}()
	}
	pair := strategyProposalAuthorityPair{observedAt: schedule.observedAt}
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

func (loader *strategyProposalAuthorityLoader) collectMarket(ctx context.Context, schedule strategyScheduleMarketAuthority, routes strategyRouteMarketAuthority, fx strategyFXMarketAuthority, observedAt time.Time) strategyProposalMarketAuthority {
	market := routes.market
	fail := func(reason StrategyProposalReason) strategyProposalMarketAuthority {
		return strategyProposalMarketAuthority{market: market, snapshot: StrategyProposalMarketSnapshot{Market: market, Reason: reason, RoutedCount: len(routes.entries)}}
	}
	if !routes.snapshot.Ready || len(routes.entries) == 0 || !schedule.snapshot.Ready || schedule.restore.Activation == nil {
		return fail(StrategyProposalRouteNotReady)
	}
	if !fx.snapshot.Ready || !fx.read.valid {
		return fail(StrategyProposalFXNotReady)
	}
	if loader.getenv == nil || loader.load == nil || loader.configDir == "." || loader.evidencePath == "." || loader.journalPath == "." || loader.accountRef == "" {
		return fail(StrategyProposalInternalFailure)
	}
	encoded := strings.TrimSpace(loader.getenv(strategyProposalPublicKeyEnv))
	key, err := base64.StdEncoding.Strict().DecodeString(encoded)
	if err != nil || base64.StdEncoding.EncodeToString(key) != encoded || len(key) != ed25519.PublicKeySize {
		return fail(StrategyProposalAuthorityInvalid)
	}
	digestEnv, evidenceEnv := strategyProposalKRManifestDigestEnv, strategyProposalKREvidenceIDEnv
	if market == StrategyMarketUS {
		digestEnv, evidenceEnv = strategyProposalUSManifestDigestEnv, strategyProposalUSEvidenceIDEnv
	}
	digest := strings.TrimSpace(loader.getenv(digestEnv))
	targets := make([]strategyproposal.ProductionTarget, 0, len(routes.entries))
	bySymbol := make(map[string]strategyRouteEntryAuthority, len(routes.entries))
	for _, entry := range routes.entries {
		symbol := entry.approved.Symbol()
		if symbol == "" || bySymbol[symbol].approved.Valid() {
			return fail(StrategyProposalInternalFailure)
		}
		bySymbol[symbol] = entry
		targets = append(targets, strategyproposal.ProductionTarget{Approved: entry.approved, Router: entry.route.Request()})
	}
	batch, err := loader.load(ctx, strategyproposal.ProductionConfig{ConfigDir: loader.configDir, EvidencePath: loader.evidencePath, JournalPath: loader.journalPath,
		AccountRef: loader.accountRef, Market: strategyrouter.Market(market), ManifestDigest: digest, TrustedKeyID: strings.TrimSpace(loader.getenv(strategyProposalKeyIDEnv)),
		TrustedKey: ed25519.PublicKey(key), ObservedAt: observedAt, RouteManifestDigest: routes.snapshot.ManifestDigest,
		ActivationDigest: schedule.snapshot.ActivationManifestDigest, CalendarGeneration: schedule.desired.CalendarVersion,
		CalendarDigest: schedule.calendar.Version, SchedulerConfigVersion: schedule.desired.ConfigVersion,
		EvidenceDBIdentity: strings.TrimSpace(loader.getenv(evidenceEnv))}, targets, fx.read.evidence)
	if err != nil || batch.ManifestDigest() != digest {
		return fail(StrategyProposalAuthorityInvalid)
	}
	// 목록을 만들기 *전에* 판단한다. 모호한 종목을 그냥 빼고 넘어가면
	// 목록이 둘에서 하나로 줄어 아래 파이프라인의 len(entries)==1 관문이
	// 오히려 만족되고, 막으려던 것과 상관없는 *다른* 종목이 풀린다.
	// routes.entries 는 심볼 순으로 정렬되어 있으므로 어느 종목이 먼저 걸리는지도 정해져 있다.
	for _, entry := range routes.entries {
		if batch.Ambiguous(entry.approved.Symbol()) {
			result := fail(StrategyProposalAmbiguousFamily)
			result.snapshot.ManifestDigest = digest
			return result
		}
	}
	entries := make([]strategyProposalEntryAuthority, 0, batch.Len())
	refused := 0
	for symbol, route := range bySymbol {
		authority, ok := batch.For(symbol)
		if !ok || !authority.Proposal().ValidProposal() {
			refused++
			continue
		}
		entries = append(entries, strategyProposalEntryAuthority{route: route, authority: authority})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].route.approved.Symbol() < entries[j].route.approved.Symbol() })
	if len(entries) == 0 {
		result := fail(StrategyProposalNoAcceptedScope)
		result.snapshot.RefusedCount = refused
		result.snapshot.ManifestDigest = digest
		return result
	}
	h := sha256.New()
	for _, entry := range entries {
		_, _ = h.Write([]byte(entry.route.approved.Symbol() + "\x00" + entry.authority.Proposal().Lineage.Identity + "\x00"))
	}
	return strategyProposalMarketAuthority{market: market, entries: entries, snapshot: StrategyProposalMarketSnapshot{Market: market, Ready: true, Reason: StrategyProposalReady,
		RoutedCount: len(routes.entries), ProposedCount: len(entries), RefusedCount: refused, ManifestDigest: digest, ProposalSetDigest: "sha256:" + hex.EncodeToString(h.Sum(nil))}}
}

func failedStrategyProposalPair(observedAt time.Time, reason StrategyProposalReason) strategyProposalAuthorityPair {
	market := func(value StrategyMarket) strategyProposalMarketAuthority {
		return strategyProposalMarketAuthority{market: value, snapshot: StrategyProposalMarketSnapshot{Market: value, Reason: reason}}
	}
	return strategyProposalAuthorityPair{observedAt: observedAt, kr: market(StrategyMarketKR), us: market(StrategyMarketUS)}
}
