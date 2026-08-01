package performance

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeJournalLineageReader struct {
	trades []Trade
	calls  int
	err    error
}

func (f *fakeJournalLineageReader) ClosedStrategyTrades(context.Context, ClosedTradeWindow) ([]Trade, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
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
		AccountRef:  "acct-1",
		ClosedAfter: at, ClosedAtOrBefore: at.Add(time.Hour),
	}, existing, at.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if reader.calls != 1 || len(got) != 1 || got[0].Markout(5).ObservationID != "existing-5" {
		t.Fatalf("calls=%d snapshots=%+v", reader.calls, got)
	}
}

func TestJournalHandoffValidatesAccountAndWindowBeforeReading(t *testing.T) {
	store := openTestStore(t)
	at := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	for _, window := range []ClosedTradeWindow{
		{ClosedAfter: at, ClosedAtOrBefore: at.Add(time.Hour)},
		{AccountRef: "acct-1", ClosedAtOrBefore: at.Add(time.Hour)},
		{AccountRef: "acct-1", ClosedAfter: at, ClosedAtOrBefore: at},
		{AccountRef: "acct-1", ClosedAfter: at.Add(time.Hour), ClosedAtOrBefore: at},
	} {
		reader := &fakeJournalLineageReader{}
		if _, err := store.CollectClosedStrategyTrades(context.Background(), reader, window, nil, at); err == nil {
			t.Fatalf("invalid window %+v was accepted", window)
		}
		if reader.calls != 0 {
			t.Fatalf("invalid window reached journal reader: calls=%d", reader.calls)
		}
	}
}

func TestJournalHandoffRequiresReader(t *testing.T) {
	store := openTestStore(t)
	at := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	if _, err := store.CollectClosedStrategyTrades(context.Background(), nil, ClosedTradeWindow{
		AccountRef: "acct-1", ClosedAfter: at, ClosedAtOrBefore: at.Add(time.Hour),
	}, nil, at.Add(time.Hour)); err == nil {
		t.Fatal("nil journal reader was accepted")
	}
}

func TestJournalHandoffReaderErrorWritesNothing(t *testing.T) {
	store := openTestStore(t)
	at := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	reader := &fakeJournalLineageReader{err: errors.New("synthetic journal failure")}
	_, err := store.CollectClosedStrategyTrades(context.Background(), reader, ClosedTradeWindow{
		AccountRef: "acct-1", ClosedAfter: at, ClosedAtOrBefore: at.Add(time.Hour),
	}, nil, at.Add(time.Hour))
	if err == nil || reader.calls != 1 {
		t.Fatalf("reader error=%v calls=%d", err, reader.calls)
	}
	var trades int
	if err := store.db.QueryRow(`SELECT count(*) FROM performance_trades`).Scan(&trades); err != nil || trades != 0 {
		t.Fatalf("reader error wrote trades=%d err=%v", trades, err)
	}
}

func TestJournalHandoffStoreBindsOneServerSelectedAccount(t *testing.T) {
	store := openTestStore(t)
	at := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	reader := &fakeJournalLineageReader{}
	for _, account := range []string{"acct-1", "acct-1"} {
		if _, err := store.CollectClosedStrategyTrades(context.Background(), reader, ClosedTradeWindow{
			AccountRef: account, ClosedAfter: at, ClosedAtOrBefore: at.Add(time.Hour),
		}, nil, at.Add(time.Hour)); err != nil {
			t.Fatalf("same account binding %q: %v", account, err)
		}
	}
	if _, err := store.CollectClosedStrategyTrades(context.Background(), reader, ClosedTradeWindow{
		AccountRef: "acct-2", ClosedAfter: at, ClosedAtOrBefore: at.Add(time.Hour),
	}, nil, at.Add(time.Hour)); err == nil {
		t.Fatal("same performance store accepted a second account")
	}
}
