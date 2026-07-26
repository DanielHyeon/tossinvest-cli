package obs_test

// escalation_test.go is the third automatic-tightening producer (task 3.2's
// succession): sustained critical-alert delivery failure → ENTRY_BLOCKED.
//
// The gate latch already existed. What this adds is the durable half, and the
// two answer different questions: the latch stops *this* process from opening a
// position, the mode row is what the next process still knows. An operator who
// notices that alerts are not arriving is exactly the person about to restart
// something.

import (
	"context"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

const escalationAccount = "acct-7"

// newEscalatingNotifier is newNotifier with the account named, which is what
// turns the escalation on.
func newEscalatingNotifier(t *testing.T, pub obs.Publisher) (*obs.Notifier, *journal.Journal, *execgw.EntryGate) {
	t.Helper()
	clk := clock.NewFake(obsNow)
	j := openJournal(t, clk)
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
	if err := j.SetModeProjector(gate); err != nil {
		t.Fatalf("SetModeProjector: %v", err)
	}
	return &obs.Notifier{
		Log:        obs.NewLogger(obs.LogOptions{Writer: newDiscard(), JSON: true, Clock: clk}),
		Publisher:  pub,
		Journal:    j,
		Gate:       gate,
		AccountRef: escalationAccount,
		Clock:      clock.System(),
		Attempts:   3,
		RetryDelay: time.Millisecond,
	}, j, gate
}

// TestAnUndeliverableCriticalAlertTightensTheOperatingMode.
func TestAnUndeliverableCriticalAlertTightensTheOperatingMode(t *testing.T) {
	pub := &failingPublisher{fail: true}
	n, j, gate := newEscalatingNotifier(t, pub)
	ctx := context.Background()

	if err := n.Notify(ctx, obs.Event{
		Type:   obs.EventOrderUnresolved,
		Title:  "UNRESOLVED_IN_DOUBT",
		Fields: map[string]any{obs.FieldAttemptID: "a-9"},
	}); err != nil {
		t.Fatalf("Notify: %v", err)
	}

	if _, undelivered := gate.Blocks()[execgw.ReasonAlertUndelivered]; !undelivered {
		t.Fatal("the gate latch is the unchanged half and it must still be there")
	}
	snapshot, err := j.CurrentOperatingMode(ctx, escalationAccount)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Mode != journal.ModeEntryBlocked {
		t.Fatalf("mode = %q, want ENTRY_BLOCKED — a latch that a restart lifts is not the whole answer",
			snapshot.Mode)
	}
	if snapshot.Cause != journal.ModeTriggerCriticalAlertUndelivered ||
		snapshot.Actor != journal.ModeActorAuto {
		t.Errorf("row = %+v, want the enumerated trigger recorded as automatic", snapshot)
	}
	// The projection reached the gate too, so the two blocks are both there for
	// different reasons and clearing one does not clear the other.
	if _, blocked := gate.OperatingModeBlocked(); !blocked {
		t.Error("the transition did not project onto the gate")
	}
}

// TestTheEscalationDoesNotAnnounceThroughTheTransportThatJustFailed.
//
// Announcing the transition would enqueue a second critical alert into the
// outbox that just failed its whole budget and spend another budget on it. The
// publisher therefore sees exactly the three attempts the original alert was
// entitled to — and, less visibly but more importantly, this test hangs rather
// than fails if the escalation ever re-enters the delivery loop, because the
// loop's mutex is not reentrant.
func TestTheEscalationDoesNotAnnounceThroughTheTransportThatJustFailed(t *testing.T) {
	pub := &failingPublisher{fail: true}
	n, j, _ := newEscalatingNotifier(t, pub)
	ctx := context.Background()

	if err := n.Notify(ctx, obs.Event{Type: obs.EventOrderUnresolved, Title: "x"}); err != nil {
		t.Fatalf("Notify: %v", err)
	}
	if pub.callCount() != 3 {
		t.Errorf("publish attempts = %d, want the 3 the original alert was entitled to and "+
			"no second alert about the transition", pub.callCount())
	}
	// One outbox row: the operator's original alert, still PENDING and waiting
	// for the transport to come back.
	remaining, err := j.UndeliveredCount(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if remaining != 1 {
		t.Errorf("undelivered = %d, want the original alert and nothing else", remaining)
	}
}

// TestRepeatedFailuresAppendOneTransition: the trigger fires on every failed
// delivery, and the no-op rule is what keeps the history readable — the account
// is already in the mode the trigger targets, so nothing is appended and nothing
// is announced. It is also what closes the feedback loop.
func TestRepeatedFailuresAppendOneTransition(t *testing.T) {
	pub := &failingPublisher{fail: true}
	n, j, _ := newEscalatingNotifier(t, pub)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		if err := n.Notify(ctx, obs.Event{
			Type:   obs.EventOrderUnresolved,
			Fields: map[string]any{obs.FieldAttemptID: string(rune('a' + i))},
		}); err != nil {
			t.Fatalf("Notify %d: %v", i, err)
		}
	}
	history, err := j.OperatingModeHistory(ctx, escalationAccount)
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 1 {
		t.Errorf("operating-mode rows = %d, want exactly 1; a 5-second failure loop must not "+
			"write a row per pass", len(history))
	}
}

// TestADeliveredAlertChangesNoMode: the trigger is *sustained failure*, and an
// alert that went out is not one.
func TestADeliveredAlertChangesNoMode(t *testing.T) {
	pub := &failingPublisher{}
	n, j, _ := newEscalatingNotifier(t, pub)
	ctx := context.Background()

	if err := n.Notify(ctx, obs.Event{Type: obs.EventOrderInDoubt, Title: "IN_DOUBT"}); err != nil {
		t.Fatalf("Notify: %v", err)
	}
	snapshot, err := j.CurrentOperatingMode(ctx, escalationAccount)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Recorded {
		t.Errorf("a delivered alert wrote an operating-mode row: %+v", snapshot)
	}
}

// TestWithoutAnAccountTheLatchIsStillTheConsequence: the durable half is opt-in
// wiring (the Notifier has no other way to learn which account it speaks for)
// and the in-memory half is not. A profile that has not named its account must
// still stop trading.
func TestWithoutAnAccountTheLatchIsStillTheConsequence(t *testing.T) {
	pub := &failingPublisher{fail: true}
	n, j, gate := newEscalatingNotifier(t, pub)
	n.AccountRef = ""
	ctx := context.Background()

	if err := n.Notify(ctx, obs.Event{Type: obs.EventOrderUnresolved, Title: "x"}); err != nil {
		t.Fatalf("Notify: %v", err)
	}
	if _, undelivered := gate.Blocks()[execgw.ReasonAlertUndelivered]; !undelivered {
		t.Error("the gate must be latched with or without an account reference")
	}
	if snapshot, err := j.CurrentOperatingMode(ctx, escalationAccount); err != nil || snapshot.Recorded {
		t.Errorf("mode = %+v err = %v, want no row for an account nobody named", snapshot, err)
	}
}
