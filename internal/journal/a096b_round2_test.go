package journal

// a096b: round 2 of the independent review. Three passes — a cross-model
// adversarial run, a mutation-tested coverage pass and a safety pass — agreed on
// two defects in the claim path, and both are in the same place: the moment
// a096 decides the operator does not need telling.
//
// The window suppresses. Everything that can make the window lie is therefore a
// way to silence the channel that reports a failed stop-loss, and the engine
// keeps opening positions believing it warned somebody.

import (
	"context"
	"testing"
	"time"
)

// TestASettledStampInTheFutureStillOwesDelivery: a delivered_at ahead of the
// clock makes now.Sub(settled) negative, and every negative duration is less
// than the window, so the row is suppressed for the whole skew plus the window.
// Nothing publishes, so the gate never latches and the escalation never runs.
//
// The decisive argument is two lines above the defect. A row this code cannot
// date fails OPEN — "it cannot claim the window has not elapsed". A row dated in
// the future has exactly the same epistemic standing and fails CLOSED. One of
// the two is wrong and it is not the one that sends.
//
// Reachable by a backward clock correction after a delivery: an NTP step-back, a
// restored snapshot, a container whose RTC was ahead at boot.
func TestASettledStampInTheFutureStillOwesDelivery(t *testing.T) {
	j, _ := outboxJournal(t)
	ctx := context.Background()
	key := "exit.proposal_refused|pos-7|LADDER_PARTIAL|1"
	alert := Alert{EventKey: key, Type: "exit.proposal_refused", Severity: "critical"}

	id, _, err := j.ClaimAlertForDelivery(ctx, alert, claimRemind)
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery: %v", err)
	}
	if err := j.MarkAlertDelivered(ctx, id); err != nil {
		t.Fatalf("MarkAlertDelivered: %v", err)
	}

	// The clock moved backwards after the delivery landed, which reads from here
	// as a stamp two hours in the future.
	future := RFC3339(j.clk.Now().Add(2 * time.Hour))
	if _, err := j.db.ExecContext(ctx,
		`UPDATE alert_outbox SET delivered_at = ? WHERE id = ?`, future, id); err != nil {
		t.Fatalf("forcing a future settlement stamp: %v", err)
	}

	_, owed, err := j.ClaimAlertForDelivery(ctx, alert, claimRemind)
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery (skewed): %v", err)
	}
	if !owed {
		t.Error("owed = false for a settlement stamp in the future — " +
			"an unusable stamp is not evidence the operator was told")
	}
	if got := alertState(t, j, id); got != AlertPending {
		t.Errorf("state = %q, want %q — an owed send walks the ordinary delivery path",
			got, AlertPending)
	}
}

// TestReArmingClearsThePreviousAcknowledgement: the re-arm writes state and
// last_error and nothing else, so acknowledged_at and acknowledged_by survive
// into the new episode. The row then reads, simultaneously, "not delivered" and
// "acknowledged by <named operator> at <time>".
//
// acknowledged_by exists to record that a human saw this. Leaving it attached to
// an episode that human never saw is not untidy bookkeeping: an operator
// triaging a backlog after an incident skips a live undelivered critical alert
// on the strength of their own name on it.
func TestReArmingClearsThePreviousAcknowledgement(t *testing.T) {
	j, clk := outboxJournal(t)
	ctx := context.Background()
	key := "order.unresolved_in_doubt|att-41"
	alert := Alert{EventKey: key, Type: "order.unresolved_in_doubt", Severity: "critical"}

	id, _, err := j.ClaimAlertForDelivery(ctx, alert, claimRemind)
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery: %v", err)
	}
	// AcknowledgeAlert moves PENDING to ACKNOWLEDGED, which is the state the
	// window applies to alongside DELIVERED.
	if err := j.AcknowledgeAlert(ctx, id, "daniel"); err != nil {
		t.Fatalf("AcknowledgeAlert: %v", err)
	}

	clk.Advance(claimRemind)
	if _, owed, err := j.ClaimAlertForDelivery(ctx, alert, claimRemind); err != nil {
		t.Fatalf("ClaimAlertForDelivery (reminder): %v", err)
	} else if !owed {
		t.Fatal("owed = false after the window elapsed on an acknowledged row")
	}

	got, err := j.LookupAlert(ctx, id)
	if err != nil {
		t.Fatalf("LookupAlert: %v", err)
	}
	if got.State != AlertPending {
		t.Fatalf("state = %q, want %q — the reminder walks the first-delivery path",
			got.State, AlertPending)
	}
	if got.AcknowledgedBy != "" {
		t.Errorf("acknowledged_by = %q on a PENDING row — "+
			"the operator never saw this episode and must not appear to have",
			got.AcknowledgedBy)
	}
	if got.AcknowledgedAt != nil {
		t.Errorf("acknowledged_at = %v on a PENDING row — "+
			"an undelivered alert cannot carry an acknowledgement", got.AcknowledgedAt)
	}
}
