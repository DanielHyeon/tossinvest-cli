package filldetect_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/filldetect"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// Durable-ledger tests (harden-execution-base task 3.2).
//
// These run the real journal under the real detector, because the requirement is
// about what survives: "동일 스냅샷의 중복 수신(폴링·SSE 재조회·재시작 후)은
// 상태를 변경하지 않는다" is only meaningful if the prior observation is on disk.

func openLedgerJournal(t *testing.T, clk clock.Clock, path string) *journal.Journal {
	t.Helper()
	j, err := journal.Open(context.Background(), journal.Options{
		Path:  path,
		Clock: clk,
		// This repository lives on ntfs, so the filesystem allowlist is satisfied
		// with a fixed probe. The guard itself is covered by internal/journal.
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		t.Fatalf("journal.Open: %v", err)
	}
	t.Cleanup(func() { _ = j.Close() })
	return j
}

func newJournalDetector(t *testing.T, clk *clock.Fake, pager *fakePager, gate *execgw.EntryGate,
	path string) (*filldetect.Detector, *journal.Journal) {
	t.Helper()
	j := openLedgerJournal(t, clk, path)
	return &filldetect.Detector{
		Orders:    pager,
		Order:     &fakeOrderReader{},
		Positions: &fakePositions{},
		Tracked:   filldetect.JournalTracked{Journal: j, AccountRef: "acct-1"},
		Ledger:    filldetect.JournalLedger{Journal: j, AccountRef: "acct-1"},
		Clock:     clk,
		Gate:      gate,
	}, j
}

func recordConfirmedLedgerOrder(t *testing.T, j *journal.Journal, orderID string) {
	t.Helper()
	ctx := context.Background()
	attempt, err := j.Prepare(ctx, journal.PrepareRequest{
		Intent: journal.Intent{
			ID: "intent-" + orderID, Market: "us", TradingDay: "2026-03-30", AccountRef: "acct-1",
			Symbol: "AAPL", Side: "BUY", OrderType: "LIMIT", Quantity: "10",
			Price: "200", Currency: "USD", Source: "engine", Fingerprint: "fp-" + orderID,
		},
		Kind: journal.KindPlace, AttemptID: "attempt-" + orderID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := attempt.MarkDispatchStarted(ctx); err != nil {
		t.Fatal(err)
	}
	if err := attempt.MarkAcked(ctx, orderID); err != nil {
		t.Fatal(err)
	}
	if err := attempt.Settle(ctx, journal.StateConfirmed, journal.ReasonBrokerAcknowledged, "acked"); err != nil {
		t.Fatal(err)
	}
}

// TestTheSameSnapshotFromPollAndSSERefetchAppliesOnce is the spec scenario
// verbatim: the poll loop sees a fill, an SSE hint triggers an immediate
// re-fetch of the identical state, and the fill is reflected exactly once.
func TestTheSameSnapshotFromPollAndSSERefetchAppliesOnce(t *testing.T) {
	clk := clock.NewFake(pollStart)
	pager := newPager(page("", rawOrder{
		id: "o-1", quantity: "10", filled: "4", avgPrice: "213.4",
		filledAt: pollStart.Add(-2 * time.Second).Format(time.RFC3339),
	}))
	d, j := newJournalDetector(t, clk, pager, nil, filepath.Join(t.TempDir(), "journal.db"))
	ctx := context.Background()

	first, err := d.PollOnce(ctx)
	if err != nil {
		t.Fatalf("poll: %v", err)
	}
	if first.Fills != 1 {
		t.Fatalf("first poll fills = %d, want 1", first.Fills)
	}

	// The SSE hint's re-fetch: same state, moments later.
	clk.Advance(200 * time.Millisecond)
	second, err := d.PollOnce(ctx)
	if err != nil {
		t.Fatalf("re-fetch: %v", err)
	}
	if second.Fills != 0 {
		t.Fatalf("re-fetch fills = %d, want 0 — the same snapshot is not a second fill", second.Fills)
	}
	if len(second.Applied) != 1 || second.Applied[0].Changed {
		t.Fatalf("re-fetch applied = %+v, want a no-op", second.Applied)
	}

	events, err := j.FillEvents(ctx, "o-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("durable fill events = %d, want 1", len(events))
	}
	if events[0].DeltaQuantity != "4" {
		t.Fatalf("recorded delta = %q, want 4", events[0].DeltaQuantity)
	}
}

// TestOnlyThePositiveDeltaIsReflected walks a partial fill up to a full one and
// checks the durable record adds up.
func TestOnlyThePositiveDeltaIsReflected(t *testing.T) {
	clk := clock.NewFake(pollStart)
	pager := newPager(page("", rawOrder{id: "o-1", quantity: "10", filled: "4"}))
	d, j := newJournalDetector(t, clk, pager, nil, filepath.Join(t.TempDir(), "journal.db"))
	ctx := context.Background()

	if _, err := d.PollOnce(ctx); err != nil {
		t.Fatalf("poll 1: %v", err)
	}
	clk.Advance(3 * time.Second)
	pager.mu.Lock()
	pager.pages[""] = page("", rawOrder{id: "o-1", quantity: "10", filled: "9"})
	pager.mu.Unlock()

	cycle, err := d.PollOnce(ctx)
	if err != nil {
		t.Fatalf("poll 2: %v", err)
	}
	if cycle.Fills != 1 || cycle.Applied[0].Delta != 5 {
		t.Fatalf("second poll = %+v, want a delta of 5", cycle.Applied)
	}

	events, err := j.FillEvents(ctx, "o-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 {
		t.Fatalf("fill events = %d, want 2", len(events))
	}
	if events[1].CumulativeQuantity != "9" {
		t.Fatalf("cumulative = %q, want 9", events[1].CumulativeQuantity)
	}
}

// TestShrinkingFillBlocksEntries is the spec's "filledQuantity 감소 관측"
// scenario at detector level: the symbol is blocked and the reason is the stable
// UNKNOWN_BROKER_STATE code an operator and the alerting path both read.
func TestShrinkingFillBlocksEntries(t *testing.T) {
	clk := clock.NewFake(pollStart)
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
	pager := newPager(page("", rawOrder{id: "o-1", quantity: "10", filled: "7"}))
	d, j := newJournalDetector(t, clk, pager, gate, filepath.Join(t.TempDir(), "journal.db"))
	ctx := context.Background()

	if _, err := d.PollOnce(ctx); err != nil {
		t.Fatalf("poll 1: %v", err)
	}
	if rejected := gate.CheckEntry(); rejected != nil {
		t.Fatalf("a healthy poll must not block: %v", rejected)
	}

	clk.Advance(3 * time.Second)
	pager.mu.Lock()
	pager.pages[""] = page("", rawOrder{id: "o-1", quantity: "10", filled: "3"})
	pager.mu.Unlock()

	cycle, err := d.PollOnce(ctx)
	if err != nil {
		t.Fatalf("poll 2: %v", err)
	}
	if len(cycle.FailedClosed) != 1 || cycle.FailedClosed[0] != "o-1" {
		t.Fatalf("FailedClosed = %v, want o-1", cycle.FailedClosed)
	}
	rejected := gate.CheckEntry()
	if rejected == nil || rejected.Reason != execgw.ReasonBrokerStateUnknown {
		t.Fatalf("a shrinking fill must block new entries, got %v", rejected)
	}
	if !strings.Contains(rejected.Detail, journal.ReasonFillDecreased) {
		t.Fatalf("the block should name the rule that fired, got %q", rejected.Detail)
	}

	stored, err := j.LookupFill(ctx, "o-1")
	if err != nil {
		t.Fatal(err)
	}
	if stored.FilledQuantity != "7" {
		t.Fatalf("stored quantity = %q, want the last trusted 7", stored.FilledQuantity)
	}
}

// TestUnknownDerivationFailsClosed: a CLOSED order with a partial fill, no
// cancellation and no successor is indeterminate, and the priority table says so.
// The ledger must not quietly bank the partial fill as final.
func TestUnknownDerivationFailsClosed(t *testing.T) {
	clk := clock.NewFake(pollStart)
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
	pager := newPager(page("", rawOrder{id: "o-1", status: "CLOSED", quantity: "10", filled: "4"}))
	d, j := newJournalDetector(t, clk, pager, gate, filepath.Join(t.TempDir(), "journal.db"))
	ctx := context.Background()

	cycle, err := d.PollOnce(ctx)
	if err != nil {
		t.Fatalf("poll: %v", err)
	}
	if len(cycle.FailedClosed) != 1 {
		t.Fatalf("FailedClosed = %v, want the indeterminate order", cycle.FailedClosed)
	}
	if rejected := gate.CheckEntry(); rejected == nil ||
		rejected.Reason != execgw.ReasonBrokerStateUnknown {
		t.Fatalf("an indeterminate order must block new entries, got %v", rejected)
	}
	events, err := j.FillEvents(ctx, "o-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 0 {
		t.Fatalf("fill events = %d, want none from a refused snapshot", len(events))
	}
}

// TestRestartDoesNotReapplyKnownFills is the durability payoff: a fresh process
// reads the same cumulative quantity and reports no new fill.
func TestRestartDoesNotReapplyKnownFills(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	clk := clock.NewFake(pollStart)
	pager := newPager(page("", rawOrder{id: "o-1", quantity: "10", filled: "6"}))
	ctx := context.Background()

	first, j1 := newJournalDetector(t, clk, pager, nil, path)
	cycle, err := first.PollOnce(ctx)
	if err != nil {
		t.Fatalf("pre-restart poll: %v", err)
	}
	if cycle.Fills != 1 {
		t.Fatalf("pre-restart fills = %d, want 1", cycle.Fills)
	}
	if err := j1.Close(); err != nil {
		t.Fatal(err)
	}

	// A new process, same account, same open order.
	clk.Advance(time.Minute)
	second, j2 := newJournalDetector(t, clk, pager, nil, path)
	cycle, err = second.PollOnce(ctx)
	if err != nil {
		t.Fatalf("post-restart poll: %v", err)
	}
	if cycle.Fills != 0 {
		t.Fatalf("post-restart fills = %d, want 0 — the prior observation is on disk", cycle.Fills)
	}
	events, err := j2.FillEvents(ctx, "o-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("fill events across the restart = %d, want 1", len(events))
	}
}

// TestTerminalOrdersLeaveTheTrackedSet stops a filled order being read forever,
// and proves the tracked source is the journal's own view rather than a cache the
// detector keeps in memory.
func TestTerminalOrdersLeaveTheTrackedSet(t *testing.T) {
	clk := clock.NewFake(pollStart)
	path := filepath.Join(t.TempDir(), "journal.db")
	pager := newPager(page("", rawOrder{id: "o-1", quantity: "10", filled: "4"}))
	d, j := newJournalDetector(t, clk, pager, nil, path)
	ctx := context.Background()
	recordConfirmedLedgerOrder(t, j, "o-1")

	if _, err := d.PollOnce(ctx); err != nil {
		t.Fatalf("poll 1: %v", err)
	}
	tracked, err := filldetect.JournalTracked{Journal: j, AccountRef: "acct-1"}.TrackedOrders(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(tracked) != 1 {
		t.Fatalf("tracked = %+v, want the live order", tracked)
	}

	// The order fills and leaves the open list; the tracked read finds it CLOSED.
	clk.Advance(3 * time.Second)
	d.Order = &fakeOrderReader{orders: rawOrders(rawOrder{
		id: "o-1", status: "CLOSED", quantity: "10", filled: "10",
	})}
	pager.mu.Lock()
	pager.pages[""] = page("")
	pager.mu.Unlock()

	cycle, err := d.PollOnce(ctx)
	if err != nil {
		t.Fatalf("poll 2: %v", err)
	}
	if cycle.TrackedReads != 1 || cycle.Fills != 1 {
		t.Fatalf("cycle = %+v, want one tracked read producing the final fill", cycle)
	}
	tracked, err = filldetect.JournalTracked{Journal: j, AccountRef: "acct-1"}.TrackedOrders(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(tracked) != 0 {
		t.Fatalf("tracked after the order filled = %+v, want none", tracked)
	}
}

func TestJournalTrackedPreservesCanonicalOrderScope(t *testing.T) {
	clk := clock.NewFake(pollStart)
	j := openLedgerJournal(t, clk, filepath.Join(t.TempDir(), "journal.db"))
	recordConfirmedLedgerOrder(t, j, "o-scoped")

	source := filldetect.JournalTracked{Journal: j, AccountRef: "acct-1"}
	if got := source.SelectedAccountRef(); got != "acct-1" {
		t.Fatalf("selected account = %q, want acct-1", got)
	}
	tracked, err := source.TrackedOrders(context.Background())
	if err != nil {
		t.Fatalf("TrackedOrders: %v", err)
	}
	if len(tracked) != 1 {
		t.Fatalf("tracked = %+v, want one order", tracked)
	}
	got := tracked[0]
	if got.OrderID != "o-scoped" || got.AccountRef != "acct-1" || got.Market != "us" ||
		got.TradingDay != "2026-03-30" || got.Symbol != "AAPL" || got.Side != "BUY" {
		t.Fatalf("tracked canonical scope = %+v, want acct-1/us/2026-03-30/AAPL/BUY/o-scoped", got)
	}
}

func TestJournalLedgerRejectsSnapshotFromAnotherAccount(t *testing.T) {
	clk := clock.NewFake(pollStart)
	j := openLedgerJournal(t, clk, filepath.Join(t.TempDir(), "journal.db"))
	ledger := filldetect.JournalLedger{Journal: j, AccountRef: "acct-1"}
	_, err := ledger.Apply(context.Background(), filldetect.Snapshot{
		OrderID: "o-cross-account", AccountRef: "acct-2", Market: "us",
		TradingDay: "2026-03-30", Symbol: "AAPL", Side: "BUY",
	})
	if err == nil || !strings.Contains(err.Error(), "ledger account") {
		t.Fatalf("Apply error = %v, want account-scope rejection", err)
	}
}
