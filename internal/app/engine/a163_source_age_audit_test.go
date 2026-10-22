package engine

// a163 tasks 2.16 — 보고 토글의 기동 audit 행(§0.5, design D8 「audit」).
//
// recordGateSettings 는 changes 목록 끝(per_trade_risk_budget 뒤)에 engine.exit_source_age_report 행을 덧붙이고, 값이 바뀐 기동에만
// old→new 한 행을 씀(RecordChange). 앞부분 순서는 a074 시험이 무수정으로 지킴. 생산 호출자(NewContext)가 로드된 값을 넘기는지는
// engine_test 의 기동 시험이 잼.

import (
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/audit"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
)

func TestA163TheReportToggleIsTheLastAuditedSettingAndOnlyItsChangeIsWritten(t *testing.T) {
	log := auditFixture(t)
	_, resolution := resolveNotificationPublisher(config.Notifications{}, fixedEnv(nil))
	record := func(on bool) []audit.Entry {
		t.Helper()
		if err := recordGateSettings(log, config.AutomationGate{LimitCurrency: "KRW"}, config.Adoption{}, resolution, "", on); err != nil {
			t.Fatalf("recordGateSettings: %v", err)
		}
		entries, err := log.Entries()
		if err != nil {
			t.Fatalf("Entries: %v", err)
		}
		return entries
	}

	first := record(false)
	last := first[len(first)-1]
	if last.Setting != "engine.exit_source_age_report" || last.Action != audit.ActionExitSourceAgeReport || last.New != "false" {
		t.Fatalf("last entry = %+v, want the report toggle recorded off with its own action", last)
	}
	if prev := first[len(first)-2]; prev.Setting != "engine.automation_gate.per_trade_risk_budget" {
		t.Fatalf("entry before the toggle = %q, want per_trade_risk_budget — the toggle is appended, never interleaved", prev.Setting)
	}
	if audit.ActionExitSourceAgeReport != "exit.source_age_report" {
		t.Fatalf("action = %q, want exit.source_age_report", audit.ActionExitSourceAgeReport)
	}

	turnedOn := record(true)
	if got := len(turnedOn) - len(first); got != 1 {
		t.Fatalf("turning the toggle on wrote %d entries, want exactly one", got)
	}
	if e := turnedOn[len(turnedOn)-1]; e.Setting != "engine.exit_source_age_report" || e.Old != "false" || e.New != "true" {
		t.Fatalf("toggle change entry = %+v, want old false → new true", e)
	}

	if again := record(true); len(again) != len(turnedOn) {
		t.Fatalf("a restart with the same value wrote %d entries, want 0", len(again)-len(turnedOn))
	}
}
