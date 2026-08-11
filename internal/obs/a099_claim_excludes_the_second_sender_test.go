package obs_test

// a099 R2 · R3 · R4: exclusion has to outlive the process that holds it.
//
// a096 put the claim and the send under one mutex, and that closed the hole it
// was aimed at: two observations inside one Notifier can no longer both publish.
// The mutex is a field on a Notifier value. Two Notifier values have two
// mutexes and exclude nothing between them — and a098 adds a second sender by
// design, so this stops being a theoretical second value.
//
// All three tests below are RED today, and they are RED for one reason:
// claimAndDeliver publishes whenever the outbox says a send is owed
// (notifier.go:276), and the outbox says that to everyone who asks.
//
//   R2 — the sender that loses the claim does not publish. Today nobody loses.
//   R3 — the exclusion holds between senders with different mutexes.
//   R4 — Flush goes through the claim too, so it skips a row someone holds.
//
// Overlap is forced with a gated publisher rather than left to the scheduler.
// A race that only sometimes overlaps is a test that only sometimes fails, and
// a RED that passes on a good day is not an observation.

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

const (
	a099Remind = time.Hour
	a099Wait   = 5 * time.Second
)

// a099Event is one recurring critical condition — the same one the storm was
// made of, and the same one 272210 is producing right now.
func a099Event() obs.Event {
	return obs.Event{
		Type:   obs.EventExitProposalRefused,
		Key:    string(obs.EventExitProposalRefused) + "|pos-1|LADDER_PARTIAL|2",
		Title:  "한화시스템(272210) 청산 주문이 제출되지 않았다",
		Fields: map[string]any{obs.FieldSymbol: "272210"},
	}
}

// a099Pair builds two notifiers over one journal: two senders, two mutexes, one
// outbox. That is the shape a098 puts into production.
func a099Pair(t *testing.T, pub obs.Publisher) (*obs.Notifier, *obs.Notifier, *journal.Journal) {
	t.Helper()
	clk := clock.NewFake(obsNow)
	j := openJournal(t, clk)
	build := func() *obs.Notifier {
		return &obs.Notifier{
			Log:         obs.NewLogger(obs.LogOptions{Writer: newDiscard(), JSON: true, Clock: clk}),
			Publisher:   pub,
			Journal:     j,
			Gate:        execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{}),
			Clock:       clock.System(),
			Attempts:    1,
			RetryDelay:  time.Millisecond,
			RemindAfter: a099Remind,
		}
	}
	return build(), build(), j
}

// a099Publisher counts sends and can hold the first one open, which is how a
// second sender is made to arrive while the first is still inside Publish.
type a099Publisher struct {
	mu      sync.Mutex
	calls   int
	hold    bool
	entered chan struct{}
	release chan struct{}
	once    sync.Once
}

func newA099Publisher(hold bool) *a099Publisher {
	return &a099Publisher{
		hold:    hold,
		entered: make(chan struct{}),
		release: make(chan struct{}),
	}
}

func (p *a099Publisher) Publish(_ context.Context, _ obs.Notification) error {
	p.mu.Lock()
	p.calls++
	first := p.calls == 1
	p.mu.Unlock()
	if p.hold && first {
		close(p.entered)
		<-p.release
	}
	return nil
}

func (p *a099Publisher) count() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.calls
}

func (p *a099Publisher) releaseAll() { p.once.Do(func() { close(p.release) }) }

// waitInside blocks until a sender is inside Publish, so the caller knows the
// row is held.
func (p *a099Publisher) waitInside(t *testing.T) {
	t.Helper()
	select {
	case <-p.entered:
	case <-time.After(a099Wait):
		p.releaseAll()
		t.Fatal("no sender reached Publish")
	}
}

// TestTheSenderThatLosesTheClaimDoesNotPublish is R2.
func TestTheSenderThatLosesTheClaimDoesNotPublish(t *testing.T) {
	pub := newA099Publisher(true)
	defer pub.releaseAll()
	holder, latecomer, _ := a099Pair(t, pub)
	ctx := context.Background()

	done := make(chan error, 1)
	go func() { done <- holder.Notify(ctx, a099Event()) }()
	pub.waitInside(t)

	// The condition is still true, so it is observed again — by a sender whose
	// mutex the holder is not holding.
	if err := latecomer.Notify(ctx, a099Event()); err != nil {
		t.Fatalf("the second sender's Notify: %v", err)
	}
	sends := pub.count()

	pub.releaseAll()
	if err := <-done; err != nil {
		t.Fatalf("the holder's Notify: %v", err)
	}

	if sends != 1 {
		t.Errorf("publishes while one sender held the row = %d, want 1 — "+
			"the second sender lost nothing, because today there is nothing to lose",
			sends)
	}
}

// TestExclusionHoldsBetweenSendersWithDifferentMutexes is R3. Run under -race:
// two senders share one journal and nothing else.
func TestExclusionHoldsBetweenSendersWithDifferentMutexes(t *testing.T) {
	pub := newA099Publisher(true)
	defer pub.releaseAll()
	first, second, j := a099Pair(t, pub)
	ctx := context.Background()

	firstDone := make(chan error, 1)
	secondDone := make(chan error, 1)
	go func() { firstDone <- first.Notify(ctx, a099Event()) }()
	go func() {
		// Arrive while the other is inside Publish. Without this the two may
		// serialise and the second would find the row already delivered — a pass
		// that proves nothing.
		<-pub.entered
		secondDone <- second.Notify(ctx, a099Event())
	}()

	pub.waitInside(t)
	if err := <-secondDone; err != nil {
		t.Fatalf("the second sender's Notify: %v", err)
	}
	sends := pub.count()

	pub.releaseAll()
	if err := <-firstDone; err != nil {
		t.Fatalf("the first sender's Notify: %v", err)
	}

	if sends != 1 {
		t.Errorf("sends = %d, want 1 — n.mu excludes senders inside one Notifier "+
			"and there are two Notifiers here", sends)
	}
	// One condition, one row: the dedup a096 built is not what is broken.
	if count, err := j.UndeliveredCount(ctx); err != nil {
		t.Fatalf("UndeliveredCount: %v", err)
	} else if count > 1 {
		t.Errorf("undelivered rows = %d, want at most 1 — one condition is one row", count)
	}
}

// TestFlushDoesNotPublishARowAnotherSenderHolds is R4. Flush is the other way
// into the same rows, and a098 turns it into a loop that runs on a timer.
func TestFlushDoesNotPublishARowAnotherSenderHolds(t *testing.T) {
	pub := newA099Publisher(true)
	defer pub.releaseAll()
	holder, flusher, _ := a099Pair(t, pub)
	ctx := context.Background()

	done := make(chan error, 1)
	go func() { done <- holder.Notify(ctx, a099Event()) }()
	pub.waitInside(t)

	// Today Flush reads PendingAlerts and publishes what it finds. The held row
	// is still PENDING — it has to be, because the send has not been settled —
	// so Flush finds it.
	if _, _, err := flusher.Flush(ctx); err != nil {
		t.Logf("Flush: %v", err)
	}
	sends := pub.count()

	pub.releaseAll()
	if err := <-done; err != nil {
		t.Fatalf("the holder's Notify: %v", err)
	}

	if sends != 1 {
		t.Errorf("publishes = %d, want 1 — Flush sent a row another sender was "+
			"already sending", sends)
	}
}
