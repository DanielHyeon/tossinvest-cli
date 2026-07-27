package main

// adoptionsettings_test.go pins the wiring end of the settings seam (독립 검증
// finding 2): the save-time audit actually appends, the surgical write holds
// through the real seam, and a typed-nil never reaches the interface.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
)

func TestTheSeamSavesAuditsAndPreservesTheFile(t *testing.T) {
	cfgDir := t.TempDir()
	dataDir := t.TempDir()
	t.Setenv("TOSSOS_DATA_DIR", dataDir)

	body := `{
  "schema_version": 5,
  "$unknown": {"kept": true},
  "trading": {"note": "keep me"},
  "engine": {"automation_gate": {"enabled": false}}
}`
	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	seam := newAdoptionSettingsSeam(&rootOptions{configDir: cfgDir})
	if seam == nil {
		t.Fatal("the seam did not resolve a --config-dir profile")
	}
	if err := seam.Save(config.Adoption{
		DefaultStopPct: 0.05, IncludeSymbols: []string{"005930"},
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// The dual record's save-time half exists (operator-console SHALL — 기동
	// 없는 flip도 기록에 남는다).
	auditData, err := os.ReadFile(filepath.Join(dataDir, "audit.log"))
	if err != nil {
		t.Fatalf("no audit entry was appended at save time: %v", err)
	}
	for _, want := range []string{"console.adoption.include_symbols", "005930"} {
		if !strings.Contains(string(auditData), want) {
			t.Errorf("the audit log does not carry %q", want)
		}
	}

	// The surgical write held through the wiring: everything outside the block
	// survives.
	after, _ := os.ReadFile(filepath.Join(cfgDir, "config.json"))
	for _, want := range []string{`"$unknown"`, "keep me", `"automation_gate"`} {
		if !strings.Contains(string(after), want) {
			t.Errorf("the save dropped %q from the file", want)
		}
	}
	block, verdict, err := seam.Load()
	if err != nil || verdict != "" || !block.Included("005930") {
		t.Errorf("round trip: block=%+v verdict=%q err=%v", block, verdict, err)
	}
}

func TestATypedNilSeamNeverReachesTheInterface(t *testing.T) {
	if s := consoleSettingsSeam(nil); s != nil {
		// DefaultPaths resolves on every supported platform, so a real seam is
		// fine — what must never happen is a non-nil interface wrapping a nil
		// pointer.
		if p, ok := s.(*consoleAdoptionSettings); ok && p == nil {
			t.Fatal("a typed nil escaped into the console.AdoptionSettings interface")
		}
	}
}
