package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
)

type countedSnapshotSource struct {
	calls   atomic.Int32
	started chan struct{}
	release chan struct{}
}

func (s *countedSnapshotSource) Snapshot(context.Context) ([]byte, error) {
	call := s.calls.Add(1)
	if call == 1 && s.started != nil {
		close(s.started)
		<-s.release
	}
	return []byte(fmt.Sprintf(`{"snapshot":%d}`, call)), nil
}

func TestHTTPAPISnapshotCacheCoalescesThirtyTwoReconnects(t *testing.T) {
	source := &countedSnapshotSource{started: make(chan struct{}), release: make(chan struct{})}
	cache, err := newHTTPAPISnapshotCache(source)
	if err != nil {
		t.Fatal(err)
	}

	const clients = 32
	results := make(chan string, clients)
	var group sync.WaitGroup
	for range clients {
		group.Add(1)
		go func() {
			defer group.Done()
			body, readErr := cache.Snapshot(context.Background())
			if readErr != nil {
				results <- "error: " + readErr.Error()
				return
			}
			results <- string(body)
		}()
	}
	<-source.started
	close(source.release)
	group.Wait()
	close(results)

	if calls := source.calls.Load(); calls != 1 {
		t.Fatalf("official snapshot calls=%d want=1", calls)
	}
	for result := range results {
		if result != `{"snapshot":1}` {
			t.Fatalf("reconnect snapshot=%q", result)
		}
	}
}

func TestHTTPAPISnapshotCacheRefreshesOnlyOnServerSchedule(t *testing.T) {
	source := &countedSnapshotSource{}
	cache, err := newHTTPAPISnapshotCache(source)
	if err != nil {
		t.Fatal(err)
	}
	first, err := cache.Snapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	second, err := cache.Snapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) || source.calls.Load() != 1 {
		t.Fatalf("cached snapshots=%s/%s calls=%d", first, second, source.calls.Load())
	}
	refreshed, err := cache.Refresh(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if string(refreshed) != `{"snapshot":2}` || source.calls.Load() != 2 {
		t.Fatalf("scheduled refresh=%s calls=%d", refreshed, source.calls.Load())
	}
}
