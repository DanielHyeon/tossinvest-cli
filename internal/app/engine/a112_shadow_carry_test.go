//go:build tossos_testseams

package engine

// a112 7.3.1 SHADOW 운반(브리프 v3.3 §4).
//
// 반사실 입력은 관문 **앞** 의 전체 제안이다(선언 시장에서 OFF 가족만 제안한 종목이 SHADOW 의 대상 모집단). 그 묶음은 authority 밖 별도 값으로
// 조정자 → collectMarket → collect → 조립까지 나란히 간다. 「관측 없음」 은 빈 묶음이 아니라 **부재 값**이다:
//   - 조정 앞 닫힘 일곱과 계보 충돌(루프 도중 return — 부분 묶음) → 부재 값.
//   - 조정 뒤 닫힘 다섯과 성공 → 루프를 끝까지 돈 묶음.
// 계보 충돌 갈래는 입력으로 닿지 않는다(조정자가 같은 신원을 접음 — a112_proposal_closure_carriage_test 주석) — 부재 값 리터럴은
// a112_shadow_structure_test 의 AST 핀 ⓑ 가 잰다.

import (
	"context"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reversallane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyproposal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

var a112ShadowNow = time.Date(2026, 9, 3, 1, 2, 3, 0, time.UTC)

// a112ShadowCoordinated 는 KR 에서 005930 이 laneIDs 로, 000660 이 continuation 으로 제안한 배치를 관문 gate 아래 조정한 결과다.
func a112ShadowCoordinated(t *testing.T, gate strategyFamilyGate, laneIDs ...string) (strategyMarketArbitration, strategyShadowBatch) {
	t.Helper()
	routes := arbitrationRoutePair(t, a112ShadowNow, familyScoresForTest(strategyrouter.MarketKR), "005930", laneIDs...).forMarket(StrategyMarketKR)
	targets := make([]strategyproposal.ProductionTarget, 0, len(routes.entries))
	for _, entry := range routes.entries {
		targets = append(targets, strategyproposal.ProductionTarget{Approved: entry.approved, Router: entry.route.Request()})
	}
	config := strategyproposal.ProductionConfig{AccountRef: "acct", Market: strategyrouter.MarketKR, ManifestDigest: "sha256:proposal-KR"}
	batch := arbitrationBatch(t, config, targets, a112ShadowNow, "005930", laneIDs)
	arbitration, _, shadow := coordinateMarketProposals("acct", StrategyMarketKR, routes.entries, batch, a112ShadowNow, gate)
	return arbitration, shadow
}

// 정상 경로: 관문 앞 제안 전부(관문이 멈춘 OFF 가족의 제안 포함)를 싣고 관측 표시가 선다.
func TestTheCoordinatorCarriesEveryPreGateProposalAsTheShadowBatch(t *testing.T) {
	runtime, _ := familyGateFixture(t)
	// 선언 시장, REVERSAL 만 ON: 005930 의 continuation 제안은 관문이 멈춘다(OFF-단독 종목).
	onlyReversal := strategyrouter.FamilyActivationForTest(strategyrouter.MarketKR, 1, map[string]bool{reversallane.KRReversalLaneID: true})
	gate := strategyFamilyGate{activation: onlyReversal, lanes: runtime.lanesFor(StrategyMarketKR)}
	arbitration, shadow := a112ShadowCoordinated(t, gate, continuationlane.KRContinuationLaneID)
	if len(arbitration.gated) == 0 {
		t.Fatal("arrangement: the gate stopped nothing, so the OFF-only population is not exercised")
	}
	if !shadow.observed || len(shadow.inputs) != 2 {
		t.Fatalf("shadow batch observed=%v inputs=%d, want observed with both pre-gate proposals (gated ones included)", shadow.observed, len(shadow.inputs))
	}
	// 미선언(오늘 생산): 관문 없음 — 같은 묶음이 선다.
	if _, undeclared := a112ShadowCoordinated(t, strategyFamilyGate{}, continuationlane.KRContinuationLaneID); !undeclared.observed || len(undeclared.inputs) != 2 {
		t.Fatalf("undeclared market shadow batch observed=%v inputs=%d, want 2", undeclared.observed, len(undeclared.inputs))
	}
}

// collectMarket: 조정 앞 닫힘은 부재 값, 조정 뒤 닫힘(FAMILY_GATE_CLOSED — OFF-단독 종목)과 성공은 묶음을 싣고 결속 설정이 붙는다.
func TestCollectMarketCarriesTheShadowBatchOnlyFromCoordination(t *testing.T) {
	runtime, _ := familyGateFixture(t)
	onlyReversal := strategyrouter.FamilyActivationForTest(strategyrouter.MarketKR, 1, map[string]bool{reversallane.KRReversalLaneID: true})
	collect := func(t *testing.T, activation strategyrouter.FamilyActivation, err error, mutate func(*strategyRouteMarketAuthority, *strategyFXMarketAuthority)) (strategyProposalMarketAuthority, strategyShadowBatch) {
		t.Helper()
		loader := testStrategyProposalLoader(t).withStrategyLanes(runtime)
		loader.load = func(_ context.Context, config strategyproposal.ProductionConfig, targets []strategyproposal.ProductionTarget, _ interfaceOfficialFX) (strategyproposal.ProductionBatchAuthority, error) {
			return arbitrationBatch(t, config, targets, a112ShadowNow, "005930", []string{continuationlane.KRContinuationLaneID}), nil
		}
		loader.loadActivation = func(context.Context, StrategyMarket, strategyScheduleMarketAuthority, strategyRouteMarketAuthority, time.Time) (strategyrouter.FamilyActivation, error) {
			return activation, err
		}
		routes := arbitrationRoutePair(t, a112ShadowNow, familyScoresForTest(strategyrouter.MarketKR), "005930", continuationlane.KRContinuationLaneID).forMarket(StrategyMarketKR)
		fx := proposalFXPair(a112ShadowNow).forMarket(StrategyMarketKR)
		if mutate != nil {
			mutate(&routes, &fx)
		}
		shadow := strategyShadowBatch{observed: true, inputs: nil} // 쓰레기 시작값 — collectMarket 이 덮어써야 한다
		got := loader.collectMarket(context.Background(), routeReadySchedulePair(a112ShadowNow).forMarket(StrategyMarketKR), routes, fx, a112ShadowNow, &shadow)
		return got, shadow
	}
	// 조정 앞 닫힘: FX 미준비 → 부재 값.
	closed, shadow := collect(t, onlyReversal, nil, func(_ *strategyRouteMarketAuthority, fx *strategyFXMarketAuthority) { fx.snapshot.Ready = false })
	if closed.snapshot.Reason != StrategyProposalFXNotReady || shadow.observed || len(shadow.inputs) != 0 {
		t.Fatalf("pre-coordination closure %s carried shadow observed=%v inputs=%d, want the absent value", closed.snapshot.Reason, shadow.observed, len(shadow.inputs))
	}
	// 조정 뒤 닫힘: OFF-단독 종목이 범위를 지움 → FAMILY_GATE_CLOSED, 묶음은 싣는다.
	gated, shadow := collect(t, onlyReversal, nil, nil)
	if gated.snapshot.Reason != StrategyProposalFamilyGateClosed {
		t.Fatalf("arrangement: reason=%s, want FAMILY_GATE_CLOSED (OFF-only symbol)", gated.snapshot.Reason)
	}
	if !shadow.observed || len(shadow.inputs) != 2 {
		t.Fatalf("FAMILY_GATE_CLOSED carried shadow observed=%v inputs=%d, want the full pre-gate batch", shadow.observed, len(shadow.inputs))
	}
	if shadow.config.Market != strategyrouter.MarketKR || shadow.config.ConfigDir == "" || shadow.config.RouteManifestDigest == "" {
		t.Fatalf("the carried batch lacks its binding config: %+v", shadow.config)
	}
	// 성공(미선언 시장): 묶음을 싣는다.
	ready, shadow := collect(t, strategyrouter.FamilyActivation{}, strategyrouter.ErrProductionFamilyActivationUndeclared, nil)
	if !ready.snapshot.Ready || !shadow.observed || len(shadow.inputs) != 2 {
		t.Fatalf("success reason=%s carried shadow observed=%v inputs=%d, want the batch", ready.snapshot.Reason, shadow.observed, len(shadow.inputs))
	}
}

// collect 는 시장별 묶음을 authority 짝 옆의 별도 짝으로 돌려준다. authority 에는 shadow 가 없다(§2 ② census 가 타입으로도 잰다).
func TestCollectReturnsTheShadowPairBesideTheAuthorityPair(t *testing.T) {
	loader := testStrategyProposalLoader(t)
	loader.load = func(_ context.Context, config strategyproposal.ProductionConfig, targets []strategyproposal.ProductionTarget, _ interfaceOfficialFX) (strategyproposal.ProductionBatchAuthority, error) {
		return arbitrationBatch(t, config, targets, a112ShadowNow, "005930", []string{continuationlane.KRContinuationLaneID}), nil
	}
	pair, shadow := loader.collect(context.Background(), routeReadySchedulePair(a112ShadowNow),
		arbitrationRoutePair(t, a112ShadowNow, familyScoresForTest(strategyrouter.MarketKR), "005930", continuationlane.KRContinuationLaneID),
		proposalFXPair(a112ShadowNow))
	if !pair.kr.snapshot.Ready {
		t.Fatalf("arrangement: KR reason=%s", pair.kr.snapshot.Reason)
	}
	if !shadow.forMarket(StrategyMarketKR).observed || !shadow.forMarket(StrategyMarketUS).observed {
		t.Fatalf("shadow pair KR=%v US=%v, want both observed", shadow.forMarket(StrategyMarketKR).observed, shadow.forMarket(StrategyMarketUS).observed)
	}
	if shadow.forMarket("JP").observed {
		t.Fatal("an unknown market must get the absent value")
	}
}

// a112PairOnly 는 collect 의 두 반환값 중 authority 짝만 고른다 — shadow 짝을 재지 않는 기존 시험이 쓴다(a112 7.3.1 서명 변경).
func a112PairOnly(pair strategyProposalAuthorityPair, _ strategyShadowPair) strategyProposalAuthorityPair {
	return pair
}
