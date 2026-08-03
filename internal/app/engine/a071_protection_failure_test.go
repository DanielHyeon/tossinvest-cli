package engine_test

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProtectionStorageFailureCannotPreventSafetyRuntimeAssembly(t *testing.T) {
	dir := isolate(t)
	writeEngineConfig(t, dir)
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	// A path collision that made the former protection SQLite supervisor fail
	// startup. The dormant read-only assembly must not depend on this path.
	if err := os.Mkdir(filepath.Join(dir, "protection.db"), 0o700); err != nil {
		t.Fatal(err)
	}
	srv, _ := engineStub(t, "123-45")
	eng, err := startEngine(t, dir, srv)
	if err != nil {
		t.Fatalf("protection storage failure stopped exit/reconcile/fill runtime assembly: %v", err)
	}
	defer eng.Close()
	if eng.Journal == nil || eng.Entry == nil || eng.Reconcile == nil || eng.Gateway == nil {
		t.Fatalf("safety runtime is incomplete: journal=%v entry=%v reconcile=%v gateway=%v", eng.Journal, eng.Entry, eng.Reconcile, eng.Gateway)
	}
	if eng.Automation.EntryPermitted {
		t.Fatal("dormant protection storage unexpectedly opened entry")
	}
}
