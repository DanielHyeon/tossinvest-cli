package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
	"unsafe"

	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func TestA166MBCachedOff(t *testing.T) {
	t.Setenv("TOSSOS_ACCOUNT_BUDGET_MODE", "off")
	if err := os.Unsetenv("TOSSOS_ACCOUNT_BUDGET_MANIFEST"); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	cache := filepath.Join(dir, "token")
	if err := os.WriteFile(cache, []byte(fmt.Sprintf(`{"access_token":"fixture","expires_at":%q}`, time.Now().Add(time.Hour).Format(time.RFC3339))), 0600); err != nil {
		t.Fatal(err)
	}
	var calls atomic.Int32
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/candles" || r.Header.Get("Authorization") != "Bearer fixture" {
			t.Errorf("unexpected request method/path/auth")
		}
		w.Header().Set("X-Ratelimit-Limit", "10")
		w.Header().Set("X-Ratelimit-Remaining", "9")
		w.Header().Set("X-Ratelimit-Reset", "1")
		_, _ = w.Write([]byte(`{"result":{"nextBefore":null}}`))
	}))
	defer s.Close()
	d, factory, err := openProductionDependencies(context.Background(), config{tokenCache: cache})
	if err != nil {
		t.Fatal(err)
	}
	defer factory.Close()
	c := d.reader.(officialReader).client
	field := reflect.ValueOf(c).Elem().FieldByName("authorityTransport")
	tr := *(**http.Transport)(unsafe.Pointer(field.UnsafeAddr()))
	tr.Proxy = nil
	tr.TLSClientConfig = s.Client().Transport.(*http.Transport).TLSClientConfig.Clone()
	tr.TLSClientConfig.ServerName = "127.0.0.1"
	want := s.Listener.Addr().String()
	tr.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		if address != "openapi.tossinvest.com:443" {
			return nil, errors.New("fixture refused address")
		}
		return (&net.Dialer{}).DialContext(ctx, network, want)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, readErr := d.reader.Candle(ctx, nil)
	t.Logf("fixture read error: %v", readErr)
	if readErr != nil || calls.Load() != 1 {
		t.Fatalf("err=%v wire calls=%d", readErr, calls.Load())
	}
}

func TestA166MBCachedRequiredMissing(t *testing.T) {
	t.Setenv("TOSSOS_ACCOUNT_BUDGET_MODE", "required")
	t.Setenv("TOSSOS_ACCOUNT_BUDGET_MANIFEST", filepath.Join(t.TempDir(), "missing"))
	_, factory, err := openProductionDependencies(context.Background(), config{tokenCache: filepath.Join(t.TempDir(), "token")})
	if !errors.Is(err, accountbudget.ErrUnavailable) || factory != nil {
		t.Fatalf("expected required refusal, factory=%v err=%v", factory, err)
	}
}

func TestA166MBCachedBusyNoWire(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "budget")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	if err := accountbudget.Provision(root, "scope", "instance"); err != nil {
		t.Fatal(err)
	}
	id := func(path string) accountbudget.FileID {
		t.Helper()
		s, e := os.Stat(path)
		if e != nil {
			t.Fatal(e)
		}
		v := s.Sys().(*syscall.Stat_t)
		return accountbudget.FileID{Device: uint64(v.Dev), Inode: v.Ino}
	}
	ident := accountbudget.Identity{Root: id(root), Queue: id(filepath.Join(root, "queue.lock")), Wire: id(filepath.Join(root, "wire.lock"))}
	raw, e := json.Marshal(map[string]any{"version": 1, "root": root, "scope": "scope", "instance": "instance", "identity": ident})
	if e != nil {
		t.Fatal(e)
	}
	manifest := filepath.Join(base, "manifest.json")
	if e = os.WriteFile(manifest, raw, 0600); e != nil {
		t.Fatal(e)
	}
	t.Setenv("TOSSOS_ACCOUNT_BUDGET_MODE", "required")
	t.Setenv("TOSSOS_ACCOUNT_BUDGET_MANIFEST", manifest)
	b, e := accountbudget.OpenContext(context.Background(), accountbudget.Config{Root: root, Scope: "scope", Instance: "instance", Clock: clock.System()})
	if e != nil {
		t.Fatal(e)
	}
	defer b.Close()
	lease, e := b.Acquire(context.Background(), accountbudget.Critical)
	if e != nil {
		t.Fatal(e)
	}
	defer lease.Release()
	cache := filepath.Join(base, "token")
	if err := os.WriteFile(cache, []byte(fmt.Sprintf(`{"access_token":"fixture","expires_at":%q}`, time.Now().Add(time.Hour).Format(time.RFC3339))), 0600); err != nil {
		t.Fatal(err)
	}
	var calls atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { calls.Add(1) }))
	defer server.Close()
	d, factory, err := openProductionDependencies(context.Background(), config{tokenCache: cache})
	if err != nil {
		t.Fatal(err)
	}
	defer factory.Close()
	c := d.reader.(officialReader).client
	field := reflect.ValueOf(c).Elem().FieldByName("authorityTransport")
	tr := *(**http.Transport)(unsafe.Pointer(field.UnsafeAddr()))
	tr.Proxy = nil
	tr.TLSClientConfig = server.Client().Transport.(*http.Transport).TLSClientConfig.Clone()
	tr.TLSClientConfig.ServerName = ""
	want := server.Listener.Addr().String()
	tr.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		if address != "openapi.tossinvest.com:443" {
			return nil, errors.New("fixture refused address")
		}
		return (&net.Dialer{}).DialContext(ctx, network, want)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	started := time.Now()
	_, err = d.reader.Candle(ctx, nil)
	t.Logf("busy error: %v", err)
	if time.Since(started) > time.Second {
		t.Fatal("auxiliary did not refuse promptly")
	}
	if err == nil || !strings.Contains(err.Error(), "one-shot GET") || calls.Load() != 0 {
		t.Fatalf("busy err=%v calls=%d", err, calls.Load())
	}
}

func TestA166MBCachedRequiredOneGET(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "budget")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	if err := accountbudget.Provision(root, "scope", "instance"); err != nil {
		t.Fatal(err)
	}
	fileID := func(path string) accountbudget.FileID {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		st := info.Sys().(*syscall.Stat_t)
		return accountbudget.FileID{Device: uint64(st.Dev), Inode: st.Ino}
	}
	identity := accountbudget.Identity{Root: fileID(root), Queue: fileID(filepath.Join(root, "queue.lock")), Wire: fileID(filepath.Join(root, "wire.lock"))}
	raw, err := json.Marshal(map[string]any{"version": 1, "root": root, "scope": "scope", "instance": "instance", "identity": identity})
	if err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(base, "manifest")
	if err = os.WriteFile(manifest, raw, 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TOSSOS_ACCOUNT_BUDGET_MODE", "required")
	t.Setenv("TOSSOS_ACCOUNT_BUDGET_MANIFEST", manifest)
	cache := filepath.Join(base, "token")
	if err = os.WriteFile(cache, []byte(fmt.Sprintf(`{"access_token":"fixture","expires_at":%q}`, time.Now().Add(time.Hour).Format(time.RFC3339))), 0600); err != nil {
		t.Fatal(err)
	}
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{SerialNumber: big.NewInt(1), DNSNames: []string{"openapi.tossinvest.com"}, NotBefore: time.Now().Add(-time.Minute), NotAfter: time.Now().Add(time.Hour), KeyUsage: x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, IsCA: true, BasicConstraintsValid: true}
	der, err := x509.CreateCertificate(rand.Reader, template, template, pub, priv)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	var calls atomic.Int32
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.Method != "GET" || r.URL.Path != "/api/v1/candles" {
			t.Error("unexpected request")
		}
		w.Header().Set("X-Ratelimit-Limit", "10")
		w.Header().Set("X-Ratelimit-Remaining", "9")
		w.Header().Set("X-Ratelimit-Reset", "1")
		_, _ = w.Write([]byte(`{"result":{"nextBefore":null}}`))
	}))
	server.TLS = &tls.Config{Certificates: []tls.Certificate{{Certificate: [][]byte{der}, PrivateKey: priv}}}
	server.StartTLS()
	defer server.Close()
	d, f, err := openProductionDependencies(context.Background(), config{tokenCache: cache})
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	c := d.reader.(officialReader).client
	field := reflect.ValueOf(c).Elem().FieldByName("authorityTransport")
	tr := *(**http.Transport)(unsafe.Pointer(field.UnsafeAddr()))
	pool := x509.NewCertPool()
	pool.AddCert(cert)
	tr.Proxy = nil
	tr.TLSClientConfig = &tls.Config{RootCAs: pool}
	want := server.Listener.Addr().String()
	tr.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		if address != "openapi.tossinvest.com:443" {
			return nil, errors.New("fixture refused address")
		}
		return (&net.Dialer{}).DialContext(ctx, network, want)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = d.reader.Candle(ctx, nil)
	if err != nil || calls.Load() != 1 {
		t.Fatalf("err=%v calls=%d", err, calls.Load())
	}
}
