package engine

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/officialfx"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyarbiter"
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
	// 보정 중재가 그 종목에서 하나를 고르지 못했다. 어떤 이유로 닫혔는지는
	// 스냅샷의 ArbitrationRefusal 이 그대로 들고 있다(태스크 5.4).
	StrategyProposalArbitrationRefused StrategyProposalReason = "ARBITRATION_REFUSED"
	// 조정자 큐가 넘쳐 그 시장을 닫았다. 동결 골든의 refusal_enums 에는 큐 코드가
	// 없으므로 중재 여섯 코드 중 하나를 빌려 쓰지 않고 엔진 자신의 이름을 쓴다 —
	// 빌려 쓰면 큐가 넘친 일이 봉인이 깨진 일로 보고된다(태스크 5.4.2).
	StrategyProposalQueueOverflow StrategyProposalReason = "PROPOSAL_QUEUE_OVERFLOW"
	// 매니페스트와 자격 집합이 **둘 다 받아들인** 스코프가 제안을 만들지 못했다.
	// 그 종목만 빼면 목록이 짧아지고, 짧아진 목록이 아래 파이프라인의
	// len(entries)==1 관문을 오히려 만족시켜 상관없는 다른 종목이 대신 풀린다 —
	// 고장 하나가 시스템을 더 관대하게 만든다. 그래서 시장을 닫는다(태스크 5.4.3).
	//
	// "제안이 원래 없는 종목"은 여기 해당하지 않는다. 그것은 예전처럼 거절로 세고
	// 시장은 열어 둔다. 둘을 가르는 판정은 strategyproposal 이 하고 여기서는 읽기만 한다.
	StrategyProposalProductionFault StrategyProposalReason = "PROPOSAL_PRODUCTION_FAULT"
)

// 아래 두 문장은 INTERNAL_FAILURE 로 닫히는 여러 원인 중 조정 경로가 낸 둘을
// 서로, 그리고 나머지와 구별하는 진단이다. 계약이 아니다 — INTERNAL_FAILURE 는
// 이 함수 안에서만도 다섯 가지 서로 다른 일에 붙는 이름이라, 코드만 남기면
// 운영자가 무엇이 닫았는지 알 방법이 없다.
const (
	strategyProposalDetailLineageCollision    = "two lanes claim the same sealed lineage identity"
	strategyProposalDetailUnresolvedSelection = "a selection has no lane to come back to"
)

type StrategyProposalMarketSnapshot struct {
	Market                                   StrategyMarket
	Ready                                    bool
	Reason                                   StrategyProposalReason
	RoutedCount, ProposedCount, RefusedCount int
	ManifestDigest, ProposalSetDigest        string
	// ArbitrationRefusal 은 Reason 이 ARBITRATION_REFUSED 일 때 중재가 돌려준
	// 계약 코드다(동결 골든 refusal_enums.arbitration 의 여섯 개 중 하나).
	// ArbitrationDetail 은 그 코드 안에서 무엇이 발화했는지 좁혀 주는 진단이며
	// 계약이 아니다 — 여섯 코드는 여러 원인을 한 이름으로 묶기 때문이다.
	ArbitrationRefusal, ArbitrationDetail string
	// QueueDropCount 는 조정자 큐에서 접히거나 들어가지 못한 봉투의 수다.
	// 유계 계수기다.
	//
	// **이 숫자는 지금 배선에서 언제나 0 이고, 아직 운영자에게 닿지도 않는다.**
	// 두 가지가 다 사실이므로 둘 다 적는다.
	//
	//  1. 언제나 0 인 이유: 계수기가 오르는 곳은 접힘과 넘침 둘뿐인데, 지금
	//     배선은 둘 다 닿지 못한다. collectMarket 이 종목 중복을 먼저 거절하고
	//     매니페스트가 (종목, 레인)마다 하나만 싣기 때문에 같은 레인 칸이 두 번
	//     오지 않으며(접힘 없음), 큐 상한이 매니페스트 상한과 같아서 정상
	//     매니페스트는 넘칠 수 없다(넘침 없음).
	//  2. 닿지 않는 이유: 읽는 소비자가 없고, 지금의 유일한 운영자 표면은 이
	//     시장의 닫힘을 EVIDENCE_STALE 하나로 뭉뚱그린다(strategy_runtime_projection.go).
	//
	// 그러므로 이 수를 "조용한 유실이 없음의 증거"로 인용하면 안 된다. 언제나
	// 0 인 계수기는 아무것도 증언하지 않는다. 투영까지 잇는 일은 태스크 7.3 이고,
	// 여러 worker 가 같은 칸을 두고 다투게 만드는 일은 태스크 5.2·5.7 이다.
	QueueDropCount uint64
	// ProductionFault 는 Reason 이 PROPOSAL_PRODUCTION_FAULT 일 때 어느 종목의
	// 어느 레인이 무엇 때문에 사라졌는지다. 종목만으로는 부족하다 — 한 종목이
	// 네 가족을 동시에 낼 수 있다. 계약이 아니라 진단이다.
	ProductionFault string
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
		// 몇 개까지 넘길 수 있는지는 여기서 정하지 않는다. handoff 한 곳이 정한다.
		handoff := value.dispatchHandoff()
		if !handoff.Admitted() || !handoff.result.ValidProposal() {
			return strategyResultMarketAuthority{market: market}
		}
		return strategyResultMarketAuthority{market: market, ready: true, result: handoff.result}
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
	// 받아들여진 스코프가 제안을 잃었으면 여기서 닫는다. 조정까지 가면 그 종목은
	// 그냥 없는 종목처럼 보이고, 짧아진 목록이 관문을 오히려 만족시킨다.
	if absence, lost := batch.Fault(); lost {
		result := fail(StrategyProposalProductionFault)
		result.snapshot.ManifestDigest = digest
		result.snapshot.ProductionFault = absence.String()
		result.snapshot.RefusedCount = result.snapshot.RoutedCount
		return result
	}
	// 한 종목이 여러 가족을 제안하면 시장 조정자가 소유자 범위마다 하나만 고른다.
	// 조정자가 한 범위를 닫으면 시장 전체를 닫는다. 닫힌 종목만 목록에서 빼면
	// 목록이 둘에서 하나로 줄어 아래 파이프라인의 len(entries)==1 관문이 오히려
	// 만족되고, 막으려던 것과 상관없는 *다른* 종목이 대신 풀린다.
	// 목록의 순서는 조정자가 정한다 — 소유자 범위 사전순이라 종목 오름차순이다.
	// 아래 닫힘 가지들이 RefusedCount 에 RoutedCount 를 그대로 넣는 이유:
	// 시장이 닫히면 **경로에 오른 종목 전부**가 제안을 못 낸 것이다. "레인이
	// 없던 종목 수 + 1" 같은 수를 넣으면 10,001 개가 경로에 오르고 하나도
	// 나가지 못한 주기가 "1 건 거절"로 보고된다.
	arbitration, refused := coordinateMarketProposals(loader.accountRef, market, routes.entries, batch, observedAt)
	if arbitration.collision {
		result := fail(StrategyProposalInternalFailure)
		result.snapshot.ManifestDigest = digest
		result.snapshot.ArbitrationDetail = strategyProposalDetailLineageCollision
		result.snapshot.RefusedCount = result.snapshot.RoutedCount
		result.snapshot.QueueDropCount = arbitration.outcome.Drops
		return result
	}
	outcome := arbitration.outcome
	if outcome.Overflow {
		result := fail(StrategyProposalQueueOverflow)
		result.snapshot.ManifestDigest = digest
		result.snapshot.ArbitrationDetail = outcome.Detail
		result.snapshot.RefusedCount = result.snapshot.RoutedCount
		result.snapshot.QueueDropCount = outcome.Drops
		return result
	}
	if outcome.Refusal != strategyarbiter.RefusalNone {
		result := fail(StrategyProposalArbitrationRefused)
		result.snapshot.ManifestDigest = digest
		result.snapshot.ArbitrationRefusal = string(outcome.Refusal)
		result.snapshot.ArbitrationDetail = outcome.Detail
		result.snapshot.RefusedCount = result.snapshot.RoutedCount
		result.snapshot.QueueDropCount = outcome.Drops
		return result
	}
	entries, resolved := arbitration.entries()
	if !resolved {
		result := fail(StrategyProposalInternalFailure)
		result.snapshot.ManifestDigest = digest
		result.snapshot.ArbitrationDetail = strategyProposalDetailUnresolvedSelection
		result.snapshot.RefusedCount = result.snapshot.RoutedCount
		result.snapshot.QueueDropCount = outcome.Drops
		return result
	}
	if len(entries) == 0 {
		result := fail(StrategyProposalNoAcceptedScope)
		result.snapshot.RefusedCount = refused
		result.snapshot.ManifestDigest = digest
		result.snapshot.QueueDropCount = outcome.Drops
		return result
	}
	h := sha256.New()
	for _, entry := range entries {
		_, _ = h.Write([]byte(entry.route.approved.Symbol() + "\x00" + entry.authority.Proposal().Lineage.Identity + "\x00"))
	}
	return strategyProposalMarketAuthority{market: market, entries: entries, snapshot: StrategyProposalMarketSnapshot{Market: market, Ready: true, Reason: StrategyProposalReady,
		RoutedCount: len(routes.entries), ProposedCount: len(entries), RefusedCount: refused, ManifestDigest: digest,
		ProposalSetDigest: "sha256:" + hex.EncodeToString(h.Sum(nil)), QueueDropCount: outcome.Drops}}
}

func failedStrategyProposalPair(observedAt time.Time, reason StrategyProposalReason) strategyProposalAuthorityPair {
	market := func(value StrategyMarket) strategyProposalMarketAuthority {
		return strategyProposalMarketAuthority{market: value, snapshot: StrategyProposalMarketSnapshot{Market: value, Reason: reason}}
	}
	return strategyProposalAuthorityPair{observedAt: observedAt, kr: market(StrategyMarketKR), us: market(StrategyMarketUS)}
}
