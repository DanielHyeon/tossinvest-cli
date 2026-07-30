package main

// consolesignals_test.go covers the wiring behind the console's /signals screen.
//
// Two claims worth a test, and both are about what the seam does NOT do: it must
// not read a source, and it must not leave the store open. The first is spec
// Requirement 7's rate budget; the second is the write-ahead log D16 says the
// watch loop has to be able to reclaim.

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/candidate"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// signalsFixtureStore opens a temporary discovery store and points the seam's
// opener at that path. It hands back a fresh handle on every call, which is what
// the production opener does.
func signalsFixtureStore(t *testing.T, clk clock.Clock) *consoleSignals {
	t.Helper()
	path := filepath.Join(t.TempDir(), candidate.DBFileName)
	open := func(ctx context.Context) (*candidate.Store, func(), error) {
		store, err := candidate.Open(ctx, candidate.Options{
			Path:  path,
			Clock: clk,
			FSProber: candidate.FixedFSProber(candidate.FSInfo{
				Name: "ext4", FreeBytes: 1 << 40, FreeMeasured: true,
			}),
		})
		if err != nil {
			return nil, func() {}, err
		}
		return store, func() { _ = store.Close() }, nil
	}
	return &consoleSignals{open: open}
}

// TestTheSignalsSeamReadsTheStoreAndCallsNoSource.
//
// spec Requirement 7 ranks the account's rate budget engine > verification >
// discovery, and a console tab reloading every discovery tick would be a second
// discoverer competing with `tossctl candidate watch` for the undocumented RANKING
// limit — where a single 429 costs the whole source, because the rankings have no
// WTS fallback (D14). The seam is handed no panel at all, which is how that is
// enforced rather than promised: there is nothing for it to call.
func TestTheSignalsSeamReadsTheStoreAndCallsNoSource(t *testing.T) {
	at := time.Date(2026, 7, 28, 9, 0, 0, 0, time.UTC)
	seam := signalsFixtureStore(t, clock.NewFake(at))
	ctx := context.Background()

	// Seed one candidate through the store's own writer.
	store, _, err := seam.open(ctx)
	if err != nil {
		t.Fatalf("opening the fixture store: %v", err)
	}
	if err := store.RecordObservations(ctx, []candidate.Observation{{
		Market: candidate.MarketKR, Symbol: "005930",
		Source: candidate.SourceOfficialTradingValue, ObservedAt: at,
		Reported: candidate.Reported{Rank: 12, RankTotal: 100, TradingValue: "1000000"},
	}}); err != nil {
		t.Fatalf("RecordObservations: %v", err)
	}
	if _, err := store.Promote(ctx, candidate.MarketKR, "005930", at); err != nil {
		t.Fatalf("Promote: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	got, err := seam.Signals(ctx)
	if err != nil {
		t.Fatalf("Signals: %v", err)
	}
	if !got.At.Equal(at) {
		t.Errorf("the reading is stamped %s, want the store's clock %s", got.At, at)
	}
	if len(got.Markets) != len(consoleSignalsMarkets) {
		t.Fatalf("the reading covers %d markets, want %d; a market missing from the page is "+
			"indistinguishable from a market with nothing in it",
			len(got.Markets), len(consoleSignalsMarkets))
	}
	kr := got.Markets[0]
	if kr.Market != candidate.MarketKR {
		t.Fatalf("the first market is %q, want %q", kr.Market, candidate.MarketKR)
	}
	if kr.Why != "" {
		t.Fatalf("KR could not be assessed: %s", kr.Why)
	}
	if len(kr.Verdicts) != 1 || kr.Verdicts[0].Summary.Symbol != "005930" {
		t.Fatalf("the assessment holds %d verdicts, want the one candidate that was stored",
			len(kr.Verdicts))
	}
	if kr.Vetoes.Total != 1 {
		t.Errorf("the veto tally counts %d candidates, want 1", kr.Vetoes.Total)
	}
	// D18: two of the three vetoes have no approved threshold, so nothing can pass.
	// The seam must not invent one to make the column non-zero.
	if kr.Vetoes.Passed != 0 {
		t.Errorf("the seam reports %d passed candidates; seen_late and extended have no "+
			"approved threshold, so a pass here means a number was invented", kr.Vetoes.Passed)
	}
	if kr.Vetoes.Unmeasured != 1 {
		t.Errorf("the seam reports %d unmeasured candidates, want 1", kr.Vetoes.Unmeasured)
	}
}

// TestTheSignalsSeamDoesNotHoldTheStoreOpenAcrossReads.
//
// Store.Checkpoint names this screen as the long-lived reader that stops
// `wal_checkpoint(TRUNCATE)` from reclaiming anything, and the write-ahead log
// grows beside the table on the filesystem the order ledger writes to (D16). A
// console that kept a handle would quietly disable the sweep the watch loop runs
// every turn — while nothing failed.
//
// The measurement is that a checkpoint taken after a read is not busy.
func TestTheSignalsSeamDoesNotHoldTheStoreOpenAcrossReads(t *testing.T) {
	at := time.Date(2026, 7, 28, 9, 0, 0, 0, time.UTC)
	seam := signalsFixtureStore(t, clock.NewFake(at))
	ctx := context.Background()

	if _, err := seam.Signals(ctx); err != nil {
		t.Fatalf("Signals: %v", err)
	}
	store, release, err := seam.open(ctx)
	if err != nil {
		t.Fatalf("opening the fixture store: %v", err)
	}
	defer release()
	busy, _, err := store.Checkpoint(ctx)
	if err != nil {
		t.Fatalf("Checkpoint: %v", err)
	}
	if busy {
		t.Error("the write-ahead log could not be reclaimed after a screen read; the seam is " +
			"holding the store open and the watch loop's cleanup is off")
	}
}

// TestTheSignalsSeamSaysWhereTheMissingSourceNamesAreRatherThanClaimingNoneAreMissing.
//
// ScanResult.Missing is the only thing that pairs a source id with the reason it
// did not answer, and it lives for one cycle inside the process that ran it. The
// console is not that process. What must not happen is for that gap to render as
// "no source is missing", which is the same confusion as an unmeasured veto
// reading as a pass — so the panel comes back unmeasured, with the sentence that
// says where the names are.
func TestTheSignalsSeamSaysWhereTheMissingSourceNamesAreRatherThanClaimingNoneAreMissing(t *testing.T) {
	seam := signalsFixtureStore(t, clock.NewFake(time.Date(2026, 7, 28, 9, 0, 0, 0, time.UTC)))
	got, err := seam.Signals(context.Background())
	if err != nil {
		t.Fatalf("Signals: %v", err)
	}
	for _, m := range got.Markets {
		if m.Panel.Known {
			t.Errorf("%s reports a known source panel; this process ran no scan, so it has no "+
				"panel report to give and a measured one would be invented", m.Market)
		}
		if m.Panel.Why == "" {
			t.Errorf("%s reports an unmeasured panel with no reason; a reason-less unmeasured "+
				"is the display task 5.7 exists to remove", m.Market)
		}
	}
}
