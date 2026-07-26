package execgw

// modegate.go is the entry gate's half of "모드의 강제 지점은 EntryGate 투영이다"
// (risk-management, add-core-domain task 3.1).
//
// # Why the mode is a latch and not a new kind of check
//
// The submission sequence is sealed (engine-safety): the gateway asks
// CheckEntryFor before every exposure-raising mutation and that call site does
// not change. Adding a mode check anywhere in that sequence would mean editing
// it; projecting the mode onto the latch map the sequence *already* consults
// means the enforcement arrives without the sequence knowing there is a new
// concept. That is the whole reason D3 chose a projection over a check.
//
// # The projection replaces, it does not accumulate
//
// One account has one mode, so the latch is set or cleared to match the latest
// row — never merged with what was there before. A mode latch this process
// raised in a previous life that the journal does not carry is a latch nothing
// can release, and a HALT_ALL whose detail still reads "ENTRY_BLOCKED: daily
// loss" is a latch an operator cannot act on.

import "github.com/JungHoonGhae/tossinvest-cli/internal/journal"

// ProjectOperatingMode makes the gate agree with one operating-mode row.
//
// It implements journal.ModeProjector, which is what makes "전환 = 영속 + 투영"
// a single flow rather than two things a caller has to remember to do in order:
// journal.TransitionOperatingMode calls this after the commit, and
// RestoreOperatingModeProjection calls it at startup with the latest row.
//
// A mode that permits entries clears the latch. That is a relaxation, and the
// journal has already refused to reach here without an operator and an audit
// line — the gate does not re-litigate the approval, it applies the decision.
func (g *EntryGate) ProjectOperatingMode(rec journal.OperatingModeRecord) {
	g.mu.Lock()
	delete(g.latches, ReasonOperatingModeBlocked)
	g.mu.Unlock()

	if !rec.BlocksEntry() {
		return
	}
	detail := rec.Mode
	if rec.Actor != "" {
		detail += " (" + rec.Actor + ")"
	}
	if rec.Cause != "" {
		detail += ": " + rec.Cause
	}
	g.Block(ReasonOperatingModeBlocked, detail)
}

// OperatingModeBlocked reports whether the gate currently carries a mode latch,
// and its detail. It is a read for status output; the authority is the journal.
func (g *EntryGate) OperatingModeBlocked() (string, bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	detail, blocked := g.latches[ReasonOperatingModeBlocked]
	return detail, blocked
}
