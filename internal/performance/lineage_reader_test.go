package performance

import (
	"context"
	"testing"
	"time"
)

type fakeJournalLineageReader struct {
	trades []Trade
	calls  int
}

func (f *fakeJournalLineageReader) ClosedStrategyTrades(context.Context, ClosedTradeWindow) ([]Trade, error) {
	f.calls++
	return append([]Trade(nil), f.trades...), nil
}

func TestJournalHandoffConsumesOneExactLineageReadAndCallerOwnedObservations(t *testing.T) {
	store := openTestStore(t)
	at := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	trade := measuredTrade(at)
	reader := &fakeJournalLineageReader{trades: []Trade{trade}}
	existing := map[string][]Observation{"position-1": {{
		ID: "existing-5", PositionID: "position-1", At: at.Add(5 * time.Minute),
		Price: "105", Source: "existing-position", SourceVersion: "v1",
	}}}
	got, err := store.CollectClosedStrategyTrades(context.Background(), reader, ClosedTradeWindow{
		ClosedAfter: at, ClosedAtOrBefore: at.Add(time.Hour),
	}, existing, at.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if reader.calls != 1 || len(got) != 1 || got[0].Markout(5).ObservationID != "existing-5" {
		t.Fatalf("calls=%d snapshots=%+v", reader.calls, got)
	}
}
