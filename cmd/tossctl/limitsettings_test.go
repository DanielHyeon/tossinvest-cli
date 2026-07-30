package main

// limitsettings_test.go pins the wiring end of the Guardian-limit seam (change
// console-sets-guardian-limits, task 8.1): the save-time audit appends the
// before AND after value engine-safety asks for, the surgical write holds
// through the real seam, and the automation gate's own switch survives untouched.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
)

func TestTheLimitSeamSavesAuditsAndLeavesTheSwitchAlone(t *testing.T) {
	cfgDir := t.TempDir()
	dataDir := t.TempDir()
	t.Setenv("TOSSOS_DATA_DIR", dataDir)

	// The gate is ON and already carries a hand-written ceiling. Both facts are
	// load-bearing: the switch must survive, and the audit's before-image must be
	// the file's value rather than "nothing recorded yet".
	body := `{
  "schema_version": 5,
  "$unknown": {"kept": true},
  "trading": {"note": "keep me"},
  "engine": {
    "automation_gate": {"enabled": true, "max_order_notional": 700000},
    "adoption": {"enabled": true, "default_stop_pct": 0.05}
  }
}`
	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	seam := newLimitSettingsSeam(&rootOptions{configDir: cfgDir})
	if seam == nil {
		t.Fatal("the seam did not resolve a --config-dir profile")
	}
	tier, ok := config.GuardianTierByID("kr-small-live")
	if !ok {
		t.Fatal("the default tier is not registered")
	}
	if err := seam.Save(tier.Limits); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// engine-safety: 게이트 토글·한도 변경 등 운영 설정 변경은 변경 전후 값·시각·
	// 주체를 audit 로그로 기록해야 한다(SHALL). The before-image is the file's, so
	// a hand-edit between two console saves cannot make the trail lie.
	auditData, err := os.ReadFile(filepath.Join(dataDir, "audit.log"))
	if err != nil {
		t.Fatalf("no audit entry was appended at save time: %v", err)
	}
	for _, want := range []string{
		"console.automation_gate.max_order_notional",
		"console.automation_gate.limit_currency",
		"1000000", // the value it moved to
		"700000",  // the value it moved from, read off the file
	} {
		if !strings.Contains(string(auditData), want) {
			t.Errorf("the audit log does not carry %q", want)
		}
	}

	after, err := os.ReadFile(filepath.Join(cfgDir, "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"$unknown"`, "keep me", `"adoption"`, `"enabled": true`} {
		if !strings.Contains(string(after), want) {
			t.Errorf("the save dropped %q from the file:\n%s", want, after)
		}
	}

	// The switch itself: still on, and still on because nobody wrote it.
	gate, err := seam.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !gate.Enabled {
		t.Error("a limit save turned the automation gate off")
	}
	if gate.Limits() != tier.Limits {
		t.Errorf("round trip: %+v, want %+v", gate.Limits(), tier.Limits)
	}
}

// TestAnUnsetCeilingIsAuditedAsUnsetRatherThanZero: "0" in a trail reads as a
// limit of zero, which is the one thing this codebase insists an absent limit is
// not.
func TestAnUnsetCeilingIsAuditedAsUnsetRatherThanZero(t *testing.T) {
	cfgDir := t.TempDir()
	dataDir := t.TempDir()
	t.Setenv("TOSSOS_DATA_DIR", dataDir)

	body := `{"schema_version":5,"trading":{},"engine":{"automation_gate":{"enabled":false}}}`
	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	seam := newLimitSettingsSeam(&rootOptions{configDir: cfgDir})
	tier, _ := config.GuardianTierByID("us-smoke")
	if err := seam.Save(tier.Limits); err != nil {
		t.Fatalf("Save: %v", err)
	}

	auditData, err := os.ReadFile(filepath.Join(dataDir, "audit.log"))
	if err != nil {
		t.Fatalf("no audit entry: %v", err)
	}
	if !strings.Contains(string(auditData), "미설정") {
		t.Error("a first-time ceiling was audited without saying it came from nothing")
	}
}

func TestATypedNilLimitSeamNeverReachesTheInterface(t *testing.T) {
	if s := consoleLimitSettingsSeam(nil); s != nil {
		if p, ok := s.(*consoleLimitSettings); ok && p == nil {
			t.Fatal("a typed nil escaped into the console.LimitSettings interface")
		}
	}
}
