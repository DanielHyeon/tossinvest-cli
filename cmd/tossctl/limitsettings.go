package main

// limitsettings.go wires the console's Guardian-limit editor (change
// console-sets-guardian-limits, task 8.2): a two-method seam over
// config.Service's raw gate read and its limit-only surgical write, plus a
// save-time audit entry per field.
//
// The audit append lives HERE for the same reason recordAdoptionSave does:
// internal/console writes nothing (its static guards say so), and engine-safety
// requires "게이트 토글·한도 변경 등 운영 설정 변경은 변경 전후 값·시각·주체를
// audit 로그로 기록해야 한다(SHALL)" at the moment the person makes the change.
// The engine's startup record says what it started with; this one says when
// somebody moved it and from what.

import (
	"path/filepath"
	"strconv"

	"github.com/JungHoonGhae/tossinvest-cli/internal/audit"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// newLimitSettingsSeam resolves the same config file every other per-profile
// path resolves to, so the screen that displays a ceiling and the screen that
// edits one cannot end up on two different files.
func newLimitSettingsSeam(root *rootOptions) *consoleLimitSettings {
	svc := configServiceFor(root)
	if svc == nil {
		return nil
	}
	return &consoleLimitSettings{svc: svc}
}

type consoleLimitSettings struct{ svc *config.Service }

func (s consoleLimitSettings) Load() (config.AutomationGate, error) {
	return s.svc.LoadRawEngineGate()
}

// Save writes the five ceilings and the currency, and records the change.
//
// The before-image is read here rather than left to audit.RecordChange's own
// "last recorded value" lookup. The two differ exactly when it matters: if the
// file was hand-edited between two console saves, the last audited value is not
// what the operator is actually changing away from, and an audit trail that
// says so is worse than one that admits it does not know.
func (s consoleLimitSettings) Save(l config.GuardianLimits) error {
	before, beforeErr := s.svc.LoadRawEngineGate()
	if err := s.svc.SaveEngineGateLimits(l); err != nil {
		return err
	}
	// Best-effort by design: the save is durable and the engine's startup record
	// is the second trail. A console save must not be rolled back because an
	// audit disk write failed.
	if beforeErr == nil {
		recordLimitSave(before.Limits(), l)
	} else {
		recordLimitSave(config.GuardianLimits{}, l)
	}
	return nil
}

// recordLimitSave appends one entry per field, with the value it moved from.
func recordLimitSave(before, after config.GuardianLimits) {
	dir, err := journal.DataDir()
	if err != nil {
		return
	}
	log, err := audit.Open(audit.Options{Path: filepath.Join(dir, audit.FileName)})
	if err != nil {
		return
	}
	number := func(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }
	for _, field := range []struct {
		setting  string
		from, to string
	}{
		{"console.automation_gate.max_order_quantity",
			number(before.MaxOrderQuantity), number(after.MaxOrderQuantity)},
		{"console.automation_gate.max_order_notional",
			number(before.MaxOrderNotional), number(after.MaxOrderNotional)},
		{"console.automation_gate.max_total_exposure",
			number(before.MaxTotalExposure), number(after.MaxTotalExposure)},
		{"console.automation_gate.max_daily_loss_amount",
			number(before.MaxDailyLossAmount), number(after.MaxDailyLossAmount)},
		{"console.automation_gate.max_daily_loss_ratio",
			number(before.MaxDailyLossRatio), number(after.MaxDailyLossRatio)},
		{"console.automation_gate.limit_currency", before.Currency, after.Currency},
	} {
		_, _ = log.RecordChange(audit.ActionLimitChange, field.setting, field.to,
			"saved from the console settings screen; file value before this save was "+
				quotedOrUnset(field.from))
	}
}

// quotedOrUnset renders the before-image, naming the empty case rather than
// rendering it as an empty string nobody can tell from a missing record.
func quotedOrUnset(v string) string {
	if v == "" || v == "0" {
		return "미설정"
	}
	return v
}
