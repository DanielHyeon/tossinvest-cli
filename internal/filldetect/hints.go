package filldetect

// hints.go consumes the WTS SSE channel (harden-execution-base task 3.3).
//
// # What an event is allowed to do
//
// Exactly one thing: make the next poll happen sooner. The fill-detection spec
// puts it as a SHALL NOT — "이벤트 페이로드를 상태 변경 근거로 사용해서는 안 된다"
// — and the reason is not distrust of the payload, it is provenance. The SSE
// channel is a browser-session feature: it needs WTS cookies, it can vanish, and
// Toss itself documents it as a "something changed, re-fetch" notification rather
// than a state feed (docs/reverse-engineering/push-events.md). An engine that
// updated a position from a frame would have two sources of truth for a live
// account, and no way to say which one is right when they disagree.
//
// So this file maps event *types* onto topics and throws the payload away. It is
// structurally impossible for a frame's contents to reach the ledger: the only
// thing that crosses the boundary is a topic name.
//
// # Why coalescing is a safety property, not a nicety
//
// A single cancel emits pending-order-refresh and purchase-price-refresh three
// times each (same doc). Re-fetching per frame would turn one user action into
// six full account reads and spend the rate-limit budget the retry matrix
// accounts for (§0.4) — and being throttled is precisely how fill detection goes
// blind. So each topic has one in-flight re-fetch and at most one queued behind
// it, and a minimum interval between them.
//
// The queue depth of one is deliberate: an event that arrives *during* a re-fetch
// may describe a change that re-fetch did not see, so it must not be dropped.
// Anything beyond that is describing a change the queued re-fetch will observe
// anyway.

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/push"
)

// HintTopic is a family of broker state a re-fetch can be triggered for.
type HintTopic string

const (
	// TopicPendingOrders — the pending-order list changed (prepare, create,
	// cancel, amend).
	TopicPendingOrders HintTopic = "pending-order-refresh"
	// TopicHoldings — a holdings delta, which is what a fill looks like from the
	// account's side.
	TopicHoldings HintTopic = "share-holdings"
	// TopicPurchasePrice — the average price or buying power of a stock moved.
	TopicPurchasePrice HintTopic = "purchase-price-refresh"
)

// hintTopics is the whole vocabulary. Everything else — web-push, icon-refresh,
// price-refresh — is ignored: they carry no order-state implication, and
// re-fetching on them would spend budget to learn nothing.
var hintTopics = []HintTopic{TopicPendingOrders, TopicHoldings, TopicPurchasePrice}

// TopicFor maps an SSE event onto a topic.
//
// It reads ev.Type and nothing else. That is the enforcement point for "the
// payload is not evidence": there is no path from ev.Msg to a caller.
func TopicFor(ev push.Event) (HintTopic, bool) {
	for _, topic := range hintTopics {
		if ev.Type == string(topic) {
			return topic, true
		}
	}
	return "", false
}

// DefaultHintInterval is the minimum gap between two re-fetches of one topic.
//
// One second against a 3-second poll loop: a sustained burst on all three topics
// can at most treble the read rate, and the coalescing means a real burst (Toss
// emits each type three times per action) produces one extra cycle, not six.
const DefaultHintInterval = time.Second

// HintStats is what the consumer has done, for status output and alerting.
type HintStats struct {
	// Events is every frame offered to Observe.
	Events int
	// Refreshes is how many re-fetches were actually run.
	Refreshes int
	// Coalesced is how many events were merged into an already-queued re-fetch.
	Coalesced int
	// Ignored is how many events had no topic.
	Ignored int
	// Failures is how many re-fetches returned an error.
	Failures int
}

// Hints turns SSE events into coalesced re-fetches.
type Hints struct {
	// Refresh runs one poll cycle. Detector.Refresh is the intended value.
	// Required.
	Refresh func(ctx context.Context, topic HintTopic) error
	// Clock drives the minimum interval. Defaults to clock.System().
	Clock clock.Clock
	// MinInterval overrides DefaultHintInterval. Zero takes the default — the
	// zero value of this struct must be throttled, because an unthrottled hint
	// path is a rate-limit incident waiting for a busy day. A negative value
	// disables the wait, which only the tests want.
	MinInterval time.Duration
	// Logf receives re-fetch failures. Optional.
	Logf func(format string, args ...any)

	once   sync.Once
	queues map[HintTopic]chan struct{}

	mu    sync.Mutex
	stats HintStats
}

// Observe offers one SSE event. It never blocks and never fails: a hint consumer
// that could stall the listener would turn a fast channel into a slow one.
func (h *Hints) Observe(ev push.Event) {
	h.ensure()

	topic, ok := TopicFor(ev)
	h.mu.Lock()
	h.stats.Events++
	if !ok {
		h.stats.Ignored++
		h.mu.Unlock()
		return
	}
	h.mu.Unlock()

	select {
	case h.queues[topic] <- struct{}{}:
	default:
		// A re-fetch for this topic is already queued behind the running one,
		// and it will observe whatever this event is about.
		h.mu.Lock()
		h.stats.Coalesced++
		h.mu.Unlock()
	}
}

// ErrNoRefreshWired means a hint consumer has nothing to route its hints to.
//
// It is exported because the engine runtime checks it at *assembly* time rather
// than discovering it when Run returns (engine-safety, add-engine-runtime: 힌트
// 라우팅을 포함하는 경우 Refresh 미배선은 감독이 아니라 조립 시점 검증으로
// 거부한다 — SHALL). The reason for the earlier check is the supervision
// contract's own premise: a loop that returns immediately would be read as a
// defensive-termination incident, when what actually happened is that nobody
// wired it.
var ErrNoRefreshWired = errors.New("filldetect: a hint consumer needs a refresh function — " +
	"an unrouted hint is a silently dropped one")

// Validate reports whether this consumer can do anything at all. It is Run's own
// precondition, exposed so a caller can ask before starting a goroutine.
func (h *Hints) Validate() error {
	if h == nil || h.Refresh == nil {
		return ErrNoRefreshWired
	}
	return nil
}

// Run works the topic queues until ctx is done. One worker per topic, so a slow
// re-fetch of one family never delays another.
func (h *Hints) Run(ctx context.Context) error {
	if err := h.Validate(); err != nil {
		return err
	}
	h.ensure()

	var wg sync.WaitGroup
	for _, topic := range hintTopics {
		wg.Add(1)
		go func(topic HintTopic) {
			defer wg.Done()
			h.work(ctx, topic)
		}(topic)
	}
	wg.Wait()
	return ctx.Err()
}

// Listen subscribes to the SSE channel and feeds Observe.
//
// The listener is passed in rather than built here: it needs a WTS session, and
// this package must keep working when there is none. A caller without a session
// simply does not call Listen — the poll loop is unaffected, which is the whole
// point of SSE being a hint.
func (h *Hints) Listen(ctx context.Context, listener *push.Listener) error {
	if listener == nil {
		return errors.New("filldetect: Listen needs a push listener")
	}
	return listener.ListenWithRetry(ctx, h.Observe)
}

// Stats returns a snapshot of the counters.
func (h *Hints) Stats() HintStats {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.stats
}

// work is one topic's single-flight loop.
func (h *Hints) work(ctx context.Context, topic HintTopic) {
	clk := h.clk()
	interval := h.interval()
	var last time.Time

	for {
		select {
		case <-ctx.Done():
			return
		case <-h.queues[topic]:
		}

		// The minimum interval is applied before the call, not after, so a hint
		// that arrives long after the previous one is served immediately.
		if interval > 0 && !last.IsZero() {
			if wait := interval - clk.Now().Sub(last); wait > 0 {
				if err := clk.Sleep(ctx, wait); err != nil {
					return
				}
			}
		}
		if ctx.Err() != nil {
			return
		}

		err := h.Refresh(ctx, topic)
		last = clk.Now()

		h.mu.Lock()
		h.stats.Refreshes++
		if err != nil {
			h.stats.Failures++
		}
		h.mu.Unlock()

		if err != nil && h.Logf != nil {
			// A failed re-fetch is not escalated here. The poll loop's own
			// outage and staleness handling already blocks entries when reads
			// stop working, and duplicating that judgement in the hint path
			// would give the same condition two owners.
			h.Logf("filldetect: %s hint re-fetch failed: %v", topic, err)
		}
	}
}

func (h *Hints) ensure() {
	h.once.Do(func() {
		h.queues = make(map[HintTopic]chan struct{}, len(hintTopics))
		for _, topic := range hintTopics {
			// Capacity one *is* the coalescing: a second event while one is
			// queued has nothing new to ask for.
			h.queues[topic] = make(chan struct{}, 1)
		}
	})
}

func (h *Hints) clk() clock.Clock {
	if h.Clock == nil {
		return clock.System()
	}
	return h.Clock
}

func (h *Hints) interval() time.Duration {
	switch {
	case h.MinInterval < 0:
		return 0
	case h.MinInterval == 0:
		return DefaultHintInterval
	default:
		return h.MinInterval
	}
}

// Refresh is the pipe the hints join: it runs one full poll cycle.
//
// The topic is not used to pick what to read. Every cycle reads the open list,
// the tracked orders and the account anyway, and a topic-specific partial read
// would produce a snapshot assembled from two different instants — exactly what
// the reconciliation contract forbids for its own snapshot (task 3.4).
func (d *Detector) Refresh(ctx context.Context, _ HintTopic) error {
	_, err := d.PollOnce(ctx)
	return err
}
