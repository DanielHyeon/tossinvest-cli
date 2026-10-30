//go:build linux

package accountbudgetruntime

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
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"

	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"golang.org/x/sys/unix"
)

func TestA166ConsumerTwoFactoriesPriceLimits(t *testing.T) {
	_, path, _ := setupGroup(t)
	cfg := Config{Mode: ModeRequired, Manifest: path, CleanupError: func(error) {}}
	a, err := Open(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	z, err := Open(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	ba, err := a.Budget()
	if err != nil {
		t.Fatal(err)
	}
	bz, err := z.Budget()
	if err != nil {
		t.Fatal(err)
	}
	if ba == bz {
		t.Fatal("independent factories reused pointer")
	}
	var tokens, prices atomic.Int32
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			tokens.Add(1)
			fmt.Fprint(w, `{"access_token":"fake","expires_in":3600,"token_type":"Bearer"}`)
		case "/api/v1/price-limits":
			if prices.Add(1) == 1 {
				w.WriteHeader(429)
				return
			}
			fmt.Fprint(w, `{"result":{"lowerLimitPrice":"50400","timestamp":"2026-03-30T09:30:00+09:00"}}`)
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))
	defer srv.Close()
	roots := x509.NewCertPool()
	roots.AddCert(srv.Certificate())
	want := srv.Listener.Addr().String()
	tr := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: roots}, DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
		if network != "tcp" || address != want {
			return nil, errors.New("fixture refused address")
		}
		return (&net.Dialer{}).DialContext(ctx, network, want)
	}}
	opts := []official.Option{official.WithBaseURL(srv.URL), official.WithHTTPClient(&http.Client{Transport: tr})}
	ca, err := a.NewClient(official.Credentials{APIKey: "fake", SecretKey: "fake"}, "", opts...)
	if err != nil {
		t.Fatal(err)
	}
	cz, err := z.NewClient(official.Credentials{APIKey: "fake", SecretKey: "fake"}, "", opts...)
	if err != nil {
		t.Fatal(err)
	}
	if !ca.UsesAccountBudget(ba) || ca.UsesAccountBudget(bz) || !cz.UsesAccountBudget(bz) {
		t.Fatal("factory client mismatch")
	}
	initial, err := bz.ObservationSnapshot(context.Background(), accountbudget.ObservationCursor{})
	if err != nil || initial.Cursor.Event != 0 {
		t.Fatalf("initial %+v %v", initial, err)
	}
	lease, err := bz.Acquire(context.Background(), accountbudget.Critical)
	if err != nil {
		t.Fatal(err)
	}
	ctx := accountbudget.WithPriority(context.Background(), accountbudget.Auxiliary)
	_, err = ca.ProtectivePriceLimits(ctx, "005930")
	if err == nil || tokens.Load() != 0 || prices.Load() != 0 {
		t.Fatalf("busy err=%v token=%d get=%d", err, tokens.Load(), prices.Load())
	}
	if err = lease.Release(); err != nil {
		t.Fatal(err)
	}
	_, err = ca.ProtectivePriceLimits(ctx, "005930")
	if !errors.Is(err, official.ErrRateLimited) || tokens.Load() != 1 || prices.Load() != 1 {
		t.Fatalf("429 err=%v token=%d get=%d", err, tokens.Load(), prices.Load())
	}
	view, err := bz.ObservationSnapshot(context.Background(), initial.Cursor)
	if err != nil || len(view.Events) != 1 || view.Events[0].Kind != accountbudget.ObservationOther || view.Cursor.Event != 1 {
		t.Fatalf("view=%+v err=%v", view, err)
	}
	result, err := ca.ProtectivePriceLimits(ctx, "005930")
	if err != nil || result.LowerLimitPrice == nil || *result.LowerLimitPrice != "50400" || tokens.Load() != 1 || prices.Load() != 2 {
		t.Fatalf("success err=%v token=%d get=%d", err, tokens.Load(), prices.Load())
	}
	after, err := bz.ObservationSnapshot(context.Background(), view.Cursor)
	if err != nil || len(after.Events) != 0 || after.Cursor.Event != 1 {
		t.Fatalf("after=%+v err=%v", after, err)
	}
}

func TestA166ConsumerPriceLimits401EachWire(t *testing.T) {
	_, path, _ := setupGroup(t)
	cfg := Config{Mode: ModeRequired, Manifest: path, CleanupError: func(error) {}}
	a, err := Open(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	z, err := Open(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	ba, err := a.Budget()
	if err != nil {
		t.Fatal(err)
	}
	bz, err := z.Budget()
	if err != nil {
		t.Fatal(err)
	}
	var tokens, prices atomic.Int32
	var mu sync.Mutex
	var attempts []accountbudget.ObservationPending
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		view, e := bz.ObservationSnapshot(r.Context(), accountbudget.ObservationCursor{})
		if e != nil || view.Pending == nil {
			t.Errorf("missing pending: %v", e)
			return
		}
		mu.Lock()
		attempts = append(attempts, *view.Pending)
		mu.Unlock()
		switch r.URL.Path {
		case "/oauth2/token":
			tokens.Add(1)
			fmt.Fprint(w, `{"access_token":"fake","expires_in":3600,"token_type":"Bearer"}`)
		case "/api/v1/price-limits":
			if prices.Add(1) == 1 {
				w.WriteHeader(401)
				return
			}
			fmt.Fprint(w, `{"result":{"lowerLimitPrice":"50400","timestamp":"2026-03-30T09:30:00+09:00"}}`)
		default:
			t.Error("unexpected request")
		}
	}))
	defer srv.Close()
	roots := x509.NewCertPool()
	roots.AddCert(srv.Certificate())
	want := srv.Listener.Addr().String()
	tr := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: roots}, DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
		if network != "tcp" || address != want {
			return nil, errors.New("fixture refused address")
		}
		return (&net.Dialer{}).DialContext(ctx, network, want)
	}}
	c, err := a.NewClient(official.Credentials{APIKey: "fake", SecretKey: "fake"}, "", official.WithBaseURL(srv.URL), official.WithHTTPClient(&http.Client{Transport: tr}))
	if err != nil {
		t.Fatal(err)
	}
	if !c.UsesAccountBudget(ba) {
		t.Fatal("client budget mismatch")
	}
	field := reflect.ValueOf(c).Elem().FieldByName("budgetTransport")
	guard := *(**accountbudget.Transport)(unsafe.Pointer(field.UnsafeAddr()))
	first := reflect.ValueOf(guard).Elem().FieldByName("firstWrite")
	callback := (*func())(unsafe.Pointer(first.UnsafeAddr()))
	var writes []time.Time
	*callback = func() { mu.Lock(); writes = append(writes, time.Now()); mu.Unlock() }
	ctx := accountbudget.WithPriority(context.Background(), accountbudget.Auxiliary)
	result, err := c.ProtectivePriceLimits(ctx, "005930")
	if err != nil || result.LowerLimitPrice == nil || tokens.Load() != 2 || prices.Load() != 2 {
		t.Fatalf("err=%v token=%d prices=%d", err, tokens.Load(), prices.Load())
	}
	mu.Lock()
	defer mu.Unlock()
	if len(writes) != 4 || len(attempts) != 4 {
		t.Fatalf("writes=%d attempts=%d", len(writes), len(attempts))
	}
	for i := range writes {
		if attempts[i].Sequence != uint64(i+1) {
			t.Fatalf("attempts=%+v", attempts)
		}
		if i > 0 && writes[i].Sub(writes[i-1]) < 350*time.Millisecond {
			t.Fatalf("write gap %v", writes[i].Sub(writes[i-1]))
		}
	}
	if attempts[0].Kind != accountbudget.ObservationToken || attempts[1].Kind != accountbudget.ObservationOther || attempts[2].Kind != accountbudget.ObservationToken || attempts[3].Kind != accountbudget.ObservationOther {
		t.Fatalf("kinds=%+v", attempts)
	}
}

func TestA166ConsumerPriceLimitsBodyOwnership(t *testing.T) {
	for _, mode := range []string{"eof", "cancel"} {
		t.Run(mode, func(t *testing.T) {
			_, path, _ := setupGroup(t)
			f, err := Open(context.Background(), Config{Mode: ModeRequired, Manifest: path, CleanupError: func(error) {}})
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			b, err := f.Budget()
			if err != nil {
				t.Fatal(err)
			}
			started := make(chan struct{})
			release := make(chan struct{})
			var once sync.Once
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/oauth2/token":
					fmt.Fprint(w, `{"access_token":"fake","expires_in":3600,"token_type":"Bearer"}`)
				case "/api/v1/price-limits":
					w.Header().Set("Content-Type", "application/json")
					_, _ = fmt.Fprint(w, `{"result":{"lowerLimitPrice":"50400",`)
					w.(http.Flusher).Flush()
					once.Do(func() { close(started) })
					select {
					case <-release:
						_, _ = fmt.Fprint(w, `"timestamp":"2026-03-30T09:30:00+09:00"}}`)
					case <-r.Context().Done():
					}
				default:
					t.Error("unexpected request")
				}
			}))
			defer srv.Close()
			roots := x509.NewCertPool()
			roots.AddCert(srv.Certificate())
			want := srv.Listener.Addr().String()
			tr := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: roots}, DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
				if network != "tcp" || address != want {
					return nil, errors.New("fixture refused address")
				}
				return (&net.Dialer{}).DialContext(ctx, network, want)
			}}
			c, err := f.NewClient(official.Credentials{APIKey: "fake", SecretKey: "fake"}, "", official.WithBaseURL(srv.URL), official.WithHTTPClient(&http.Client{Transport: tr}))
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(accountbudget.WithPriority(context.Background(), accountbudget.Auxiliary))
			defer cancel()
			done := make(chan error, 1)
			go func() { _, e := c.ProtectivePriceLimits(ctx, "005930"); done <- e }()
			select {
			case <-started:
			case <-time.After(5 * time.Second):
				t.Fatal("no price request")
			}
			bounded, stop := context.WithTimeout(context.Background(), 100*time.Millisecond)
			lease, e := b.Acquire(bounded, accountbudget.Critical)
			stop()
			if lease != nil {
				_ = lease.Release()
			}
			if !errors.Is(e, context.DeadlineExceeded) {
				t.Fatalf("critical escaped active body: %v", e)
			}
			if mode == "cancel" {
				cancel()
			} else {
				close(release)
			}
			select {
			case e := <-done:
				if mode == "eof" && e != nil {
					t.Fatal(e)
				}
				if mode == "cancel" && e == nil {
					t.Fatal("cancel returned success")
				}
			case <-time.After(5 * time.Second):
				t.Fatal("request did not finish")
			}
			if mode == "cancel" {
				close(release)
			}
			bounded, stop = context.WithTimeout(context.Background(), 3*time.Second)
			lease, e = b.Acquire(bounded, accountbudget.Critical)
			stop()
			if e != nil {
				t.Fatal(e)
			}
			if e = lease.Release(); e != nil {
				t.Fatal(e)
			}
		})
	}
}

func TestA166ConsumerTokenSentThenAuxGETRefused(t *testing.T) {
	root, path, _ := setupGroup(t)
	f, err := Open(context.Background(), Config{Mode: ModeRequired, Manifest: path, CleanupError: func(error) {}})
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	b, err := f.Budget()
	if err != nil {
		t.Fatal(err)
	}
	var tokens, gets atomic.Int32
	leaseCh := make(chan *accountbudget.Lease, 1)
	errorCh := make(chan error, 1)
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			tokens.Add(1)
			go func() {
				lease, e := b.Acquire(context.Background(), accountbudget.Critical)
				if e != nil {
					errorCh <- e
					return
				}
				leaseCh <- lease
			}()
			deadline := time.Now().Add(3 * time.Second)
			queued := false
			for time.Now().Before(deadline) {
				raw, e := os.ReadFile(filepath.Join(root, "state"))
				if e == nil {
					var state struct {
						Tickets []json.RawMessage `json:"tickets"`
					}
					if json.Unmarshal(raw, &state) == nil && len(state.Tickets) > 0 {
						queued = true
						break
					}
				}
				time.Sleep(time.Millisecond)
			}
			if !queued {
				t.Error("critical did not queue")
				return
			}
			fmt.Fprint(w, `{"access_token":"fake","expires_in":3600,"token_type":"Bearer"}`)
		case "/api/v1/price-limits":
			gets.Add(1)
			fmt.Fprint(w, `{"result":{"lowerLimitPrice":"50400"}}`)
		default:
			t.Error("unexpected request")
		}
	}))
	defer srv.Close()
	roots := x509.NewCertPool()
	roots.AddCert(srv.Certificate())
	want := srv.Listener.Addr().String()
	tr := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: roots}, DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
		if network != "tcp" || address != want {
			return nil, errors.New("fixture refused address")
		}
		return (&net.Dialer{}).DialContext(ctx, network, want)
	}}
	c, err := f.NewClient(official.Credentials{APIKey: "fake", SecretKey: "fake"}, "", official.WithBaseURL(srv.URL), official.WithHTTPClient(&http.Client{Transport: tr}))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(accountbudget.WithPriority(context.Background(), accountbudget.Auxiliary), 5*time.Second)
	defer cancel()
	_, err = c.ProtectivePriceLimits(ctx, "005930")
	if err == nil || tokens.Load() != 1 || gets.Load() != 0 {
		t.Fatalf("err=%v token=%d get=%d", err, tokens.Load(), gets.Load())
	}
	select {
	case lease := <-leaseCh:
		if e := lease.Release(); e != nil {
			t.Fatal(e)
		}
	case e := <-errorCh:
		t.Fatal(e)
	case <-time.After(4 * time.Second):
		t.Fatal("critical did not acquire")
	}
}

func TestA166ConsumerFinishFaultPreservesHTTP429(t *testing.T) {
	_, path, _ := setupGroup(t)
	fault := errors.New("fixture postcommit fsync")
	reported := make(chan error, 2)
	f, err := Open(context.Background(), Config{Mode: ModeRequired, Manifest: path, CleanupError: func(e error) { reported <- e }})
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	b, err := f.Budget()
	if err != nil {
		t.Fatal(err)
	}
	observation := reflect.ValueOf(b).Elem().FieldByName("observation")
	store := *(**accountbudget.ObservationStore)(unsafe.Pointer(observation.UnsafeAddr()))
	syncField := reflect.ValueOf(store).Elem().FieldByName("syncDir")
	syncFn := (*func(int) error)(unsafe.Pointer(syncField.UnsafeAddr()))
	var syncCalls atomic.Int32
	*syncFn = func(fd int) error {
		if syncCalls.Add(1) == 12 {
			return fault
		}
		return unix.Fsync(fd)
	}
	var tokens, gets atomic.Int32
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			tokens.Add(1)
			fmt.Fprint(w, `{"access_token":"fake","expires_in":3600,"token_type":"Bearer"}`)
		case "/api/v1/price-limits":
			gets.Add(1)
			w.WriteHeader(429)
		default:
			t.Error("unexpected request")
		}
	}))
	defer srv.Close()
	roots := x509.NewCertPool()
	roots.AddCert(srv.Certificate())
	want := srv.Listener.Addr().String()
	tr := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: roots}, DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
		if network != "tcp" || address != want {
			return nil, errors.New("fixture refused address")
		}
		return (&net.Dialer{}).DialContext(ctx, network, want)
	}}
	c, err := f.NewClient(official.Credentials{APIKey: "fake", SecretKey: "fake"}, "", official.WithBaseURL(srv.URL), official.WithHTTPClient(&http.Client{Transport: tr}))
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.ProtectivePriceLimits(accountbudget.WithPriority(context.Background(), accountbudget.Auxiliary), "005930")
	if !errors.Is(err, official.ErrRateLimited) || errors.Is(err, accountbudget.ErrNotSent) || tokens.Load() != 1 || gets.Load() != 1 {
		t.Fatalf("err=%v token=%d get=%d", err, tokens.Load(), gets.Load())
	}
	select {
	case e := <-reported:
		if !errors.Is(e, fault) {
			t.Fatal(e)
		}
	default:
		t.Fatal("missing cleanup report")
	}
	view, err := b.ObservationSnapshot(context.Background(), accountbudget.ObservationCursor{})
	if err != nil || len(view.Events) != 1 || view.Pending != nil {
		t.Fatalf("view=%+v err=%v", view, err)
	}
}

func TestA166ConsumerToken429CrossFactorySnapshot(t *testing.T) {
	_, path, _ := setupGroup(t)
	cfg := Config{Mode: ModeRequired, Manifest: path, CleanupError: func(error) {}}
	a, err := Open(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	z, err := Open(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	ba, err := a.Budget()
	if err != nil {
		t.Fatal(err)
	}
	bz, err := z.Budget()
	if err != nil {
		t.Fatal(err)
	}
	if ba == bz {
		t.Fatal("factory budget pointers equal")
	}
	var tokens, gets atomic.Int32
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			tokens.Add(1)
			w.WriteHeader(429)
		case "/api/v1/price-limits":
			gets.Add(1)
		default:
			t.Error("unexpected request")
		}
	}))
	defer srv.Close()
	roots := x509.NewCertPool()
	roots.AddCert(srv.Certificate())
	want := srv.Listener.Addr().String()
	tr := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: roots}, DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
		if network != "tcp" || address != want {
			return nil, errors.New("fixture refused address")
		}
		return (&net.Dialer{}).DialContext(ctx, network, want)
	}}
	c, err := a.NewClient(official.Credentials{APIKey: "fake", SecretKey: "fake"}, "", official.WithBaseURL(srv.URL), official.WithHTTPClient(&http.Client{Transport: tr}))
	if err != nil {
		t.Fatal(err)
	}
	if !c.UsesAccountBudget(ba) || c.UsesAccountBudget(bz) {
		t.Fatal("incorrect budget binding")
	}
	before, err := bz.ObservationSnapshot(context.Background(), accountbudget.ObservationCursor{})
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.ProtectivePriceLimits(accountbudget.WithPriority(context.Background(), accountbudget.Auxiliary), "005930")
	if !errors.Is(err, official.ErrRateLimited) || tokens.Load() != 1 || gets.Load() != 0 {
		t.Fatalf("err=%v tokens=%d gets=%d", err, tokens.Load(), gets.Load())
	}
	after, err := bz.ObservationSnapshot(context.Background(), before.Cursor)
	if err != nil || len(after.Events) != 1 || after.Events[0].Kind != accountbudget.ObservationToken || after.Cursor.Event != before.Cursor.Event+1 {
		t.Fatalf("view=%+v err=%v", after, err)
	}
}

func TestA166ConsumerPriceLimitsTwo401AdoptAndMint(t *testing.T) {
	_, path, _ := setupGroup(t)
	f, err := Open(context.Background(), Config{Mode: ModeRequired, Manifest: path, CleanupError: func(error) {}})
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	b, err := f.Budget()
	if err != nil {
		t.Fatal(err)
	}
	cache := filepath.Join(t.TempDir(), "cache.json")
	writeToken := func(value string) {
		data, e := json.Marshal(map[string]any{"access_token": value, "expires_at": time.Now().Add(time.Hour)})
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(cache, data, 0600); e != nil {
			t.Fatal(e)
		}
	}
	writeToken("fake-initial")
	var tokens, gets atomic.Int32
	var mu sync.Mutex
	var attempts []accountbudget.ObservationPending
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		view, e := b.ObservationSnapshot(r.Context(), accountbudget.ObservationCursor{})
		if e != nil || view.Pending == nil {
			t.Errorf("missing pending: %v", e)
			return
		}
		mu.Lock()
		attempts = append(attempts, *view.Pending)
		mu.Unlock()
		switch r.URL.Path {
		case "/oauth2/token":
			tokens.Add(1)
			fmt.Fprint(w, `{"access_token":"fake-minted","expires_in":3600,"token_type":"Bearer"}`)
		case "/api/v1/price-limits":
			n := gets.Add(1)
			if n == 1 {
				writeToken("fake-adopted")
				w.WriteHeader(401)
				return
			}
			if n == 2 {
				w.WriteHeader(401)
				return
			}
			fmt.Fprint(w, `{"result":{"lowerLimitPrice":"50400","timestamp":"2026-03-30T09:30:00+09:00"}}`)
		default:
			t.Error("unexpected request")
		}
	}))
	defer srv.Close()
	roots := x509.NewCertPool()
	roots.AddCert(srv.Certificate())
	want := srv.Listener.Addr().String()
	tr := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: roots}, DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
		if network != "tcp" || address != want {
			return nil, errors.New("fixture refused address")
		}
		return (&net.Dialer{}).DialContext(ctx, network, want)
	}}
	c, err := f.NewClient(official.Credentials{APIKey: "fake", SecretKey: "fake"}, cache, official.WithBaseURL(srv.URL), official.WithHTTPClient(&http.Client{Transport: tr}))
	if err != nil {
		t.Fatal(err)
	}
	field := reflect.ValueOf(c).Elem().FieldByName("budgetTransport")
	guard := *(**accountbudget.Transport)(unsafe.Pointer(field.UnsafeAddr()))
	first := reflect.ValueOf(guard).Elem().FieldByName("firstWrite")
	callback := (*func())(unsafe.Pointer(first.UnsafeAddr()))
	var writes []time.Time
	*callback = func() { mu.Lock(); writes = append(writes, time.Now()); mu.Unlock() }
	collector := &accountbudget.ObservationCollector{}
	ctx := accountbudget.WithObservation(accountbudget.WithPriority(context.Background(), accountbudget.Auxiliary), accountbudget.ObservationOther, accountbudget.ObservationKey{}, collector)
	result, err := c.ProtectivePriceLimits(ctx, "005930")
	if err != nil || result.LowerLimitPrice == nil || gets.Load() != 3 || tokens.Load() != 1 {
		t.Fatalf("err=%v gets=%d tokens=%d", err, gets.Load(), tokens.Load())
	}
	if got := collector.Snapshot(); got.RateLimited || got.LocalError != nil {
		t.Fatalf("collector %+v", got)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(writes) != 4 || len(attempts) != 4 {
		t.Fatalf("writes=%d attempts=%d", len(writes), len(attempts))
	}
	kinds := []accountbudget.ObservationKind{accountbudget.ObservationOther, accountbudget.ObservationOther, accountbudget.ObservationToken, accountbudget.ObservationOther}
	for i := range writes {
		if attempts[i].Sequence != uint64(i+1) || attempts[i].Kind != kinds[i] {
			t.Fatalf("attempts=%+v", attempts)
		}
		if i > 0 && writes[i].Sub(writes[i-1]) < 350*time.Millisecond {
			t.Fatalf("write gap %v", writes[i].Sub(writes[i-1]))
		}
	}
}
