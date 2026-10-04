//go:build tossos_testseams

package engine

// a112 태스크 8.8.4 항목 5(Manager 판정 2026-10-04): 레인 런타임의 `promotion` 인자를 **검증된 활성화**로 몰아, 승격이 실제로 일어남을
// (소유 입력을 받은 레인이 EMITTED) 잰다. 앞선 시험들은 검증된 활성화로 돌 때 입력이 없거나(REFUSED — 봉인 불일치,
// `TestTheLaneDesiredAndEffectiveAreTheActivationTheWaveRanWith`) 영값 활성화만 넘겼다(DORMANT). 대조군: 같은 입력 · 같은 런타임에
// 영값 활성화면 같은 레인이 DORMANT 이고 봉투 0 이다 — EMITTED 가 활성화 때문임을 보인다.
// 생산 빌드의 `laneStepFor`(strategy_lane_step.go)를 무태그로 `evaluate` 경유 도는 시험은 따로 있다(예: a112_lane_runtime_test.go
// `TestEveryFamilyLaneIsDormantUntilASignedManifestPromotesIt` · `TestTheProductionStepNeverLatchesSoTheLedgerStaysEmpty` — 커버리지 실측).

import (
	"context"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyworker"
)

func TestAVerifiedPromotionLetsTheOwningLaneEmitThroughTheLaneRuntime(t *testing.T) {
	inputs := laneOwnedInputs(t)
	owning := map[strategyworker.Key]bool{}
	for _, lane := range laneRuntimeForDependency(t).lanesFor(StrategyMarketKR) {
		for _, input := range inputs {
			if lane.Owns(input.Proposal) {
				owning[lane.Key()] = true
			}
		}
	}
	if len(owning) != 1 {
		t.Fatalf("arrangement: %d KR lanes own the sealed proposals, want exactly one", len(owning))
	}
	for _, tc := range []struct {
		name       string
		activation strategyrouter.FamilyActivation
		want       strategyworker.Outcome
	}{
		{"verified activation of all four KR families", strategyrouter.FamilyActivationForTest(strategyrouter.MarketKR, 1,
			strategyrouter.AllFourFamiliesForTest(strategyrouter.MarketKR)), strategyworker.OutcomeEmitted},
		{"control: no activation", strategyrouter.FamilyActivation{}, strategyworker.OutcomeDormant},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runtime := laneRuntimeForDependency(t)
			if err := runtime.evaluate(context.Background(), StrategyMarketKR, 1, tc.activation, inputs, strategyShadowBatch{}); err != nil {
				t.Fatal(err)
			}
			emitted := 0
			for _, observation := range runtime.observations() {
				if observation.Emitted {
					emitted++
				}
				if !owning[observation.Key] {
					continue
				}
				wantState := strategyrouter.StateOff
				if tc.want == strategyworker.OutcomeEmitted {
					wantState = strategyrouter.StateOn
				}
				if observation.Outcome != tc.want || observation.Emitted != (tc.want == strategyworker.OutcomeEmitted) ||
					observation.Desired != wantState || observation.Effective != wantState {
					t.Fatalf("owning lane %v: outcome=%s emitted=%v desired=%s effective=%s detail=%q, want %s with %s/%s",
						observation.Key.Parts(), observation.Outcome, observation.Emitted, observation.Desired, observation.Effective,
						observation.Detail, tc.want, wantState, wantState)
				}
			}
			if want := map[bool]int{true: 1, false: 0}[tc.want == strategyworker.OutcomeEmitted]; emitted != want {
				t.Fatalf("lanes that emitted=%d, want %d (only the owning lane, only under the verified activation)", emitted, want)
			}
		})
	}
}
