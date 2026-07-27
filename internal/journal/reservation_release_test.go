package journal

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// reservation_release_test.go pins the four exits from a held reservation and,
// more importantly, the exit that does not exist (task 3.2, design D5).
//
// The load-bearing test is TestUnknownBrokerStateKeepsTheReservationHeld: a
// CLOSED order with nothing filled and no cancellation is what an expiry would
// look like — and also what a rejection we never saw would look like — so the
// derivation keeps it at UNKNOWN_BROKER_STATE and the hold must survive.

// reserveOne records a decision, reserves against it and returns the
// reservation id. Everything after this point is about letting the hold go.
func reserveOne(t *testing.T, j *Journal, decisionID, account string) string {
	t.Helper()
	recordEntryDecision(t, j, decisionID, account)
	out, err := j.Reserve(context.Background(),
		exposureReserve(j, decisionID, account, "100", "0", "10000", mustVersion(t, j, account)))
	if err != nil {
		t.Fatalf("Reserve(%s): %v", decisionID, err)
	}
	return out.Reservations[0].ID
}

func reservationState(t *testing.T, j *Journal, id string) Reservation {
	t.Helper()
	res, err := j.LookupReservation(context.Background(), id)
	if err != nil {
		t.Fatalf("LookupReservation(%s): %v", id, err)
	}
	return res
}

// --- (a) derived terminal states --------------------------------------------

func TestNotDispatchedReleasesInTheSameTransaction(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()

	dec := recordEntryDecision(t, j, "d-nd", "acct-1")
	out, err := j.Reserve(ctx, exposureReserve(j, dec.ID, "acct-1", "100", "0", "10000",
		mustVersion(t, j, "acct-1")))
	if err != nil {
		t.Fatalf("Reserve: %v", err)
	}
	attempt, err := j.Prepare(ctx, reservationPrepare(dec, "attempt-nd"))
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}

	if err := attempt.Settle(ctx, StateNotDispatched, "test", "never sent"); err != nil {
		t.Fatalf("Settle: %v", err)
	}
	res := reservationState(t, j, out.Reservations[0].ID)
	if res.Held() {
		t.Fatal("a NOT_DISPATCHED attempt sent nothing; its hold must be released")
	}
	if res.ReleaseReason != ReleaseReasonBrokerTerminal {
		t.Fatalf("release reason = %q, want %s", res.ReleaseReason, ReleaseReasonBrokerTerminal)
	}
}

func TestFailedConfirmedReleasesAndConfirmedDoesNot(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()

	// FAILED_CONFIRMED: the broker definitively refused, so there is no exposure.
	failedDec := recordEntryDecision(t, j, "d-failed", "acct-1")
	failedOut, err := j.Reserve(ctx, exposureReserve(j, failedDec.ID, "acct-1", "100", "0", "10000",
		mustVersion(t, j, "acct-1")))
	if err != nil {
		t.Fatalf("Reserve: %v", err)
	}
	failed, err := j.Prepare(ctx, reservationPrepare(failedDec, "attempt-failed"))
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	if err := failed.MarkDispatchStarted(ctx); err != nil {
		t.Fatalf("MarkDispatchStarted: %v", err)
	}
	if err := failed.MarkInDoubt(ctx, "test", "no answer"); err != nil {
		t.Fatalf("MarkInDoubt: %v", err)
	}
	if err := failed.ResolveFailed(ctx, "", "proven absent"); err != nil {
		t.Fatalf("ResolveFailed: %v", err)
	}
	if reservationState(t, j, failedOut.Reservations[0].ID).Held() {
		t.Fatal("FAILED_CONFIRMED means nothing exists at the broker; the hold must be released")
	}

	// CONFIRMED: the order exists. The hold stands until the *order* reaches a
	// derived terminal state, which arrives through the fill ledger.
	okDec := recordEntryDecision(t, j, "d-confirmed", "acct-1")
	okOut, err := j.Reserve(ctx, exposureReserve(j, okDec.ID, "acct-1", "100", "0", "10000",
		mustVersion(t, j, "acct-1")))
	if err != nil {
		t.Fatalf("Reserve: %v", err)
	}
	ok, err := j.Prepare(ctx, reservationPrepare(okDec, "attempt-confirmed"))
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	if err := ok.MarkDispatchStarted(ctx); err != nil {
		t.Fatalf("MarkDispatchStarted: %v", err)
	}
	if err := ok.MarkAcked(ctx, "order-1"); err != nil {
		t.Fatalf("MarkAcked: %v", err)
	}
	if err := ok.ResolveConfirmed(ctx, "order-1", "", "found"); err != nil {
		t.Fatalf("ResolveConfirmed: %v", err)
	}
	if !reservationState(t, j, okOut.Reservations[0].ID).Held() {
		t.Fatal("a CONFIRMED order exists; releasing its hold would free headroom for live exposure")
	}
}

func TestUnresolvedInDoubtDoesNotRelease(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()

	dec := recordEntryDecision(t, j, "d-unres", "acct-1")
	out, err := j.Reserve(ctx, exposureReserve(j, dec.ID, "acct-1", "100", "0", "10000",
		mustVersion(t, j, "acct-1")))
	if err != nil {
		t.Fatalf("Reserve: %v", err)
	}
	attempt, err := j.Prepare(ctx, reservationPrepare(dec, "attempt-unres"))
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	if err := attempt.MarkDispatchStarted(ctx); err != nil {
		t.Fatalf("MarkDispatchStarted: %v", err)
	}
	if err := attempt.MarkInDoubt(ctx, "test", "no answer"); err != nil {
		t.Fatalf("MarkInDoubt: %v", err)
	}
	if err := attempt.ResolveUnresolved(ctx, "", "neither presence nor absence proven"); err != nil {
		t.Fatalf("ResolveUnresolved: %v", err)
	}

	if !reservationState(t, j, out.Reservations[0].ID).Held() {
		t.Fatal("UNRESOLVED_IN_DOUBT is terminal but proves nothing; the hold must survive it")
	}
	alerts, err := j.ReservationsAwaitingOperator(ctx)
	if err != nil {
		t.Fatalf("ReservationsAwaitingOperator: %v", err)
	}
	if len(alerts) != 1 || alerts[0].Cause != AlertCauseUnresolved {
		t.Fatalf("want one UNRESOLVED_IN_DOUBT alert, got %+v", alerts)
	}
}

func TestDerivedTerminalFillReleasesTheHold(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()

	dec := recordEntryDecision(t, j, "d-fill", "acct-1")
	out, err := j.Reserve(ctx, exposureReserve(j, dec.ID, "acct-1", "100", "0", "10000",
		mustVersion(t, j, "acct-1")))
	if err != nil {
		t.Fatalf("Reserve: %v", err)
	}
	attempt := confirmAttempt(t, j, dec, "attempt-fill", "order-fill-1")
	_ = attempt

	res, err := j.RecordFill(ctx, FillObservation{
		OrderID: "order-fill-1", Symbol: "005930", Market: "kr",
		State: "FILLED", Terminal: true,
		Quantity: "10", FilledQuantity: "10", AveragePrice: "70000",
		ObservedAt: j.nowString(),
	})
	if err != nil {
		t.Fatalf("RecordFill: %v", err)
	}
	if len(res.ReleasedReservations) != 1 {
		t.Fatalf("a derived terminal fill must release the hold in the same transaction, got %+v",
			res.ReleasedReservations)
	}
	if reservationState(t, j, out.Reservations[0].ID).Held() {
		t.Fatal("the reservation is still held after its order filled")
	}
}

// TestUnknownBrokerStateKeepsTheReservationHeld is the "만료 추정 해제 금지"
// case the task names: CLOSED, nothing filled, no cancellation. That is what an
// expiry would look like — and the broker's status vocabulary for an expiry is
// [미측정 — 2b 2.1] — so the derivation refuses to call it terminal, the hold
// survives, and an operator is told it exists.
func TestUnknownBrokerStateKeepsTheReservationHeld(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()

	dec := recordEntryDecision(t, j, "d-unknown", "acct-1")
	out, err := j.Reserve(ctx, exposureReserve(j, dec.ID, "acct-1", "100", "0", "10000",
		mustVersion(t, j, "acct-1")))
	if err != nil {
		t.Fatalf("Reserve: %v", err)
	}
	confirmAttempt(t, j, dec, "attempt-unknown", "order-unknown-1")

	// What internal/brokerstate produces for that combination: not terminal,
	// fail-closed, with the reason naming the ambiguity.
	res, err := j.RecordFill(ctx, FillObservation{
		OrderID: "order-unknown-1", Symbol: "005930", Market: "kr",
		State: "UNKNOWN_BROKER_STATE", Terminal: false, FailClosed: true,
		Reason:   "closed_without_fill_or_cancel",
		Detail:   "closed with no fill, no cancellation timestamp and no successor",
		Quantity: "10", FilledQuantity: "0",
		ObservedAt: j.nowString(),
	})
	if err != nil {
		t.Fatalf("RecordFill: %v", err)
	}
	if len(res.ReleasedReservations) != 0 {
		t.Fatalf("an unexplained close must release nothing, got %+v", res.ReleasedReservations)
	}
	if !reservationState(t, j, out.Reservations[0].ID).Held() {
		t.Fatal("the hold was released on an assumed expiry; that is the one exit that must not exist")
	}

	// The alert is raised at the moment of the observation, not on some later
	// sweep: a ratchet nobody can see is not operable.
	if len(res.ReservationAlerts) != 1 {
		t.Fatalf("want one operator alert on the refused snapshot, got %+v", res.ReservationAlerts)
	}
	if res.ReservationAlerts[0].Reservation.ID != out.Reservations[0].ID {
		t.Fatalf("the alert names reservation %q, want %q",
			res.ReservationAlerts[0].Reservation.ID, out.Reservations[0].ID)
	}

	awaiting, err := j.ReservationsAwaitingOperator(ctx)
	if err != nil {
		t.Fatalf("ReservationsAwaitingOperator: %v", err)
	}
	if len(awaiting) != 1 || awaiting[0].Cause != AlertCauseBrokerStateUnknown {
		t.Fatalf("want one UNKNOWN_BROKER_STATE alert, got %+v", awaiting)
	}
}

// confirmAttempt drives an attempt to CONFIRMED against a broker order id, so
// the reservation has an order to be released by.
func confirmAttempt(t *testing.T, j *Journal, dec Decision, attemptID, orderID string) *Attempt {
	t.Helper()
	ctx := context.Background()
	attempt, err := j.Prepare(ctx, reservationPrepare(dec, attemptID))
	if err != nil {
		t.Fatalf("Prepare(%s): %v", attemptID, err)
	}
	if err := attempt.MarkDispatchStarted(ctx); err != nil {
		t.Fatalf("MarkDispatchStarted(%s): %v", attemptID, err)
	}
	if err := attempt.MarkAcked(ctx, orderID); err != nil {
		t.Fatalf("MarkAcked(%s): %v", attemptID, err)
	}
	if err := attempt.ResolveConfirmed(ctx, orderID, "", "acked"); err != nil {
		t.Fatalf("ResolveConfirmed(%s): %v", attemptID, err)
	}
	return attempt
}

// --- (b) expiry, only with an unspent nonce ---------------------------------

func TestExpiryReleasesOnlyWhenTheNonceWasNeverSpent(t *testing.T) {
	j, fake := openReservationJournal(t)
	ctx := context.Background()

	unsent := reserveOne(t, j, "d-unsent", "acct-1")

	sentDec := recordEntryDecision(t, j, "d-sent", "acct-1")
	sentOut, err := j.Reserve(ctx, exposureReserve(j, sentDec.ID, "acct-1", "100", "0", "10000",
		mustVersion(t, j, "acct-1")))
	if err != nil {
		t.Fatalf("Reserve: %v", err)
	}
	attempt, err := j.Prepare(ctx, reservationPrepare(sentDec, "attempt-sent"))
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	// Spending the nonce is what "a request left the process" means.
	if err := attempt.MarkDispatchStarted(ctx); err != nil {
		t.Fatalf("MarkDispatchStarted: %v", err)
	}

	fake.Advance(11 * time.Minute) // both decisions are now past their validity

	// The sweep is lazy: it runs when the ledger is next written.
	third := recordEntryDecision(t, j, "d-after", "acct-1")
	req := exposureReserve(j, third.ID, "acct-1", "1", "0", "10000", mustVersion(t, j, "acct-1"))
	req.SnapshotAsOf = j.Now()
	if _, err := j.Reserve(ctx, req); err != nil {
		t.Fatalf("Reserve after the expiries: %v", err)
	}

	if reservationState(t, j, unsent).Held() {
		t.Fatal("a decision that expired with its nonce unspent sent nothing; its hold must lapse")
	}
	if got := reservationState(t, j, unsent).ReleaseReason; got != ReleaseReasonExpiredUnconsumed {
		t.Fatalf("release reason = %q, want %s", got, ReleaseReasonExpiredUnconsumed)
	}
	if !reservationState(t, j, sentOut.Reservations[0].ID).Held() {
		t.Fatal("the nonce was spent, so a request was sent and may have been accepted; " +
			"expiry must not release that hold")
	}
}

// --- (c) the operator's exit -------------------------------------------------

// recordingAuditor is the audit sink, captured rather than written to disk.
type recordingAuditor struct {
	actions []string
	err     error
}

func (a *recordingAuditor) RecordAction(action, setting, value, detail string) error {
	if a.err != nil {
		return a.err
	}
	a.actions = append(a.actions, strings.Join([]string{action, setting, value, detail}, "|"))
	return nil
}

func TestOperatorReleaseIsTheExitForAnUnknownHold(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()
	id := reserveOne(t, j, "d-op", "acct-1")

	auditor := &recordingAuditor{}
	rel, err := j.OperatorReleaseReservation(ctx, OperatorReleaseRequest{
		ReservationID: id,
		Operator:      "sre-1",
		Reason:        "the order was cancelled at the broker",
		Evidence:      "order detail screen shows CANCELED at 09:31",
		Auditor:       auditor,
	})
	if err != nil {
		t.Fatalf("OperatorReleaseReservation: %v", err)
	}
	if rel.Reason != ReleaseReasonOperator {
		t.Fatalf("release reason = %q, want %s", rel.Reason, ReleaseReasonOperator)
	}
	if reservationState(t, j, id).Held() {
		t.Fatal("the reservation is still held after an operator released it")
	}
	if len(auditor.actions) != 1 {
		t.Fatalf("want one audit line, got %v", auditor.actions)
	}
	line := auditor.actions[0]
	for _, want := range []string{AuditActionReservationRelease, id, "sre-1",
		"cancelled at the broker", "CANCELED at 09:31"} {
		if !strings.Contains(line, want) {
			t.Errorf("audit line %q must contain %q", line, want)
		}
	}
}

func TestOperatorReleaseRefusesWithoutReasonEvidenceOrAudit(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()
	id := reserveOne(t, j, "d-op-bad", "acct-1")

	full := OperatorReleaseRequest{
		ReservationID: id, Operator: "sre-1", Reason: "why", Evidence: "what",
		Auditor: &recordingAuditor{},
	}
	missing := map[string]func(OperatorReleaseRequest) OperatorReleaseRequest{
		"operator": func(r OperatorReleaseRequest) OperatorReleaseRequest { r.Operator = " "; return r },
		"reason":   func(r OperatorReleaseRequest) OperatorReleaseRequest { r.Reason = ""; return r },
		"evidence": func(r OperatorReleaseRequest) OperatorReleaseRequest { r.Evidence = ""; return r },
		"auditor":  func(r OperatorReleaseRequest) OperatorReleaseRequest { r.Auditor = nil; return r },
	}
	for name, mutate := range missing {
		if _, err := j.OperatorReleaseReservation(ctx, mutate(full)); !errors.Is(err, ErrInvalidRequest) {
			t.Errorf("an operator release without a %s must be refused, got %v", name, err)
		}
		if !reservationState(t, j, id).Held() {
			t.Fatalf("a refused release (%s) must not have released anything", name)
		}
	}
}

// TestAFailedAuditWriteAbortsTheRelease pins the ordering: the record is
// written before the commit, so a released hold whose audit line is missing is
// not a state this path can produce.
func TestAFailedAuditWriteAbortsTheRelease(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()
	id := reserveOne(t, j, "d-op-audit", "acct-1")

	_, err := j.OperatorReleaseReservation(ctx, OperatorReleaseRequest{
		ReservationID: id, Operator: "sre-1", Reason: "why", Evidence: "what",
		Auditor: &recordingAuditor{err: errors.New("disk full")},
	})
	if err == nil {
		t.Fatal("a failed audit write must fail the release")
	}
	if !reservationState(t, j, id).Held() {
		t.Fatal("the hold was released even though its audit line could not be written")
	}
}

func TestOperatorReleaseRefusesAnAlreadyReleasedHold(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()
	id := reserveOne(t, j, "d-op-twice", "acct-1")

	req := OperatorReleaseRequest{
		ReservationID: id, Operator: "sre-1", Reason: "why", Evidence: "what",
		Auditor: &recordingAuditor{},
	}
	if _, err := j.OperatorReleaseReservation(ctx, req); err != nil {
		t.Fatalf("first release: %v", err)
	}
	if _, err := j.OperatorReleaseReservation(ctx, req); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("releasing an already-released hold must be refused, got %v", err)
	}
}

// --- (d) the trading-day boundary -------------------------------------------

// TestDailyLossReservationLapsesAtTheMarketsDayBoundary is the requirement's
// scenario: 일일 손실 예약을 보유한 채 거래일이 바뀌면 그 예약은 소멸하고 새
// 거래일의 한도가 온전히 사용 가능하다.
func TestDailyLossReservationLapsesAtTheMarketsDayBoundary(t *testing.T) {
	j, fake := openReservationJournal(t)
	ctx := context.Background()

	// 2026-03-30T00:30:00Z is 09:30 in Seoul: the Korean trading day is the 30th.
	dec := recordEntryDecision(t, j, "d-loss", "acct-1")
	req := exposureReserve(j, dec.ID, "acct-1", "500", "0", "1000", mustVersion(t, j, "acct-1"))
	req.Reservations[0].Kind = ReservationKindDailyLoss
	req.Reservations[0].TradingDay = "2026-03-30"
	req.SnapshotUsage[0].Kind = ReservationKindDailyLoss
	req.Limits[0].Kind = ReservationKindDailyLoss
	out, err := j.Reserve(ctx, req)
	if err != nil {
		t.Fatalf("Reserve: %v", err)
	}
	// Spend the decision's nonce, so the hold is one the *expiry* exit cannot
	// touch: a request was sent. The day boundary is then the only thing that
	// can lapse it, which is exactly what this test is about.
	spendNonce(t, j, dec, "attempt-loss")

	// A second 500 would reach the 1000 limit while the first is still held.
	blocked := recordEntryDecision(t, j, "d-loss-2", "acct-1")
	second := exposureReserve(j, blocked.ID, "acct-1", "500", "0", "1000", mustVersion(t, j, "acct-1"))
	second.Reservations[0].Kind = ReservationKindDailyLoss
	second.Reservations[0].TradingDay = "2026-03-30"
	second.SnapshotUsage[0].Kind = ReservationKindDailyLoss
	second.Limits[0].Kind = ReservationKindDailyLoss
	if _, err := j.Reserve(ctx, second); !errors.Is(err, ErrReservationLimitExceeded) {
		t.Fatalf("the day's loss limit must still be consumed, got %v", err)
	}

	// Move to the next Seoul trading day.
	fake.Advance(24 * time.Hour)

	third := recordEntryDecision(t, j, "d-loss-3", "acct-1")
	next := exposureReserve(j, third.ID, "acct-1", "500", "0", "1000", mustVersion(t, j, "acct-1"))
	next.SnapshotAsOf = j.Now()
	next.Reservations[0].Kind = ReservationKindDailyLoss
	next.Reservations[0].TradingDay = "2026-03-31"
	next.SnapshotUsage[0].Kind = ReservationKindDailyLoss
	next.Limits[0].Kind = ReservationKindDailyLoss
	if _, err := j.Reserve(ctx, next); err != nil {
		t.Fatalf("the new trading day's limit must be fully available: %v", err)
	}

	lapsed := reservationState(t, j, out.Reservations[0].ID)
	if lapsed.Held() {
		t.Fatal("yesterday's daily-loss hold must lapse at the day boundary")
	}
	if lapsed.ReleaseReason != ReleaseReasonDayBoundary {
		t.Fatalf("release reason = %q, want %s", lapsed.ReleaseReason, ReleaseReasonDayBoundary)
	}
}

// spendNonce drives an attempt to DISPATCH_STARTED, which is what consuming a
// decision's one-shot nonce means: a request left the process.
func spendNonce(t *testing.T, j *Journal, dec Decision, attemptID string) {
	t.Helper()
	ctx := context.Background()
	attempt, err := j.Prepare(ctx, reservationPrepare(dec, attemptID))
	if err != nil {
		t.Fatalf("Prepare(%s): %v", attemptID, err)
	}
	if err := attempt.MarkDispatchStarted(ctx); err != nil {
		t.Fatalf("MarkDispatchStarted(%s): %v", attemptID, err)
	}
}

// TestOpenExposureDoesNotLapseWithTheDay: only the daily-loss hold is a
// property of a day. An open position is still open tomorrow.
func TestOpenExposureDoesNotLapseWithTheDay(t *testing.T) {
	j, fake := openReservationJournal(t)
	ctx := context.Background()

	carried := recordEntryDecision(t, j, "d-carry", "acct-1")
	out, err := j.Reserve(ctx, exposureReserve(j, carried.ID, "acct-1", "100", "0", "10000",
		mustVersion(t, j, "acct-1")))
	if err != nil {
		t.Fatalf("Reserve: %v", err)
	}
	id := out.Reservations[0].ID
	spendNonce(t, j, carried, "attempt-carry")

	fake.Advance(24 * time.Hour)

	dec := recordEntryDecision(t, j, "d-carry-2", "acct-1")
	req := exposureReserve(j, dec.ID, "acct-1", "1", "0", "10000", mustVersion(t, j, "acct-1"))
	req.SnapshotAsOf = j.Now()
	if _, err := j.Reserve(ctx, req); err != nil {
		t.Fatalf("Reserve: %v", err)
	}
	if !reservationState(t, j, id).Held() {
		t.Fatal("an open-exposure hold does not lapse with the trading day")
	}
}
