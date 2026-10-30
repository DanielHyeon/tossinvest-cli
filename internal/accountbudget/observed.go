package accountbudget

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ObservationCollector belongs to one logical request; Snapshot is concurrency safe.
type ObservationCollector struct {
	mu          sync.Mutex
	rateLimited bool
	localErr    error
}
type ObservationResult struct {
	RateLimited bool
	LocalError  error
}

func (c *ObservationCollector) Snapshot() ObservationResult {
	if c == nil {
		return ObservationResult{}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return ObservationResult{c.rateLimited, c.localErr}
}
func (c *ObservationCollector) record(rate bool, err error) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.rateLimited = c.rateLimited || rate
	c.localErr = errors.Join(c.localErr, err)
}

type observedScope struct {
	kind      ObservationKind
	key       ObservationKey
	collector *ObservationCollector
}
type observedScopeKey struct{}

// WithObservation supplies explicit request kind; default is non-poll.
func WithObservation(ctx context.Context, kind ObservationKind, key ObservationKey, collector *ObservationCollector) context.Context {
	return context.WithValue(ctx, observedScopeKey{}, observedScope{kind, key, collector})
}
func requestObservation(ctx context.Context) observedScope {
	s, ok := ctx.Value(observedScopeKey{}).(observedScope)
	if !ok {
		return observedScope{kind: ObservationOther}
	}
	return s
}

var ErrObservationDeferred = errors.New("observation poll deferred")

func retryAfterHint(value string, now time.Time) *time.Duration {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	if n, err := strconv.ParseUint(value, 10, 64); err == nil {
		if n > 30 {
			return nil
		}
		d := time.Duration(n) * time.Second
		return &d
	}
	if strings.HasPrefix(value, "+") || strings.HasPrefix(value, "-") {
		return nil
	}
	when, err := http.ParseTime(value)
	if err != nil {
		return nil
	}
	d := when.Sub(now)
	if d < 0 || d > 30*time.Second {
		return nil
	}
	return &d
}
