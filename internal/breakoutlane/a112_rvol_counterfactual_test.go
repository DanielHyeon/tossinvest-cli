package breakoutlane

// a112 태스크 2.3 — 1.2 반사실(Manager 판정 2026-10-01 (b), design.md 「1.2 반사실의 의미」): 입장 문턱이 1.2 였다면 이 setup 에 돌파 봉이 있었는가.
// 입장하지 못한 봉 중 close buffer 와 wick 상한을 통과하고 RVOL 만 1.5 미달 · 1.2 이상인 봉이 있으면 RVOLAt1200000 이 참이다.
// 기록 전용: 그 봉이 결정(단계 · 전이 · 거절 · 수량 · 제안 · 봉인)을 바꾸면 안 된다 — 각 경우를 RVOL 1.0 의 같은 스냅숏과 비교한다.

import (
	"reflect"
	"testing"
)

func a112WithBreakoutBar(t *testing.T, set func(*ClosedBarInput)) Decision {
	t.Helper()
	i := fixtureInput(t)
	i.Bars = append([]ClosedBar(nil), i.Bars...)
	b := i.Bars[15].value
	set(&b)
	i.Bars[15] = ClosedBar{value: b}
	return Evaluate(snapshot(t, i), nil)
}

func TestTheOnePointTwoCounterfactualRecordsABarThatOnlyTheLowerThresholdWouldAdmit(t *testing.T) {
	for _, tc := range []struct {
		name string
		set  func(*ClosedBarInput)
		want bool
	}{
		{"RVOL exactly 1.2, close and wick qualify", func(b *ClosedBarInput) { b.RVOLPPM = 1_200_000 }, true},
		{"RVOL one ppm under admission", func(b *ClosedBarInput) { b.RVOLPPM = 1_499_999 }, true},
		{"RVOL one ppm under 1.2", func(b *ClosedBarInput) { b.RVOLPPM = 1_199_999 }, false},
		{"wick one ppm over the veto", func(b *ClosedBarInput) { b.RVOLPPM, b.UpperWickRangePPM = 1_300_000, 350_001 }, false},
		{"close inside the buffer (no breakout close)", func(b *ClosedBarInput) { b.RVOLPPM, b.CloseMinor = 1_300_000, 100 }, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := a112WithBreakoutBar(t, tc.set)
			p := d.Provenance()
			if p.RVOLAt1200000 != tc.want || p.RVOLAdmission || p.RVOLAt2000000 || p.RVOLAt2500000 {
				t.Fatalf("provenance=%+v, want at1.2=%v and no admission or higher counterfactual", p, tc.want)
			}
			// 기록 전용: RVOL 1.0(반사실 없음)인 같은 모양과 결정이 같다 — 깃발만 다르다.
			base := a112WithBreakoutBar(t, func(b *ClosedBarInput) {
				tc.set(b)
				b.RVOLPPM = 1_000_000
			})
			if base.Provenance().RVOLAt1200000 {
				t.Fatalf("arrangement: the RVOL 1.0 twin recorded a counterfactual")
			}
			if d.Phase() != base.Phase() || d.Refusal() != base.Refusal() || d.ProposalID() != base.ProposalID() || d.FinalQuantity() != base.FinalQuantity() ||
				d.CandidateQuantity() != base.CandidateQuantity() || !reflect.DeepEqual(p.Transitions, base.Provenance().Transitions) || d.seal != decisionSeal(d) {
				t.Fatalf("the counterfactual changed the decision: %+v vs twin %+v", d, base)
			}
		})
	}
}

// 입장 경로는 그대로다: 1.5 이상 입장 봉의 기록(1.2 · 2.0 · 2.5)과 제안은 이 편집 전과 같은 식으로 나온다.
func TestTheAdmittedBreakoutPathIsUnchangedByTheCounterfactual(t *testing.T) {
	d := a112WithBreakoutBar(t, func(b *ClosedBarInput) { b.RVOLPPM = 1_500_000 })
	p := d.Provenance()
	if d.Phase() != "PROPOSED" || !p.RVOLAdmission || !p.RVOLAt1200000 || p.RVOLAt2000000 || p.RVOLAt2500000 || len(p.Transitions) != 7 {
		t.Fatalf("admitted path: %s %+v", d.Phase(), p)
	}
	// 1.2 만 통과하는 봉 뒤에 입장 봉이 오면: 입장 봉이 돌파이고 반사실은 참(입장 = 1.2 에서도 입장).
	i := fixtureInput(t)
	i.Bars = append([]ClosedBar(nil), i.Bars[:15]...)
	i.Bars = append(i.Bars,
		fixtureBar(t, 16, 102, 99, 101, 1_300_000, 100_000), // 1.2 만 통과
		fixtureBar(t, 17, 111, 99, 101, 1_500_000, 350_000), // 입장
		fixtureBar(t, 18, 100, 98, 99, 1_000_000, 100_000),  // retest
		fixtureBar(t, 19, 102, 99, 101, 1_000_000, 100_000), // reclaim
	)
	later := Evaluate(snapshot(t, i), nil)
	if lp := later.Provenance(); later.Phase() != "PROPOSED" || !lp.RVOLAdmission || !lp.RVOLAt1200000 {
		t.Fatalf("a 1.2-only bar before the admitted bar: %s %+v", later.Phase(), lp)
	}
}
