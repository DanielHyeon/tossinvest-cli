package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
)

// a136: 콘솔 저장의 하루 매수 상한 audit 는 저장 뒤 **파일 값**의 변화를 적음 — 폼 값 저장과 통화 전환 지움 둘 다.
func TestA136TheDailyBuyCeilingSaveIsAuditedFromTheFile(t *testing.T) {
	cfgDir, dataDir := t.TempDir(), t.TempDir()
	t.Setenv("TOSSOS_DATA_DIR", dataDir)
	body := `{"schema_version": 5, "engine": {"automation_gate": {"enabled": true, "limit_currency": "KRW", "max_daily_buy_notional": 3000000}}}`
	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	seam := newLimitSettingsSeam(&rootOptions{configDir: cfgDir})
	tier, _ := config.GuardianTierByID("kr-small-live")
	five := 5_000_000.0
	next := tier.Limits
	next.DailyBuy = &five
	if err := seam.Save(next); err != nil {
		t.Fatal(err)
	}
	// 통화 전환 프리셋 — writer 가 지우고 audit 는 그 지움을 적어야 함.
	usd, _ := config.GuardianTierByID("us-smoke")
	if err := seam.Save(usd.Limits); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dataDir, "audit.log"))
	if err != nil {
		t.Fatal(err)
	}
	log := string(data)
	if strings.Count(log, "console.automation_gate.max_daily_buy_notional") != 2 || !strings.Contains(log, "3000000") ||
		!strings.Contains(log, "5000000") {
		t.Fatalf("the daily buy ceiling moves (3000000 → 5000000 → 0) must both be audited:\n%s", log)
	}
	gate, err := config.NewService(filepath.Join(cfgDir, "config.json")).LoadRawEngineGate()
	if err != nil || gate.MaxDailyBuyNotional != 0 || gate.LimitCurrency != "USD" {
		t.Fatalf("after switching to USD the ceiling must be cleared, got %+v (%v)", gate, err)
	}
	if !gate.Enabled {
		t.Fatal("the switch must survive the save")
	}
}
