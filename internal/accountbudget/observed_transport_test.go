//go:build linux

package accountbudget

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
	"time"
)

func TestObservedOpenRejectsMissingEvent(t *testing.T) {
	b, root := fixture(t)
	if err := b.Close(); err != nil {
		t.Fatal(err)
	}
	_, err := OpenObserved(context.Background(), Config{Root: root, Scope: "scope-a", Instance: "instance-a", Clock: clock.System()}, FileID{Device: 1, Inode: 1}, clock.SystemBootSource())
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("missing event: %v", err)
	}
}

func observedFixture(t *testing.T) (*Budget, *ObservationStore) {
	t.Helper()
	old, store, boot := a166Store(t)
	cfg := Config{Root: old.root, Scope: old.scope, Instance: old.instance, Clock: clock.System()}
	id := store.lockID
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}
	b, err := OpenObserved(context.Background(), cfg, id, boot)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if e := b.Close(); e != nil {
			t.Error(e)
		}
	})
	return b, b.observation
}

func observedTLS(t *testing.T, b *Budget, handler http.HandlerFunc) (*Transport, string, chan error) {
	t.Helper()
	server := httptest.NewTLSServer(handler)
	t.Cleanup(server.Close)
	pool := x509.NewCertPool()
	pool.AddCert(server.Certificate())
	expected := server.Listener.Addr().String()
	dial := func(ctx context.Context, network, address string) (net.Conn, error) {
		if network != "tcp" || address != expected {
			return nil, errors.New("unexpected destination")
		}
		return (&net.Dialer{}).DialContext(ctx, network, address)
	}
	faults := make(chan error, 8)
	tr, err := NewTransport(b, &http.Transport{DialContext: dial, TLSClientConfig: &tls.Config{RootCAs: pool}}, func(e error) { faults <- e })
	if err != nil {
		t.Fatal(err)
	}
	return tr, server.URL, faults
}

func TestObservedTransportRateLimitAndPoison(t *testing.T) {
	b, s := observedFixture(t)
	var calls atomic.Int32
	tr, url, _ := observedTLS(t, b, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Retry-After", "2")
		w.WriteHeader(429)
		_, _ = w.Write([]byte("original"))
	})
	var key ObservationKey
	key[0] = 7
	c := &ObservationCollector{}
	ctx := WithObservation(context.Background(), ObservationPoll, key, c)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := tr.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 429 || string(body) != "original" || !c.Snapshot().RateLimited {
		t.Fatal("response or observation lost")
	}
	remaining, err := s.PollRemaining(context.Background(), key)
	if err != nil || remaining <= 0 {
		t.Fatalf("remaining %v %v", remaining, err)
	}
	req, _ = http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	_, err = tr.RoundTrip(req)
	if !errors.Is(err, ErrObservationDeferred) || !errors.Is(err, ErrNotSent) || calls.Load() != 1 {
		t.Fatalf("defer: %v calls %d", err, calls.Load())
	}
	snap, err := s.Snapshot(context.Background(), ObservationCursor{})
	if err != nil || len(snap.Events) != 1 || snap.Pending != nil {
		t.Fatalf("snapshot %+v %v", snap, err)
	}
}

func TestObservedTransportOrphanRecheck(t *testing.T) {
	b, s := observedFixture(t)
	var calls atomic.Int32
	tr, url, _ := observedTLS(t, b, func(w http.ResponseWriter, r *http.Request) { calls.Add(1) })
	var key ObservationKey
	key[0] = 9
	lease, err := b.Acquire(context.Background(), Normal)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.Begin(context.Background(), lease, ObservationPoll, key); err != nil {
		t.Fatal(err)
	}
	if err = lease.Release(); err != nil {
		t.Fatal(err)
	}
	c := &ObservationCollector{}
	req, _ := http.NewRequestWithContext(WithObservation(context.Background(), ObservationPoll, key, c), http.MethodGet, url, nil)
	_, err = tr.RoundTrip(req)
	if !errors.Is(err, ErrObservationPoisoned) || !errors.Is(err, ErrNotSent) || calls.Load() != 0 || !errors.Is(c.Snapshot().LocalError, ErrObservationPoisoned) {
		t.Fatalf("orphan: %v %+v", err, c.Snapshot())
	}
	snap, e := s.Snapshot(context.Background(), ObservationCursor{})
	if e != nil || len(snap.Gaps) != 1 || snap.Pending != nil {
		t.Fatalf("snapshot %+v %v", snap, e)
	}
}

func TestObservedTransportFinishFaultPreservesResponse(t *testing.T) {
	b, s := observedFixture(t)
	tr, url, faults := observedTLS(t, b, func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(429); _, _ = w.Write([]byte("body")) })
	fail := errors.New("finish fault")
	var calls int
	s.testFault = func(stage string) error {
		if stage == "state-write" {
			calls++
			if calls == 2 {
				return fail
			}
		}
		return nil
	}
	c := &ObservationCollector{}
	req, _ := http.NewRequestWithContext(WithObservation(context.Background(), ObservationOther, ObservationKey{}, c), http.MethodPost, url, nil)
	resp, err := tr.RoundTrip(req)
	if err != nil || resp == nil || resp.StatusCode != 429 {
		t.Fatalf("response %v %v", resp, err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil || string(body) != "body" {
		t.Fatalf("body %q %v", body, err)
	}
	got := c.Snapshot()
	if !got.RateLimited || !errors.Is(got.LocalError, fail) {
		t.Fatalf("collector %+v", got)
	}
	select {
	case err := <-faults:
		if !errors.Is(err, fail) {
			t.Fatal(err)
		}
	default:
		t.Fatal("missing callback")
	}
}

type observedBody struct {
	closed   atomic.Int32
	closeErr error
}

func (*observedBody) Read([]byte) (int, error) { return 0, io.EOF }
func (b *observedBody) Close() error           { b.closed.Add(1); return b.closeErr }

func TestObservedTransportBeginFaultClosesBody(t *testing.T) {
	b, s := observedFixture(t)
	var calls atomic.Int32
	tr, url, _ := observedTLS(t, b, func(w http.ResponseWriter, r *http.Request) { calls.Add(1) })
	fault := errors.New("begin fault")
	closeFault := errors.New("body close fault")
	s.testFault = func(stage string) error {
		if stage == "state-write" {
			return fault
		}
		return nil
	}
	c := &ObservationCollector{}
	body := &observedBody{closeErr: closeFault}
	req, _ := http.NewRequestWithContext(WithObservation(context.Background(), ObservationOther, ObservationKey{}, c), http.MethodPost, url, body)
	_, err := tr.RoundTrip(req)
	if !errors.Is(err, ErrNotSent) || !errors.Is(err, fault) || !errors.Is(err, closeFault) || body.closed.Load() != 1 || calls.Load() != 0 || !errors.Is(c.Snapshot().LocalError, fault) || !errors.Is(c.Snapshot().LocalError, closeFault) {
		t.Fatalf("fault=%v closes=%d calls=%d collector=%+v", err, body.closed.Load(), calls.Load(), c.Snapshot())
	}
}
func TestObservedTransportPollRejectsPost(t *testing.T) {
	b, _ := observedFixture(t)
	var calls atomic.Int32
	tr, url, _ := observedTLS(t, b, func(w http.ResponseWriter, r *http.Request) { calls.Add(1) })
	var key ObservationKey
	key[0] = 1
	body := &observedBody{}
	c := &ObservationCollector{}
	req, _ := http.NewRequestWithContext(WithObservation(context.Background(), ObservationPoll, key, c), http.MethodPost, url, body)
	_, err := tr.RoundTrip(req)
	if !errors.Is(err, ErrNotSent) || body.closed.Load() != 1 || calls.Load() != 0 || c.Snapshot().LocalError == nil {
		t.Fatalf("err %v", err)
	}
}
func TestObservedRetryAfterParsing(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		input string
		valid bool
		want  time.Duration
	}{{"0", true, 0}, {"30", true, 30 * time.Second}, {"31", false, 0}, {"-1", false, 0}, {"+1", false, 0}, {"18446744073709551616", false, 0}, {"", false, 0}, {now.Add(2 * time.Second).Format(http.TimeFormat), true, 2 * time.Second}, {now.Add(-time.Second).Format(http.TimeFormat), false, 0}} {
		got := retryAfterHint(tc.input, now)
		if (got != nil) != tc.valid || got != nil && *got != tc.want {
			t.Errorf("%q: %v", tc.input, got)
		}
	}
}

func TestObservedTransportNoResponseAfterSent(t *testing.T) {
	b, s := observedFixture(t)
	tr, url, _ := observedTLS(t, b, func(w http.ResponseWriter, r *http.Request) {
		h, ok := w.(http.Hijacker)
		if !ok {
			t.Error("hijack unsupported")
			return
		}
		conn, _, err := h.Hijack()
		if err != nil {
			t.Error(err)
			return
		}
		_ = conn.Close()
	})
	c := &ObservationCollector{}
	req, _ := http.NewRequestWithContext(WithObservation(context.Background(), ObservationOther, ObservationKey{}, c), http.MethodGet, url, nil)
	_, err := tr.RoundTrip(req)
	if err == nil || errors.Is(err, ErrNotSent) {
		t.Fatalf("sent classification %v", err)
	}
	snap, e := s.Snapshot(context.Background(), ObservationCursor{})
	if e != nil || snap.Pending != nil || snap.Last == nil || snap.Last.Result != ObservationNoResponse {
		t.Fatalf("snapshot %+v err %v", snap, e)
	}
}
func TestObservedTransportInvalid429HintStillEvent(t *testing.T) {
	b, s := observedFixture(t)
	tr, url, _ := observedTLS(t, b, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "999999999999999999999999999")
		w.WriteHeader(429)
	})
	var key ObservationKey
	key[0] = 2
	c := &ObservationCollector{}
	req, _ := http.NewRequestWithContext(WithObservation(context.Background(), ObservationPoll, key, c), http.MethodGet, url, nil)
	resp, err := tr.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	remaining, err := s.PollRemaining(context.Background(), key)
	if err != nil || remaining != 0 {
		t.Fatalf("hint %v %v", remaining, err)
	}
	snap, err := s.Snapshot(context.Background(), ObservationCursor{})
	if err != nil || len(snap.Events) != 1 || !c.Snapshot().RateLimited {
		t.Fatalf("event %+v %v", snap, err)
	}
}

func TestObservedTransportCancelDuringBeginLeavesGap(t *testing.T) {
	b, s := observedFixture(t)
	var calls atomic.Int32
	tr, url, _ := observedTLS(t, b, func(w http.ResponseWriter, r *http.Request) { calls.Add(1) })
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var writes int
	s.testFault = func(stage string) error {
		if stage == "state-write" {
			writes++
			if writes == 1 {
				cancel()
			}
		}
		return nil
	}
	body := &observedBody{}
	c := &ObservationCollector{}
	req, _ := http.NewRequestWithContext(WithObservation(ctx, ObservationOther, ObservationKey{}, c), http.MethodPost, url, body)
	_, err := tr.RoundTrip(req)
	if !errors.Is(err, context.Canceled) || !errors.Is(err, ErrNotSent) || calls.Load() != 0 || body.closed.Load() != 1 {
		t.Fatalf("err %v calls %d closed %d", err, calls.Load(), body.closed.Load())
	}
	snap, e := s.Snapshot(context.Background(), ObservationCursor{})
	if e != nil || snap.Pending == nil || snap.PendingStatus != ObservationPendingOrphan {
		t.Fatalf("snapshot %+v err %v", snap, e)
	}
}

type cancelBoot struct {
	source clock.BootSource
	cancel context.CancelFunc
}

func (b cancelBoot) Snapshot() (clock.BootStamp, error) { b.cancel(); return b.source.Snapshot() }
func TestObservedTransportCancelAfterBeginSettles(t *testing.T) {
	b, s := observedFixture(t)
	var calls atomic.Int32
	tr, url, _ := observedTLS(t, b, func(w http.ResponseWriter, r *http.Request) { calls.Add(1) })
	var key ObservationKey
	key[0] = 8
	lease, err := b.Acquire(context.Background(), Normal)
	if err != nil {
		t.Fatal(err)
	}
	seq, err := s.Begin(context.Background(), lease, ObservationPoll, key)
	if err != nil {
		t.Fatal(err)
	}
	hint := time.Second
	if err = s.Finish(context.Background(), lease, seq, ObservationRateLimited, &hint); err != nil {
		t.Fatal(err)
	}
	if err = lease.Release(); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.boot = cancelBoot{s.boot, cancel}
	body := &observedBody{}
	c := &ObservationCollector{}
	req, _ := http.NewRequestWithContext(WithObservation(ctx, ObservationPoll, key, c), http.MethodGet, url, body)
	_, err = tr.RoundTrip(req)
	if !errors.Is(err, ErrNotSent) || !errors.Is(err, context.Canceled) || !errors.Is(c.Snapshot().LocalError, context.Canceled) || calls.Load() != 0 || body.closed.Load() != 1 {
		t.Fatalf("err %v calls %d closed %d", err, calls.Load(), body.closed.Load())
	}
	snap, e := s.Snapshot(context.Background(), ObservationCursor{})
	if e != nil || snap.Pending != nil || snap.Last == nil || snap.Last.Result != ObservationNoResponse {
		t.Fatalf("snapshot %+v err %v", snap, e)
	}
}

func TestObservedTransportPreSendReleaseError(t *testing.T) {
	b, _ := observedFixture(t)
	tr, url, _ := observedTLS(t, b, func(w http.ResponseWriter, r *http.Request) { t.Error("unexpected wire") })
	releaseErr := errors.New("release error")
	b.testWireRelease = func(f *os.File) error { return errors.Join(release(f), releaseErr) }
	var key ObservationKey
	key[0] = 1
	c := &ObservationCollector{}
	req, _ := http.NewRequestWithContext(WithObservation(context.Background(), ObservationPoll, key, c), http.MethodPost, url, nil)
	_, err := tr.RoundTrip(req)
	if !errors.Is(err, ErrNotSent) || !errors.Is(err, releaseErr) || !errors.Is(c.Snapshot().LocalError, releaseErr) {
		t.Fatalf("error %v collector %+v", err, c.Snapshot())
	}
}

func TestObserved429BodyFailureRetainsEvent(t *testing.T) {
	b, s := observedFixture(t)
	tr, url, _ := observedTLS(t, b, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "10")
		w.WriteHeader(429)
		_, _ = w.Write([]byte("x"))
	})
	c := &ObservationCollector{}
	req, _ := http.NewRequestWithContext(WithObservation(context.Background(), ObservationOther, ObservationKey{}, c), http.MethodGet, url, nil)
	resp, err := tr.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	_, err = io.ReadAll(resp.Body)
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("body error %v", err)
	}
	snap, err := s.Snapshot(context.Background(), ObservationCursor{})
	if err != nil || len(snap.Events) != 1 || !c.Snapshot().RateLimited {
		t.Fatalf("event %+v %v", snap, err)
	}
}
