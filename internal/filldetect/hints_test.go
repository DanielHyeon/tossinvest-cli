package filldetect_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/filldetect"
	"github.com/JungHoonGhae/tossinvest-cli/internal/push"
	"github.com/JungHoonGhae/tossinvest-cli/internal/session"
)

// SSE hint tests (harden-execution-base task 3.3).
//
// The requirement is fill-detection's "SSE는 지연 단축 힌트": an event may make
// the next poll happen sooner and may do nothing else. Its payload is never
// evidence, bursts collapse into one in-flight re-fetch per topic, and a minimum
// interval keeps a chatty channel from spending the API budget.

// blockingRefresh is a Refresh function a test can hold open, which is what makes
// "while one refresh is in flight" a deterministic state rather than a race.
type blockingRefresh struct {
	mu      sync.Mutex
	calls   int
	topics  []filldetect.HintTopic
	release chan struct{}
	entered chan struct{}
}

func newBlockingRefresh() *blockingRefresh {
	return &blockingRefresh{
		release: make(chan struct{}),
		entered: make(chan struct{}, 64),
	}
}

func (r *blockingRefresh) fn(ctx context.Context, topic filldetect.HintTopic) error {
	r.mu.Lock()
	r.calls++
	r.topics = append(r.topics, topic)
	r.mu.Unlock()
	select {
	case r.entered <- struct{}{}:
	default:
	}
	if r.release == nil {
		return nil
	}
	select {
	case <-r.release:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *blockingRefresh) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.calls
}

func (r *blockingRefresh) seen() []filldetect.HintTopic {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]filldetect.HintTopic(nil), r.topics...)
}

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(time.Millisecond)
	}
}

func event(eventType string) push.Event {
	return push.Event{Type: eventType, Received: time.Now().UTC()}
}

// TestEventBurstCollapsesIntoOneRefetch is the spec's "이벤트 폭주" scenario:
// five events arriving while one re-fetch is running produce exactly one more,
// not five.
func TestEventBurstCollapsesIntoOneRefetch(t *testing.T) {
	clk := clock.NewFake(pollStart)
	refresh := newBlockingRefresh()
	h := &filldetect.Hints{
		Refresh:     refresh.fn,
		Clock:       clk,
		MinInterval: -1, // the interval has its own test
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- h.Run(ctx) }()

	h.Observe(event("pending-order-refresh"))
	<-refresh.entered // one re-fetch is now in flight and blocked

	for i := 0; i < 5; i++ {
		h.Observe(event("pending-order-refresh"))
	}
	close(refresh.release)

	waitFor(t, "the queued re-fetch", func() bool { return refresh.count() >= 2 })
	// Give any surplus a chance to appear before asserting there is none.
	time.Sleep(20 * time.Millisecond)
	if got := refresh.count(); got != 2 {
		t.Fatalf("re-fetches = %d, want 2 (the in-flight one plus a single coalesced follow-up)", got)
	}

	stats := h.Stats()
	if stats.Coalesced < 4 {
		t.Fatalf("coalesced = %d, want at least 4 of the 5 burst events merged", stats.Coalesced)
	}

	cancel()
	if err := <-done; err != context.Canceled {
		t.Fatalf("Run returned %v, want context.Canceled", err)
	}
}

// TestMinimumIntervalIsEnforcedPerTopic keeps a chatty channel from turning into
// a poll loop: the API budget (§0.4) is calculated on the polling cadence, and an
// unbounded hint path would silently multiply it.
func TestMinimumIntervalIsEnforcedPerTopic(t *testing.T) {
	clk := clock.NewFake(pollStart)
	refresh := newBlockingRefresh()
	close(refresh.release) // never block; this test is about the wait before the call
	h := &filldetect.Hints{Refresh: refresh.fn, Clock: clk, MinInterval: 2 * time.Second}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = h.Run(ctx) }()

	h.Observe(event("share-holdings"))
	waitFor(t, "the first re-fetch", func() bool { return refresh.count() == 1 })

	h.Observe(event("share-holdings"))
	if !clk.WaitForSleepers(1, 2*time.Second) {
		t.Fatal("the second re-fetch should be waiting out the minimum interval")
	}
	if got := refresh.count(); got != 1 {
		t.Fatalf("re-fetches = %d during the interval, want 1", got)
	}

	clk.Advance(2 * time.Second)
	waitFor(t, "the throttled re-fetch", func() bool { return refresh.count() == 2 })
}

// TestTopicsCoalesceIndependently: a holdings event must not be swallowed by an
// in-flight pending-order re-fetch. The spec says coalescing is per topic.
func TestTopicsCoalesceIndependently(t *testing.T) {
	clk := clock.NewFake(pollStart)
	refresh := newBlockingRefresh()
	close(refresh.release)
	h := &filldetect.Hints{Refresh: refresh.fn, Clock: clk, MinInterval: -1}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = h.Run(ctx) }()

	h.Observe(event("pending-order-refresh"))
	h.Observe(event("share-holdings"))
	h.Observe(event("purchase-price-refresh"))

	waitFor(t, "all three topics", func() bool { return refresh.count() >= 3 })
	seen := map[filldetect.HintTopic]bool{}
	for _, topic := range refresh.seen() {
		seen[topic] = true
	}
	for _, want := range []filldetect.HintTopic{
		filldetect.TopicPendingOrders,
		filldetect.TopicHoldings,
		filldetect.TopicPurchasePrice,
	} {
		if !seen[want] {
			t.Fatalf("topic %s never triggered a re-fetch (saw %v)", want, refresh.seen())
		}
	}
}

// TestUnrelatedEventsAreIgnored: web-push is a human notification, not a state
// signal, and re-fetching on it would spend budget for nothing.
func TestUnrelatedEventsAreIgnored(t *testing.T) {
	clk := clock.NewFake(pollStart)
	refresh := newBlockingRefresh()
	close(refresh.release)
	h := &filldetect.Hints{Refresh: refresh.fn, Clock: clk}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = h.Run(ctx) }()

	for _, unrelated := range []string{"web-push", "icon-refresh", "setting-refresh", ""} {
		h.Observe(event(unrelated))
	}
	time.Sleep(20 * time.Millisecond)
	if got := refresh.count(); got != 0 {
		t.Fatalf("re-fetches = %d, want 0 for events this consumer does not act on", got)
	}
	if h.Stats().Ignored != 4 {
		t.Fatalf("ignored = %d, want 4", h.Stats().Ignored)
	}
}

// TestEventPayloadIsNeverUsedAsState is the spec's SHALL NOT, made concrete: an
// event claiming a 999-share fill changes nothing, because the state comes from
// the re-fetch and the re-fetch says the order is untouched.
func TestEventPayloadIsNeverUsedAsState(t *testing.T) {
	clk := clock.NewFake(pollStart)
	pager := newPager(page("", rawOrder{id: "o-1", quantity: "10", filled: "0"}))
	d, j := newJournalDetector(t, clk, pager, nil, filepath.Join(t.TempDir(), "journal.db"))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h := &filldetect.Hints{Refresh: d.Refresh, Clock: clk, MinInterval: -1}
	go func() { _ = h.Run(ctx) }()

	lying := push.Event{
		Type: "share-holdings",
		Msg: map[string]any{
			"stockCode":      "US20181228002",
			"filledQuantity": "999",
			"quantity":       "999",
		},
	}
	h.Observe(lying)
	waitFor(t, "the hint-triggered re-fetch", func() bool { return h.Stats().Refreshes == 1 })

	events, err := j.FillEvents(ctx, "o-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 0 {
		t.Fatalf("fill events = %d, want 0 — the payload is a hint, not evidence", len(events))
	}
	stored, err := j.LookupFill(ctx, "o-1")
	if err != nil {
		t.Fatal(err)
	}
	if stored.FilledQuantity != "0" {
		t.Fatalf("stored quantity = %q, want the re-fetched 0", stored.FilledQuantity)
	}
}

// TestHintShortensDetectionLatency is the point of the whole file: an event makes
// the poll happen now, and the fill is picked up from the official API.
func TestHintShortensDetectionLatency(t *testing.T) {
	clk := clock.NewFake(pollStart)
	pager := newPager(page("", rawOrder{id: "o-1", quantity: "10", filled: "6"}))
	d, j := newJournalDetector(t, clk, pager, nil, filepath.Join(t.TempDir(), "journal.db"))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h := &filldetect.Hints{Refresh: d.Refresh, Clock: clk, MinInterval: -1}
	go func() { _ = h.Run(ctx) }()

	h.Observe(event("pending-order-refresh"))
	waitFor(t, "the re-fetch", func() bool { return h.Stats().Refreshes == 1 })

	events, err := j.FillEvents(ctx, "o-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].DeltaQuantity != "6" {
		t.Fatalf("fill events = %+v, want the fill the hint made us look for", events)
	}
}

// TestListenConsumesALiveSSEStream wires the real push.Listener to the consumer
// over an httptest server — no real Toss endpoint, per the change's test rules.
func TestListenConsumesALiveSSEStream(t *testing.T) {
	clk := clock.NewFake(pollStart)
	refresh := newBlockingRefresh()
	close(refresh.release)
	h := &filldetect.Hints{Refresh: refresh.fn, Clock: clk, MinInterval: -1}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Error("ResponseWriter is not a Flusher")
			return
		}
		for _, frame := range []string{
			`{"type":"purchase-price-refresh","msg":{"stockCode":"US20181228002"},"key":"1"}`,
			`{"type":"pending-order-refresh","key":"1"}`,
			`{"type":"share-holdings","msg":{"stockCode":"US20181228002"}}`,
			`{"type":"web-push","msg":{"title":"주문 성공"}}`,
		} {
			fmt.Fprintf(w, "id: %d\ndata: %s\n\n", len(frame), frame)
			flusher.Flush()
		}
		<-r.Context().Done()
	}))
	defer srv.Close()

	listener := push.NewListener(
		&session.Session{Cookies: map[string]string{"SESSION": "test"}},
		push.WithStreamURL(srv.URL),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = h.Run(ctx) }()
	go func() { _ = h.Listen(ctx, listener) }()

	waitFor(t, "the three actionable frames", func() bool { return h.Stats().Refreshes >= 3 })
	if h.Stats().Ignored != 1 {
		t.Fatalf("ignored = %d, want 1 (the web-push frame)", h.Stats().Ignored)
	}
}

// TestHintsRefuseIncompleteWiring: a consumer with nowhere to send a refresh
// would silently swallow every hint.
func TestHintsRefuseIncompleteWiring(t *testing.T) {
	h := &filldetect.Hints{Clock: clock.NewFake(pollStart)}
	if err := h.Run(context.Background()); err == nil {
		t.Fatal("a hint consumer with no refresh function must refuse to run")
	}
}
