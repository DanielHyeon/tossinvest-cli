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

// dailyBuyOnlyGuardian 은 5필드와 하루 매수 상한은 진술하지만 최대 보유 포지션 수는 진술하지 않는 Guardian(a138 이전 모양).
type dailyBuyOnlyGuardian struct{ silentGuardian }

func (dailyBuyOnlyGuardian) ExposureLimits() execgw.Limits { return fullGateLimits() }
func (dailyBuyOnlyGuardian) DailyBuyNotionalLimit() riskcalc.Money {
	return riskcalc.Money{Amount: "10000000", Currency: "KRW"}
}

func TestA138TheInterlockAuditsTheOpenPositionMaximum(t *testing.T) {
	dir := isolate(t)
	writeGateConfig(t, dir, fullGate())
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	writeAttestation(t, dir, nil)
	srv, _ := interlockServer(t, "123-45")

	eng, err := openProtectedGateEngine(t, dir, srv, matchedGuardian())
	if err != nil {
		t.Fatalf("a gate without the key resolves to the default 6 and must start: %v", err)
	}
	if got := eng.Automation.MaxOpenPositions; got != 6 {
		t.Fatalf("resolved open position maximum = %v, want 6", got)
	}
	entries := readAudit(t, filepath.Join(dir, "audit.log"))
	accepted := lastEntryFor(entries, "engine.automation_gate.enabled", audit.ActionGateAccepted)
	if accepted == nil || !strings.Contains(accepted.Detail, "max_open_positions=6") {
		t.Fatalf("the acceptance must name the open position maximum, got %+v", accepted)
	}
	line := lastEntryFor(entries, "engine.automation_gate.max_open_positions", audit.ActionLimitChange)
	if line == nil {
		t.Fatal("the open position maximum must be an audited setting (hand edits are traced at startup)")
	}
}

func TestA138TheInterlockRefusesAGuardianThatDisagreesOnTheOpenPositionMaximum(t *testing.T) {
	for name, guardian := range map[string]execgw.Guardian{
		"does not declare":       dailyBuyOnlyGuardian{},
		"declares another value": stubGuardian{limits: fullGateLimits(), maxOpenPositions: 7},
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

func TestA138AnUnusableOpenPositionMaximumRefusesToStart(t *testing.T) {
	for _, bad := range []float64{-1, 2.5} {
		dir := isolate(t)
		gate := fullGate()
		gate.MaxOpenPositions = bad
		writeGateConfig(t, dir, gate)
		writeCredentials(t, dir, "test-api-key-000000", "test-secret")
		writeAttestation(t, dir, nil)
		srv, _ := interlockServer(t, "123-45")
		_, err := openProtectedGateEngine(t, dir, srv, matchedGuardian())
		if !errors.Is(err, engine.ErrLimitsRequired) {
			t.Fatalf("max_open_positions %v must refuse to start as an absent limit, got %v", bad, err)
		}
	}
}
