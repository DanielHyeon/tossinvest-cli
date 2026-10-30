package prototype

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
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
	return c.Conn.Write(p)
}

type limiter struct {
	mu          sync.Mutex
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
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.lockPath != "" {
		f, e := os.OpenFile(l.lockPath, os.O_CREATE|os.O_RDWR, 0600)
		if e != nil {
			return nil, e
		}
		defer f.Close()
		if e = syscall.Flock(int(f.Fd()), syscall.LOCK_EX); e != nil {
			return nil, e
		}
		defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	}
	time.Sleep(350 * time.Millisecond)
	lease := &lease{starts: &l.starts, logMu: &l.logMu, pause: l.beforeWrite, resume: l.resume, evidence: l.evidence}
	defer lease.revoke()
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
	resp, err := tr.RoundTrip(req)
	if resp != nil && resp.TLS == nil {
		lease.mu.Lock()
		if lease.state != nil && len(lease.state.VerifiedChains) > 0 {
			state := *lease.state
			resp.TLS = &state
		}
		lease.mu.Unlock()
	}
	return resp, err
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
	go func() { lim.mu.Lock(); lim.mu.Unlock(); close(next) }()
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
