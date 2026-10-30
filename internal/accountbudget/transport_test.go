//go:build linux

package accountbudget

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func fixture(t *testing.T) (*Budget, string) {
	t.Helper()
	root := t.TempDir()
	if err := os.Chmod(root, 0700); err != nil {
		t.Fatal(err)
	}
	if err := Provision(root, "scope-a", "instance-a"); err != nil {
		t.Fatal(err)
	}
	b, err := Open(Config{Root: root, Scope: "scope-a", Instance: "instance-a", Clock: clock.System()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := b.Close(); err != nil {
			t.Error(err)
		}
	})
	return b, root
}
func TestTransportOwnsFullBody(t *testing.T) {
	b, _ := fixture(t)
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("payload")) }))
	defer s.Close()
	pool := x509.NewCertPool()
	pool.AddCert(s.Certificate())
	async := make(chan error, 1)
	tr, e := NewTransport(b, &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool}}, func(e error) { async <- e })
	if e != nil {
		t.Fatal(e)
	}
	req, _ := http.NewRequestWithContext(WithPriority(context.Background(), Normal), "GET", s.URL, nil)
	resp, e := tr.RoundTrip(req)
	if e != nil {
		t.Fatal(e)
	}
	if resp.TLS == nil || len(resp.TLS.VerifiedChains) == 0 {
		t.Fatal("missing verified TLS")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()
	_, e = b.Acquire(ctx, Critical)
	if !errors.Is(e, context.DeadlineExceeded) {
		t.Fatalf("wire released before body: %v", e)
	}
	got, e := io.ReadAll(resp.Body)
	if e != nil {
		t.Fatal(e)
	}
	if string(got) != "payload" {
		t.Fatalf("body %q", got)
	}
	if e = resp.Body.Close(); e != nil {
		t.Fatal(e)
	}
	lease, e := b.Acquire(context.Background(), Critical)
	if e != nil {
		t.Fatal(e)
	}
	if e = lease.Release(); e != nil {
		t.Fatal(e)
	}
	select {
	case e = <-async:
		t.Fatal(e)
	default:
	}
}
func TestTransportPreAdmissionIsNotSent(t *testing.T) {
	b, _ := fixture(t)
	tr, e := NewTransport(b, &http.Transport{}, func(error) {})
	if e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithCancel(WithPriority(context.Background(), Critical))
	cancel()
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://localhost/", nil)
	_, e = tr.RoundTrip(req)
	if !errors.Is(e, ErrNotSent) || !errors.Is(e, context.Canceled) {
		t.Fatalf("error %v", e)
	}
}

func TestAttemptClassificationAfterDrain(t *testing.T) {
	b, _ := fixture(t)
	tr, e := NewTransport(b, &http.Transport{}, func(error) {})
	if e != nil {
		t.Fatal(e)
	}
	tr.beforeStop = func(g *connectionGate) {
		writer, reader := net.Pipe()
		g.mu.Lock()
		g.conn = writer
		g.mu.Unlock()
		done := make(chan error, 1)
		go func() { buf := make([]byte, 1); _, e := reader.Read(buf); done <- e }()
		c := &gatedConn{Conn: writer, gate: g}
		if _, e := c.Write([]byte("x")); e != nil {
			t.Fatal(e)
		}
		if e := <-done; e != nil {
			t.Fatal(e)
		}
		reader.Close()
	}
	s := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer s.Close()
	req, _ := http.NewRequest("GET", s.URL, nil)
	_, e = tr.RoundTrip(req)
	if e == nil || errors.Is(e, ErrNotSent) {
		t.Fatalf("first write must not be marked NOT_SENT: %v", e)
	}
}

func TestProxySelectionRefusedBeforeAdmission(t *testing.T) {
	for _, mode := range []string{"proxy", "error"} {
		t.Run(mode, func(t *testing.T) {
			b, _ := fixture(t)
			base := &http.Transport{Proxy: func(*http.Request) (*url.URL, error) {
				if mode == "error" {
					return nil, errors.New("proxy lookup error")
				}
				return url.Parse("http://127.0.0.1:10000")
			}}
			tr, e := NewTransport(b, base, func(error) {})
			if e != nil {
				t.Fatal(e)
			}
			req, _ := http.NewRequest("GET", "https://example.invalid/", nil)
			ctx, cancel := context.WithTimeout(req.Context(), 100*time.Millisecond)
			defer cancel()
			req = req.WithContext(ctx)
			_, e = tr.RoundTrip(req)
			if !errors.Is(e, ErrNotSent) || !errors.Is(e, ErrUnavailable) {
				t.Fatalf("proxy not refused: %v", e)
			}
		})
	}
}

func TestTLSProcessHelper(t *testing.T) {
	root := os.Getenv("BUDGET_TLS_ROOT")
	if root == "" {
		return
	}
	b, e := Open(Config{Root: root, Scope: "scope-a", Instance: "instance-a", Clock: clock.System()})
	if e != nil {
		t.Fatal(e)
	}
	defer b.Close()
	pemBytes, e := os.ReadFile(filepath.Join(root, "test-cert.pem"))
	if e != nil {
		t.Fatal(e)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pemBytes) {
		t.Fatal("certificate")
	}
	tr, e := NewTransport(b, &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool}}, func(e error) { t.Error(e) })
	if e != nil {
		t.Fatal(e)
	}
	for i := 0; i < 2; i++ {
		index := strconv.Itoa(i)
		tr.firstWrite = func() {
			value := []byte(strconv.FormatInt(time.Now().UnixNano(), 10))
			if e := os.WriteFile(filepath.Join(root, "test-start-"+os.Getenv("BUDGET_TLS_ID")+"-"+index), value, 0600); e != nil {
				t.Error(e)
			}
		}
		req, _ := http.NewRequest("GET", os.Getenv("BUDGET_TLS_URL"), nil)
		resp, e := tr.RoundTrip(req)
		if e != nil {
			t.Fatal(e)
		}
		_, e = io.Copy(io.Discard, resp.Body)
		if e != nil {
			t.Fatal(e)
		}
		if e = resp.Body.Close(); e != nil {
			t.Fatal(e)
		}
	}

}
func TestTLSFirstWriteSpacingAcrossProcesses(t *testing.T) {
	_, root := fixture(t)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) }))
	defer server.Close()
	cert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw})
	if e := os.WriteFile(filepath.Join(root, "test-cert.pem"), cert, 0600); e != nil {
		t.Fatal(e)
	}
	children := make([]*exec.Cmd, 2)
	for i := range children {
		c := exec.Command(os.Args[0], "-test.run=^TestTLSProcessHelper$")
		c.Env = append(os.Environ(), "BUDGET_TLS_ROOT="+root, "BUDGET_TLS_URL="+server.URL, "BUDGET_TLS_ID="+strconv.Itoa(i))
		if e := c.Start(); e != nil {
			t.Fatal(e)
		}
		children[i] = c
	}
	for _, c := range children {
		if e := c.Wait(); e != nil {
			t.Fatal(e)
		}
	}
	times := make([]int64, 0, 4)
	for process := 0; process < 2; process++ {
		for request := 0; request < 2; request++ {
			v, e := os.ReadFile(filepath.Join(root, fmt.Sprintf("test-start-%d-%d", process, request)))
			if e != nil {
				t.Fatal(e)
			}
			stamp, e := strconv.ParseInt(string(v), 10, 64)
			if e != nil {
				t.Fatal(e)
			}
			times = append(times, stamp)
		}
	}
	sort.Slice(times, func(i, j int) bool { return times[i] < times[j] })
	for i := 1; i < len(times); i++ {
		if gap := time.Duration(times[i] - times[i-1]); gap < interval {
			t.Fatalf("TLS first-write gap %v", gap)
		}
	}
	for i := 3; i < len(times); i++ {
		if time.Duration(times[i]-times[i-3]) < time.Second {
			t.Fatal("four first writes within one second")
		}
	}

}

type zeros struct{}

func (zeros) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 0
	}
	return len(p), nil
}

type countBody struct {
	io.Reader
	closes atomic.Int32
}

func (b *countBody) Close() error { b.closes.Add(1); return nil }
func TestCancelBlockedTLSWriteDrains(t *testing.T) {
	b, _ := fixture(t)
	entered := make(chan struct{})
	hold := make(chan struct{})
	var calls atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			close(entered)
			select {
			case <-r.Context().Done():
			case <-hold:
			}
			return
		}
		w.Write([]byte("ok"))
	}))
	defer server.Close()
	defer close(hold)
	pool := x509.NewCertPool()
	pool.AddCert(server.Certificate())
	tr, e := NewTransport(b, &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool}}, func(e error) { t.Error(e) })
	if e != nil {
		t.Fatal(e)
	}
	var inflight atomic.Int32
	tr.inflight = &inflight
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	body := &countBody{Reader: io.LimitReader(zeros{}, 128<<20)}
	req, _ := http.NewRequestWithContext(ctx, "POST", server.URL, body)
	req.ContentLength = 128 << 20
	done := make(chan error, 1)
	go func() {
		resp, e := tr.RoundTrip(req)
		if resp != nil {
			resp.Body.Close()
		}
		done <- e
	}()
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("server not reached")
	}
	time.Sleep(200 * time.Millisecond)
	if inflight.Load() != 1 {
		t.Fatal("TLS write not blocked")
	}
	cancel()
	select {
	case e := <-done:
		if !errors.Is(e, context.Canceled) {
			t.Fatalf("cancel error %v", e)
		}
		if errors.Is(e, ErrNotSent) {
			t.Fatal("sent request classified not sent")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("write did not drain")
	}
	if body.closes.Load() != 1 {
		t.Fatalf("request body close count %d", body.closes.Load())
	}
	req2, _ := http.NewRequest("GET", server.URL, nil)
	resp, e := tr.RoundTrip(req2)
	if e != nil {
		t.Fatal(e)
	}
	v, e := io.ReadAll(resp.Body)
	if e != nil || string(v) != "ok" {
		t.Fatalf("next body %q %v", v, e)
	}
	if e = resp.Body.Close(); e != nil {
		t.Fatal(e)
	}
}

func TestCancelUnreadResponseReleasesOwnership(t *testing.T) {
	b, _ := fixture(t)
	hold := make(chan struct{})
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("part"))
		w.(http.Flusher).Flush()
		select {
		case <-r.Context().Done():
		case <-hold:
		}
	}))
	defer server.Close()
	defer close(hold)
	pool := x509.NewCertPool()
	pool.AddCert(server.Certificate())
	async := make(chan error, 1)
	tr, e := NewTransport(b, &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool}}, func(e error) { async <- e })
	if e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithCancel(context.Background())
	req, _ := http.NewRequestWithContext(ctx, "GET", server.URL, nil)
	resp, e := tr.RoundTrip(req)
	if e != nil {
		cancel()
		t.Fatal(e)
	}
	cancel()
	nextCtx, nextCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer nextCancel()
	l, e := b.Acquire(nextCtx, Critical)
	if e != nil {
		t.Fatal(e)
	}
	if e = l.Release(); e != nil {
		t.Fatal(e)
	}
	if e = resp.Body.Close(); e != nil {
		t.Logf("body close after cancel: %v", e)
	}
	select {
	case e := <-async:
		t.Logf("async cleanup: %v", e)
	default:
	}
}
