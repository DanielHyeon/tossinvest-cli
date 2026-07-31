package markout

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestMeasureUsesFirstExistingObservationWithinTargetTolerance(t *testing.T) {
	decision := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)
	report := Measure(decision, "100", []Observation{
		{At: decision.Add(31*time.Minute + time.Second), Price: "131"},
		{At: decision.Add(5*time.Minute + 40*time.Second), Price: "106"},
		{At: decision.Add(4*time.Minute + 59*time.Second), Price: "104"},
		{At: decision.Add(5*time.Minute + 20*time.Second), Price: "105"},
		{At: decision.Add(16 * time.Minute), Price: "115"},
	})

	five, ok := report.ForMinutes(5)
	if !ok || five.Status != StatusMeasured || five.ObservedAt == nil ||
		!five.ObservedAt.Equal(decision.Add(5*time.Minute+20*time.Second)) {
		t.Fatalf("5m = %+v, want the first observation at/after target", five)
	}
	if five.ReturnPct != "5" {
		t.Errorf("5m return = %q, want 5", five.ReturnPct)
	}
	fifteen, ok := report.ForMinutes(15)
	if !ok || fifteen.Status != StatusMeasured || fifteen.ObservedAt == nil ||
		!fifteen.ObservedAt.Equal(decision.Add(16*time.Minute)) {
		t.Fatalf("15m boundary = %+v, want +60s observation measured", fifteen)
	}
	thirty, ok := report.ForMinutes(30)
	if !ok || thirty.Status != StatusNotMeasured || thirty.ReturnPct != "" {
		t.Fatalf("30m = %+v, want not_measured with no synthetic zero return", thirty)
	}
}

func TestMeasureRefusesMissingOrUnreadableExistingPrices(t *testing.T) {
	decision := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name       string
		base       string
		observed   string
		wantReason Reason
	}{
		{name: "missing decision price", observed: "101", wantReason: ReasonDecisionPriceAbsent},
		{name: "unreadable decision price", base: "NaN", observed: "101", wantReason: ReasonDecisionPriceUnreadable},
		{name: "unreadable first observation", base: "100", observed: "bad", wantReason: ReasonObservationPriceUnreadable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := Measure(decision, tc.base, []Observation{{
				At: decision.Add(5 * time.Minute), Price: tc.observed,
			}}).ForMinutes(5)
			if got.Status != StatusNotMeasured || got.Reason != tc.wantReason || got.ReturnPct != "" {
				t.Fatalf("measurement = %+v, want not_measured/%s and no numeric return", got, tc.wantReason)
			}
		})
	}
}

func TestA049GoldenFixturePinsFiveFifteenThirtyMinuteContract(t *testing.T) {
	data, err := os.ReadFile("testdata/a049_markout.golden.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		DecisionAt    time.Time     `json:"decision_at"`
		DecisionPrice string        `json:"decision_price"`
		Observations  []Observation `json:"observations"`
		Expected      Report        `json:"expected"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatalf("decode golden: %v", err)
	}
	got := Measure(fixture.DecisionAt, fixture.DecisionPrice, fixture.Observations)
	if encodedGot, encodedWant := mustJSON(t, got), mustJSON(t, fixture.Expected); encodedGot != encodedWant {
		t.Fatalf("golden mismatch\n got: %s\nwant: %s", encodedGot, encodedWant)
	}
}

func mustJSON(t *testing.T, value any) string {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
