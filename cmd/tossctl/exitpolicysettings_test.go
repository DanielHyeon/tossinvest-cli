package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

func TestExitPolicySeamSavesAuditsAndPreservesUnrelatedConfig(t *testing.T) {
	cfgDir, dataDir := t.TempDir(), t.TempDir()
	t.Setenv("TOSSOS_DATA_DIR", dataDir)
	body := `{"schema_version":5,"$future":{"keep":true},"trading":{"sell":false},
	  "engine":{"automation_gate":{"enabled":false},"adoption":{"enabled":false}}}`
	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	seam := newExitPolicySettingsSeam(&rootOptions{configDir: cfgDir})
	if err := seam.Save(config.ExitPolicy{CommonPolicy: exitpolicy.CommonLadderHybrid50}); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(filepath.Join(cfgDir, "config.json"))
	for _, want := range []string{`"$future":{"keep":true}`, `"automation_gate":{"enabled":false}`,
		`"adoption":{"enabled":false}`, exitpolicy.CommonLadderHybrid50} {
		if !strings.Contains(string(after), want) {
			t.Errorf("config missing %q:\n%s", want, after)
		}
	}
	auditData, err := os.ReadFile(filepath.Join(dataDir, "audit.log"))
	if err != nil {
		t.Fatalf("audit missing: %v", err)
	}
	for _, want := range []string{"exit_policy.common", "engine.exit_policy.common_policy",
		exitpolicy.CommonLadderHybrid50} {
		if !strings.Contains(string(auditData), want) {
			t.Errorf("audit missing %q:\n%s", want, auditData)
		}
	}
}
