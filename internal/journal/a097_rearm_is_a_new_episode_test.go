package journal

// a097: a re-armed row is a new episode, and every column of it has to say so.
//
// a096 re-armed the state. a096b added the acknowledgement, because a row that
// reads "not delivered" and "acknowledged by daniel" at the same time makes an
// operator skip a live critical alert on the strength of their own name.
//
// The same argument reaches four more columns and stopped there. A row whose
// body describes an episode that is over, whose attempt count belongs to that
// episode, and whose delivery timestamp claims the body it is no longer holding
// was delivered, is not evidence. It is two episodes wearing one row.

import (
	"context"
	"testing"
	"time"
)

// a097Alert is one observation of a condition. The key is the condition; the
// title, body and payload are the cause, and the cause is what changes between
// episodes of the same key.
func a097Alert(cause string) Alert {
	return Alert{
		EventKey: "exit.proposal_refused|pos-7|LADDER_PARTIAL|1",
		Type:     "exit.proposal_refused",
		Severity: "critical",
		Title:    "protective exit refused: " + cause,
		Body:     "the broker refused the exit order: " + cause,
		Payload:  `{"reason":"` + cause + `"}`,
	}
}

// TestReArmingCarriesTheCauseThatReArmedIt: the event key holds the condition
// and not its cause — `exit.proposal_refused` is keyed by position, action and
// rung, never by the refusal. So the second episode of one key is routinely a
// *different* reason, and the re-arm has to carry it.
//
// Today the re-arm writes state, last_error and the acknowledgement columns and
// nothing else, so the row keeps the first episode's text. The backlog-draining
// path builds what it sends from the row (notifier.go:409-414), not from the
// live event, and an operator reading the outbox after an incident reads the
// row too.
func TestReArmingCarriesTheCauseThatReArmedIt(t *testing.T) {
	j, clk := outboxJournal(t)
	ctx := context.Background()

	first := a097Alert("order-hours-closed")
	claim, err := j.ClaimAlertForDelivery(ctx, first, claimRemind, testClaimant)
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery: %v", err)
	}
	id := claim.ID
	if _, err := j.MarkAlertDelivered(ctx, id, claim.Token); err != nil {
		t.Fatalf("MarkAlertDelivered: %v", err)
	}

	clk.Advance(claimRemind)
	second := a097Alert("insufficient-quantity")
	if again, err := j.ClaimAlertForDelivery(ctx, second, claimRemind, testClaimant); err != nil {
		t.Fatalf("ClaimAlertForDelivery (reminder): %v", err)
	} else if again.Disposition != ClaimAcquired {
		t.Fatalf("disposition = %v after the window elapsed", again.Disposition)
	}

	got, err := j.LookupAlert(ctx, id)
	if err != nil {
		t.Fatalf("LookupAlert: %v", err)
	}
	if got.Title != second.Title {
		t.Errorf("title = %q, want %q — the row a caller is about to send must "+
			"describe the observation that re-armed it", got.Title, second.Title)
	}
	if got.Body != second.Body {
		t.Errorf("body = %q, want %q", got.Body, second.Body)
	}
	if got.Payload != second.Payload {
		t.Errorf("payload = %q, want %q", got.Payload, second.Payload)
	}
}

// TestReArmingResetsTheAttemptCount: the re-arm already rewrites state,
// last_error and the acknowledgement for this episode. Leaving attempts alone
// makes it the one column still counting the previous one, and an operator
// reading `state=PENDING, attempts=7` reads "this alert has failed seven times".
//
// Safe because nothing decides on this column: deliver's retry budget is the
// in-memory n.Attempts (notifier.go:306), PendingAlerts filters on state and
// orders by id, and no production SQL filters or orders by attempts.
func TestReArmingResetsTheAttemptCount(t *testing.T) {
	j, clk := outboxJournal(t)
	ctx := context.Background()

	alert := a097Alert("order-hours-closed")
	claim, err := j.ClaimAlertForDelivery(ctx, alert, claimRemind, testClaimant)
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery: %v", err)
	}
	id := claim.ID
	// Two failed sends and then a successful one: attempts = 3 for episode one.
	// All three settle under the one claim the sender took, which is what a
	// sender working through its retry budget actually does.
	for i := 0; i < 2; i++ {
		if _, err := j.MarkAlertAttemptFailed(ctx, id, claim.Token, "transport down"); err != nil {
			t.Fatalf("MarkAlertAttemptFailed: %v", err)
		}
	}
	if _, err := j.MarkAlertDelivered(ctx, id, claim.Token); err != nil {
		t.Fatalf("MarkAlertDelivered: %v", err)
	}
	if before, err := j.LookupAlert(ctx, id); err != nil {
		t.Fatalf("LookupAlert: %v", err)
	} else if before.Attempts == 0 {
		t.Fatalf("attempts = 0 before the re-arm, so this test would pass vacuously")
	}

	clk.Advance(claimRemind)
	if again, err := j.ClaimAlertForDelivery(ctx, alert, claimRemind, testClaimant); err != nil {
		t.Fatalf("ClaimAlertForDelivery (reminder): %v", err)
	} else if again.Disposition != ClaimAcquired {
		t.Fatalf("disposition = %v after the window elapsed", again.Disposition)
	}

	got, err := j.LookupAlert(ctx, id)
	if err != nil {
		t.Fatalf("LookupAlert: %v", err)
	}
	if got.Attempts != 0 {
		t.Errorf("attempts = %d on a re-armed row, want 0 — this episode has not "+
			"been attempted, and a non-zero count says it has", got.Attempts)
	}
}

// TestReArmingClearsThePreviousEpisodeTimestamps: this is the one the first
// draft of a097 got wrong. It kept delivered_at, calling it "the record that the
// previous episode really was delivered".
//
// It cannot be. The re-arm overwrites the body, so a surviving delivered_at
// makes the row say "the content I am holding was delivered at T1" — false, and
// with no episode history anywhere to recover the content that actually was.
// last_attempt_at is the same shape: attempts=0 beside last_attempt_at=T1 reads
// "never attempted, last attempted at T1".
//
// A row is evidence only when every column points at one event. Past episodes
// live in the structured log; the outbox row is a delivery task.
//
// Nothing reads these two for a decision on a PENDING row: latestStamp is
// reached only from the settled case (claimOwed B3@255), and a re-armed row is
// PENDING, so it returns from B2@252 without dating anything.
func TestReArmingClearsThePreviousEpisodeTimestamps(t *testing.T) {
	j, clk := outboxJournal(t)
	ctx := context.Background()

	alert := a097Alert("order-hours-closed")
	claim, err := j.ClaimAlertForDelivery(ctx, alert, claimRemind, testClaimant)
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery: %v", err)
	}
	id := claim.ID
	if _, err := j.MarkAlertDelivered(ctx, id, claim.Token); err != nil {
		t.Fatalf("MarkAlertDelivered: %v", err)
	}
	if before, err := j.LookupAlert(ctx, id); err != nil {
		t.Fatalf("LookupAlert: %v", err)
	} else if before.DeliveredAt == nil || before.LastAttemptAt == nil {
		t.Fatalf("delivered_at or last_attempt_at unset after delivery, so this "+
			"test would pass vacuously (delivered=%v attempt=%v)",
			before.DeliveredAt, before.LastAttemptAt)
	}

	clk.Advance(claimRemind)
	if again, err := j.ClaimAlertForDelivery(ctx, a097Alert("insufficient-quantity"),
		claimRemind, testClaimant); err != nil {
		t.Fatalf("ClaimAlertForDelivery (reminder): %v", err)
	} else if again.Disposition != ClaimAcquired {
		t.Fatalf("disposition = %v after the window elapsed", again.Disposition)
	}

	got, err := j.LookupAlert(ctx, id)
	if err != nil {
		t.Fatalf("LookupAlert: %v", err)
	}
	if got.State != AlertPending {
		t.Fatalf("state = %q, want %q", got.State, AlertPending)
	}
	if got.DeliveredAt != nil {
		t.Errorf("delivered_at = %v on a re-armed row holding new content — "+
			"the row would claim this content was delivered then, and it was not",
			got.DeliveredAt)
	}
	if got.LastAttemptAt != nil {
		t.Errorf("last_attempt_at = %v beside attempts = %d — a row cannot have "+
			"never been attempted and last been attempted at a time",
			got.LastAttemptAt, got.Attempts)
	}
}

// TestRecoveringAnUnknownStateIgnoresTheReminderWindow pins R3's behaviour
// rather than changing it. It arrives green.
//
// The independent review read "the default branch ignores remindAfter <= 0" as
// a defect. The AST says otherwise: the remindAfter check is B4@256, inside
// B3@255 (`case AlertDelivered, AlertAcknowledged`), and `default` is B8@284 —
// its sibling. A sibling cannot guard a sibling, so these are two rules, not one
// rule with a hole:
//
//	timed reminder — off when remindAfter <= 0;
//	state repair   — independent of the window.
//
// Taking repair away from a record-only caller would leave the row outside
// PENDING, so MarkAlertDelivered could not move it and every later observation
// would publish again. That is the storm a096 exists to stop. The comment saying
// so can be deleted; this test cannot.
func TestRecoveringAnUnknownStateIgnoresTheReminderWindow(t *testing.T) {
	j, _ := outboxJournal(t)
	ctx := context.Background()

	alert := a097Alert("order-hours-closed")
	claim, err := j.ClaimAlertForDelivery(ctx, alert, claimRemind, testClaimant)
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery: %v", err)
	}
	id := claim.ID
	if _, err := j.db.ExecContext(ctx,
		`UPDATE alert_outbox SET state = 'QUARANTINED' WHERE id = ?`, id); err != nil {
		t.Fatalf("forcing an unknown state: %v", err)
	}

	// remindAfter = 0 is what EnqueueAlert passes: a caller that records without
	// delivering and therefore holds no reminder policy.
	again, err := j.ClaimAlertForDelivery(ctx, alert, 0, testClaimant)
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery (record-only): %v", err)
	}
	if again.Disposition != ClaimAcquired {
		t.Errorf("disposition = %v for an unrecognised state under remindAfter = 0 — "+
			"an unknown state is not evidence of delivery whatever the window is", again.Disposition)
	}
	if got := alertState(t, j, id); got != AlertPending {
		t.Errorf("state = %q, want %q — without the repair MarkAlertDelivered "+
			"cannot move this row and every observation publishes it again", got, AlertPending)
	}
}

// TestARecordOnlyCallerDoesNotReArmADeliveredRow is the other half of R3, and
// it also arrives green. The two rules must not contaminate each other: repair
// ignoring the window does not license a record-only caller to restart the
// delivery of a row that was genuinely delivered.
func TestARecordOnlyCallerDoesNotReArmADeliveredRow(t *testing.T) {
	j, clk := outboxJournal(t)
	ctx := context.Background()

	alert := a097Alert("order-hours-closed")
	claim, err := j.ClaimAlertForDelivery(ctx, alert, claimRemind, testClaimant)
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery: %v", err)
	}
	id := claim.ID
	if _, err := j.MarkAlertDelivered(ctx, id, claim.Token); err != nil {
		t.Fatalf("MarkAlertDelivered: %v", err)
	}

	// Well past any window, to show it is the zero and not the elapsed time that
	// decides.
	clk.Advance(24 * time.Hour)
	again, err := j.ClaimAlertForDelivery(ctx, alert, 0, testClaimant)
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery (record-only): %v", err)
	}
	if again.Disposition != ClaimSettled {
		t.Errorf("disposition = %v under remindAfter = 0 on a delivered row — "+
			"a caller that does not deliver holds no reminder policy", again.Disposition)
	}
	if got := alertState(t, j, id); got != AlertDelivered {
		t.Errorf("state = %q, want %q — recording must not restart a settled delivery",
			got, AlertDelivered)
	}
}
