package strategymarket

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

func minutes() []official.RawMinuteCandle {
	out := make([]official.RawMinuteCandle, 5)
	for i := range out {
		out[i] = official.RawMinuteCandle{Timestamp: time.Date(2026, 7, 31, 9, i, 0, 0, time.FixedZone("KST", 9*3600)).Format(time.RFC3339), Open: "100.01", High: "101.10", Low: "99.90", Close: "100.20", Volume: "0.1", Currency: "KRW"}
	}
	out[4].Close = "100.55"
	return out
}
func TestAggregateClosedKRXFiveMinutePreservesExactDecimals(t *testing.T) {
	got, err := AggregateClosedKRXFiveMinute(minutes(), time.Date(2026, 7, 31, 9, 5, 0, 0, time.FixedZone("KST", 9*3600)))
	if err != nil {
		t.Fatal(err)
	}
	if got.Open != "100.01" || got.High != "101.1" || got.Low != "99.9" || got.Close != "100.55" || got.Volume != "0.5" {
		t.Fatalf("bar=%+v", got)
	}
}
func TestAggregateClosedKRXFiveMinuteFailsClosed(t *testing.T) {
	base := minutes()
	now := time.Date(2026, 7, 31, 9, 5, 0, 0, time.FixedZone("KST", 9*3600))
	tests := []struct {
		name string
		rows []official.RawMinuteCandle
		now  time.Time
		kind Refusal
	}{
		{"missing", base[:4], now, RefusalIncompleteBucket}, {"open", base, now.Add(-time.Nanosecond), RefusalOpenBucket},
		{"naive", func() []official.RawMinuteCandle { v := minutes(); v[0].Timestamp = "2026-07-31T09:00:00"; return v }(), now, RefusalNaiveTimestamp},
		{"gap", func() []official.RawMinuteCandle {
			v := minutes()
			v[2].Timestamp = time.Date(2026, 7, 31, 9, 7, 0, 0, time.FixedZone("KST", 9*3600)).Format(time.RFC3339)
			return v
		}(), now.Add(3 * time.Minute), RefusalMinuteGap},
		{"outside", func() []official.RawMinuteCandle {
			v := minutes()
			for i := range v {
				v[i].Timestamp = time.Date(2026, 7, 31, 8, 55+i, 0, 0, time.FixedZone("KST", 9*3600)).Format(time.RFC3339)
			}
			return v
		}(), now, RefusalOutsideRegularSession},
		{"decimal", func() []official.RawMinuteCandle { v := minutes(); v[0].Open = "1e2"; return v }(), now, RefusalInvalidDecimal},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := AggregateClosedKRXFiveMinute(tc.rows, tc.now)
			var ie *IntegrityError
			if !errors.As(err, &ie) || ie.Kind != tc.kind {
				t.Fatalf("err=%v want=%s", err, tc.kind)
			}
		})
	}
}

type stateStub struct {
	reading StateReading
	err     error
}

func (s stateStub) ReadSymbolState(context.Context, string, string) (StateReading, error) {
	return s.reading, s.err
}
func TestAuthoritativeSymbolStateMustBePresentFreshAndNormal(t *testing.T) {
	now := time.Now().UTC()
	if _, err := RequireFreshNormalState(context.Background(), stateStub{reading: StateReading{State: StateNormal, ObservedAt: now, Authority: "official-status-v1"}}, "KR", "005930", now); err != nil {
		t.Fatal(err)
	}
	for _, source := range []SymbolStateSource{nil, stateStub{err: context.Canceled}, stateStub{reading: StateReading{State: StateNormal, ObservedAt: now.Add(-31 * time.Second), Authority: "official-status-v1"}}, stateStub{reading: StateReading{State: StateHalt, ObservedAt: now, Authority: "official-status-v1"}}} {
		if _, err := RequireFreshNormalState(context.Background(), source, "KR", "005930", now); err == nil {
			t.Fatal("blocked state accepted")
		}
	}
}
