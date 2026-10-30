//go:build linux

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget"
	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudgetruntime"
	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/monitor"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/session"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

func TestA166RootBudgetEnvMatrix(t *testing.T) {
	cases := []struct {
		mode        string
		modeSet     bool
		manifest    string
		manifestSet bool
		valid       bool
	}{
		{"", false, "", false, true}, {"off", true, "", false, true}, {"required", true, "/x", true, true},
		{"", true, "", false, false}, {"wat", true, "", false, false}, {"off", true, "/x", true, false}, {"off", true, "", true, false}, {"", false, "", true, false}, {"required", true, "", false, false}, {"required", true, "", true, false}, {"", false, "/x", true, false},
	}
	for _, v := range cases {
		_, e := parseBudgetEnv(v.mode, v.modeSet, v.manifest, v.manifestSet)
		if (e == nil) != v.valid {
			t.Fatalf("case %+v: %v", v, e)
		}
	}
}

func a166RootManifest(t *testing.T) string {
	t.Helper()
	base := t.TempDir()
	root := filepath.Join(base, "group")
	if e := os.Mkdir(root, 0700); e != nil {
		t.Fatal(e)
	}
	if e := accountbudget.Provision(root, "scope", "instance"); e != nil {
		t.Fatal(e)
	}
	eventID := func() accountbudget.FileID {
		b, err := accountbudget.Open(accountbudget.Config{Root: root, Scope: "scope", Instance: "instance", Clock: clock.System()})
		if err != nil {
			t.Fatal(err)
		}
		id, err := accountbudget.ProvisionObservation(b)
		err = errors.Join(err, b.Close())
		if err != nil {
			t.Fatal(err)
		}
		return id
	}()
	id := func(path string) accountbudget.FileID {
		st, e := os.Stat(path)
		if e != nil {
			t.Fatal(e)
		}
		v := st.Sys().(*syscall.Stat_t)
		return accountbudget.FileID{Device: uint64(v.Dev), Inode: v.Ino}
	}
	m := map[string]any{"event_lock": eventID, "version": 2, "root": root, "scope": "scope", "instance": "instance", "identity": accountbudget.Identity{Root: id(root), Queue: id(filepath.Join(root, "queue.lock")), Wire: id(filepath.Join(root, "wire.lock"))}}
	data, e := json.Marshal(m)
	if e != nil {
		t.Fatal(e)
	}
	path := filepath.Join(base, "launch.json")
	if e = os.WriteFile(path, data, 0600); e != nil {
		t.Fatal(e)
	}
	return path
}
func a166Owner(path string) *rootBudgetOwner {
	return newRootBudgetOwner(func(k string) (string, bool) {
		if k == "TOSSOS_ACCOUNT_BUDGET_MODE" {
			return "required", true
		}
		return path, true
	}, io.Discard)
}
func TestA166RootFactoryConcurrentReuseAndClose(t *testing.T) {
	o := a166Owner(a166RootManifest(t))
	var wg sync.WaitGroup
	results := make(chan *accountbudgetruntime.Factory, 12)
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			f, e := o.get(context.Background())
			if e != nil {
				t.Error(e)
				return
			}
			results <- f
		}()
	}
	wg.Wait()
	close(results)
	var first *accountbudgetruntime.Factory
	for f := range results {
		if first == nil {
			first = f
		} else if first != f {
			t.Fatal("duplicate owner")
		}
	}
	b, e := first.Budget()
	if e != nil {
		t.Fatal(e)
	}
	lease, e := b.Acquire(context.Background(), accountbudget.Critical)
	if e != nil {
		t.Fatal(e)
	}
	if e = o.Close(); e == nil {
		t.Fatal("active close accepted")
	}
	if _, e = o.get(context.Background()); e == nil {
		t.Fatal("closing owner reopened")
	}
	if e = lease.Release(); e != nil {
		t.Fatal(e)
	}
	if e = o.Close(); e != nil {
		t.Fatal(e)
	}
}
func TestA166RootFailureCanRetry(t *testing.T) {
	path := a166RootManifest(t)
	backup := path + ".saved"
	if e := os.Rename(path, backup); e != nil {
		t.Fatal(e)
	}
	o := a166Owner(path)
	if _, e := o.get(context.Background()); e == nil {
		t.Fatal("missing accepted")
	}
	if e := os.Rename(backup, path); e != nil {
		t.Fatal(e)
	}
	if _, e := o.get(context.Background()); e != nil {
		t.Fatal(e)
	}
	if e := o.Close(); e != nil {
		t.Fatal(e)
	}
}
func TestA166ExecuteClosesOnSuccessAndFailure(t *testing.T) {
	for _, fail := range []bool{false, true} {
		o := a166Owner(a166RootManifest(t))
		if _, e := o.get(context.Background()); e != nil {
			t.Fatal(e)
		}
		sentinel := errors.New("command failure")
		root := &cobra.Command{Use: "test", SilenceErrors: true, RunE: func(*cobra.Command, []string) error {
			if fail {
				return sentinel
			}
			return nil
		}}
		e := executeRoot(root, &rootOptions{budgetOwner: o})
		if fail && !errors.Is(e, sentinel) {
			t.Fatalf("lost error %v", e)
		}
		if !fail && e != nil {
			t.Fatal(e)
		}
		if _, e = o.get(context.Background()); e == nil {
			t.Fatal("owner not closed")
		}
	}
}

func TestA166RootWaitHonorsContext(t *testing.T) {
	o := a166Owner(a166RootManifest(t))
	<-o.gate
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e := o.get(ctx); !errors.Is(e, context.Canceled) {
		t.Fatalf("want canceled: %v", e)
	}
	o.gate <- struct{}{}
	if e := o.Close(); e != nil {
		t.Fatal(e)
	}
}
func TestA166ExecutePreservesCommandAndCloseErrors(t *testing.T) {
	o := a166Owner(a166RootManifest(t))
	f, e := o.get(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	b, e := f.Budget()
	if e != nil {
		t.Fatal(e)
	}
	lease, e := b.Acquire(context.Background(), accountbudget.Critical)
	if e != nil {
		t.Fatal(e)
	}
	cause := errors.New("command error")
	root := &cobra.Command{Use: "test", SilenceErrors: true, SilenceUsage: true, RunE: func(*cobra.Command, []string) error { return cause }}
	e = executeRoot(root, &rootOptions{budgetOwner: o})
	if !errors.Is(e, cause) || !strings.Contains(e.Error(), "active owners") {
		t.Fatalf("lost combined error: %v", e)
	}
	if e = lease.Release(); e != nil {
		t.Fatal(e)
	}
	if e = o.Close(); e != nil {
		t.Fatal(e)
	}
}
func TestA166MonitorSentinelIsSilent(t *testing.T) {
	var out bytes.Buffer
	root := &cobra.Command{Use: "test", SilenceUsage: true, RunE: func(c *cobra.Command, _ []string) error { c.SilenceErrors = true; return errMonitorFailed }}
	root.SetErr(&out)
	if e := executeRoot(root, &rootOptions{}); e != errMonitorFailed {
		t.Fatalf("wrong result %v", e)
	}
	if out.Len() != 0 {
		t.Fatalf("unexpected output %q", out.String())
	}
}

func TestA166RootCachedFactoryRejectsCanceledContext(t *testing.T) {
	o := a166Owner(a166RootManifest(t))
	if _, e := o.get(context.Background()); e != nil {
		t.Fatal(e)
	}
	defer o.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for i := 0; i < 100; i++ {
		if f, e := o.get(ctx); f != nil || !errors.Is(e, context.Canceled) {
			t.Fatalf("canceled cache accepted: %v", e)
		}
	}
}

func TestA166EngineAndFlattenRequiredMissingFailBeforeBootstrap(t *testing.T) {
	o := a166Owner(filepath.Join(t.TempDir(), "missing.json"))
	root := &rootOptions{budgetOwner: o, configDir: t.TempDir()}
	if c, e := assembleEngine(context.Background(), root, clock.System(), nil, nil); c != nil || !errors.Is(e, accountbudget.ErrUnavailable) {
		t.Fatalf("engine accepted missing manifest: %v", e)
	}
	if w, e := buildFlattenWiring(context.Background(), root, &flattenOptions{}); w != nil || !errors.Is(e, accountbudget.ErrUnavailable) {
		t.Fatalf("flatten accepted missing manifest: %v", e)
	}
	if e := o.Close(); e != nil {
		t.Fatal(e)
	}
}
func TestA166RootEnvSnapshot(t *testing.T) {
	env := map[string]string{"TOSSOS_ACCOUNT_BUDGET_MODE": "off"}
	o := newRootBudgetOwner(func(k string) (string, bool) { v, ok := env[k]; return v, ok }, io.Discard)
	env["TOSSOS_ACCOUNT_BUDGET_MODE"] = "invalid"
	f, e := o.get(context.Background())
	if e != nil || f == nil {
		t.Fatalf("env not snapshotted: %v", e)
	}
	if e = o.Close(); e != nil {
		t.Fatal(e)
	}
}

func TestA166FlattenBootstrapUsesBorrowedBudget(t *testing.T) {
	t.Setenv("TOSSCTL_OPENAPI_KEY", "fake-key")
	t.Setenv("TOSSCTL_OPENAPI_SECRET", "fake-secret")
	var calls atomic.Int32
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1); http.Error(w, "unexpected", 500) }))
	defer s.Close()
	tr := s.Client().Transport.(*http.Transport).Clone()
	tr.Proxy = nil
	o := a166Owner(a166RootManifest(t))
	defer o.Close()
	f, e := o.get(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	b, e := f.Budget()
	if e != nil {
		t.Fatal(e)
	}
	lease, e := b.Acquire(context.Background(), accountbudget.Critical)
	if e != nil {
		t.Fatal(e)
	}
	defer lease.Release()
	root := &rootOptions{budgetOwner: o, configDir: t.TempDir(), officialOptions: []official.Option{official.WithBaseURL(s.URL), official.WithHTTPClient(&http.Client{Transport: tr})}}
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()
	if _, e = buildFlattenWiring(ctx, root, &flattenOptions{}); e == nil || ctx.Err() == nil {
		t.Fatalf("flatten bootstrap bypass: %v", e)
	}
	if calls.Load() != 0 {
		t.Fatal("bootstrap hit fake server")
	}
}

func TestA166MonitorFailureOutputAndNoBudgetOpen(t *testing.T) {
	dir := t.TempDir()
	if e := session.WriteFile(filepath.Join(dir, "session.json"), &session.Session{Cookies: map[string]string{"SESSION": "fake"}}); e != nil {
		t.Fatal(e)
	}
	old := runMonitor
	runMonitor = func(context.Context, *session.Session) []monitor.Result {
		return []monitor.Result{{Probe: monitor.Probe{Name: "fixture"}, Status: 503, Detail: "unavailable"}}
	}
	defer func() { runMonitor = old }()
	owner := a166Owner(filepath.Join(dir, "missing-manifest"))
	opts := &rootOptions{budgetOwner: owner}
	root := newRootCmdWithOptions(opts)
	var out, errout bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errout)
	root.SetArgs([]string{"--config-dir", dir, "monitor", "api"})
	if e := executeRoot(root, opts); e != errMonitorFailed {
		t.Fatalf("wrong exit: %v", e)
	}
	if !strings.Contains(out.String(), "0 passed, 1 failed") || !strings.Contains(out.String()+errout.String(), "fixture") || strings.Contains(out.String()+errout.String(), "monitor api failed") {
		t.Fatalf("output: %q %q", out.String(), errout.String())
	}
	if owner.factory != nil || !owner.closing {
		t.Fatal("monitor opened budget or failed to close")
	}
}

func TestA166MonitorMainExitOne(t *testing.T) {
	if os.Getenv("A166_MONITOR_CHILD") == "1" {
		runMonitor = func(context.Context, *session.Session) []monitor.Result {
			return []monitor.Result{{Probe: monitor.Probe{Name: "fixture"}, Status: 503, Detail: "unavailable"}}
		}
		os.Args = []string{"tossctl", "--config-dir", os.Getenv("A166_MONITOR_DIR"), "monitor", "api"}
		main()
		os.Exit(0)
	}
	dir := t.TempDir()
	if e := session.WriteFile(filepath.Join(dir, "session.json"), &session.Session{Cookies: map[string]string{"SESSION": "fake"}}); e != nil {
		t.Fatal(e)
	}
	c := exec.Command(os.Args[0], "-test.run=^TestA166MonitorMainExitOne$")
	c.Env = append(os.Environ(), "A166_MONITOR_CHILD=1", "A166_MONITOR_DIR="+dir, "TOSSOS_ACCOUNT_BUDGET_MODE=required", "TOSSOS_ACCOUNT_BUDGET_MANIFEST="+filepath.Join(dir, "missing"))
	out, e := c.CombinedOutput()
	var exit *exec.ExitError
	if !errors.As(e, &exit) || exit.ExitCode() != 1 {
		t.Fatalf("exit=%v output=%s", e, out)
	}
	if !strings.Contains(string(out), "fixture") || !strings.Contains(string(out), "0 passed, 1 failed") || strings.Contains(string(out), "monitor api failed") {
		t.Fatalf("output=%q", out)
	}
}

func TestA166EngineDecoratorCannotRemoveRequiredBudget(t *testing.T) {
	t.Setenv("TOSSCTL_OPENAPI_KEY", "fake-key")
	t.Setenv("TOSSCTL_OPENAPI_SECRET", "fake-secret")
	var calls atomic.Int32
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1); http.Error(w, "unexpected", 500) }))
	defer s.Close()
	tr := s.Client().Transport.(*http.Transport).Clone()
	tr.Proxy = nil
	o := a166Owner(a166RootManifest(t))
	defer o.Close()
	f, e := o.get(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	b, e := f.Budget()
	if e != nil {
		t.Fatal(e)
	}
	lease, e := b.Acquire(context.Background(), accountbudget.Critical)
	if e != nil {
		t.Fatal(e)
	}
	defer lease.Release()
	root := &rootOptions{budgetOwner: o, configDir: t.TempDir()}
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()
	_, e = assembleEngine(ctx, root, clock.System(), nil, func(o *engine.Options) { o.BudgetFactory = nil }, official.WithBaseURL(s.URL), official.WithHTTPClient(&http.Client{Transport: tr}))
	if e == nil || ctx.Err() == nil {
		t.Fatalf("engine bootstrap bypass: %v", e)
	}
	if calls.Load() != 0 {
		t.Fatal("bootstrap hit fake server")
	}
}

func TestA166DirectAppBuilderMissingContextFailsClosed(t *testing.T) {
	t.Setenv("TOSSCTL_OPENAPI_KEY", "fake-key")
	t.Setenv("TOSSCTL_OPENAPI_SECRET", "fake-secret")
	dir := t.TempDir()
	cfg := config.DefaultFile()
	cfg.OpenAPI = config.OpenAPI{Enabled: true, Prefer: "auto"}
	data, e := json.Marshal(cfg)
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(dir, "config.json"), data, 0600); e != nil {
		t.Fatal(e)
	}
	o := a166Owner(a166RootManifest(t))
	defer o.Close()
	app, e := newAppContext(&rootOptions{budgetOwner: o, configDir: dir, outputFormat: "table"})
	if app != nil || !errors.Is(e, accountbudget.ErrUnavailable) {
		t.Fatalf("missing context accepted: %v", e)
	}
}
func TestA166CanceledRequiredStartupNoWire(t *testing.T) {
	t.Setenv("TOSSCTL_OPENAPI_KEY", "fake-key")
	t.Setenv("TOSSCTL_OPENAPI_SECRET", "fake-secret")
	var calls atomic.Int32
	s := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { calls.Add(1) }))
	defer s.Close()
	o := a166Owner(a166RootManifest(t))
	defer o.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	root := &rootOptions{budgetOwner: o, configDir: t.TempDir()}
	c, e := assembleEngine(ctx, root, clock.System(), nil, nil, official.WithBaseURL(s.URL), official.WithHTTPClient(s.Client()))
	if c != nil || !errors.Is(e, context.Canceled) || calls.Load() != 0 {
		t.Fatalf("startup cancellation err=%v calls=%d", e, calls.Load())
	}
}

func TestA166MonitorAndCloseErrorPreserveBothCauses(t *testing.T) {
	closeErr := errors.New("close fixture")
	joined := monitorCloseError{cause: closeErr}
	if !errors.Is(joined, errMonitorFailed) || !errors.Is(joined, closeErr) {
		t.Fatal("lost cause")
	}
	if strings.Contains(joined.Error(), errMonitorFailed.Error()) || joined.Error() != closeErr.Error() {
		t.Fatalf("unexpected display: %q", joined.Error())
	}

	o := a166Owner(a166RootManifest(t))
	f, e := o.get(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	b, e := f.Budget()
	if e != nil {
		t.Fatal(e)
	}
	lease, e := b.Acquire(context.Background(), accountbudget.Critical)
	if e != nil {
		t.Fatal(e)
	}
	root := &cobra.Command{Use: "test", SilenceErrors: true, SilenceUsage: true, RunE: func(*cobra.Command, []string) error { return errMonitorFailed }}
	e = executeRoot(root, &rootOptions{budgetOwner: o})
	if !errors.Is(e, errMonitorFailed) || !strings.Contains(e.Error(), "active owners") || strings.Contains(e.Error(), errMonitorFailed.Error()) {
		t.Fatalf("wrong combined monitor result: %v", e)
	}
	if e = lease.Release(); e != nil {
		t.Fatal(e)
	}
	if e = o.Close(); e != nil {
		t.Fatal(e)
	}
}
