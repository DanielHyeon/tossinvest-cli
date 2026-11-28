package engine_test

// a208 design 8.2 T1~T6 — 편입 전용 보호선 정책(engine.adoption.floor_only_exit).
//
// 켜져 있으면 자동 편입(adoptOne)은 공통 정책과 무관하게 RUNNER 편입 사본을 기록함 → 부분익절 0, 보호선만 상향.
// 꺼져 있으면 기록 정책 · 알림 문구가 a208 이전과 바이트로 같음(§0.2). 손절(합성 손절가)은 토글과 무관함.

import (
	"context"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// a208 시험 전용 합성 심볼 — 실제 종목과 무관함.
const a208Symbol = "999990"

// a208BaseAdoptionBody 는 a208 이전 편입 알림 본문 바이트 그대로임(adoption.go — 「보호선 승격,」 뒤 공백 없음까지 원문).
const a208BaseAdoptionBody = "계좌가 보유 중인데 이를 설명하는 진입 결정이 없어 편입 기록을 남겼다. 지금부터 보호선 승격," +
	"중간 익절, 손절이 엔진이 직접 연 포지션과 동일하게 적용된다.\n" +
	"기준선은 매입 원가가 아니라 **편입 시점에 관측된 가격**으로 잡는다 — 오래 들고 있던 " +
	"수익 종목은 지금부터 보호되고, 과거 수익은 R 척도에 포함되지 않는다."

// a208Driver 는 공통 정책과 토글을 정한 편입 드라이버 하네스를 만듦.
func a208Driver(t *testing.T, commonPolicy string, floorOnly bool) *driverHarness {
	t.Helper()
	return newDriverHarness(t, func(o *engine.ReconcileDriverOptions) {
		o.CommonPolicy = commonPolicy
		o.Adoption.FloorOnlyExit = floorOnly
	})
}

// a208AdoptOnce 는 한 주기를 돌려 편입이 일어났음을 확인하고 편입 기록 · exit state 를 돌려줌.
func a208AdoptOnce(t *testing.T, h *driverHarness, symbol string) (journal.PositionAdoption, journal.ExitState) {
	t.Helper()
	cycle := h.cycle()
	if cycle.Err != nil || cycle.Adopted != 1 {
		t.Fatalf("cycle = %+v, want exactly one adoption", cycle)
	}
	p := h.position(symbol)
	adoption, err := h.journal.AdoptionOf(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("AdoptionOf: %v", err)
	}
	state, err := h.journal.ExitState(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("ExitState: %v", err)
	}
	return adoption, state
}

func a208AdoptionAlert(t *testing.T, h *driverHarness) obs.Event {
	t.Helper()
	event, ok := h.alerts.first(obs.EventExitPositionAdopted)
	if !ok {
		t.Fatal("no adoption alert was sent")
	}
	return event
}

// T1 — ON + BALANCED: 기록 · exit state 정책이 RUNNER 편입 사본이고 그 표의 부분익절은 전부 0, 합성 손절은 그대로.
func TestA208FloorOnlyAdoptionRecordsTheRunnerCopy(t *testing.T) {
	h := a208Driver(t, exitpolicy.CommonLadderBalanced, true)
	h.holds(a208Symbol, "10", "12000", 10000)

	adoption, state := a208AdoptOnce(t, h, a208Symbol)
	if adoption.ExitPolicyID != exitpolicy.CommonLadderRunner {
		t.Fatalf("adoption policy = %q, want %q regardless of the common policy", adoption.ExitPolicyID,
			exitpolicy.CommonLadderRunner)
	}
	if state.PolicyKind != journal.ExitPolicyLadder || state.PolicyID != exitpolicy.CommonLadderRunner {
		t.Fatalf("exit state = %s/%s, want LADDER/%s", state.PolicyKind, state.PolicyID, exitpolicy.CommonLadderRunner)
	}
	if state.PolicyIdentity.Version != "1.0.0-adopted.1" {
		t.Fatalf("policy identity = %+v, want the adopted RUNNER copy 1.0.0-adopted.1", state.PolicyIdentity)
	}
	ladder, err := exitpolicy.CommonLadderForPosition(state.PolicyID, true)
	if err != nil {
		t.Fatalf("CommonLadderForPosition: %v", err)
	}
	for i, rung := range ladder.Rungs {
		if rung.PartialRatio != "0" {
			t.Errorf("rung %d partial = %s, want 0 — a floor-only adoption never takes profit", i, rung.PartialRatio)
		}
	}
	// 손절 기준은 토글과 무관하게 편입 관측가 × (1 − pct) — 평균단가(12000)가 아님.
	if adoption.SyntheticStop != "9500" || state.InitialStop != "9500" || state.EntryPrice != "10000" {
		t.Fatalf("stop = %s / %s, entry = %s, want 9500 / 9500 from the 10000 observation",
			adoption.SyntheticStop, state.InitialStop, state.EntryPrice)
	}
}

// T2 — ON + 공통 정책 없음: RATCHET(+1.0R 40% 부분익절)이 아니라 LADDER RUNNER 사본.
func TestA208FloorOnlyAdoptionReplacesRatchet(t *testing.T) {
	h := a208Driver(t, "", true)
	h.holds(a208Symbol, "10", "", 10000)

	adoption, state := a208AdoptOnce(t, h, a208Symbol)
	if adoption.ExitPolicyID != exitpolicy.CommonLadderRunner || state.PolicyKind != journal.ExitPolicyLadder {
		t.Fatalf("adoption = %q, state kind = %q, want the LADDER RUNNER copy instead of RATCHET",
			adoption.ExitPolicyID, state.PolicyKind)
	}
}

// T3 — OFF: 기록 정책 = 공통 정책, 알림 본문 = a208 이전 바이트.
func TestA208OffAdoptionIsUnchanged(t *testing.T) {
	h := a208Driver(t, exitpolicy.CommonLadderBalanced, false)
	h.holds(a208Symbol, "10", "12000", 10000)

	adoption, state := a208AdoptOnce(t, h, a208Symbol)
	if adoption.ExitPolicyID != exitpolicy.CommonLadderBalanced || state.PolicyID != exitpolicy.CommonLadderBalanced {
		t.Fatalf("adoption = %q, state = %q, want the common policy when the toggle is off",
			adoption.ExitPolicyID, state.PolicyID)
	}
	if body := a208AdoptionAlert(t, h).Body; body != a208BaseAdoptionBody {
		t.Fatalf("alert body changed with the toggle off:\n got %q\nwant %q", body, a208BaseAdoptionBody)
	}
}

// T4 — ON: 알림이 중간 익절을 하지 않는다고 말하고, 「동일하게 적용」 문장을 말하지 않음.
func TestA208FloorOnlyAlertSaysNoPartialTakeProfit(t *testing.T) {
	h := a208Driver(t, exitpolicy.CommonLadderHybrid50, true)
	h.holds(a208Symbol, "10", "12000", 10000)

	a208AdoptOnce(t, h, a208Symbol)
	event := a208AdoptionAlert(t, h)
	if !strings.Contains(event.Body, "보호선 승격과 손절만 적용되고 중간 익절은 하지 않는다") {
		t.Fatalf("alert body does not say the floor-only policy takes no partial profit: %q", event.Body)
	}
	if strings.Contains(event.Body, "중간 익절, 손절이 엔진이 직접 연 포지션과 동일하게") {
		t.Fatalf("alert body still claims the engine's partial take-profit applies: %q", event.Body)
	}
	if got := event.Fields["exit_policy_id"]; got != exitpolicy.CommonLadderRunner {
		t.Fatalf("alert exit_policy_id = %v, want %s", got, exitpolicy.CommonLadderRunner)
	}
}

// T6 — ON 으로 편입된 포지션은 토글을 끈 뒤에도 RUNNER 사본으로 남고, 이후 편입만 공통 정책을 따름.
func TestA208TurningTheToggleOffKeepsTheAdoptedPolicy(t *testing.T) {
	var opts engine.ReconcileDriverOptions
	h := newDriverHarness(t, func(o *engine.ReconcileDriverOptions) {
		o.CommonPolicy = exitpolicy.CommonLadderBalanced
		o.Adoption.FloorOnlyExit = true
		opts = *o
	})
	h.holds(a208Symbol, "10", "12000", 10000)
	a208AdoptOnce(t, h, a208Symbol)
	first := h.position(a208Symbol)

	// 같은 원장 위에 토글을 끈 드라이버로 다시 세움(재기동과 같음).
	opts.Adoption.FloorOnlyExit = false
	restarted, err := engine.NewReconcileDriver(opts)
	if err != nil {
		t.Fatalf("NewReconcileDriver: %v", err)
	}
	h.driver = restarted
	h.holds("999991", "5", "8000", 8000)
	cycle := h.cycle()
	if cycle.Err != nil || cycle.Adopted != 1 {
		t.Fatalf("second cycle = %+v, want exactly the new holding adopted", cycle)
	}

	state, err := h.journal.ExitState(context.Background(), first.ID)
	if err != nil {
		t.Fatalf("ExitState: %v", err)
	}
	if state.PolicyID != exitpolicy.CommonLadderRunner || state.PolicyIdentity.Version != "1.0.0-adopted.1" {
		t.Fatalf("the earlier adoption was reinterpreted: %+v", state.PolicyIdentity)
	}
	later, err := h.journal.AdoptionOf(context.Background(), h.position("999991").ID)
	if err != nil {
		t.Fatalf("AdoptionOf: %v", err)
	}
	if later.ExitPolicyID != exitpolicy.CommonLadderBalanced {
		t.Fatalf("the adoption after turning the toggle off = %q, want the common policy", later.ExitPolicyID)
	}
}

// T5 — RUNNER 편입 기록(T1 이 쓰는 값)의 포지션을 관측기로 구동: 모든 rung 을 지나도 제출 0, 기준선만 단조 상승,
// 첫 rung 위 첫 관측에서 바로 승격(유예 없음 — 사람 결정 2026-11-28 Q1).
func TestA208FloorOnlyAdoptedPositionNeverTakesProfit(t *testing.T) {
	h := newExitHarness(t, nil)
	ctx := context.Background()
	watermark, err := h.journal.FillWatermark(ctx, a208Symbol)
	if err != nil {
		t.Fatal(err)
	}
	fold, err := h.journal.ApplyPositionAdjustment(ctx, journal.AdjustmentRequest{
		AccountRef: exitAccount, Market: "kr", Symbol: a208Symbol, Kind: journal.AdjustmentExternal,
		ExpectedPrevQuantity: "0", ExpectedFillWatermark: watermark,
		NewQuantity: "10", NewAvgPrice: "12000",
		BrokerAsOf: h.clk.Now().Format("2006-01-02T15:04:05Z07:00"),
		Evidence:   "a208 floor-only fixture",
	})
	if err != nil {
		t.Fatalf("fold: %v", err)
	}
	if _, err := h.journal.AdoptPosition(ctx, journal.AdoptionRequest{
		PositionID: fold.Position.ID, Symbol: a208Symbol, Market: "kr", Quantity: "10",
		CostBasis: "12000", ObservedPrice: "10000", SyntheticStop: "9500",
		ObservedAt: journal.RFC3339(h.clk.Now()), ExitPolicyID: exitpolicy.CommonLadderRunner,
	}); err != nil {
		t.Fatalf("adopt: %v", err)
	}

	// +2.6% → t0 잠금(10000), +4.6% → +2.0%(10200), +7.1% → +3.5%(10350), +12% → 그대로 보호선만.
	steps := []struct {
		price    float64
		baseline string
	}{{10260, "10000"}, {10460, "10200"}, {10710, "10350"}, {11200, "10350"}}
	for i, step := range steps {
		h.prices.last[a208Symbol] = step.price
		cycle := h.observer.ObserveOnce(ctx)
		if cycle.Err != nil {
			t.Fatalf("step %d: %v", i, cycle.Err)
		}
		if cycle.Proposed != 0 {
			t.Fatalf("step %d at %.0f proposed %d exits, want 0 — the floor-only copy never takes profit",
				i, step.price, cycle.Proposed)
		}
		state, err := h.journal.ExitState(ctx, fold.Position.ID)
		if err != nil {
			t.Fatal(err)
		}
		if state.Baseline != step.baseline {
			t.Fatalf("step %d at %.0f baseline = %s, want %s (raised on this very observation, no grace)",
				i, step.price, state.Baseline, step.baseline)
		}
	}
	if n := len(h.submit.places); n != 0 {
		t.Fatalf("submitted %d orders, want 0", n)
	}
}
