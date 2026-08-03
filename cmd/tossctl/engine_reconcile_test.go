package main

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

type reconcileResolveTestLock struct {
	path     string
	released *bool
}

func (l *reconcileResolveTestLock) Path() string { return l.path }
func (l *reconcileResolveTestLock) Release()     { *l.released = true }

func executeReconcileResolve(t *testing.T, deps reconcileResolveDeps, args ...string) (string, error) {
	t.Helper()
	cmd := newEngineReconcileResolveCmdWithDeps(&rootOptions{}, deps)
	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(io.Discard)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func cleanReconcileResolveRuntime(snapshots []reconcile.Snapshot, resolved *bool) reconcileResolveRuntime {
	next := 0
	return reconcileResolveRuntime{
		accountRef: "acct-7",
		collect: func(context.Context) (reconcile.Snapshot, error) {
			if next >= len(snapshots) {
				return reconcile.Snapshot{}, errors.New("unexpected snapshot read")
			}
			snap := snapshots[next]
			next++
			return snap, nil
		},
		localState: func(context.Context, string) (reconcile.LocalState, error) {
			return reconcile.LocalState{AccountRef: "acct-7", Positions: map[string]string{}, OpenOrders: map[string]reconcile.LocalOrder{}}, nil
		},
		compare: func(reconcile.Snapshot, reconcile.LocalState) reconcile.Diff {
			return reconcile.Diff{AccountRef: "acct-7"}
		},
		validate: func() error { return nil },
		resolve: func(_ context.Context, operator, note string) error {
			if operator != "ops-1" || note != "fresh official snapshots agree" {
				return errors.New("identity or note was not forwarded")
			}
			*resolved = true
			return nil
		},
		close: func() error { return nil },
	}
}

func stableReconcileResolveSnapshots() []reconcile.Snapshot {
	first := time.Date(2026, 8, 3, 4, 0, 0, 0, time.UTC)
	return []reconcile.Snapshot{
		{AsOf: first, CompletedAt: first.Add(time.Millisecond), AccountRef: "acct-7"},
		{AsOf: first.Add(reconcile.DefaultStabilisationInterval), CompletedAt: first.Add(reconcile.DefaultStabilisationInterval + time.Millisecond), AccountRef: "acct-7"},
		{AsOf: first.Add(2 * reconcile.DefaultStabilisationInterval), CompletedAt: first.Add(2*reconcile.DefaultStabilisationInterval + time.Millisecond), AccountRef: "acct-7"},
	}
}

func TestEngineReconcileResolveIsRegisteredAsOfficialReadOnly(t *testing.T) {
	engineCmd := newEngineCmd(&rootOptions{})
	cmd, _, err := engineCmd.Find([]string{"reconcile-resolve"})
	if err != nil {
		t.Fatalf("Find(reconcile-resolve): %v", err)
	}
	if cmd == engineCmd || cmd.Name() != "reconcile-resolve" {
		t.Fatalf("engine child = %q, want reconcile-resolve", cmd.CommandPath())
	}
	if got := cmd.Annotations["source"]; got != "official" {
		t.Fatalf("source annotation = %q, want official", got)
	}
	if _, mutating := cmd.Annotations["mutating"]; mutating {
		t.Fatal("reconcile-resolve must not be annotated as a broker mutation")
	}
	for _, name := range []string{"confirm", "operator", "note"} {
		if cmd.Flag(name) == nil {
			t.Errorf("missing required --%s flag", name)
		}
	}
}

func TestEngineReconcileResolveProductionWiringHasNoBrokerMutationOrHTTPDoor(t *testing.T) {
	src, err := os.ReadFile("engine_reconcile.go")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	text := string(src)
	for _, required := range []string{
		"enginelock.Acquire", "reconcile.LocalStateFromJournal", "tracker.Resolve",
	} {
		if !strings.Contains(text, required) {
			t.Errorf("production wiring does not contain %q", required)
		}
	}
	for _, forbidden := range []string{
		"PlaceOrder", "PlacePendingOrder", "CancelOrder", "AmendPendingOrder",
		"ModifyOrder", "NewOrderPath", "httpapi", `"mutating": "true"`,
	} {
		if strings.Contains(text, forbidden) {
			t.Errorf("production recovery wiring contains forbidden mutation surface %q", forbidden)
		}
	}
}

func TestEngineReconcileResolveRequiresExplicitConfirmationIdentityAndNote(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{name: "confirmation", args: []string{"--confirm=false", "--operator", "ops-1", "--note", "checked"}, want: "--confirm"},
		{name: "operator", args: []string{"--confirm", "--operator", " ", "--note", "checked"}, want: "--operator"},
		{name: "note", args: []string{"--confirm", "--operator", "ops-1", "--note", " "}, want: "--note"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			_, err := executeReconcileResolve(t, reconcileResolveDeps{
				journalDir: func(*rootOptions) (string, error) { called = true; return "", nil },
			}, tc.args...)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want refusal naming %s", err, tc.want)
			}
			if called {
				t.Fatal("validation must refuse before resolving or locking the live journal directory")
			}
		})
	}
}

func TestEngineReconcileResolveAcquiresTheEngineLockBeforeAssembly(t *testing.T) {
	lockErr := errors.New("engine is already running")
	assembled := false
	_, err := executeReconcileResolve(t, reconcileResolveDeps{
		journalDir: func(*rootOptions) (string, error) { return "/journal", nil },
		acquire: func(dir string) (reconcileResolveLock, error) {
			if dir != "/journal" {
				t.Fatalf("lock dir = %q, want /journal", dir)
			}
			return nil, lockErr
		},
		build: func(context.Context, *rootOptions) (reconcileResolveRuntime, error) {
			assembled = true
			return reconcileResolveRuntime{}, nil
		},
		sleep: func(context.Context, time.Duration) error { return nil },
	}, "--confirm", "--operator", "ops-1", "--note", "checked")
	if !errors.Is(err, lockErr) {
		t.Fatalf("error = %v, want lock error", err)
	}
	if assembled {
		t.Fatal("official reads or journal assembly happened before engine exclusion")
	}
}

func TestEngineReconcileResolveRefusesUnstableSnapshots(t *testing.T) {
	snaps := stableReconcileResolveSnapshots()
	snaps[1].Holdings = []reconcile.Holding{{Symbol: "005930", Quantity: "1"}}
	resolved := false
	released := false
	closed := false
	slept := time.Duration(0)
	runtime := cleanReconcileResolveRuntime(snaps, &resolved)
	runtime.close = func() error { closed = true; return nil }

	_, err := executeReconcileResolve(t, reconcileResolveDeps{
		journalDir: func(*rootOptions) (string, error) { return "/journal", nil },
		acquire: func(string) (reconcileResolveLock, error) {
			return &reconcileResolveTestLock{path: "/journal/engine.lock", released: &released}, nil
		},
		build: func(context.Context, *rootOptions) (reconcileResolveRuntime, error) { return runtime, nil },
		sleep: func(_ context.Context, d time.Duration) error { slept = d; return nil },
	}, "--confirm", "--operator", "ops-1", "--note", "fresh official snapshots agree")
	if err == nil || !strings.Contains(err.Error(), "stable") {
		t.Fatalf("error = %v, want stable-snapshot refusal", err)
	}
	if resolved {
		t.Fatal("unstable account evidence reached operator resolution")
	}
	if slept < reconcile.DefaultStabilisationInterval {
		t.Fatalf("snapshot interval = %s, want at least %s", slept, reconcile.DefaultStabilisationInterval)
	}
	if !released || !closed {
		t.Fatalf("cleanup: lock released=%v runtime closed=%v", released, closed)
	}
}

func TestEngineReconcileResolveRefusesABlockingCorrectedComparison(t *testing.T) {
	resolved := false
	runtime := cleanReconcileResolveRuntime(stableReconcileResolveSnapshots(), &resolved)
	runtime.compare = func(reconcile.Snapshot, reconcile.LocalState) reconcile.Diff {
		return reconcile.Diff{AccountRef: "acct-7", Quantities: []reconcile.QuantityMismatch{{Symbol: "005930", Local: "1", Broker: "2"}}}
	}
	_, err := executeReconcileResolve(t, reconcileResolveDeps{
		journalDir: func(*rootOptions) (string, error) { return "/journal", nil },
		acquire: func(string) (reconcileResolveLock, error) {
			released := false
			return &reconcileResolveTestLock{path: "/journal/engine.lock", released: &released}, nil
		},
		build: func(context.Context, *rootOptions) (reconcileResolveRuntime, error) { return runtime, nil },
		sleep: func(context.Context, time.Duration) error { return nil },
	}, "--confirm", "--operator", "ops-1", "--note", "fresh official snapshots agree")
	if err == nil || !strings.Contains(err.Error(), "blocking") {
		t.Fatalf("error = %v, want blocking-diff refusal", err)
	}
	if resolved {
		t.Fatal("blocking comparison reached operator resolution")
	}
}

func TestEngineReconcileResolveRefusesAFinalSnapshotThatChanged(t *testing.T) {
	snaps := stableReconcileResolveSnapshots()
	snaps[2].Holdings = []reconcile.Holding{{Symbol: "005930", Quantity: "1"}}
	resolved := false
	runtime := cleanReconcileResolveRuntime(snaps, &resolved)
	_, err := executeReconcileResolve(t, reconcileResolveDeps{
		journalDir: func(*rootOptions) (string, error) { return "/journal", nil },
		acquire: func(string) (reconcileResolveLock, error) {
			released := false
			return &reconcileResolveTestLock{path: "/journal/engine.lock", released: &released}, nil
		},
		build: func(context.Context, *rootOptions) (reconcileResolveRuntime, error) { return runtime, nil },
		sleep: func(context.Context, time.Duration) error { return nil },
	}, "--confirm", "--operator", "ops-1", "--note", "fresh official snapshots agree")
	if err == nil || !strings.Contains(err.Error(), "final bounded-fresh") {
		t.Fatalf("error = %v, want final-snapshot refusal", err)
	}
	if resolved {
		t.Fatal("changed final account evidence reached operator resolution")
	}
}

func TestEngineReconcileResolveRefusesNonQuantityCauses(t *testing.T) {
	resolved := false
	runtime := cleanReconcileResolveRuntime(stableReconcileResolveSnapshots(), &resolved)
	runtime.validate = func() error {
		return errors.New("active cause IDENTIFIER_CONFLICT requires cause-specific recovery evidence")
	}
	_, err := executeReconcileResolve(t, reconcileResolveDeps{
		journalDir: func(*rootOptions) (string, error) { return "/journal", nil },
		acquire: func(string) (reconcileResolveLock, error) {
			released := false
			return &reconcileResolveTestLock{path: "/journal/engine.lock", released: &released}, nil
		},
		build: func(context.Context, *rootOptions) (reconcileResolveRuntime, error) { return runtime, nil },
		sleep: func(context.Context, time.Duration) error { return nil },
	}, "--confirm", "--operator", "ops-1", "--note", "fresh official snapshots agree")
	if err == nil || !strings.Contains(err.Error(), "IDENTIFIER_CONFLICT") {
		t.Fatalf("error = %v, want cause-specific recovery refusal", err)
	}
	if resolved {
		t.Fatal("quantity-only evidence released a foreign reconcile cause")
	}
}

func TestEngineReconcileResolveRecordsDurableReleaseBeforeSuccess(t *testing.T) {
	resolved := false
	released := false
	closed := false
	runtime := cleanReconcileResolveRuntime(stableReconcileResolveSnapshots(), &resolved)
	runtime.close = func() error { closed = true; return nil }
	output, err := executeReconcileResolve(t, reconcileResolveDeps{
		journalDir: func(*rootOptions) (string, error) { return "/journal", nil },
		acquire: func(string) (reconcileResolveLock, error) {
			return &reconcileResolveTestLock{path: "/journal/engine.lock", released: &released}, nil
		},
		build: func(context.Context, *rootOptions) (reconcileResolveRuntime, error) { return runtime, nil },
		sleep: func(context.Context, time.Duration) error { return nil },
	}, "--confirm", "--operator", "ops-1", "--note", "fresh official snapshots agree")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !resolved {
		t.Fatal("the existing operator-resolution contract was not called")
	}
	if !strings.Contains(output, "released") {
		t.Fatalf("output = %q, want durable release acknowledgement", output)
	}
	if !released || !closed {
		t.Fatalf("cleanup: lock released=%v runtime closed=%v", released, closed)
	}
}

func TestEngineReconcileResolveDoesNotReportSuccessWhenPersistenceFails(t *testing.T) {
	persistErr := errors.New("journal fsync failed")
	resolved := false
	runtime := cleanReconcileResolveRuntime(stableReconcileResolveSnapshots(), &resolved)
	runtime.resolve = func(context.Context, string, string) error { return persistErr }
	output, err := executeReconcileResolve(t, reconcileResolveDeps{
		journalDir: func(*rootOptions) (string, error) { return "/journal", nil },
		acquire: func(string) (reconcileResolveLock, error) {
			released := false
			return &reconcileResolveTestLock{path: "/journal/engine.lock", released: &released}, nil
		},
		build: func(context.Context, *rootOptions) (reconcileResolveRuntime, error) { return runtime, nil },
		sleep: func(context.Context, time.Duration) error { return nil },
	}, "--confirm", "--operator", "ops-1", "--note", "fresh official snapshots agree")
	if !errors.Is(err, persistErr) {
		t.Fatalf("error = %v, want persistence error", err)
	}
	if strings.Contains(output, "released") {
		t.Fatalf("output = %q, must not claim release before durable persistence", output)
	}
}
