package config

// operating_io.go is the console's third and fourth config write surfaces: the
// automation gate's switch, and the four trading-policy toggles the engine's
// exit path actually uses (change console-owns-the-operating-toggles).
//
// # Why each write is its own function with its own closed member list
//
// limits_io.go established the rule and the reason: re-emitting a block means
// re-emitting keys the screen did not edit, sourced from a read the handler took
// outside the file lock. On the gate block its worst case was a flipped switch,
// which is why `enabled` was excluded from `gateMembersOf` and the console could
// not write it at all.
//
// This change gives the console the switch — the prohibition was an over-reading
// of §0.7, and it pushed the most consequential toggle in the system onto a
// hand-edit path with no validation and no audit. What it does NOT give up is
// the mechanism: `gateSwitchMembers` emits one key and `tradingMembersOf` emits
// four, so a limit save still cannot move the switch and a switch save still
// cannot move a limit. A reviewer checks three functions, not every caller.
//
//	gateMembersOf        five ceilings + currency   (limits_io.go)
//	gateSwitchMembers    enabled                    (here)
//	tradingMembersOf     place, sell, cancel,
//	                     allow_live_order_actions   (here)
//
// The three lists are disjoint, and that is the whole guarantee.

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// TradingPolicy is the console's editable slice of the `trading` block.
//
// Four fields and not seven. `amend`, `conditional` and `fractional` are what
// this build's engine loops never reach, so the screen does not offer them and
// this type cannot carry them — a save leaves whatever the file spells for those
// three exactly as it found it.
type TradingPolicy struct {
	Place                 bool
	Sell                  bool
	Cancel                bool
	AllowLiveOrderActions bool
}

// TradingPolicyOf projects the editable four out of a full trading block.
func TradingPolicyOf(t Trading) TradingPolicy {
	return TradingPolicy{
		Place:                 t.Place,
		Sell:                  t.Sell,
		Cancel:                t.Cancel,
		AllowLiveOrderActions: t.AllowLiveOrderActions,
	}
}

// Complete reports that all four are on, which is what interlock clause 3
// requires. The console renders the missing ones by name rather than saying
// "incomplete".
func (p TradingPolicy) Complete() bool {
	return p.Place && p.Sell && p.Cancel && p.AllowLiveOrderActions
}

// Missing names the toggles that are off, in the order the interlock enumerates
// them, so the screen's list and the engine's refusal read the same.
func (p TradingPolicy) Missing() []string {
	var out []string
	for _, c := range []struct {
		on   bool
		name string
	}{
		{p.Place, "trading.place"},
		{p.Sell, "trading.sell"},
		{p.Cancel, "trading.cancel"},
		{p.AllowLiveOrderActions, "trading.allow_live_order_actions"},
	} {
		if !c.on {
			out = append(out, c.name)
		}
	}
	return out
}

// LoadRawTrading reads the `trading` block as written.
//
// A missing file is the zero policy and no error, for the same reason
// LoadRawEngineGate says: the screen renders every toggle off, which is the
// truth, rather than an error nobody can act on.
func (s *Service) LoadRawTrading() (Trading, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return Trading{}, nil
	}
	if err != nil {
		return Trading{}, err
	}
	var doc struct {
		Trading Trading `json:"trading"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return Trading{}, fmt.Errorf("config: parsing %s: %w", s.path, err)
	}
	return doc.Trading, nil
}

// LoadEngineAutostart reads the operator's process-lifecycle approval.
//
// Missing files, engine blocks and keys all mean false. Autostart is not
// inferred from the automation gate: the machine may be armed for a deliberate
// manual start without being approved to start after every reboot.
func (s *Service) LoadEngineAutostart() (bool, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	var doc struct {
		Engine struct {
			Autostart bool `json:"autostart"`
		} `json:"engine"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return false, fmt.Errorf("config: parsing %s: %w", s.path, err)
	}
	return doc.Engine.Autostart, nil
}

// SaveTradingPolicy writes the four toggles into config.json, surgically.
func (s *Service) SaveTradingPolicy(p TradingPolicy) error {
	return s.spliceInto(func(data []byte) ([]byte, error) {
		return spliceMembers(data, tradingValueSpan, insertEmptyTrading, tradingMembersOf(p))
	})
}

// SaveEngineGateEnabled writes the automation gate's switch, and nothing else.
//
// There is no validation here on purpose. Turning the gate ON with an
// unsatisfiable interlock is not a corrupt config — it is a configuration the
// engine will refuse to start on, loudly, enumerating what is unmet. Refusing
// the save instead would mean this function had to reproduce the interlock's
// judgement, and two implementations of one rule is how they drift.
//
// What the console does do is show the operator the clauses it CAN judge from
// the file before they press it (design D3). That is advice; this is a record.
func (s *Service) SaveEngineGateEnabled(on bool) error {
	return s.spliceInto(func(data []byte) ([]byte, error) {
		return spliceMembers(data, gateValueSpan, insertEmptyGate, gateSwitchMembers(on))
	})
}

// SaveEngineAutostart writes engine.autostart and no other key.
func (s *Service) SaveEngineAutostart(on bool) error {
	return s.spliceInto(func(data []byte) ([]byte, error) {
		return spliceMembers(data, engineValueSpan, insertEmptyEngine,
			[]gateMember{{"autostart", []byte(boolLiteral(on))}})
	})
}

// gateSwitchMembers is the complete, closed list of what a switch save writes.
func gateSwitchMembers(on bool) []gateMember {
	return []gateMember{{"enabled", []byte(boolLiteral(on))}}
}

// tradingMembersOf is the complete, closed list of what a policy save writes.
//
// The three absent keys — amend, conditional, fractional — are absent for the
// same reason `enabled` is absent from gateMembersOf: a save emits no bytes for
// what the screen does not edit.
func tradingMembersOf(p TradingPolicy) []gateMember {
	return []gateMember{
		{"place", []byte(boolLiteral(p.Place))},
		{"sell", []byte(boolLiteral(p.Sell))},
		{"cancel", []byte(boolLiteral(p.Cancel))},
		{"allow_live_order_actions", []byte(boolLiteral(p.AllowLiveOrderActions))},
	}
}

func boolLiteral(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

// spliceInto is the lock-read-validate-write shell both savers share.
func (s *Service) spliceInto(edit func([]byte) ([]byte, error)) error {
	lock, err := acquireConfigLock(s.path + ".lock")
	if err != nil {
		return err
	}
	defer lock.release()

	data, err := os.ReadFile(s.path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		// The skeleton's trading block is all-false and its gate is off, so a
		// first save into a non-existent file writes the operator's choice on top
		// of defaults that grant nothing.
		file := DefaultFile()
		skeleton, mErr := json.MarshalIndent(file, "", "  ")
		if mErr != nil {
			return mErr
		}
		data = append(skeleton, '\n')
	case err != nil:
		return err
	}

	if !json.Valid(data) {
		return fmt.Errorf("config: %s is not valid JSON; refusing to overwrite a file that may be "+
			"a hand-edit in progress", s.path)
	}
	out, err := edit(data)
	if err != nil {
		return err
	}
	if !json.Valid(out) {
		return errors.New("config: the surgical write produced invalid JSON; nothing was written")
	}
	return s.installBytes(out)
}

// spliceMembers replaces or inserts each member inside the block `span` locates,
// creating the block with `create` when it is absent.
//
// The span is re-located before every member because the previous member's edit
// moved the offsets. Re-scanning is not the fast way; it is the way that cannot
// write into a stale offset (limits_io.go says the same about its own loop).
func spliceMembers(
	data []byte,
	span func([]byte) (int64, int64, bool, error),
	create func([]byte) ([]byte, error),
	members []gateMember,
) ([]byte, error) {
	for _, m := range members {
		bStart, bEnd, found, err := span(data)
		if err != nil {
			return nil, err
		}
		if !found {
			if data, err = create(data); err != nil {
				return nil, err
			}
			if bStart, bEnd, found, err = span(data); err != nil {
				return nil, err
			} else if !found {
				return nil, errors.New("config: the target block could not be created")
			}
		}
		kStart, kEnd, kFound, err := valueSpan(data[bStart:bEnd], m.key)
		if err != nil {
			return nil, err
		}
		if kFound {
			out := make([]byte, 0, len(data)+len(m.value))
			out = append(out, data[:bStart+kStart]...)
			out = append(out, m.value...)
			out = append(out, data[bStart+kEnd:]...)
			data = out
			continue
		}
		if data, err = insertKey(data, bStart, bEnd, m.key, m.value); err != nil {
			return nil, err
		}
	}
	return data, nil
}

// tradingValueSpan locates the top-level `trading` object.
func tradingValueSpan(data []byte) (start, end int64, found bool, err error) {
	return valueSpan(data, "trading")
}

// engineValueSpan locates the top-level engine object.
func engineValueSpan(data []byte) (start, end int64, found bool, err error) {
	return valueSpan(data, "engine")
}

// insertEmptyEngine creates the top-level engine block without granting any
// capability. The caller splices the one requested key afterwards.
func insertEmptyEngine(data []byte) ([]byte, error) {
	return insertKey(data, 0, int64(len(data)), "engine", []byte("{}"))
}

// insertEmptyTrading adds an empty top-level `trading` object.
//
// The whole document is the object to insert into, spelled the way
// insertEmptyGate spells the same case for a missing `engine`.
func insertEmptyTrading(data []byte) ([]byte, error) {
	return insertKey(data, 0, int64(len(data)), "trading", []byte("{}"))
}
