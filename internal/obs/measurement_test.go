package obs_test

import (
	"context"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// measurement_test.go is change add-net-rr-measurement tasks 4.6 and 4.7.
//
// # What this file guards, and why the existing grading test was not enough
//
// TestGradingMatchesTheSpec asserts that every event engine-safety calls critical
// *is* in the grading table. That is the right test for those events and it is
// blind to the failure this change can cause: an event added to the table by
// mistake. The consequence of that mistake is not a noisy alert — it is a live
// account that stops taking entries and needs a human to unblock it, because a
// critical alert goes to the durable outbox, and an outbox entry that cannot be
// delivered latches the gate with ReasonAlertUndelivered and escalates the
// operating mode to ENTRY_BLOCKED.
//
// Membership in criticalEvents is the whole of that switch. So these tests assert
// non-membership, for the *class* rather than for one constant, so a second
// measurement event added later is covered without anybody remembering this file.

// TestMeasurementEventsAreNeverCritical is the rule, applied to the whole subject.
func TestMeasurementEventsAreNeverCritical(t *testing.T) {
	for _, e := range []obs.EventType{
		obs.EventEntryObservationFailed,
		obs.EventEntryObservationReconstructed,
	} {
		if got := e.Subject(); got != "measurement" {
			t.Errorf("%s has subject %q; the class rule below is keyed on the subject, so an "+
				"event outside it would not be covered", e, got)
		}
		if got := obs.SeverityOf(e); got != obs.SeverityNormal {
			t.Errorf("SeverityOf(%s) = %s, want normal. A critical measurement event reaches "+
				"the durable outbox, and an undelivered outbox entry blocks new entries until "+
				"an operator clears it — a lost analysis row must not do that", e, got)
		}
	}

	// The class rule. This is what catches the event somebody adds next year.
	for _, e := range obs.CriticalEvents() {
		if e.Subject() == "measurement" {
			t.Errorf("%s is in the critical grading table. risk-management: 분석·성과 작업 "+
				"실패는 트리거가 아니다 (SHALL NOT) — a measurement fault must not be able to "+
				"stop a live account trading", e)
		}
	}
}

// TestNoThirdSeverityWasInvented is round 3's R3-10. The round-2 disposition
// described the observation-failure grade as "비강화 등급", a non-escalating tier —
// and obs has exactly two grades. Inventing a third would be a grade every
// existing consumer's switch falls through.
func TestNoThirdSeverityWasInvented(t *testing.T) {
	for _, e := range []obs.EventType{
		obs.EventEntryObservationFailed,
		obs.EventEntryObservationReconstructed,
		obs.EventOrderInDoubt,
		obs.EventFillObserved,
	} {
		switch got := obs.SeverityOf(e); got {
		case obs.SeverityCritical, obs.SeverityNormal:
		default:
			t.Errorf("SeverityOf(%s) = %q, which is neither of the two grades this build has", e, got)
		}
	}
}

// TestObservationFailureAlertNeverReachesTheGateOrTheMode is task 4.7, run through
// the same harness the critical path is tested with.
//
// The comparison is the point. newEscalatingNotifier is the wiring that turns an
// undelivered critical alert into ENTRY_BLOCKED — a real journal, a real entry
// gate, a real mode projector — and
// TestAnUndeliverableCriticalAlertTightensTheOperatingMode shows it doing exactly
// that. Same harness, same dead transport, same exhausted retries, measurement
// event: nothing moves.
func TestObservationFailureAlertNeverReachesTheGateOrTheMode(t *testing.T) {
	pub := &failingPublisher{fail: true}
	n, j, gate := newEscalatingNotifier(t, pub)
	ctx := context.Background()

	err := n.Notify(ctx, obs.Event{
		Type:  obs.EventEntryObservationFailed,
		Key:   string(obs.EventEntryObservationFailed) + "|decision-1",
		Title: "an entry observation could not be written",
		Body:  "the verdict stands; the measurement of it was lost",
		Fields: map[string]any{
			obs.FieldAccount: escalationAccount,
			"decision_id":    "decision-1",
		},
	})
	if err != nil {
		t.Fatalf("a normal-grade alert must not report its delivery failure to the caller: %v", err)
	}
	if pub.callCount() == 0 {
		t.Error("the alert must still be attempted: the requirement is 'not critical', " +
			"not 'not sent' — 조용히 넘어가서도 안 된다")
	}

	// (a) The in-process latch.
	if blocks := gate.Blocks(); len(blocks) != 0 {
		t.Fatalf("the entry gate is blocked with %v. A lost analysis row must never require "+
			"an operator to unblock trading", blocks)
	}

	// (b) The durable half. Still whatever it was, with no automatic tightening.
	snapshot, err := j.CurrentOperatingMode(ctx, escalationAccount)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Mode == journal.ModeEntryBlocked || snapshot.Mode == journal.ModeHaltAll {
		t.Fatalf("operating mode = %q after a measurement failure; the mode must be untouched",
			snapshot.Mode)
	}
	if snapshot.Cause == journal.ModeTriggerCriticalAlertUndelivered {
		t.Errorf("the measurement failure borrowed the %s trigger",
			journal.ModeTriggerCriticalAlertUndelivered)
	}

	// (c) The outbox. A critical event's first act is to make the alert durable,
	// and an undelivered row there is what the escalation reads. There must be
	// nothing.
	undelivered, err := j.UndeliveredCount(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if undelivered != 0 {
		t.Errorf("undelivered outbox entries = %d, want 0. UndeliveredCount is global — one "+
			"measurement alert sitting in it would also stop an operator clearing a genuine "+
			"IN_DOUBT latch", undelivered)
	}
}

// TestACriticalAlertStillEscalatesThroughTheSameNotifier is the control. Without
// it the test above could pass because the harness was inert rather than because
// the grading worked.
func TestACriticalAlertStillEscalatesThroughTheSameNotifier(t *testing.T) {
	pub := &failingPublisher{fail: true}
	n, _, gate := newEscalatingNotifier(t, pub)

	if err := n.Notify(context.Background(), obs.Event{
		Type:   obs.EventOrderUnresolved,
		Title:  "UNRESOLVED_IN_DOUBT",
		Fields: map[string]any{obs.FieldAttemptID: "a-9"},
	}); err != nil {
		t.Fatalf("Notify: %v", err)
	}
	if _, blocked := gate.Blocks()[execgw.ReasonAlertUndelivered]; !blocked {
		t.Fatal("the harness must still escalate a genuine critical alert, or the " +
			"measurement test above proves nothing")
	}
}

// TestObservationFailureIsLoggedEvenWhenUndeliverable: the requirement has two
// halves and the second is the easy one to lose. Not escalating is one; not being
// silent is the other, and with the transport down the structured log is what is
// left.
func TestObservationFailureIsLoggedEvenWhenUndeliverable(t *testing.T) {
	var logged strings.Builder
	notifier := &obs.Notifier{
		Publisher: &failingPublisher{fail: true},
		Log:       obs.NewLogger(obs.LogOptions{Writer: &logged}),
	}
	if err := notifier.Notify(context.Background(), obs.Event{
		Type:   obs.EventEntryObservationFailed,
		Title:  "an entry observation could not be written",
		Fields: map[string]any{"decision_id": "decision-1"},
	}); err != nil {
		t.Fatal(err)
	}
	out := logged.String()
	if !strings.Contains(out, string(obs.EventEntryObservationFailed)) {
		t.Errorf("the loss must be countable from the log stream:\n%s", out)
	}
	if !strings.Contains(out, "decision-1") {
		t.Errorf("the log line must carry which verdict lost its measurement:\n%s", out)
	}
}
