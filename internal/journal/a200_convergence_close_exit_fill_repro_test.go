package journal

// a200_convergence_close_exit_fill_repro_test.go 는 a200 1단계(1.4·1.6)의 재현 시험임 — 편집 전 코드의 **현재 동작**을 기록함.
//
// 수렴 조정(position.Classify(LocalInstance) 의 kind, 수량 0)이 엔진 관리 인스턴스를 먼저 닫은 **뒤** 그 인스턴스의 청산 발의
// 주문(exit_states.pending_intent_id 에 붙은 intent)의 매도 체결이 도착하는 순서를 임시 원장과 운영 배선과 같은 훅 리터럴
// (app/engine bindApplyHooks: Project · Campaign · Exit · Costs)로 만듦. 조정 요청은 reconcile.Converger.ConvergeQuantities
// (converge.go:209-227)가 보내는 값과 같은 모양으로 직접 씀 — journal 패키지 내부 시험은 reconcile 을 들일 수 없어서임
// (실제 Converger · Tracker 로 도는 창 재현은 internal/reconcile/a200_convergence_close_exit_fill_cycles_test.go).
//
// 이름의 「CurrentBehaviour」 는 편집 전부터 같은 동작을 뜻함 — 이 파일에 남은 시험(W0 · 조정을 사이에 둔 큰 주문 · 재매수 둘 ·
// 체결로 닫힌 뒤 둘째 매도)은 a200 뒤에도 그대로 맞는 계약임. 뒤집힌 셋은 a200_recognition_test.go 로 옮겨 갔음.
// 픽스처는 합성 값만 씀(계좌 · 종목 · 주문 번호 · 가격 모두 지어낸 값).

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/position"
)

// a200Journal 은 운영 배선과 같은 훅 리터럴을 묶은 임시 원장과 그 가짜 시계를 돌려줌(OutcomeRefused 싱크만 없음 — 시험 배선과 같음).
func a200Journal(t *testing.T) (*Journal, *clock.Fake) {
	t.Helper()
	j := openTestJournal(t)
	if err := j.SetApplyHooks(ApplyHooks{
		Project: ProjectPosition, Campaign: ApplyPositionCampaignFill, Exit: ApplyExitFill,
		Costs: costs.DefaultModel(),
	}); err != nil {
		t.Fatalf("SetApplyHooks: %v", err)
	}
	clk, ok := j.clk.(*clock.Fake)
	if !ok {
		t.Fatalf("openTestJournal 의 시계가 가짜 시계가 아님: %T", j.clk)
	}
	return j, clk
}

// a200ArmStop 은 손절 발의를 무장하고 그 매도 주문(접수 확정)을 그 발의에 붙임 — ExitObserver.submit 의 arm → attach → submit 순서.
func a200ArmStop(t *testing.T, j *Journal, positionID string, sell order) order {
	t.Helper()
	if err := j.RecordExitJudgement(context.Background(), ExitJudgement{
		PositionID: positionID, ObservedPrice: "67900", HighWater: "70000",
		Baseline: "68000", RatchetLevel: RatchetNone, ActiveRung: exitpolicy.NoRung,
		Proposal: &ExitProposal{Action: string(exitpolicy.ActionBaselineBreach), Level: RatchetNone},
	}); err != nil {
		t.Fatalf("손절 발의 무장: %v", err)
	}
	sell = place(t, j, sell)
	if err := j.AttachExitIntent(context.Background(), positionID, sell.intentID); err != nil {
		t.Fatalf("AttachExitIntent: %v", err)
	}
	return sell
}

// a200Converge 는 Converger 가 보내는 것과 같은 모양의 조정(수량 → newQuantity)을 씀.
func a200Converge(t *testing.T, j *Journal, symbol, local, newQuantity string, asOf time.Time) AdjustmentResult {
	t.Helper()
	ctx := context.Background()
	watermark, err := j.FillWatermark(ctx, symbol)
	if err != nil {
		t.Fatal(err)
	}
	result, err := j.ApplyPositionAdjustment(ctx, AdjustmentRequest{
		AccountRef: "acct-1", Market: "kr", Symbol: symbol,
		Kind:                  string(position.Classify(position.ProvenanceInputs{LocalInstance: true})),
		ExpectedPrevQuantity:  local,
		ExpectedFillWatermark: watermark,
		NewQuantity:           newQuantity,
		BrokerAsOf:            asOf.UTC().Format(time.RFC3339),
		Evidence:              "the account reports " + newQuantity + " of " + symbol + " and the projection held " + local,
	})
	if err != nil {
		t.Fatalf("수렴 조정: %v", err)
	}
	return result
}

// a200Ledger 는 한 인스턴스와 그 종목에 대해 원장에 남은 것을 읽음.
type a200Ledger struct {
	Reconcile   []string // "cause|evidence|active"
	State       string
	Quantity    string
	ClosedAt    string
	ExitActions []string
	Completed   bool
	PendingID   string
	Taken       string
	Outcomes    int
	Adjustments []string // "kind|prev→new|created_at"
	Fills       []string // "id|order|delta|cumulative|committed_at"
}

func readA200Ledger(t *testing.T, j *Journal, positionID string) a200Ledger {
	t.Helper()
	ctx := context.Background()
	var out a200Ledger
	rows, err := j.db.QueryContext(ctx, `SELECT cause, evidence, released_at IS NULL FROM reconcile_states
		WHERE account_ref = 'acct-1' AND symbol = '005930' ORDER BY rowid`)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var cause, evidence string
		var active bool
		if err := rows.Scan(&cause, &evidence, &active); err != nil {
			t.Fatal(err)
		}
		state := "released"
		if active {
			state = "active"
		}
		out.Reconcile = append(out.Reconcile, cause+"|"+evidence+"|"+state)
	}
	_ = rows.Close()
	if err := j.db.QueryRowContext(ctx, `SELECT state, quantity, coalesce(closed_at,'') FROM positions WHERE id = ?`,
		positionID).Scan(&out.State, &out.Quantity, &out.ClosedAt); err != nil {
		t.Fatal(err)
	}
	events, err := j.ExitEvents(ctx, positionID)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range events {
		out.ExitActions = append(out.ExitActions, e.Action)
	}
	state, err := j.ExitState(ctx, positionID)
	if err != nil {
		t.Fatal(err)
	}
	out.Completed, out.PendingID, out.Taken = state.Completed, state.PendingIntentID, state.TakenRatioTotal
	if err := j.db.QueryRowContext(ctx, `SELECT count(*) FROM trade_outcomes WHERE position_id = ?`,
		positionID).Scan(&out.Outcomes); err != nil {
		t.Fatal(err)
	}
	adjustments, err := j.PositionAdjustments(ctx, positionID)
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range adjustments {
		out.Adjustments = append(out.Adjustments, a.Kind+"|"+a.PrevQuantity+"→"+a.NewQuantity+"|"+a.CreatedAt)
	}
	fills, err := j.db.QueryContext(ctx, `SELECT id, order_id, delta_quantity, cumulative_quantity, committed_at
		FROM fill_events WHERE symbol = '005930' ORDER BY id`)
	if err != nil {
		t.Fatal(err)
	}
	for fills.Next() {
		var id int64
		var orderID, delta, cumulative, committed string
		if err := fills.Scan(&id, &orderID, &delta, &cumulative, &committed); err != nil {
			t.Fatal(err)
		}
		out.Fills = append(out.Fills, strings.Join([]string{strconv.FormatInt(id, 10), orderID, delta, cumulative, committed}, "|"))
	}
	_ = fills.Close()
	return out
}

func (l a200Ledger) log(t *testing.T, label string) {
	t.Helper()
	t.Logf("[%s] reconcile_states=%q", label, l.Reconcile)
	t.Logf("[%s] position state=%s quantity=%s closed_at=%s", label, l.State, l.Quantity, l.ClosedAt)
	t.Logf("[%s] exit_events=%q completed=%v pending_intent_id=%q taken_ratio_total=%s", label,
		l.ExitActions, l.Completed, l.PendingID, l.Taken)
	t.Logf("[%s] trade_outcomes=%d adjustments=%q", label, l.Outcomes, l.Adjustments)
	t.Logf("[%s] fill_events=%q", label, l.Fills)
}

func countOf(values []string, want string) int {
	n := 0
	for _, v := range values {
		if v == want {
			n++
		}
	}
	return n
}

// probeFreeze 는 「이 원장에서 지금 성과를 동결하면 무엇이 나오는가」를 읽기만으로 계산함(체결 트랜잭션 안의 freeze 와 같은 함수 ·
// 같은 nil 분봉). design D5 의 실현 가능성 증거임 — 원장에 쓰지 않음.
func probeFreeze(t *testing.T, j *Journal, positionID string) (TradeOutcome, TradeOutcomeRefusalReason) {
	t.Helper()
	return computeTradeOutcome(context.Background(), journalReader{j: j}, positionID, costs.DefaultModel(), nil)
}

// (a200 2.2) 1단계의 본 사고 모양 · 편입 · 부분 체결 「현재 동작」 시험 셋은 새 계약으로 뒤집혀
// a200_recognition_test.go 의 TestA200AStopFillAfterAConvergenceCloseIsRecognised · …AnAdoptedStopFillIsRecognisedWithoutAnOutcome ·
// …PartialStopFillsAfterAConvergenceCloseAreEachRecognised 로 옮겨 갔다. 편집 전 기대는 analysis/repro/w4.txt 원문에 남는다.

// TestA200CurrentBehaviourTheSameStopFilledBeforeAnyAdjustmentFreezesTheOutcome 은 같은 입력의 정상 순서(W0) — 성과 값 비교 기준.
func TestA200CurrentBehaviourTheSameStopFilledBeforeAnyAdjustmentFreezesTheOutcome(t *testing.T) {
	j, clk := a200Journal(t)
	entry, _ := openedPosition(t, j, "31")
	p := currentPosition(t, j, entry)
	sell := a200ArmStop(t, j, p.ID, order{intentID: "i-stop", attemptID: "a-stop", orderID: "o-stop",
		side: "SELL", quantity: "31"})
	clk.Advance(time.Second)
	if _, err := j.RecordFill(context.Background(), terminalFill(sell, "31", "67900")); err != nil {
		t.Fatalf("RecordFill(stop): %v", err)
	}
	got := readA200Ledger(t, j, p.ID)
	got.log(t, "W0")
	if len(got.Reconcile) != 0 || countOf(got.ExitActions, ExitEventCompleted) != 1 || got.Outcomes != 1 {
		t.Fatalf("W0 = %+v, 정상 청산은 차단 0 · COMPLETED 1 · 성과 1", got)
	}
	frozen := outcomeOf(t, j, p.ID)
	t.Logf("[W0 성과] %+v", frozen)
}

// TestA200CurrentBehaviourAStopFillStraddlingTheConvergenceIsAnAttributionFailure 는 design §2 가 이름으로 적은 「이 change 가
// 거절하는 정상 입력」: 손절 31 의 부분 체결 10 이 조정 **전에** 투영에 반영(31→21)되고, 수렴이 21→0 으로 닫은 뒤 나머지 21 이 옴.
// 상태 없는 판정 (라)는 그 intent 주문의 누적 합 31 을 감소분 21 과 견주므로 인정하지 않음 — 구현 뒤에도 ATTRIBUTION_FAILED 로 남아야 함.
func TestA200CurrentBehaviourAStopFillStraddlingTheConvergenceIsAnAttributionFailure(t *testing.T) {
	j, clk := a200Journal(t)
	entry, _ := openedPosition(t, j, "31")
	p := currentPosition(t, j, entry)
	sell := a200ArmStop(t, j, p.ID, order{intentID: "i-stop", attemptID: "a-stop", orderID: "o-stop",
		side: "SELL", quantity: "31"})
	clk.Advance(time.Second)
	if _, err := j.RecordFill(context.Background(), fillOf(sell, "10", "67900")); err != nil {
		t.Fatalf("RecordFill(partial before the adjustment): %v", err)
	}
	clk.Advance(time.Second)
	a200Converge(t, j, "005930", "21", "0", clk.Now())
	if _, err := j.RecordFill(context.Background(), terminalFill(sell, "31", "67900")); err != nil {
		t.Fatalf("RecordFill(rest after the adjustment): %v", err)
	}
	got := readA200Ledger(t, j, p.ID)
	got.log(t, "조정을 사이에 둔 부분 체결")
	if len(got.Reconcile) != 1 || !strings.Contains(got.Reconcile[0], "X59 (SELL_ON_CLOSED)") ||
		!strings.HasPrefix(got.Reconcile[0], ReconcileCauseAttributionFailed+"|") {
		t.Fatalf("reconcile_states = %q, 현재 동작은 ATTRIBUTION_FAILED(X59) 1", got.Reconcile)
	}
	if len(got.Adjustments) != 1 || !strings.HasPrefix(got.Adjustments[0], AdjustmentUnknown+"|21→0|") {
		t.Fatalf("adjustments = %q, 수렴 조정 21→0 하나", got.Adjustments)
	}
}

// TestA200CurrentBehaviourAStopFillAfterAReBuyLandsOnTheNewInstance 는 전제 (f) — 체결 투영은 주문이 결속된 인스턴스가 아니라
// (계좌 · 시장 · 종목)의 최신 인스턴스(currentInstance)를 읽는가. 수렴이 31→0 으로 닫은 뒤 사람의 재매수 5주가 새 EXTERNAL
// 인스턴스로 접히고, 그 다음에 옛 손절 매도의 체결이 옴.
func TestA200CurrentBehaviourAStopFillAfterAReBuyLandsOnTheNewInstance(t *testing.T) {
	j, clk := a200Journal(t)
	ctx := context.Background()
	entry, _ := openedPosition(t, j, "31")
	old := currentPosition(t, j, entry)
	sell := a200ArmStop(t, j, old.ID, order{intentID: "i-stop", attemptID: "a-stop", orderID: "o-stop",
		side: "SELL", quantity: "31"})
	clk.Advance(time.Second)
	a200Converge(t, j, "005930", "31", "0", clk.Now())

	// 외부 보유 접기(external.go Ingest 가 보내는 EXTERNAL 조정 — 닫힌 인스턴스 위에 새 인스턴스를 엶).
	clk.Advance(time.Second)
	watermark, err := j.FillWatermark(ctx, "005930")
	if err != nil {
		t.Fatal(err)
	}
	folded, err := j.ApplyPositionAdjustment(ctx, AdjustmentRequest{
		AccountRef: "acct-1", Market: "kr", Symbol: "005930", Kind: AdjustmentExternal,
		ExpectedPrevQuantity: "0", ExpectedFillWatermark: watermark, NewQuantity: "5",
		BrokerAsOf: clk.Now().UTC().Format(time.RFC3339), Evidence: "a holding with no local instance (synthetic)",
	})
	if err != nil {
		t.Fatalf("외부 보유 접기: %v", err)
	}
	if folded.Position.ID == old.ID {
		t.Fatalf("접기가 닫힌 인스턴스를 다시 열었음: %+v", folded.Position)
	}

	if _, err := j.RecordFill(ctx, terminalFill(sell, "31", "67900")); err != nil {
		t.Fatalf("RecordFill(stop): %v", err)
	}
	oldAfter := readA200Ledger(t, j, old.ID)
	oldAfter.log(t, "옛 인스턴스")
	var newState, newQuantity string
	if err := j.db.QueryRowContext(ctx, `SELECT state, quantity FROM positions WHERE id = ?`, folded.Position.ID).
		Scan(&newState, &newQuantity); err != nil {
		t.Fatal(err)
	}
	t.Logf("[새 인스턴스] id=%s seq=%d state=%s quantity=%s", folded.Position.ID, folded.Position.InstanceSeq, newState, newQuantity)
	// 현재 동작: 매도 31 이 새 인스턴스(5)에 닿아 OVERSHOOTS → OVERSELL → QUANTITY_MISMATCH. SELL_ON_CLOSED 가 아님.
	if len(oldAfter.Reconcile) != 1 || !strings.HasPrefix(oldAfter.Reconcile[0], ReconcileCauseQuantityMismatch+"|") ||
		strings.Contains(oldAfter.Reconcile[0], "SELL_ON_CLOSED") {
		t.Fatalf("reconcile_states = %q, 현재 동작은 새 인스턴스 기준 OVERSELL → QUANTITY_MISMATCH", oldAfter.Reconcile)
	}
}

// TestA200CurrentBehaviourAPartialStopFillAfterAReBuyReducesTheNewInstance 는 (f) 의 둘째 모양: 옛 매도의 부분 체결 3 이
// 새 외부 인스턴스(5)보다 작으면 거절도 차단도 없이 새 인스턴스를 줄임.
func TestA200CurrentBehaviourAPartialStopFillAfterAReBuyReducesTheNewInstance(t *testing.T) {
	j, clk := a200Journal(t)
	ctx := context.Background()
	entry, _ := openedPosition(t, j, "31")
	old := currentPosition(t, j, entry)
	sell := a200ArmStop(t, j, old.ID, order{intentID: "i-stop", attemptID: "a-stop", orderID: "o-stop",
		side: "SELL", quantity: "31"})
	clk.Advance(time.Second)
	a200Converge(t, j, "005930", "31", "0", clk.Now())
	clk.Advance(time.Second)
	watermark, err := j.FillWatermark(ctx, "005930")
	if err != nil {
		t.Fatal(err)
	}
	folded, err := j.ApplyPositionAdjustment(ctx, AdjustmentRequest{
		AccountRef: "acct-1", Market: "kr", Symbol: "005930", Kind: AdjustmentExternal,
		ExpectedPrevQuantity: "0", ExpectedFillWatermark: watermark, NewQuantity: "5",
		BrokerAsOf: clk.Now().UTC().Format(time.RFC3339), Evidence: "a holding with no local instance (synthetic)",
	})
	if err != nil {
		t.Fatalf("외부 보유 접기: %v", err)
	}
	if _, err := j.RecordFill(ctx, fillOf(sell, "3", "67900")); err != nil {
		t.Fatalf("RecordFill(stop partial): %v", err)
	}
	oldAfter := readA200Ledger(t, j, old.ID)
	oldAfter.log(t, "옛 인스턴스")
	var newState, newQuantity string
	if err := j.db.QueryRowContext(ctx, `SELECT state, quantity FROM positions WHERE id = ?`, folded.Position.ID).
		Scan(&newState, &newQuantity); err != nil {
		t.Fatal(err)
	}
	t.Logf("[새 인스턴스] state=%s quantity=%s", newState, newQuantity)
	if len(oldAfter.Reconcile) != 0 || newQuantity != "2" {
		t.Fatalf("reconcile=%q 새 인스턴스 수량=%s, 현재 동작은 차단 없이 새 외부 인스턴스를 5→2 로 줄임",
			oldAfter.Reconcile, newQuantity)
	}
}

// TestA200CurrentBehaviourASecondEngineSellOnAFillClosedInstance 는 design §2 「계속 거절 2」 의 모양: 조정 없이 체결로 닫힌
// 인스턴스에 둘째 엔진 매도 체결 — 같은 SELL_ON_CLOSED 행이 생기고 조정 행은 0.
func TestA200CurrentBehaviourASecondEngineSellOnAFillClosedInstance(t *testing.T) {
	j, clk := a200Journal(t)
	entry, _ := openedPosition(t, j, "31")
	p := currentPosition(t, j, entry)
	first := a200ArmStop(t, j, p.ID, order{intentID: "i-stop", attemptID: "a-stop", orderID: "o-stop",
		side: "SELL", quantity: "31"})
	second := place(t, j, order{intentID: "i-dup", attemptID: "a-dup", orderID: "o-dup", side: "SELL", quantity: "31"})
	clk.Advance(time.Second)
	if _, err := j.RecordFill(context.Background(), terminalFill(first, "31", "67900")); err != nil {
		t.Fatal(err)
	}
	if _, err := j.RecordFill(context.Background(), terminalFill(second, "31", "67900")); err != nil {
		t.Fatal(err)
	}
	got := readA200Ledger(t, j, p.ID)
	got.log(t, "체결로 닫힌 뒤 둘째 매도")
	if len(got.Reconcile) != 1 || !strings.Contains(got.Reconcile[0], "SELL_ON_CLOSED") || len(got.Adjustments) != 0 {
		t.Fatalf("reconcile=%q adjustments=%q, 현재 동작은 SELL_ON_CLOSED 행 1 · 조정 0", got.Reconcile, got.Adjustments)
	}
}
