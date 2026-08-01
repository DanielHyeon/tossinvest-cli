package main

import (
	"context"
	"errors"
	"sync"
)

type httpAPISnapshotSource interface {
	Snapshot(context.Context) ([]byte, error)
}

// httpAPISnapshotCache prevents reconnect storms from becoming broker request
// storms. One server-owned refresh may touch the official read seams; every SSE
// fallback receives the last complete immutable snapshot. The first request is
// singleflight so an empty cache also costs one refresh, not one per client.
type httpAPISnapshotCache struct {
	source httpAPISnapshotSource

	mu         sync.Mutex
	data       []byte
	ready      bool
	refreshing bool
	done       chan struct{}
	lastErr    error
}

func newHTTPAPISnapshotCache(source httpAPISnapshotSource) (*httpAPISnapshotCache, error) {
	if source == nil {
		return nil, errors.New("httpapi: snapshot source is required")
	}
	return &httpAPISnapshotCache{source: source}, nil
}

func (c *httpAPISnapshotCache) Snapshot(ctx context.Context) ([]byte, error) {
	return c.load(ctx, false)
}

func (c *httpAPISnapshotCache) Refresh(ctx context.Context) ([]byte, error) {
	return c.load(ctx, true)
}

func (c *httpAPISnapshotCache) load(ctx context.Context, force bool) ([]byte, error) {
	c.mu.Lock()
	if !force && c.ready {
		data := append([]byte(nil), c.data...)
		c.mu.Unlock()
		return data, nil
	}
	if c.refreshing {
		done := c.done
		c.mu.Unlock()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-done:
		}
		c.mu.Lock()
		defer c.mu.Unlock()
		if c.ready {
			return append([]byte(nil), c.data...), nil
		}
		return nil, c.lastErr
	}
	c.refreshing = true
	c.done = make(chan struct{})
	done := c.done
	c.mu.Unlock()

	data, err := c.source.Snapshot(ctx)

	c.mu.Lock()
	if err == nil {
		c.data = append(c.data[:0], data...)
		c.ready = true
	}
	c.lastErr = err
	c.refreshing = false
	close(done)
	c.mu.Unlock()
	if err != nil {
		return nil, err
	}
	return append([]byte(nil), data...), nil
}
