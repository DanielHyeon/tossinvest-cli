//go:build linux

package accountbudget

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
)

type priorityKey struct{}

// WithPriority attaches the explicitly chosen class to one request context.
func WithPriority(ctx context.Context, p Priority) context.Context {
	return context.WithValue(ctx, priorityKey{}, p)
}
func PriorityFromContext(ctx context.Context) Priority {
	p, ok := ctx.Value(priorityKey{}).(Priority)
	if !ok {
		return Normal
	}
	return p
}

type Transport struct {
	budget       *Budget
	base         *http.Transport
	cleanupError func(error)
	beforeStop   func(*connectionGate) // 시험에서 RoundTrip 반환 직후 경합을 동기화함.
	firstWrite   func()
	inflight     *atomic.Int32
}

// NewTransport does not grant trading authority. The official constructor owns
// that decision and must not unwrap a caller-supplied transport as authority.
// cleanupError receives close/revoke/release errors from asynchronous context
// cancellation after response handoff. It must be concurrency-safe and return
// promptly. Synchronous RoundTrip and Body.Close errors return to the caller.
func NewTransport(b *Budget, base *http.Transport, cleanupError func(error)) (*Transport, error) {
	if b == nil || base == nil || cleanupError == nil {
		return nil, ErrUnavailable
	}
	// 공식 constructor 소유 base의 지원 필드만 복사하며 원본 Clone을 호출하지 않음.
	if base.DialTLSContext != nil || base.DialTLS != nil || base.Dial != nil || base.TLSNextProto != nil || base.OnProxyConnectResponse != nil || base.GetProxyConnectHeader != nil || len(base.ProxyConnectHeader) != 0 || base.HTTP2 != nil || base.Protocols != nil {
		return nil, ErrUnavailable
	}
	var config *tls.Config
	if c := base.TLSClientConfig; c != nil {
		if c.InsecureSkipVerify || c.VerifyPeerCertificate != nil || c.VerifyConnection != nil || c.ServerName != "" || c.Time != nil || c.GetCertificate != nil || c.GetClientCertificate != nil || c.GetConfigForClient != nil || c.UnwrapSession != nil || c.WrapSession != nil || c.EncryptedClientHelloRejectionVerify != nil || c.GetEncryptedClientHelloKeys != nil {
			return nil, ErrUnavailable
		}
		if len(c.NextProtos) != 0 && !(len(c.NextProtos) == 2 && c.NextProtos[0] == "h2" && c.NextProtos[1] == "http/1.1") {
			return nil, ErrUnavailable
		}
		config = c.Clone()
		config.NextProtos = nil
	}
	template := &http.Transport{
		Proxy:                  base.Proxy,
		DialContext:            base.DialContext,
		TLSClientConfig:        config,
		TLSHandshakeTimeout:    base.TLSHandshakeTimeout,
		DisableKeepAlives:      true,
		DisableCompression:     base.DisableCompression,
		MaxIdleConns:           base.MaxIdleConns,
		MaxIdleConnsPerHost:    base.MaxIdleConnsPerHost,
		MaxConnsPerHost:        base.MaxConnsPerHost,
		IdleConnTimeout:        base.IdleConnTimeout,
		ResponseHeaderTimeout:  base.ResponseHeaderTimeout,
		ExpectContinueTimeout:  base.ExpectContinueTimeout,
		MaxResponseHeaderBytes: base.MaxResponseHeaderBytes,
		WriteBufferSize:        base.WriteBufferSize,
		ReadBufferSize:         base.ReadBufferSize,
		ForceAttemptHTTP2:      false,
		TLSNextProto:           map[string]func(string, *tls.Conn) http.RoundTripper{},
	}
	return &Transport{budget: b, base: template, cleanupError: cleanupError}, nil
}

type connectionGate struct {
	ctx        context.Context
	mu         sync.Mutex
	writeMu    sync.Mutex
	conn       net.Conn
	stopped    bool
	sent       atomic.Bool
	tlsState   *tls.ConnectionState
	closeOnce  sync.Once
	closeErr   error
	firstWrite func()
	inflight   *atomic.Int32
}

func (g *connectionGate) closeConn() error {
	g.closeOnce.Do(func() {
		if g.conn != nil {
			g.closeErr = g.conn.Close()
		}
	})
	return g.closeErr
}
func (g *connectionGate) register(c *tls.Conn) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.stopped {
		c.Close()
		return false
	}
	g.conn = c
	state := c.ConnectionState()
	g.tlsState = &state
	return true
}
func (g *connectionGate) stop() error {
	g.mu.Lock()
	g.stopped = true
	err := g.closeConn()
	g.mu.Unlock()
	// in-flight Write가 종료될 때까지 소유권을 유지함.
	g.writeMu.Lock()
	g.writeMu.Unlock()
	return err
}

type gatedConn struct {
	net.Conn
	gate *connectionGate
}

func (c *gatedConn) Close() error {
	c.gate.mu.Lock()
	defer c.gate.mu.Unlock()
	return c.gate.closeConn()
}
func (c *gatedConn) Write(p []byte) (int, error) {
	c.gate.writeMu.Lock()
	defer c.gate.writeMu.Unlock()
	c.gate.mu.Lock()
	if c.gate.stopped {
		c.gate.mu.Unlock()
		return 0, io.ErrClosedPipe
	}
	if err := c.gate.ctx.Err(); err != nil {
		c.gate.mu.Unlock()
		return 0, err
	}
	first := len(p) > 0 && !c.gate.sent.Swap(true)
	c.gate.mu.Unlock()
	if first && c.gate.firstWrite != nil {
		c.gate.firstWrite()
	}
	if c.gate.inflight != nil {
		c.gate.inflight.Add(1)
		defer c.gate.inflight.Add(-1)
	}
	return c.Conn.Write(p)
}

type responseOwner struct {
	body     io.ReadCloser
	once     sync.Once
	done     chan struct{}
	closeErr error
	finish   func() error
}

func (o *responseOwner) Close() error {
	o.once.Do(func() { o.closeErr = errors.Join(o.body.Close(), o.finish()); close(o.done) })
	return o.closeErr
}
func (o *responseOwner) Read(p []byte) (int, error) {
	n, e := o.body.Read(p)
	if e != nil {
		if closeErr := o.Close(); closeErr != nil {
			e = errors.Join(e, closeErr)
		}
	}
	return n, e
}

func (t *Transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	sent := false
	defer func() {
		if err != nil && !sent {
			err = errors.Join(ErrNotSent, err)
		}
	}()
	p := PriorityFromContext(req.Context())
	if p > Auxiliary || req.URL == nil || req.URL.Scheme != "https" || req.URL.Host == "" {
		if req.Body != nil {
			err = errors.Join(ErrUnavailable, req.Body.Close())
		} else {
			err = ErrUnavailable
		}
		return nil, err
	}
	if t.base.Proxy != nil {
		proxy, proxyErr := t.base.Proxy(req)
		if proxyErr != nil || proxy != nil {
			err = errors.Join(ErrUnavailable, proxyErr)
			if req.Body != nil {
				err = errors.Join(err, req.Body.Close())
			}
			return nil, err
		}
	}
	lease, e := t.budget.Acquire(req.Context(), p)
	if e != nil || lease == nil {
		if req.Body != nil {
			e = errors.Join(e, req.Body.Close())
		}
		if e == nil {
			e = ErrUnavailable
		}
		return nil, e
	}
	gate := &connectionGate{ctx: req.Context(), firstWrite: t.firstWrite, inflight: t.inflight}
	tr := t.base.Clone()
	tr.Proxy = nil
	tr.DisableKeepAlives = true
	tr.ForceAttemptHTTP2 = false
	tr.TLSNextProto = map[string]func(string, *tls.Conn) http.RoundTripper{}
	tr.DialTLSContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, _, e := net.SplitHostPort(addr)
		if e != nil {
			return nil, e
		}
		host = strings.Trim(host, "[]")
		cfg := tr.TLSClientConfig
		if cfg == nil {
			cfg = &tls.Config{}
		} else {
			cfg = cfg.Clone()
		}
		cfg.ServerName = host
		cfg.NextProtos = []string{"http/1.1"}
		dial := t.base.DialContext
		if dial == nil {
			dial = (&net.Dialer{KeepAlive: -1}).DialContext
		}
		raw, e := dial(ctx, network, addr)
		if e != nil {
			return nil, e
		}
		tc := tls.Client(raw, cfg)
		handshakeCtx := ctx
		cancel := func() {}
		if tr.TLSHandshakeTimeout > 0 {
			handshakeCtx, cancel = context.WithTimeout(ctx, tr.TLSHandshakeTimeout)
		}
		defer cancel()
		if e = tc.HandshakeContext(handshakeCtx); e != nil {
			raw.Close()
			return nil, e
		}
		if !gate.register(tc) {
			return nil, io.ErrClosedPipe
		}
		return &gatedConn{Conn: tc, gate: gate}, nil
	}
	resp, e = tr.RoundTrip(req)
	if resp == nil {
		if t.beforeStop != nil {
			t.beforeStop(gate)
		}
		stopErr := gate.stop()
		// revoke/drain 뒤에만 현재 시도의 첫 Write 여부를 확정함.
		sent = gate.sent.Load()
		cleanup := errors.Join(stopErr, lease.Release())
		tr.CloseIdleConnections()
		return nil, errors.Join(e, cleanup)
	}
	sent = gate.sent.Load()
	gate.mu.Lock()
	if resp.TLS == nil && gate.tlsState != nil {
		state := *gate.tlsState
		resp.TLS = &state
	}
	gate.mu.Unlock()
	owner := &responseOwner{body: resp.Body, done: make(chan struct{}), finish: func() error { e := errors.Join(gate.stop(), lease.Release()); tr.CloseIdleConnections(); return e }}
	resp.Body = owner
	go func() {
		select {
		case <-req.Context().Done():
			if e := owner.Close(); e != nil {
				t.cleanupError(e)
			}
		case <-owner.done:
		}
	}()
	return resp, e
}
