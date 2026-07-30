package config

// limits_io.go is the console's second config write surface: the automation
// gate's five ceilings and their currency, and nothing else in that block
// (change console-sets-guardian-limits, task 4.x).
//
// # Why this is not "marshal the block and splice it"
//
// SaveEngineAdoption replaces the whole engine.adoption value in one go, which
// is right there: every key in that block is one the screen edits. The
// automation gate is not like that. It also carries `enabled` — the §0.7
// switch that decides whether the engine may place orders unattended — and the
// console is not allowed to write it.
//
// Re-emitting the block would mean re-emitting `enabled` on every save, sourced
// from a read the handler took OUTSIDE the file lock. That is the lost-update
// property the adoption seam already documents, and on the adoption block its
// worst case is a dropped symbol. Here its worst case is a gate that flips.
//
// So the write is per key: each of the six values is replaced or inserted where
// it sits, and `enabled` and `attestation_file` are never in the member list.
// The console cannot write the switch — not because this file is careful, but
// because it emits no bytes for it (design D6).

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

// LoadRawEngineGate reads the engine.automation_gate block as written.
//
// It returns the whole block, `enabled` included, because the screen has to say
// whether the gate is on — a limit that is about to matter reads differently
// from one that is not. The write side takes GuardianLimits, which has no such
// field, so there is no shape in which this value travels back.
//
// A missing file is the zero gate and no error: the screen renders 미설정, which
// is the truth, rather than an error nobody can act on.
func (s *Service) LoadRawEngineGate() (AutomationGate, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return AutomationGate{}, nil
	}
	if err != nil {
		return AutomationGate{}, err
	}

	var doc struct {
		Engine struct {
			Gate rawAutomationGate `json:"automation_gate"`
		} `json:"engine"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return AutomationGate{}, fmt.Errorf("config: parsing %s: %w", s.path, err)
	}

	raw := doc.Engine.Gate
	gate := AutomationGate{
		AttestationFile:    raw.AttestationFile,
		MaxOrderQuantity:   raw.MaxOrderQuantity,
		MaxOrderNotional:   raw.MaxOrderNotional,
		MaxTotalExposure:   raw.MaxTotalExposure,
		MaxDailyLossAmount: raw.MaxDailyLossAmount,
		MaxDailyLossRatio:  raw.MaxDailyLossRatio,
		// The file's own spelling, not normalised: this is a read of what is
		// written, and the screen shows what is written.
		LimitCurrency: raw.LimitCurrency,
	}
	if raw.Enabled != nil {
		gate.Enabled = *raw.Enabled
	}
	return gate, nil
}

// Limits projects the editable six out of a gate block.
func (g AutomationGate) Limits() GuardianLimits {
	return GuardianLimits{
		MaxOrderQuantity:   g.MaxOrderQuantity,
		MaxOrderNotional:   g.MaxOrderNotional,
		MaxTotalExposure:   g.MaxTotalExposure,
		MaxDailyLossAmount: g.MaxDailyLossAmount,
		MaxDailyLossRatio:  g.MaxDailyLossRatio,
		Currency:           g.LimitCurrency,
	}
}

// SaveEngineGateLimits writes the six values into config.json, surgically.
//
// Two refusals come before any byte is written, and they are different claims:
// Validate is "the engine would refuse to start on this" (the same rule the
// startup interlock applies), and the ceiling is this package's own fat-finger
// backstop. Neither is the caller's to skip — the console checks them too so it
// can show a useful message, but the check that matters is this one.
func (s *Service) SaveEngineGateLimits(l GuardianLimits) error {
	next := l
	next.Currency = normaliseCurrency(l.Currency)

	if err := next.Validate(); err != nil {
		return fmt.Errorf(
			"config: refusing to save Guardian limits the engine would refuse to start on: %w", err)
	}
	if violations := next.CeilingViolations(); len(violations) > 0 {
		return fmt.Errorf("config: refusing to save Guardian limits above the registered ceiling: %s",
			strings.Join(violations, "; "))
	}

	lock, err := acquireConfigLock(s.path + ".lock")
	if err != nil {
		return err
	}
	defer lock.release()

	data, err := os.ReadFile(s.path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		// Skeleton creation is for a file that does not exist, only. It goes
		// through DefaultFile, whose gate is off — so the one `enabled` this file
		// can ever emit is the default's false, into a document that had no
		// previous value to preserve.
		file := DefaultFile()
		file.Engine.AutomationGate.MaxOrderQuantity = next.MaxOrderQuantity
		file.Engine.AutomationGate.MaxOrderNotional = next.MaxOrderNotional
		file.Engine.AutomationGate.MaxTotalExposure = next.MaxTotalExposure
		file.Engine.AutomationGate.MaxDailyLossAmount = next.MaxDailyLossAmount
		file.Engine.AutomationGate.MaxDailyLossRatio = next.MaxDailyLossRatio
		file.Engine.AutomationGate.LimitCurrency = next.Currency
		skeleton, err := json.MarshalIndent(file, "", "  ")
		if err != nil {
			return err
		}
		return s.installBytes(append(skeleton, '\n'))
	case err != nil:
		return err
	}

	if !json.Valid(data) {
		return fmt.Errorf("config: %s is not valid JSON; refusing to overwrite a file that may be "+
			"a hand-edit in progress", s.path)
	}

	out, err := spliceGateLimits(data, next)
	if err != nil {
		return err
	}
	return s.installBytes(out)
}

// gateMember is one key/value pair the writer is allowed to touch.
type gateMember struct {
	key   string
	value []byte
}

// gateMembersOf is the complete, closed list of what a limit save writes.
//
// `enabled` and `attestation_file` are absent, and their absence is the whole
// mechanism of design D6 — a reviewer checks this one function rather than
// auditing every caller.
func gateMembersOf(l GuardianLimits) ([]gateMember, error) {
	currency, err := json.Marshal(l.Currency)
	if err != nil {
		return nil, err
	}
	return []gateMember{
		{"max_order_quantity", []byte(formatLimit(l.MaxOrderQuantity))},
		{"max_order_notional", []byte(formatLimit(l.MaxOrderNotional))},
		{"max_total_exposure", []byte(formatLimit(l.MaxTotalExposure))},
		{"max_daily_loss_amount", []byte(formatLimit(l.MaxDailyLossAmount))},
		{"max_daily_loss_ratio", []byte(formatLimit(l.MaxDailyLossRatio))},
		{"limit_currency", currency},
	}, nil
}

// spliceGateLimits replaces or inserts each member inside
// engine.automation_gate, leaving every other byte of the document alone.
//
// The block's span is re-located before each member because the previous
// member's edit moved the offsets. Re-scanning six times is not the fast way to
// do this; it is the way that cannot write into a stale offset.
func spliceGateLimits(data []byte, l GuardianLimits) ([]byte, error) {
	members, err := gateMembersOf(l)
	if err != nil {
		return nil, err
	}

	for _, m := range members {
		gStart, gEnd, found, err := gateValueSpan(data)
		if err != nil {
			return nil, err
		}
		if !found {
			if data, err = insertEmptyGate(data); err != nil {
				return nil, err
			}
			if gStart, gEnd, found, err = gateValueSpan(data); err != nil {
				return nil, err
			} else if !found {
				return nil, errors.New("config: the automation_gate block could not be created")
			}
		}

		kStart, kEnd, kFound, err := valueSpan(data[gStart:gEnd], m.key)
		if err != nil {
			return nil, err
		}
		if kFound {
			out := make([]byte, 0, len(data)+len(m.value))
			out = append(out, data[:gStart+kStart]...)
			out = append(out, m.value...)
			out = append(out, data[gStart+kEnd:]...)
			data = out
			continue
		}
		if data, err = insertKey(data, gStart, gEnd, m.key, m.value); err != nil {
			return nil, err
		}
	}
	return data, nil
}

// insertEmptyGate adds an empty engine.automation_gate object, creating the
// engine block too when it is absent. It is empty on purpose: the members are
// spliced in afterwards by the same loop that handles an existing block, so
// there is exactly one code path that decides which keys a save writes.
func insertEmptyGate(data []byte) ([]byte, error) {
	eStart, eEnd, eFound, err := valueSpan(data, "engine")
	if err != nil {
		return nil, err
	}
	if eFound {
		return insertKey(data, eStart, eEnd, "automation_gate", []byte("{}"))
	}
	return insertKey(data, 0, int64(len(data)), "engine", []byte(`{"automation_gate": {}}`))
}

// gateValueSpan locates the byte span of the engine.automation_gate value.
func gateValueSpan(data []byte) (start, end int64, found bool, err error) {
	eStart, eEnd, eFound, err := valueSpan(data, "engine")
	if err != nil || !eFound {
		return 0, 0, false, err
	}
	gStart, gEnd, gFound, err := valueSpan(data[eStart:eEnd], "automation_gate")
	if err != nil || !gFound {
		return 0, 0, false, err
	}
	return eStart + gStart, eStart + gEnd, true, nil
}
