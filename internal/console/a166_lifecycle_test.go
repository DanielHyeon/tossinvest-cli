package console

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"sync/atomic"
	"testing"
	"time"
)

func TestA166FinishedFlagDoesNotReplaceLiveWorker(t *testing.T) {
	done := make(chan struct{})
	old := &runState{finishedFlag: true, done: done}
	c := &Console{run: old, now: time.Now}
	got, err := c.startRun("KR", nil)
	if !errors.Is(err, errRunInProgress) || got != old {
		t.Fatalf("live worker replaced: got=%p old=%p err=%v", got, old, err)
	}
	close(done)
}

func TestA166RelaunchWaitsForActualWorkerDone(t *testing.T) {
	done := make(chan struct{})
	var called atomic.Int32
	c := &Console{out: io.Discard, now: time.Now, handler: http.NotFoundHandler(), relaunch: make(chan int, 1), run: &runState{finishedFlag: true, done: done}, settleWait: 20 * time.Millisecond, shutdownWait: 100 * time.Millisecond, opts: Options{Relaunch: func(int) error { called.Add(1); return nil }}}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	c.relaunch <- 12345
	err = c.Serve(context.Background(), ln)
	if err == nil || called.Load() != 0 {
		t.Fatalf("unfinished worker: err=%v relaunch=%d", err, called.Load())
	}
	close(done)
}

func TestA166ShutdownTimeoutPreventsRelaunch(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	defer close(release)
	var called atomic.Int32
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { close(started); <-release })
	c := &Console{out: io.Discard, now: time.Now, handler: handler, relaunch: make(chan int, 1), settleWait: 20 * time.Millisecond, shutdownWait: 30 * time.Millisecond, opts: Options{Relaunch: func(int) error { called.Add(1); return nil }}}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	result := make(chan error, 1)
	go func() { result <- c.Serve(context.Background(), ln) }()
	go func() {
		resp, err := http.Get("http://" + ln.Addr().String())
		if err == nil {
			resp.Body.Close()
		}
	}()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("handler did not start")
	}
	c.relaunch <- 12345
	select {
	case err = <-result:
	case <-time.After(time.Second):
		t.Fatal("serve did not terminate")
	}
	if !errors.Is(err, context.DeadlineExceeded) || called.Load() != 0 {
		t.Fatalf("shutdown err=%v relaunch=%d", err, called.Load())
	}
}

type a166FailListener struct{ err error }

func (l a166FailListener) Accept() (net.Conn, error) { return nil, l.err }
func (l a166FailListener) Close() error              { return nil }
func (l a166FailListener) Addr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 12345}
}

func TestA166ServerFailureStillSettlesWorker(t *testing.T) {
	sentinel := errors.New("fake listener failure")
	done := make(chan struct{})
	defer close(done)
	c := &Console{out: io.Discard, now: time.Now, handler: http.NotFoundHandler(), relaunch: make(chan int, 1), run: &runState{finishedFlag: true, done: done}, settleWait: 20 * time.Millisecond, shutdownWait: 100 * time.Millisecond}
	err := c.Serve(context.Background(), a166FailListener{err: sentinel})
	if !errors.Is(err, sentinel) || err == sentinel {
		t.Fatalf("worker timeout and server error not preserved: %v", err)
	}
}

func TestA166RelaunchAfterWorkerDone(t *testing.T) {
	done := make(chan struct{})
	var called atomic.Int32
	c := &Console{out: io.Discard, now: time.Now, handler: http.NotFoundHandler(), relaunch: make(chan int, 1), run: &runState{finishedFlag: true, done: done}, settleWait: time.Second, shutdownWait: time.Second, opts: Options{Relaunch: func(int) error { called.Add(1); return nil }}}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	c.relaunch <- 12345
	go func() { time.Sleep(20 * time.Millisecond); close(done) }()
	if err = c.Serve(context.Background(), ln); err != nil || called.Load() != 1 {
		t.Fatalf("completed worker: err=%v relaunch=%d", err, called.Load())
	}
}

func TestA166ActiveHandlerFinishesBeforeOwnerCloseAndExec(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	events := make(chan string, 3)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { close(started); <-release; events <- "handler-done" })
	c := &Console{out: io.Discard, now: time.Now, handler: handler, relaunch: make(chan int, 1), settleWait: time.Second, shutdownWait: time.Second, opts: Options{Relaunch: func(int) error { events <- "owner-close"; events <- "fake-exec"; return nil }}}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	result := make(chan error, 1)
	go func() { result <- c.Serve(context.Background(), ln) }()
	go func() {
		resp, err := http.Get("http://" + ln.Addr().String())
		if err == nil {
			resp.Body.Close()
		}
	}()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("handler did not start")
	}
	c.relaunch <- 12345
	select {
	case err := <-result:
		t.Fatalf("returned before handler release: %v", err)
	case <-time.After(30 * time.Millisecond):
	}
	close(release)
	select {
	case err := <-result:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("serve did not finish")
	}
	for _, want := range []string{"handler-done", "owner-close", "fake-exec"} {
		if got := <-events; got != want {
			t.Fatalf("got %q want %q", got, want)
		}
	}
}
