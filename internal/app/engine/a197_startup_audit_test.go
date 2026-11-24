package engine_test

// a197 design D8 · Manager 결정 O-1 — NewContext 의 기동 audit 단계가 로드된 engine.cancel_external_opposite_buy 값을 값이 바뀐
// 기동에만 한 줄 남기고, 기존 audit 순서(a074 · a163 행)는 그대로임.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/audit"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
)

func a197WriteToggledConfig(t *testing.T, dir string, on bool) {
	t.Helper()
	cfg := config.DefaultFile()
	cfg.Trading = openTradingPolicy()
	cfg.Engine.AutomationGate = smallLiveGate()
	cfg.Engine.CancelExternalOppositeBuy = on
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), data, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
}

func TestA197StartupAuditsTheLoadedExternalBuyToggle(t *testing.T) {
	dir := isolate(t)
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	writeAttestation(t, dir, nil)
	srv, _ := interlockServer(t, "123-45")
	for i, on := range []bool{false, true, true} {
		a197WriteToggledConfig(t, dir, on)
		eng, err := openProtectedGateEngine(t, dir, srv, nil)
		if err != nil {
			t.Fatalf("start %d: %v", i, err)
		}
		if eng.Config.Engine.CancelExternalOppositeBuy != on {
			t.Fatalf("start %d loaded the toggle as %v, want %v", i, eng.Config.Engine.CancelExternalOppositeBuy, on)
		}
		_ = eng.Close()
	}
	entries := readAudit(t, filepath.Join(dir, "audit.log"))
	var rows []audit.Entry
	lastSourceAge := -1
	for i, e := range entries {
		if e.Setting == "engine.cancel_external_opposite_buy" {
			rows = append(rows, e)
		}
		if e.Setting == "engine.exit_source_age_report" {
			lastSourceAge = i
		}
	}
	if len(rows) != 2 || rows[0].New != "false" || rows[1].Old != "false" || rows[1].New != "true" ||
		rows[1].Action != audit.ActionExternalOppositeBuyCancel {
		t.Fatalf("toggle audit rows = %+v, want the first start (→false) and the one change (false→true), none for the repeat", rows)
	}
	if lastSourceAge < 0 {
		t.Fatal("the a163 report row is missing from the startup audit")
	}
}
