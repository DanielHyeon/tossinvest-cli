//go:build linux

package accountbudgetruntime

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

func a169FenceClientTransport(t *testing.T, addr string, cert []byte) *http.Transport {
	t.Helper()
	host, _, e := net.SplitHostPort(addr)
	if e != nil || net.ParseIP(host) == nil || !net.ParseIP(host).IsLoopback() {
		t.Fatal("fixture requires loopback")
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(cert) {
		t.Fatal("invalid fixture certificate")
	}
	return &http.Transport{TLSClientConfig: &tls.Config{RootCAs: roots}, DialContext: func(ctx context.Context, network, got string) (net.Conn, error) {
		if network != "tcp" || got != addr {
			return nil, errors.New("fixture destination rejected")
		}
		return (&net.Dialer{}).DialContext(ctx, network, addr)
	}}
}
func TestA169FenceOfficial401TokenAndSticky(t *testing.T) {
	_, path, _ := setupGroup(t)
	f, e := Open(context.Background(), Config{Mode: ModeRequired, Manifest: path, CleanupError: func(error) {}})
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	b, e := f.Budget()
	if e != nil {
		t.Fatal(e)
	}
	fence, e := b.NewObservationFence(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	ctx, e := fence.Context(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	var tokens, gets atomic.Int32
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			tokens.Add(1)
			fmt.Fprint(w, `{"access_token":"fake-minted","expires_in":3600,"token_type":"Bearer"}`)
		case "/api/v1/price-limits":
			n := gets.Add(1)
			if n == 1 {
				w.WriteHeader(401)
			} else if n == 2 {
				fmt.Fprint(w, `{"result":{"lowerLimitPrice":"50400","timestamp":"2026-03-30T09:30:00+09:00"}}`)
			} else {
				w.WriteHeader(429)
			}
		default:
			t.Error("unexpected fixture path")
		}
	}))
	defer srv.Close()
	cert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: srv.Certificate().Raw})
	base := a169FenceClientTransport(t, srv.Listener.Addr().String(), cert)
	c, e := f.NewClient(official.Credentials{APIKey: "fake", SecretKey: "fake"}, "", official.WithBaseURL(srv.URL), official.WithHTTPClient(&http.Client{Transport: base, Timeout: 5 * time.Second}))
	if e != nil {
		t.Fatal(e)
	}
	if _, e = c.ProtectivePriceLimits(ctx, "005930"); e != nil {
		t.Fatalf("401/mint: %v", e)
	}
	if tokens.Load() != 2 || gets.Load() != 2 {
		t.Fatalf("token=%d GET=%d", tokens.Load(), gets.Load())
	}
	if _, e = c.ProtectivePriceLimits(ctx, "005930"); e == nil || gets.Load() != 3 {
		t.Fatalf("429 err=%v GET=%d", e, gets.Load())
	}
	if _, e = c.ProtectivePriceLimits(ctx, "005930"); e == nil || gets.Load() != 3 || tokens.Load() != 2 {
		t.Fatalf("sticky err=%v token=%d GET=%d", e, tokens.Load(), gets.Load())
	}
}
func TestA169FenceCrossProcessChild(t *testing.T) {
	path := os.Getenv("A169_FENCE_CHILD_MANIFEST")
	if path == "" {
		return
	}
	cert, e := os.ReadFile(os.Getenv("A169_FENCE_CHILD_CERT"))
	if e != nil {
		t.Fatal(e)
	}
	url := os.Getenv("A169_FENCE_CHILD_URL")
	addr := strings.TrimPrefix(url, "https://")
	base := a169FenceClientTransport(t, addr, cert)
	f, e := Open(context.Background(), Config{Mode: ModeRequired, Manifest: path, CleanupError: func(error) {}})
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	c, e := f.NewClient(official.Credentials{APIKey: "fake-child", SecretKey: "fake"}, "", official.WithBaseURL(url), official.WithHTTPClient(&http.Client{Transport: base, Timeout: 5 * time.Second}))
	if e != nil {
		t.Fatal(e)
	}
	if _, e = c.AuthHeaders(context.Background()); e == nil {
		t.Fatal("expected token 429")
	}
}
func TestA169FenceCrossProcessToken429(t *testing.T) {
	root, path, _ := setupGroup(t)
	f, e := Open(context.Background(), Config{Mode: ModeRequired, Manifest: path, CleanupError: func(error) {}})
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	b, e := f.Budget()
	if e != nil {
		t.Fatal(e)
	}
	fence, e := b.NewObservationFence(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	ctx, e := fence.Context(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	var tokens atomic.Int32
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokens.Add(1)
		if r.URL.Path != "/oauth2/token" {
			t.Error("unexpected path")
		}
		w.WriteHeader(429)
	}))
	defer srv.Close()
	cert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: srv.Certificate().Raw})
	certPath := filepath.Join(root, "fence-fixture.pem")
	if e = os.WriteFile(certPath, cert, 0600); e != nil {
		t.Fatal(e)
	}
	cmdCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(cmdCtx, os.Args[0], "-test.run=^TestA169FenceCrossProcessChild$", "-test.v")
	cmd.Env = append(os.Environ(), "A169_FENCE_CHILD_MANIFEST="+path, "A169_FENCE_CHILD_CERT="+certPath, "A169_FENCE_CHILD_URL="+srv.URL)
	out, e := cmd.CombinedOutput()
	if e != nil {
		t.Fatalf("child %v: %s", e, out)
	}
	base := a169FenceClientTransport(t, srv.Listener.Addr().String(), cert)
	c, e := f.NewClient(official.Credentials{APIKey: "fake-parent", SecretKey: "fake"}, "", official.WithBaseURL(srv.URL), official.WithHTTPClient(&http.Client{Transport: base, Timeout: 5 * time.Second}))
	if e != nil {
		t.Fatal(e)
	}
	if _, e = c.AuthHeaders(ctx); e == nil || tokens.Load() != 1 {
		t.Fatalf("parent err=%v tokens=%d", e, tokens.Load())
	}
}

func TestA169FenceOfficialCachedAdoptAndMint(t *testing.T) {
	_, path, _ := setupGroup(t)
	f, e := Open(context.Background(), Config{Mode: ModeRequired, Manifest: path, CleanupError: func(error) {}})
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	b, e := f.Budget()
	if e != nil {
		t.Fatal(e)
	}
	fence, e := b.NewObservationFence(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	ctx, e := fence.Context(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	cache := filepath.Join(t.TempDir(), "token.json")
	writeCache := func(value string) error {
		p, e := json.Marshal(map[string]any{"access_token": value, "expires_at": time.Now().Add(time.Hour)})
		if e != nil {
			return e
		}
		return os.WriteFile(cache, p, 0600)
	}
	if e = writeCache("fake-cached"); e != nil {
		t.Fatal(e)
	}
	var tokens, gets atomic.Int32
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			tokens.Add(1)
			fmt.Fprint(w, `{"access_token":"fake-minted","expires_in":3600,"token_type":"Bearer"}`)
		case "/api/v1/price-limits":
			n := gets.Add(1)
			if n == 1 {
				if e := writeCache("fake-adopted"); e != nil {
					t.Error(e)
				}
				w.WriteHeader(401)
			} else if n == 2 {
				w.WriteHeader(401)
			} else {
				fmt.Fprint(w, `{"result":{"lowerLimitPrice":"50400","timestamp":"2026-03-30T09:30:00+09:00"}}`)
			}
		default:
			t.Error("unexpected fixture path")
		}
	}))
	defer srv.Close()
	cert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: srv.Certificate().Raw})
	base := a169FenceClientTransport(t, srv.Listener.Addr().String(), cert)
	c, e := f.NewClient(official.Credentials{APIKey: "fake", SecretKey: "fake"}, cache, official.WithBaseURL(srv.URL), official.WithHTTPClient(&http.Client{Transport: base, Timeout: 5 * time.Second}))
	if e != nil {
		t.Fatal(e)
	}
	if _, e = c.ProtectivePriceLimits(ctx, "005930"); e != nil || tokens.Load() != 1 || gets.Load() != 3 {
		t.Fatalf("err=%v token=%d GET=%d", e, tokens.Load(), gets.Load())
	}
}
