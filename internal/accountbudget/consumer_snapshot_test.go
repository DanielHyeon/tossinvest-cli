//go:build linux

package accountbudget

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestA166ConsumerSnapshotBoundary(t *testing.T) {
	b, _ := observedFixture(t)
	view, err := b.ObservationSnapshot(context.Background(), ObservationCursor{})
	if err != nil || view.Cursor.Scope != "scope" {
		t.Fatalf("view=%+v err=%v", view, err)
	}
	var nilBudget *Budget
	if _, err = nilBudget.ObservationSnapshot(context.Background(), ObservationCursor{}); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("nil budget: %v", err)
	}
	if _, err = b.ObservationSnapshot(nil, ObservationCursor{}); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("nil context: %v", err)
	}
	plain, _ := fixture(t)
	if _, err = plain.ObservationSnapshot(context.Background(), ObservationCursor{}); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("unobserved: %v", err)
	}
	if _, err = b.ObservationSnapshot(context.Background(), ObservationCursor{Scope: view.Cursor.Scope, Instance: view.Cursor.Instance, Event: 1}); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("future cursor: %v", err)
	}
}

func TestA166ConsumerSnapshotCopyAndPastCursor(t *testing.T) {
	b, store := observedFixture(t)
	lease, err := b.Acquire(context.Background(), Normal)
	if err != nil {
		t.Fatal(err)
	}
	seq, err := store.Begin(context.Background(), lease, ObservationOther, ObservationKey{})
	if err != nil {
		t.Fatal(err)
	}
	err = store.Finish(context.Background(), lease, seq, ObservationRateLimited, nil)
	err = errors.Join(err, lease.Release())
	if err != nil {
		t.Fatal(err)
	}
	view, err := b.ObservationSnapshot(context.Background(), ObservationCursor{})
	if err != nil || len(view.Events) != 1 {
		t.Fatalf("event=%+v err=%v", view, err)
	}
	cursor := view.Cursor
	view.Events[0].Sequence = 999
	view.Last.Attempt = 999
	again, err := b.ObservationSnapshot(context.Background(), ObservationCursor{Scope: cursor.Scope, Instance: cursor.Instance})
	if err != nil || len(again.Events) != 1 || again.Events[0].Sequence != 1 || again.Last.Attempt != seq {
		t.Fatalf("copy=%+v err=%v", again, err)
	}
	if _, err = b.ObservationSnapshot(context.Background(), ObservationCursor{Scope: "other", Instance: cursor.Instance}); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("scope: %v", err)
	}
	if err = b.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = b.ObservationSnapshot(context.Background(), ObservationCursor{}); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("closed: %v", err)
	}
}

func TestA166ConsumerSnapshotPinsDuringClose(t *testing.T) {
	b, store := observedFixture(t)
	held, err := store.lock(context.Background(), true)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, e := b.ObservationSnapshot(ctx, ObservationCursor{}); done <- e }()
	deadline := time.Now().Add(time.Second)
	active := false
	for time.Now().Before(deadline) {
		b.mu.Lock()
		active = b.active > 0
		b.mu.Unlock()
		if active {
			break
		}
		time.Sleep(time.Millisecond)
	}
	if !active {
		cancel()
		_ = release(held)
		t.Fatal("snapshot did not pin")
	}
	if err = b.Close(); err == nil {
		cancel()
		_ = release(held)
		t.Fatal("close passed active snapshot")
	}
	cancel()
	if err = release(held); err != nil {
		t.Fatal(err)
	}
	if err = <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("snapshot cancel: %v", err)
	}
}

func TestA166ConsumerObservedBodyCloseReleasesWire(t *testing.T) {
	b, _ := observedFixture(t)
	hold := make(chan struct{})
	tr, url, _ := observedTLS(t, b, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("partial"))
		w.(http.Flusher).Flush()
		select {
		case <-hold:
		case <-r.Context().Done():
		}
	})
	defer close(hold)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := tr.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	lease, err := b.Acquire(ctx, Critical)
	cancel()
	if lease != nil {
		_ = lease.Release()
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		_ = resp.Body.Close()
		t.Fatalf("critical escaped body: %v", err)
	}
	if err = resp.Body.Close(); err != nil {
		t.Fatal(err)
	}
	ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	lease, err = b.Acquire(ctx, Critical)
	if err != nil {
		t.Fatal(err)
	}
	if err = lease.Release(); err != nil {
		t.Fatal(err)
	}
}
