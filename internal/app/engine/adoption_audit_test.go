package engine_test

// adoption_audit_test.go is §0.5/§0.7 for the adoption toggle (change
// adopt-external-positions task 2.4).
//
// Turning adoption on is the moment the engine starts placing sell orders for
// shares its owner bought by hand. That is exactly the class of change §0.5
// requires to be traceable and §0.7 requires a person to make, and the audit
// trail is the only artefact that survives a restart to say when it happened.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/audit"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
)

// writeAdoptionConfig writes a gate-off config carrying an adoption block.
func writeAdoptionConfig(t *testing.T, dir string, adoption config.Adoption) {
	t.Helper()
	cfg := config.DefaultFile()
	cfg.Trading = openTradingPolicy()
	cfg.Engine.Adoption = adoption
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), data, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
}

// TestAdoptionToggleIsAudited: the flip carries the previous value, the new
// value, the operator and the time.
func TestAdoptionToggleIsAudited(t *testing.T) {
	dir := isolate(t)
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	srv, _ := interlockServer(t, "123-45")

	// A first start with adoption off, so the later change has a baseline.
	writeAdoptionConfig(t, dir, config.Adoption{})
	if _, err := openGateEngine(t, dir, srv, matchedGuardian()); err != nil {
		t.Fatalf("first start: %v", err)
	}

	writeAdoptionConfig(t, dir, config.Adoption{
		Enabled: true, DefaultStopPct: 0.05, ExcludeSymbols: []string{"AAPL"},
	})
	if _, err := openGateEngine(t, dir, srv, matchedGuardian()); err != nil {
		t.Fatalf("second start: %v", err)
	}

	entries := readAudit(t, filepath.Join(dir, "audit.log"))
	toggle := lastEntryFor(entries, "engine.adoption.enabled", audit.ActionAdoptionToggle)
	if toggle == nil {
		t.Fatalf("no adoption toggle recorded; entries = %+v", entries)
	}
	if toggle.Old != "false" || toggle.New != "true" {
		t.Errorf("toggle old/new = %q/%q, want false/true", toggle.Old, toggle.New)
	}
	if toggle.At.IsZero() {
		t.Error("the toggle entry has no timestamp")
	}
	if toggle.Subject != "test-operator" {
		t.Errorf("subject = %q, want the operator identity", toggle.Subject)
	}

	pct := lastEntryFor(entries, "engine.adoption.default_stop_pct", audit.ActionAdoptionSetting)
	if pct == nil || pct.New != "0.05" {
		t.Errorf("stop fraction entry = %+v, want a change to 0.05", pct)
	}
	excludes := lastEntryFor(entries, "engine.adoption.exclude_symbols", audit.ActionAdoptionSetting)
	if excludes == nil || excludes.New != "AAPL" {
		t.Errorf("exclusion list entry = %+v, want a change to AAPL", excludes)
	}
}

// TestARefusedAdoptionBlockIsAudited: an operator who wrote a half-percent stop
// and found adoption off must be able to learn from the trail that the number
// was refused, not that their edit was lost.
func TestARefusedAdoptionBlockIsAudited(t *testing.T) {
	dir := isolate(t)
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	srv, _ := interlockServer(t, "123-45")

	writeAdoptionConfig(t, dir, config.Adoption{Enabled: true, DefaultStopPct: 0.005})
	eng, err := openGateEngine(t, dir, srv, matchedGuardian())
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	if eng.Config.Engine.Adoption.Enabled {
		t.Error("a refused block left adoption enabled")
	}

	entries := readAudit(t, filepath.Join(dir, "audit.log"))
	toggle := lastEntryFor(entries, "engine.adoption.enabled", audit.ActionAdoptionToggle)
	if toggle == nil {
		t.Fatalf("no adoption entry recorded; entries = %+v", entries)
	}
	if toggle.New != "false" {
		t.Errorf("recorded value = %q, want false: the block was refused", toggle.New)
	}
	if toggle.Detail == "" {
		t.Error("a refused block must record why; without it the trail says the operator asked for " +
			"nothing rather than that the engine declined what they asked for")
	}
}

// TestAdoptionStaysOffOnAGateOffEngine is the §0.2 half at the wiring level: the
// engine everybody runs today is gate-off and has no adoption block, and adding
// the feature must not have changed what it does.
func TestAdoptionStaysOffOnAGateOffEngine(t *testing.T) {
	dir := isolate(t)
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	srv, _ := interlockServer(t, "123-45")
	writeGateConfig(t, dir, config.AutomationGate{})

	eng, err := openGateEngine(t, dir, srv, nil)
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	if eng.Config.Engine.Adoption.Enabled {
		t.Error("adoption is on in an engine whose config has no adoption block")
	}
	if eng.Automation.Verified {
		t.Error("the automation gate is verified on a gate-off engine")
	}
}
