package obs

// mode.go is add-core-domain task 3.3: an operating-mode transition is announced
// as a critical alert and as one structured log line.
//
// # Why critical
//
// The grading table's rule is "would an operator want to be woken up". A mode
// transition qualifies in both directions and for different reasons:
//
//   - a tightening means the engine has stopped opening positions, automatically,
//     because something went wrong. Nobody finds that out from a dashboard they
//     are not looking at, and every one of the four triggers is a condition that
//     stays broken until a human acts;
//   - a relaxation means somebody re-enabled entries on a live account. That is
//     precisely the change §0.7 wants a second pair of eyes on.
//
// Critical also means durable: the alert is written to the journal's outbox
// before any send, so a transition announced by a process that then dies is
// still an alert somebody receives.
//
// # The feedback loop, and why it terminates
//
// A critical alert that cannot be delivered latches the gate, and sustained
// outbox failure is itself an escalation trigger. So: transition → alert →
// delivery fails → escalate → transition? No — the escalation's target is
// ENTRY_BLOCKED, the account is already there, and journal.TransitionOperatingMode
// reports changed=false and appends nothing. Nothing calls this function for a
// transition that did not happen, so the loop closes after one pass. The
// property is worth naming because it is a property of the *no-op rule*, not of
// anything in this file.

import (
	"context"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// AnnounceOperatingMode reports one committed mode transition.
//
// It implements journal.ModeAnnouncer, so the journal calls it as part of the
// transition flow rather than every call site remembering to. previous is the
// mode the account was in; rec is the row that replaced it.
//
// A nil Notifier is a no-op: an engine wired without alerting still transitions,
// and refusing to would make the safety mechanism depend on the notification
// transport.
func (n *Notifier) AnnounceOperatingMode(ctx context.Context, previous string, rec journal.OperatingModeRecord) error {
	if n == nil {
		return nil
	}
	// 사건 구성은 기록 전용 통지(RecordOnly.AnnounceOperatingMode)와 같은 함수를 씀(a092 K1).
	return n.Notify(ctx, operatingModeEvent(previous, rec))
}

// OperatingModeEvent 는 모드 전이 하나의 통지 사건을 만드는 순수 함수의 공개 입구임(a090 — 기록 전용 공지자가 obs 밖에서
// 같은 내용을 씀). 내용은 operatingModeEvent 한 곳에만 있음: 두 공지가 같은 전이를 다르게 적지 않게.
func OperatingModeEvent(previous string, rec journal.OperatingModeRecord) Event {
	return operatingModeEvent(previous, rec)
}

func modeLabel(mode string) string {
	if mode == "" {
		return journal.ModeNormal
	}
	return mode
}

func permission(rec journal.OperatingModeRecord) string {
	if rec.BlocksEntry() {
		return "refused"
	}
	return "permitted"
}
