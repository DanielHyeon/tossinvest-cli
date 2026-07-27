package filldetect_test

// payload_test.go pins what the observation payload carries into the ledger
// (extend-execution-contract task 5.3).
//
// Two properties, both about not losing information the correction rule needs:
// `execution.filledAmount` reaches the Snapshot as the decimal string it arrived
// as, and `orderId` reaches it unmodified because it is an opaque token.

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/filldetect"
)

// appliedSnapshots returns the snapshots the detector offered, in order.
func (l *fakeLedger) appliedSnapshots() []filldetect.Snapshot {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]filldetect.Snapshot(nil), l.applied...)
}

// pollOne runs one cycle over a single open order and returns the snapshot the
// ledger was offered.
func pollOne(t *testing.T, order rawOrder) filldetect.Snapshot {
	t.Helper()
	pager := newPager(page("", order))
	d, _, ledger := newDetector(t, pager, &fakeOrderReader{}, &fakePositions{}, fakeTracked{}, nil)
	if _, err := d.PollOnce(context.Background()); err != nil {
		t.Fatalf("PollOnce: %v", err)
	}
	snaps := ledger.appliedSnapshots()
	if len(snaps) != 1 {
		t.Fatalf("applied snapshots = %d, want 1", len(snaps))
	}
	return snaps[0]
}

// TestSnapshotCarriesTheFilledAmount: without it, an amount-only restatement of
// an execution is invisible to the ledger (design D7). openapi types
// `execution.filledAmount` as a nullable decimal string.
func TestSnapshotCarriesTheFilledAmount(t *testing.T) {
	snap := pollOne(t, rawOrder{id: "o-1", quantity: "10", filled: "10",
		avgPrice: "70000", amount: "700000"})
	if snap.FilledAmount != "700000" {
		t.Errorf("FilledAmount = %q, want 700000", snap.FilledAmount)
	}
}

// TestFilledAmountKeepsItsDecimalString: the amount is compared byte-for-byte
// against the stored previous value, so a float round trip would manufacture
// corrections out of nothing.
func TestFilledAmountKeepsItsDecimalString(t *testing.T) {
	snap := pollOne(t, rawOrder{id: "o-1", quantity: "10", filled: "10",
		avgPrice: "37.05", amount: "370.50"})
	if snap.FilledAmount != "370.50" {
		t.Errorf("FilledAmount = %q, want the string as sent (370.50, not 370.5)", snap.FilledAmount)
	}
}

// TestAbsentFilledAmountIsEmptyNotZero: the broker documents the field as
// nullable, and "" means "not observed" all the way through — the same
// convention the average price already uses. A 0 would read as a real amount.
func TestAbsentFilledAmountIsEmptyNotZero(t *testing.T) {
	for _, tc := range []struct {
		name  string
		order rawOrder
	}{
		{name: "field missing", order: rawOrder{id: "o-1", filled: "0"}},
		{name: "field null", order: rawOrder{id: "o-1", filled: "0", rawAmount: "null"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := pollOne(t, tc.order).FilledAmount; got != "" {
				t.Errorf("FilledAmount = %q, want empty", got)
			}
		})
	}
}

// TestUnreadableFilledAmountFailsTheCycle keeps this package's rule: everything
// unreadable is an error, never a zero.
func TestUnreadableFilledAmountFailsTheCycle(t *testing.T) {
	clk := clock.NewFake(pollStart)
	pager := newPager(page("", rawOrder{id: "o-1", filled: "1", amount: "seven hundred"}))
	d := &filldetect.Detector{
		Orders: pager, Order: &fakeOrderReader{}, Positions: &fakePositions{},
		Tracked: fakeTracked{}, Ledger: newLedger(clk), Clock: clk,
	}
	if _, err := d.PollOnce(context.Background()); err == nil {
		t.Fatal("an unreadable filledAmount must fail the cycle rather than become an empty amount")
	}
}

// TestSnapshotKeepsTheOrderIdVerbatim closes the residual violation issues.md
// assigned here (payload.go:84). `orderId` is opaque — openapi contracts no
// shape — so normalising it on the way in produces an identifier the broker
// never issued, which no later byte comparison can match.
func TestSnapshotKeepsTheOrderIdVerbatim(t *testing.T) {
	const padded = "  o-1 " // the test helper writes the id into JSON literally
	snap := pollOne(t, rawOrder{id: padded, filled: "1"})
	if snap.OrderID != padded {
		t.Errorf("OrderID = %q, want %q", snap.OrderID, padded)
	}
}

// TestBlankPayloadIdFallsBackToTheTrackedId: trimming survives where it belongs,
// in the emptiness judgement. A tracked order whose payload names no id is still
// attributed to the order that was read.
func TestBlankPayloadIdFallsBackToTheTrackedId(t *testing.T) {
	reader := &fakeOrderReader{orders: map[string]json.RawMessage{
		"o-tracked": rawOrder{id: "  ", status: "CLOSED", quantity: "4", filled: "4"}.json(),
	}}
	tracked := fakeTracked{orders: []filldetect.TrackedOrder{
		{OrderID: "o-tracked", Symbol: "AAPL", Market: "us"},
	}}
	d, _, ledger := newDetector(t, newPager(page("")), reader, &fakePositions{}, tracked, nil)
	if _, err := d.PollOnce(context.Background()); err != nil {
		t.Fatalf("PollOnce: %v", err)
	}
	snaps := ledger.appliedSnapshots()
	if len(snaps) != 1 || snaps[0].OrderID != "o-tracked" {
		t.Fatalf("applied = %+v, want the tracked id to fill in for a blank payload id", snaps)
	}
}
