//go:build linux

package accountbudgetruntime

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"
)

func TestA166RuntimeV2RejectsLegacyManifest(t *testing.T) {
	_, path, m := setupGroup(t)
	m.Version = 1
	writeManifest(t, path, m)
	f, err := Open(context.Background(), Config{Mode: ModeRequired, Manifest: path, CleanupError: func(error) {}})
	if f != nil {
		_ = f.Close()
	}
	if err == nil {
		t.Fatal("v1 manifest accepted")
	}
}

func TestA166RuntimeV2EventPinAndDirty(t *testing.T) {
	root, path, m := setupGroup(t)
	for _, change := range []func(*manifest){func(m *manifest) { m.EventLock = accountbudget.FileID{} }, func(m *manifest) { m.EventLock.Inode++ }, func(m *manifest) { m.Version = 1 }} {
		bad := m
		change(&bad)
		writeManifest(t, path, bad)
		f, err := Open(context.Background(), Config{Mode: ModeRequired, Manifest: path, CleanupError: func(error) {}})
		if f != nil || err == nil {
			t.Fatalf("accepted invalid pin: %v", err)
		}
	}
	writeManifest(t, path, m)
	if err := os.WriteFile(filepath.Join(root, "event.dirty"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	if f, err := Open(context.Background(), Config{Mode: ModeRequired, Manifest: path, CleanupError: func(error) {}}); f != nil || err == nil {
		t.Fatal("dirty event accepted")
	}
}

func TestA166RuntimeV2StrictEventShape(t *testing.T) {
	_, path, m := setupGroup(t)
	raw, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	var object map[string]json.RawMessage
	if err = json.Unmarshal(raw, &object); err != nil {
		t.Fatal(err)
	}
	cases := [][]byte{bytes.Replace(raw, []byte(`"event_lock":{`), []byte(`"event_lock":{"device":1,`), 1), bytes.Replace(raw, []byte(`"event_lock":{`), []byte(`"event_lock":{"unknown":1,`), 1)}
	for _, value := range []json.RawMessage{nil, []byte(`null`), []byte(`[]`), []byte(`"bad"`), []byte(`{"device":"bad","inode":1}`)} {
		copyObject := make(map[string]json.RawMessage, len(object))
		for k, v := range object {
			copyObject[k] = v
		}
		if value == nil {
			delete(copyObject, "event_lock")
		} else {
			copyObject["event_lock"] = value
		}
		data, err := json.Marshal(copyObject)
		if err != nil {
			t.Fatal(err)
		}
		cases = append(cases, data)
	}
	for _, data := range cases {
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
		if f, err := Open(context.Background(), Config{Mode: ModeRequired, Manifest: path, CleanupError: func(error) {}}); f != nil || err == nil {
			t.Fatalf("invalid event accepted %s", data)
		}
	}
}

func a166RuntimeClient(t *testing.T, f *Factory, server *httptest.Server, cache string) *official.Client {
	t.Helper()
	pool := x509.NewCertPool()
	pool.AddCert(server.Certificate())
	want := server.Listener.Addr().String()
	tr := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool}, DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
		if network != "tcp" || address != want {
			return nil, errors.New("unexpected destination")
		}
		return (&net.Dialer{}).DialContext(ctx, network, address)
	}}
	c, err := f.NewClient(official.Credentials{APIKey: "fake", SecretKey: "fake"}, cache, official.WithBaseURL(server.URL), official.WithHTTPClient(&http.Client{Transport: tr}))
	if err != nil {
		t.Fatal(err)
	}
	return c
}
func a166RuntimeStore(t *testing.T, f *Factory, m manifest) *accountbudget.ObservationStore {
	t.Helper()
	b, err := f.Budget()
	if err != nil {
		t.Fatal(err)
	}
	s, err := accountbudget.OpenObservation(context.Background(), b, m.EventLock, clock.SystemBootSource())
	if err != nil {
		t.Fatal(err)
	}
	return s
}
func TestA166RuntimeToken429PreservesPollCollector(t *testing.T) {
	_, path, m := setupGroup(t)
	f, err := Open(context.Background(), Config{Mode: ModeRequired, Manifest: path, CleanupError: func(error) {}})
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var hits atomic.Int32
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		if r.URL.Path != "/oauth2/token" {
			t.Error("unexpected route")
		}
		w.Header().Set("Retry-After", "20")
		w.WriteHeader(429)
	}))
	defer srv.Close()
	c := a166RuntimeClient(t, f, srv, filepath.Join(t.TempDir(), "token"))
	var key accountbudget.ObservationKey
	key[0] = 9
	collector := &accountbudget.ObservationCollector{}
	ctx := accountbudget.WithPriority(context.Background(), accountbudget.Critical)
	ctx = accountbudget.WithObservation(ctx, accountbudget.ObservationPoll, key, collector)
	_, err = c.Accounts(ctx)
	if !errors.Is(err, official.ErrRateLimited) || hits.Load() != 1 || !collector.Snapshot().RateLimited {
		t.Fatalf("err %v hits %d result %+v", err, hits.Load(), collector.Snapshot())
	}
	s := a166RuntimeStore(t, f, m)
	view, err := s.Snapshot(context.Background(), accountbudget.ObservationCursor{})
	if err != nil || len(view.Events) != 1 || view.Events[0].Kind != accountbudget.ObservationToken {
		t.Fatalf("events %+v err %v", view, err)
	}
	remaining, err := s.PollRemaining(context.Background(), key)
	if err != nil || remaining != 0 {
		t.Fatalf("poll hint %v %v", remaining, err)
	}
	child := exec.Command(os.Args[0], "-test.run=^TestA166RuntimeEventChild$")
	child.Env = append(os.Environ(), "A166_EVENT_CHILD_MANIFEST="+path)
	if output, err := child.CombinedOutput(); err != nil {
		t.Fatalf("event child: %v %s", err, output)
	}
}
func TestA166RuntimeEventChild(t *testing.T) {
	path := os.Getenv("A166_EVENT_CHILD_MANIFEST")
	if path == "" {
		return
	}
	m, err := readManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	f, err := Open(context.Background(), Config{Mode: ModeRequired, Manifest: path, CleanupError: func(error) {}})
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	view, err := a166RuntimeStore(t, f, m).Snapshot(context.Background(), accountbudget.ObservationCursor{})
	if err != nil || len(view.Events) != 1 || view.Events[0].Kind != accountbudget.ObservationToken {
		t.Fatalf("child event %+v %v", view, err)
	}
}
func TestA166RuntimeToken401AndCached(t *testing.T) {
	_, path, m := setupGroup(t)
	f, err := Open(context.Background(), Config{Mode: ModeRequired, Manifest: path, CleanupError: func(error) {}})
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var token, gets atomic.Int32
	store := a166RuntimeStore(t, f, m)
	kinds := make(chan accountbudget.ObservationKind, 5)
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		view, err := store.Snapshot(r.Context(), accountbudget.ObservationCursor{})
		if err != nil || view.Pending == nil {
			t.Errorf("pending observation: %v", err)
			return
		}
		kinds <- view.Pending.Kind
		switch r.URL.Path {
		case "/oauth2/token":
			fmt.Fprintf(w, `{"access_token":"fake-%d","expires_in":3600,"token_type":"Bearer"}`, token.Add(1))
		case "/api/v1/accounts":
			if gets.Add(1) == 1 {
				w.WriteHeader(401)
				return
			}
			fmt.Fprint(w, `{"result":[{"accountNo":"fake","accountSeq":7,"accountType":"BROKERAGE"}]}`)
		default:
			t.Error("unexpected route")
		}
	}))
	defer srv.Close()
	c := a166RuntimeClient(t, f, srv, filepath.Join(t.TempDir(), "token"))
	for i := 0; i < 2; i++ {
		if _, err = c.Accounts(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	if token.Load() != 2 || gets.Load() != 3 {
		t.Fatalf("token %d gets %d", token.Load(), gets.Load())
	}
	for _, want := range []accountbudget.ObservationKind{accountbudget.ObservationToken, accountbudget.ObservationOther, accountbudget.ObservationToken, accountbudget.ObservationOther, accountbudget.ObservationOther} {
		if got := <-kinds; got != want {
			t.Fatalf("kind %v want %v", got, want)
		}
	}
	view, err := store.Snapshot(context.Background(), accountbudget.ObservationCursor{})
	if err != nil || view.Pending != nil || view.Last == nil || view.Last.Attempt != 5 || len(view.Events) != 0 {
		t.Fatalf("view %+v err %v", view, err)
	}
}

func TestA166RuntimeV2ReplacedEventLock(t *testing.T) {
	root, path, _ := setupGroup(t)
	lock := filepath.Join(root, "event.lock")
	if err := os.Rename(lock, lock+".old"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(lock, nil, 0600); err != nil {
		t.Fatal(err)
	}
	if f, err := Open(context.Background(), Config{Mode: ModeRequired, Manifest: path, CleanupError: func(error) {}}); f != nil || err == nil {
		t.Fatal("replacement event lock accepted")
	}
}
func TestA166RuntimeGet429BodyFailure(t *testing.T) {
	_, path, m := setupGroup(t)
	f, err := Open(context.Background(), Config{Mode: ModeRequired, Manifest: path, CleanupError: func(error) {}})
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			fmt.Fprint(w, `{"access_token":"fake","expires_in":3600,"token_type":"Bearer"}`)
		case "/api/v1/accounts":
			w.Header().Set("Content-Length", "10")
			w.Header().Set("Retry-After", "5")
			w.WriteHeader(429)
			fmt.Fprint(w, "x")
		default:
			t.Error("unexpected route")
		}
	}))
	defer srv.Close()
	c := a166RuntimeClient(t, f, srv, filepath.Join(t.TempDir(), "token"))
	var key accountbudget.ObservationKey
	key[0] = 3
	collector := &accountbudget.ObservationCollector{}
	_, err = c.Accounts(accountbudget.WithObservation(context.Background(), accountbudget.ObservationPoll, key, collector))
	if err == nil || !collector.Snapshot().RateLimited {
		t.Fatalf("body error %v collector %+v", err, collector.Snapshot())
	}
	store := a166RuntimeStore(t, f, m)
	view, err := store.Snapshot(context.Background(), accountbudget.ObservationCursor{})
	if err != nil || len(view.Events) != 1 || view.Events[0].Kind != accountbudget.ObservationPoll {
		t.Fatalf("event %+v %v", view, err)
	}
	remaining, err := store.PollRemaining(context.Background(), key)
	if err != nil || remaining <= 0 {
		t.Fatalf("hint %v %v", remaining, err)
	}
}

func TestA166RuntimeTwoClientsShareEvents(t *testing.T) {
	_, path, m := setupGroup(t)
	f, err := Open(context.Background(), Config{Mode: ModeRequired, Manifest: path, CleanupError: func(error) {}})
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth2/token" {
			t.Error("unexpected route")
		}
		w.WriteHeader(429)
	}))
	defer srv.Close()
	b, err := f.Budget()
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		c := a166RuntimeClient(t, f, srv, filepath.Join(t.TempDir(), "token"))
		if !c.UsesAccountBudget(b) {
			t.Fatal("different budget")
		}
		if _, err = c.Accounts(context.Background()); !errors.Is(err, official.ErrRateLimited) {
			t.Fatal(err)
		}
	}
	view, err := a166RuntimeStore(t, f, m).Snapshot(context.Background(), accountbudget.ObservationCursor{})
	if err != nil || len(view.Events) != 2 || view.Events[0].Kind != accountbudget.ObservationToken || view.Events[1].Kind != accountbudget.ObservationToken {
		t.Fatalf("events %+v %v", view, err)
	}
}

func TestA166RuntimeV2FailureClosesRootFD(t *testing.T) {
	root, path, m := setupGroup(t)
	count := func() int {
		entries, err := os.ReadDir("/proc/self/fd")
		if err != nil {
			t.Fatal(err)
		}
		n := 0
		for _, entry := range entries {
			target, err := os.Readlink(filepath.Join("/proc/self/fd", entry.Name()))
			if err == nil && (target == root || strings.HasPrefix(target, root+"/")) {
				n++
			}
		}
		return n
	}
	before := count()
	for _, bad := range []manifest{func() manifest { x := m; x.EventLock.Inode++; return x }(), func() manifest { x := m; x.Identity.Wire.Inode++; return x }()} {
		writeManifest(t, path, bad)
		for i := 0; i < 4; i++ {
			if f, err := Open(context.Background(), Config{Mode: ModeRequired, Manifest: path, CleanupError: func(error) {}}); f != nil || err == nil {
				t.Fatal("invalid identity opened")
			}
		}
		if got := count(); got != before {
			t.Fatalf("root fd leaked: before=%d after=%d", before, got)
		}
	}
}

func TestA166RuntimeToken429FinishFaultPreservesStatus(t *testing.T) {
	root, path, _ := setupGroup(t)
	faults := make(chan error, 2)
	f, err := Open(context.Background(), Config{Mode: ModeRequired, Manifest: path, CleanupError: func(err error) { faults <- err }})
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth2/token" {
			t.Error("unexpected route")
		}
		if err := os.WriteFile(filepath.Join(root, "event.dirty"), nil, 0600); err != nil {
			t.Error(err)
		}
		w.WriteHeader(429)
	}))
	defer srv.Close()
	c := a166RuntimeClient(t, f, srv, filepath.Join(t.TempDir(), "token"))
	collector := &accountbudget.ObservationCollector{}
	ctx := accountbudget.WithObservation(context.Background(), accountbudget.ObservationOther, accountbudget.ObservationKey{}, collector)
	_, err = c.Accounts(ctx)
	if !errors.Is(err, official.ErrRateLimited) || errors.Is(err, accountbudget.ErrNotSent) || !collector.Snapshot().RateLimited || collector.Snapshot().LocalError == nil {
		t.Fatalf("err %v collector %+v", err, collector.Snapshot())
	}
	select {
	case <-faults:
	default:
		t.Fatal("missing cleanup callback")
	}
}

func TestA166RuntimeV2A112CloneRecordsPending(t *testing.T) {
	_, path, m := setupGroup(t)
	f, err := Open(context.Background(), Config{Mode: ModeRequired, Manifest: path, CleanupError: func(error) {}})
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	cache := filepath.Join(t.TempDir(), "token")
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
	store := a166RuntimeStore(t, f, m)
	var calls atomic.Int32
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		view, err := store.Snapshot(r.Context(), accountbudget.ObservationCursor{})
		if err != nil || view.Pending == nil || view.Pending.Kind != accountbudget.ObservationOther {
			t.Errorf("pending %+v %v", view.Pending, err)
		}
		if r.Method != "GET" || r.URL.Path != "/api/v1/candles" {
			t.Error("unexpected route")
		}
		w.Header().Set("X-Ratelimit-Limit", "10")
		w.Header().Set("X-Ratelimit-Remaining", "9")
		w.Header().Set("X-Ratelimit-Reset", "1")
		fmt.Fprint(w, `{"result":{"nextBefore":null}}`)
	}))
	srv.TLS = &tls.Config{Certificates: []tls.Certificate{{Certificate: [][]byte{der}, PrivateKey: priv}}}
	srv.StartTLS()
	defer srv.Close()
	c, err := f.NewClient(official.Credentials{APIKey: "fake", SecretKey: "fake"}, cache)
	if err != nil {
		t.Fatal(err)
	}
	b, err := f.Budget()
	if err != nil || !c.UsesAccountBudget(b) {
		t.Fatal("factory binding")
	}
	field := reflect.ValueOf(c).Elem().FieldByName("authorityTransport")
	if !field.IsValid() || !field.CanAddr() || field.Type() != reflect.TypeOf((*http.Transport)(nil)) {
		t.Fatal("authority field shape changed")
	}
	tr := *(**http.Transport)(unsafe.Pointer(field.UnsafeAddr()))
	if tr == nil {
		t.Fatal("authority transport missing")
	}
	pool := x509.NewCertPool()
	pool.AddCert(cert)
	tr.TLSClientConfig = &tls.Config{RootCAs: pool}
	want := srv.Listener.Addr().String()
	tr.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		if network != "tcp" || address != "openapi.tossinvest.com:443" {
			return nil, errors.New("unexpected authority address")
		}
		return (&net.Dialer{}).DialContext(ctx, network, want)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = official.A112MBUSCandle(ctx, c, nil)
	if err != nil || calls.Load() != 1 {
		t.Fatalf("a112 %v calls %d", err, calls.Load())
	}
	view, err := store.Snapshot(context.Background(), accountbudget.ObservationCursor{})
	if err != nil || view.Pending != nil || view.Last == nil || view.Last.Attempt != 1 {
		t.Fatalf("final %+v %v", view, err)
	}
}
