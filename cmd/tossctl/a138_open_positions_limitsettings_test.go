package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
)

// a138: 콘솔 저장의 최대 보유 포지션 수 audit 는 저장 뒤 파일 값의 변화를 적고, 0 의 뜻(미설정 — 기본 6)을 드러냄.
// 통화를 바꾸는 프리셋 저장은 이 값을 지우지 않음.
func TestA138TheOpenPositionMaximumSaveIsAuditedFromTheFile(t *testing.T) {
	cfgDir, dataDir := t.TempDir(), t.TempDir()
	t.Setenv("TOSSOS_DATA_DIR", dataDir)
	body := `{"schema_version": 5, "engine": {"automation_gate": {"enabled": true, "limit_currency": "KRW"}}}`
	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	seam := newLimitSettingsSeam(&rootOptions{configDir: cfgDir})
	tier, _ := config.GuardianTierByID("kr-small-live")
	four := 4.0
	next := tier.Limits
	next.OpenPositions = &four
	if err := seam.Save(next); err != nil {
		t.Fatal(err)
	}
	usd, _ := config.GuardianTierByID("us-smoke")
	if err := seam.Save(usd.Limits); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dataDir, "audit.log"))
	if err != nil {
		t.Fatal(err)
	}
	log := string(data)
	if strings.Count(log, "console.automation_gate.max_open_positions") != 1 || !strings.Contains(log, "0 (미설정 — 기본 6)") {
		t.Fatalf("the move 0 → 4 must be audited once, naming what 0 meant:\n%s", log)
	}
	gate, err := config.NewService(filepath.Join(cfgDir, "config.json")).LoadRawEngineGate()
	if err != nil || gate.MaxOpenPositions != 4 || gate.LimitCurrency != "USD" {
		t.Fatalf("a currency switch must not clear the count, got %+v (%v)", gate, err)
	}
}
