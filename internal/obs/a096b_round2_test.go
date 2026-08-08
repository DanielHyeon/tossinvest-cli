package obs_test

// a096b: round 2 of the independent review, notifier side.
//
// Two findings, and they are opposites. One is a send that happened and was not
// recorded, which leaves the row PENDING so every later observation sends again
// — the storm a096 exists to kill, restored through the one path that reports
// success. The other is that the window production actually runs on had no test
// at all: no non-test code assigns Notifier.RemindAfter, so DefaultRemindAfter
// is the only value an operator ever sees, and every existing test overrides it.

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// cancellingPublisher accepts the notification and then kills the context the
// caller is holding. The push landed; the recording that follows it will not.
type cancellingPublisher struct {
	mu     sync.Mutex
	calls  int
	cancel context.CancelFunc
}

func (p *cancellingPublisher) Publish(_ context.Context, _ obs.Notification) error {
	p.mu.Lock()
	p.calls++
	cancel := p.cancel
	p.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	return nil
}

func (p *cancellingPublisher) callCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.calls
}

// TestASendThatCannotBeRecordedLatchesTheGate: deliver logs a MarkAlertDelivered
// failure and returns true anyway, so notifyCritical sees a success. The row
// stays PENDING, which means the next observation finds the send owed and sends
// again, and the one after that, for as long as the condition lasts. That is the
// 2026-08-08 storm, reached through the success path.
//
// Reachable whenever the write fails after the push lands: a cancelled context
// during shutdown, SQLITE_BUSY past the busy timeout, a full disk.
//
// "The push went out" is not the same as "the operator is covered", because the
// system's own record says it is not. The conservative reading of a send it
// cannot account for is to latch the gate: new entries stop, exits do not.
func TestASendThatCannotBeRecordedLatchesTheGate(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pub := &cancellingPublisher{cancel: cancel}
	n, j, gate, _ := a096Notifier(t, pub)

	// The error the notifier returns is not the subject: a caller that logs and
	// moves on is exactly what production does.
	_ = n.Notify(ctx, a096Event())

	if got := pub.callCount(); got != 1 {
		t.Fatalf("sends = %d, want 1 — the transport accepted the push", got)
	}

	bg := context.Background()
	if count, err := j.UndeliveredCount(bg); err != nil {
		t.Fatalf("UndeliveredCount: %v", err)
	} else if count != 1 {
		t.Fatalf("undelivered = %d, want 1 — the mark failed, so the row is still owed", count)
	}

	if rejected := gate.CheckEntry(); rejected == nil {
		t.Error("the entry gate is open after a send that could not be recorded — " +
			"the row is PENDING and will be sent again on every observation, " +
			"and nothing tells the operator that the outbox stopped working")
	}
}

// TestTheDefaultReminderWindowIsAnHour pins the value production runs on.
//
// This one does not go red against the shipped code, and that is the finding:
// `rg RemindAfter -g '!*_test.go'` finds no assignment anywhere outside
// notifier.go's own default, so DefaultRemindAfter is the only window that has
// ever reached an operator — while every a096 test sets RemindAfter explicitly
// and therefore never touches it. Mutating `return DefaultRemindAfter` to
// `return 0` — permanent suppression, the design round 1 rejected — left the
// whole obs suite green. This test is what makes that mutation die.
func TestTheDefaultReminderWindowIsAnHour(t *testing.T) {
	pub := &failingPublisher{}
	clk := clock.NewFake(obsNow)
	j := openJournal(t, clk)
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
	// Constructed the way production constructs it: RemindAfter unset.
	n := &obs.Notifier{
		Log:        obs.NewLogger(obs.LogOptions{Writer: newDiscard(), JSON: true, Clock: clk}),
		Publisher:  pub,
		Journal:    j,
		Gate:       gate,
		Clock:      clock.System(),
		Attempts:   3,
		RetryDelay: time.Millisecond,
	}
	ctx := context.Background()

	if err := n.Notify(ctx, a096Event()); err != nil {
		t.Fatalf("Notify: %v", err)
	}
	if got := pub.callCount(); got != 1 {
		t.Fatalf("sends = %d, want 1 — the first observation is always owed", got)
	}

	clk.Advance(obs.DefaultRemindAfter - time.Second)
	if err := n.Notify(ctx, a096Event()); err != nil {
		t.Fatalf("Notify (inside the default window): %v", err)
	}
	if got := pub.callCount(); got != 1 {
		t.Errorf("sends = %d one second inside DefaultRemindAfter, want 1 — "+
			"the default window suppresses", got)
	}

	clk.Advance(2 * time.Second)
	if err := n.Notify(ctx, a096Event()); err != nil {
		t.Fatalf("Notify (past the default window): %v", err)
	}
	if got := pub.callCount(); got != 2 {
		t.Errorf("sends = %d past DefaultRemindAfter, want 2 — "+
			"the default window reminds, it does not tombstone", got)
	}
}
