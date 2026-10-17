package strategyprojection

// a112 7.3.1 SHADOW(브리프 v3.3 §6 · §7 · §9) — SHADOW 는 projection 어휘에만 있고, 외부 경계 검증기(validateLane)는 교차 규칙 하나로
// 그 값을 묶는다: runtime=SHADOW ⇒ desired=OFF ∧ effective=OFF ∧ health≠nil ∧ cycleGeneration>0, 그리고 shadowOutcome≠null ⇔ runtime=SHADOW.
// 거부되는 정상 입력은 표로 적는다(fail-closed 는 무엇을 거부하는지 말해야 한다).

import (
	"reflect"
	"testing"
)

// a112ShadowLane 은 관측된 KR continuation 레인을 SHADOW · WOULD_EMIT 으로 둔 유효한 스냅숏이다.
func a112ShadowLane() Snapshot {
	snapshot := a112ObservedSnapshot()
	outcome := LaneShadowWouldEmit
	snapshot.Lanes[0].Runtime, snapshot.Lanes[0].ShadowOutcome = LaneRuntimeShadow, &outcome
	return snapshot
}

func TestValidateAcceptsAnObservedOffLaneInShadow(t *testing.T) {
	if err := Validate(a112ShadowLane()); err != nil {
		t.Fatalf("an observed OFF/OFF lane in SHADOW must validate: %v", err)
	}
	for _, outcome := range LaneShadowOutcomes() {
		snapshot := a112ShadowLane()
		value := LaneShadowOutcome(outcome)
		snapshot.Lanes[0].ShadowOutcome = &value
		if err := Validate(snapshot); err != nil {
			t.Errorf("shadow outcome %s must validate: %v", outcome, err)
		}
	}
}

// 거부 표 — 각 줄이 무엇을 거부하는지 이름으로 적는다.
func TestValidateRefusesShadowOutsideTheOneCrossRule(t *testing.T) {
	on, latched := StateOn, LaneLatched
	bogus := LaneShadowOutcome("MAYBE")
	for _, tc := range []struct {
		name   string
		mutate func(*LaneRuntimeProjection)
	}{
		{"SHADOW with desired ON", func(lane *LaneRuntimeProjection) { lane.Desired = on }},
		{"SHADOW with effective ON", func(lane *LaneRuntimeProjection) { lane.Desired, lane.Effective = on, on }},
		{"SHADOW on an unobserved lane (health null)", func(lane *LaneRuntimeProjection) { *lane = unobservedShadow(lane) }},
		{"SHADOW before any wave (cycleGeneration 0)", func(lane *LaneRuntimeProjection) {
			lane.CycleGeneration, lane.Trigger, lane.Start, lane.Outcome = 0, nil, nil, nil
		}},
		{"SHADOW without an outcome", func(lane *LaneRuntimeProjection) { lane.ShadowOutcome = nil }},
		{"an outcome without SHADOW", func(lane *LaneRuntimeProjection) { lane.Runtime = LaneRuntimeUnobserved }},
		{"an unknown shadow outcome", func(lane *LaneRuntimeProjection) { lane.ShadowOutcome = &bogus }},
		{"an unknown runtime", func(lane *LaneRuntimeProjection) { lane.Runtime = "ACTIVE" }},
	} {
		snapshot := a112ShadowLane()
		tc.mutate(&snapshot.Lanes[0])
		if err := Validate(snapshot); err == nil {
			t.Errorf("%s: Validate accepted it", tc.name)
		}
	}
	// 거부하지 않는 쪽(정상 입력): 잠긴 레인도 SHADOW 를 관측할 수 있다(관측만 — 브리프 §6).
	snapshot := a112ShadowLane()
	snapshot.Lanes[0].Health = &latched
	if err := Validate(snapshot); err != nil {
		t.Fatalf("a latched observed OFF lane in SHADOW must validate: %v", err)
	}
}

func unobservedShadow(lane *LaneRuntimeProjection) LaneRuntimeProjection {
	return LaneRuntimeProjection{Market: lane.Market, Family: lane.Family, LaneID: lane.LaneID, LaneVersion: lane.LaneVersion,
		Horizon: lane.Horizon, Desired: StateOff, Effective: StateOff, Runtime: LaneRuntimeShadow, ShadowOutcome: lane.ShadowOutcome}
}

// 어휘: runtime 은 정확히 {UNOBSERVED, SHADOW}, shadowOutcome 은 정확히 셋 — OpenAPI enum 동기 시험(httpapi)이 이 두 함수와 대조한다.
// wire: 핀 0(오늘) = 레인 전부 runtime UNOBSERVED · shadowOutcome null 이고, shadowOutcome 은 언제나 직렬화된다(required).
func TestTheShadowVocabularyAndTheWireShape(t *testing.T) {
	if got, want := LaneRuntimes(), []string{"UNOBSERVED", "SHADOW"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("LaneRuntimes=%v, want %v", got, want)
	}
	if got, want := LaneShadowOutcomes(), []string{"WOULD_EMIT", "NOT_THIS_LANE", "NO_INPUT"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("LaneShadowOutcomes=%v, want %v", got, want)
	}
	found := false
	for _, name := range LaneJSONFields() {
		found = found || name == "shadowOutcome"
	}
	if !found {
		t.Fatal("shadowOutcome is not a lane JSON field")
	}
	for _, lane := range DormantSnapshot(a112At).Lanes {
		if lane.Runtime != LaneRuntimeUnobserved || lane.ShadowOutcome != nil {
			t.Fatalf("a dormant lane carries runtime=%s shadowOutcome=%v, want UNOBSERVED/null", lane.Runtime, lane.ShadowOutcome)
		}
	}
	original := a112ShadowLane().Lanes
	cloned := cloneLanes(original)
	if cloned[0].ShadowOutcome == original[0].ShadowOutcome || *cloned[0].ShadowOutcome != *original[0].ShadowOutcome {
		t.Fatal("cloneLanes must deep-copy shadowOutcome")
	}
}
