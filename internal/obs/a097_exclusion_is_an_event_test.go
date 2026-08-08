package obs_test

// a097 R4: prove exclusion with an event, not with a clock.
//
// Flush and the notify path share one mutex, and that mutex is the only thing
// stopping the two from publishing the same row at once. It has no test: a096's
// third edition removed it by mutation and the whole obs suite stayed green.
//
// A lock is not a branch, so coverage cannot see its absence — `n.mu.Lock` reads
// as covered whether or not it is needed. What can see it is a publisher that
// notices it has been entered twice at once.

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// reentrantPublisher records the largest number of Publish calls that were ever
// inside it simultaneously. The first call parks so a second one has something
// to be simultaneous with; later calls return at once, because the point is to
// observe the overlap, not to build a queue.
type reentrantPublisher struct {
	mu       sync.Mutex
	calls    int
	inFlight int
	peak     int

	entered chan struct{}
	release chan struct{}
	once    sync.Once
}

func newReentrantPublisher() *reentrantPublisher {
	return &reentrantPublisher{
		entered: make(chan struct{}),
		release: make(chan struct{}),
	}
}

func (p *reentrantPublisher) Publish(_ context.Context, _ obs.Notification) error {
	p.mu.Lock()
	p.calls++
	p.inFlight++
	if p.inFlight > p.peak {
		p.peak = p.inFlight
	}
	first := p.calls == 1
	p.mu.Unlock()

	if first {
		p.once.Do(func() { close(p.entered) })
		<-p.release
	}

	p.mu.Lock()
	p.inFlight--
	p.mu.Unlock()
	return nil
}

func (p *reentrantPublisher) stats() (calls, peak int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.calls, p.peak
}

// TestFlushCannotPublishBesideASend.
//
// While a send is inside Publish its outbox row is still PENDING —
// MarkAlertDelivered runs after Publish returns. So a Flush that does not take
// the same mutex reads that row, decides it is undelivered, and publishes it a
// second time. The operator gets the same alert twice and the second
// MarkAlertDelivered fails against an already-delivered row, which is the
// `no such alert` line the 2026-08-08 storm left behind.
//
// # What this test proves and what it does not
//
// The failure is an event: two concurrent entries into Publish actually
// happened, and no amount of scheduling luck manufactures that. The pass is
// weaker — it needs the Flush goroutine to have had the opportunity to reach
// Publish. The settle below buys that opportunity; it is not the assertion, and
// lengthening it would not make a passing run more true.
//
// That residual is why the value of this test is established by mutation
// (a097 task 5.8) rather than by argument: remove the lock, run it repeatedly,
// and count. A test whose only evidence is "it did not fail" is the thing a097
// exists to stop shipping.
func TestFlushCannotPublishBesideASend(t *testing.T) {
	pub := newReentrantPublisher()
	n, _, _, _ := a096Notifier(t, pub)
	ctx := context.Background()

	sendDone := make(chan error, 1)
	go func() { sendDone <- n.Notify(ctx, a096Event()) }()
	// Bounded, so a send that never publishes fails here with a sentence rather
	// than hanging until the package-wide test timeout kills every other test.
	select {
	case <-pub.entered: // a send holds the delivery mutex and is inside Publish
	case <-time.After(10 * time.Second):
		t.Fatal("no publish began within 10s — the send never reached the transport")
	}

	flushDone := make(chan error, 1)
	go func() {
		_, _, err := n.Flush(ctx)
		flushDone <- err
	}()

	// Give the flush goroutine time to reach Publish if nothing stops it. With
	// the mutex in place it parks on the lock instead and this simply expires.
	time.Sleep(50 * time.Millisecond)

	close(pub.release)
	if err := <-sendDone; err != nil {
		t.Fatalf("Notify: %v", err)
	}
	if err := <-flushDone; err != nil {
		t.Fatalf("Flush: %v", err)
	}

	calls, peak := pub.stats()
	if peak > 1 {
		t.Errorf("%d publishes were in flight at once (calls=%d) — the flush path and "+
			"the send path published the same row together, which is the double send "+
			"the shared mutex exists to prevent", peak, calls)
	}
	if calls != 1 {
		t.Errorf("publishes = %d, want 1 — one condition observed once is one alert, "+
			"and draining the backlog must not add a second", calls)
	}
}
