//go:build tossos_testseams

package strategyworker

// a112 태스크 8.8.4 항목 3(Manager 판정 2026-10-04): 골든의 desired/effective 단언을 공허하지 않게 바꾼다. 앞 판(golden_contract_test.go)은
// 영값 활성화로 물었다 — 영값은 `FamilyActivation.lookup` 이 첫 줄에서 OFF 를 돌려주므로 골든의 OFF 와 상수-대-상수였다.
// 여기서는 골든이 얼린 OFF 가 **각 worker 열쇠의 기본값**이며 서명 활성화가 정확히 그 열쇠에서만 그것을 ON 으로 뒤집음을 잰다:
//   - 활성화 없음 → 여덟 다 골든 값(OFF/OFF).
//   - worker i 의 레인 하나만 승격한 그 시장의 활성화 → worker i 만 ON/ON, 나머지 일곱은 골든 값 그대로(다른 시장 넷 포함).
//   - 다른 시장의 네 레인을 전부 승격한 활성화 → 이 worker 는 골든 값(시장 대조 · 레인 ID 시장별 분리).
// 마지막 둘이 영값 lookup 의 단락이 아니라 열쇠 대조가 답을 정함을 보인다.

import (
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

func TestTheGoldenOffIsEachWorkersDefaultAndOnlyItsSignedActivationFlipsIt(t *testing.T) {
	golden := readGolden(t)
	workers := ProductionWorkers()
	if len(workers) != len(golden.Descriptors) || len(workers) != 8 {
		t.Fatalf("workers=%d golden descriptors=%d, want the eight", len(workers), len(golden.Descriptors))
	}
	for index, want := range golden.Descriptors {
		if want.Desired != string(strategyrouter.StateOff) || want.Effective != string(strategyrouter.StateOff) {
			t.Fatalf("golden descriptor %d is %s/%s — this test reads OFF as the frozen default", index, want.Desired, want.Effective)
		}
	}
	assert := func(name string, activation strategyrouter.FamilyActivation, on func(int) bool) {
		t.Helper()
		for index, worker := range workers {
			wantDesired, wantEffective := golden.Descriptors[index].Desired, golden.Descriptors[index].Effective
			if on(index) {
				wantDesired, wantEffective = string(strategyrouter.StateOn), string(strategyrouter.StateOn)
			}
			if string(worker.Desired(activation)) != wantDesired || string(worker.Effective(activation)) != wantEffective {
				t.Errorf("%s: worker %d (%v) is %s/%s, want %s/%s", name, index, worker.Key().Parts(),
					worker.Desired(activation), worker.Effective(activation), wantDesired, wantEffective)
			}
		}
	}
	assert("no activation", strategyrouter.FamilyActivation{}, func(int) bool { return false })
	for index, worker := range workers {
		key := worker.Key()
		promoted := map[string]bool{key.LaneID: true}
		own := strategyrouter.FamilyActivationForTest(key.Market, 1, promoted)
		if !own.Verified() {
			t.Fatalf("arrangement: the seam did not mint a verified activation for %v", key.Parts())
		}
		assert("own market, one lane promoted", own, func(i int) bool { return i == index })
		other := strategyrouter.MarketUS
		if key.Market == strategyrouter.MarketUS {
			other = strategyrouter.MarketKR
		}
		// 다른 시장의 네 레인을 전부 켠 활성화: 그 시장 넷만 ON, 이 worker 를 포함한 이 시장 넷은 골든 값.
		assert("other market with all four of its lanes promoted", strategyrouter.FamilyActivationForTest(other, 1, strategyrouter.AllFourFamiliesForTest(other)),
			func(i int) bool { return workers[i].Key().Market == other })
	}
	// 위 「다른 시장」 단언은 활성화의 시장 대조(`lookup` 의 `market != activation.market`)와 레인 ID 의 시장별 분리 **둘 다**가 지킨다 —
	// 시장 대조를 지우는 변이(A1)는 레인 ID 가 시장마다 달라서 동등이다. 그 동등이 우연이 아니게, 분리 자체를 못 박는다: 한 레인 ID 가
	// 두 시장에 나타나면 시장 대조가 유일한 방어가 되므로 이 단언이 먼저 뒤집혀 검토를 강제한다.
	byID := map[string]strategyrouter.Market{}
	for _, worker := range workers {
		key := worker.Key()
		if market, seen := byID[key.LaneID]; seen && market != key.Market {
			t.Fatalf("lane id %q appears in %s and %s — the market check in FamilyActivation.lookup becomes the only guard", key.LaneID, market, key.Market)
		}
		byID[key.LaneID] = key.Market
	}
}
