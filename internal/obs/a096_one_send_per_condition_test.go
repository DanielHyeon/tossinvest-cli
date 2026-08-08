package obs_test

// a096: "the same condition observed on three consecutive polls is one alert,
// not three" is what the outbox says it does. It deduplicated the row and sent
// every time.
//
// TestTheSameConditionEnqueuesOnce counts outbox rows, which the UNIQUE
// constraint on event_key already guarantees, so it passed throughout. These
// tests count sends, which is the thing the operator's phone counts.
//
// Round 1 of the independent review rejected permanent suppression, so these
// also pin what must still happen *after* the reminder window: the transport
// gets tested again, and a later occurrence of the same key is not swallowed on
// the strength of an earlier one.

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

const a096Remind = time.Hour

// a096Event is one recurring condition: the same position, the same rung, the
// same refusal, observed again because the exit loop observes again.
func a096Event() obs.Event {
	return obs.Event{
		Type:   obs.EventExitProposalRefused,
		Key:    string(obs.EventExitProposalRefused) + "|pos-1|LADDER_PARTIAL|2",
		Title:  "한화시스템(272210) 청산 주문이 제출되지 않았다",
		Fields: map[string]any{obs.FieldSymbol: "272210"},
	}
}

// a096Notifier is newNotifier plus the journal's clock, which is what drives the
// reminder window. The notifier's own clock stays real because it drives the
// retry sleeps, and advancing those from outside the call they are inside is a
// deadlock rather than a test.
func a096Notifier(t *testing.T, pub obs.Publisher) (*obs.Notifier, *journal.Journal, *execgw.EntryGate, *clock.Fake) {
	t.Helper()
	clk := clock.NewFake(obsNow)
	j := openJournal(t, clk)
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
	return &obs.Notifier{
		Log:         obs.NewLogger(obs.LogOptions{Writer: newDiscard(), JSON: true, Clock: clk}),
		Publisher:   pub,
		Journal:     j,
		Gate:        gate,
		Clock:       clock.System(),
		Attempts:    3,
		RetryDelay:  time.Millisecond,
		RemindAfter: a096Remind,
	}, j, gate, clk
}

// TestOneConditionIsOneSend is the storm's cure. On 2026-08-08 this condition
// was observed every 5.6 seconds and every observation became a push.
func TestOneConditionIsOneSend(t *testing.T) {
	pub := &failingPublisher{}
	n, j, _, _ := a096Notifier(t, pub)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		if err := n.Notify(ctx, a096Event()); err != nil {
			t.Fatalf("Notify #%d: %v", i+1, err)
		}
	}

	if got := pub.callCount(); got != 1 {
		t.Errorf("sends = %d, want 1 — the operator was told once and the condition did not change",
			got)
	}
	// The row assertion the old test made is still true and still worth keeping.
	if count, err := j.UndeliveredCount(ctx); err != nil {
		t.Fatalf("UndeliveredCount: %v", err)
	} else if count != 0 {
		t.Errorf("undelivered = %d, want 0 — the one send succeeded", count)
	}
}

// TestTheSameConditionIsRemindedOncePerWindow: suppression is a window, not a
// tombstone. A condition that will not go away is worth saying again eventually,
// and how often is bounded rather than "every observation".
func TestTheSameConditionIsRemindedOncePerWindow(t *testing.T) {
	pub := &failingPublisher{}
	n, _, _, clk := a096Notifier(t, pub)
	ctx := context.Background()

	// Two hours of a persistent condition, observed every five seconds: 1441
	// observations. Under the storm every one of them was a push.
	observations := 0
	for elapsed := time.Duration(0); elapsed <= 2*a096Remind; elapsed += 5 * time.Second {
		if err := n.Notify(ctx, a096Event()); err != nil {
			t.Fatalf("Notify at %s: %v", elapsed, err)
		}
		observations++
		clk.Advance(5 * time.Second)
	}

	// One when it started, one at each window boundary.
	if got := pub.callCount(); got != 3 {
		t.Errorf("sends = %d over %d observations, want 3 — bounded, not per-observation",
			got, observations)
	}
}

// TestADeadTransportIsStillFoundAfterASuccessfulDelivery is round 1's blocker 2.
//
// Under permanent suppression a delivered row never touched the transport
// again, so an engine could keep trading with a dead alert channel: the retry
// budget was never spent, the gate never latched, and the operating mode never
// tightened. The window is what gives that detector back.
func TestADeadTransportIsStillFoundAfterASuccessfulDelivery(t *testing.T) {
	pub := &failingPublisher{}
	n, _, gate, clk := a096Notifier(t, pub)
	ctx := context.Background()

	if err := n.Notify(ctx, a096Event()); err != nil {
		t.Fatalf("Notify (transport alive): %v", err)
	}
	if pub.callCount() != 1 {
		t.Fatalf("sends = %d, want 1", pub.callCount())
	}
	if rejected := gate.CheckEntry(); rejected != nil {
		t.Fatalf("entries blocked while delivery was working: %v", rejected)
	}

	// The transport dies after the alert landed.
	pub.setFail(true)
	clk.Advance(a096Remind)
	if err := n.Notify(ctx, a096Event()); err != nil {
		t.Fatalf("Notify (transport dead): %v", err)
	}

	if pub.callCount() <= 1 {
		t.Fatalf("sends = %d — the reminder never tested the transport", pub.callCount())
	}
	if rejected := gate.CheckEntry(); rejected == nil {
		t.Error("entries are still permitted after the reminder exhausted its retries — " +
			"a dead alert channel must stop new positions")
	}
}

// TestALaterOccurrenceOfTheSameKeyIsNotSwallowed is round 1's blocker 3.
//
// An event key carries the condition, not its cause: exit.proposal_refused is
// keyed by position, action and rung and says nothing about why the broker
// refused. A weekend "market is shut" refusal must not permanently silence a
// later refusal of the same rung for a reason that needs an operator.
func TestALaterOccurrenceOfTheSameKeyIsNotSwallowed(t *testing.T) {
	pub := &failingPublisher{}
	n, _, _, clk := a096Notifier(t, pub)
	ctx := context.Background()

	shut := a096Event()
	shut.Body = "order-hours-closed: 주문가능일이 아닙니다."
	if err := n.Notify(ctx, shut); err != nil {
		t.Fatalf("Notify (market shut): %v", err)
	}

	clk.Advance(a096Remind)
	actionable := a096Event()
	actionable.Body = "insufficient-quantity: 매도 가능 수량이 부족합니다."
	if err := n.Notify(ctx, actionable); err != nil {
		t.Fatalf("Notify (actionable cause): %v", err)
	}

	if got := pub.callCount(); got != 2 {
		t.Fatalf("sends = %d, want 2 — a later occurrence must reach the operator", got)
	}
	pub.mu.Lock()
	msgs := append([]obs.Notification(nil), pub.messages...)
	pub.mu.Unlock()
	if len(msgs) != 2 || !strings.Contains(msgs[1].Body, "insufficient-quantity") {
		t.Errorf("the second send did not carry the later cause: %+v", msgs)
	}
}

// TestConcurrentObservationsOfOneConditionSendOnce is round 1's blocker 1.
//
// Claiming outside the delivery mutex let two observations both read a row that
// had not been delivered yet, both conclude the send was owed, and both
// publish. The second then failed to mark a row that was already DELIVERED —
// the `no such alert` line the storm left behind.
func TestConcurrentObservationsOfOneConditionSendOnce(t *testing.T) {
	pub := &failingPublisher{}
	n, _, _, _ := a096Notifier(t, pub)
	ctx := context.Background()

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := n.Notify(ctx, a096Event()); err != nil {
				t.Errorf("Notify: %v", err)
			}
		}()
	}
	wg.Wait()

	if got := pub.callCount(); got != 1 {
		t.Errorf("sends = %d, want 1 — eight observations of one condition are one alert", got)
	}
}

// TestAnUndeliveredConditionIsStillRetried is the counterweight, and it must
// pass both before and after the fix. Suppressing a resend is only safe when
// the first send actually landed; a failed send is unfinished work, not a
// duplicate, and the retry budget plus its entry block are a074's contract.
func TestAnUndeliveredConditionIsStillRetried(t *testing.T) {
	pub := &failingPublisher{fail: true}
	n, j, _, _ := a096Notifier(t, pub)
	ctx := context.Background()

	if err := n.Notify(ctx, a096Event()); err != nil {
		t.Fatalf("Notify (transport down): %v", err)
	}
	spent := pub.callCount()
	if spent == 0 {
		t.Fatalf("sends = 0 during the outage, want the retry budget to have been spent")
	}
	if count, err := j.UndeliveredCount(ctx); err != nil {
		t.Fatalf("UndeliveredCount: %v", err)
	} else if count != 1 {
		t.Fatalf("undelivered = %d, want 1 — a failed send is preserved, not abandoned", count)
	}

	pub.setFail(false)
	if err := n.Notify(ctx, a096Event()); err != nil {
		t.Fatalf("Notify (transport back): %v", err)
	}

	if got := pub.callCount(); got <= spent {
		t.Errorf("sends = %d, want more than %d — a PENDING row must still be tried", got, spent)
	}
	if count, err := j.UndeliveredCount(ctx); err != nil {
		t.Fatalf("UndeliveredCount: %v", err)
	} else if count != 0 {
		t.Errorf("undelivered = %d, want 0 — the retry landed", count)
	}
}

// TestSuppressingTheSendKeepsTheRecord: what is suppressed is the transport,
// not the observation. logEvent runs in Notify ahead of the grading branch, so
// the structured line is written every time either way — that is what keeps the
// log honest about how long a condition persisted. On 2026-08-08 that was 60
// WARN lines; the operator should have received one push, not 60, and still be
// able to read all 60 lines.
func TestSuppressingTheSendKeepsTheRecord(t *testing.T) {
	var buf bytes.Buffer
	clk := clock.NewFake(obsNow)
	n := &obs.Notifier{
		Log:         obs.NewLogger(obs.LogOptions{Writer: &buf, JSON: true, Clock: clk}),
		Publisher:   &failingPublisher{},
		Journal:     openJournal(t, clk),
		Clock:       clock.System(),
		Attempts:    3,
		RetryDelay:  time.Millisecond,
		RemindAfter: a096Remind,
	}
	ctx := context.Background()

	const observations = 5
	for i := 0; i < observations; i++ {
		if err := n.Notify(ctx, a096Event()); err != nil {
			t.Fatalf("Notify #%d: %v", i+1, err)
		}
	}

	pub := n.Publisher.(*failingPublisher)
	if got := pub.callCount(); got != 1 {
		t.Errorf("sends = %d, want 1", got)
	}
	lines := strings.Count(strings.TrimSpace(buf.String()), "\n") + 1
	if lines != observations {
		t.Errorf("log lines = %d, want %d — suppressing the push must not suppress the record",
			lines, observations)
	}
}

// blockingPublisher stops inside Publish until it is released, so a test can
// hold the delivery path open and see what another goroutine is allowed to do
// meanwhile.
type blockingPublisher struct {
	entered chan struct{}
	release chan struct{}
	once    sync.Once
}

func (p *blockingPublisher) Publish(_ context.Context, _ obs.Notification) error {
	p.once.Do(func() { close(p.entered) })
	<-p.release
	return nil
}

// TestAcknowledgeCannotClearTheGateMidSend: the release is a read-then-decide —
// count what is still pending, clear the gate only if the count is zero. A096
// re-arms settled rows, so a send running alongside an acknowledgement can turn
// a settled row back into a pending one. If that lands between the count and
// the clear, the gate opens while an undelivered critical alert exists.
//
// The two must therefore exclude each other. This test pins that: while a send
// is inside Publish, Acknowledge must not have finished.
func TestAcknowledgeCannotClearTheGateMidSend(t *testing.T) {
	pub := &blockingPublisher{entered: make(chan struct{}), release: make(chan struct{})}
	n, _, _, _ := a096Notifier(t, pub)
	ctx := context.Background()

	go func() { _ = n.Notify(ctx, a096Event()) }()
	<-pub.entered // a send is in flight and holds the delivery mutex

	done := make(chan error, 1)
	go func() { done <- n.Acknowledge(ctx, "operator") }()

	select {
	case err := <-done:
		t.Fatalf("Acknowledge finished while a send was in flight (err=%v) — "+
			"it can clear the gate against a count taken before a re-arm", err)
	case <-time.After(50 * time.Millisecond):
	}

	close(pub.release)
	if err := <-done; err != nil {
		t.Fatalf("Acknowledge after the send completed: %v", err)
	}
}
