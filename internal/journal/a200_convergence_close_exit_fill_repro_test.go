package journal

// a200_convergence_close_exit_fill_repro_test.go 는 a200 1단계(1.4·1.6)의 재현 시험임 — 편집 전 코드의 **현재 동작**을 기록함.
//
// 수렴 조정(position.Classify(LocalInstance) 의 kind, 수량 0)이 엔진 관리 인스턴스를 먼저 닫은 **뒤** 그 인스턴스의 청산 발의
// 주문(exit_states.pending_intent_id 에 붙은 intent)의 매도 체결이 도착하는 순서를 임시 원장과 운영 배선과 같은 훅 리터럴
// (app/engine bindApplyHooks: Project · Campaign · Exit · Costs)로 만듦. 조정 요청은 reconcile.Converger.ConvergeQuantities
// (converge.go:209-227)가 보내는 값과 같은 모양으로 직접 씀 — journal 패키지 내부 시험은 reconcile 을 들일 수 없어서임
// (실제 Converger · Tracker 로 도는 창 재현은 internal/reconcile/a200_convergence_close_exit_fill_windows_test.go).
//
// 이름의 「CurrentBehaviour」 는 이 시험이 지금 동작을 못 박는다는 뜻임 — a200 구현 단계(2.2)가 인정 갈래에서 RED 로 뒤집음.
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

// TestA200CurrentBehaviourAConvergenceCloseThenTheStopFillIsAnAttributionFailure 는 본 사고 모양(W4 · W5 의 원장 결과):
// 엔진 진입 · 관리 포지션 31주, 손절 발의 무장 + 매도 주문 붙임, 수렴 조정 31→0 커밋, 그 뒤 같은 초에 그 매도 31주 종결 체결.
func TestA200CurrentBehaviourAConvergenceCloseThenTheStopFillIsAnAttributionFailure(t *testing.T) {
	j, clk := a200Journal(t)
	entry, _ := openedPosition(t, j, "31")
	p := currentPosition(t, j, entry)
	sell := a200ArmStop(t, j, p.ID, order{intentID: "i-stop", attemptID: "a-stop", orderID: "o-stop",
		side: "SELL", quantity: "31"})

	clk.Advance(time.Second)
	result := a200Converge(t, j, "005930", "31", "0", clk.Now())
	if !result.ClosedExitState || result.Position.State != PositionClosed {
		t.Fatalf("수렴 조정 결과 = %+v, 관리 인스턴스를 닫아야 함", result)
	}
	afterAdjustment := readA200Ledger(t, j, p.ID)
	afterAdjustment.log(t, "조정 직후")

	// 같은 초(시계를 움직이지 않음)에 그 매도의 종결 체결.
	if _, err := j.RecordFill(context.Background(), terminalFill(sell, "31", "67900")); err != nil {
		t.Fatalf("RecordFill(stop): %v", err)
	}
	got := readA200Ledger(t, j, p.ID)
	got.log(t, "체결 뒤")

	if len(got.Reconcile) != 1 || !strings.HasPrefix(got.Reconcile[0], ReconcileCauseAttributionFailed+"|") ||
		!strings.Contains(got.Reconcile[0], "X59 (SELL_ON_CLOSED)") || !strings.HasSuffix(got.Reconcile[0], "|active") {
		t.Fatalf("reconcile_states = %q, 현재 동작은 ATTRIBUTION_FAILED(X59 SELL_ON_CLOSED) 활성 행 1", got.Reconcile)
	}
	if got.State != PositionClosed || got.Quantity != "0" {
		t.Fatalf("투영 = %s/%s, 현재 동작은 CLOSED/0 그대로", got.State, got.Quantity)
	}
	if countOf(got.ExitActions, ExitEventAdjustmentClosed) != 1 || countOf(got.ExitActions, ExitEventCompleted) != 0 {
		t.Fatalf("exit_events = %q, 현재 동작은 ADJUSTMENT_CLOSED 1 · COMPLETED 0", got.ExitActions)
	}
	if countOf(got.ExitActions, ExitEventProposalFilled) != 1 || got.PendingID != "" {
		t.Fatalf("exit_events = %q pending=%q, 현재 동작은 PROPOSAL_FILLED 1(종결 체결이 발의를 해소)", got.ExitActions, got.PendingID)
	}
	if got.Outcomes != 0 {
		t.Fatalf("trade_outcomes = %d, 현재 동작은 0", got.Outcomes)
	}
	if len(got.Adjustments) != 1 || !strings.HasPrefix(got.Adjustments[0], AdjustmentUnknown+"|31→0|") {
		t.Fatalf("adjustments = %q, 수렴 조정 UNKNOWN 31→0 하나", got.Adjustments)
	}
	if adjAt, fillAt := strings.Split(got.Adjustments[0], "|")[2], got.Fills[len(got.Fills)-1]; !strings.HasSuffix(fillAt, "|"+adjAt) {
		t.Logf("조정 시각 %s 와 체결 커밋 시각이 다른 초임: %s", adjAt, fillAt)
	} else {
		t.Logf("조정 created_at 과 체결 committed_at 이 같은 초(%s) — D1 의 「시각 비교 불가」 모양", adjAt)
	}

	// D5 실현 가능성 탐침: 지금 이 원장에서 동결하면 무엇이 나오는가(엔진 진입 포지션).
	outcome, reason := probeFreeze(t, j, p.ID)
	t.Logf("[성과 탐침 · 엔진 진입] reason=%q outcome=%+v", reason, outcome)
}

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

// TestA200CurrentBehaviourAnAdoptedPositionsOutcomeIsRefusedAsAPostAdoptionDisposal 은 편입 포지션의 같은 모양 —
// D5(편입 · 엔진 진입 둘 다 성과 동결)의 실현 가능성을 탐침으로 잼.
func TestA200CurrentBehaviourAnAdoptedPositionsOutcomeIsRefusedAsAPostAdoptionDisposal(t *testing.T) {
	j, clk := a200Journal(t)
	ctx := context.Background()
	adoptable(t, j, "p-ext")
	if _, err := j.AdoptPosition(ctx, sampleAdoption("p-ext")); err != nil {
		t.Fatal(err)
	}
	if _, err := j.OpenAdoptedExitState(ctx, "p-ext"); err != nil {
		t.Fatal(err)
	}
	clk.Advance(time.Second)
	sell := a200ArmStop(t, j, "p-ext", order{intentID: "i-stop", attemptID: "a-stop", orderID: "o-stop",
		side: "SELL", quantity: "10"})
	clk.Advance(time.Second)
	a200Converge(t, j, "005930", "10", "0", clk.Now())
	if _, err := j.RecordFill(ctx, terminalFill(sell, "10", "67900")); err != nil {
		t.Fatalf("RecordFill(stop): %v", err)
	}
	got := readA200Ledger(t, j, "p-ext")
	got.log(t, "편입 W4")
	if len(got.Reconcile) != 1 || !strings.HasPrefix(got.Reconcile[0], ReconcileCauseAttributionFailed+"|") {
		t.Fatalf("reconcile_states = %q, 편입 포지션도 현재 동작은 ATTRIBUTION_FAILED", got.Reconcile)
	}
	outcome, reason := probeFreeze(t, j, "p-ext")
	t.Logf("[성과 탐침 · 편입] reason=%q outcome=%+v", reason, outcome)
	if reason != OutcomeRefusalPostAdoptionDisposal {
		t.Fatalf("편입 포지션 성과 탐침 사유 = %q, 현재 산식은 수렴 조정(편입 뒤 감소)을 %q 로 거절함", reason,
			OutcomeRefusalPostAdoptionDisposal)
	}
}

// TestA200CurrentBehaviourPartialStopFillsAfterAConvergenceClose 는 D4 의 부분 체결 모양: 31주 매도가 20(Working) 뒤 31(Done)로
// 옴. 첫 거절이 행을 쓰고 둘째는 활성 행 때문에 아무것도 안 씀(F10).
func TestA200CurrentBehaviourPartialStopFillsAfterAConvergenceClose(t *testing.T) {
	j, clk := a200Journal(t)
	entry, _ := openedPosition(t, j, "31")
	p := currentPosition(t, j, entry)
	sell := a200ArmStop(t, j, p.ID, order{intentID: "i-stop", attemptID: "a-stop", orderID: "o-stop",
		side: "SELL", quantity: "31"})
	clk.Advance(time.Second)
	a200Converge(t, j, "005930", "31", "0", clk.Now())

	if _, err := j.RecordFill(context.Background(), fillOf(sell, "20", "67900")); err != nil {
		t.Fatalf("RecordFill(working): %v", err)
	}
	first := readA200Ledger(t, j, p.ID)
	first.log(t, "부분 체결 1(Working)")
	if first.PendingID != "i-stop" {
		t.Fatalf("Working 체결 뒤 pending=%q, 발의는 종결 체결 전까지 살아 있어야 함(D4 전제)", first.PendingID)
	}
	clk.Advance(time.Second)
	if _, err := j.RecordFill(context.Background(), terminalFill(sell, "31", "67900")); err != nil {
		t.Fatalf("RecordFill(done): %v", err)
	}
	second := readA200Ledger(t, j, p.ID)
	second.log(t, "부분 체결 2(Done)")
	if len(first.Reconcile) != 1 || !strings.Contains(first.Reconcile[0], "X58 (SELL_ON_CLOSED)") {
		t.Fatalf("첫 체결 reconcile_states = %q, 현재 동작은 X58 SELL_ON_CLOSED 행 1", first.Reconcile)
	}
	if len(second.Reconcile) != 1 {
		t.Fatalf("둘째 체결 뒤 reconcile_states = %q, 활성 행이 있으면 더 쓰지 않음(F10)", second.Reconcile)
	}
	if len(second.Fills) != 3 {
		t.Fatalf("fill_events = %q, 매수 1 + 거절된 매도 2 가 모두 기록됨(F11)", second.Fills)
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
