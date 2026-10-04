//go:build tossos_testseams

package engine

// a112 8.5 응답 로트(Manager 최종 판정 2026-10-04).
//
// ① F1(보이스 2 P1 · 보이스 1 P2-1, 판정 Y): 8.8.4 항목 2 가 관문 계산을 경로 준비 가드 직후로 **옮긴** 뒤, 관문의 스냅숏이 제안 적재 시간만큼
// 낡았다 — `familyGateFor` 의 입력 중 시간에 대해 변하는 둘(ctx 취소 · 디스크의 매니페스트)이 적재 중에 바뀌면 그 주기 판정에 안 보였다
// (편집 전 FAMILY_GATE_CLOSED · handoff 0 → 편집 뒤 READY · handoff 2, 보이스 2 실측). 철회 쪽은 하류에 다시 막는 자리가 없다(활성화를
// 다시 읽는 소비자 0 — grep). Y: 조기 계산은 조정 앞 닫힘 여섯의 **진단** carry 로 두고, 판정은 편집 전 자리(`coordinateMarketProposals`
// 바로 앞)에서 다시 계산한다 — 같은 함수 · 같은 순간이라 판정 동등성이 구성으로 선다.
//
// X(조정 직전 `ctx.Err()` 재확인)를 고르지 않은 이유가 시험 2 의 첫 모양이다: X 는 취소만 닫고 철회를 남긴다. 또 X 는 「선언된 시장인가」를
// 다시 판별해야 해 미선언 규칙의 사본이 둘째 판정 자리에 생긴다(시험 3 이 미선언 쪽을 잰다).
//
// ② P2-a(codex r2): getenv 가 nil 인 적재기는 활성화 적재에서 공황했고 collect 의 recover 가 그것을 INTERNAL_FAILURE 로 바꿔, 같은 주기의
// FX 미준비 사유를 덮었다. 이제 nil getenv 는 Unavailable(선언됐는데 쓸 수 없음 → 관문 되돌림)이다.

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reversallane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyproposal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

type a112DecisionRun struct {
	authority strategyProposalMarketAuthority
	loads     int
	handed    int
}

// a112CollectWithActivationAnswers 는 KR 한 시장을 관문 아래에서 수집한다. answer 는 활성화 적재가 n 번째 호출에 낼 답이고, onLoad 는 제안
// 적재가 배치를 만든 **뒤**(반환 직전)에 돈다 — 적재 중에 일어나는 취소 · 철회의 자리다.
func a112CollectWithActivationAnswers(t *testing.T, ctx context.Context, lanes []string, fxReady bool, onLoad func(),
	answer func(call int, ctx context.Context) (strategyrouter.FamilyActivation, error),
) a112DecisionRun {
	t.Helper()
	now := time.Date(2026, 9, 3, 1, 2, 3, 0, time.UTC)
	runtime, _ := familyGateFixture(t)
	loader := testStrategyProposalLoader(t).withStrategyLanes(runtime)
	loader.load = func(_ context.Context, config strategyproposal.ProductionConfig, targets []strategyproposal.ProductionTarget, _ interfaceOfficialFX) (strategyproposal.ProductionBatchAuthority, error) {
		batch := arbitrationBatch(t, config, targets, now, "005930", lanes)
		if onLoad != nil {
			onLoad()
		}
		return batch, nil
	}
	calls := 0
	loader.loadActivation = func(c context.Context, _ StrategyMarket, _ strategyScheduleMarketAuthority, _ strategyRouteMarketAuthority, _ time.Time) (strategyrouter.FamilyActivation, error) {
		calls++
		return answer(calls, c)
	}
	fx := proposalFXPair(now).forMarket(StrategyMarketKR)
	if !fxReady {
		fx.snapshot.Ready = false
	}
	got := loader.collectMarket(ctx, routeReadySchedulePair(now).forMarket(StrategyMarketKR),
		arbitrationRoutePair(t, now, familyScoresForTest(strategyrouter.MarketKR), "005930", lanes...).forMarket(StrategyMarketKR), fx, now)
	handed := 0
	for _, handoff := range got.dispatchHandoffs() {
		if _, ok := handoff.Single(); ok {
			handed++
		}
	}
	return a112DecisionRun{authority: got, loads: calls, handed: handed}
}

func a112ActivationGeneration(generation uint64) strategyrouter.FamilyActivation {
	return strategyrouter.FamilyActivationForTest(strategyrouter.MarketKR, generation, strategyrouter.AllFourFamiliesForTest(strategyrouter.MarketKR))
}

// a112LaneShapes 는 보이스 2 가 잰 두 모양(한 레인 · 여러 가족)이다.
var a112LaneShapes = map[string][]string{
	"one lane":     {continuationlane.KRContinuationLaneID},
	"two families": {continuationlane.KRContinuationLaneID, reversallane.KRReversalLaneID},
}

// 시험 1(보이스 2 ctx-race 승격): 제안 적재가 끝난 직후 주기 ctx 가 취소되면, 판정 관문(생산 적재기처럼 취소된 ctx 에 ctx 오류를 내는)이
// 그것을 보고 시장을 닫는다 — 편집 전과 같은 답.
func TestACancelDuringTheProposalLoadClosesTheMarketAtTheDecision(t *testing.T) {
	for name, lanes := range a112LaneShapes {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			run := a112CollectWithActivationAnswers(t, ctx, lanes, true, cancel, func(_ int, c context.Context) (strategyrouter.FamilyActivation, error) {
				if err := c.Err(); err != nil {
					return strategyrouter.FamilyActivation{}, err
				}
				return a112ActivationGeneration(7), nil
			})
			snapshot := run.authority.snapshot
			if snapshot.Ready || snapshot.Reason != StrategyProposalFamilyGateClosed {
				t.Fatalf("ready=%v reason=%s, want closed with %s — a cancel during the proposal load must reach the decision",
					snapshot.Ready, snapshot.Reason, StrategyProposalFamilyGateClosed)
			}
			if snapshot.ProposedCount != 0 || run.handed != 0 || len(run.authority.entries) != 0 {
				t.Fatalf("proposed=%d handed=%d entries=%d, want 0, 0, 0", snapshot.ProposedCount, run.handed, len(run.authority.entries))
			}
			if run.authority.familyActivation().Verified() {
				t.Fatal("the closed market carries a verified activation — the decision gate rolled back")
			}
			if run.loads != 2 {
				t.Fatalf("activation loads=%d, want 2 (diagnostic + decision)", run.loads)
			}
		})
	}
}

// 시험 2(X/Y 를 가르는 판별 시험): 판정 관문은 **적재 뒤의** 디스크를 본다.
func TestARevocationDuringTheProposalLoadClosesTheMarketAtTheDecision(t *testing.T) {
	for name, lanes := range a112LaneShapes {
		t.Run(name+"/revoked mid-load", func(t *testing.T) {
			// 1회차(조기 · 진단) 검증 gen 7, 2회차(판정) 철회 — X(ctx 재확인)는 이 모양을 READY 로 통과시킨다.
			run := a112CollectWithActivationAnswers(t, context.Background(), lanes, true, nil, func(call int, _ context.Context) (strategyrouter.FamilyActivation, error) {
				if call == 1 {
					return a112ActivationGeneration(7), nil
				}
				return strategyrouter.FamilyActivation{}, strategyrouter.ErrProductionFamilyActivationRevoked
			})
			snapshot := run.authority.snapshot
			if snapshot.Ready || snapshot.Reason != StrategyProposalFamilyGateClosed || run.handed != 0 {
				t.Fatalf("ready=%v reason=%s handed=%d, want closed with %s and nothing handed off",
					snapshot.Ready, snapshot.Reason, run.handed, StrategyProposalFamilyGateClosed)
			}
			if run.authority.familyActivation().Verified() {
				t.Fatal("the closed market carries the stale (pre-revocation) activation — it must carry the decision gate's")
			}
		})
		t.Run(name+"/replaced mid-load", func(t *testing.T) {
			// 1회차 gen 7, 2회차 gen 8: 조정 · 성공은 판정 값(gen 8)을 싣는다.
			run := a112CollectWithActivationAnswers(t, context.Background(), lanes, true, nil, func(call int, _ context.Context) (strategyrouter.FamilyActivation, error) {
				return a112ActivationGeneration(uint64(6 + call)), nil
			})
			if !run.authority.snapshot.Ready {
				t.Fatalf("reason=%s, want READY", run.authority.snapshot.Reason)
			}
			if got := run.authority.familyActivation().Generation(); got != 8 || run.loads != 2 {
				t.Fatalf("carried generation=%d loads=%d, want 8 (the decision gate's) and 2", got, run.loads)
			}
		})
	}
	t.Run("a closure before arbitration carries the diagnostic value", func(t *testing.T) {
		// 같은 답 순서에서 FX 미준비 닫힘은 조정에 닿지 않는다 — 적재 1 회, 조기(진단) 값 gen 7. 잔여 비대칭(관측 전용)의 자리다.
		run := a112CollectWithActivationAnswers(t, context.Background(), a112LaneShapes["one lane"], false, nil, func(call int, _ context.Context) (strategyrouter.FamilyActivation, error) {
			return a112ActivationGeneration(uint64(6 + call)), nil
		})
		if run.authority.snapshot.Reason != StrategyProposalFXNotReady || run.loads != 1 {
			t.Fatalf("reason=%s loads=%d, want %s and 1", run.authority.snapshot.Reason, run.loads, StrategyProposalFXNotReady)
		}
		if got := run.authority.familyActivation().Generation(); got != 7 {
			t.Fatalf("carried generation=%d, want 7 (the early, diagnostic gate's)", got)
		}
	})
}

// 시험 3(토글 OFF): 미선언 시장은 적재 중 취소에도 기존 경로 그대로다 — 취소 없는 같은 주기와 스냅숏이 같고, 관문이 아무것도 멈추지 않는다.
// 생산 적재기는 미선언을 ctx 보다 먼저 본다(`LoadProductionFamilyActivation` 첫 줄) — 이 스텁도 ctx 를 보지 않는다.
func TestAnUndeclaredMarketCancelledDuringTheLoadKeepsTheLegacyPath(t *testing.T) {
	undeclared := func(int, context.Context) (strategyrouter.FamilyActivation, error) {
		return strategyrouter.FamilyActivation{}, strategyrouter.ErrProductionFamilyActivationUndeclared
	}
	for name, lanes := range a112LaneShapes {
		t.Run(name, func(t *testing.T) {
			plain := a112CollectWithActivationAnswers(t, context.Background(), lanes, true, nil, undeclared)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			cancelled := a112CollectWithActivationAnswers(t, ctx, lanes, true, cancel, undeclared)
			if cancelled.authority.snapshot.Reason == StrategyProposalFamilyGateClosed || cancelled.authority.snapshot.GatedCount != 0 {
				t.Fatalf("undeclared market closed by the gate on a cancelled cycle: %+v", cancelled.authority.snapshot)
			}
			if !reflect.DeepEqual(plain.authority.snapshot, cancelled.authority.snapshot) || plain.handed != cancelled.handed {
				t.Fatalf("cancel changed the undeclared (legacy) path:\n plain     %+v handed=%d\n cancelled %+v handed=%d",
					plain.authority.snapshot, plain.handed, cancelled.authority.snapshot, cancelled.handed)
			}
			if cancelled.authority.familyActivation().Verified() {
				t.Fatal("an undeclared market carries a verified activation")
			}
		})
	}
}

// ② P2-a: nil getenv 는 공황이 아니라 Unavailable(관문 되돌림)이고, 같은 주기의 FX 미준비 사유를 덮지 않는다(동시 다중 실패).
func TestANilEnvironmentReaderRollsTheGateBackInsteadOfPanicking(t *testing.T) {
	now := time.Date(2026, 9, 3, 1, 2, 3, 0, time.UTC)
	routes := arbitrationRoutePair(t, now, familyScoresForTest(strategyrouter.MarketKR), "005930", continuationlane.KRContinuationLaneID)
	schedule := routeReadySchedulePair(now)
	t.Run("the loader answers Unavailable", func(t *testing.T) {
		loader := testStrategyProposalLoader(t)
		loader.getenv = nil
		var err error
		func() {
			defer func() {
				if recovered := recover(); recovered != nil {
					t.Fatalf("loadFamilyActivation panicked on a nil getenv: %v", recovered)
				}
			}()
			_, err = loader.loadFamilyActivation(context.Background(), StrategyMarketKR, schedule.forMarket(StrategyMarketKR), routes.forMarket(StrategyMarketKR), now)
		}()
		if !errors.Is(err, strategyrouter.ErrProductionFamilyActivationUnavailable) || errors.Is(err, strategyrouter.ErrProductionFamilyActivationUndeclared) {
			t.Fatalf("err=%v, want Unavailable and not Undeclared (a nil reader cannot prove the market undeclared)", err)
		}
	})
	t.Run("nil getenv and FX not ready report FX_NOT_READY", func(t *testing.T) {
		runtime, _ := familyGateFixture(t)
		loader := testStrategyProposalLoader(t).withStrategyLanes(runtime)
		loader.getenv = nil
		fx := proposalFXPair(now)
		fx.kr.snapshot.Ready = false
		pair := loader.collect(context.Background(), schedule, routes, fx)
		if pair.kr.snapshot.Reason != StrategyProposalFXNotReady {
			t.Fatalf("KR reason=%s, want %s (the activation load must not turn the cycle into an internal failure)", pair.kr.snapshot.Reason, StrategyProposalFXNotReady)
		}
		if pair.kr.familyActivation().Verified() {
			t.Fatal("a nil environment reader produced a verified activation")
		}
	})
}
