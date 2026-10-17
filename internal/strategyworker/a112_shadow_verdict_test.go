//go:build tossos_testseams

package strategyworker

// a112 7.3.1 SHADOW(브리프 v3.3 §4 · §6 · §7) — worker 의 반사실 판정과 「OFF 레인」 술어 하나.
//
// SHADOW 는 관측만 한다: 판정은 봉투도 승격도 만들지 않는 값(ShadowOutcome)이고, 입력은 opaque ShadowInput 이라 admit · Submit 이
// 받는 strategyworker.Input 과 섞일 수 없다. 술어는 하나다 — ShadowEligible = Desired OFF ∧ Effective OFF.

import (
	"reflect"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reversallane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

// 판정 세 갈래: 이 레인이 소유하면 WOULD_EMIT, 이 레인을 주장하지만 봉인이 성립하지 않으면 NOT_THIS_LANE, 이 레인을 주장하지도 않으면
// NO_INPUT. 판정은 활성화를 보지 않는다(OFF 레인의 반사실이므로 — 같은 입력이 ON 이었다면 Run 이 봉투를 냈을 것인가).
func TestTheShadowVerdictAsksWhetherThisLaneWouldHaveEmitted(t *testing.T) {
	f := newFixture(allKRFamilies()...)
	worker := workerFor(t, strategyrouter.MarketKR, strategyrouter.FamilyContinuation)
	own := f.input(t, continuationlane.KRContinuationLaneID)
	// 대조: 같은 입력이 켜진 worker 에서 실제로 봉투가 된다 — WOULD_EMIT 이 「Run 이 냈을 것」 과 같은 뜻임을 잰다.
	if cycle := worker.Run(promoting(t, strategyrouter.MarketKR, strategyrouter.FamilyContinuation), own); cycle.Outcome != OutcomeEmitted {
		t.Fatalf("control: the owned input is not emittable (%s/%s)", cycle.Outcome, cycle.Detail)
	}
	if got := worker.ShadowVerdict(NewShadowInput(own.Proposal)); got != ShadowWouldEmit {
		t.Fatalf("owned input verdict=%s, want %s", got, ShadowWouldEmit)
	}
	tampered := own
	tampered.Proposal.Result.Quantity++
	if got := worker.ShadowVerdict(NewShadowInput(tampered.Proposal)); got != ShadowNotThisLane {
		t.Fatalf("a broken seal claiming this lane: verdict=%s, want %s", got, ShadowNotThisLane)
	}
	other := f.input(t, reversallane.KRReversalLaneID)
	if got := worker.ShadowVerdict(NewShadowInput(other.Proposal)); got != ShadowNoInput {
		t.Fatalf("another lane's proposal: verdict=%s, want %s", got, ShadowNoInput)
	}
	// 활성화를 보지 않는다: 잠든 worker(오늘 생산)도 같은 판정이다 — Run 은 DORMANT 지만 반사실은 WOULD_EMIT.
	if cycle := worker.Run(noActivation(), own); cycle.Outcome != OutcomeDormant {
		t.Fatalf("control: an unactivated worker must stay DORMANT, got %s", cycle.Outcome)
	}
}

// 레인 묶음 판정: 한 레인이 여러 종목의 입력을 받으면 WOULD_EMIT > NOT_THIS_LANE > NO_INPUT 순으로 하나를 고른다. 빈 묶음은 NO_INPUT.
// 잠긴 레인도 판정한다(SHADOW 는 관측만 — 브리프 §6 「LATCHED 허용, health 그대로」).
func TestALaneFoldsItsShadowVerdictsAndALatchedLaneStillObserves(t *testing.T) {
	f := newFixture(allKRFamilies()...)
	own := f.input(t, continuationlane.KRContinuationLaneID)
	tampered := f.withSymbol("000660").input(t, continuationlane.KRContinuationLaneID)
	tampered.Proposal.Result.Quantity++
	other := f.input(t, reversallane.KRReversalLaneID)
	lane := laneFor(t, strategyrouter.MarketKR, strategyrouter.FamilyContinuation)
	for _, tc := range []struct {
		name   string
		inputs []ShadowInput
		want   ShadowOutcome
	}{
		{"empty", nil, ShadowNoInput},
		{"other lane only", []ShadowInput{NewShadowInput(other.Proposal)}, ShadowNoInput},
		{"broken seal and other", []ShadowInput{NewShadowInput(other.Proposal), NewShadowInput(tampered.Proposal)}, ShadowNotThisLane},
		{"owned beats broken", []ShadowInput{NewShadowInput(tampered.Proposal), NewShadowInput(own.Proposal)}, ShadowWouldEmit},
		{"owned beats a later broken one", []ShadowInput{NewShadowInput(own.Proposal), NewShadowInput(tampered.Proposal)}, ShadowWouldEmit},
	} {
		if got := lane.ShadowOutcomeOver(tc.inputs); got != tc.want {
			t.Errorf("%s: outcome=%s, want %s", tc.name, got, tc.want)
		}
	}
	for range lane.Policy().FailureThreshold() {
		lane.Fail("induced", false)
	}
	if !lane.Latched() {
		t.Fatal("arrangement: the lane did not latch")
	}
	if got := lane.ShadowOutcomeOver([]ShadowInput{NewShadowInput(own.Proposal)}); got != ShadowWouldEmit {
		t.Fatalf("a latched lane's shadow verdict=%s, want %s (observation only)", got, ShadowWouldEmit)
	}
}

// 술어 하나: ShadowEligible = Desired OFF ∧ Effective OFF. 레인 위임은 같은 함수다.
func TestShadowEligibilityIsExactlyOffAndOff(t *testing.T) {
	worker := workerFor(t, strategyrouter.MarketKR, strategyrouter.FamilyContinuation)
	lane := laneFor(t, strategyrouter.MarketKR, strategyrouter.FamilyContinuation)
	for _, tc := range []struct {
		name       string
		activation strategyrouter.FamilyActivation
		want       bool
	}{
		{"no activation (today)", noActivation(), true},
		{"this family ON", promoting(t, strategyrouter.MarketKR, strategyrouter.FamilyContinuation), false},
		{"a neighbour ON, this OFF", promoting(t, strategyrouter.MarketKR, strategyrouter.FamilyReversal), true},
		{"the other market's activation", promoting(t, strategyrouter.MarketUS, strategyrouter.FamilyContinuation), true},
		{"desired ON but effective OFF (not yet up)", strategyrouter.FamilyActivationDesiredOnlyForTest(strategyrouter.MarketKR, 1,
			map[string]bool{worker.Key().LaneID: true}), false},
	} {
		if got := ShadowEligible(tc.activation, worker); got != tc.want {
			t.Errorf("%s: ShadowEligible=%v, want %v", tc.name, got, tc.want)
		}
		if got := lane.ShadowEligible(tc.activation); got != tc.want {
			t.Errorf("%s: lane.ShadowEligible=%v, want %v (the lane must delegate to the one predicate)", tc.name, got, tc.want)
		}
	}
}

// ShadowInput 은 opaque 다(필드 전부 비공개) — 이 패키지 밖에서는 NewShadowInput 으로만 만들고, Input 이 아니므로 Run 에 들어가지 않는다.
// 어휘는 정확히 셋.
func TestTheShadowInputIsOpaqueAndTheVocabularyIsExactlyThree(t *testing.T) {
	typ := reflect.TypeOf(ShadowInput{})
	for index := range typ.NumField() {
		if typ.Field(index).IsExported() {
			t.Errorf("ShadowInput exports field %s — it must stay opaque", typ.Field(index).Name)
		}
	}
	if reflect.TypeOf(ShadowInput{}) == reflect.TypeOf(Input{}) {
		t.Fatal("ShadowInput must not be the cycle Input")
	}
	got := ShadowOutcomes()
	want := []ShadowOutcome{ShadowWouldEmit, ShadowNotThisLane, ShadowNoInput}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ShadowOutcomes=%v, want %v", got, want)
	}
}

// laneFor 는 생산 레인 하나(열린 채로 태어남)다.
func laneFor(t *testing.T, market strategyrouter.Market, family strategyrouter.Family) *Lane {
	t.Helper()
	for _, lane := range ProductionLanes(clock.NewFake(newFixture().now)) {
		if lane.Key().Market == market && lane.Key().Family == family {
			return lane
		}
	}
	t.Fatalf("no production lane for %s/%s", market, family)
	return nil
}
