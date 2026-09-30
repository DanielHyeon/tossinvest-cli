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
	detail := rec.Mode
	if rec.Actor != "" {
		detail += " (" + rec.Actor + ")"
	}
	if rec.Cause != "" {
		detail += ": " + rec.Cause
	}

	// 한 번의 잠금 안에서 교체함(a092 AC2) — 지우고 잠금을 놓은 뒤 다시 넣으면 그 사이의 진입 점검이 모드 사유를 못 봄.
	g.mu.Lock()
	defer g.mu.Unlock()
	// 커밋 순서 울타리(a092 C5): 투영은 커밋 뒤 잠금 없이 불리므로 겹친 두 전이의 투영이 뒤바뀌어 도착할 수 있음.
	// 마지막으로 적용한 순번보다 큰 것만 적용 — 초기값 0, 「보다 큰」(순번 0 은 원장 행이 아님).
	if rec.Seq <= g.modeSeq {
		return
	}
	g.modeSeq = rec.Seq
	_, had := g.latches[ReasonOperatingModeBlocked]
	if !rec.BlocksEntry() {
		if had {
			delete(g.latches, ReasonOperatingModeBlocked)
			g.revision++
		}
		return
	}
	g.latches[ReasonOperatingModeBlocked] = detail
	// 상태 세대는 모드 사유의 **존재**가 바뀔 때만 +1(a092 C20 — a124 「실제로 상태가 바뀐 때만」). 설명만 바뀌는 교체는 +0.
	// 해제 세대(clearEpochs)는 건드리지 않음 — 그것은 그 사유의 Clear 만 바꿈(a124).
	if !had {
		g.revision++
	}
}

// OperatingModeBlocked reports whether the gate currently carries a mode latch,
// and its detail. It is a read for status output; the authority is the journal.
func (g *EntryGate) OperatingModeBlocked() (string, bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	detail, blocked := g.latches[ReasonOperatingModeBlocked]
	return detail, blocked
}
