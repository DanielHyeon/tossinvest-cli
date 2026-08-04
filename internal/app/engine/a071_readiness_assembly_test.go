package engine_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProductionEngineAssemblesPairedUnwiredReadinessProvider(t *testing.T) {
	dir := isolate(t)
	writeEngineConfig(t, dir)
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	srv, _ := engineStub(t, "123-45")
	eng, err := startEngine(t, dir, srv)
	if err != nil {
		t.Fatal(err)
	}
	if !eng.Gateway.HasProtectionReadinessProvider() {
		t.Fatal("production gateway omitted the sealed readiness provider")
	}
	if eng.Automation.EntryPermitted {
		t.Fatal("default paired-UNWIRED readiness opened entry")
	}
}

func TestProtectionReadinessIsAbsentFromExitReconcileAndFillLoopSources(t *testing.T) {
	for _, name := range []string{"exitloop.go", "exitwiring.go", "reconcileloop.go", "runtime_wiring.go"} {
		data, err := os.ReadFile(filepath.Join(".", name))
		if err != nil {
			t.Fatal(err)
		}
		text := string(data)
		for _, forbidden := range []string{"ProtectionReadiness", "ReadinessAdapter", "checkProtection"} {
			if strings.Contains(text, forbidden) {
				t.Fatalf("%s routes stop/emergency/reconcile/fill work through readiness token %q", name, forbidden)
			}
		}
	}
}
