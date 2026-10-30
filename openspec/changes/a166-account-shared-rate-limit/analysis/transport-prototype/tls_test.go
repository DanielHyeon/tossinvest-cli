package prototype

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

type lease struct {
	mu       sync.Mutex
	closed   bool
	conn     net.Conn
	state    *tls.ConnectionState
	starts   *[]time.Time
	logMu    *sync.Mutex
	once     bool
	pause    chan struct{}
	resume   chan struct{}
	evidence string
	inFlight *atomic.Int32
}

func (l *lease) register(c net.Conn) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		c.Close()
		return false
	}
	l.conn = c
	if tc, ok := c.(*tls.Conn); ok {
		state := tc.ConnectionState()
		l.state = &state
	}
	return true
}
func (l *lease) revoke() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.closed = true
}

type guarded struct {
	net.Conn
	l *lease
}

func (c *guarded) Write(p []byte) (int, error) {
	c.l.mu.Lock()
	defer c.l.mu.Unlock()
	if c.l.closed {
		return 0, io.ErrClosedPipe
	}
	if !c.l.once {
		if c.l.pause != nil {
			close(c.l.pause)
			<-c.l.resume
		}
		c.l.logMu.Lock()
		*c.l.starts = append(*c.l.starts, time.Now())
		c.l.logMu.Unlock()
		if c.l.evidence != "" {
			f, e := os.OpenFile(c.l.evidence, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
			if e != nil {
				return 0, e
			}
			_, e = fmt.Fprintf(f, "%d\n", time.Now().UnixNano())
			f.Close()
			if e != nil {
				return 0, e
			}
		}
		c.l.once = true
	}
	if c.l.inFlight != nil {
		c.l.inFlight.Add(1)
		defer c.l.inFlight.Add(-1)
	}
	return c.Conn.Write(p)
}

type limiter struct {
	inFlight    atomic.Int32
	gateOnce    sync.Once
	gate        chan struct{}
	logMu       sync.Mutex
	starts      []time.Time
	tls         *tls.Config
	beforeWrite chan struct{}
	resume      chan struct{}
	dialStart   chan struct{}
	dialResume  chan struct{}
	lockPath    string
	evidence    string
}

func (l *limiter) RoundTrip(req *http.Request) (*http.Response, error) {
	handedOff := false
	defer func() {
		if !handedOff && req.Body != nil {
			req.Body.Close()
		}
	}()
	l.gateOnce.Do(func() { l.gate = make(chan struct{}, 1) })
	var f *os.File
	if l.lockPath != "" {
		var e error
		f, e = os.OpenFile(l.lockPath, os.O_CREATE|os.O_RDWR, 0600)
		if e != nil {
			return nil, e
		}
	}
	release, e := acquireContext(req.Context(), l.gate, f)
	if e != nil {
		if f != nil {
			f.Close()
		}
		return nil, e
	}

	lease := &lease{starts: &l.starts, logMu: &l.logMu, pause: l.beforeWrite, resume: l.resume, evidence: l.evidence, inFlight: &l.inFlight}
	tr := &http.Transport{DisableKeepAlives: true, ForceAttemptHTTP2: false, TLSNextProto: map[string]func(string, *tls.Conn) http.RoundTripper{}, Proxy: nil}
	tr.DialTLSContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		if l.dialStart != nil {
			close(l.dialStart)
			<-l.dialResume
		}
		raw, err := (&net.Dialer{}).DialContext(ctx, network, addr)
		if err != nil {
			return nil, err
		}
		cfg := l.tls.Clone()
		if cfg.ServerName == "" {
			cfg.ServerName, _, err = net.SplitHostPort(addr)
			if err != nil {
				raw.Close()
				return nil, err
			}
		}
		tc := tls.Client(raw, cfg)
		if err = tc.HandshakeContext(ctx); err != nil {
			raw.Close()
			return nil, err
		}
		if !lease.register(tc) {
			return nil, io.ErrClosedPipe
		}
		return &guarded{Conn: tc, l: lease}, nil
	}

	defer tr.CloseIdleConnections()
	handedOff = true
	resp, err := tr.RoundTrip(req)
	if resp != nil && resp.TLS == nil {
		lease.mu.Lock()
		if lease.state != nil && len(lease.state.VerifiedChains) > 0 {
			state := *lease.state
			resp.TLS = &state
		}
		lease.mu.Unlock()
	}
	if resp == nil {
		lease.revoke()
		release()
		if f != nil {
			f.Close()
		}
		return nil, err
	}
	owner := &responseOwner{body: resp.Body, done: make(chan struct{}), finish: func() {
		lease.revoke()
		release()
		if f != nil {
			f.Close()
		}
	}}
	resp.Body = owner
	go func() {
		select {
		case <-req.Context().Done():
			owner.Close()
		case <-owner.done:
		}
	}()
	return resp, err
}

type responseOwner struct {
	body     io.ReadCloser
	once     sync.Once
	done     chan struct{}
	finish   func()
	closeErr error
}

func (o *responseOwner) Close() error {
	o.once.Do(func() { o.closeErr = o.body.Close(); o.finish(); close(o.done) })
	return o.closeErr
}
func (o *responseOwner) Read(p []byte) (int, error) {
	n, e := o.body.Read(p)
	if e != nil {
		o.Close()
	}
	return n, e
}
func TestTLSRedirectAnd401AttemptsAreSpaced(t *testing.T) {
	var mu sync.Mutex
	count := 0
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		count++
		n := count
		mu.Unlock()
		if n == 1 {
			w.Header().Set("Location", "/again")
			w.WriteHeader(307)
			return
		}
		if n == 2 {
			w.WriteHeader(401)
			return
		}
		w.WriteHeader(200)
	}))
	defer s.Close()
	lim := &limiter{tls: s.Client().Transport.(*http.Transport).TLSClientConfig}
	client := &http.Client{Transport: lim, Timeout: 5 * time.Second}
	resp, err := client.Get(s.URL)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 401 {
		t.Fatalf("status %d", resp.StatusCode)
	}
	resp, err = client.Get(s.URL)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	lim.logMu.Lock()
	defer lim.logMu.Unlock()
	if len(lim.starts) != 3 {
		t.Fatalf("starts=%d", len(lim.starts))
	}
	for i := 1; i < len(lim.starts); i++ {
		if lim.starts[i].Sub(lim.starts[i-1]) < 350*time.Millisecond {
			t.Fatal("spacing")
		}
	}
	mu.Lock()
	defer mu.Unlock()
	if count != 3 {
		t.Fatalf("requests=%d", count)
	}
}
func TestLateDialAfterCancelCannotWrite(t *testing.T) {
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { t.Error("unexpected request") }))
	defer s.Close()
	lim := &limiter{tls: s.Client().Transport.(*http.Transport).TLSClientConfig, dialStart: make(chan struct{}), dialResume: make(chan struct{})}
	ctx, cancel := context.WithCancel(context.Background())
	req, _ := http.NewRequestWithContext(ctx, "GET", s.URL, nil)
	done := make(chan struct{})
	go func() { lim.RoundTrip(req); close(done) }()
	<-lim.dialStart
	cancel()
	<-done
	close(lim.dialResume)
	time.Sleep(100 * time.Millisecond)
	lim.logMu.Lock()
	defer lim.logMu.Unlock()
	if len(lim.starts) != 0 {
		t.Fatal("late dial wrote")
	}
}

func TestPausedFirstWritePreventsNextAttempt(t *testing.T) {
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }))
	defer s.Close()
	lim := &limiter{tls: s.Client().Transport.(*http.Transport).TLSClientConfig, beforeWrite: make(chan struct{}), resume: make(chan struct{})}
	req, _ := http.NewRequest("GET", s.URL, nil)
	done := make(chan struct{})
	go func() {
		resp, _ := lim.RoundTrip(req)
		if resp != nil {
			resp.Body.Close()
		}
		close(done)
	}()
	<-lim.beforeWrite
	next := make(chan struct{})
	go func() { lim.gate <- struct{}{}; <-lim.gate; close(next) }()
	select {
	case <-next:
		t.Fatal("lock released during paused write")
	case <-time.After(50 * time.Millisecond):
	}
	close(lim.resume)
	<-done
	<-next
}

type slowBody struct {
	entered chan struct{}
	release chan struct{}
	once    sync.Once
}

func (b *slowBody) Read(p []byte) (int, error) {
	b.once.Do(func() { close(b.entered) })
	<-b.release
	return 0, io.EOF
}
func (b *slowBody) Close() error { return nil }
func TestEarlyResponseWithSlowBody(t *testing.T) {
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(401); w.(http.Flusher).Flush() }))
	defer s.Close()
	lim := &limiter{tls: s.Client().Transport.(*http.Transport).TLSClientConfig}
	body := &slowBody{entered: make(chan struct{}), release: make(chan struct{})}
	req, _ := http.NewRequest("POST", s.URL, body)
	req.ContentLength = -1
	done := make(chan struct{})
	go func() {
		resp, _ := lim.RoundTrip(req)
		if resp != nil {
			resp.Body.Close()
		}
		close(done)
	}()
	<-body.entered
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		close(body.release)
		<-done
		t.Fatal("early response did not return")
	}
	close(body.release)
}

func TestChunkedResponseBodySurvivesRevoke(t *testing.T) {
	data := strings.Repeat("body-data-", 50000)
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(data[:100]))
		w.(http.Flusher).Flush()
		time.Sleep(50 * time.Millisecond)
		w.Write([]byte(data[100:]))
	}))
	defer s.Close()
	lim := &limiter{tls: s.Client().Transport.(*http.Transport).TLSClientConfig}
	client := &http.Client{Transport: lim, Timeout: 5 * time.Second}
	resp, err := client.Get(s.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil || string(b) != data {
		t.Fatalf("body len=%d want=%d err=%v", len(b), len(data), err)
	}
}

func TestTLSRejectsUntrustedAndWrongHost(t *testing.T) {
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { t.Error("untrusted request reached handler") }))
	defer s.Close()
	for _, cfg := range []*tls.Config{{}, {RootCAs: s.Client().Transport.(*http.Transport).TLSClientConfig.RootCAs, ServerName: "wrong.invalid"}} {
		lim := &limiter{tls: cfg}
		c := &http.Client{Transport: lim, Timeout: 3 * time.Second}
		resp, err := c.Get(s.URL)
		if resp != nil {
			resp.Body.Close()
		}
		if err == nil {
			t.Error("TLS verification accepted invalid certificate or name")
		}
		lim.logMu.Lock()
		if len(lim.starts) != 0 {
			t.Error("request write after failed TLS verification")
		}
		lim.logMu.Unlock()
	}
}

func TestFreshHTTP1FailedRequestIsNotRetried(t *testing.T) {
	var count atomic.Int32
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count.Add(1)
		c, _, err := w.(http.Hijacker).Hijack()
		if err != nil {
			t.Error(err)
			return
		}
		c.Close()
	}))
	defer s.Close()
	l := &limiter{tls: s.Client().Transport.(*http.Transport).TLSClientConfig}
	resp, err := (&http.Client{Transport: l, Timeout: 3 * time.Second}).Get(s.URL)
	if resp != nil {
		resp.Body.Close()
	}
	if err == nil {
		t.Fatal("expected connection error")
	}
	if count.Load() != 1 {
		t.Fatalf("requests=%d", count.Load())
	}
	l.logMu.Lock()
	defer l.logMu.Unlock()
	if len(l.starts) != 1 {
		t.Fatalf("writes=%d", len(l.starts))
	}
}

func TestIntegratedCancellationBeforeWrite(t *testing.T) {
	for _, stage := range []string{"local", "flock", "cooldown"} {
		t.Run(stage, func(t *testing.T) {
			var received atomic.Int32
			s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { received.Add(1) }))
			defer s.Close()
			l := &limiter{tls: s.Client().Transport.(*http.Transport).TLSClientConfig}
			var other *os.File
			if stage == "local" {
				l.gateOnce.Do(func() { l.gate = make(chan struct{}, 1) })
				l.gate <- struct{}{}
				defer func() { <-l.gate }()
			}
			if stage == "flock" {
				l.lockPath = filepath.Join(t.TempDir(), "lock")
				var e error
				other, e = os.OpenFile(l.lockPath, os.O_CREATE|os.O_RDWR, 0600)
				if e != nil {
					t.Fatal(e)
				}
				defer other.Close()
				if e = syscall.Flock(int(other.Fd()), syscall.LOCK_EX); e != nil {
					t.Fatal(e)
				}
				defer syscall.Flock(int(other.Fd()), syscall.LOCK_UN)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()
			req, _ := http.NewRequestWithContext(ctx, "GET", s.URL, nil)
			resp, err := l.RoundTrip(req)
			if resp != nil {
				resp.Body.Close()
			}
			if !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("err=%v", err)
			}
			l.logMu.Lock()
			defer l.logMu.Unlock()
			if len(l.starts) != 0 || received.Load() != 0 {
				t.Fatal("request sent after cancel")
			}
		})
	}
}

func TestResponseBodyCleanup(t *testing.T) {
	for _, mode := range []string{"eof", "close", "cancel"} {
		t.Run(mode, func(t *testing.T) {
			closed := make(chan struct{}, 1)
			hold := make(chan struct{})
			server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("head"))
				w.(http.Flusher).Flush()
				if mode != "eof" {
					select {
					case <-r.Context().Done():
					case <-hold:
					}
				} else {
					w.Write([]byte("tail"))
				}
			}))
			server.Config.ConnState = func(_ net.Conn, state http.ConnState) {
				if state == http.StateClosed {
					select {
					case closed <- struct{}{}:
					default:
					}
				}
			}
			server.StartTLS()
			defer server.Close()
			defer close(hold)
			l := &limiter{tls: server.Client().Transport.(*http.Transport).TLSClientConfig}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			req, _ := http.NewRequestWithContext(ctx, "GET", server.URL, nil)
			resp, err := (&http.Client{Transport: l}).Do(req)
			if err != nil {
				t.Fatal(err)
			}
			if mode == "eof" {
				b, err := io.ReadAll(resp.Body)
				if err != nil || string(b) != "headtail" {
					t.Fatalf("body=%q err=%v", b, err)
				}
			}
			if mode == "cancel" {
				cancel()
			}
			resp.Body.Close()
			select {
			case <-closed:
			case <-time.After(3 * time.Second):
				t.Fatal("connection not closed")
			}
		})
	}
}

func TestRevokedTLSWriteFails(t *testing.T) {
	a, b := net.Pipe()
	defer a.Close()
	defer b.Close()
	go io.Copy(io.Discard, b)
	starts := []time.Time{}
	var mu sync.Mutex
	l := &lease{starts: &starts, logMu: &mu}
	if !l.register(a) {
		t.Fatal("register")
	}
	l.revoke()
	c := &guarded{Conn: a, l: l}
	if _, err := c.Write([]byte("x")); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatalf("err=%v", err)
	}
	if len(starts) != 0 {
		t.Fatal("recorded write after revoke")
	}
}

type trackedBody struct {
	io.Reader
	closed atomic.Int32
}

func (b *trackedBody) Close() error { b.closed.Add(1); return nil }
func TestPreAdmissionCancelClosesBody(t *testing.T) {
	for _, stage := range []string{"local", "flock", "cooldown"} {
		t.Run(stage, func(t *testing.T) {
			s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { t.Error("unexpected request") }))
			defer s.Close()
			l := &limiter{tls: s.Client().Transport.(*http.Transport).TLSClientConfig}
			var other *os.File
			if stage == "local" {
				l.gateOnce.Do(func() { l.gate = make(chan struct{}, 1) })
				l.gate <- struct{}{}
				defer func() { <-l.gate }()
			}
			if stage == "flock" {
				l.lockPath = filepath.Join(t.TempDir(), "lock")
				var e error
				other, e = os.OpenFile(l.lockPath, os.O_CREATE|os.O_RDWR, 0600)
				if e != nil {
					t.Fatal(e)
				}
				defer other.Close()
				if e = syscall.Flock(int(other.Fd()), syscall.LOCK_EX); e != nil {
					t.Fatal(e)
				}
				defer syscall.Flock(int(other.Fd()), syscall.LOCK_UN)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()
			body := &trackedBody{Reader: strings.NewReader("payload")}
			req, _ := http.NewRequestWithContext(ctx, "POST", s.URL, body)
			resp, err := l.RoundTrip(req)
			if resp != nil {
				resp.Body.Close()
			}
			if !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("err=%v", err)
			}
			if body.closed.Load() != 1 {
				t.Fatalf("close calls=%d", body.closed.Load())
			}
			l.logMu.Lock()
			defer l.logMu.Unlock()
			if len(l.starts) != 0 {
				t.Fatal("write after cancel")
			}
		})
	}
}

func TestCancelBlockedTLSWriteDrainsAndAllowsNext(t *testing.T) {
	entered := make(chan struct{})
	hold := make(chan struct{})
	var count atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if count.Add(1) == 1 {
			close(entered)
			select {
			case <-r.Context().Done():
			case <-hold:
			}
			return
		}
		fmt.Fprint(w, "ok")
	}))
	defer server.Close()
	defer close(hold)
	l := &limiter{tls: server.Client().Transport.(*http.Transport).TLSClientConfig}
	ctx, cancel := context.WithCancel(context.Background())
	body := &trackedBody{Reader: io.LimitReader(zeroReader{}, 128<<20)}
	req, _ := http.NewRequestWithContext(ctx, "POST", server.URL, body)
	req.ContentLength = 128 << 20
	done := make(chan error, 1)
	go func() {
		resp, err := l.RoundTrip(req)
		if resp != nil {
			resp.Body.Close()
		}
		done <- err
	}()
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		cancel()
		t.Fatal("server did not receive request")
	}
	select {
	case err := <-done:
		t.Fatalf("expected blocked write, returned %v", err)
	case <-time.After(200 * time.Millisecond):
	}
	if l.inFlight.Load() != 1 {
		cancel()
		t.Fatal("no TLS Write blocked at cancellation")
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("err=%v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("cancel did not drain write")
	}
	req2, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := l.RoundTrip(req2)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil || string(b) != "ok" {
		t.Fatalf("body=%q err=%v", b, err)
	}
	if body.closed.Load() != 1 {
		t.Fatalf("body closes=%d", body.closed.Load())
	}
}

type zeroReader struct{}

func (zeroReader) Read(p []byte) (int, error) { clear(p); return len(p), nil }

func TestCancelDuringBlockedResponseRead(t *testing.T) {
	hold := make(chan struct{})
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "first"); w.(http.Flusher).Flush(); <-hold }))
	defer s.Close()
	defer close(hold)
	l := &limiter{tls: s.Client().Transport.(*http.Transport).TLSClientConfig}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, "GET", s.URL, nil)
	resp, err := (&http.Client{Transport: l}).Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b := make([]byte, 5)
	if _, err = io.ReadFull(resp.Body, b); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { _, err := resp.Body.Read(b); done <- err }()
	select {
	case err := <-done:
		t.Fatalf("expected blocked read, got %v", err)
	case <-time.After(100 * time.Millisecond):
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("read error=%v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("read did not cancel")
	}
}

func TestResponseOwnershipBlocksNextUntilFinished(t *testing.T) {
	for _, mode := range []string{"eof", "close", "cancel", "client_timeout"} {
		t.Run(mode, func(t *testing.T) {
			var calls atomic.Int32
			secondEntered := make(chan struct{})
			bodyRelease := make(chan struct{})
			var releaseOnce sync.Once
			releaseBody := func() { releaseOnce.Do(func() { close(bodyRelease) }) }
			s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if calls.Add(1) == 1 {
					fmt.Fprint(w, "first")
					w.(http.Flusher).Flush()
					<-bodyRelease
					fmt.Fprint(w, "last")
					return
				}
				close(secondEntered)
				fmt.Fprint(w, "ok")
			}))
			defer s.Close()
			defer releaseBody()
			l := &limiter{tls: s.Client().Transport.(*http.Transport).TLSClientConfig}
			c := &http.Client{Transport: l}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if mode == "client_timeout" {
				c.Timeout = 2 * time.Second
			}
			req, _ := http.NewRequestWithContext(ctx, "GET", s.URL, nil)
			first, err := c.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			done := make(chan error, 1)
			go func() {
				resp, err := (&http.Client{Transport: l, Timeout: 5 * time.Second}).Get(s.URL)
				if resp != nil {
					io.Copy(io.Discard, resp.Body)
					resp.Body.Close()
				}
				done <- err
			}()
			select {
			case <-secondEntered:
				t.Fatal("second request during first body")
			case <-time.After(500 * time.Millisecond):
			}
			switch mode {
			case "eof":
				releaseBody()
				b, err := io.ReadAll(first.Body)
				if err != nil || string(b) != "firstlast" {
					t.Fatalf("body=%q err=%v", b, err)
				}
			case "close":
				first.Body.Close()
			case "cancel":
				cancel()
			case "client_timeout":
			}
			select {
			case err := <-done:
				if err != nil {
					t.Fatal(err)
				}
			case <-time.After(4 * time.Second):
				t.Fatal("second request remained blocked")
			}
			first.Body.Close()
		})
	}
}
