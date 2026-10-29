package engine

import (
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

func TestA087ExitLimitPriceUsesOnlyEligibleKoreanProtectiveLowerLimit(t *testing.T) {
	kst, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 3, 30, 9, 0, 0, 0, kst)
	m := managed{position: journal.Position{Market: "kr", Symbol: "005930"}, state: journal.ExitState{Baseline: "50000"}}

	tests := []struct {
		name     string
		now      time.Time
		market   string
		action   exitpolicy.Action
		observed string
		want     exitLimitChoice
	}{
		{"open boundary", now, "kr", exitpolicy.ActionBaselineBreach, "50400", exitLimitChoice{price: "50400", source: exitPriceSourceLowerLimit}},
		{"before open", now.Add(-time.Second), "kr", exitpolicy.ActionBaselineBreach, "50400", exitLimitChoice{price: "50400", source: exitPriceSourceObserved}},
		{"last second", now.Add(6*time.Hour + 29*time.Minute + 59*time.Second), "kr", exitpolicy.ActionLadderStop, "50400", exitLimitChoice{price: "50400", source: exitPriceSourceLowerLimit}},
		{"close boundary", now.Add(6*time.Hour + 30*time.Minute), "kr", exitpolicy.ActionLadderStop, "50400", exitLimitChoice{price: "50400", source: exitPriceSourceObserved}},
		{"weekend", time.Date(2026, 3, 29, 10, 0, 0, 0, kst), "kr", exitpolicy.ActionBaselineBreach, "50400", exitLimitChoice{price: "50400", source: exitPriceSourceObserved}},
		{"take profit", now, "kr", exitpolicy.ActionLadderTakeProfit, "50400", exitLimitChoice{price: "50400", source: exitPriceSourceObserved}},
		{"us", now, "us", exitpolicy.ActionBaselineBreach, "50400", exitLimitChoice{price: "50400", source: exitPriceSourceObserved}},
		{"observed higher", now, "kr", exitpolicy.ActionBaselineBreach, "50401", exitLimitChoice{price: "50400", source: exitPriceSourceLowerLimit}},
		{"observed lower", now, "kr", exitpolicy.ActionBaselineBreach, "50399", exitLimitChoice{price: "50399", source: exitPriceSourceObserved}},
		{"observed invalid", now, "kr", exitpolicy.ActionBaselineBreach, "not-a-number", exitLimitChoice{price: "not-a-number", source: exitPriceSourceObserved}},
		{"observed empty", now, "kr", exitpolicy.ActionBaselineBreach, "", exitLimitChoice{price: "50000", source: exitPriceSourceBaseline}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tableBoard := &exitLowerLimitBoard{}
			tableBoard.publish(lowerLimitSnapshot{records: map[string]lowerLimitRecord{
				lowerLimitKey(tc.market, m.position.Symbol): {price: "50400", timestamp: now},
			}})
			got := exitLimitPrice(tc.now, managed{position: journal.Position{Market: tc.market, Symbol: m.position.Symbol}, state: m.state},
				exitpolicy.Proposal{Action: tc.action}, tc.observed, tableBoard)
			if got != tc.want {
				t.Fatalf("choice = %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestA087ExitLimitPriceRejectsNonCanonicalAndStaleLowerLimit(t *testing.T) {
	kst, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 3, 30, 10, 0, 0, 0, kst)
	m := managed{position: journal.Position{Market: "kr", Symbol: "005930"}, state: journal.ExitState{Baseline: "50000"}}
	for _, price := range []string{"050400", "50400.0", " 50400", "5.04e4", "", "0"} {
		t.Run(price, func(t *testing.T) {
			board := &exitLowerLimitBoard{}
			board.publish(lowerLimitSnapshot{records: map[string]lowerLimitRecord{
				lowerLimitKey("kr", "005930"): {price: price, timestamp: now},
			}})
			got := exitLimitPrice(now, m, exitpolicy.Proposal{Action: exitpolicy.ActionBaselineBreach}, "50400", board)
			want := exitLimitChoice{price: "50400", source: exitPriceSourceObserved}
			if got != want {
				t.Fatalf("%q choice = %#v, want %#v", price, got, want)
			}
		})
	}
	board := &exitLowerLimitBoard{}
	board.publish(lowerLimitSnapshot{records: map[string]lowerLimitRecord{
		lowerLimitKey("kr", "005930"): {price: "50400", timestamp: now.Add(-24 * time.Hour)},
	}})
	if got := exitLimitPrice(now, m, exitpolicy.Proposal{Action: exitpolicy.ActionBaselineBreach}, "50400", board); got.source != exitPriceSourceObserved {
		t.Fatalf("stale source = %q, want observed", got.source)
	}
}

func TestA087FirstProtectiveAttemptIsMonotonic(t *testing.T) {
	kst, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 3, 30, 10, 0, 0, 0, kst)
	m := managed{position: journal.Position{Market: "kr", Symbol: "005930"}, state: journal.ExitState{Baseline: "50000"}}
	board := &exitLowerLimitBoard{}
	board.publish(lowerLimitSnapshot{records: map[string]lowerLimitRecord{
		lowerLimitKey("kr", "005930"): {price: "50400", timestamp: now},
	}})
	proposal := exitpolicy.Proposal{Action: exitpolicy.ActionBaselineBreach}
	if got := exitLimitPrice(now, m, proposal, "50400", board); got.source != exitPriceSourceLowerLimit {
		t.Fatalf("first source = %q, want lower limit", got.source)
	}
	board.markProtectiveAttempt("kr", "005930", now)
	if got := exitLimitPrice(now, m, proposal, "50400", board); got.source != exitPriceSourceObserved {
		t.Fatalf("second source = %q, want observed", got.source)
	}
}

func TestA087RawLowerLimitValidationFailsClosed(t *testing.T) {
	kst, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 3, 30, 10, 0, 0, 0, kst)
	valid := func(price, timestamp string) official.ProtectivePriceLimits {
		return official.ProtectivePriceLimits{LowerLimitPrice: &price, Timestamp: &timestamp}
	}
	for _, tc := range []struct {
		name string
		raw  official.ProtectivePriceLimits
	}{
		{"null price", official.ProtectivePriceLimits{Timestamp: a087String(now.Format(time.RFC3339))}},
		{"zero", valid("0", now.Format(time.RFC3339))},
		{"fraction", valid("50400.0", now.Format(time.RFC3339))},
		{"timestamp missing", official.ProtectivePriceLimits{LowerLimitPrice: a087String("50400")}},
		{"timestamp malformed", valid("50400", "not-a-time")},
		{"previous day payload", valid("50400", now.AddDate(0, 0, -1).Format(time.RFC3339))},
		{"before nine", valid("50400", time.Date(2026, 3, 30, 8, 59, 59, 0, kst).Format(time.RFC3339))},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, ok := validLowerLimitRecord(now, tc.raw); ok {
				t.Fatalf("%s was accepted", tc.name)
			}
		})
	}
}

func TestA087NilBoardKeepsLiteralFallback(t *testing.T) {
	m := managed{position: journal.Position{Market: "kr", Symbol: "005930"}, state: journal.ExitState{Baseline: "50000"}}
	now := time.Date(2026, 3, 30, 1, 0, 0, 0, time.UTC)
	got := exitLimitPrice(now, m, exitpolicy.Proposal{Action: exitpolicy.ActionBaselineBreach}, "50400", nil)
	if got != (exitLimitChoice{price: "50400", source: exitPriceSourceObserved}) {
		t.Fatalf("choice = %#v", got)
	}
}

func TestA087EmptyObservedAndBaselineKeepsEmptyLegacyFallback(t *testing.T) {
	m := managed{position: journal.Position{Market: "kr", Symbol: "005930"}}
	now := time.Date(2026, 3, 30, 1, 0, 0, 0, time.UTC)
	got := exitLimitPrice(now, m, exitpolicy.Proposal{Action: exitpolicy.ActionBaselineBreach}, "", nil)
	if got != (exitLimitChoice{source: exitPriceSourceBaseline}) {
		t.Fatalf("choice = %#v, want empty baseline fallback", got)
	}
}

func TestA087ExitOrderLineKeepsChosenPriceAndSource(t *testing.T) {
	m := managed{position: journal.Position{Market: "kr", Symbol: "005930", ID: "p-a087"}, state: journal.ExitState{Baseline: "50000"}}
	proposal := exitpolicy.Proposal{Action: exitpolicy.ActionBaselineBreach}
	for _, choice := range []exitLimitChoice{
		{price: "50400", source: exitPriceSourceLowerLimit},
		{price: "50399", source: exitPriceSourceObserved},
		{price: "50000", source: exitPriceSourceBaseline},
	} {
		line := exitOrderLineFor(m, proposal, "10", "intent-a087", "50399", choice, false)
		if line.price != choice.price || line.priceSource != choice.source {
			t.Fatalf("line = %#v, want choice %#v", line, choice)
		}
		args := line.args()
		got := map[string]any{}
		for i := 0; i < len(args); i += 2 {
			got[args[i].(string)] = args[i+1]
		}
		if got["price"] != choice.price || got["price_source"] != string(choice.source) {
			t.Fatalf("args = %#v, want price/source from %#v", got, choice)
		}
	}
}
