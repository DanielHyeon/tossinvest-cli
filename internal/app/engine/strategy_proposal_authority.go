package engine

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/officialfx"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyarbiter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyproposal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyworker"
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
	// 4-가족 관문이 한 소유자 범위의 제안을 **전부** 멈췄다 (태스크 8.8.1).
	//
	// 그 범위만 목록에서 빼면 목록이 짧아지고, 짧아진 목록이 아래 파이프라인의
	// `strategyhandoff.Capacity = 1` 관문을 오히려 만족시켜 상관없는 다른 종목이
	// 대신 풀린다 — 위 PROPOSAL_PRODUCTION_FAULT 와 정확히 같은 기전이다.
	// 그래서 같은 처리를 한다: 시장을 닫는다.
	//
	// 동결 골든의 refusal_enums 에는 엔진 이름이 없으므로 위 두 코드와 마찬가지로
	// 엔진 자신의 이름을 쓴다. 중재 여섯 코드 중 하나를 빌리면 관문이 멈춘 일이
	// 봉인이 깨진 일로 보고된다.
	StrategyProposalFamilyGateClosed StrategyProposalReason = "FAMILY_GATE_CLOSED"
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
	// GatedCount 는 4-가족 관문이 조정자 앞에서 멈춘 제안의 수이고,
	// GatedOutcomes 는 그 결과 종류를 사전순 중복 없이 담은 목록이다
	// (태스크 8.8.1).
	//
	// 수와 종류를 **함께** 내는 이유: 수만 있으면 "왜 줄었는가"에 답할 수 없고,
	// 종류만 있으면 "얼마나"에 답할 수 없다. DORMANT(안 켰다)·LATCHED(고장으로
	// 닫혔다)·REFUSED(주인 없다)는 운영자가 할 조치가 전부 다르다 — 복구 증거가
	// 필요한 상태가 "아직 안 켰다"로 보이면 안 된다.
	//
	// 이 수를 RefusedCount 에 합치지 않는다. RefusedCount 는 **경로에 오른
	// 종목** 중 제안을 못 낸 수이고 이것은 **제안** 수다. 단위가 다른 둘을 더하면
	// 종목 하나가 가족 셋을 냈다가 둘을 잃은 주기를 아무도 읽을 수 없게 된다.
	GatedCount    int
	GatedOutcomes []string
	// QueueDropCount 는 조정자 큐에서 접히거나 들어가지 못한 봉투의 수다.
	// 유계 계수기다.
	//
	// **이 숫자는 지금 배선에서 언제나 0 이다.** 계수기가 오르는 곳은 접힘과
	// 넘침 둘뿐인데, 지금 배선은 둘 다 닿지 못한다. collectMarket 이 종목 중복을
	// 먼저 거절하고 매니페스트가 (종목, 레인)마다 하나만 싣기 때문에 같은 레인
	// 칸이 두 번 오지 않으며(접힘 없음), 큐 상한이 매니페스트 상한과 같아서 정상
	// 매니페스트는 넘칠 수 없다(넘침 없음).
	//
	// 운영자에게는 조정자 투영 자식(`coordinators[].queueDropCount`, a112 7.3)으로
	// 닿는다. 시장 레코드는 여전히 이 시장의 닫힘을 EVIDENCE_STALE 하나로 뭉뚱그린다.
	//
	// 그러므로 이 수를 "조용한 유실이 없음의 증거"로 인용하면 안 된다. 언제나
	// 0 인 계수기는 아무것도 증언하지 않는다. 여러 worker 가 같은 칸을 두고
	// 다투게 만드는 일은 태스크 5.2·5.7 이다.
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
	// activation 은 이 시장의 조정 앞에 선 4-가족 관문이 쓴 서명 활성화다
	// (태스크 5.1.2.2). 영값이면 관문이 서지 않았거나(미선언) 되돌린 채로 섰다
	// (선언했는데 쓸 수 없음, 태스크 8.7.2) — 어느 쪽이든 승격은 없다.
	//
	// 권위와 함께 실어 보내는 이유: 새로 고침 뒤 도는 레인 관측 사이클이 관문과
	// **같은** 승격을 봐야 한다. 두 자리에서 따로 읽으면 만료가 그 사이에
	// 지나갔을 때 관문은 통과시킨 제안을 관측은 DORMANT 로 적는다.
	activation strategyrouter.FamilyActivation
}

// familyActivation 은 이 시장의 관문이 쓴 활성화다. 없으면 영값이다.
func (authority strategyProposalMarketAuthority) familyActivation() strategyrouter.FamilyActivation {
	return authority.activation
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
		// 몇 개까지 넘길 수 있는지는 여기서 정하지 않는다. 경계 한 곳이 정한다.
		// Single 은 값과 함께 "건너가도 되는가"를 돌려주므로, 거절을 안 보고
		// 값을 읽는 판본은 아예 쓸 수 없다.
		//
		// a112 5.2.2.2: 주문 경로와 같은 handoff 목록(dispatchHandoffs)에서 읽는다 — 활성화 없는 시장은 시장 단위 handoff 하나라 오늘과
		// 같고(토글 OFF = upstream), 서명 활성화된 시장은 소유자 범위마다 하나다. 하나라도 유효하지 않으면 그 시장의 결과 권한은 준비 안 됨
		// (경계가 승인한 목록의 일부만 넘기면 위험 · 계좌 권한이 주문 경로와 다른 범위 집합을 보게 된다).
		handoffs := value.dispatchHandoffs()
		results := make([]strategyflow.Result, 0, len(handoffs))
		for _, handoff := range handoffs {
			result, handedOff := handoff.Single()
			if !handedOff || !result.ValidProposal() {
				return strategyResultMarketAuthority{market: market}
			}
			results = append(results, result)
		}
		if len(results) == 0 {
			return strategyResultMarketAuthority{market: market}
		}
		authority := strategyResultMarketAuthority{market: market, ready: true, result: results[0]}
		if len(results) > 1 || value.familyActivation().Verified() {
			authority.scoped = results
		}
		return authority
	}
	return strategyResultAuthorityPair{observedAt: pair.observedAt, kr: convert(StrategyMarketKR, pair.kr), us: convert(StrategyMarketUS, pair.us)}
}

type loadProductionProposalBatch func(context.Context, strategyproposal.ProductionConfig, []strategyproposal.ProductionTarget, officialfx.Evidence) (strategyproposal.ProductionBatchAuthority, error)

type strategyProposalAuthorityLoader struct {
	configDir, evidencePath, journalPath, accountRef string
	getenv                                           func(string) string
	load                                             loadProductionProposalBatch
	// lanes 는 이 프로세스의 여덟 전략군 레인이다 (태스크 5.1.2.2). nil 이면
	// 4-가족 관문이 서지 않고 조정은 오늘과 같은 경로로 돈다.
	//
	// 레인을 여기서 만들지 않고 받는 이유: 레인의 latch 와 연속 실패 계수기는
	// 프로세스 수명 기억이고, 새로 고침마다 만들면 잠긴 레인이 열린 채로
	// 돌아온다(5.1.2.1). 만드는 곳은 `*Context` 하나다.
	lanes *strategyLaneRuntime
	// loadActivation 은 서명된 4-가족 활성화를 읽는 자리다. nil 이면 생산
	// 구현(`loadFamilyActivation`)이 돈다 — 같은 로더의 `load` 필드와 같은
	// 관례다.
	//
	// **이 seam 이 생산 경로를 가리지 않는다는 것을 시험이 지킨다**:
	// `TestTheProductionActivationLoaderRunsAndFindsNoManifest` 가 nil 을 둔 채
	// 실제 구현을 돌려 "매니페스트가 없으면 관문이 서지 않는다"를 잰다. 그 실행이
	// 없으면 "seam 을 건너뛴다" 변이가 모든 시험을 통과한다.
	loadActivation func(context.Context, StrategyMarket, strategyScheduleMarketAuthority,
		strategyRouteMarketAuthority, time.Time) (strategyrouter.FamilyActivation, error)
}

// withStrategyLanes 는 이 로더에 이 프로세스의 레인을 붙인다.
//
// 생성자 인자로 받지 않는 이유: 레인은 원장을 읽어야 세워지고(durable latch),
// 그 읽기는 오류를 낼 수 있다. 생성자에 넣으면 오류가 없는 로더를 만들 수 없다.
func (loader *strategyProposalAuthorityLoader) withStrategyLanes(lanes *strategyLaneRuntime) *strategyProposalAuthorityLoader {
	if loader == nil {
		return nil
	}
	loader.lanes = lanes
	return loader
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
	// gate 를 fail 보다 **먼저** 선언한다. 클로저가 참조로 잡으므로, 아래에서
	// 관문이 계산된 **뒤의** 닫힘 갈래는 전부 그 활성화를 함께 싣는다. 정확히는 13 닫힘 중
	// 첫째(ROUTE_NOT_READY — 관문 계산 전)만 영값을 싣고 나머지 열둘이 관문의 활성화를 싣는다
	// (a112 8.8.4 항목 2 정정 — 앞 판 이 주석은 「모든 닫힘」이라 했으나 관문 계산이 제안 적재 뒤라
	// 일곱이 영값이었다; 표는 TestTheThirteenProposalClosures… 가 못 박는다).
	//
	// 앞 판본은 반환값마다 `carry(...)` 를 부르게 했다. 그것은 이 change 가
	// 반복해서 고쳐 온 **무시할 수 있는 답**이다 — 갈래 하나에서 빠뜨리면
	// 그 주기의 레인 관측이 관문과 다른 승격을 보고, 아무도 그것을 못 본다.
	var gate strategyFamilyGate
	// 관문이 멈춘 것도 gate 와 같은 이유로 클로저 밖에 둔다. 관문이 돌기 전의
	// 닫힘 갈래는 0 과 빈 목록을 싣는데, 그것이 맞는 값이다 — 그 주기에는
	// 관문이 아직 아무것도 멈추지 않았다.
	var gatedCount int
	var gatedOutcomes []string
	fail := func(reason StrategyProposalReason) strategyProposalMarketAuthority {
		return strategyProposalMarketAuthority{market: market, activation: gate.activation,
			snapshot: StrategyProposalMarketSnapshot{Market: market, Reason: reason, RoutedCount: len(routes.entries),
				GatedCount: gatedCount, GatedOutcomes: gatedOutcomes}}
	}
	if !routes.snapshot.Ready || len(routes.entries) == 0 || !schedule.snapshot.Ready || schedule.restore.Activation == nil {
		// 영값을 싣는 유일한 닫힘이다(a112 8.8.4 항목 2): 경로 · 스케줄 권한이 준비되지 않은 주기에는 관문을 계산할 결속 값(경로 매니페스트
		// digest · 보정 · 달력)이 없거나 믿을 수 없어, 관문 계산 **전**에 닫는다.
		return fail(StrategyProposalRouteNotReady)
	}
	// 관문 **계산**은 결속 값이 서는 바로 이 자리다(a112 8.8.4 항목 2 — Manager 판정 Q-B2, 계산/판정 분리). 앞 판은 제안 적재 뒤에
	// 계산해 그 앞 여섯 닫힘(FX · 설정 · 열쇠 · 중복 · 적재 · 고장)이 영값을 실었다 — 그 주기의 레인 관측이 관문과 다른 승격을 봤다.
	// 판정(실패 kind · 우선순위 · FAMILY_GATE_CLOSED 의 자리)은 아래 제자리 그대로이고, 이 값은 닫힘 갈래가 싣는 활성화와 아래 조정
	// 관문으로만 쓰인다. familyGateFor 는 읽기 전용이다(env · 매니페스트 파일 읽기와 검증, 레인 목록 조회 — 원장 · 브로커 · 토글 쓰기 0).
	// 13 닫힘의 순서와 이 자리는 `TestTheThirteenProposalClosuresKeepTheirOrderAndTheGateIsComputedRightAfterRouteReadiness` 가 못 박는다.
	gate = loader.familyGateFor(ctx, market, schedule, routes, observedAt)
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
	// 4-가족 관문은 **조정 앞**에 선다. 뒤에 세우면 중재가 이미 한 범위의
	// 승자를 골라 버렸고, 그 승자의 레인이 잠겨 있으면 그 범위는 이웃 가족이
	// 이길 수 있었는데도 통째로 닫힌다.
	arbitration, refused := coordinateMarketProposals(loader.accountRef, market, routes.entries, batch, observedAt, gate)
	gatedCount, gatedOutcomes = len(arbitration.gated), distinctGatedOutcomes(arbitration.gated)
	// 관문이 한 소유자 범위를 통째로 지웠으면 시장을 닫는다 (태스크 8.8.1).
	//
	// 위 PROPOSAL_PRODUCTION_FAULT 와 같은 처리이고 같은 이유다: 그 범위만 빼면
	// 목록이 짧아지고, 짧아진 목록이 `strategyhandoff.Capacity = 1` 관문을 오히려
	// 만족시켜 상관없는 다른 종목이 대신 풀린다. 8.5 적대 리뷰가 이 경로로 실제
	// 주문이 나가는 것을 실측했다.
	//
	// 중재 결과보다 **먼저** 본다. 뒤에 두면 지워진 범위 때문에 중재가 이미
	// 다른 답을 냈고, 그 답이 무엇이든 이 시장은 닫혀야 한다.
	if arbitration.erasedScopes != 0 {
		result := fail(StrategyProposalFamilyGateClosed)
		result.snapshot.ManifestDigest = digest
		result.snapshot.RefusedCount = result.snapshot.RoutedCount
		result.snapshot.QueueDropCount = arbitration.outcome.Drops
		return result
	}
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
	// 제안 집합 digest 는 A-lite 계약(dispatchHandoffs)이 대조하는 식과 **한 함수**로 적는다(6.2 리뷰 보이스 A #6 — 두 사본이면 한쪽만
	// 바뀌어 활성화 시장이 조용히 닫히거나, 대조가 공허해짐).
	return strategyProposalMarketAuthority{market: market, entries: entries, activation: gate.activation,
		snapshot: StrategyProposalMarketSnapshot{Market: market, Ready: true, Reason: StrategyProposalReady,
			RoutedCount: len(routes.entries), ProposedCount: len(entries), RefusedCount: refused, ManifestDigest: digest,
			ProposalSetDigest: strategyProposalSetDigest(entries), QueueDropCount: outcome.Drops,
			GatedCount: gatedCount, GatedOutcomes: gatedOutcomes}}
}

// distinctGatedOutcomes 는 관문이 낸 결과 종류를 사전순 중복 없이 돌려준다.
//
// 정렬하는 이유는 운영자가 아니라 **비교**다. 순서가 입력 순서를 따라가면 같은
// 상태가 주기마다 다른 문자열로 보고되고, 그러면 두 주기의 스냅샷을 눈으로도
// 시험으로도 대조할 수 없다.
func distinctGatedOutcomes(values []strategyworker.Outcome) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(values))
	kinds := make([]string, 0, len(values))
	for _, value := range values {
		if _, exists := seen[string(value)]; exists {
			continue
		}
		seen[string(value)] = struct{}{}
		kinds = append(kinds, string(value))
	}
	sort.Strings(kinds)
	return kinds
}

func failedStrategyProposalPair(observedAt time.Time, reason StrategyProposalReason) strategyProposalAuthorityPair {
	market := func(value StrategyMarket) strategyProposalMarketAuthority {
		return strategyProposalMarketAuthority{market: value, snapshot: StrategyProposalMarketSnapshot{Market: value, Reason: reason}}
	}
	return strategyProposalAuthorityPair{observedAt: observedAt, kr: market(StrategyMarketKR), us: market(StrategyMarketUS)}
}
