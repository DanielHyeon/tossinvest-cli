package official

import (
	"context"
	"time"
)

// AttemptTrace is emitted at the HTTP boundary for an individual request
// attempt. Body is the exact response body as read, before status classification
// or envelope decoding. The timestamps come from one process-local clock; users
// that serialize this evidence must preserve their own monotonic sequence rather
// than treating the wall component as causal authority.
type AttemptTrace struct {
	RequestStart     time.Time
	BodyReadComplete time.Time
	StatusCode       int
	Body             []byte
	Err              error
}

// AttemptObserver receives additive transport evidence. It is intentionally a
// context capability: existing public readers retain their result/error contract
// and no mutation API is introduced.
type AttemptObserver func(AttemptTrace)

type attemptObserverKey struct{}

// WithAttemptObserver enables per-attempt tracing for calls made with ctx.
func WithAttemptObserver(ctx context.Context, observer AttemptObserver) context.Context {
	if observer == nil {
		return ctx
	}
	return context.WithValue(ctx, attemptObserverKey{}, observer)
}

func observeAttempt(ctx context.Context, trace AttemptTrace) {
	observer, _ := ctx.Value(attemptObserverKey{}).(AttemptObserver)
	if observer != nil {
		observer(trace)
	}
}
