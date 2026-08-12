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
//
// a099 moved these from a bool to a disposition, because "a send is owed" and
// "this send is mine" turned out to be two questions and the bool answered only
// the first. Every assertion below is the same assertion; ClaimSettled is the
// old owed=false and ClaimAcquired is the old owed=true. The one case where the
// meaning genuinely split is called out where it happens.

import (
	"context"
	"testing"
	"time"
)

const claimRemind = time.Hour

// testClaimant names the sender in these tests. Exclusion rests on the token,
// never on this string, which is why every test can share one.
const testClaimant = "test-sender"

// TestClaimingAFreshAlertOwesDelivery: a row this call inserted has never been
// sent, so its delivery is owed.
func TestClaimingAFreshAlertOwesDelivery(t *testing.T) {
	j, _ := outboxJournal(t)
	ctx := context.Background()

	claim, err := j.ClaimAlertForDelivery(ctx, Alert{
		EventKey: "exit.proposal_refused|pos-1|LADDER_PARTIAL|2",
		Type:     "exit.proposal_refused",
		Severity: "critical",
	}, claimRemind, testClaimant)
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery: %v", err)
	}
	if claim.ID == 0 {
		t.Fatalf("id = 0, want the inserted row")
	}
	if claim.Disposition != ClaimAcquired {
		t.Errorf("disposition = %v, want acquired — a row that was just inserted has not been sent",
			claim.Disposition)
	}
	if claim.Token == "" {
		t.Error("token is empty on an acquired claim — nothing could settle this send")
	}
}

// TestClaimingADeliveredAlertInsideTheWindowOwesNothing is the storm's cure.
// The condition recurs on the next observation, seconds later, and the operator
// is not told twice.
func TestClaimingADeliveredAlertInsideTheWindowOwesNothing(t *testing.T) {
	j, clk := outboxJournal(t)
	ctx := context.Background()
	key := "exit.proposal_refused|pos-1|LADDER_PARTIAL|2"

	claim, err := j.ClaimAlertForDelivery(ctx, Alert{EventKey: key,
		Type: "exit.proposal_refused", Severity: "critical"}, claimRemind, testClaimant)
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery: %v", err)
	}
	if claim.Disposition != ClaimAcquired {
		t.Fatalf("disposition = %v on the first claim, want acquired", claim.Disposition)
	}
	if _, err := j.MarkAlertDelivered(ctx, claim.ID, claim.Token); err != nil {
		t.Fatalf("MarkAlertDelivered: %v", err)
	}

	clk.Advance(claimRemind - time.Second)
	again, err := j.ClaimAlertForDelivery(ctx, Alert{EventKey: key,
		Type: "exit.proposal_refused", Severity: "critical"}, claimRemind, testClaimant)
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery (recurrence): %v", err)
	}
	if again.ID != claim.ID {
		t.Errorf("id = %d, want %d — the same condition is the same row", again.ID, claim.ID)
	}
	if again.Disposition != ClaimSettled {
		t.Errorf("disposition = %v one second inside the window — the operator was just told",
			again.Disposition)
	}
	if got := alertState(t, j, claim.ID); got != AlertDelivered {
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

	claim, err := j.ClaimAlertForDelivery(ctx, Alert{EventKey: key,
		Type: "exit.proposal_refused", Severity: "critical"}, claimRemind, testClaimant)
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery: %v", err)
	}
	if _, err := j.MarkAlertDelivered(ctx, claim.ID, claim.Token); err != nil {
		t.Fatalf("MarkAlertDelivered: %v", err)
	}

	clk.Advance(claimRemind)
	again, err := j.ClaimAlertForDelivery(ctx, Alert{EventKey: key,
		Type: "exit.proposal_refused", Severity: "critical"}, claimRemind, testClaimant)
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery (reminder): %v", err)
	}
	if again.ID != claim.ID {
		t.Errorf("id = %d, want %d — a reminder is the same row, not a new one", again.ID, claim.ID)
	}
	if again.Disposition != ClaimAcquired {
		t.Fatalf("disposition = %v after the window elapsed — the reminder never leaves", again.Disposition)
	}
	if got := alertState(t, j, claim.ID); got != AlertPending {
		t.Errorf("state = %q, want %q — the reminder must walk the first-delivery path", got, AlertPending)
	}
	// Re-armed means the ordinary mutators work again, which is the whole point:
	// a failing reminder can latch the gate exactly as a failing original did.
	res, err := j.MarkAlertAttemptFailed(ctx, claim.ID, again.Token, "transport is down")
	if err != nil {
		t.Errorf("MarkAlertAttemptFailed on a re-armed row: %v", err)
	}
	if res.Outcome != SettleApplied {
		t.Errorf("outcome = %v, want applied — the reminder's own claim must settle", res.Outcome)
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

	claim, err := j.ClaimAlertForDelivery(ctx, alert, claimRemind, testClaimant)
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery: %v", err)
	}
	// Acknowledged while a sender still holds the lease. That is a designed
	// path, not a contrived one: the operator's release deliberately ignores
	// leases, because a person having seen the alert outranks a machine being
	// midway through telling them.
	if err := j.AcknowledgeAlert(ctx, claim.ID, "operator"); err != nil {
		t.Fatalf("AcknowledgeAlert: %v", err)
	}

	clk.Advance(claimRemind - time.Second)
	if got, err := j.ClaimAlertForDelivery(ctx, alert, claimRemind, testClaimant); err != nil {
		t.Fatalf("ClaimAlertForDelivery (inside window): %v", err)
	} else if got.Disposition != ClaimSettled {
		t.Errorf("disposition = %v inside the window after acknowledgement, want settled", got.Disposition)
	}

	clk.Advance(time.Second)
	if got, err := j.ClaimAlertForDelivery(ctx, alert, claimRemind, testClaimant); err != nil {
		t.Fatalf("ClaimAlertForDelivery (past window): %v", err)
	} else if got.Disposition != ClaimAcquired {
		t.Errorf("disposition = %v past the window — a recurrence after acknowledgement is news",
			got.Disposition)
	}
}

// TestClaimingAnUndeliveredAlertStillOwesDelivery is the counterweight. A row
// whose send failed is not a duplicate, it is unfinished, and the caller must
// keep being told to try — otherwise the retry budget and the entry block that
// follows a spent one are silently dropped.
//
// This is the one place where a099 split the old bool in two. Before the lease,
// the second claim came back owed=true and the answer stopped there. Now the row
// is still unsettled *and* it is still held: recording a failed attempt keeps the
// claim on purpose, because the sender that took it has retries left and is
// about to use them. A second sender arriving between two of those attempts must
// not send, and that is what held-elsewhere says. What has not changed is the
// thing this test exists for: a failed send is never mistaken for a delivered
// one.
func TestClaimingAnUndeliveredAlertStillOwesDelivery(t *testing.T) {
	j, _ := outboxJournal(t)
	ctx := context.Background()
	key := "engine.loop_degraded|filldetect"
	alert := Alert{EventKey: key, Type: "engine.loop_degraded", Severity: "critical"}

	claim, err := j.ClaimAlertForDelivery(ctx, alert, claimRemind, testClaimant)
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery: %v", err)
	}
	res, err := j.MarkAlertAttemptFailed(ctx, claim.ID, claim.Token, "transport is down")
	if err != nil {
		t.Fatalf("MarkAlertAttemptFailed: %v", err)
	}
	if res.Outcome != SettleApplied {
		t.Fatalf("outcome = %v, want applied", res.Outcome)
	}
	if got := alertState(t, j, claim.ID); got != AlertPending {
		t.Errorf("state = %q, want %q — a failed send is unfinished, not duplicate", got, AlertPending)
	}

	again, err := j.ClaimAlertForDelivery(ctx, alert, claimRemind, testClaimant)
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery (recurrence): %v", err)
	}
	if again.Disposition == ClaimSettled {
		t.Error("disposition = settled while the row is still PENDING — a failed send is unfinished, not duplicate")
	}
	if again.Disposition != ClaimHeldElsewhere {
		t.Errorf("disposition = %v, want held-elsewhere — the first sender still has retries left",
			again.Disposition)
	}

	// And once that sender gives up, the row is available again. Otherwise the
	// lease would be a way to lose an alert rather than a way to own one.
	if _, err := j.ReleaseAlertClaim(ctx, claim.ID, claim.Token); err != nil {
		t.Fatalf("ReleaseAlertClaim: %v", err)
	}
	if got, err := j.ClaimAlertForDelivery(ctx, alert, claimRemind, testClaimant); err != nil {
		t.Fatalf("ClaimAlertForDelivery (after release): %v", err)
	} else if got.Disposition != ClaimAcquired {
		t.Errorf("disposition = %v after the holder gave up, want acquired", got.Disposition)
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

	claim, err := j.ClaimAlertForDelivery(ctx, alert, claimRemind, testClaimant)
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery: %v", err)
	}
	if _, err := j.db.ExecContext(ctx,
		`UPDATE alert_outbox SET state = 'QUARANTINED' WHERE id = ?`, claim.ID); err != nil {
		t.Fatalf("forcing an unknown state: %v", err)
	}

	again, err := j.ClaimAlertForDelivery(ctx, alert, claimRemind, testClaimant)
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery: %v", err)
	}
	if again.Disposition != ClaimAcquired {
		t.Errorf("disposition = %v for an unrecognised state — an unknown state is not evidence of delivery",
			again.Disposition)
	}
	if got := alertState(t, j, claim.ID); got != AlertPending {
		t.Errorf("state = %q, want %q — an owed send must use the ordinary delivery path", got, AlertPending)
	}
	if _, err := j.MarkAlertDelivered(ctx, claim.ID, again.Token); err != nil {
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

	claim, err := j.ClaimAlertForDelivery(ctx, alert, 0, testClaimant)
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery: %v", err)
	}
	if _, err := j.MarkAlertDelivered(ctx, claim.ID, claim.Token); err != nil {
		t.Fatalf("MarkAlertDelivered: %v", err)
	}

	clk.Advance(30 * 24 * time.Hour)
	if got, err := j.ClaimAlertForDelivery(ctx, alert, 0, testClaimant); err != nil {
		t.Fatalf("ClaimAlertForDelivery: %v", err)
	} else if got.Disposition != ClaimSettled {
		t.Errorf("disposition = %v with no reminder window, want settled", got.Disposition)
	}
	if got := alertState(t, j, claim.ID); got != AlertDelivered {
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
