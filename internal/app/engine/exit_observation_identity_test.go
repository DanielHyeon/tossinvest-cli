package engine

import (
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

func TestStableObservationIDUsesFetchedAtAndIgnoresCycleFallback(t *testing.T) {
	m := managed{position: journal.Position{
		ID: "position-secret", Market: "kr", Symbol: "005930", InstanceSeq: 7,
	}}
	quote := observedQuote{Price: "9700.00", FetchedAt: time.Date(2026, 7, 31, 1, 2, 3, 4, time.UTC)}
	first, err := stableObservationID("account-secret", m, quote,
		cycleObservation{at: time.Unix(1, 0), sequence: 1})
	if err != nil {
		t.Fatal(err)
	}
	second, err := stableObservationID("account-secret", m, quote,
		cycleObservation{at: time.Unix(99, 0), sequence: 99})
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("same fetched quote IDs differ: %q != %q", first, second)
	}
	if strings.Contains(first, "account-secret") || strings.Contains(first, "position-secret") {
		t.Fatalf("opaque observation id leaked source identifiers: %q", first)
	}
}

func TestStableObservationIDReusesOneFallbackWithinCycle(t *testing.T) {
	m := managed{position: journal.Position{
		ID: "p-1", Market: "kr", Symbol: "005930", InstanceSeq: 7,
	}}
	quote := observedQuote{Price: "9700"}
	fallback := cycleObservation{at: time.Date(2026, 7, 31, 1, 2, 3, 4, time.UTC), sequence: 41}
	first, err := stableObservationID("account", m, quote, fallback)
	if err != nil {
		t.Fatal(err)
	}
	second, err := stableObservationID("account", m, observedQuote{Price: "9700.0"}, fallback)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("one fallback cycle produced different canonical IDs: %q != %q", first, second)
	}
	next, err := stableObservationID("account", m, quote,
		cycleObservation{at: fallback.at, sequence: fallback.sequence + 1})
	if err != nil {
		t.Fatal(err)
	}
	if next == first {
		t.Fatal("two zero-timestamp cycles reused one observation id")
	}
}
