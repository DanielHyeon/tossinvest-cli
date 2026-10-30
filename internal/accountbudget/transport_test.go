//go:build linux

package accountbudget

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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
	tr.firstWrite = func() {
		value := []byte(strconv.FormatInt(time.Now().UnixNano(), 10))
		if e := os.WriteFile(filepath.Join(root, "test-start-"+os.Getenv("BUDGET_TLS_ID")), value, 0600); e != nil {
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
	times := make([]int64, 2)
	for i := range times {
		v, e := os.ReadFile(filepath.Join(root, "test-start-"+strconv.Itoa(i)))
		if e != nil {
			t.Fatal(e)
		}
		times[i], e = strconv.ParseInt(string(v), 10, 64)
		if e != nil {
			t.Fatal(e)
		}
	}
	gap := times[1] - times[0]
	if gap < 0 {
		gap = -gap
	}
	if time.Duration(gap) < interval {
		t.Fatalf("TLS first-write gap %v", time.Duration(gap))
	}
}
