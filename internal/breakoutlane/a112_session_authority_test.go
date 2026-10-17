package breakoutlane

// a112 3.8(-c) 감사 보강(Manager 판정 2026-10-04): 세션 사실은 **공식 봉** 이 정한다 — 후보(Toss rank · volume · flow) 쪽의 「정규장」 주장은
// 레인 입력에 들어올 자리 자체가 없고, 봉 둘레의 두 문이 세션을 지킨다:
//   ① NewClosedBar 는 정규장 · 마감 봉만 받는다(types.go NewClosedBar 의 `!input.RegularSession || !input.Closed`).
//   ② 스냅숏은 모든 봉의 SessionID 가 입력 세션과 같아야 선다(machine.go validStructuralEvidenceInput 의 `b.SessionID != v.SessionID`).
// 둘 다 감사 시점 행동 시험이 없던 갈래다(audit-3.8-4.5.md 3.8-c). KR · US 두 시장으로 잰다.

import "testing"

func a112MarketInput(t *testing.T, market Market) EvidenceInput {
	t.Helper()
	input := fixtureInput(t)
	if market == MarketKR {
		return input
	}
	// US: 레인 · 세션 이름만 바꾼다(봉 모양은 같다) — 봉의 SessionID 도 입력 세션에 맞춘다.
	input.Market, input.LaneID, input.SessionID = MarketUS, USLaneID, "US:2026-08-18"
	bars := make([]ClosedBar, len(input.Bars))
	for index, bar := range input.Bars {
		value := bar.value
		value.SessionID = input.SessionID
		bars[index] = ClosedBar{value: value}
	}
	input.Bars = bars
	return input
}

func TestAClosedBarOutsideTheRegularSessionOrStillOpenIsRefused(t *testing.T) {
	base := fixtureInput(t).Bars[15].value
	if _, err := NewClosedBar(base); err != nil {
		t.Fatalf("arrangement: the fixture bar itself is refused: %v", err)
	}
	for name, change := range map[string]func(*ClosedBarInput){
		"extended session": func(b *ClosedBarInput) { b.RegularSession = false },
		"not yet closed":   func(b *ClosedBarInput) { b.Closed = false },
	} {
		t.Run(name, func(t *testing.T) {
			bar := base
			change(&bar)
			if _, err := NewClosedBar(bar); err == nil {
				t.Fatal("NewClosedBar accepted the bar")
			}
		})
	}
}

func TestASnapshotRefusesABarFromAnotherSession(t *testing.T) {
	for _, market := range []Market{MarketKR, MarketUS} {
		t.Run(string(market), func(t *testing.T) {
			input := a112MarketInput(t, market)
			if _, err := NewEvidenceSnapshot(input); err != nil {
				t.Fatalf("arrangement: the %s fixture snapshot is refused: %v", market, err)
			}
			// 돌파 봉(색인 15) 하나만 다른 세션 이름 — 봉 자체는 유효하다(NewClosedBar 통과), 세션 대조만 어긋난다.
			bars := append([]ClosedBar(nil), input.Bars...)
			value := bars[15].value
			value.SessionID = input.SessionID + "-other"
			if _, err := NewClosedBar(value); err != nil {
				t.Fatalf("arrangement: the relabelled bar is itself invalid: %v", err)
			}
			bars[15] = ClosedBar{value: value}
			input.Bars = bars
			if _, err := NewEvidenceSnapshot(input); err == nil {
				t.Fatal("NewEvidenceSnapshot accepted a bar whose session differs from the snapshot's")
			}
		})
	}
}
