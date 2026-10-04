package engine

import (
	"sort"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojection"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyworker"
)

// a112 0.5 리뷰 유지#5: worker 의 반사실 판정 어휘와 projection 의 shadowOutcome 어휘를 **직접** 대조함. 앞 판은 worker 시험이 상수
// 자신을, projection 시험이 자기 문자열을 써서 순환이었고, `ShadowWouldEmit` 값을 "WOULD_FIRE" 로 바꾼 변이(M3)를 두 패키지 · httpapi
// 시험이 전부 통과시켰음(엔진에서 투영 Validate 가 스냅숏 전체를 거절해 간접으로만 잡힘). 엔진은 worker 값을 projection 값으로
// 그대로 옮기는(`strategyprojection.LaneShadowOutcome(shadow.outcome)`) 유일한 자리라 두 어휘가 같아야 함.
func TestTheWorkerShadowVocabularyIsTheProjectionShadowVocabulary(t *testing.T) {
	var worker []string
	for _, outcome := range strategyworker.ShadowOutcomes() {
		worker = append(worker, string(outcome))
	}
	projection := append([]string(nil), strategyprojection.LaneShadowOutcomes()...)
	sort.Strings(worker)
	sort.Strings(projection)
	if len(worker) == 0 || strings.Join(worker, ",") != strings.Join(projection, ",") {
		t.Fatalf("worker shadow outcomes %v, projection shadow outcomes %v — the engine copies one into the other", worker, projection)
	}
	// 양쪽 목록이 자기 상수를 빠짐없이 담는지도 셈 — 목록 함수에서 상수 하나를 빼도 위 등식은 양쪽이 같이 빼면 통과하므로.
	for _, want := range []strategyworker.ShadowOutcome{strategyworker.ShadowWouldEmit, strategyworker.ShadowNotThisLane, strategyworker.ShadowNoInput} {
		if !strings.Contains(","+strings.Join(worker, ",")+",", ","+string(want)+",") {
			t.Errorf("ShadowOutcomes() lacks %s", want)
		}
	}
	for _, want := range []strategyprojection.LaneShadowOutcome{strategyprojection.LaneShadowWouldEmit, strategyprojection.LaneShadowNotThisLane,
		strategyprojection.LaneShadowNoInput} {
		if !strings.Contains(","+strings.Join(projection, ",")+",", ","+string(want)+",") {
			t.Errorf("LaneShadowOutcomes() lacks %s", want)
		}
	}
}
