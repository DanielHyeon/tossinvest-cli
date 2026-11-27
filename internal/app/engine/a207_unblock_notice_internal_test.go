package engine

import (
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// a207 리뷰 r1 B-P2 — 재편입은 포지션 id 를 두고 세대만 올림. 옛 세대의 거절 셈이 새 세대의 첫 수용에 「해소」로 붙으면 안 됨.
func TestA207AStreakDoesNotCrossALifecycleGeneration(t *testing.T) {
	o := &ExitObserver{clk: clock.NewFake(time.Date(2026, 11, 27, 1, 0, 0, 0, time.UTC))}
	gen1 := managed{position: journal.Position{ID: "pos-a207", Symbol: "005930"}, state: journal.ExitState{LifecycleGeneration: 1}}
	gen2 := gen1
	gen2.state.LifecycleGeneration = 2
	stop := exitpolicy.Proposal{Action: exitpolicy.ActionBaselineBreach, Level: "BASELINE"}

	o.noteRefusalStreak(gen1, journal.ProposalRefused)
	o.noteRefusalStreak(gen1, journal.ProposalRefused)
	o.noteUnblocked(gen2, stop, "exit-gen2-first")
	if len(o.unblockNotices) != 0 {
		t.Fatalf("notices = %+v — generation 1 refusals were reported as generation 2's unblock", o.unblockNotices)
	}
	// 세대 1 의 셈은 떼어졌으므로 다시 붙지 않음.
	o.noteRefusalStreak(gen1, journal.ProposalRefused)
	o.noteRefusalStreak(gen2, journal.ProposalRefused)
	o.noteUnblocked(gen2, stop, "exit-gen2-second")
	if len(o.unblockNotices) != 1 || o.unblockNotices[0].Fields["refusals"] != 1 {
		t.Fatalf("notices = %+v, want one generation-2 notice counting only its own refusal", o.unblockNotices)
	}
}
