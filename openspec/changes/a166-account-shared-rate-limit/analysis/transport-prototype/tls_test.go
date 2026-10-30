package prototype

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

type lease struct {
	mu     sync.Mutex
	closed bool
	conn   net.Conn
	starts *[]time.Time
	logMu  *sync.Mutex
	once   bool
	pause  chan struct{}
	resume chan struct{}
}

func (l *lease) register(c net.Conn) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		c.Close()
		return false
	}
	l.conn = c
	return true
}
func (l *lease) revoke() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.closed = true
	if l.conn != nil {
		l.conn.Close()
	}
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
}

func (l *limiter) RoundTrip(req *http.Request) (*http.Response, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	time.Sleep(350 * time.Millisecond)
	lease := &lease{starts: &l.starts, logMu: &l.logMu, pause: l.beforeWrite, resume: l.resume}
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
	return tr.RoundTrip(req)
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
