package main

// operatingsettings.go wires the console's two operating editors (change
// console-owns-the-operating-toggles): the four trading toggles the engine's
// exit path uses, and the automation gate's switch.
//
// The audit append lives here for the same reason recordLimitSave does:
// internal/console writes nothing (its static guards say so), and engine-safety
// requires "게이트 토글·한도 변경 등 운영 설정 변경은 변경 전후 값·시각·주체를
// audit 로그로 기록해야 한다(SHALL)" at the moment the person makes the change.
//
// It matters more here than anywhere else on the settings screen. Until this
// change the gate could only be flipped by hand-editing config.json, which left
// no trail at all — the engine's startup record said what it came up with, and
// nothing said when somebody decided it. The line this file writes is the only
// answer to "when was automatic trading turned on, and from what".

import (
	"path/filepath"

	"github.com/JungHoonGhae/tossinvest-cli/internal/audit"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// newTradingPolicySeam resolves the same config file every other per-profile
// path resolves to.
func newTradingPolicySeam(root *rootOptions) *consoleTradingPolicy {
	svc := configServiceFor(root)
	if svc == nil {
		return nil
	}
	return &consoleTradingPolicy{svc: svc}
}

type consoleTradingPolicy struct{ svc *config.Service }

func (s consoleTradingPolicy) Load() (config.Trading, error) { return s.svc.LoadRawTrading() }

func (s consoleTradingPolicy) Save(p config.TradingPolicy) error {
	before, beforeErr := s.svc.LoadRawTrading()
	if err := s.svc.SaveTradingPolicy(p); err != nil {
		return err
	}
	// Best-effort, exactly as recordLimitSave is: the save is durable and a
	// console edit must not be rolled back because an audit disk write failed.
	if beforeErr != nil {
		before = config.Trading{}
	}
	recordPolicySave(config.TradingPolicyOf(before), p)
	return nil
}

// newGateSwitchSeam is the write side of engine.automation_gate.enabled.
func newGateSwitchSeam(root *rootOptions) *consoleGateSwitch {
	svc := configServiceFor(root)
	if svc == nil {
		return nil
	}
	return &consoleGateSwitch{svc: svc}
}

type consoleGateSwitch struct{ svc *config.Service }

func (s consoleGateSwitch) Save(on bool) error {
	// The before-image is read here rather than left to audit.RecordChange's own
	// lookup, for the reason consoleLimitSettings.Save gives: a hand-edit between
	// two console saves makes the last audited value the wrong "from", and a
	// trail that admits it does not know beats one that guesses.
	before, beforeErr := s.svc.LoadRawEngineGate()
	if err := s.svc.SaveEngineGateEnabled(on); err != nil {
		return err
	}
	old := "unknown"
	if beforeErr == nil {
		old = boolText(before.Enabled)
	}
	recordGateFlip(old, boolText(on))
	return nil
}

func boolText(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func openAuditLog() *audit.Log {
	dir, err := journal.DataDir()
	if err != nil {
		return nil
	}
	log, err := audit.Open(audit.Options{Path: filepath.Join(dir, audit.FileName)})
	if err != nil {
		return nil
	}
	return log
}

// recordPolicySave appends one entry per toggle that moved.
//
// Per toggle rather than one line for the block, matching recordLimitSave: a
// reader asking "when did selling become allowed" should find that sentence, not
// a diff they have to compute.
func recordPolicySave(before, after config.TradingPolicy) {
	log := openAuditLog()
	if log == nil {
		return
	}
	for _, field := range []struct {
		setting     string
		from, to    bool
	}{
		{"trading.place", before.Place, after.Place},
		{"trading.sell", before.Sell, after.Sell},
		{"trading.cancel", before.Cancel, after.Cancel},
		{"trading.allow_live_order_actions", before.AllowLiveOrderActions, after.AllowLiveOrderActions},
	} {
		if field.from == field.to {
			continue
		}
		_ = log.Record(audit.Entry{
			Action:  audit.ActionTradingPolicy,
			Setting: field.setting,
			Old:     boolText(field.from),
			New:     boolText(field.to),
			Detail:  "operator console, 거래 정책 저장",
		})
	}
}

// recordGateFlip is the one line this whole change exists to make possible.
func recordGateFlip(old, next string) {
	log := openAuditLog()
	if log == nil {
		return
	}
	detail := "operator console, 자동화 게이트 저장"
	if next == "true" {
		detail += " — 다음 엔진 기동에서 대사·exit 관측·체결 감지 루프가 시작된다"
	}
	_ = log.Record(audit.Entry{
		Action:  audit.ActionGateToggle,
		Setting: "engine.automation_gate.enabled",
		Old:     old,
		New:     next,
		Detail:  detail,
	})
}
