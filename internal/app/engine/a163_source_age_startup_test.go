package engine_test

// a163 tasks 2.16 생산 절반 — NewContext 가 로드된 engine.exit_source_age_report 값을 기동 audit 에 넘김(design D8 「audit」 · 「읽는 시점」).

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/audit"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
)

func a163WriteToggledConfig(t *testing.T, dir string, on bool) {
	t.Helper()
	cfg := config.DefaultFile()
	cfg.Trading = openTradingPolicy()
	cfg.Engine.AutomationGate = smallLiveGate()
	cfg.Engine.ExitSourceAgeReport = on
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), data, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
}

func TestA163AStartupAuditsTheLoadedReportToggle(t *testing.T) {
	dir := isolate(t)
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	writeAttestation(t, dir, nil)
	srv, _ := interlockServer(t, "123-45")
	for i, on := range []bool{false, true, true} {
		a163WriteToggledConfig(t, dir, on)
		eng, err := openProtectedGateEngine(t, dir, srv, nil)
		if err != nil {
			t.Fatalf("start %d: %v", i, err)
		}
		if eng.Config.Engine.ExitSourceAgeReport != on {
			t.Fatalf("start %d loaded the toggle as %v, want %v", i, eng.Config.Engine.ExitSourceAgeReport, on)
		}
		_ = eng.Close()
	}
	var rows []audit.Entry
	for _, e := range readAudit(t, filepath.Join(dir, "audit.log")) {
		if e.Setting == "engine.exit_source_age_report" {
			rows = append(rows, e)
		}
	}
	if len(rows) != 2 || rows[0].New != "false" || rows[1].Old != "false" || rows[1].New != "true" || rows[1].Action != audit.ActionExitSourceAgeReport {
		t.Fatalf("toggle audit rows = %+v, want the first start (→false) and the one change (false→true), none for the repeat", rows)
	}
}
