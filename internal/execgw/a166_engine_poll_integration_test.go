//go:build linux

package execgw

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
	"unsafe"

	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget"
	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudgetruntime"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"golang.org/x/sys/unix"
)

func TestA166EnginePollActual429StopsAndSharesHint(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "group")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	if err := accountbudget.Provision(root, "scope", "instance"); err != nil {
		t.Fatal(err)
	}
	b, err := accountbudget.Open(accountbudget.Config{Root: root, Scope: "scope", Instance: "instance", Clock: clock.System()})
	if err != nil {
		t.Fatal(err)
	}
	event, err := accountbudget.ProvisionObservation(b)
	err = errors.Join(err, b.Close())
	if err != nil {
		t.Fatal(err)
	}
	id := func(path string) accountbudget.FileID {
		st, e := os.Stat(path)
		if e != nil {
			t.Fatal(e)
		}
		v := st.Sys().(*syscall.Stat_t)
		return accountbudget.FileID{Device: uint64(v.Dev), Inode: v.Ino}
	}
	data, err := json.Marshal(map[string]any{"version": 2, "root": root, "scope": "scope", "instance": "instance", "event_lock": event, "identity": accountbudget.Identity{Root: id(root), Queue: id(filepath.Join(root, "queue.lock")), Wire: id(filepath.Join(root, "wire.lock"))}})
	if err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(dir, "manifest")
	if err = os.WriteFile(manifest, data, 0600); err != nil {
		t.Fatal(err)
	}
	f, err := accountbudgetruntime.Open(context.Background(), accountbudgetruntime.Config{Mode: accountbudgetruntime.ModeRequired, Manifest: manifest, CleanupError: func(error) {}})
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	b, err = f.Budget()
	if err != nil {
		t.Fatal(err)
	}
	var tokenCalls, priceCalls atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth2/token" {
			tokenCalls.Add(1)
			fmt.Fprint(w, `{"access_token":"fake","expires_in":3600,"token_type":"Bearer"}`)
			return
		}
		priceCalls.Add(1)
		w.Header().Set("Retry-After", "2")
		w.WriteHeader(429)
	}))
	defer server.Close()
	pool := x509.NewCertPool()
	pool.AddCert(server.Certificate())
	want := server.Listener.Addr().String()
	tr := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool}, DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
		if network != "tcp" || address != want {
			return nil, errors.New("unexpected destination")
		}
		return (&net.Dialer{}).DialContext(ctx, network, want)
	}}
	client, err := f.NewClient(official.Credentials{APIKey: "fake", SecretKey: "fake"}, "", official.WithBaseURL(server.URL), official.WithHTTPClient(&http.Client{Transport: tr}), official.WithAccountSeq(1))
	if err != nil {
		t.Fatal(err)
	}
	retrier, err := NewEnginePollRetrier(&Retrier{AccountRef: "fixture"}, b, client, "fixture", 1)
	if err != nil {
		t.Fatal(err)
	}
	syntheticCalls := 0
	err = retrier.Query(context.Background(), QueryOpenOrders, func(context.Context) error { syntheticCalls++; return &official.APIError{Code: 429} })
	if !errors.Is(err, official.ErrRateLimited) || syntheticCalls != 1 {
		t.Fatalf("api429 err=%v calls=%d", err, syntheticCalls)
	}
	if retrier.EnginePollReaderAllowed([]int{1}) || retrier.EnginePollReaderAllowed(official.New(official.Credentials{}, "")) || !retrier.EnginePollReaderAllowed(client) {
		t.Fatal("reader binding")
	}
	seqField := reflect.ValueOf(client).Elem().FieldByName("accountSeq")
	seqPtr := (*atomic.Int64)(unsafe.Pointer(seqField.UnsafeAddr()))
	originalSeq := seqPtr.Load()
	seqPtr.Store(originalSeq + 1)
	driftCalled := false
	driftErr := retrier.Query(context.Background(), QueryPrice, func(context.Context) error { driftCalled = true; return nil })
	seqPtr.Store(originalSeq)
	if !errors.Is(driftErr, accountbudget.ErrUnavailable) || driftCalled {
		t.Fatalf("sequence drift err=%v fn=%v", driftErr, driftCalled)
	}
	calledBad := false
	if err := retrier.Query(context.Background(), RequiredQuery("unknown"), func(context.Context) error { calledBad = true; return nil }); !errors.Is(err, accountbudget.ErrUnavailable) || calledBad {
		t.Fatalf("unknown err=%v called=%v", err, calledBad)
	}
	copied := *retrier
	copied.AccountRef = "foreign"
	if err := copied.Query(context.Background(), QueryPrice, func(context.Context) error { calledBad = true; return nil }); !errors.Is(err, accountbudget.ErrUnavailable) || calledBad {
		t.Fatalf("mismatch err=%v called=%v", err, calledBad)
	}
	read := func(ctx context.Context) error { _, err := client.Prices(ctx, []string{"005930"}); return err }
	err = retrier.Query(context.Background(), QueryPrice, read)
	if !errors.Is(err, official.ErrRateLimited) || tokenCalls.Load() != 1 || priceCalls.Load() != 1 {
		t.Fatalf("err=%v token=%d price=%d", err, tokenCalls.Load(), priceCalls.Load())
	}
	called := false
	err = retrier.Query(context.Background(), QueryPrice, func(context.Context) error { called = true; return nil })
	if !errors.Is(err, accountbudget.ErrObservationDeferred) || called {
		t.Fatalf("hint err=%v called=%v", err, called)
	}

	gate := NewEntryGate(clock.System(), map[RequiredQuery]time.Duration{})
	retrier.Gate = gate
	escalator := &pollEscalator{}
	retrier.Escalate = escalator
	err = retrier.Query(context.Background(), QueryHoldings, func(ctx context.Context) error { _, _ = client.Prices(ctx, []string{"005930"}); return nil })
	if !errors.Is(err, official.ErrRateLimited) || priceCalls.Load() != 2 {
		t.Fatalf("hidden rate limit err=%v calls=%d", err, priceCalls.Load())
	}
	err = retrier.Query(context.Background(), QueryBuyingPower, func(ctx context.Context) error {
		_, _ = client.Prices(ctx, []string{"005930"})
		return official.ErrAuth
	})
	if !errors.Is(err, official.ErrAuth) || !errors.Is(err, official.ErrRateLimited) || gate.CheckEntry() == nil || gate.CheckEntry().Reason != ReasonBrokerAuthRejected || escalator.calls != 1 {
		t.Fatalf("auth err=%v gate=%v", err, gate.CheckEntry())
	}
}

func TestA166EnginePollTokenFinishFaultStopsGET(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "group")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	if err := accountbudget.Provision(root, "scope", "instance"); err != nil {
		t.Fatal(err)
	}
	b, err := accountbudget.Open(accountbudget.Config{Root: root, Scope: "scope", Instance: "instance", Clock: clock.System()})
	if err != nil {
		t.Fatal(err)
	}
	event, err := accountbudget.ProvisionObservation(b)
	err = errors.Join(err, b.Close())
	if err != nil {
		t.Fatal(err)
	}
	id := func(path string) accountbudget.FileID {
		st, e := os.Stat(path)
		if e != nil {
			t.Fatal(e)
		}
		v := st.Sys().(*syscall.Stat_t)
		return accountbudget.FileID{Device: uint64(v.Dev), Inode: v.Ino}
	}
	data, err := json.Marshal(map[string]any{"version": 2, "root": root, "scope": "scope", "instance": "instance", "event_lock": event, "identity": accountbudget.Identity{Root: id(root), Queue: id(filepath.Join(root, "queue.lock")), Wire: id(filepath.Join(root, "wire.lock"))}})
	if err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(dir, "manifest")
	if err = os.WriteFile(manifest, data, 0600); err != nil {
		t.Fatal(err)
	}
	f, err := accountbudgetruntime.Open(context.Background(), accountbudgetruntime.Config{Mode: accountbudgetruntime.ModeRequired, Manifest: manifest, CleanupError: func(error) {}})
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	b, err = f.Budget()
	if err != nil {
		t.Fatal(err)
	}
	var tokenCalls, priceCalls atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth2/token" {
			tokenCalls.Add(1)
			fmt.Fprint(w, `{"access_token":"fake","expires_in":3600,"token_type":"Bearer"}`)
			return
		}
		priceCalls.Add(1)
		w.Header().Set("Retry-After", "2")
		w.WriteHeader(429)
	}))
	defer server.Close()
	pool := x509.NewCertPool()
	pool.AddCert(server.Certificate())
	want := server.Listener.Addr().String()
	tr := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool}, DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
		if network != "tcp" || address != want {
			return nil, errors.New("unexpected destination")
		}
		return (&net.Dialer{}).DialContext(ctx, network, want)
	}}
	client, err := f.NewClient(official.Credentials{APIKey: "fake", SecretKey: "fake"}, "", official.WithBaseURL(server.URL), official.WithHTTPClient(&http.Client{Transport: tr}), official.WithAccountSeq(1))
	if err != nil {
		t.Fatal(err)
	}
	retrier, err := NewEnginePollRetrier(&Retrier{AccountRef: "fixture"}, b, client, "fixture", 1)
	if err != nil {
		t.Fatal(err)
	}
	fail := errors.New("postcommit finish fixture")
	budgetValue := reflect.ValueOf(b).Elem().FieldByName("observation")
	store := *(**accountbudget.ObservationStore)(unsafe.Pointer(budgetValue.UnsafeAddr()))
	syncField := reflect.ValueOf(store).Elem().FieldByName("syncDir")
	syncFn := (*func(int) error)(unsafe.Pointer(syncField.UnsafeAddr()))
	var syncCalls atomic.Int32
	*syncFn = func(fd int) error {
		if syncCalls.Add(1) == 6 {
			return fail
		}
		return unix.Fsync(fd)
	}
	err = retrier.Query(context.Background(), QueryPrice, func(ctx context.Context) error { _, e := client.Prices(ctx, []string{"005930"}); return e })
	if !errors.Is(err, fail) || tokenCalls.Load() != 1 || priceCalls.Load() != 0 {
		t.Fatalf("err=%v token=%d price=%d sync=%d", err, tokenCalls.Load(), priceCalls.Load(), syncCalls.Load())
	}
}

func TestA166EnginePoll401FinishFaultStopsRefresh(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "group")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	if err := accountbudget.Provision(root, "scope", "instance"); err != nil {
		t.Fatal(err)
	}
	b, err := accountbudget.Open(accountbudget.Config{Root: root, Scope: "scope", Instance: "instance", Clock: clock.System()})
	if err != nil {
		t.Fatal(err)
	}
	event, err := accountbudget.ProvisionObservation(b)
	err = errors.Join(err, b.Close())
	if err != nil {
		t.Fatal(err)
	}
	id := func(path string) accountbudget.FileID {
		st, e := os.Stat(path)
		if e != nil {
			t.Fatal(e)
		}
		v := st.Sys().(*syscall.Stat_t)
		return accountbudget.FileID{Device: uint64(v.Dev), Inode: v.Ino}
	}
	data, err := json.Marshal(map[string]any{"version": 2, "root": root, "scope": "scope", "instance": "instance", "event_lock": event, "identity": accountbudget.Identity{Root: id(root), Queue: id(filepath.Join(root, "queue.lock")), Wire: id(filepath.Join(root, "wire.lock"))}})
	if err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(dir, "manifest")
	if err = os.WriteFile(manifest, data, 0600); err != nil {
		t.Fatal(err)
	}
	f, err := accountbudgetruntime.Open(context.Background(), accountbudgetruntime.Config{Mode: accountbudgetruntime.ModeRequired, Manifest: manifest, CleanupError: func(error) {}})
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	b, err = f.Budget()
	if err != nil {
		t.Fatal(err)
	}
	var tokenCalls, priceCalls atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth2/token" {
			tokenCalls.Add(1)
			fmt.Fprint(w, `{"access_token":"fake","expires_in":3600,"token_type":"Bearer"}`)
			return
		}
		priceCalls.Add(1)
		w.Header().Set("Retry-After", "2")
		w.WriteHeader(401)
	}))
	defer server.Close()
	pool := x509.NewCertPool()
	pool.AddCert(server.Certificate())
	want := server.Listener.Addr().String()
	tr := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool}, DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
		if network != "tcp" || address != want {
			return nil, errors.New("unexpected destination")
		}
		return (&net.Dialer{}).DialContext(ctx, network, want)
	}}
	cache := filepath.Join(dir, "cache")
	if err = os.WriteFile(cache, []byte(fmt.Sprintf(`{"access_token":"cached","expires_at":%q}`, time.Now().Add(time.Hour).Format(time.RFC3339))), 0600); err != nil {
		t.Fatal(err)
	}
	client, err := f.NewClient(official.Credentials{APIKey: "fake", SecretKey: "fake"}, cache, official.WithBaseURL(server.URL), official.WithHTTPClient(&http.Client{Transport: tr}), official.WithAccountSeq(1))
	if err != nil {
		t.Fatal(err)
	}
	retrier, err := NewEnginePollRetrier(&Retrier{AccountRef: "fixture"}, b, client, "fixture", 1)
	if err != nil {
		t.Fatal(err)
	}
	fail := errors.New("postcommit finish fixture")
	budgetValue := reflect.ValueOf(b).Elem().FieldByName("observation")
	store := *(**accountbudget.ObservationStore)(unsafe.Pointer(budgetValue.UnsafeAddr()))
	syncField := reflect.ValueOf(store).Elem().FieldByName("syncDir")
	syncFn := (*func(int) error)(unsafe.Pointer(syncField.UnsafeAddr()))
	var syncCalls atomic.Int32
	*syncFn = func(fd int) error {
		if syncCalls.Add(1) == 6 {
			return fail
		}
		return unix.Fsync(fd)
	}
	err = retrier.Query(context.Background(), QueryPrice, func(ctx context.Context) error { _, e := client.Prices(ctx, []string{"005930"}); return e })
	if !errors.Is(err, fail) || tokenCalls.Load() != 0 || priceCalls.Load() != 1 {
		t.Fatalf("err=%v token=%d price=%d sync=%d", err, tokenCalls.Load(), priceCalls.Load(), syncCalls.Load())
	}
}

type pollEscalator struct{ calls int }

func (e *pollEscalator) EscalateOperatingMode(context.Context, string, string, journal.ModeAnnouncer) (journal.OperatingModeRecord, bool, error) {
	e.calls++
	return journal.OperatingModeRecord{}, true, nil
}

func TestA166EnginePollToken429DoesNotCreatePriceHint(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "group")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	if err := accountbudget.Provision(root, "scope", "instance"); err != nil {
		t.Fatal(err)
	}
	b, err := accountbudget.Open(accountbudget.Config{Root: root, Scope: "scope", Instance: "instance", Clock: clock.System()})
	if err != nil {
		t.Fatal(err)
	}
	event, err := accountbudget.ProvisionObservation(b)
	err = errors.Join(err, b.Close())
	if err != nil {
		t.Fatal(err)
	}
	id := func(path string) accountbudget.FileID {
		st, e := os.Stat(path)
		if e != nil {
			t.Fatal(e)
		}
		v := st.Sys().(*syscall.Stat_t)
		return accountbudget.FileID{Device: uint64(v.Dev), Inode: v.Ino}
	}
	data, err := json.Marshal(map[string]any{"version": 2, "root": root, "scope": "scope", "instance": "instance", "event_lock": event, "identity": accountbudget.Identity{Root: id(root), Queue: id(filepath.Join(root, "queue.lock")), Wire: id(filepath.Join(root, "wire.lock"))}})
	if err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(dir, "manifest")
	if err = os.WriteFile(manifest, data, 0600); err != nil {
		t.Fatal(err)
	}
	f, err := accountbudgetruntime.Open(context.Background(), accountbudgetruntime.Config{Mode: accountbudgetruntime.ModeRequired, Manifest: manifest, CleanupError: func(error) {}})
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	b, err = f.Budget()
	if err != nil {
		t.Fatal(err)
	}
	var tokenCalls, priceCalls atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth2/token" {
			tokenCalls.Add(1)
			w.Header().Set("Retry-After", "2")
			w.WriteHeader(429)
			return
		}
		priceCalls.Add(1)
		w.Header().Set("Retry-After", "2")
		w.WriteHeader(429)
	}))
	defer server.Close()
	pool := x509.NewCertPool()
	pool.AddCert(server.Certificate())
	want := server.Listener.Addr().String()
	tr := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool}, DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
		if network != "tcp" || address != want {
			return nil, errors.New("unexpected destination")
		}
		return (&net.Dialer{}).DialContext(ctx, network, want)
	}}
	client, err := f.NewClient(official.Credentials{APIKey: "fake", SecretKey: "fake"}, "", official.WithBaseURL(server.URL), official.WithHTTPClient(&http.Client{Transport: tr}), official.WithAccountSeq(1))
	if err != nil {
		t.Fatal(err)
	}
	retrier, err := NewEnginePollRetrier(&Retrier{AccountRef: "fixture"}, b, client, "fixture", 1)
	if err != nil {
		t.Fatal(err)
	}
	err = retrier.Query(context.Background(), QueryPrice, func(ctx context.Context) error { _, e := client.Prices(ctx, []string{"005930"}); return e })
	if !errors.Is(err, official.ErrRateLimited) || tokenCalls.Load() != 1 || priceCalls.Load() != 0 {
		t.Fatalf("err=%v token=%d price=%d", err, tokenCalls.Load(), priceCalls.Load())
	}
	key, err := b.EnginePollKey("fixture", string(QueryPrice))
	if err != nil {
		t.Fatal(err)
	}
	remaining, err := b.PollRemaining(context.Background(), key)
	if err != nil || remaining != 0 {
		t.Fatalf("token rate created price hint: %v %v", remaining, err)
	}
}

func TestA166EnginePollConcurrentCollectorIsolation(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "group")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	if err := accountbudget.Provision(root, "scope", "instance"); err != nil {
		t.Fatal(err)
	}
	b, err := accountbudget.Open(accountbudget.Config{Root: root, Scope: "scope", Instance: "instance", Clock: clock.System()})
	if err != nil {
		t.Fatal(err)
	}
	event, err := accountbudget.ProvisionObservation(b)
	err = errors.Join(err, b.Close())
	if err != nil {
		t.Fatal(err)
	}
	id := func(path string) accountbudget.FileID {
		st, e := os.Stat(path)
		if e != nil {
			t.Fatal(e)
		}
		v := st.Sys().(*syscall.Stat_t)
		return accountbudget.FileID{Device: uint64(v.Dev), Inode: v.Ino}
	}
	data, err := json.Marshal(map[string]any{"version": 2, "root": root, "scope": "scope", "instance": "instance", "event_lock": event, "identity": accountbudget.Identity{Root: id(root), Queue: id(filepath.Join(root, "queue.lock")), Wire: id(filepath.Join(root, "wire.lock"))}})
	if err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(dir, "manifest")
	if err = os.WriteFile(manifest, data, 0600); err != nil {
		t.Fatal(err)
	}
	f, err := accountbudgetruntime.Open(context.Background(), accountbudgetruntime.Config{Mode: accountbudgetruntime.ModeRequired, Manifest: manifest, CleanupError: func(error) {}})
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	b, err = f.Budget()
	if err != nil {
		t.Fatal(err)
	}
	var tokenCalls, priceCalls atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth2/token" {
			tokenCalls.Add(1)
			fmt.Fprint(w, `{"access_token":"fake","expires_in":3600,"token_type":"Bearer"}`)
			return
		}
		priceCalls.Add(1)
		w.Header().Set("Retry-After", "2")
		w.WriteHeader(429)
	}))
	defer server.Close()
	pool := x509.NewCertPool()
	pool.AddCert(server.Certificate())
	want := server.Listener.Addr().String()
	tr := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool}, DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
		if network != "tcp" || address != want {
			return nil, errors.New("unexpected destination")
		}
		return (&net.Dialer{}).DialContext(ctx, network, want)
	}}
	client, err := f.NewClient(official.Credentials{APIKey: "fake", SecretKey: "fake"}, "", official.WithBaseURL(server.URL), official.WithHTTPClient(&http.Client{Transport: tr}), official.WithAccountSeq(1))
	if err != nil {
		t.Fatal(err)
	}
	retrier, err := NewEnginePollRetrier(&Retrier{AccountRef: "fixture"}, b, client, "fixture", 1)
	if err != nil {
		t.Fatal(err)
	}
	started := make(chan struct{})
	finish := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- retrier.Query(context.Background(), QueryOpenOrders, func(context.Context) error { close(started); <-finish; return nil })
	}()
	<-started
	err = retrier.Query(context.Background(), QueryHoldings, func(ctx context.Context) error { _, e := client.Prices(ctx, []string{"005930"}); return e })
	close(finish)
	other := <-done
	if !errors.Is(err, official.ErrRateLimited) || other != nil || tokenCalls.Load() != 1 || priceCalls.Load() != 1 {
		t.Fatalf("rate=%v unrelated=%v token=%d price=%d", err, other, tokenCalls.Load(), priceCalls.Load())
	}
}

func TestA166EnginePollInvalidHintNoRetryOrSlot(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "group")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	if err := accountbudget.Provision(root, "scope", "instance"); err != nil {
		t.Fatal(err)
	}
	b, err := accountbudget.Open(accountbudget.Config{Root: root, Scope: "scope", Instance: "instance", Clock: clock.System()})
	if err != nil {
		t.Fatal(err)
	}
	event, err := accountbudget.ProvisionObservation(b)
	err = errors.Join(err, b.Close())
	if err != nil {
		t.Fatal(err)
	}
	id := func(path string) accountbudget.FileID {
		st, e := os.Stat(path)
		if e != nil {
			t.Fatal(e)
		}
		v := st.Sys().(*syscall.Stat_t)
		return accountbudget.FileID{Device: uint64(v.Dev), Inode: v.Ino}
	}
	data, err := json.Marshal(map[string]any{"version": 2, "root": root, "scope": "scope", "instance": "instance", "event_lock": event, "identity": accountbudget.Identity{Root: id(root), Queue: id(filepath.Join(root, "queue.lock")), Wire: id(filepath.Join(root, "wire.lock"))}})
	if err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(dir, "manifest")
	if err = os.WriteFile(manifest, data, 0600); err != nil {
		t.Fatal(err)
	}
	f, err := accountbudgetruntime.Open(context.Background(), accountbudgetruntime.Config{Mode: accountbudgetruntime.ModeRequired, Manifest: manifest, CleanupError: func(error) {}})
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	b, err = f.Budget()
	if err != nil {
		t.Fatal(err)
	}
	var tokenCalls, priceCalls atomic.Int32
	hintValue := ""
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth2/token" {
			tokenCalls.Add(1)
			fmt.Fprint(w, `{"access_token":"fake","expires_in":3600,"token_type":"Bearer"}`)
			return
		}
		priceCalls.Add(1)
		if hintValue != "" {
			w.Header().Set("Retry-After", hintValue)
		}
		w.WriteHeader(429)
	}))
	defer server.Close()
	pool := x509.NewCertPool()
	pool.AddCert(server.Certificate())
	want := server.Listener.Addr().String()
	tr := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool}, DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
		if network != "tcp" || address != want {
			return nil, errors.New("unexpected destination")
		}
		return (&net.Dialer{}).DialContext(ctx, network, want)
	}}
	client, err := f.NewClient(official.Credentials{APIKey: "fake", SecretKey: "fake"}, "", official.WithBaseURL(server.URL), official.WithHTTPClient(&http.Client{Transport: tr}), official.WithAccountSeq(1))
	if err != nil {
		t.Fatal(err)
	}
	retrier, err := NewEnginePollRetrier(&Retrier{AccountRef: "fixture"}, b, client, "fixture", 1)
	if err != nil {
		t.Fatal(err)
	}
	for i, value := range []string{"", "bogus", "31", "999999999999999999999999999"} {
		hintValue = value
		calls := 0
		err = retrier.Query(context.Background(), QueryPrice, func(ctx context.Context) error { calls++; _, e := client.Prices(ctx, []string{"005930"}); return e })
		if !errors.Is(err, official.ErrRateLimited) || calls != 1 || priceCalls.Load() != int32(i+1) {
			t.Fatalf("value=%q err=%v fn=%d wire=%d", value, err, calls, priceCalls.Load())
		}
		key, e := b.EnginePollKey("fixture", string(QueryPrice))
		if e != nil {
			t.Fatal(e)
		}
		remaining, e := b.PollRemaining(context.Background(), key)
		if e != nil || remaining != 0 {
			t.Fatalf("hint %q remaining=%v err=%v", value, remaining, e)
		}
	}
	retrier.Policy = RetryPolicy{MaxAttempts: 2, Budget: 2 * time.Second, BaseBackoff: time.Millisecond, MaxBackoff: time.Millisecond}
	transient := 0
	err = retrier.Query(context.Background(), QueryOpenOrders, func(context.Context) error {
		transient++
		if transient == 1 {
			return official.ErrServer
		}
		return nil
	})
	if err != nil || transient != 2 {
		t.Fatalf("transient err=%v calls=%d", err, transient)
	}

}

func TestA166EnginePollHTTPDateHint(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "group")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	if err := accountbudget.Provision(root, "scope", "instance"); err != nil {
		t.Fatal(err)
	}
	b, err := accountbudget.Open(accountbudget.Config{Root: root, Scope: "scope", Instance: "instance", Clock: clock.System()})
	if err != nil {
		t.Fatal(err)
	}
	event, err := accountbudget.ProvisionObservation(b)
	err = errors.Join(err, b.Close())
	if err != nil {
		t.Fatal(err)
	}
	id := func(path string) accountbudget.FileID {
		st, e := os.Stat(path)
		if e != nil {
			t.Fatal(e)
		}
		v := st.Sys().(*syscall.Stat_t)
		return accountbudget.FileID{Device: uint64(v.Dev), Inode: v.Ino}
	}
	data, err := json.Marshal(map[string]any{"version": 2, "root": root, "scope": "scope", "instance": "instance", "event_lock": event, "identity": accountbudget.Identity{Root: id(root), Queue: id(filepath.Join(root, "queue.lock")), Wire: id(filepath.Join(root, "wire.lock"))}})
	if err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(dir, "manifest")
	if err = os.WriteFile(manifest, data, 0600); err != nil {
		t.Fatal(err)
	}
	f, err := accountbudgetruntime.Open(context.Background(), accountbudgetruntime.Config{Mode: accountbudgetruntime.ModeRequired, Manifest: manifest, CleanupError: func(error) {}})
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	b, err = f.Budget()
	if err != nil {
		t.Fatal(err)
	}
	var tokenCalls, priceCalls atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth2/token" {
			tokenCalls.Add(1)
			fmt.Fprint(w, `{"access_token":"fake","expires_in":3600,"token_type":"Bearer"}`)
			return
		}
		priceCalls.Add(1)
		w.Header().Set("Retry-After", time.Now().Add(3*time.Second).UTC().Format(http.TimeFormat))
		w.WriteHeader(429)
	}))
	defer server.Close()
	pool := x509.NewCertPool()
	pool.AddCert(server.Certificate())
	want := server.Listener.Addr().String()
	tr := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool}, DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
		if network != "tcp" || address != want {
			return nil, errors.New("unexpected destination")
		}
		return (&net.Dialer{}).DialContext(ctx, network, want)
	}}
	client, err := f.NewClient(official.Credentials{APIKey: "fake", SecretKey: "fake"}, "", official.WithBaseURL(server.URL), official.WithHTTPClient(&http.Client{Transport: tr}), official.WithAccountSeq(1))
	if err != nil {
		t.Fatal(err)
	}
	retrier, err := NewEnginePollRetrier(&Retrier{AccountRef: "fixture"}, b, client, "fixture", 1)
	if err != nil {
		t.Fatal(err)
	}
	err = retrier.Query(context.Background(), QueryPrice, func(ctx context.Context) error { _, e := client.Prices(ctx, []string{"005930"}); return e })
	if !errors.Is(err, official.ErrRateLimited) || tokenCalls.Load() != 1 || priceCalls.Load() != 1 {
		t.Fatalf("err=%v token=%d price=%d", err, tokenCalls.Load(), priceCalls.Load())
	}
	key, e := b.EnginePollKey("fixture", string(QueryPrice))
	if e != nil {
		t.Fatal(e)
	}
	remaining, e := b.PollRemaining(context.Background(), key)
	if e != nil || remaining <= 0 || remaining > 3*time.Second {
		t.Fatalf("date hint remaining=%v err=%v", remaining, e)
	}
}

func TestA166EnginePoll429BodyErrorStillRateLimited(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "group")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	if err := accountbudget.Provision(root, "scope", "instance"); err != nil {
		t.Fatal(err)
	}
	b, err := accountbudget.Open(accountbudget.Config{Root: root, Scope: "scope", Instance: "instance", Clock: clock.System()})
	if err != nil {
		t.Fatal(err)
	}
	event, err := accountbudget.ProvisionObservation(b)
	err = errors.Join(err, b.Close())
	if err != nil {
		t.Fatal(err)
	}
	id := func(path string) accountbudget.FileID {
		st, e := os.Stat(path)
		if e != nil {
			t.Fatal(e)
		}
		v := st.Sys().(*syscall.Stat_t)
		return accountbudget.FileID{Device: uint64(v.Dev), Inode: v.Ino}
	}
	data, err := json.Marshal(map[string]any{"version": 2, "root": root, "scope": "scope", "instance": "instance", "event_lock": event, "identity": accountbudget.Identity{Root: id(root), Queue: id(filepath.Join(root, "queue.lock")), Wire: id(filepath.Join(root, "wire.lock"))}})
	if err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(dir, "manifest")
	if err = os.WriteFile(manifest, data, 0600); err != nil {
		t.Fatal(err)
	}
	f, err := accountbudgetruntime.Open(context.Background(), accountbudgetruntime.Config{Mode: accountbudgetruntime.ModeRequired, Manifest: manifest, CleanupError: func(error) {}})
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	b, err = f.Budget()
	if err != nil {
		t.Fatal(err)
	}
	var tokenCalls, priceCalls atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth2/token" {
			tokenCalls.Add(1)
			fmt.Fprint(w, `{"access_token":"fake","expires_in":3600,"token_type":"Bearer"}`)
			return
		}
		priceCalls.Add(1)
		w.Header().Set("Retry-After", "2")
		w.Header().Set("Content-Length", "10")
		w.WriteHeader(429)
		_, _ = w.Write([]byte("x"))
	}))
	defer server.Close()
	pool := x509.NewCertPool()
	pool.AddCert(server.Certificate())
	want := server.Listener.Addr().String()
	tr := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool}, DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
		if network != "tcp" || address != want {
			return nil, errors.New("unexpected destination")
		}
		return (&net.Dialer{}).DialContext(ctx, network, want)
	}}
	client, err := f.NewClient(official.Credentials{APIKey: "fake", SecretKey: "fake"}, "", official.WithBaseURL(server.URL), official.WithHTTPClient(&http.Client{Transport: tr}), official.WithAccountSeq(1))
	if err != nil {
		t.Fatal(err)
	}
	retrier, err := NewEnginePollRetrier(&Retrier{AccountRef: "fixture"}, b, client, "fixture", 1)
	if err != nil {
		t.Fatal(err)
	}
	err = retrier.Query(context.Background(), QueryPrice, func(ctx context.Context) error { _, e := client.Prices(ctx, []string{"005930"}); return e })
	if !errors.Is(err, official.ErrRateLimited) || !errors.Is(err, official.ErrTransport) || tokenCalls.Load() != 1 || priceCalls.Load() != 1 {
		t.Fatalf("err=%v token=%d price=%d", err, tokenCalls.Load(), priceCalls.Load())
	}
	key, e := b.EnginePollKey("fixture", string(QueryPrice))
	if e != nil {
		t.Fatal(e)
	}
	remaining, e := b.PollRemaining(context.Background(), key)
	if e != nil || remaining <= 0 || remaining > 2*time.Second {
		t.Fatalf("date hint remaining=%v err=%v", remaining, e)
	}
}

func TestA166EnginePoll503RetriesWithout429Event(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "group")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	if err := accountbudget.Provision(root, "scope", "instance"); err != nil {
		t.Fatal(err)
	}
	b, err := accountbudget.Open(accountbudget.Config{Root: root, Scope: "scope", Instance: "instance", Clock: clock.System()})
	if err != nil {
		t.Fatal(err)
	}
	event, err := accountbudget.ProvisionObservation(b)
	err = errors.Join(err, b.Close())
	if err != nil {
		t.Fatal(err)
	}
	id := func(path string) accountbudget.FileID {
		st, e := os.Stat(path)
		if e != nil {
			t.Fatal(e)
		}
		v := st.Sys().(*syscall.Stat_t)
		return accountbudget.FileID{Device: uint64(v.Dev), Inode: v.Ino}
	}
	data, err := json.Marshal(map[string]any{"version": 2, "root": root, "scope": "scope", "instance": "instance", "event_lock": event, "identity": accountbudget.Identity{Root: id(root), Queue: id(filepath.Join(root, "queue.lock")), Wire: id(filepath.Join(root, "wire.lock"))}})
	if err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(dir, "manifest")
	if err = os.WriteFile(manifest, data, 0600); err != nil {
		t.Fatal(err)
	}
	f, err := accountbudgetruntime.Open(context.Background(), accountbudgetruntime.Config{Mode: accountbudgetruntime.ModeRequired, Manifest: manifest, CleanupError: func(error) {}})
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	b, err = f.Budget()
	if err != nil {
		t.Fatal(err)
	}
	var tokenCalls, priceCalls atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth2/token" {
			tokenCalls.Add(1)
			fmt.Fprint(w, `{"access_token":"fake","expires_in":3600,"token_type":"Bearer"}`)
			return
		}
		priceCalls.Add(1)
		w.WriteHeader(503)
	}))
	defer server.Close()
	pool := x509.NewCertPool()
	pool.AddCert(server.Certificate())
	want := server.Listener.Addr().String()
	tr := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool}, DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
		if network != "tcp" || address != want {
			return nil, errors.New("unexpected destination")
		}
		return (&net.Dialer{}).DialContext(ctx, network, want)
	}}
	client, err := f.NewClient(official.Credentials{APIKey: "fake", SecretKey: "fake"}, "", official.WithBaseURL(server.URL), official.WithHTTPClient(&http.Client{Transport: tr}), official.WithAccountSeq(1))
	if err != nil {
		t.Fatal(err)
	}
	retrier, err := NewEnginePollRetrier(&Retrier{AccountRef: "fixture"}, b, client, "fixture", 1)
	if err != nil {
		t.Fatal(err)
	}
	retrier.Policy = RetryPolicy{MaxAttempts: 2, Budget: 4 * time.Second, BaseBackoff: time.Millisecond, MaxBackoff: time.Millisecond}
	err = retrier.Query(context.Background(), QueryPrice, func(ctx context.Context) error { _, e := client.Prices(ctx, []string{"005930"}); return e })
	if !errors.Is(err, official.ErrServer) || tokenCalls.Load() != 1 || priceCalls.Load() != 2 {
		t.Fatalf("err=%v token=%d price=%d", err, tokenCalls.Load(), priceCalls.Load())
	}
	key, e := b.EnginePollKey("fixture", string(QueryPrice))
	if e != nil {
		t.Fatal(e)
	}
	remaining, e := b.PollRemaining(context.Background(), key)
	if e != nil || remaining != 0 {
		t.Fatalf("503 hint remaining=%v err=%v", remaining, e)
	}
	store, e := accountbudget.OpenObservation(context.Background(), b, event, clock.SystemBootSource())
	if e != nil {
		t.Fatal(e)
	}
	snapshot, e := store.Snapshot(context.Background(), accountbudget.ObservationCursor{})
	if e != nil || len(snapshot.Events) != 0 {
		t.Fatalf("503 event=%d err=%v", len(snapshot.Events), e)
	}

}

func TestA166EnginePollHintExpiryRestoresFreshness(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "group")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	if err := accountbudget.Provision(root, "scope", "instance"); err != nil {
		t.Fatal(err)
	}
	b, err := accountbudget.Open(accountbudget.Config{Root: root, Scope: "scope", Instance: "instance", Clock: clock.System()})
	if err != nil {
		t.Fatal(err)
	}
	event, err := accountbudget.ProvisionObservation(b)
	err = errors.Join(err, b.Close())
	if err != nil {
		t.Fatal(err)
	}
	id := func(path string) accountbudget.FileID {
		st, e := os.Stat(path)
		if e != nil {
			t.Fatal(e)
		}
		v := st.Sys().(*syscall.Stat_t)
		return accountbudget.FileID{Device: uint64(v.Dev), Inode: v.Ino}
	}
	data, err := json.Marshal(map[string]any{"version": 2, "root": root, "scope": "scope", "instance": "instance", "event_lock": event, "identity": accountbudget.Identity{Root: id(root), Queue: id(filepath.Join(root, "queue.lock")), Wire: id(filepath.Join(root, "wire.lock"))}})
	if err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(dir, "manifest")
	if err = os.WriteFile(manifest, data, 0600); err != nil {
		t.Fatal(err)
	}
	f, err := accountbudgetruntime.Open(context.Background(), accountbudgetruntime.Config{Mode: accountbudgetruntime.ModeRequired, Manifest: manifest, CleanupError: func(error) {}})
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	b, err = f.Budget()
	if err != nil {
		t.Fatal(err)
	}
	var tokenCalls, priceCalls atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth2/token" {
			tokenCalls.Add(1)
			fmt.Fprint(w, `{"access_token":"fake","expires_in":3600,"token_type":"Bearer"}`)
			return
		}
		n := priceCalls.Add(1)
		if n == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(429)
			return
		}
		fmt.Fprint(w, `{"result":[{"symbol":"005930","lastPrice":"72000","currency":"KRW"}]}`)
	}))
	defer server.Close()
	pool := x509.NewCertPool()
	pool.AddCert(server.Certificate())
	want := server.Listener.Addr().String()
	tr := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool}, DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
		if network != "tcp" || address != want {
			return nil, errors.New("unexpected destination")
		}
		return (&net.Dialer{}).DialContext(ctx, network, want)
	}}
	client, err := f.NewClient(official.Credentials{APIKey: "fake", SecretKey: "fake"}, "", official.WithBaseURL(server.URL), official.WithHTTPClient(&http.Client{Transport: tr}), official.WithAccountSeq(1))
	if err != nil {
		t.Fatal(err)
	}
	retrier, err := NewEnginePollRetrier(&Retrier{AccountRef: "fixture"}, b, client, "fixture", 1)
	if err != nil {
		t.Fatal(err)
	}
	gate := NewEntryGate(clock.System(), map[RequiredQuery]time.Duration{QueryPrice: time.Minute})
	retrier.Gate = gate
	read := func(ctx context.Context) error { _, e := client.Prices(ctx, []string{"005930"}); return e }
	err = retrier.Query(context.Background(), QueryPrice, read)
	if !errors.Is(err, official.ErrRateLimited) || gate.CheckEntry() == nil {
		t.Fatalf("429 err=%v gate=%v", err, gate.CheckEntry())
	}
	err = retrier.Query(context.Background(), QueryPrice, read)
	if !errors.Is(err, accountbudget.ErrObservationDeferred) || priceCalls.Load() != 1 || gate.CheckEntry() == nil {
		t.Fatalf("deferred err=%v wire=%d", err, priceCalls.Load())
	}
	time.Sleep(1100 * time.Millisecond)
	err = retrier.Query(context.Background(), QueryPrice, read)
	if err != nil || priceCalls.Load() != 2 || gate.CheckEntry() != nil {
		t.Fatalf("recovered err=%v wire=%d gate=%v", err, priceCalls.Load(), gate.CheckEntry())
	}
}
