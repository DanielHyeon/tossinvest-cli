//go:build linux

package accountbudgetruntime

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

func TestRequiredMissingAndOffContradiction(t *testing.T) {
	for _, cfg := range []Config{{}, {Mode: ModeRequired}, {Mode: ModeOff, Manifest: "/unexpected"}, {Mode: Mode(99)}} {
		f, e := Open(context.Background(), cfg)
		if f != nil || !errors.Is(e, accountbudget.ErrUnavailable) {
			t.Fatalf("accepted invalid mode %v", cfg.Mode)
		}
	}
}

func setupGroup(t *testing.T) (string, string, manifest) {
	t.Helper()
	base := t.TempDir()
	root := filepath.Join(base, "group")
	if e := os.Mkdir(root, 0700); e != nil {
		t.Fatal(e)
	}
	if e := accountbudget.Provision(root, "scope", "instance"); e != nil {
		t.Fatal(e)
	}
	id := func(p string) accountbudget.FileID {
		st, e := os.Stat(p)
		if e != nil {
			t.Fatal(e)
		}
		v := st.Sys().(*syscall.Stat_t)
		return accountbudget.FileID{Device: uint64(v.Dev), Inode: v.Ino}
	}
	m := manifest{Version: 1, Root: root, Scope: "scope", Instance: "instance", Identity: accountbudget.Identity{Root: id(root), Queue: id(filepath.Join(root, "queue.lock")), Wire: id(filepath.Join(root, "wire.lock"))}}
	path := filepath.Join(base, "launch.json")
	writeManifest(t, path, m)
	return root, path, m
}
func writeManifest(t *testing.T, path string, m manifest) {
	t.Helper()
	b, e := json.Marshal(m)
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(path, b, 0600); e != nil {
		t.Fatal(e)
	}
}
func TestFactoryReuseAndIdentity(t *testing.T) {
	_, path, _ := setupGroup(t)
	f, e := Open(context.Background(), Config{Mode: ModeRequired, Manifest: path, CleanupError: func(error) {}})
	if e != nil {
		t.Fatal(e)
	}
	borrowed, e := f.Budget()
	if e != nil || borrowed == nil {
		t.Fatal("missing borrowed budget")
	}
	a, e := f.NewClient(official.Credentials{}, "")
	if e != nil {
		t.Fatal(e)
	}
	b, e := f.NewClient(official.Credentials{}, "")
	if e != nil {
		t.Fatal(e)
	}
	if !a.UsesAccountBudget(borrowed) || !b.UsesAccountBudget(borrowed) {
		t.Fatal("not shared")
	}
	if e = f.Close(); e != nil {
		t.Fatal(e)
	}
	if borrowed, e := f.Budget(); borrowed != nil || e == nil {
		t.Fatal("closed accessor accepted")
	}
	if c, e := f.NewClient(official.Credentials{}, ""); c != nil || e == nil {
		t.Fatal("closed factory accepted")
	}
}
func TestIdentityMismatchFails(t *testing.T) {
	_, path, m := setupGroup(t)
	for _, change := range []func(*manifest){func(m *manifest) { m.Identity.Root.Inode++ }, func(m *manifest) { m.Identity.Queue.Inode++ }, func(m *manifest) { m.Identity.Wire.Inode++ }} {
		bad := m
		change(&bad)
		writeManifest(t, path, bad)
		f, e := Open(context.Background(), Config{Mode: ModeRequired, Manifest: path, CleanupError: func(error) {}})
		if f != nil || !errors.Is(e, accountbudget.ErrUnavailable) {
			t.Fatalf("accepted mismatch: %v", e)
		}
	}
}
func TestManifestInvalid(t *testing.T) {
	_, path, m := setupGroup(t)
	valid, _ := json.Marshal(m)
	cases := [][]byte{[]byte(`{"version":1,"version":1}`), append(valid, []byte(` {}`)...), []byte(`{"version":1,"unknown":true}`), []byte(`{"Version":1}`), bytes.Repeat([]byte("x"), 4097)}
	for _, data := range cases {
		if e := os.WriteFile(path, data, 0600); e != nil {
			t.Fatal(e)
		}
		if f, e := Open(context.Background(), Config{Mode: ModeRequired, Manifest: path, CleanupError: func(error) {}}); f != nil || e == nil {
			t.Fatal("invalid accepted")
		}
	}
}
func TestEnvMode(t *testing.T) {
	lookup := func(string) (string, bool) { return "", false }
	c, e := FromEnv(lookup, "BUDGET", nil)
	if e != nil || c.Mode != ModeOff {
		t.Fatal("missing not OFF")
	}
	_, e = FromEnv(func(string) (string, bool) { return "", true }, "BUDGET", nil)
	if e == nil {
		t.Fatal("empty accepted")
	}
}
func TestChildOpen(t *testing.T) {
	if os.Getenv("RUNTIME_CHILD") == "" {
		return
	}
	f, e := Open(context.Background(), Config{Mode: ModeRequired, Manifest: os.Getenv("RUNTIME_MANIFEST"), CleanupError: func(error) {}})
	if os.Getenv("RUNTIME_CHILD") == "valid" {
		if e != nil {
			t.Fatal(e)
		}
		f.Close()
	} else if f != nil || e == nil {
		t.Fatal("copy accepted")
	}
}
func TestSameManifestAcrossProcessesRejectsCopiedRoot(t *testing.T) {
	root, path, _ := setupGroup(t)
	run := func(kind string) {
		c := exec.Command(os.Args[0], "-test.run=^TestChildOpen$")
		c.Env = append(os.Environ(), "RUNTIME_CHILD="+kind, "RUNTIME_MANIFEST="+path)
		if out, e := c.CombinedOutput(); e != nil {
			t.Fatalf("child %s: %v %s", kind, e, out)
		}
	}
	run("valid")
	copyRoot := root + "-copy"
	if e := os.Mkdir(copyRoot, 0700); e != nil {
		t.Fatal(e)
	}
	if e := accountbudget.Provision(copyRoot, "scope", "instance"); e != nil {
		t.Fatal(e)
	}
	if e := os.Rename(root, root+"-original"); e != nil {
		t.Fatal(e)
	}
	if e := os.Rename(copyRoot, root); e != nil {
		t.Fatal(e)
	}
	run("copy")
}

func TestManifestSymlinkModeAndInsideRoot(t *testing.T) {
	root, path, m := setupGroup(t)
	link := path + "-link"
	if e := os.Symlink(path, link); e != nil {
		t.Fatal(e)
	}
	cfg := Config{Mode: ModeRequired, Manifest: link, CleanupError: func(error) {}}
	if f, e := Open(context.Background(), cfg); f != nil || e == nil {
		t.Fatal("symlink accepted")
	}
	if e := os.Chmod(path, 0644); e != nil {
		t.Fatal(e)
	}
	cfg.Manifest = path
	if f, e := Open(context.Background(), cfg); f != nil || e == nil {
		t.Fatal("mode accepted")
	}
	inside := filepath.Join(root, "launch.json")
	writeManifest(t, inside, m)
	cfg.Manifest = inside
	if f, e := Open(context.Background(), cfg); f != nil || e == nil {
		t.Fatal("self approval accepted")
	}
}
func TestFactoryCloseWithActiveOwner(t *testing.T) {
	_, path, _ := setupGroup(t)
	f, e := Open(context.Background(), Config{Mode: ModeRequired, Manifest: path, CleanupError: func(error) {}})
	if e != nil {
		t.Fatal(e)
	}
	lease, e := f.budget.Acquire(context.Background(), accountbudget.Critical)
	if e != nil {
		t.Fatal(e)
	}
	if e = f.Close(); e == nil {
		t.Fatal("closed active budget")
	}
	if e = lease.Release(); e != nil {
		t.Fatal(e)
	}
	if e = f.Close(); e != nil {
		t.Fatal(e)
	}
}
func TestTwoRealClientsShareGroup(t *testing.T) {
	_, path, _ := setupGroup(t)
	f, e := Open(context.Background(), Config{Mode: ModeRequired, Manifest: path, CleanupError: func(error) {}})
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	var calls atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.URL.Path == "/oauth2/token" {
			if e := r.ParseForm(); e != nil {
				t.Error(e)
				return
			}
			token := r.Form.Get("client_id")
			if token != "fake-a" && token != "fake-b" {
				t.Error("unexpected fake key")
				return
			}
			json.NewEncoder(w).Encode(map[string]any{"access_token": token, "expires_in": 3600, "token_type": "Bearer"})
			return
		}
		if r.URL.Path == "/api/v1/accounts" {
			seq := 17
			if r.Header.Get("Authorization") == "Bearer fake-b" {
				seq = 29
			}
			json.NewEncoder(w).Encode(map[string]any{"result": []map[string]any{{"accountNo": "fake", "accountSeq": seq, "accountType": "BROKERAGE"}}})
			return
		}
		t.Errorf("unexpected path %s", r.URL.Path)
	}))
	defer server.Close()
	tr := server.Client().Transport.(*http.Transport).Clone()
	tr.Proxy = nil
	a, e := f.NewClient(official.Credentials{APIKey: "fake-a", SecretKey: "fake-a"}, "", official.WithBaseURL(server.URL), official.WithHTTPClient(&http.Client{Transport: tr}))
	if e != nil {
		t.Fatal(e)
	}
	b, e := f.NewClient(official.Credentials{APIKey: "fake-b", SecretKey: "fake-b"}, "", official.WithBaseURL(server.URL), official.WithHTTPClient(&http.Client{Transport: tr}))
	if e != nil {
		t.Fatal(e)
	}
	lease, e := f.budget.Acquire(context.Background(), accountbudget.Critical)
	if e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Millisecond)
	defer cancel()
	if _, e = a.Accounts(ctx); e == nil {
		t.Fatal("bootstrap bypassed budget")
	}
	if calls.Load() != 0 {
		t.Fatal("bootstrap reached network")
	}
	if e = lease.Release(); e != nil {
		t.Fatal(e)
	}
	for i, c := range []*official.Client{a, b} {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		accounts, e := c.Accounts(ctx)
		cancel()
		if e != nil || len(accounts) != 1 {
			t.Fatalf("accounts: %v", e)
		}
		seq, ok := c.SelectedAccountSeq()
		if !ok || seq != []int{17, 29}[i] {
			t.Fatalf("unexpected fake sequence %d", seq)
		}
	}
	if calls.Load() != 4 {
		t.Fatalf("token/accounts count %d", calls.Load())
	}
}

func TestContextAndRequiredCleanup(t *testing.T) {
	_, path, _ := setupGroup(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if f, e := Open(ctx, Config{Mode: ModeRequired, Manifest: path, CleanupError: func(error) {}}); f != nil || !errors.Is(e, context.Canceled) {
		t.Fatal("canceled context accepted")
	}
	if f, e := Open(context.Background(), Config{Mode: ModeRequired, Manifest: path}); f != nil || !errors.Is(e, accountbudget.ErrUnavailable) {
		t.Fatal("missing cleanup accepted")
	}
}
func TestParentSymlinkRefused(t *testing.T) {
	_, path, _ := setupGroup(t)
	parent := filepath.Dir(path)
	link := filepath.Join(t.TempDir(), "alias")
	if e := os.Symlink(parent, link); e != nil {
		t.Fatal(e)
	}
	if f, e := Open(context.Background(), Config{Mode: ModeRequired, Manifest: filepath.Join(link, "launch.json"), CleanupError: func(error) {}}); f != nil || e == nil {
		t.Fatal("parent symlink accepted")
	}
}

func TestOffFactoryDoesNotSupplyBudget(t *testing.T) {
	f, e := Open(context.Background(), Config{Mode: ModeOff})
	if e != nil {
		t.Fatal(e)
	}
	b, e := f.Budget()
	if e != nil || b != nil {
		t.Fatal("OFF gave budget")
	}
	c, e := f.NewClient(official.Credentials{}, "")
	if e != nil || c == nil {
		t.Fatal("OFF client unavailable")
	}
	if e = f.Close(); e != nil {
		t.Fatal(e)
	}
	if _, e = f.Budget(); e == nil {
		t.Fatal("closed OFF accessor")
	}
}

func TestFactoryCloseConcurrentCreation(t *testing.T) {
	_, path, _ := setupGroup(t)
	f, e := Open(context.Background(), Config{Mode: ModeRequired, Manifest: path, CleanupError: func(error) {}})
	if e != nil {
		t.Fatal(e)
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 30; j++ {
				f.Budget()
				f.NewClient(official.Credentials{}, "")
			}
		}()
	}
	if e = f.Close(); e != nil {
		t.Fatal(e)
	}
	wg.Wait()
	if c, e := f.NewClient(official.Credentials{}, ""); c != nil || e == nil {
		t.Fatal("closed factory accepted")
	}
}

func TestIncompleteFactoryFailsClosed(t *testing.T) {
	for _, f := range []*Factory{{}, {mode: ModeRequired}, {mode: ModeOff, budget: &accountbudget.Budget{}}} {
		if c, e := f.NewClient(official.Credentials{}, ""); c != nil || e == nil {
			t.Fatal("incomplete client accepted")
		}
		if b, e := f.Budget(); b != nil || e == nil {
			t.Fatal("incomplete budget accepted")
		}
	}
}
