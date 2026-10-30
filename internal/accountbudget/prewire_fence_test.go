//go:build linux

package accountbudget

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sys/unix"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestA169FenceBaselineAndEvents(t *testing.T) {
	b, s := observedFixture(t)
	ctx := context.Background()
	lease, e := b.Acquire(ctx, Normal)
	if e != nil {
		t.Fatal(e)
	}
	seq, e := s.Begin(ctx, lease, ObservationOther, ObservationKey{})
	if e != nil {
		t.Fatal(e)
	}
	if e = s.Finish(ctx, lease, seq, ObservationRateLimited, nil); e != nil {
		t.Fatal(e)
	}
	if e = lease.Release(); e != nil {
		t.Fatal(e)
	}
	f, e := b.NewObservationFence(ctx)
	if e != nil {
		t.Fatal(e)
	}
	lease, e = b.Acquire(ctx, Normal)
	if e != nil {
		t.Fatal(e)
	}
	seq, e = s.Begin(ctx, lease, ObservationOther, ObservationKey{})
	if e != nil {
		t.Fatal(e)
	}
	if e = s.Finish(ctx, lease, seq, ObservationResponse, nil); e != nil {
		t.Fatal(e)
	}
	if e = lease.Release(); e != nil {
		t.Fatal(e)
	}
	lease, e = b.Acquire(ctx, Normal)
	if e != nil {
		t.Fatal(e)
	}
	if e = f.check(ctx, b, lease); e != nil {
		t.Fatalf("historical event: %v", e)
	}
	seq, e = s.Begin(ctx, lease, ObservationOther, ObservationKey{})
	if e != nil {
		t.Fatal(e)
	}
	if e = s.Finish(ctx, lease, seq, ObservationRateLimited, nil); e != nil {
		t.Fatal(e)
	}
	if e = lease.Release(); e != nil {
		t.Fatal(e)
	}
	lease, e = b.Acquire(ctx, Normal)
	if e != nil {
		t.Fatal(e)
	}
	if e = f.check(ctx, b, lease); !errors.Is(e, ErrUnavailable) {
		t.Fatalf("new event %v", e)
	}
	if e = lease.Release(); e != nil {
		t.Fatal(e)
	}
}
func TestA169FenceInvalidOwnerAndPending(t *testing.T) {
	b, s := observedFixture(t)
	ctx := context.Background()
	if _, e := b.NewObservationFence(nil); !errors.Is(e, ErrUnavailable) {
		t.Fatalf("nil ctx %v", e)
	}
	var f ObservationFence
	if _, e := f.Context(ctx); !errors.Is(e, ErrUnavailable) {
		t.Fatalf("zero fence %v", e)
	}
	l, e := b.Acquire(ctx, Normal)
	if e != nil {
		t.Fatal(e)
	}
	_, e = s.Begin(ctx, l, ObservationOther, ObservationKey{})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = b.NewObservationFence(ctx); !errors.Is(e, ErrUnavailable) {
		t.Fatalf("pending %v", e)
	}
	if e = l.Release(); e != nil {
		t.Fatal(e)
	}
	if _, e = b.NewObservationFence(ctx); !errors.Is(e, ErrUnavailable) {
		t.Fatalf("orphan %v", e)
	}
}

func TestA169FenceTransportAfterEventAndOwn429(t *testing.T) {
	b, s := observedFixture(t)
	var calls atomic.Int32
	tr, url, _ := observedTLS(t, b, func(w http.ResponseWriter, _ *http.Request) { calls.Add(1); w.WriteHeader(429) })
	f, e := b.NewObservationFence(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	ctx, e := f.Context(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, e := tr.RoundTrip(req)
	if e != nil {
		t.Fatal(e)
	}
	if resp.StatusCode != 429 {
		t.Fatal(resp.StatusCode)
	}
	if e = resp.Body.Close(); e != nil {
		t.Fatal(e)
	}
	req, _ = http.NewRequestWithContext(WithObservation(ctx, ObservationOther, ObservationKey{}, &ObservationCollector{}), http.MethodGet, url, nil)
	_, e = tr.RoundTrip(req)
	if !errors.Is(e, ErrNotSent) || calls.Load() != 1 {
		t.Fatalf("sticky e=%v calls=%d", e, calls.Load())
	}
	f2, e := b.NewObservationFence(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	l, e := b.Acquire(context.Background(), Normal)
	if e != nil {
		t.Fatal(e)
	}
	seq, e := s.Begin(context.Background(), l, ObservationOther, ObservationKey{})
	if e != nil {
		t.Fatal(e)
	}
	if e = s.Finish(context.Background(), l, seq, ObservationRateLimited, nil); e != nil {
		t.Fatal(e)
	}
	if e = l.Release(); e != nil {
		t.Fatal(e)
	}
	ctx, e = f2.Context(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	req, _ = http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	_, e = tr.RoundTrip(req)
	if !errors.Is(e, ErrNotSent) || calls.Load() != 1 {
		t.Fatalf("new event e=%v calls=%d", e, calls.Load())
	}
}

func TestA169FenceReleaseFailureConcurrentWaiter(t *testing.T) {
	b, _ := observedFixture(t)
	var calls atomic.Int32
	tr, url, _ := observedTLS(t, b, func(w http.ResponseWriter, _ *http.Request) { calls.Add(1); w.WriteHeader(200) })
	f, e := b.NewObservationFence(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	ctx, e = f.Context(ctx)
	if e != nil {
		t.Fatal(e)
	}
	var initial, acquired atomic.Int32
	initialB := make(chan struct{})
	acquiredB := make(chan struct{})
	tr.testFenceInitial = func() {
		if initial.Add(1) == 2 {
			close(initialB)
		}
	}
	tr.testFenceAcquired = func() {
		if acquired.Add(1) == 2 {
			close(acquiredB)
		}
	}
	marker := errors.New("injected release error")
	unlocked := make(chan struct{})
	allow := make(chan struct{})
	var unblockOnce sync.Once
	unblock := func() { unblockOnce.Do(func() { close(allow) }) }
	defer unblock()
	var releases atomic.Int32
	b.testWireRelease = func(file *os.File) error {
		err := release(file)
		if releases.Add(1) == 1 {
			close(unlocked)
			<-allow
			return errors.Join(err, marker)
		}
		return err
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, e := tr.RoundTrip(req)
	if e != nil {
		t.Fatal(e)
	}
	bdone := make(chan error, 1)
	go func() {
		r, _ := http.NewRequestWithContext(WithObservation(ctx, ObservationOther, ObservationKey{}, &ObservationCollector{}), http.MethodGet, url, nil)
		rr, err := tr.RoundTrip(r)
		if rr != nil {
			err = errors.Join(err, rr.Body.Close())
		}
		bdone <- err
	}()
	wait := func(ch <-chan struct{}, name string) {
		t.Helper()
		select {
		case <-ch:
		case <-ctx.Done():
			t.Fatalf("timeout %s", name)
		}
	}
	wait(initialB, "B initial")
	adone := make(chan error, 1)
	go func() { adone <- resp.Body.Close() }()
	wait(unlocked, "unlock")
	wait(acquiredB, "B acquired")
	if f.mu.TryLock() {
		f.mu.Unlock()
		unblock()
		t.Fatal("release failure not serialized with B")
	}
	unblock()
	if e = <-adone; !errors.Is(e, marker) {
		t.Fatalf("close error %v", e)
	}
	if e = <-bdone; !errors.Is(e, ErrNotSent) || !errors.Is(e, marker) || calls.Load() != 1 {
		t.Fatalf("B %v calls=%d", e, calls.Load())
	}
}

func TestA169FencePostcommitFinishConcurrentWaiter(t *testing.T) {
	b, s := observedFixture(t)
	var calls atomic.Int32
	tr, url, _ := observedTLS(t, b, func(w http.ResponseWriter, _ *http.Request) { calls.Add(1); w.WriteHeader(200) })
	f, e := b.NewObservationFence(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	ctx, e = f.Context(ctx)
	if e != nil {
		t.Fatal(e)
	}
	marker := errors.New("finish postcommit error")
	entered := make(chan struct{})
	proceed := make(chan struct{})
	var syncs atomic.Int32
	s.syncDir = func(fd int) error {
		if syncs.Add(1) == 6 {
			close(entered)
			select {
			case <-proceed:
				return marker
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return unix.Fsync(fd)
	}
	initialB := make(chan struct{})
	var initial atomic.Int32
	tr.testFenceInitial = func() {
		if initial.Add(1) == 2 {
			close(initialB)
		}
	}
	type answer struct {
		resp *http.Response
		err  error
	}
	adone := make(chan answer, 1)
	go func() {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		r, err := tr.RoundTrip(req)
		adone <- answer{r, err}
	}()
	select {
	case <-entered:
	case <-ctx.Done():
		t.Fatal("finish barrier")
	}
	bdone := make(chan error, 1)
	go func() {
		req, _ := http.NewRequestWithContext(WithObservation(ctx, ObservationOther, ObservationKey{}, &ObservationCollector{}), http.MethodGet, url, nil)
		r, err := tr.RoundTrip(req)
		if r != nil {
			err = errors.Join(err, r.Body.Close())
		}
		bdone <- err
	}()
	select {
	case <-initialB:
	case <-ctx.Done():
		close(proceed)
		t.Fatal("B initial barrier")
	}
	close(proceed)
	var a answer
	select {
	case a = <-adone:
	case <-ctx.Done():
		t.Fatal("A response")
	}
	if a.err != nil || a.resp == nil || a.resp.StatusCode != 200 {
		t.Fatalf("response lost: %+v", a)
	}
	if e = a.resp.Body.Close(); e != nil {
		t.Fatal(e)
	}
	select {
	case e = <-bdone:
		if !errors.Is(e, ErrNotSent) || !errors.Is(e, marker) || calls.Load() != 1 {
			t.Fatalf("B=%v calls=%d", e, calls.Load())
		}
	case <-ctx.Done():
		t.Fatal("B result")
	}
}

type a169FailClose struct {
	err    error
	closes int
}

func (*a169FailClose) Read([]byte) (int, error) { return 0, io.EOF }
func (c *a169FailClose) Close() error           { c.closes++; return c.err }
func TestA169FenceBodyCloseBeforeRelease(t *testing.T) {
	b, _ := observedFixture(t)
	f, e := b.NewObservationFence(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	marker := errors.New("body close failure")
	body := &a169FailClose{err: marker}
	collector := &ObservationCollector{}
	count := 0
	owner := &responseOwner{body: &observationFenceBody{ReadCloser: body, fence: f, collector: collector}, done: make(chan struct{}), finish: func() error {
		count++
		if !errors.Is(f.initial(b), marker) {
			t.Error("body failure not recorded before release")
		}
		return nil
	}}
	if e = owner.Close(); !errors.Is(e, marker) {
		t.Fatal(e)
	}
	if e = owner.Close(); !errors.Is(e, marker) || body.closes != 1 || count != 1 || !errors.Is(collector.Snapshot().LocalError, marker) {
		t.Fatalf("close=%v body=%d finish=%d", e, body.closes, count)
	}
}
func TestA169FenceAuxBusyAndCancellationNotSticky(t *testing.T) {
	b, _ := observedFixture(t)
	tr, url, _ := observedTLS(t, b, func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) })
	f, e := b.NewObservationFence(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	ctx, e := f.Context(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	held, e := b.Acquire(ctx, Normal)
	if e != nil {
		t.Fatal(e)
	}
	req, _ := http.NewRequestWithContext(WithPriority(ctx, Auxiliary), http.MethodGet, url, nil)
	_, e = tr.RoundTrip(req)
	if !errors.Is(e, ErrNotSent) {
		t.Fatalf("aux busy %v", e)
	}
	if e = held.Release(); e != nil {
		t.Fatal(e)
	}
	started := make(chan struct{})
	dial := tr.base.DialContext
	tr.base.DialContext = func(c context.Context, _, _ string) (net.Conn, error) {
		close(started)
		<-c.Done()
		return nil, c.Err()
	}
	cancelCtx, cancel := context.WithCancel(ctx)
	done := make(chan error, 1)
	go func() {
		req, _ := http.NewRequestWithContext(cancelCtx, http.MethodGet, url, nil)
		_, err := tr.RoundTrip(req)
		done <- err
	}()
	<-started
	cancel()
	if e = <-done; !errors.Is(e, ErrNotSent) || !errors.Is(e, context.Canceled) {
		t.Fatalf("cancel %v", e)
	}
	tr.base.DialContext = dial
	req, _ = http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, e := tr.RoundTrip(req)
	if e != nil {
		t.Fatalf("after cancellation %v", e)
	}
	if e = resp.Body.Close(); e != nil {
		t.Fatal(e)
	}
}

func TestA169FenceHistoricalGapOverrunAndRollback(t *testing.T) {
	b, s := observedFixture(t)
	ctx := context.Background()
	l, e := b.Acquire(ctx, Normal)
	if e != nil {
		t.Fatal(e)
	}
	lock, e := s.lock(ctx, true)
	if e != nil {
		t.Fatal(e)
	}
	state, e := s.read()
	if e != nil {
		t.Fatal(e)
	}
	state.Attempt = 129
	state.Gap = 129
	state.Generation++
	state.Gaps = nil
	for i := uint64(2); i <= 129; i++ {
		state.Gaps = append(state.Gaps, ObservationGap{Sequence: i, Attempt: i, Kind: ObservationOther})
	}
	if e = s.write(state); e != nil {
		t.Fatal(e)
	}
	if e = release(lock); e != nil {
		t.Fatal(e)
	}
	if e = l.Release(); e != nil {
		t.Fatal(e)
	}
	view, e := b.ObservationSnapshot(ctx, ObservationCursor{})
	if e != nil || !view.GapOverrun {
		t.Fatalf("baseline overrun %+v %v", view, e)
	}
	f, e := b.NewObservationFence(ctx)
	if e != nil {
		t.Fatal(e)
	}
	l, e = b.Acquire(ctx, Normal)
	if e != nil {
		t.Fatal(e)
	}
	if e = f.check(ctx, b, l); e != nil {
		t.Fatalf("historical gap %v", e)
	}
	lock, e = s.lock(ctx, true)
	if e != nil {
		t.Fatal(e)
	}
	state, e = s.read()
	if e != nil {
		t.Fatal(e)
	}
	state.Attempt = 128
	state.Gap = 128
	state.Generation++
	state.Gaps = nil
	for i := uint64(1); i <= 128; i++ {
		state.Gaps = append(state.Gaps, ObservationGap{Sequence: i, Attempt: i, Kind: ObservationOther})
	}
	if e = s.write(state); e != nil {
		t.Fatal(e)
	}
	if e = release(lock); e != nil {
		t.Fatal(e)
	}
	if e = f.check(ctx, b, l); !errors.Is(e, ErrUnavailable) {
		t.Fatalf("rollback %v", e)
	}
	f2, e := b.NewObservationFence(ctx)
	if e != nil {
		t.Fatal(e)
	}
	lock, e = s.lock(ctx, true)
	if e != nil {
		t.Fatal(e)
	}
	state, e = s.read()
	if e != nil {
		t.Fatal(e)
	}
	state.Attempt++
	state.Gap++
	state.Generation++
	state.Gaps = append(state.Gaps[1:], ObservationGap{Sequence: state.Gap, Attempt: state.Attempt, Kind: ObservationOther})
	if e = s.write(state); e != nil {
		t.Fatal(e)
	}
	if e = release(lock); e != nil {
		t.Fatal(e)
	}
	if e = f2.check(ctx, b, l); !errors.Is(e, ErrUnavailable) {
		t.Fatalf("new gap %v", e)
	}
	if e = l.Release(); e != nil {
		t.Fatal(e)
	}
}
func TestA169FenceWrongBudgetAndDirty(t *testing.T) {
	a, _ := observedFixture(t)
	b, _ := observedFixture(t)
	f, e := a.NewObservationFence(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	ctx, e := f.Context(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	var count atomic.Int32
	tr, url, _ := observedTLS(t, b, func(w http.ResponseWriter, _ *http.Request) { count.Add(1) })
	body := &observedBody{}
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	_, e = tr.RoundTrip(req)
	if !errors.Is(e, ErrNotSent) || !errors.Is(e, ErrUnavailable) || count.Load() != 0 || body.closed.Load() != 1 {
		t.Fatalf("wrong budget %v count=%d closes=%d", e, count.Load(), body.closed.Load())
	}
	dirty := filepath.Join(a.root, observationDirty)
	if e = os.WriteFile(dirty, []byte("x"), 0600); e != nil {
		t.Fatal(e)
	}
	if _, e = a.NewObservationFence(context.Background()); !errors.Is(e, ErrUnavailable) {
		t.Fatalf("dirty %v", e)
	}
}

func TestA169FenceBeginFailureSticky(t *testing.T) {
	b, s := observedFixture(t)
	var calls atomic.Int32
	tr, url, _ := observedTLS(t, b, func(w http.ResponseWriter, _ *http.Request) { calls.Add(1) })
	f, e := b.NewObservationFence(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	ctx, e := f.Context(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	marker := errors.New("begin fault")
	s.syncDir = func(int) error { return marker }
	body := &observedBody{}
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	_, e = tr.RoundTrip(req)
	if !errors.Is(e, ErrNotSent) || !errors.Is(e, marker) || body.closed.Load() != 1 {
		t.Fatalf("first=%v closes=%d", e, body.closed.Load())
	}
	req, _ = http.NewRequestWithContext(WithObservation(ctx, ObservationOther, ObservationKey{}, &ObservationCollector{}), http.MethodGet, url, nil)
	_, e = tr.RoundTrip(req)
	if !errors.Is(e, ErrNotSent) || !errors.Is(e, marker) || calls.Load() != 0 {
		t.Fatalf("sticky=%v calls=%d", e, calls.Load())
	}
}

func TestA169FenceNilResponseCleanupSticky(t *testing.T) {
	b, _ := observedFixture(t)
	tr, url, _ := observedTLS(t, b, func(http.ResponseWriter, *http.Request) { t.Error("wire must stay zero") })
	f, e := b.NewObservationFence(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	ctx, e := f.Context(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	marker := errors.New("release failure")
	b.testWireRelease = func(file *os.File) error { return errors.Join(release(file), marker) }
	tr.base.DialContext = func(context.Context, string, string) (net.Conn, error) {
		return nil, errors.New("synthetic dial failure")
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	_, e = tr.RoundTrip(req)
	if !errors.Is(e, ErrNotSent) || !errors.Is(e, marker) {
		t.Fatalf("first %v", e)
	}
	req, _ = http.NewRequestWithContext(WithObservation(ctx, ObservationOther, ObservationKey{}, &ObservationCollector{}), http.MethodGet, url, nil)
	_, e = tr.RoundTrip(req)
	if !errors.Is(e, ErrNotSent) || !errors.Is(e, marker) {
		t.Fatalf("sticky %v", e)
	}
}

type a169ResponseCloseFault struct {
	io.ReadCloser
	err error
}

func (r *a169ResponseCloseFault) Close() error { return errors.Join(r.ReadCloser.Close(), r.err) }
func TestA169FenceBodyFailureConcurrentWaiter(t *testing.T) {
	b, _ := observedFixture(t)
	var calls atomic.Int32
	tr, url, _ := observedTLS(t, b, func(w http.ResponseWriter, _ *http.Request) { calls.Add(1); w.WriteHeader(200) })
	f, e := b.NewObservationFence(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	ctx, e = f.Context(ctx)
	if e != nil {
		t.Fatal(e)
	}
	initialB := make(chan struct{})
	var initial atomic.Int32
	tr.testFenceInitial = func() {
		if initial.Add(1) == 2 {
			close(initialB)
		}
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, e := tr.RoundTrip(req)
	if e != nil {
		t.Fatal(e)
	}
	owner, ok := resp.Body.(*responseOwner)
	if !ok {
		t.Fatal("response owner changed")
	}
	body, ok := owner.body.(*observationFenceBody)
	if !ok {
		t.Fatal("fence body changed")
	}
	marker := errors.New("response close error")
	body.ReadCloser = &a169ResponseCloseFault{ReadCloser: body.ReadCloser, err: marker}
	done := make(chan error, 1)
	go func() {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		r, err := tr.RoundTrip(req)
		if r != nil {
			err = errors.Join(err, r.Body.Close())
		}
		done <- err
	}()
	select {
	case <-initialB:
	case <-ctx.Done():
		t.Fatal("B initial")
	}
	if e = resp.Body.Close(); !errors.Is(e, marker) {
		t.Fatalf("close %v", e)
	}
	select {
	case e = <-done:
		if !errors.Is(e, ErrNotSent) || !errors.Is(e, marker) || calls.Load() != 1 {
			t.Fatalf("B=%v calls=%d", e, calls.Load())
		}
	case <-ctx.Done():
		t.Fatal("B result")
	}
}

func TestA169FenceBaselineWhileResponseOwnsWire(t *testing.T) {
	b, _ := observedFixture(t)
	var calls atomic.Int32
	tr, url, _ := observedTLS(t, b, func(w http.ResponseWriter, _ *http.Request) { calls.Add(1); fmt.Fprint(w, "held") })
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	resp, e := tr.RoundTrip(req)
	if e != nil {
		t.Fatal(e)
	}
	defer resp.Body.Close()
	f, e := b.NewObservationFence(context.Background())
	if e != nil {
		t.Fatalf("stable ledger snapshot while body held: %v", e)
	}
	ctx, e := f.Context(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	req, _ = http.NewRequestWithContext(WithPriority(ctx, Auxiliary), http.MethodGet, url, nil)
	_, e = tr.RoundTrip(req)
	if !errors.Is(e, ErrNotSent) || calls.Load() != 1 {
		t.Fatalf("held wire aux=%v calls=%d", e, calls.Load())
	}
	if e = resp.Body.Close(); e != nil {
		t.Fatal(e)
	}
	req, _ = http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	next, e := tr.RoundTrip(req)
	if e != nil {
		t.Fatal(e)
	}
	if e = next.Body.Close(); e != nil {
		t.Fatal(e)
	}
	if calls.Load() != 2 {
		t.Fatalf("calls=%d", calls.Load())
	}
}

func TestA169FenceEarlyBodyCloseFailureSticky(t *testing.T) {
	b, _ := observedFixture(t)
	tr, _, _ := observedTLS(t, b, func(http.ResponseWriter, *http.Request) { t.Error("wire must stay zero") })
	f, e := b.NewObservationFence(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	ctx, e := f.Context(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	marker := errors.New("early body close failure")
	collector := &ObservationCollector{}
	body := &observedBody{closeErr: marker}
	req, _ := http.NewRequestWithContext(WithObservation(ctx, ObservationOther, ObservationKey{}, collector), http.MethodPost, "http://127.0.0.1/", body)
	_, e = tr.RoundTrip(req)
	if !errors.Is(e, ErrNotSent) || !errors.Is(e, marker) || !errors.Is(collector.Snapshot().LocalError, marker) {
		t.Fatalf("first=%v collector=%v", e, collector.Snapshot().LocalError)
	}
	req, _ = http.NewRequestWithContext(ctx, http.MethodGet, "https://127.0.0.1/", nil)
	_, e = tr.RoundTrip(req)
	if !errors.Is(e, ErrNotSent) || !errors.Is(e, marker) {
		t.Fatalf("sticky=%v", e)
	}
}

func TestA169FenceGenerationRollbackAndClosed(t *testing.T) {
	b, s := observedFixture(t)
	ctx := context.Background()
	l, e := b.Acquire(ctx, Normal)
	if e != nil {
		t.Fatal(e)
	}
	seq, e := s.Begin(ctx, l, ObservationOther, ObservationKey{})
	if e != nil {
		t.Fatal(e)
	}
	if e = s.Finish(ctx, l, seq, ObservationResponse, nil); e != nil {
		t.Fatal(e)
	}
	if e = l.Release(); e != nil {
		t.Fatal(e)
	}
	f, e := b.NewObservationFence(ctx)
	if e != nil {
		t.Fatal(e)
	}
	l, e = b.Acquire(ctx, Normal)
	if e != nil {
		t.Fatal(e)
	}
	lock, e := s.lock(ctx, true)
	if e != nil {
		t.Fatal(e)
	}
	state, e := s.read()
	if e != nil {
		t.Fatal(e)
	}
	state.Generation = 0
	if e = s.write(state); e != nil {
		t.Fatal(e)
	}
	if e = release(lock); e != nil {
		t.Fatal(e)
	}
	if e = f.check(ctx, b, l); !errors.Is(e, ErrUnavailable) {
		t.Fatalf("generation rollback %v", e)
	}
	if e = l.Release(); e != nil {
		t.Fatal(e)
	}
	if e = b.Close(); e != nil {
		t.Fatal(e)
	}
	if _, e = b.NewObservationFence(ctx); !errors.Is(e, ErrUnavailable) {
		t.Fatalf("closed %v", e)
	}
}
