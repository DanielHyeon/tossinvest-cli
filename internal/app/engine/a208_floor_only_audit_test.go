package engine

// a208 design 8.2 T10 — engine.adoption.floor_only_exit 의 기동 audit 행(§0.5).
//
// recordGateSettings 표 끝(a163 행 뒤)에 덧붙이고, 값이 바뀐 기동에만 old→new 한 행을 씀(RecordChange).
// 생산 호출자(NewContext)는 cfg.Engine.Adoption 을 통째로 넘기므로 호출 줄은 바뀌지 않음.

import (
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/audit"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
)

func TestA208FloorOnlyIsAuditedAtStartup(t *testing.T) {
	log := auditFixture(t)
	_, resolution := resolveNotificationPublisher(config.Notifications{}, fixedEnv(nil))
	record := func(on bool) []audit.Entry {
		t.Helper()
		adoption := config.Adoption{Enabled: true, DefaultStopPct: 0.05, FloorOnlyExit: on}
		if err := recordGateSettings(log, config.AutomationGate{LimitCurrency: "KRW"}, adoption, resolution, "", false); err != nil {
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
	if last.Setting != "engine.adoption.floor_only_exit" || last.Action != audit.ActionAdoptionToggle || last.New != "false" {
		t.Fatalf("last entry = %+v, want the floor-only toggle recorded off as an adoption toggle", last)
	}
	if prev := first[len(first)-2]; prev.Setting != "engine.exit_source_age_report" {
		t.Fatalf("entry before the toggle = %q, want the a163 row — the toggle is appended, never interleaved", prev.Setting)
	}

	turnedOn := record(true)
	if got := len(turnedOn) - len(first); got != 1 {
		t.Fatalf("turning the toggle on wrote %d entries, want exactly one", got)
	}
	if e := turnedOn[len(turnedOn)-1]; e.Setting != "engine.adoption.floor_only_exit" || e.Old != "false" || e.New != "true" {
		t.Fatalf("toggle change entry = %+v, want old false → new true", e)
	}

	if again := record(true); len(again) != len(turnedOn) {
		t.Fatalf("a restart with the same value wrote %d entries, want 0", len(again)-len(turnedOn))
	}
}
