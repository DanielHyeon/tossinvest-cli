//go:build linux

package accountbudget

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
)

type pollCountingBody struct {
	io.Reader
	closes   atomic.Int32
	closeErr error
}

func (b *pollCountingBody) Close() error { b.closes.Add(1); return b.closeErr }

func TestA166PollCollectorStopsSubsequentWire(t *testing.T) {
	b, store := observedFixture(t)
	var calls atomic.Int32
	tr, url, _ := observedTLS(t, b, func(http.ResponseWriter, *http.Request) { calls.Add(1) })
	cause := errors.New("fixture observation fault")
	collector := &ObservationCollector{}
	collector.record(false, cause)
	closeErr := errors.New("fixture body close")
	body := &pollCountingBody{Reader: strings.NewReader("fixture"), closeErr: closeErr}
	ctx := WithObservation(context.Background(), ObservationToken, ObservationKey{}, collector)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := tr.RoundTrip(req)
	if resp != nil {
		_ = resp.Body.Close()
	}
	if !errors.Is(err, cause) || !errors.Is(err, closeErr) || !errors.Is(err, ErrNotSent) || body.closes.Load() != 1 || calls.Load() != 0 {
		t.Fatalf("err=%v closes=%d calls=%d", err, body.closes.Load(), calls.Load())
	}
	state, err := store.Snapshot(context.Background(), ObservationCursor{})
	if err != nil || state.Last != nil || state.Pending != nil {
		t.Fatalf("state=%+v err=%v", state, err)
	}
}

func TestA166PollCollectorRateLimitStopsWire(t *testing.T) {
	b, store := observedFixture(t)
	var calls atomic.Int32
	tr, url, _ := observedTLS(t, b, func(http.ResponseWriter, *http.Request) { calls.Add(1) })
	collector := &ObservationCollector{}
	collector.record(true, nil)
	ctx := WithTokenObservation(WithObservation(context.Background(), ObservationPoll, ObservationKey{1}, collector))
	body := &pollCountingBody{Reader: strings.NewReader("fixture")}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := tr.RoundTrip(req)
	if resp != nil {
		_ = resp.Body.Close()
	}
	if !errors.Is(err, ErrObservationRateLimited) || !errors.Is(err, ErrNotSent) || body.closes.Load() != 1 || calls.Load() != 0 {
		t.Fatalf("err=%v closes=%d wire=%d", err, body.closes.Load(), calls.Load())
	}
	state, err := store.Snapshot(context.Background(), ObservationCursor{})
	if err != nil || state.Last != nil || state.Pending != nil {
		t.Fatalf("state=%+v err=%v", state, err)
	}
}
