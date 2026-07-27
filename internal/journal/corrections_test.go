package journal

// corrections_test.go pins the EXECUTION_CORRECTION rule
// (extend-execution-contract task 5.3) and the verbatim storage of the order
// identifiers this table is keyed by (the residual violations issues.md assigned
// to 5.3).
//
// The requirement is order-execution "체결 정정 이벤트": an observation with the
// same cumulative quantity but a different average price or filled amount is
// recorded as a correction — no quantity delta, no fill event — inside the same
// transaction that holds the previous snapshot.

import (
	"context"
	"path/filepath"
	"testing"
)

// correctionObservation is `observation` plus the amount field, which is the
// half of a correction the P1 schema could not see.
func correctionObservation(orderID, filled, avg, amount string) FillObservation {
	obs := observation(orderID, filled)
	obs.AveragePrice = avg
	obs.FilledAmount = amount
	return obs
}

func correctionsOf(t *testing.T, j *Journal, orderID string) []ExecutionCorrection {
	t.Helper()
	rows, err := j.ExecutionCorrections(context.Background(), orderID)
	if err != nil {
		t.Fatalf("ExecutionCorrections(%q): %v", orderID, err)
	}
	return rows
}

// TestAveragePriceCorrectionIsRecordedWithoutAQuantityDelta is the spec's
// "수량 동일·평균가 변경" scenario.
func TestAveragePriceCorrectionIsRecordedWithoutAQuantityDelta(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	if _, err := j.RecordFill(ctx, correctionObservation("o-1", "3", "213.4", "640.2")); err != nil {
		t.Fatal(err)
	}
	res, err := j.RecordFill(ctx, correctionObservation("o-1", "3", "214.05", "640.2"))
	if err != nil {
		t.Fatal(err)
	}
	if res.Delta != "0" || res.DeltaQuantity != 0 {
		t.Fatalf("delta = %q, want 0 — a correction moves no quantity", res.Delta)
	}
	if !res.Corrected {
		t.Fatal("the result must report the correction so a caller can alert on it")
	}

	events, err := j.FillEvents(ctx, "o-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("fill events = %d, want the single original fill — a correction is not a fill", len(events))
	}

	rows := correctionsOf(t, j, "o-1")
	if len(rows) != 1 {
		t.Fatalf("corrections = %d, want 1", len(rows))
	}
	got := rows[0]
	if got.PrevAveragePrice != "213.4" || got.NewAveragePrice != "214.05" {
		t.Errorf("average price %q -> %q, want 213.4 -> 214.05", got.PrevAveragePrice, got.NewAveragePrice)
	}
	if got.CumulativeQuantity != "3" {
		t.Errorf("cumulative quantity = %q, want the unchanged 3", got.CumulativeQuantity)
	}

	stored, err := j.LookupFill(ctx, "o-1")
	if err != nil {
		t.Fatal(err)
	}
	if stored.AveragePrice != "214.05" {
		t.Errorf("stored average price = %q, want the corrected value", stored.AveragePrice)
	}
}

// TestAmountOnlyCorrectionIsDetected is why fill_snapshots grew filled_amount:
// with an unchanged quantity and an unchanged average, the amount is the only
// evidence left that the broker restated the execution.
func TestAmountOnlyCorrectionIsDetected(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	if _, err := j.RecordFill(ctx, correctionObservation("o-1", "3", "213.4", "640.2")); err != nil {
		t.Fatal(err)
	}
	res, err := j.RecordFill(ctx, correctionObservation("o-1", "3", "213.4", "640.35"))
	if err != nil {
		t.Fatal(err)
	}
	if !res.Corrected || res.Delta != "0" {
		t.Fatalf("res = %+v, want an amount-only correction with no delta", res)
	}
	rows := correctionsOf(t, j, "o-1")
	if len(rows) != 1 {
		t.Fatalf("corrections = %d, want 1", len(rows))
	}
	if rows[0].PrevFilledAmount != "640.2" || rows[0].NewFilledAmount != "640.35" {
		t.Errorf("amount %q -> %q, want 640.2 -> 640.35",
			rows[0].PrevFilledAmount, rows[0].NewFilledAmount)
	}
	stored, err := j.LookupFill(ctx, "o-1")
	if err != nil {
		t.Fatal(err)
	}
	if stored.FilledAmount != "640.35" {
		t.Errorf("stored filled amount = %q, want the corrected value", stored.FilledAmount)
	}
}

// TestUnobservedCorrectionValuesAreStoredAsEmptyNotNull is the Manager's ruling
// on issues.md: the UNIQUE key that stops a re-observation double-inserting is
// disabled by a NULL, because SQLite treats NULLs in an index as distinct. The
// empty string means "the broker gave none", as it does in fill_snapshots.
func TestUnobservedCorrectionValuesAreStoredAsEmptyNotNull(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	// The broker reported an average but never an amount.
	if _, err := j.RecordFill(ctx, correctionObservation("o-1", "3", "213.4", "")); err != nil {
		t.Fatal(err)
	}
	if _, err := j.RecordFill(ctx, correctionObservation("o-1", "3", "214.05", "")); err != nil {
		t.Fatal(err)
	}

	var nulls int
	if err := j.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM execution_corrections
		 WHERE new_avg_price IS NULL OR new_filled_amount IS NULL`).Scan(&nulls); err != nil {
		t.Fatal(err)
	}
	if nulls != 0 {
		t.Fatalf("%d correction rows carry a NULL dedup column; the UNIQUE is off for them", nulls)
	}
	rows := correctionsOf(t, j, "o-1")
	if len(rows) != 1 || rows[0].NewFilledAmount != "" {
		t.Fatalf("corrections = %+v, want one row with an empty amount", rows)
	}
}

// TestRepeatedPollOfACorrectedSnapshotChangesNothing is the spec's "동일 관측
// 반복" scenario: the poller reads the corrected order every cycle, and only the
// first of those reads is a correction.
func TestRepeatedPollOfACorrectedSnapshotChangesNothing(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	if _, err := j.RecordFill(ctx, correctionObservation("o-1", "3", "213.4", "640.2")); err != nil {
		t.Fatal(err)
	}
	corrected := correctionObservation("o-1", "3", "214.05", "640.35")
	if _, err := j.RecordFill(ctx, corrected); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 4; i++ {
		repeat := corrected
		repeat.ObservedAt = "2026-03-30T00:31:0" + string(rune('0'+i)) + "Z"
		res, err := j.RecordFill(ctx, repeat)
		if err != nil {
			t.Fatalf("poll %d: %v", i, err)
		}
		if res.Changed || res.Corrected || res.FailClosed {
			t.Fatalf("poll %d = %+v, want a no-op", i, res)
		}
	}
	if rows := correctionsOf(t, j, "o-1"); len(rows) != 1 {
		t.Fatalf("corrections after four repeat polls = %d, want 1", len(rows))
	}
}

// TestReObservedCorrectionIsDedupedByTheUniqueKey is the crash case D9's UNIQUE
// exists for. A restart (or a broker that restates an execution back and forth)
// can present a triple that was already recorded; the second insert of the same
// (order, cumulative quantity, new average, new amount) is one row, not two.
func TestReObservedCorrectionIsDedupedByTheUniqueKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	ctx := context.Background()

	j := openTestJournalAt(t, path)
	if _, err := j.RecordFill(ctx, correctionObservation("o-1", "3", "213.4", "640.2")); err != nil {
		t.Fatal(err)
	}
	if _, err := j.RecordFill(ctx, correctionObservation("o-1", "3", "214.05", "640.35")); err != nil {
		t.Fatal(err)
	}
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}

	// A fresh process. The broker restates the execution back to what it first
	// said, and then to the correction again: the third correction repeats the
	// second one's key.
	reopened := openTestJournalAt(t, path)
	if _, err := reopened.RecordFill(ctx, correctionObservation("o-1", "3", "213.4", "640.2")); err != nil {
		t.Fatal(err)
	}
	res, err := reopened.RecordFill(ctx, correctionObservation("o-1", "3", "214.05", "640.35"))
	if err != nil {
		t.Fatalf("the replayed correction must not error: %v", err)
	}
	if res.FailClosed {
		t.Fatalf("res = %+v, want the replay absorbed, not refused", res)
	}

	rows := correctionsOf(t, reopened, "o-1")
	if len(rows) != 2 {
		t.Fatalf("corrections = %d, want 2 (214.05 and the restatement back to 213.4); "+
			"the repeat of 214.05 is deduped by the UNIQUE key", len(rows))
	}
	// The snapshot still tracks the latest observation.
	stored, err := reopened.LookupFill(ctx, "o-1")
	if err != nil {
		t.Fatal(err)
	}
	if stored.AveragePrice != "214.05" || stored.FilledAmount != "640.35" {
		t.Errorf("stored snapshot = %+v, want the latest observation", stored)
	}
}

// TestCorrectionCarriesTheAccountItBelongsTo: execution_corrections.account_ref
// is NOT NULL because a correction on a live account has to be attributable. The
// detector's Snapshot has no account dimension, so the journal derives it from
// the confirmed attempt that named the order.
func TestCorrectionCarriesTheAccountItBelongsTo(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	attempt, err := j.Prepare(ctx, testRequest())
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	if _, err := attempt.Dispatch(ctx, ackDispatch("o-1")); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}

	if _, err := j.RecordFill(ctx, correctionObservation("o-1", "3", "213.4", "640.2")); err != nil {
		t.Fatal(err)
	}
	if _, err := j.RecordFill(ctx, correctionObservation("o-1", "3", "214.05", "640.2")); err != nil {
		t.Fatal(err)
	}
	rows := correctionsOf(t, j, "o-1")
	if len(rows) != 1 {
		t.Fatalf("corrections = %d, want 1", len(rows))
	}
	if rows[0].AccountRef != "acct-1" {
		t.Errorf("account_ref = %q, want acct-1 from the intent behind the order", rows[0].AccountRef)
	}
}

// TestCallerSuppliedAccountRefWins: a caller that knows the account does not have
// to rely on the derivation.
func TestCallerSuppliedAccountRefWins(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	first := correctionObservation("o-1", "3", "213.4", "640.2")
	first.AccountRef = "acct-9"
	if _, err := j.RecordFill(ctx, first); err != nil {
		t.Fatal(err)
	}
	second := correctionObservation("o-1", "3", "214.05", "640.2")
	second.AccountRef = "acct-9"
	if _, err := j.RecordFill(ctx, second); err != nil {
		t.Fatal(err)
	}
	rows := correctionsOf(t, j, "o-1")
	if len(rows) != 1 || rows[0].AccountRef != "acct-9" {
		t.Fatalf("corrections = %+v, want one row for acct-9", rows)
	}
}

// TestQuantityChangeIsAFillNotACorrection guards the boundary from the other
// side: a moving quantity is the P1 fill path, untouched.
func TestQuantityChangeIsAFillNotACorrection(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	if _, err := j.RecordFill(ctx, correctionObservation("o-1", "3", "213.4", "640.2")); err != nil {
		t.Fatal(err)
	}
	res, err := j.RecordFill(ctx, correctionObservation("o-1", "7", "214.05", "1498.35"))
	if err != nil {
		t.Fatal(err)
	}
	if res.Corrected {
		t.Fatal("a quantity delta is a fill; recording it as a correction as well would double count")
	}
	if res.Delta != "4" {
		t.Fatalf("delta = %q, want 4", res.Delta)
	}
	if rows := correctionsOf(t, j, "o-1"); len(rows) != 0 {
		t.Fatalf("corrections = %d, want none", len(rows))
	}
}

// TestDecreasedQuantityStillFailsClosedWithAnAmount: P1's fail-closed rule is
// unchanged by the amount column, and a refused snapshot records no correction.
func TestDecreasedQuantityStillFailsClosedWithAnAmount(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	if _, err := j.RecordFill(ctx, correctionObservation("o-1", "7", "213.4", "1493.8")); err != nil {
		t.Fatal(err)
	}
	res, err := j.RecordFill(ctx, correctionObservation("o-1", "3", "213.4", "640.2"))
	if err != nil {
		t.Fatal(err)
	}
	if !res.FailClosed || res.Reason != ReasonFillDecreased {
		t.Fatalf("res = %+v, want a fail-closed decrease", res)
	}
	if res.Corrected {
		t.Fatal("a refused snapshot must not also be recorded as a correction")
	}
	if rows := correctionsOf(t, j, "o-1"); len(rows) != 0 {
		t.Fatalf("corrections = %d, want none", len(rows))
	}
	stored, err := j.LookupFill(ctx, "o-1")
	if err != nil {
		t.Fatal(err)
	}
	if stored.FilledAmount != "1493.8" {
		t.Errorf("stored amount = %q, want the last trusted 1493.8", stored.FilledAmount)
	}
}

// TestFillSnapshotStoresTheOrderIdVerbatim closes the residual opaque-identifier
// violation issues.md assigned here (fills.go:173/385/393). `orderId` has no
// contracted shape (openapi), so trimming on the way in stores an identifier the
// broker never issued and a later byte comparison silently misses it.
func TestFillSnapshotStoresTheOrderIdVerbatim(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	const padded = " o-1\t"
	if _, err := j.RecordFill(ctx, observation(padded, "3")); err != nil {
		t.Fatal(err)
	}

	var stored string
	if err := j.db.QueryRowContext(ctx, `SELECT order_id FROM fill_snapshots`).Scan(&stored); err != nil {
		t.Fatal(err)
	}
	if stored != padded {
		t.Errorf("stored order id = %q, want %q", stored, padded)
	}

	if _, err := j.LookupFill(ctx, padded); err != nil {
		t.Errorf("LookupFill with the identifier as received: %v", err)
	}
	events, err := j.FillEvents(ctx, padded)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].OrderID != padded {
		t.Errorf("fill events = %+v, want one keyed by the verbatim id", events)
	}
}

// TestBlankOrderIdIsStillRefused: trimming survives in the emptiness check,
// which is the only thing it was ever for.
func TestBlankOrderIdIsStillRefused(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	if _, err := j.RecordFill(ctx, observation("  \t ", "3")); err == nil {
		t.Fatal("a blank order id is not a name and must be refused")
	}
}

// TestResolveConfirmedWithLineageStoresTheIdentifierVerbatim closes the
// asymmetry 5.2 left behind: ResolveConfirmed stores the identifier as received,
// and this — the amend half of the same resolution write path — trimmed it.
func TestResolveConfirmedWithLineageStoresTheIdentifierVerbatim(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	attempt, err := j.Prepare(ctx, testRequest())
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	if err := attempt.MarkDispatchStarted(ctx); err != nil {
		t.Fatalf("MarkDispatchStarted: %v", err)
	}
	if err := attempt.MarkInDoubt(ctx, "test", "parked for the test"); err != nil {
		t.Fatalf("MarkInDoubt: %v", err)
	}

	const padded = " child-7 "
	if err := attempt.ResolveConfirmedWithLineage(ctx, LineageEdge{
		ParentOrderID:        "parent-1",
		ChildOrderID:         padded,
		ParentFilledQuantity: "2",
		RequestedQuantity:    "8",
	}, ReasonResolvedFound, "the amendment created a new order"); err != nil {
		t.Fatalf("ResolveConfirmedWithLineage: %v", err)
	}

	rec, err := j.LookupAttempt(ctx, "attempt-1")
	if err != nil {
		t.Fatal(err)
	}
	if rec.BrokerOrderID != padded {
		t.Errorf("stored broker order id = %q, want %q — the same rule ResolveConfirmed follows",
			rec.BrokerOrderID, padded)
	}
}
