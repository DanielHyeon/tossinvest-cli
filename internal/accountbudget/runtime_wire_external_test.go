//go:build linux

package accountbudget_test

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
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
	"unsafe"

	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget"
	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudgetruntime"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"golang.org/x/sys/unix"
)

func TestA166RuntimeWireSeamAvailable(t *testing.T) {
	if accountbudget.InstallFirstWriteForTest(nil, nil) == nil {
		t.Fatal("nil seam accepted")
	}
}

// private official guard의 실제 포인터만 검사함. 호출 전에 hook 설정 외 제품 객체 변경 없음.
func guardedTransport(t *testing.T, c *official.Client, b *accountbudget.Budget) *accountbudget.Transport {
	t.Helper()
	if c == nil || !c.UsesAccountBudget(b) {
		t.Fatal("factory binding invalid")
	}
	v := reflect.ValueOf(c)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		t.Fatal("client shape changed")
	}
	field := v.Elem().FieldByName("budgetTransport")
	if !field.IsValid() || !field.CanAddr() || field.Type() != reflect.TypeOf((*accountbudget.Transport)(nil)) {
		t.Fatal("private guard field changed")
	}
	guard, ok := reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Interface().(*accountbudget.Transport)
	if !ok || guard == nil {
		t.Fatal("private guard unavailable")
	}
	return guard
}
func TestA166RuntimeWireChild(t *testing.T) {
	root := os.Getenv("RUNTIME_WIRE_ROOT")
	if root == "" {
		return
	}
	id := os.Getenv("RUNTIME_WIRE_ID")
	f, e := accountbudgetruntime.Open(context.Background(), accountbudgetruntime.Config{Mode: accountbudgetruntime.ModeRequired, Manifest: os.Getenv("RUNTIME_WIRE_MANIFEST"), CleanupError: func(e error) { t.Errorf("cleanup: %v", e) }})
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	b, e := f.Budget()
	if e != nil || b == nil {
		t.Fatalf("budget: %v", e)
	}
	cert, e := os.ReadFile(filepath.Join(root, "fixture-cert.pem"))
	if e != nil {
		t.Fatal(e)
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(cert) {
		t.Fatal("fixture certificate")
	}
	want := strings.TrimPrefix(os.Getenv("RUNTIME_WIRE_URL"), "https://")
	host, _, err := net.SplitHostPort(want)
	if err != nil || net.ParseIP(host) == nil || !net.ParseIP(host).IsLoopback() {
		t.Fatal("invalid local fixture address")
	}
	base := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: roots}, DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
		if network != "tcp" || addr != want {
			return nil, errors.New("unexpected fixture address")
		}
		return (&net.Dialer{}).DialContext(ctx, network, addr)
	}}
	client, e := f.NewClient(official.Credentials{APIKey: "fake-" + id, SecretKey: "fake-secret"}, "", official.WithBaseURL(os.Getenv("RUNTIME_WIRE_URL")), official.WithHTTPClient(&http.Client{Transport: base, Timeout: 10 * time.Second}))
	if e != nil {
		t.Fatal(e)
	}
	guard := guardedTransport(t, client, b)
	path := filepath.Join(root, "first-write-"+id)
	e = accountbudget.InstallFirstWriteForTest(guard, func() {
		var ts unix.Timespec
		if err := unix.ClockGettime(unix.CLOCK_MONOTONIC, &ts); err != nil {
			panic(err)
		}
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			panic(err)
		}
		_, err = fmt.Fprintf(f, "%d\n", ts.Sec*int64(time.Second)+ts.Nsec)
		err = errors.Join(err, f.Close())
		if err != nil {
			panic(err)
		}
	})
	if e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	accounts, e := client.Accounts(ctx)
	if e != nil || len(accounts) != 1 {
		t.Fatalf("accounts: %v count=%d", e, len(accounts))
	}
}
func TestA166RuntimeFactoryTwoProcessesTLSStarts(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "budget")
	if e := os.Mkdir(root, 0700); e != nil {
		t.Fatal(e)
	}
	if e := accountbudget.Provision(root, "opaque-scope", "opaque-instance"); e != nil {
		t.Fatal(e)
	}
	b, err := accountbudget.Open(accountbudget.Config{Root: root, Scope: "opaque-scope", Instance: "opaque-instance", Clock: clock.System()})
	if err != nil {
		t.Fatal(err)
	}
	eventID, err := accountbudget.ProvisionObservation(b)
	if err = errors.Join(err, b.Close()); err != nil {
		t.Fatal(err)
	}
	id := func(path string) accountbudget.FileID {
		t.Helper()
		var s unix.Stat_t
		if e := unix.Stat(path, &s); e != nil {
			t.Fatal(e)
		}
		return accountbudget.FileID{Device: uint64(s.Dev), Inode: s.Ino}
	}
	manifest := struct {
		Version   int                    `json:"version"`
		Root      string                 `json:"root"`
		Scope     string                 `json:"scope"`
		Instance  string                 `json:"instance"`
		Identity  accountbudget.Identity `json:"identity"`
		EventLock accountbudget.FileID   `json:"event_lock"`
	}{2, root, "opaque-scope", "opaque-instance", accountbudget.Identity{Root: id(root), Queue: id(filepath.Join(root, "queue.lock")), Wire: id(filepath.Join(root, "wire.lock"))}, eventID}
	data, e := json.Marshal(manifest)
	if e != nil {
		t.Fatal(e)
	}
	manifestPath := filepath.Join(base, "launch.json")
	if e = os.WriteFile(manifestPath, data, 0600); e != nil {
		t.Fatal(e)
	}
	var mu sync.Mutex
	tokens := map[string]int{}
	accounts := map[string]int{}
	var bad []string
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		switch r.URL.Path {
		case "/oauth2/token":
			if e := r.ParseForm(); e != nil {
				bad = append(bad, e.Error())
				w.WriteHeader(500)
				return
			}
			key := r.Form.Get("client_id")
			if key != "fake-0" && key != "fake-1" {
				bad = append(bad, "unexpected key")
				w.WriteHeader(400)
				return
			}
			tokens[key]++
			fmt.Fprintf(w, `{"access_token":%q,"expires_in":3600,"token_type":"Bearer"}`, key)
		case "/api/v1/accounts":
			key := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if key != "fake-0" && key != "fake-1" {
				bad = append(bad, "unexpected bearer")
				w.WriteHeader(400)
				return
			}
			accounts[key]++
			if accounts[key] == 1 {
				w.WriteHeader(401)
				return
			}
			fmt.Fprint(w, `{"result":[{"accountNo":"fake","accountSeq":17,"accountType":"BROKERAGE"}]}`)
		default:
			bad = append(bad, "unexpected path")
			w.WriteHeader(404)
		}
	}))
	defer server.Close()
	if e = os.WriteFile(filepath.Join(base, "fixture-cert.pem"), pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw}), 0600); e != nil {
		t.Fatal(e)
	}
	children := make([]*exec.Cmd, 2)
	for i := range children {
		c := exec.Command(os.Args[0], "-test.run=^TestA166RuntimeWireChild$")
		c.Env = append(os.Environ(), "RUNTIME_WIRE_ROOT="+base, "RUNTIME_WIRE_ID="+strconv.Itoa(i), "RUNTIME_WIRE_MANIFEST="+manifestPath, "RUNTIME_WIRE_URL="+server.URL)
		if e = c.Start(); e != nil {
			t.Fatal(e)
		}
		children[i] = c
	}
	for _, c := range children {
		if e = c.Wait(); e != nil {
			t.Fatal(e)
		}
	}
	stamps := make([]int64, 0, 8)
	for i := 0; i < 2; i++ {
		v, e := os.ReadFile(filepath.Join(base, "first-write-"+strconv.Itoa(i)))
		if e != nil {
			t.Fatal(e)
		}
		lines := strings.Fields(string(v))
		if len(lines) != 4 {
			t.Fatalf("child %d seam count %d", i, len(lines))
		}
		for _, line := range lines {
			ns, e := strconv.ParseInt(line, 10, 64)
			if e != nil {
				t.Fatal(e)
			}
			stamps = append(stamps, ns)
		}
	}
	mu.Lock()
	defer mu.Unlock()
	if len(bad) != 0 || tokens["fake-0"] != 2 || tokens["fake-1"] != 2 || accounts["fake-0"] != 2 || accounts["fake-1"] != 2 {
		t.Fatalf("fixture paths tokens=%v accounts=%v bad=%v", tokens, accounts, bad)
	}
	sort.Slice(stamps, func(i, j int) bool { return stamps[i] < stamps[j] })
	gaps := make([]time.Duration, 0, len(stamps)-1)
	for i := 1; i < len(stamps); i++ {
		gap := time.Duration(stamps[i] - stamps[i-1])
		gaps = append(gaps, gap)
		if gap < 350*time.Millisecond {
			t.Fatalf("first Write gap %v", gap)
		}
	}
	t.Logf("first-write samples=%d adjacent=%v; per process token=2 accounts=2", len(stamps), gaps)
	for i := 3; i < len(stamps); i++ {
		if time.Duration(stamps[i]-stamps[i-3]) < time.Second {
			t.Fatal("four first Writes within rolling second")
		}
	}
}
