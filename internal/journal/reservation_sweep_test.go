package journal

import (
	"context"
	"errors"
	"testing"
	"time"
)

// reservation_sweep_test.go covers the restart sweep and the as-of contention
// (task 3.3).
//
// # Why the contention test injects a version rather than racing goroutines
//
// The journal holds one connection (SetMaxOpenConns(1)), so two reservation
// transactions can never interleave: running them concurrently under -race
// proves that the mutex works, not that the as-of condition does. What has to
// be shown is that a snapshot which *predates a committed reservation* is
// detected and re-collected — so the tests below make the ledger move between
// the caller reading its version and the transaction running, which is exactly
// the state a slow broker round trip produces in production.

func TestStartupSweepReleasesExpiredUnconsumedHolds(t *testing.T) {
	j, fake := openReservationJournal(t)
	ctx := context.Background()

	unsent := reserveOne(t, j, "d-sweep-unsent", "acct-1")
	fake.Advance(11 * time.Minute)

	sweep, err := j.SweepReservations(ctx)
	if err != nil {
		t.Fatalf("SweepReservations: %v", err)
	}
	if len(sweep.Released) != 1 || sweep.Released[0].ReservationID != unsent {
		t.Fatalf("want the unconsumed expiry released, got %+v", sweep.Released)
	}
	if sweep.Released[0].Reason != ReleaseReasonExpiredUnconsumed {
		t.Fatalf("release reason = %q, want %s", sweep.Released[0].Reason, ReleaseReasonExpiredUnconsumed)
	}
	if reservationState(t, j, unsent).Held() {
		t.Fatal("the hold survived the startup sweep")
	}

	// Idempotent: a second pass has nothing left to do.
	again, err := j.SweepReservations(ctx)
	if err != nil {
		t.Fatalf("second SweepReservations: %v", err)
	}
	if len(again.Released) != 0 {
		t.Fatalf("the sweep must be idempotent, second pass released %+v", again.Released)
	}
}

func TestStartupSweepPreservesUnresolvedHolds(t *testing.T) {
	j, fake := openReservationJournal(t)
	ctx := context.Background()

	dec := recordEntryDecision(t, j, "d-sweep-unres", "acct-1")
	out, err := j.Reserve(ctx, exposureReserve(j, dec.ID, "acct-1", "100", "0", "10000",
		mustVersion(t, j, "acct-1")))
	if err != nil {
		t.Fatalf("Reserve: %v", err)
	}
	attempt, err := j.Prepare(ctx, reservationPrepare(dec, "attempt-sweep-unres"))
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	if err := attempt.MarkDispatchStarted(ctx); err != nil {
		t.Fatalf("MarkDispatchStarted: %v", err)
	}
	if err := attempt.MarkInDoubt(ctx, "test", "no answer"); err != nil {
		t.Fatalf("MarkInDoubt: %v", err)
	}
	if err := attempt.ResolveUnresolved(ctx, "", "neither proven"); err != nil {
		t.Fatalf("ResolveUnresolved: %v", err)
	}

	// The decision has also expired by now, so this is the case that matters:
	// expiry does not release a hold whose nonce was spent, and the sweep says
	// so rather than quietly leaving it out of the report.
	fake.Advance(11 * time.Minute)

	sweep, err := j.SweepReservations(ctx)
	if err != nil {
		t.Fatalf("SweepReservations: %v", err)
	}
	if len(sweep.Released) != 0 {
		t.Fatalf("an unresolved attempt's hold must not be released, got %+v", sweep.Released)
	}
	if !reservationState(t, j, out.Reservations[0].ID).Held() {
		t.Fatal("the hold of an unresolved attempt was released by the startup sweep")
	}
	if len(sweep.Preserved) == 0 {
		t.Fatal("a preserved hold must be reported; an invisible ratchet is not operable")
	}
	var causes []string
	for _, alert := range sweep.Preserved {
		causes = append(causes, alert.Cause)
	}
	if !containsCause(causes, AlertCauseUnresolved) && !containsCause(causes, AlertCauseNonceSpentExpired) {
		t.Fatalf("want an operator-facing cause, got %v", causes)
	}
	// One reservation, one line: the hold qualifies on two counts and an
	// operator should see it once.
	if len(sweep.Preserved) != 1 {
		t.Fatalf("want one preserved entry, got %d: %+v", len(sweep.Preserved), sweep.Preserved)
	}
}

func containsCause(causes []string, want string) bool {
	for _, c := range causes {
		if c == want {
			return true
		}
	}
	return false
}

// TestStartupSweepRecoversAnOrphanedTerminalHold models a row from before the
// release rode in the terminal transaction: the attempt is terminal, the hold
// is not. Nothing in the live path can produce it now, and a hold nothing will
// ever release shrinks the account's limits forever.
func TestStartupSweepRecoversAnOrphanedTerminalHold(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()

	dec := recordEntryDecision(t, j, "d-orphan", "acct-1")
	out, err := j.Reserve(ctx, exposureReserve(j, dec.ID, "acct-1", "100", "0", "10000",
		mustVersion(t, j, "acct-1")))
	if err != nil {
		t.Fatalf("Reserve: %v", err)
	}
	if _, err := j.Prepare(ctx, reservationPrepare(dec, "attempt-orphan")); err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	// Straight to the row: this is the state an older build left behind, and
	// going through Settle would release the hold as part of the transition —
	// which is the behaviour the live path already has.
	if _, err := j.db.ExecContext(ctx,
		"UPDATE mutation_attempts SET state = ? WHERE id = ?",
		string(StateFailedConfirmed), "attempt-orphan"); err != nil {
		t.Fatalf("orphaning the attempt: %v", err)
	}
	if !reservationState(t, j, out.Reservations[0].ID).Held() {
		t.Fatal("precondition: the hold should still be held before the sweep")
	}

	sweep, err := j.SweepReservations(ctx)
	if err != nil {
		t.Fatalf("SweepReservations: %v", err)
	}
	if len(sweep.Released) != 1 || sweep.Released[0].Reason != ReleaseReasonBrokerTerminal {
		t.Fatalf("want the orphaned hold released as BROKER_TERMINAL, got %+v", sweep.Released)
	}
	if reservationState(t, j, out.Reservations[0].ID).Held() {
		t.Fatal("the orphaned hold survived the sweep")
	}
}

// TestOrphanSweepDoesNotReleaseAFailClosedObservation keeps the assumed-expiry
// case out of the startup path too: a snapshot the derivation could not explain
// is not terminal, so the sweep must leave its hold alone.
func TestOrphanSweepDoesNotReleaseAFailClosedObservation(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()

	dec := recordEntryDecision(t, j, "d-orphan-unknown", "acct-1")
	out, err := j.Reserve(ctx, exposureReserve(j, dec.ID, "acct-1", "100", "0", "10000",
		mustVersion(t, j, "acct-1")))
	if err != nil {
		t.Fatalf("Reserve: %v", err)
	}
	confirmAttempt(t, j, dec, "attempt-orphan-unknown", "order-orphan-1")
	if _, err := j.RecordFill(ctx, FillObservation{
		OrderID: "order-orphan-1", Symbol: "005930", Market: "kr",
		State: "UNKNOWN_BROKER_STATE", Terminal: false, FailClosed: true,
		Reason: "closed_without_fill_or_cancel", Detail: "closed with nothing filled and no cancellation",
		Quantity: "10", FilledQuantity: "0", ObservedAt: j.nowString(),
	}); err != nil {
		t.Fatalf("RecordFill: %v", err)
	}

	sweep, err := j.SweepReservations(ctx)
	if err != nil {
		t.Fatalf("SweepReservations: %v", err)
	}
	if len(sweep.Released) != 0 {
		t.Fatalf("a fail-closed observation is not terminal; nothing may be released, got %+v", sweep.Released)
	}
	if !reservationState(t, j, out.Reservations[0].ID).Held() {
		t.Fatal("the startup sweep released a hold on an assumed expiry")
	}
	if len(sweep.Preserved) != 1 || sweep.Preserved[0].Cause != AlertCauseBrokerStateUnknown {
		t.Fatalf("want one UNKNOWN_BROKER_STATE alert, got %+v", sweep.Preserved)
	}
}

// --- the as-of contention ---------------------------------------------------

// TestASnapshotThatPredatesACommittedReservationIsRecollected is the injected
// contention the task asks for: the ledger moves between the caller reading its
// version and the reservation transaction running.
//
// Without the version check the second decision would be measured against a
// snapshot that does not know about the first one's hold, and both would fit
// under a limit only one of them fits under.
func TestASnapshotThatPredatesACommittedReservationIsRecollected(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()

	recordEntryDecision(t, j, "d-interfering", "acct-1")
	recordEntryDecision(t, j, "d-late", "acct-1")

	collected := 0
	out, err := j.ReserveWithRecollection(ctx,
		func(_ context.Context, attempt int) (ReserveRequest, error) {
			collected = attempt
			// The version is read first, as a caller must: before the broker
			// round trip that produces the snapshot.
			version := mustVersion(t, j, "acct-1")
			if attempt == 1 {
				// …and during that round trip, another decision reserves.
				if _, err := j.Reserve(ctx, exposureReserve(j, "d-interfering", "acct-1",
					"400", "0", "1000", mustVersion(t, j, "acct-1"))); err != nil {
					t.Fatalf("the interfering reservation: %v", err)
				}
			}
			return exposureReserve(j, "d-late", "acct-1", "400", "0", "1000", version), nil
		}, RecollectPolicy{})
	if err != nil {
		t.Fatalf("ReserveWithRecollection: %v", err)
	}
	if collected != 2 {
		t.Fatalf("the superseded snapshot must trigger exactly one re-collection, collected %d times", collected)
	}
	if len(out.Reservations) != 1 {
		t.Fatalf("want one reservation from the second snapshot, got %d", len(out.Reservations))
	}

	// Both holds are counted, which is what the re-collection bought: the
	// second snapshot was measured against a ledger that included the first.
	held, err := j.HeldReservations(ctx, "acct-1")
	if err != nil {
		t.Fatalf("HeldReservations: %v", err)
	}
	if len(held) != 2 {
		t.Fatalf("want two held reservations, got %d", len(held))
	}
	// A third 400 now reaches the 1000 limit, proving the aggregate saw both.
	recordEntryDecision(t, j, "d-third", "acct-1")
	if _, err := j.Reserve(ctx, exposureReserve(j, "d-third", "acct-1", "400", "0", "1000",
		mustVersion(t, j, "acct-1"))); !errors.Is(err, ErrReservationLimitExceeded) {
		t.Fatalf("400 + 400 + 400 must reach the 1000 limit, got %v", err)
	}
}

// TestAnUnchangedLedgerNeedsNoRecollection is the negative control: without it,
// a version check that always failed would pass the test above.
func TestAnUnchangedLedgerNeedsNoRecollection(t *testing.T) {
	j, _ := openReservationJournal(t)
	recordEntryDecision(t, j, "d-quiet", "acct-1")

	collected := 0
	if _, err := j.ReserveWithRecollection(context.Background(),
		func(_ context.Context, attempt int) (ReserveRequest, error) {
			collected = attempt
			return exposureReserve(j, "d-quiet", "acct-1", "100", "0", "1000",
				mustVersion(t, j, "acct-1")), nil
		}, RecollectPolicy{}); err != nil {
		t.Fatalf("ReserveWithRecollection: %v", err)
	}
	if collected != 1 {
		t.Fatalf("a ledger nobody moved must not force a re-collection, collected %d times", collected)
	}
}

// TestALapsedHoldReleasedByTheSweepMovesTheVersion pins the other half of the
// version's definition: releases move it too, so a caller that sized against a
// ledger the sweep has since changed re-collects rather than proceeding.
func TestALapsedHoldReleasedByTheSweepMovesTheVersion(t *testing.T) {
	j, fake := openReservationJournal(t)
	ctx := context.Background()

	reserveOne(t, j, "d-version-lapse", "acct-1")
	before := mustVersion(t, j, "acct-1")

	fake.Advance(11 * time.Minute)
	if _, err := j.SweepReservations(ctx); err != nil {
		t.Fatalf("SweepReservations: %v", err)
	}
	after := mustVersion(t, j, "acct-1")
	if after <= before {
		t.Fatalf("a release must move the ledger version: %d → %d", before, after)
	}
}
