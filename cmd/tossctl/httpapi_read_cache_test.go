package main

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestHTTPAPITimedReadCacheBoundsConcurrentOfficialReadsAndFailureRetries(t *testing.T) {
	var cache httpAPITimedReadCache[int]
	var calls atomic.Int32
	now := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	read := func(context.Context) (int, error) {
		return int(calls.Add(1)), nil
	}

	const clients = 32
	var group sync.WaitGroup
	for range clients {
		group.Add(1)
		go func() {
			defer group.Done()
			value, err := cache.Get(context.Background(), now, httpAPIBrokerReadTTL, httpAPIBrokerReadErrorTTL, read)
			if err != nil || value != 1 {
				t.Errorf("coalesced value/error=%d/%v", value, err)
			}
		}()
	}
	group.Wait()
	if calls.Load() != 1 {
		t.Fatalf("concurrent official reads=%d want=1", calls.Load())
	}
	value, err := cache.Get(context.Background(), now.Add(httpAPIBrokerReadTTL), httpAPIBrokerReadTTL, httpAPIBrokerReadErrorTTL, read)
	if err != nil || value != 2 || calls.Load() != 2 {
		t.Fatalf("scheduled refresh value/error/calls=%d/%v/%d", value, err, calls.Load())
	}

	var failures httpAPITimedReadCache[int]
	failureCalls := 0
	fail := func(context.Context) (int, error) {
		failureCalls++
		return 0, errors.New("broker unavailable")
	}
	for range 8 {
		_, _ = failures.Get(context.Background(), now, httpAPIBrokerReadTTL, httpAPIBrokerReadErrorTTL, fail)
	}
	if failureCalls != 1 {
		t.Fatalf("failure retries=%d want=1 during backoff", failureCalls)
	}
}
