package main

import (
	"context"
	"sync"
	"time"
)

const (
	httpAPIBrokerReadTTL      = 30 * time.Second
	httpAPIBrokerReadErrorTTL = 5 * time.Second
)

// httpAPITimedReadCache bounds official API work independently of request
// volume. Concurrent refreshes coalesce, successful reads remain immutable for
// one server-owned interval, and failures get a short backoff so an outage does
// not turn no-token VPN traffic into a retry storm.
type httpAPITimedReadCache[T any] struct {
	mu         sync.Mutex
	value      T
	ready      bool
	lastAt     time.Time
	lastErr    error
	refreshing bool
	done       chan struct{}
}

func (c *httpAPITimedReadCache[T]) Get(
	ctx context.Context,
	now time.Time,
	successTTL time.Duration,
	errorTTL time.Duration,
	read func(context.Context) (T, error),
) (T, error) {
	for {
		c.mu.Lock()
		age := now.Sub(c.lastAt)
		if c.ready && !c.lastAt.IsZero() && age >= 0 && age < successTTL {
			value := c.value
			c.mu.Unlock()
			return value, nil
		}
		if c.lastErr != nil && !c.lastAt.IsZero() && age >= 0 && age < errorTTL {
			err := c.lastErr
			var zero T
			c.mu.Unlock()
			return zero, err
		}
		if c.refreshing {
			done := c.done
			c.mu.Unlock()
			select {
			case <-ctx.Done():
				var zero T
				return zero, ctx.Err()
			case <-done:
				continue
			}
		}
		c.refreshing = true
		c.done = make(chan struct{})
		done := c.done
		c.mu.Unlock()

		value, err := read(ctx)

		c.mu.Lock()
		if err == nil {
			c.value = value
			c.ready = true
		} else {
			c.ready = false
		}
		c.lastAt = now
		c.lastErr = err
		c.refreshing = false
		close(done)
		c.mu.Unlock()
		return value, err
	}
}
