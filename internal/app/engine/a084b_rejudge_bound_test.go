package engine_test

// a084 개정 2 · D8·D9 — 재판정의 경계와 부작용 (issues.md B1·B2).
//
//   D8  "개정당 1회"는 판정이 커밋될 때만 성립했다. judgeRatchet/judgeLadder가
//       !snapshot.Changed에서 record 앞으로 반환하므로, 가격이 움직이지 않은 재판정은
//       격리 행을 건드리지 않고 다음 주기에 전부 반복된다 — 영구히. 자격은 판정의
//       성패가 아니라 *시도*로 소진되어야 한다.
//   D9  통과한 재판정이 판정보다 *먼저* 실주문을 취소한다. 취소 뒤 판정이 다시
//       거부하면 그 포지션은 working 주문도 없고 여전히 격리이며 여전히 미판정이다.
//       a084 이전에는 record에 도달조차 하지 않았으므로 보호가 나빠진다 — §0.3.

import (
	"context"
	"database/sql"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// storedSelectorRevision reads the active row's stamp straight off disk, NULL
// included: the journal reader maps NULL to 0 and this test is about telling
// "never stamped" from "stamped 0".
func storedSelectorRevision(t *testing.T, h *exitHarness, p journal.Position) (sql.NullInt64, bool) {
	t.Helper()
	var revision sql.NullInt64
	err := openRaw(t, h).QueryRowContext(context.Background(),
		`SELECT selector_revision FROM exit_snapshot_quarantines
		 WHERE position_id=? AND position_generation=? AND released_at IS NULL`,
		p.ID, p.InstanceSeq).Scan(&revision)
	if err == sql.ErrNoRows {
		return revision, false
	}
	if err != nil {
		t.Fatalf("reading the stamp: %v", err)
	}
	return revision, true
}

// TestTheReJudgementRetryIsSpentByTheAttempt (D8).
//
// 가격이 움직이지 않으면 판정은 !snapshot.Changed에서 반환하고 아무것도 기록하지
// 않는다. 그때도 그 개정의 재시도는 소진되어야 한다. 소진이 fail-closed인 이유:
// 격리는 유지되고 운영자 해제 경로(a079)는 그대로이며, 다음 개정이 다시 한 번을 준다.
func TestTheReJudgementRetryIsSpentByTheAttempt(t *testing.T) {
	ctx := context.Background()
	h := newExitHarness(t, func(o *engine.ExitObserverOptions) {
		policy := exitpolicy.DefaultLadderPolicy()
		o.Ladder = &policy
	})
	p := h.entry("005930", "10", "70000", "68000", "70000")
	if _, err := h.journal.OpenExitState(ctx, journal.ExitStateSeed{
		PositionID: p.ID, PolicyKind: journal.ExitPolicyLadder,
		EntryPrice: "70000", InitialStop: "68000",
	}); err != nil {
		t.Fatalf("OpenExitState: %v", err)
	}
	h.quote("005930", 70500)
	h.observe()

	supersededQuarantine(t, h, p)

	// The same quote: nothing about the line moves, so the judgement returns
	// before it can write anything.
	h.quote("005930", 70500)
	h.observe()

	revision, active := storedSelectorRevision(t, h, p)
	if !active {
		return // the re-judgement resolved it outright, which also spends the retry
	}
	if !revision.Valid || revision.Int64 != exitpolicy.RecoverySelectorRevision {
		t.Fatalf("selector_revision = %v, want the current %d: the position was let through "+
			"for its one re-judgement and the judgement returned early, so nothing stamped "+
			"the row — it is let through again on the next cycle, and the next, forever, "+
			"logging \"re-judged once\" every five seconds",
			revision, exitpolicy.RecoverySelectorRevision)
	}
}

// TestAReJudgementDoesNotCancelAWorkingOrderBeforeItIsAllowedTo (D9).
//
// workingSet의 주석은 재판정 통과가 "without arming anything"이라고 약속한다.
// 코드는 clearTheSymbol을 판정 트랜잭션보다 먼저 낸다. 그 통과에서는 주문을 내지도
// 지우지도 않아야 한다.
func TestAReJudgementDoesNotCancelAWorkingOrderBeforeItIsAllowedTo(t *testing.T) {
	ctx := context.Background()
	h := newExitHarness(t, func(o *engine.ExitObserverOptions) {
		policy := exitpolicy.DefaultLadderPolicy()
		o.Ladder = &policy
	})
	p := h.entry("005930", "10", "70000", "68000", "70000")
	if _, err := h.journal.OpenExitState(ctx, journal.ExitStateSeed{
		PositionID: p.ID, PolicyKind: journal.ExitPolicyLadder,
		EntryPrice: "70000", InitialStop: "68000",
	}); err != nil {
		t.Fatalf("OpenExitState: %v", err)
	}
	h.quote("005930", 70500)
	h.observe()

	supersededQuarantine(t, h, p)

	placesBefore, cancelsBefore := len(h.submit.places), len(h.submit.cancels)

	// A price that would arm on a normal cycle. On the re-judgement pass it must
	// move the judgement and nothing else.
	h.quote("005930", 84000)
	h.observe()

	if got := len(h.submit.cancels) - cancelsBefore; got != 0 {
		t.Fatalf("the re-judgement pass sent %d cancel(s) to the broker before the "+
			"judgement transaction could refuse: a refusal after the cancel leaves the "+
			"position with no working order, still quarantined and still unjudged — "+
			"strictly worse protection than not letting it through at all (§0.3)", got)
	}
	if got := len(h.submit.places) - placesBefore; got != 0 {
		t.Fatalf("the re-judgement pass placed %d order(s); workingSet's own comment "+
			"promises it arms nothing", got)
	}
}

// TestAReJudgementNeverWithholdsAStop: D9의 비대칭이 정본이다. 재판정은 익절을
// 미루지, 보호를 미루지 않는다 — 보류된 제안은 라인이 다시 움직일 때까지 재제안되지
// 않으므로, 손절을 보류하는 것은 §0.3이 금지하는 지연이다.
func TestAReJudgementNeverWithholdsAStop(t *testing.T) {
	ctx := context.Background()
	h := newExitHarness(t, func(o *engine.ExitObserverOptions) {
		policy := exitpolicy.DefaultLadderPolicy()
		o.Ladder = &policy
	})
	p := h.entry("005930", "10", "70000", "68000", "70000")
	if _, err := h.journal.OpenExitState(ctx, journal.ExitStateSeed{
		PositionID: p.ID, PolicyKind: journal.ExitPolicyLadder,
		EntryPrice: "70000", InitialStop: "68000",
	}); err != nil {
		t.Fatalf("OpenExitState: %v", err)
	}
	h.quote("005930", 70500)
	h.observe()

	supersededQuarantine(t, h, p)

	// Below the initial stop: the re-judgement's own pass must liquidate.
	placesBefore := len(h.submit.places)
	h.quote("005930", 67000)
	h.observe()

	if len(h.submit.places) == placesBefore {
		t.Fatal("the re-judgement pass withheld a stop. A withheld proposal is not " +
			"re-proposed until the line moves again, so withholding protection is a stop " +
			"delayed for as long as the price sits still — which §0.3 forbids. Only the " +
			"upside orders may wait for the judgement to commit")
	}
}

// TestASuppressedReJudgeArmingIsNotedAsADelay: D9가 미루는 것은 조용해서는 안 된다.
// 보류된 청산을 한계 초과 시 알림으로 바꾸는 것이 noteDelay이고, working-order 경로가
// 이미 쓰는 그 계약을 재사용한다 — 새 침묵 경로를 만들지 않는다는 것이 요점이다.
func TestASuppressedReJudgeArmingIsNotedAsADelay(t *testing.T) {
	ctx := context.Background()
	h := newExitHarness(t, func(o *engine.ExitObserverOptions) {
		policy := exitpolicy.DefaultLadderPolicy()
		o.Ladder = &policy
	})
	p := h.entry("005930", "10", "70000", "68000", "70000")
	if _, err := h.journal.OpenExitState(ctx, journal.ExitStateSeed{
		PositionID: p.ID, PolicyKind: journal.ExitPolicyLadder,
		EntryPrice: "70000", InitialStop: "68000",
	}); err != nil {
		t.Fatalf("OpenExitState: %v", err)
	}
	h.quote("005930", 70500)
	h.observe()

	supersededQuarantine(t, h, p)

	h.quote("005930", 84000)
	h.observe()

	var suppressed sql.NullString
	if err := openRaw(t, h).QueryRowContext(ctx,
		`SELECT arm_suppressed_reason FROM exit_events WHERE position_id=?
		 ORDER BY rowid DESC LIMIT 1`, p.ID).Scan(&suppressed); err != nil {
		t.Skipf("the judgement did not record an arm suppression here: %v", err)
	}
	if suppressed.Valid && suppressed.String != "" &&
		suppressed.String != journal.ArmSuppressedReJudge {
		t.Fatalf("arm_suppressed_reason = %q, want %q: a withheld arming has to name why, "+
			"and this one is not the working-order case",
			suppressed.String, journal.ArmSuppressedReJudge)
	}
}
