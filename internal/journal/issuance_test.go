package journal

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// issuance_test.go pins the atomic issuance API (task 0.2, design D1).
//
// The assertion that carries the whole design is negative: after a refusal,
// there is no decisions row. Everything else here exists to make that one
// unfakeable — the nonce becomes reusable (so the row is gone rather than
// merely invisible), the ledger version has not moved, and the concurrent case
// leaves exactly one decision for exactly one winner.

// issueRequest builds a one-row OPEN_EXPOSURE issuance for an entry decision.
func issueRequest(j *Journal, id, account, amount, usage, limit string, version int64) IssueRequest {
	issued := j.Now()
	return IssueRequest{
		Decision: DecisionRequest{
			ID:          id,
			AccountRef:  account,
			SafetyClass: SafetyClassExposureRaising,
			Kind:        KindPlace,
			Preimage: RiskIntent{
				AccountRef: account, Market: "kr", Symbol: "005930", Side: "BUY",
				Quantity: "10", EntryPrice: "70000", StopPrice: "68000",
				PolicyVersion: "test-1",
			},
			LimitsJSON: `{"max_notional":"1000000"}`,
			Nonce:      "nonce-" + id,
			IssuedAt:   issued,
			ExpiresAt:  issued.Add(60 * time.Second),
		},
		Reserve: ReserveRequest{
			SnapshotAsOf:    issued,
			ObservedVersion: version,
			SnapshotUsage: []AggregateAmount{
				{Kind: ReservationKindOpenExposure, Amount: usage, Currency: "KRW"},
			},
			Limits: []AggregateAmount{
				{Kind: ReservationKindOpenExposure, Amount: limit, Currency: "KRW"},
			},
			Reservations: []ReservationRequest{
				{ID: "res-" + id, Kind: ReservationKindOpenExposure, Amount: amount, Currency: "KRW"},
			},
		},
	}
}

func decisionExists(t *testing.T, j *Journal, id string) bool {
	t.Helper()
	_, err := j.LookupDecision(context.Background(), id)
	switch {
	case err == nil:
		return true
	case errors.Is(err, ErrDecisionNotFound):
		return false
	default:
		t.Fatalf("LookupDecision(%s): %v", id, err)
		return false
	}
}

// TestIssuanceWritesTheDecisionAndItsReservation is the success path: both rows
// exist, the reservation names the decision, and the returned version is the
// one a caller can hand to the next issuance without re-reading.
func TestIssuanceWritesTheDecisionAndItsReservation(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()

	before := mustVersion(t, j, "acct-1")
	out, err := j.RecordDecisionAndReserve(ctx, issueRequest(j, "d-1", "acct-1", "600", "0", "1000", before))
	if err != nil {
		t.Fatalf("RecordDecisionAndReserve: %v", err)
	}
	if out.Decision.ID != "d-1" || out.Decision.RiskHash == "" {
		t.Fatalf("issued decision = %+v, want d-1 with a bound hash", out.Decision)
	}
	if len(out.Reservations) != 1 || out.Reservations[0].DecisionID != "d-1" {
		t.Fatalf("reservations = %+v, want one bound to d-1", out.Reservations)
	}
	if !decisionExists(t, j, "d-1") {
		t.Fatal("the decision must be readable after a successful issuance")
	}
	held, err := j.HeldReservations(ctx, "acct-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(held) != 1 || held[0].Amount != "600" {
		t.Fatalf("held = %+v, want the one 600 reservation", held)
	}
	if after := mustVersion(t, j, "acct-1"); after != out.Version || after <= before {
		t.Errorf("ledger version = %d (returned %d), want it advanced past %d", after, out.Version, before)
	}
}

// TestRefusedReservationLeavesNoDecision is the requirement's own scenario
// (risk-management: 예약 거부 시 결정 부재). The limit is already exhausted, the
// reservation is refused, and the decision the same transaction wrote must be
// gone with it — otherwise a submittable authorisation exists that nothing paid
// for.
func TestRefusedReservationLeavesNoDecision(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()

	// Fill the account: 600 held against a 1000 limit leaves room for 399.
	if _, err := j.RecordDecisionAndReserve(ctx,
		issueRequest(j, "d-first", "acct-1", "600", "0", "1000", mustVersion(t, j, "acct-1"))); err != nil {
		t.Fatalf("seeding the held reservation: %v", err)
	}

	beforeVersion := mustVersion(t, j, "acct-1")
	_, err := j.RecordDecisionAndReserve(ctx,
		issueRequest(j, "d-refused", "acct-1", "500", "0", "1000", beforeVersion))
	if !errors.Is(err, ErrReservationLimitExceeded) {
		t.Fatalf("want the limit refusal, got %v", err)
	}
	if reason, ok := IssueRefusalReason(err); !ok || reason != IssueReasonLimitReached {
		t.Errorf("reason = %q (%v), want %s", reason, ok, IssueReasonLimitReached)
	}

	// The point of the whole file: no orphan decision.
	if decisionExists(t, j, "d-refused") {
		t.Fatal("a refused issuance left a submittable decision on disk")
	}
	reservations, err := j.ReservationsForDecision(ctx, "d-refused")
	if err != nil {
		t.Fatal(err)
	}
	if len(reservations) != 0 {
		t.Errorf("refused issuance left %d reservations", len(reservations))
	}
	if after := mustVersion(t, j, "acct-1"); after != beforeVersion {
		t.Errorf("ledger version moved from %d to %d on a refusal", beforeVersion, after)
	}

	// The nonce is free again, which is only true if the row was rolled back and
	// not merely hidden: nonce is UNIQUE across the whole table.
	retry := issueRequest(j, "d-retry", "acct-1", "100", "0", "1000", mustVersion(t, j, "acct-1"))
	retry.Decision.Nonce = "nonce-d-refused"
	if _, err := j.RecordDecisionAndReserve(ctx, retry); err != nil {
		t.Fatalf("the refused decision's nonce must be free again: %v", err)
	}
}

// TestIssuanceRefusesAnExpiredDecision covers the reason that needed a new
// sentinel. The decision is well-formed — it expires after it was issued — but
// its expiry is already in the past when the transaction reads it, so it
// authorises nothing and nothing is written.
func TestIssuanceRefusesAnExpiredDecision(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()

	req := issueRequest(j, "d-expired", "acct-1", "100", "0", "1000", mustVersion(t, j, "acct-1"))
	req.Decision.IssuedAt = j.Now().Add(-2 * time.Minute)
	req.Decision.ExpiresAt = j.Now().Add(-1 * time.Minute)

	_, err := j.RecordDecisionAndReserve(ctx, req)
	if !errors.Is(err, ErrDecisionExpired) {
		t.Fatalf("want ErrDecisionExpired, got %v", err)
	}
	// The new sentinel does not weaken the old contract: every caller that
	// already tests for ErrInvalidRequest still sees one.
	if !errors.Is(err, ErrInvalidRequest) {
		t.Error("ErrDecisionExpired must still satisfy errors.Is(err, ErrInvalidRequest)")
	}
	if reason, ok := IssueRefusalReason(err); !ok || reason != IssueReasonDecisionExpired {
		t.Errorf("reason = %q (%v), want %s", reason, ok, IssueReasonDecisionExpired)
	}
	if decisionExists(t, j, "d-expired") {
		t.Fatal("an expired decision must not survive its own refused issuance")
	}
}

// TestIssuanceRefusesAStaleOrSupersededSnapshot: both are VERSION_CONFLICT for
// a single-shot issuance, and both leave nothing behind.
func TestIssuanceRefusesAStaleOrSupersededSnapshot(t *testing.T) {
	j, fake := openReservationJournal(t)
	ctx := context.Background()

	superseded := issueRequest(j, "d-superseded", "acct-1", "100", "0", "1000",
		mustVersion(t, j, "acct-1")+7)
	_, err := j.RecordDecisionAndReserve(ctx, superseded)
	if !errors.Is(err, ErrSnapshotSuperseded) {
		t.Fatalf("want ErrSnapshotSuperseded, got %v", err)
	}
	if reason, ok := IssueRefusalReason(err); !ok || reason != IssueReasonVersionConflict {
		t.Errorf("reason = %q (%v), want %s", reason, ok, IssueReasonVersionConflict)
	}
	if decisionExists(t, j, "d-superseded") {
		t.Fatal("a superseded issuance left a decision behind")
	}

	stale := issueRequest(j, "d-stale", "acct-1", "100", "0", "1000", mustVersion(t, j, "acct-1"))
	fake.Advance(11 * time.Second) // past riskcalc.AccountSnapshotStaleness
	stale.Decision.IssuedAt = j.Now()
	stale.Decision.ExpiresAt = j.Now().Add(60 * time.Second)
	_, err = j.RecordDecisionAndReserve(ctx, stale)
	if !errors.Is(err, ErrSnapshotStale) {
		t.Fatalf("want ErrSnapshotStale, got %v", err)
	}
	if reason, ok := IssueRefusalReason(err); !ok || reason != IssueReasonVersionConflict {
		t.Errorf("reason = %q (%v), want %s", reason, ok, IssueReasonVersionConflict)
	}
	if decisionExists(t, j, "d-stale") {
		t.Fatal("a stale issuance left a decision behind")
	}
}

// TestIssuanceWithRecollectionConverges: the loop retries the two retryable
// refusals and issues on the attempt whose snapshot holds.
func TestIssuanceWithRecollectionConverges(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()

	var attempts int
	out, err := j.RecordDecisionAndReserveWithRecollection(ctx,
		func(_ context.Context, attempt int) (IssueRequest, error) {
			attempts = attempt
			version := mustVersion(t, j, "acct-1")
			if attempt < 3 {
				// A version nobody is at: the ledger has moved under this snapshot.
				version += 5
			}
			return issueRequest(j, "d-1", "acct-1", "100", "0", "1000", version), nil
		}, RecollectPolicy{})
	if err != nil {
		t.Fatalf("RecordDecisionAndReserveWithRecollection: %v", err)
	}
	if attempts != 3 {
		t.Errorf("attempts = %d, want the third to be the one that held", attempts)
	}
	if out.Decision.ID != "d-1" || !decisionExists(t, j, "d-1") {
		t.Fatal("the converged attempt must have written the decision")
	}
	// Exactly one issuance happened, not one per attempt.
	reservations, err := j.ReservationsForDecision(ctx, "d-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(reservations) != 1 {
		t.Fatalf("reservations = %d, want 1: the refused attempts must have rolled back", len(reservations))
	}
}

// TestIssuanceWithRecollectionExhausts is the other end of the loop: the cap is
// spent, the answer is a refusal and never the last snapshot seen, and no
// attempt left a decision behind.
func TestIssuanceWithRecollectionExhausts(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()

	_, err := j.RecordDecisionAndReserveWithRecollection(ctx,
		func(_ context.Context, _ int) (IssueRequest, error) {
			return issueRequest(j, "d-1", "acct-1", "100", "0", "1000",
				mustVersion(t, j, "acct-1")+5), nil
		}, RecollectPolicy{MaxAttempts: 3})
	if !errors.Is(err, ErrRecollectionExhausted) {
		t.Fatalf("want ErrRecollectionExhausted, got %v", err)
	}
	// The exhaustion wraps the last refusal, so the reason mapping has to test
	// the outer sentinel first or it reports the race the loop already survived.
	if !errors.Is(err, ErrSnapshotSuperseded) {
		t.Error("the exhaustion must carry the last reason with it")
	}
	if reason, ok := IssueRefusalReason(err); !ok || reason != IssueReasonRecollectionExhausted {
		t.Errorf("reason = %q (%v), want %s", reason, ok, IssueReasonRecollectionExhausted)
	}
	if decisionExists(t, j, "d-1") {
		t.Fatal("an exhausted re-collection left a decision behind")
	}
}

// TestIssuanceWithRecollectionStopsOnALimit: a limit refusal is not retryable.
// Collecting the same account again will not make it fit, and retrying a limit
// is how a cap becomes a suggestion.
func TestIssuanceWithRecollectionStopsOnALimit(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()

	var attempts int
	_, err := j.RecordDecisionAndReserveWithRecollection(ctx,
		func(_ context.Context, attempt int) (IssueRequest, error) {
			attempts = attempt
			return issueRequest(j, "d-1", "acct-1", "2000", "0", "1000",
				mustVersion(t, j, "acct-1")), nil
		}, RecollectPolicy{MaxAttempts: 3})
	if !errors.Is(err, ErrReservationLimitExceeded) {
		t.Fatalf("want the limit refusal, got %v", err)
	}
	if attempts != 1 {
		t.Errorf("attempts = %d, want 1: a limit is not a reason to re-collect", attempts)
	}
	if decisionExists(t, j, "d-1") {
		t.Fatal("a limit refusal left a decision behind")
	}
}

// TestIssuanceWithRecollectionExpiresWithTheDecision: the re-collection budget
// is bounded by the decision's TTL, not the other way round. A loop that
// outlives the decision refuses with DECISION_EXPIRED rather than issuing an
// authorisation whose reasoning is older than its own freshness bound.
func TestIssuanceWithRecollectionExpiresWithTheDecision(t *testing.T) {
	j, fake := openReservationJournal(t)
	ctx := context.Background()

	issued := j.Now()
	_, err := j.RecordDecisionAndReserveWithRecollection(ctx,
		func(_ context.Context, attempt int) (IssueRequest, error) {
			if attempt > 1 {
				// The clock moves between attempts, as a real re-collection's
				// broker round trip would move it.
				fake.Advance(4 * time.Second)
			}
			// The first two attempts race the ledger; the third finally carries a
			// snapshot that holds — by which time the decision behind it has run
			// out, which is the case this test exists for.
			version := mustVersion(t, j, "acct-1")
			if attempt < 3 {
				version += 5
			}
			req := issueRequest(j, "d-1", "acct-1", "100", "0", "1000", version)
			req.Decision.IssuedAt = issued
			// A TTL shorter than the loop's budget, so the decision runs out first.
			req.Decision.ExpiresAt = issued.Add(5 * time.Second)
			req.Reserve.SnapshotAsOf = j.Now()
			return req, nil
		}, RecollectPolicy{MaxAttempts: 5, Budget: 30 * time.Second})
	if !errors.Is(err, ErrDecisionExpired) {
		t.Fatalf("want ErrDecisionExpired, got %v", err)
	}
	if reason, ok := IssueRefusalReason(err); !ok || reason != IssueReasonDecisionExpired {
		t.Errorf("reason = %q (%v), want %s", reason, ok, IssueReasonDecisionExpired)
	}
	if decisionExists(t, j, "d-1") {
		t.Fatal("an expired re-collection left a decision behind")
	}
}

// TestConcurrentIssuancesCannotBothTakeTheLastSlot is the atomic property under
// contention: two entries race for the last slot of an aggregate limit, exactly
// one is issued, and the loser leaves neither a reservation nor a decision.
func TestConcurrentIssuancesCannotBothTakeTheLastSlot(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()

	// Room for one 600 reservation under a 1000 limit; a second would reach it.
	version := mustVersion(t, j, "acct-1")

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		winners []string
		errs    []error
	)
	for _, id := range []string{"d-a", "d-b"} {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			_, err := j.RecordDecisionAndReserve(ctx,
				issueRequest(j, id, "acct-1", "600", "0", "1000", version))
			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				winners = append(winners, id)
			} else {
				errs = append(errs, err)
			}
		}(id)
	}
	wg.Wait()

	if len(winners) != 1 {
		t.Fatalf("winners = %v, want exactly one issuance to succeed (errors: %v)", winners, errs)
	}
	// The loser is refused for one of the two reasons this transaction can
	// produce under contention, and both are classified.
	for _, err := range errs {
		reason, ok := IssueRefusalReason(err)
		if !ok {
			t.Errorf("an unclassified refusal under contention: %v", err)
			continue
		}
		if reason != IssueReasonLimitReached && reason != IssueReasonVersionConflict {
			t.Errorf("refusal reason = %s, want LIMIT_REACHED or VERSION_CONFLICT (%v)", reason, err)
		}
	}

	held, err := j.HeldReservations(ctx, "acct-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(held) != 1 {
		t.Fatalf("held reservations = %d, want 1", len(held))
	}
	for _, id := range []string{"d-a", "d-b"} {
		if got := decisionExists(t, j, id); got != (id == winners[0]) {
			t.Errorf("decision %s exists = %v, want %v", id, got, id == winners[0])
		}
	}
}

// TestIssuanceRefusesAMismatchedPairing keeps the two halves from naming
// different things. A reservation pointing at another decision is a bug worth
// stopping on, not a field worth silently correcting.
func TestIssuanceRefusesAMismatchedPairing(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()

	wrongDecision := issueRequest(j, "d-1", "acct-1", "100", "0", "1000", mustVersion(t, j, "acct-1"))
	wrongDecision.Reserve.DecisionID = "d-other"
	if _, err := j.RecordDecisionAndReserve(ctx, wrongDecision); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("a reservation against another decision must be refused, got %v", err)
	}
	if decisionExists(t, j, "d-1") {
		t.Fatal("a refused pairing must not have written anything")
	}

	wrongAccount := issueRequest(j, "d-2", "acct-1", "100", "0", "1000", mustVersion(t, j, "acct-1"))
	wrongAccount.Reserve.AccountRef = "acct-2"
	if _, err := j.RecordDecisionAndReserve(ctx, wrongAccount); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("a reservation on another account must be refused, got %v", err)
	}
	if decisionExists(t, j, "d-2") {
		t.Fatal("a refused pairing must not have written anything")
	}
}

// TestIssuanceRejectsAMalformedDecisionBeforeReserving: the decision is
// validated first, so a bad request never takes the write lock and never
// reaches the limit arithmetic.
func TestIssuanceRejectsAMalformedDecisionBeforeReserving(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()

	req := issueRequest(j, "d-1", "acct-1", "100", "0", "1000", mustVersion(t, j, "acct-1"))
	req.Decision.Preimage = RiskIntent{
		AccountRef: "acct-1", Market: "kr", Symbol: "005930", Side: "BUY",
		Quantity: "10", EntryPrice: "70000", // no stop price
		PolicyVersion: "test-1",
	}
	if _, err := j.RecordDecisionAndReserve(ctx, req); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("a decision with no stop must be refused, got %v", err)
	}
	if decisionExists(t, j, "d-1") {
		t.Fatal("a malformed decision must not have been written")
	}
	if held, err := j.HeldReservations(ctx, "acct-1"); err != nil || len(held) != 0 {
		t.Fatalf("held = %v (err %v), want none", held, err)
	}
}

// TestIssueRefusalReasonIgnoresWhatIsNotARiskRefusal: a malformed request or a
// database failure is a bug or an outage. Classifying either as a risk refusal
// would file it under "the account is full" and hide it.
func TestIssueRefusalReasonIgnoresWhatIsNotARiskRefusal(t *testing.T) {
	for _, err := range []error{
		nil,
		ErrInvalidRequest,
		ErrDecisionNotFound,
		errors.New("disk full"),
	} {
		if reason, ok := IssueRefusalReason(err); ok {
			t.Errorf("IssueRefusalReason(%v) = %q, want no classification", err, reason)
		}
	}
}

// TestSingleShotReserveStillRefusesTheExpiredDecision keeps the landed API's
// behaviour: the new sentinel changed what the error *is*, not what the
// single-shot Reserve does with it.
func TestSingleShotReserveStillRefusesTheExpiredDecision(t *testing.T) {
	j, fake := openReservationJournal(t)
	ctx := context.Background()

	recordEntryDecision(t, j, "d-expired", "acct-1")
	fake.Advance(11 * time.Minute)
	req := exposureReserve(j, "d-expired", "acct-1", "1", "0", "1000", mustVersion(t, j, "acct-1"))
	req.SnapshotAsOf = j.Now()

	_, err := j.Reserve(ctx, req)
	if !errors.Is(err, ErrDecisionExpired) || !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("want an expiry refusal that is still an invalid request, got %v", err)
	}
	// The decision was written by its own transaction here, so it survives: only
	// the atomic API rolls a decision back, and only the one it wrote itself.
	if !decisionExists(t, j, "d-expired") {
		t.Fatal("Reserve must not delete a decision it did not write")
	}
}
