package engine_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/audit"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
)

// a139: 유도된 건당 위험 예산이 기동 audit 와 인터록 수용 문구에 남음 — 규칙(1/3)이 바뀐 기동을 기록으로 가를 수 있게.
func TestA139TheInterlockAuditsTheDerivedRiskBudget(t *testing.T) {
	dir := isolate(t)
	writeGateConfig(t, dir, fullGate()) // 일일 손실 200,000 KRW → 66,666
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	writeAttestation(t, dir, nil)
	srv, _ := interlockServer(t, "123-45")
	if _, err := openProtectedGateEngine(t, dir, srv, matchedGuardian()); err != nil {
		t.Fatalf("start: %v", err)
	}
	entries := readAudit(t, filepath.Join(dir, "audit.log"))
	accepted := lastEntryFor(entries, "engine.automation_gate.enabled", audit.ActionGateAccepted)
	if accepted == nil || !strings.Contains(accepted.Detail, "per_trade_risk_budget=66666") {
		t.Fatalf("the acceptance must name the derived risk budget, got %+v", accepted)
	}
	line := lastEntryFor(entries, "engine.automation_gate.per_trade_risk_budget", audit.ActionLimitChange)
	if line == nil || line.New != "66666" || !strings.Contains(line.Detail, "floor(max_daily_loss_amount / 3") {
		t.Fatalf("the derived risk budget must be an audited setting, got %+v", line)
	}
}

// freeze 리뷰 P1-2: 게이트 OFF 설정은 일일 손실이 없을 수 있음 — 유도가 불가능해도 기동은 지금처럼 성공하고(OFF = upstream),
// audit 줄은 빈 값과 사유를 남김.
func TestA139AGateOffStartWithoutADailyLossStillStarts(t *testing.T) {
	dir := isolate(t)
	writeGateConfig(t, dir, config.AutomationGate{})
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	srv, _ := interlockServer(t, "123-45")
	if _, err := openGateEngine(t, dir, srv, nil); err != nil {
		t.Fatalf("a gate-off engine with no daily loss must start as before: %v", err)
	}
	line := lastEntryFor(readAudit(t, filepath.Join(dir, "audit.log")), "engine.automation_gate.per_trade_risk_budget", audit.ActionLimitChange)
	if line == nil || line.New != "" || !strings.Contains(line.Detail, "cannot be derived") {
		t.Fatalf("an underivable budget must be audited as empty with its reason, got %+v", line)
	}
}
