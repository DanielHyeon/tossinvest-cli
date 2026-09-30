package engine_test

// a094 4.3b · 4.3d — 기동 따라잡기: 무장 발의마다 판정 하나를 한 번. 비수용 종결뿐이거나 attempt 가 없으면 풀고,
// 접수 · park · 미종결이면 두며, 한 행의 실패는 그 행만 두고 다음으로 간다(포지션 key 의 critical).

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

func TestA094TheBootCatchUpReleasesOnlyUnacceptedProposals(t *testing.T) {
	h, _, _ := a094Harness(t, nil)
	ctx := context.Background()
	symbols := []struct {
		symbol string
		state  journal.AttemptState // "" = attempt 없음
		free   bool
	}{
		{"005930", journal.StateFailedConfirmed, true},
		{"000660", "", true},
		{"035420", journal.StateConfirmed, false},
		{"051910", journal.StateUnresolvedInDoubt, false},
		{"068270", journal.StateInDoubt, false},
	}
	ids := map[string]string{}
	for i, s := range symbols {
		p := h.entry(s.symbol, "10", "70000", "68000", "70000")
		h.quote(s.symbol, 70100)
		h.observe()
		intent := fmt.Sprintf("exit-boot-%d", i)
		h.a094ArmOn(p, s.symbol, intent, string(exitpolicy.ActionBaselineBreach))
		if s.state != "" {
			h.a094SellAttemptOn(s.symbol, intent, s.state)
		}
		ids[s.symbol] = p.ID
	}
	released := engine.CatchUpExitProposalsForTest(ctx, h.journal, exitAccount, nil)
	if released != 2 {
		t.Errorf("released = %d, want 2", released)
	}
	for _, s := range symbols {
		if got := h.state(ids[s.symbol]).Pending(); got == s.free {
			t.Errorf("%s (%q): armed = %v, want %v", s.symbol, s.state, got, !s.free)
		}
	}
	// 멱등.
	if again := engine.CatchUpExitProposalsForTest(ctx, h.journal, exitAccount, nil); again != 0 {
		t.Errorf("a second catch-up released %d, want 0", again)
	}
}

// 4.3d — 한 행의 쓰기 실패: 그 행은 그대로, 나머지는 처리, 포지션 · intent key 의 critical.
func TestA094ABootCatchUpRowFailureIsNamedAndSkipped(t *testing.T) {
	h, _, _ := a094Harness(t, nil)
	ctx := context.Background()
	var bad, good journal.Position
	for i, symbol := range []string{"005930", "000660"} {
		p := h.entry(symbol, "10", "70000", "68000", "70000")
		h.quote(symbol, 70100)
		h.observe()
		h.a094ArmOn(p, symbol, fmt.Sprintf("exit-row-%d", i), string(exitpolicy.ActionBaselineBreach))
		if i == 0 {
			bad = p
		} else {
			good = p
		}
	}
	db, err := sql.Open("sqlite", "file:"+h.journal.Path())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(fmt.Sprintf(`CREATE TRIGGER a094_refuse_release BEFORE UPDATE ON exit_states
		WHEN OLD.position_id = '%s' AND NEW.pending_action IS NULL BEGIN SELECT RAISE(ABORT, 'a094 fixture'); END`, bad.ID)); err != nil {
		t.Fatalf("fixture trigger: %v", err)
	}
	_ = db.Close()
	crit := &a094Criticals{}
	released := engine.CatchUpExitProposalsForTest(ctx, h.journal, exitAccount, crit)
	if released != 1 || h.state(good.ID).Pending() {
		t.Fatalf("released = %d (good armed %v), want the other row released", released, h.state(good.ID).Pending())
	}
	if !h.state(bad.ID).Pending() {
		t.Fatal("the failing row changed")
	}
	named := crit.withKey("|" + bad.ID + "|intent:exit-row-0")
	if len(named) != 1 || crit.remind[0] != 0 {
		t.Fatalf("row-failure alerts = %+v, want one keyed by position and intent, window 0", crit.events)
	}
}

func (h *exitHarness) a094SellAttemptOn(symbol, intentID string, to journal.AttemptState) {
	h.t.Helper()
	if symbol == "005930" {
		h.a094RecordSell(intentID, 10, 67000, to)
		return
	}
	ctx := context.Background()
	h.ids++
	id := fmt.Sprintf("a-boot-%d", h.ids)
	attempt, err := h.journal.Prepare(ctx, journal.PrepareRequest{
		Intent: journal.Intent{ID: intentID, Market: "kr", TradingDay: "2026-03-30", AccountRef: exitAccount,
			Symbol: symbol, Side: "SELL", OrderType: "LIMIT", TimeInForce: "DAY", Quantity: "10", Price: "67000",
			Currency: "KRW", Source: "engine/test", Fingerprint: "fp-" + intentID},
		Kind: journal.KindPlace, AttemptID: id, AccountRef: exitAccount,
	})
	if err != nil {
		h.t.Fatal(err)
	}
	must := func(err error) {
		h.t.Helper()
		if err != nil {
			h.t.Fatal(err)
		}
	}
	must(attempt.MarkDispatchStarted(ctx))
	switch to {
	case journal.StateFailedConfirmed:
		must(attempt.Settle(ctx, journal.StateFailedConfirmed, "x", "x"))
	case journal.StateConfirmed:
		must(attempt.MarkAcked(ctx, "O-"+id))
		must(attempt.Settle(ctx, journal.StateConfirmed, "broker_accepted", ""))
	case journal.StateInDoubt, journal.StateUnresolvedInDoubt:
		must(attempt.MarkInDoubt(ctx, "dispatch_outcome_unknown", "x"))
		if to == journal.StateUnresolvedInDoubt {
			must(attempt.ResolveUnresolved(ctx, "in_doubt_unresolved", "parked"))
		}
	}
}
