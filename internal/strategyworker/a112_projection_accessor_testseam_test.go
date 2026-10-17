//go:build tossos_testseams

package strategyworker

import (
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

// a112 7.3: 서명 활성화가 KR 네 가족을 켜면 KR 레인 넷만 ON 으로 읽히고 US 넷은 OFF 다 — 접근자가 레인 자기 열쇠로 묻는다
// (호출자가 시장 · 가족을 고르지 않음). 상수 OFF 를 돌려주는 변이는 이 시험이 잡는다.
func TestTheLaneDesiredAndEffectiveFollowTheSignedActivationForItsOwnKey(t *testing.T) {
	activation := strategyrouter.FamilyActivationForTest(strategyrouter.MarketKR, 1, strategyrouter.AllFourFamiliesForTest(strategyrouter.MarketKR))
	on := 0
	for _, lane := range ProductionLanes(clock.NewFake(laneNow)) {
		want := strategyrouter.StateOff
		if lane.Key().Market == strategyrouter.MarketKR {
			want = strategyrouter.StateOn
			on++
		}
		if lane.Desired(activation) != want || lane.Effective(activation) != want {
			t.Fatalf("lane %v desired=%s effective=%s, want %s", lane.Key(), lane.Desired(activation), lane.Effective(activation), want)
		}
	}
	if on != 4 {
		t.Fatalf("KR lanes=%d, want 4", on)
	}
}
