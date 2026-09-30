package engine_test

// a094 다각 리뷰(2026-09-30) 수리의 시험 — 보이스 A · B · C 가 짚은 구멍마다 하나.

import (
	"context"
	"database/sql"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

func (h *exitHarness) a094Exec(t *testing.T, stmt string) {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+h.journal.Path())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(stmt); err != nil {
		t.Fatalf("fixture %q: %v", stmt, err)
	}
}

// B P1-c · 3.R9 — 종결 증거 대기 critical 은 재시작 뒤에도 같은 취소면 같은 원장 행(실제 알림기 · outbox).
func TestA094TheCloseWaitEpisodeIsOneRowAcrossRestarts(t *testing.T) {
	h, sub, _ := a094Harness(t, func(o *engine.ExitObserverOptions) {
		o.Critical = &obs.Notifier{Journal: o.Journal}
	})
	p := a094ArmTakeProfit(t, h, sub, journal.StateConfirmed)
	h.submit.settle = nil
	sub.cancelRecords = true
	h.quote("005930", 68500)
	h.observe()
	h.clk.Advance(31 * time.Second)
	h.observe()
	if n := len(h.a094Rows("|" + p.ID + "|cancel:")); n != 1 {
		t.Fatalf("close-wait rows = %d, want 1", n)
	}
	restarted, err := engine.NewExitObserver(h.observer.OptionsForTest())
	if err != nil {
		t.Fatal(err)
	}
	restarted.ObserveOnce(context.Background())
	if n := len(h.a094Rows("|" + p.ID + "|cancel:")); n != 1 {
		t.Fatalf("close-wait rows after the restart = %d, want the same single row", n)
	}
}

// B P1-d · 4.N4f③ — 다른 attempt 가 park 되면 새 행.
func TestA094AnotherParkedAttemptIsANewRow(t *testing.T) {
	h, _, crit := a094Harness(t, nil)
	p := h.entry("005930", "10", "70000", "68000", "70000")
	h.quote("005930", 70100)
	h.observe()
	h.a094ArmOn(p, "005930", "exit-two", string(exitpolicy.ActionBaselineBreach))
	first, _ := h.a094RecordSell("exit-two", 10, 67000, journal.StateUnresolvedInDoubt)
	h.quote("005930", 67000)
	h.observe()
	second, _ := h.a094RecordSell("exit-two", 10, 67000, journal.StateUnresolvedInDoubt)
	h.observe()
	if len(crit.withKey("|attempt:"+first)) != 1 || len(crit.withKey("|attempt:"+second)) != 1 {
		t.Fatalf("park rows = %+v, want one per parked attempt", crit.events)
	}
}

// B P2-b · 4.N4g — 알림기 입구의 기록이 실패하면 진입이 잠기고(입구의 몫) 관측 루프는 계속, 다른 포지션 손절 무영향.
func TestA094ARecordFailureLocksEntriesAndTheLoopGoesOn(t *testing.T) {
	var notifier *obs.Notifier
	h, sub, _ := a094Harness(t, func(o *engine.ExitObserverOptions) {
		notifier = &obs.Notifier{Journal: o.Journal}
		o.Critical = notifier
	})
	notifier.Gate = h.gate
	a094ArmTakeProfit(t, h, sub, journal.StateUnresolvedInDoubt)
	if _, locked := h.gate.Blocks()[execgw.ReasonAlertUndelivered]; locked {
		t.Fatal("control: the gate was already locked by the undelivered-alert reason")
	}
	h.a094Exec(t, `CREATE TRIGGER a094_refuse_alert BEFORE INSERT ON alert_outbox BEGIN SELECT RAISE(ABORT, 'a094 fixture: outbox full'); END`)
	h.entry("000660", "10", "70000", "68000", "70000")
	places := len(h.submit.places)
	h.quote("005930", 68500)
	h.quote("000660", 67000)
	if c := h.observe(); c.Judged != 2 {
		t.Fatalf("judged = %d, want both positions judged despite the record failure", c.Judged)
	}
	if _, locked := h.gate.Blocks()[execgw.ReasonAlertUndelivered]; !locked {
		t.Fatalf("entry gate blocks = %v, want the undelivered-alert reason", h.gate.Blocks())
	}
	if len(h.submit.places) != places+1 {
		t.Fatalf("places %d→%d, want the other position's stop placed", places, len(h.submit.places))
	}
}

// A P2#3 · C P2#2 — 결과 쓰기가 실패해 입증된 비수용 attempt 위에 발의가 남으면, 재시작 없이 다음 관측이 같은 판정으로 풀고
// 손절을 다시 발의함(「종결 후에는 반드시 푼다」).
func TestA094AnUnacceptedProposalLeftArmedIsReleasedNextObservation(t *testing.T) {
	h, sub, _ := a094Harness(t, nil)
	p := h.entry("005930", "10", "70000", "68000", "70000")
	h.quote("005930", 70100)
	h.observe()
	h.a094Exec(t, fmt.Sprintf(`CREATE TRIGGER a094_refuse_release BEFORE UPDATE ON exit_states
		WHEN OLD.position_id = '%s' AND OLD.pending_action IS NOT NULL AND NEW.pending_action IS NULL
		BEGIN SELECT RAISE(ABORT, 'a094 fixture'); END`, p.ID))
	sub.placeState = journal.StateFailedConfirmed
	h.quote("005930", 67000)
	h.observe()
	if !h.state(p.ID).Pending() {
		t.Fatal("control: the release did not fail")
	}
	h.a094Exec(t, `DROP TRIGGER a094_refuse_release`)
	sub.placeState = ""
	h.quote("005930", 66900)
	h.observe() // 판정 진입이 풂(이 주기는 억제된 평가)
	h.quote("005930", 66800)
	h.observe() // 다시 발의
	if len(h.submit.places) != 2 {
		t.Fatalf("places = %d, want the stop proposed again without a restart", len(h.submit.places))
	}
}

// B P2-c · A P3#6 — intent 없는 무장 발의는 침묵하지 않음(명명 critical 한 번).
func TestA094AnIntentlessArmedProposalIsNamed(t *testing.T) {
	h, _, crit := a094Harness(t, nil)
	p := h.entry("005930", "10", "70000", "68000", "70000")
	h.quote("005930", 70100)
	h.observe()
	h.a094ArmOn(p, "005930", "", string(exitpolicy.ActionBaselineBreach))
	h.quote("005930", 67000)
	h.observe()
	h.observe()
	if got := len(crit.withKey("|" + p.ID + "|nointent:")); got != 1 {
		t.Fatalf("intentless alerts = %d, want 1 (%+v)", got, crit.events)
	}
}

// B P2-a · A P3#5 — 청소가 취소한 **다른 intent** 의 매도가 청산 지연 한계 넘게 종결 기록이 없어도 그 취소를 명명함(D−9.3).
func TestA094AnotherIntentsCloseWaitIsNamedToo(t *testing.T) {
	h, sub, crit := a094Harness(t, nil)
	p := h.entry("005930", "10", "70000", "68000", "70000")
	h.quote("005930", 70100)
	h.observe()
	h.a094ArmOn(p, "005930", "exit-tp-armed", string(exitpolicy.ActionRatchetPartial))
	h.a094RecordSell("old-sell", 4, 72000, journal.StateConfirmed)
	h.submit.settle = nil
	sub.cancelRecords = true
	h.quote("005930", 67900)
	h.observe()
	h.clk.Advance(31 * time.Second)
	h.observe()
	if got := len(crit.withKey("|" + p.ID + "|cancel:")); got != 1 {
		t.Fatalf("close-wait alerts = %d, want the other intent's cancelled sell named", got)
	}
}

// A P3#7 · C P2#1 — 판정 없이 비우는 ResolveExitProposal 의 비시험 호출자는 0(해제 판정은 한 곳).
func TestA094TheUnjudgedResolveHasNoProductionCaller(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	var callers []string
	checked := 0
	_ = filepath.Walk(filepath.Join(root, "internal"), func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		checked++
		fset := token.NewFileSet()
		f, perr := parser.ParseFile(fset, path, nil, 0)
		if perr != nil {
			return nil
		}
		ast.Inspect(f, func(n ast.Node) bool {
			if c, ok := n.(*ast.CallExpr); ok {
				if s, ok := c.Fun.(*ast.SelectorExpr); ok && s.Sel.Name == "ResolveExitProposal" {
					callers = append(callers, fset.Position(c.Pos()).String())
				}
			}
			return true
		})
		return nil
	})
	if checked < 100 {
		t.Fatalf("control: only %d files walked", checked)
	}
	if len(callers) != 0 {
		t.Errorf("ResolveExitProposal has production callers %v — releases must go through ReleaseUnacceptedExitProposal", callers)
	}
}

// A P2#2 — 연속 실패 경보의 래치는 기록이 성공했을 때만: 첫 기록이 실패하면 다음 실패 주기가 다시 기록함.
func TestA094AFailedStreakRecordIsRetried(t *testing.T) {
	h, sub, crit := a094Harness(t, nil)
	p := h.entry("005930", "10", "70000", "68000", "70000")
	h.workingEntry("005930", "5", "69500")
	sub.cancelFails = true
	h.quote("005930", 67900)
	crit.err = fmt.Errorf("outbox write failed")
	for i := 0; i < obs.DefaultCriticalAttempts; i++ {
		h.observe()
	}
	if got := len(crit.withKey("|streak:")); got != 0 {
		t.Fatalf("control: %d streak records while the recorder fails", got)
	}
	crit.err = nil
	h.observe()
	if got := len(crit.withKey("|" + p.ID + "|streak:")); got != 1 {
		t.Fatalf("streak records after the recorder recovered = %d, want the retry to land once", got)
	}
}

// C P2#4 — 기동 따라잡기가 목록을 못 읽으면 이 기동을 에피소드로 한 critical 하나(로그만 두지 않음).
func TestA094ABootCatchUpListFailureIsNamed(t *testing.T) {
	h, _, _ := a094Harness(t, nil)
	h.a094Exec(t, `ALTER TABLE exit_states RENAME TO exit_states_gone`)
	crit := &a094Criticals{}
	if released := engine.CatchUpExitProposalsForTest(context.Background(), h.journal, exitAccount, crit); released != 0 {
		t.Fatalf("released = %d on a list failure", released)
	}
	if got := len(crit.withKey("|catchup-list:")); got != 1 || crit.remind[0] != 0 {
		t.Fatalf("catch-up list alerts = %+v, want one window-0 record", crit.events)
	}
}
