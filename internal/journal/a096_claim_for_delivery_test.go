package journal

// a096: the outbox deduplicates the row. Until now that was all it did — a
// caller got an id back and could not tell a fresh row from one whose alert had
// already been sent, so it sent again. Every 5.6 seconds, in production, on
// 2026-08-08: one row, sixty pushes.
//
// ClaimAlertForDelivery answers the question the id cannot: is a send owed? The
// answer is read inside the transaction that resolves the id.
//
// Round 1 of the independent review killed the first answer, which was "never
// again once it landed". Permanent suppression removed the only thing that
// periodically proved the transport still worked, and it silenced a later,
// unlike occurrence on the strength of an earlier one, because an event key
// carries the condition and not its cause. The answer is now a window.

import (
	"context"
	"testing"
	"time"
)

const claimRemind = time.Hour

// TestClaimingAFreshAlertOwesDelivery: a row this call inserted has never been
// sent, so its delivery is owed.
func TestClaimingAFreshAlertOwesDelivery(t *testing.T) {
	j, _ := outboxJournal(t)
	ctx := context.Background()

	id, owed, err := j.ClaimAlertForDelivery(ctx, Alert{
		EventKey: "exit.proposal_refused|pos-1|LADDER_PARTIAL|2",
		Type:     "exit.proposal_refused",
		Severity: "critical",
	}, claimRemind)
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery: %v", err)
	}
	if id == 0 {
		t.Fatalf("id = 0, want the inserted row")
	}
	if !owed {
		t.Error("owed = false, want true — a row that was just inserted has not been sent")
	}
}

// TestClaimingADeliveredAlertInsideTheWindowOwesNothing is the storm's cure.
// The condition recurs on the next observation, seconds later, and the operator
// is not told twice.
func TestClaimingADeliveredAlertInsideTheWindowOwesNothing(t *testing.T) {
	j, clk := outboxJournal(t)
	ctx := context.Background()
	key := "exit.proposal_refused|pos-1|LADDER_PARTIAL|2"

	id, owed, err := j.ClaimAlertForDelivery(ctx, Alert{EventKey: key,
		Type: "exit.proposal_refused", Severity: "critical"}, claimRemind)
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery: %v", err)
	}
	if !owed {
		t.Fatalf("owed = false on the first claim, want true")
	}
	if err := j.MarkAlertDelivered(ctx, id); err != nil {
		t.Fatalf("MarkAlertDelivered: %v", err)
	}

	clk.Advance(claimRemind - time.Second)
	again, owed, err := j.ClaimAlertForDelivery(ctx, Alert{EventKey: key,
		Type: "exit.proposal_refused", Severity: "critical"}, claimRemind)
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery (recurrence): %v", err)
	}
	if again != id {
		t.Errorf("id = %d, want %d — the same condition is the same row", again, id)
	}
	if owed {
		t.Error("owed = true one second inside the window — the operator was just told")
	}
	if got := alertState(t, j, id); got != AlertDelivered {
		t.Errorf("state = %q, want %q — a suppressed claim must not disturb the row", got, AlertDelivered)
	}
}

// TestClaimingADeliveredAlertPastTheWindowIsReArmed is what makes the
// suppression survivable. The row goes back to PENDING, so the reminder walks
// the ordinary first-delivery path: retry budget, gate latch, escalation.
//
// Without this the engine can trade on with a dead alert channel, because the
// only thing that would have discovered the channel was dead is the send this
// row keeps declining to make.
func TestClaimingADeliveredAlertPastTheWindowIsReArmed(t *testing.T) {
	j, clk := outboxJournal(t)
	ctx := context.Background()
	key := "exit.proposal_refused|pos-1|LADDER_PARTIAL|2"

	id, _, err := j.ClaimAlertForDelivery(ctx, Alert{EventKey: key,
		Type: "exit.proposal_refused", Severity: "critical"}, claimRemind)
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery: %v", err)
	}
	if err := j.MarkAlertDelivered(ctx, id); err != nil {
		t.Fatalf("MarkAlertDelivered: %v", err)
	}

	clk.Advance(claimRemind)
	again, owed, err := j.ClaimAlertForDelivery(ctx, Alert{EventKey: key,
		Type: "exit.proposal_refused", Severity: "critical"}, claimRemind)
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery (reminder): %v", err)
	}
	if again != id {
		t.Errorf("id = %d, want %d — a reminder is the same row, not a new one", again, id)
	}
	if !owed {
		t.Fatal("owed = false after the window elapsed — the reminder never leaves")
	}
	if got := alertState(t, j, id); got != AlertPending {
		t.Errorf("state = %q, want %q — the reminder must walk the first-delivery path", got, AlertPending)
	}
	// Re-armed means the ordinary mutators work again, which is the whole point:
	// a failing reminder can latch the gate exactly as a failing original did.
	if err := j.MarkAlertAttemptFailed(ctx, id, "transport is down"); err != nil {
		t.Errorf("MarkAlertAttemptFailed on a re-armed row: %v", err)
	}
}

// TestClaimingAnAcknowledgedAlertFollowsTheSameWindow: acknowledgement is the
// operator saying they have seen something the machine could not prove was
// seen. It ends the debt for the episode, not for all time — the condition
// coming back an hour later is news again.
func TestClaimingAnAcknowledgedAlertFollowsTheSameWindow(t *testing.T) {
	j, clk := outboxJournal(t)
	ctx := context.Background()
	key := "order.unresolved_in_doubt|att-1"
	alert := Alert{EventKey: key, Type: "order.unresolved_in_doubt", Severity: "critical"}

	id, _, err := j.ClaimAlertForDelivery(ctx, alert, claimRemind)
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery: %v", err)
	}
	if err := j.AcknowledgeAlert(ctx, id, "operator"); err != nil {
		t.Fatalf("AcknowledgeAlert: %v", err)
	}

	clk.Advance(claimRemind - time.Second)
	if _, owed, err := j.ClaimAlertForDelivery(ctx, alert, claimRemind); err != nil {
		t.Fatalf("ClaimAlertForDelivery (inside window): %v", err)
	} else if owed {
		t.Error("owed = true inside the window after acknowledgement, want false")
	}

	clk.Advance(time.Second)
	if _, owed, err := j.ClaimAlertForDelivery(ctx, alert, claimRemind); err != nil {
		t.Fatalf("ClaimAlertForDelivery (past window): %v", err)
	} else if !owed {
		t.Error("owed = false past the window — a recurrence after acknowledgement is news")
	}
}

// TestClaimingAnUndeliveredAlertStillOwesDelivery is the counterweight. A row
// whose send failed is not a duplicate, it is unfinished, and the caller must
// keep being told to try — otherwise the retry budget and the entry block that
// follows a spent one are silently dropped.
func TestClaimingAnUndeliveredAlertStillOwesDelivery(t *testing.T) {
	j, _ := outboxJournal(t)
	ctx := context.Background()
	key := "engine.loop_degraded|filldetect"
	alert := Alert{EventKey: key, Type: "engine.loop_degraded", Severity: "critical"}

	id, _, err := j.ClaimAlertForDelivery(ctx, alert, claimRemind)
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery: %v", err)
	}
	if err := j.MarkAlertAttemptFailed(ctx, id, "transport is down"); err != nil {
		t.Fatalf("MarkAlertAttemptFailed: %v", err)
	}

	if _, owed, err := j.ClaimAlertForDelivery(ctx, alert, claimRemind); err != nil {
		t.Fatalf("ClaimAlertForDelivery (recurrence): %v", err)
	} else if !owed {
		t.Error("owed = false while the row is still PENDING — a failed send is unfinished, not duplicate")
	}
}

// TestAnUnrecognisedAlertStateOwesDelivery: the column has no CHECK constraint,
// so an unknown value is a row written by something this build does not
// understand. "I cannot tell whether the operator got this" reads as send.
func TestAnUnrecognisedAlertStateOwesDelivery(t *testing.T) {
	j, _ := outboxJournal(t)
	ctx := context.Background()
	key := "order.unresolved_in_doubt|att-2"
	alert := Alert{EventKey: key, Type: "order.unresolved_in_doubt", Severity: "critical"}

	id, _, err := j.ClaimAlertForDelivery(ctx, alert, claimRemind)
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery: %v", err)
	}
	if _, err := j.db.ExecContext(ctx,
		`UPDATE alert_outbox SET state = 'QUARANTINED' WHERE id = ?`, id); err != nil {
		t.Fatalf("forcing an unknown state: %v", err)
	}

	if _, owed, err := j.ClaimAlertForDelivery(ctx, alert, claimRemind); err != nil {
		t.Fatalf("ClaimAlertForDelivery: %v", err)
	} else if !owed {
		t.Error("owed = false for an unrecognised state — an unknown state is not evidence of delivery")
	}
	if got := alertState(t, j, id); got != AlertPending {
		t.Errorf("state = %q, want %q — an owed send must use the ordinary delivery path", got, AlertPending)
	}
	if err := j.MarkAlertDelivered(ctx, id); err != nil {
		t.Errorf("MarkAlertDelivered after claiming an unknown state: %v", err)
	}
}

// TestAZeroReminderWindowNeverReArms: a caller that records without delivering
// holds no reminder policy and must not re-arm a row on somebody else's behalf.
func TestAZeroReminderWindowNeverReArms(t *testing.T) {
	j, clk := outboxJournal(t)
	ctx := context.Background()
	key := "order.unresolved_in_doubt|att-3"
	alert := Alert{EventKey: key, Type: "order.unresolved_in_doubt", Severity: "critical"}

	id, _, err := j.ClaimAlertForDelivery(ctx, alert, 0)
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery: %v", err)
	}
	if err := j.MarkAlertDelivered(ctx, id); err != nil {
		t.Fatalf("MarkAlertDelivered: %v", err)
	}

	clk.Advance(30 * 24 * time.Hour)
	if _, owed, err := j.ClaimAlertForDelivery(ctx, alert, 0); err != nil {
		t.Fatalf("ClaimAlertForDelivery: %v", err)
	} else if owed {
		t.Error("owed = true with no reminder window, want false")
	}
	if got := alertState(t, j, id); got != AlertDelivered {
		t.Errorf("state = %q, want %q — a zero window must not re-arm", got, AlertDelivered)
	}
}

// TestEnqueueAlertKeepsItsContract: the recording-only caller is untouched. It
// still gets the deduplicated id and still does not have to reason about
// delivery, because it does not deliver.
func TestEnqueueAlertKeepsItsContract(t *testing.T) {
	j, _ := outboxJournal(t)
	ctx := context.Background()
	key := "order.unresolved_in_doubt|att-9"

	first, err := j.EnqueueAlert(ctx, Alert{EventKey: key,
		Type: "order.unresolved_in_doubt", Severity: "critical"})
	if err != nil {
		t.Fatalf("EnqueueAlert: %v", err)
	}
	second, err := j.EnqueueAlert(ctx, Alert{EventKey: key,
		Type: "order.unresolved_in_doubt", Severity: "critical"})
	if err != nil {
		t.Fatalf("EnqueueAlert (repeat): %v", err)
	}
	if first != second {
		t.Errorf("ids = %d, %d — the same event key is one row", first, second)
	}
}

func alertState(t *testing.T, j *Journal, id int64) string {
	t.Helper()
	var state string
	if err := j.db.QueryRowContext(context.Background(),
		`SELECT state FROM alert_outbox WHERE id = ?`, id).Scan(&state); err != nil {
		t.Fatalf("reading alert %d state: %v", id, err)
	}
	return state
}
