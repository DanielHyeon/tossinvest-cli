package engine_test

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/audit"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// fiveFieldGuardian 은 5필드만 진술하고 하루 매수 상한은 진술하지 않는 Guardian(a136 이전 모양).
type fiveFieldGuardian struct{ silentGuardian }

func (fiveFieldGuardian) ExposureLimits() execgw.Limits { return fullGateLimits() }

func TestA136TheInterlockAuditsTheDailyBuyCeiling(t *testing.T) {
	dir := isolate(t)
	writeGateConfig(t, dir, fullGate())
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	writeAttestation(t, dir, nil)
	srv, _ := interlockServer(t, "123-45")

	eng, err := openProtectedGateEngine(t, dir, srv, matchedGuardian())
	if err != nil {
		t.Fatalf("a KRW gate without the key resolves to the default and must start: %v", err)
	}
	if got := eng.Automation.DailyBuyNotional; !got.Set || got.Value != 10_000_000 {
		t.Fatalf("resolved daily buy ceiling = %+v, want 10,000,000", got)
	}
	entries := readAudit(t, filepath.Join(dir, "audit.log"))
	accepted := lastEntryFor(entries, "engine.automation_gate.enabled", audit.ActionGateAccepted)
	if accepted == nil || !strings.Contains(accepted.Detail, "max_daily_buy_notional=10000000") {
		t.Fatalf("the acceptance must name the daily buy ceiling, got %+v", accepted)
	}
	if lastEntryFor(entries, "engine.automation_gate.max_daily_buy_notional", audit.ActionLimitChange) == nil {
		t.Fatal("the daily buy ceiling must be an audited setting (hand edits are traced at startup)")
	}
}

func TestA136TheInterlockRefusesAGuardianThatDisagreesOnTheDailyBuyCeiling(t *testing.T) {
	for name, guardian := range map[string]execgw.Guardian{
		"does not declare": fiveFieldGuardian{},
		"declares another value": stubGuardian{limits: fullGateLimits(),
			dailyBuy: riskcalc.Money{Amount: "5000000", Currency: "KRW"}},
		"declares another currency": stubGuardian{limits: fullGateLimits(),
			dailyBuy: riskcalc.Money{Amount: "10000000", Currency: "USD"}},
	} {
		t.Run(name, func(t *testing.T) {
			dir := isolate(t)
			writeGateConfig(t, dir, fullGate())
			writeCredentials(t, dir, "test-api-key-000000", "test-secret")
			writeAttestation(t, dir, nil)
			srv, _ := interlockServer(t, "123-45")
			_, err := openProtectedGateEngine(t, dir, srv, guardian)
			if !errors.Is(err, engine.ErrGuardianLimitsMismatch) {
				t.Fatalf("want a single-source mismatch, got %v", err)
			}
		})
	}
}

func TestA136AUSDGateWithoutTheDailyBuyCeilingRefusesToStart(t *testing.T) {
	dir := isolate(t)
	gate := fullGate()
	gate.LimitCurrency = "USD"
	writeGateConfig(t, dir, gate)
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	writeAttestation(t, dir, nil)
	srv, _ := interlockServer(t, "123-45")
	_, err := openProtectedGateEngine(t, dir, srv, stubGuardian{limits: gateLimitsWith(func(l *execgw.Limits) { l.Currency = "USD" })})
	if !errors.Is(err, engine.ErrLimitsRequired) {
		t.Fatalf("a USD gate with no daily buy ceiling must refuse to start (D5), got %v", err)
	}
}
