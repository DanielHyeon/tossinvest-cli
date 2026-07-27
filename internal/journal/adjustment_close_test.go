package journal

import (
	"context"
	"testing"
)

// adjustment_close_test.go covers the adjustment-to-zero rule (change
// adopt-external-positions task 1.5; position-ledger and design A7).
//
// A position can reach zero two ways. A fill closes it inside the transaction
// that records the fill, which completes the exit state and freezes the outcome.
// An *adjustment* closes it because the account says the shares are gone and no
// local fill explains that — somebody sold them by hand. Then:
//
//	the exit state   is completed, with an ADJUSTMENT_CLOSED event saying why.
//	                 Leaving it open would keep a position that no longer exists
//	                 in the observation loop's working set forever.
//	the outcome      is not written. There is no sell leg the engine can price,
//	                 and a round trip priced with an empty one records the whole
//	                 position as a total loss — a fabricated number that then
//	                 flows into every aggregate an operator reads.
//
// The rule is common to engine-entered and adopted positions: an orphan is an
// orphan whichever record justified it.

func adjustToZero(t *testing.T, j *Journal, symbol, prevQuantity string) AdjustmentResult {
	t.Helper()
	ctx := context.Background()
	watermark, err := j.FillWatermark(ctx, symbol)
	if err != nil {
		t.Fatal(err)
	}
	result, err := j.ApplyPositionAdjustment(ctx, AdjustmentRequest{
		AccountRef:            "acct-1",
		Market:                "kr",
		Symbol:                symbol,
		Kind:                  AdjustmentExternal,
		ExpectedPrevQuantity:  prevQuantity,
		ExpectedFillWatermark: watermark,
		NewQuantity:           "0",
		BrokerAsOf:            "2026-03-30T01:00:00Z",
		Evidence:              "the account no longer holds it and no local fill explains that",
	})
	if err != nil {
		t.Fatalf("ApplyPositionAdjustment: %v", err)
	}
	return result
}

func assertClosedByAdjustment(t *testing.T, j *Journal, positionID string) {
	t.Helper()
	ctx := context.Background()

	state, err := j.ExitState(ctx, positionID)
	if err != nil {
		t.Fatalf("ExitState: %v", err)
	}
	if !state.Completed {
		t.Error("the exit state is still open over a position that holds nothing; the observation " +
			"loop would keep it in its working set forever")
	}

	events, err := j.ExitEvents(ctx, positionID)
	if err != nil {
		t.Fatal(err)
	}
	closes := 0
	for _, e := range events {
		if e.Action == ExitEventAdjustmentClosed {
			closes++
		}
	}
	if closes != 1 {
		t.Errorf("ADJUSTMENT_CLOSED events = %d, want exactly 1 (%+v)", closes, events)
	}

	var outcomes int
	if err := j.db.QueryRowContext(ctx,
		"SELECT count(*) FROM trade_outcomes WHERE position_id = ?", positionID).Scan(&outcomes); err != nil {
		t.Fatal(err)
	}
	if outcomes != 0 {
		t.Errorf("trade_outcomes rows = %d, want 0: the sell happened outside the engine, so there is "+
			"no sell leg to price and a row would record the position as a total loss", outcomes)
	}
}

func TestAdjustmentToZeroClosesAnAdoptedPosition(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	adoptable(t, j, "p-ext")
	if _, err := j.AdoptPosition(ctx, sampleAdoption("p-ext")); err != nil {
		t.Fatal(err)
	}
	if _, err := j.OpenAdoptedExitState(ctx, "p-ext"); err != nil {
		t.Fatal(err)
	}

	result := adjustToZero(t, j, "005930", "10")
	if !result.ClosedExitState {
		t.Error("ClosedExitState = false; the caller has no cue to alert that the engine stopped " +
			"protecting a position")
	}
	if result.Position.State != PositionClosed {
		t.Errorf("position state = %q, want CLOSED", result.Position.State)
	}
	assertClosedByAdjustment(t, j, "p-ext")
}

// TestAdjustmentToZeroClosesAnEngineEnteredPosition is the "공통" half: the rule
// is about orphaned exit states, not about how the position was acquired.
func TestAdjustmentToZeroClosesAnEngineEnteredPosition(t *testing.T) {
	j := openTestJournal(t)
	insertDecision(t, j, "decision-1", "nonce-close-1")
	insertPosition(t, j, "p-engine", "decision-1")
	insertExitState(t, j, "p-engine")

	result := adjustToZero(t, j, "005930", "10")
	if !result.ClosedExitState {
		t.Error("ClosedExitState = false for an engine-entered position")
	}
	assertClosedByAdjustment(t, j, "p-engine")
}

// TestAdjustmentToZeroLeavesAnUnmanagedPositionAlone keeps the rule narrow: a
// holding nobody adopted has no exit state, and inventing an event for it would
// make the judgement history claim a protection decision that never existed.
func TestAdjustmentToZeroLeavesAnUnmanagedPositionAlone(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	adoptable(t, j, "p-bare")

	result := adjustToZero(t, j, "005930", "10")
	if result.ClosedExitState {
		t.Error("ClosedExitState = true for a position that was never managed")
	}
	events, err := j.ExitEvents(ctx, "p-bare")
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 0 {
		t.Errorf("exit events on an unmanaged position = %+v, want none", events)
	}
}

// TestAdjustmentToZeroDoesNotCloseTwice pins the idempotence: a re-collection
// that recomputes the same difference, or a second adjustment on an instance
// whose policy has already finished, must not append a second closing event.
func TestAdjustmentToZeroDoesNotCloseTwice(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	adoptable(t, j, "p-ext")
	if _, err := j.AdoptPosition(ctx, sampleAdoption("p-ext")); err != nil {
		t.Fatal(err)
	}
	if _, err := j.OpenAdoptedExitState(ctx, "p-ext"); err != nil {
		t.Fatal(err)
	}

	adjustToZero(t, j, "005930", "10")

	// The same adjustment again: recognised by its derived id, so nothing is
	// written at all.
	again := adjustToZero(t, j, "005930", "10")
	if again.ClosedExitState {
		t.Error("a replayed adjustment reported a second close")
	}
	assertClosedByAdjustment(t, j, "p-ext")

	// And a *different* adjustment landing on the closed instance: it opens the
	// next instance rather than reopening this one, so this one's history stays
	// as it was.
	events, err := j.ExitEvents(ctx, "p-ext")
	if err != nil {
		t.Fatal(err)
	}
	if got := len(events); got != 2 {
		t.Errorf("exit events = %d (%+v), want OPENED and ADJUSTMENT_CLOSED", got, events)
	}
}
