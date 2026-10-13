package journal

// a130_adopted_outcome_test.go 는 a130(4판)의 원장 쪽 시험임 — 묶음별 편입가(브로커 1분봉)와 편입 시기 · 인스턴스 매도 거절(design D1),
// 사유 코드 타입(D2), 커밋 뒤에만 남는 거절 기록(D3).
//
// 사건: 편입 3주 → 63초 뒤 대사 조정 3→6 → 손절 6주 매도로 닫힌 포지션이 성과 행도 흔적도 남기지 않았음. 4판(사용자 결정 2026-10-13
// 3차): 체결 트랜잭션 안에서는 네트워크를 부르지 않으므로 needs_intake_price 로 거절하고, backfill 이 그 분의 1분봉 종가로 매입 처리 —
// 분자 · 분모는 같은 6주 기준.

import (
	"context"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"math/big"
	"os/exec"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
)

// a130Refusals 는 거절 싱크가 받은 것을 모음.
type a130Refusals struct {
	mu   sync.Mutex
	seen []TradeOutcomeRefusal
}

func (r *a130Refusals) sink(x TradeOutcomeRefusal) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seen = append(r.seen, x)
}

func (r *a130Refusals) all() []TradeOutcomeRefusal {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]TradeOutcomeRefusal(nil), r.seen...)
}

// a130Journal 은 생산과 같은 훅(Project · Exit · 비용 모델)에 거절 싱크를 묶은 저널임.
func a130Journal(t *testing.T, refusals *a130Refusals) (*Journal, *clock.Fake, string) {
	t.Helper()
	path := t.TempDir() + "/journal.db"
	j := openTestJournalAt(t, path)
	hooks := ApplyHooks{Project: ProjectPosition, Exit: ApplyExitFill, Costs: costs.DefaultModel()}
	if refusals != nil {
		hooks.OutcomeRefused = refusals.sink
	}
	if err := j.SetApplyHooks(hooks); err != nil {
		t.Fatalf("SetApplyHooks: %v", err)
	}
	fake, ok := j.clk.(*clock.Fake)
	if !ok {
		t.Fatal("the test journal must run on the fake clock")
	}
	return j, fake, path
}

// a130Adjust 는 대사 조정 한 건을 그 순간의 시계로 기록함.
func a130Adjust(t *testing.T, j *Journal, account, kind, prev, next, avg string) {
	t.Helper()
	ctx := context.Background()
	watermark, err := j.FillWatermark(ctx, "005930")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := j.ApplyPositionAdjustment(ctx, AdjustmentRequest{
		AccountRef: account, Market: "kr", Symbol: "005930", Kind: kind,
		ExpectedPrevQuantity: prev, ExpectedFillWatermark: watermark,
		NewQuantity: next, NewAvgPrice: avg,
		BrokerAsOf: j.nowString(), Evidence: "a130 test adjustment",
	}); err != nil {
		t.Fatalf("adjust %s→%s: %v", prev, next, err)
	}
}

// a130Adopt 는 지금 프로젝션의 수량으로 편입하고 exit 상태를 엶. observedAt 은 지금 시계.
func a130Adopt(t *testing.T, j *Journal, account, quantity, observed, stop string) string {
	t.Helper()
	ctx := context.Background()
	p, err := j.CurrentPosition(ctx, account, "kr", "005930")
	if err != nil {
		t.Fatalf("CurrentPosition: %v", err)
	}
	if _, err := j.AdoptPosition(ctx, AdoptionRequest{
		PositionID: p.ID, Symbol: "005930", Market: "kr",
		Quantity: quantity, CostBasis: "55000",
		ObservedPrice: observed, SyntheticStop: stop,
		ObservedAt: j.nowString(),
	}); err != nil {
		t.Fatalf("AdoptPosition: %v", err)
	}
	if _, err := j.OpenAdoptedExitState(ctx, p.ID); err != nil {
		t.Fatalf("OpenAdoptedExitState: %v", err)
	}
	return p.ID
}

// a130Sell 는 exit 루프가 낸 매도를 계좌를 지정해 체결시킴(exitLoopSell 과 같되 계좌 센티널을 씀).
func a130Sell(t *testing.T, j *Journal, account, positionID, tag, quantity, price string) {
	t.Helper()
	ctx := context.Background()
	sell := place(t, j, order{
		intentID: "i-" + tag, attemptID: "a-" + tag, orderID: "o-" + tag,
		decisionID: "d-" + tag, side: "SELL", quantity: quantity, account: account,
	})
	state, err := j.ExitState(ctx, positionID)
	if err != nil {
		t.Fatal(err)
	}
	if err := j.RecordExitJudgement(ctx, ExitJudgement{
		PositionID: positionID, ObservedPrice: price,
		HighWater: state.HighWater, Baseline: state.Baseline,
		RatchetLevel: state.RatchetLevel, ActiveRung: exitpolicy.NoRung,
		Proposal: &ExitProposal{Action: string(exitpolicy.ActionBaselineBreach), Level: RatchetNone, IntentID: "i-" + tag},
	}); err != nil {
		t.Fatalf("RecordExitJudgement: %v", err)
	}
	if _, err := j.RecordFill(ctx, terminalFill(sell, quantity, price)); err != nil {
		t.Fatalf("RecordFill(%s): %v", tag, err)
	}
}

// a130ExpectedR 는 design D1 의 산식을 시험 안에서 독립으로 계산함 — 같은 비용 모델, 같은 12자리 표기.
func a130ExpectedR(t *testing.T, buyNotional, sellNotional, risk, quantity string) (pnl, r string) {
	t.Helper()
	model := costs.DefaultModel()
	buyCost, err := model.EstimateTradeCost(buyNotional, costs.SideBuy, costs.Market("kr"))
	if err != nil {
		t.Fatal(err)
	}
	sellCost, err := model.EstimateTradeCost(sellNotional, costs.SideSell, costs.Market("kr"))
	if err != nil {
		t.Fatal(err)
	}
	p := new(big.Rat).Sub(ratOf(sellNotional), ratOf(buyNotional))
	p.Sub(p, ratOf(buyCost.Total))
	p.Sub(p, ratOf(sellCost.Total))
	den := new(big.Rat).Mul(ratOf(risk), ratOf(quantity))
	return ratText(p), ratText(new(big.Rat).Quo(p, den))
}

func a130Only(t *testing.T, refusals *a130Refusals, want TradeOutcomeRefusalReason) {
	t.Helper()
	seen := refusals.all()
	if len(seen) != 1 || seen[0].Reason != want {
		t.Fatalf("refusals = %+v, want exactly one %s", seen, want)
	}
}

func a130NoRow(t *testing.T, j *Journal, id string) {
	t.Helper()
	if _, err := j.TradeOutcomeOf(context.Background(), id); !errors.Is(err, ErrTradeOutcomeNotFound) {
		t.Fatalf("TradeOutcomeOf: %v, want no row", err)
	}
}

func a130Silent(t *testing.T, refusals *a130Refusals) {
	t.Helper()
	if n := len(refusals.all()); n != 0 {
		t.Errorf("a frozen close reported %d refusals, want 0", n)
	}
}

// a130Grown 은 사건의 모양을 만듦: 편입 3 @70000(위험 3500) → 63초 뒤 대사 3→6 → 1분 뒤 엔진 6주 매도 @72000.
func a130Grown(t *testing.T, j *Journal, fake *clock.Fake) string {
	t.Helper()
	return a130GrownTagged(t, j, fake, "stop")
}

// a130GrownTagged 는 a130Grown 과 같되 매도 intent 의 이름표를 받음(한 저널에 여럿을 만들 때).
func a130GrownTagged(t *testing.T, j *Journal, fake *clock.Fake, tag string) string {
	t.Helper()
	a130Adjust(t, j, "acct-1", AdjustmentExternal, "0", "3", "70100")
	id := a130Adopt(t, j, "acct-1", "3", "70000", "66500")
	fake.Advance(63 * time.Second)
	a130Adjust(t, j, "acct-1", AdjustmentUnknown, "3", "6", "")
	fake.Advance(time.Minute)
	a130Sell(t, j, "acct-1", id, tag, "6", "72000")
	return id
}

// a130Prices 는 이 포지션의 늘어난 묶음마다 그 분의 봉(이름 = BarLabel)을 close 로 매긴 분봉 집합을 만듦 — 도구가 브로커에서 읽는 것의 대역.
func a130Prices(t *testing.T, j *Journal, id, closePrice string) IntakePrices {
	t.Helper()
	lots, reason := adoptionEpoch(context.Background(), journalReader{j: j}, id)
	if reason != "" {
		t.Fatalf("adoptionEpoch: %s", reason)
	}
	prices := IntakePrices{}
	for _, lot := range lots {
		prices[lot.Key()] = IntakeCandle{Label: lot.BarLabel(), Close: closePrice, Source: IntakeSourceBrokerCandle1m}
	}
	return prices
}

// TestA130AGrownPositionIsRefusedInsideTheFillTransaction: 체결 트랜잭션 안에서는 분봉을 읽지 않음 — 행 없음, 커밋 뒤 거절 한 번
// {종목, 시장, 회차, needs_intake_price}, 계좌 참조 · 포지션 id 없음.
func TestA130AGrownPositionIsRefusedInsideTheFillTransaction(t *testing.T) {
	refusals := &a130Refusals{}
	j, fake, _ := a130Journal(t, refusals)
	id := a130Grown(t, j, fake)

	a130NoRow(t, j, id)
	seen := refusals.all()
	want := TradeOutcomeRefusal{Symbol: "005930", Market: "kr", InstanceSeq: 1, Reason: OutcomeRefusalNeedsIntakePrice}
	if len(seen) != 1 || seen[0] != want {
		t.Fatalf("refusals = %+v, want exactly %+v", seen, want)
	}
	text := strings.ToUpper(strings.Join([]string{seen[0].Symbol, seen[0].Market, string(seen[0].Reason)}, "|"))
	if strings.Contains(text, "ACCT-1") || strings.Contains(text, strings.ToUpper(id)) {
		t.Errorf("the refusal carries the account or the position id: %q", text)
	}
}

// TestA130AGrownPositionPricesEachLotAtItsMinuteBar 는 사건의 재계산임: 매수 = 3 × 70000(편입 관측가) + 3 × 69000(조정이 든 1분봉 종가),
// 초기 수량 = 6, 실현 R = 손익 ÷ (3500 × 6) — 편입 수량 3 을 분모로 둔 2배 값도, 조정 행 평균가로 매긴 값도 아님.
func TestA130AGrownPositionPricesEachLotAtItsMinuteBar(t *testing.T) {
	j, fake, _ := a130Journal(t, nil)
	id := a130Grown(t, j, fake)

	got, reason := computeTradeOutcome(context.Background(), journalReader{j: j}, id, costs.DefaultModel(),
		a130Prices(t, j, id, "69000"))
	if reason != "" {
		t.Fatalf("reason = %s, want a priced outcome", reason)
	}
	wantPnL, wantR := a130ExpectedR(t, "417000", "432000", "3500", "6")
	if got.InitialQuantity != "6" || got.InitialRisk != "3500" {
		t.Errorf("quantity/risk = %s/%s, want 6/3500", got.InitialQuantity, got.InitialRisk)
	}
	if got.RealizedPnLAfterCosts != wantPnL || got.RealizedR != wantR {
		t.Errorf("pnl/R = %s/%s, want %s/%s", got.RealizedPnLAfterCosts, got.RealizedR, wantPnL, wantR)
	}
	if _, inflated := a130ExpectedR(t, "417000", "432000", "3500", "3"); got.RealizedR == inflated {
		t.Errorf("realised R = %s is the adoption-quantity denominator (2x inflated)", got.RealizedR)
	}
	if _, costBasis := a130ExpectedR(t, "420300", "432000", "3500", "6"); got.RealizedR == costBasis {
		t.Error("the added lot was priced at the carried-over average, not at the minute bar")
	}
}

// TestA130AMissingOrWrongMinuteBarIsRefused: 분봉 집합에 그 묶음이 없거나 이름이 그 분이 아니면 no_intake_candle(가격을 지어내지 않음).
func TestA130AMissingOrWrongMinuteBarIsRefused(t *testing.T) {
	j, fake, _ := a130Journal(t, nil)
	id := a130Grown(t, j, fake)
	ctx := context.Background()
	if _, reason := computeTradeOutcome(ctx, journalReader{j: j}, id, costs.DefaultModel(), IntakePrices{}); reason != OutcomeRefusalNoIntakeCandle {
		t.Errorf("empty prices: reason = %q, want no_intake_candle", reason)
	}
	wrong := a130Prices(t, j, id, "69000")
	for k, c := range wrong {
		c.Label = c.Label.Add(-time.Minute) // 앞 분의 봉
		wrong[k] = c
	}
	if _, reason := computeTradeOutcome(ctx, journalReader{j: j}, id, costs.DefaultModel(), wrong); reason != OutcomeRefusalNoIntakeCandle {
		t.Errorf("previous-minute bar: reason = %q, want no_intake_candle", reason)
	}
}

// TestA130TheBarLabelIsTheMinuteThatClosesAfterTheAdjustment: 공식 분봉의 이름은 봉 종료 시각이고 t 봉은 [t−1분, t) — 조정 00:18:54 는
// 00:19:00 봉, 정각 00:19:00 은 00:20:00 봉.
func TestA130TheBarLabelIsTheMinuteThatClosesAfterTheAdjustment(t *testing.T) {
	for _, c := range []struct{ at, want time.Time }{
		{time.Date(2026, 10, 7, 0, 18, 54, 0, time.UTC), time.Date(2026, 10, 7, 0, 19, 0, 0, time.UTC)},
		{time.Date(2026, 10, 7, 0, 19, 0, 0, time.UTC), time.Date(2026, 10, 7, 0, 20, 0, 0, time.UTC)},
		{time.Date(2026, 10, 7, 9, 18, 54, 500, time.FixedZone("KST", 9*3600)), time.Date(2026, 10, 7, 0, 19, 0, 0, time.UTC)},
	} {
		if got := (IntakeLot{CreatedAt: c.at}).BarLabel(); !got.Equal(c.want) {
			t.Errorf("BarLabel(%v) = %v, want %v", c.at, got, c.want)
		}
	}
}

// TestA130TheGrowThenShrinkRoundTripIsRefusedAsADisposal 는 이름 붙인 회귀(F-9)임: 편입 3 → +3 → −3 → 엔진 3주 매도. a130 이전엔
// 동결(매도 = 편입 수량)이었고 a130 뒤엔 편입 뒤 감소로 거절.
func TestA130TheGrowThenShrinkRoundTripIsRefusedAsADisposal(t *testing.T) {
	refusals := &a130Refusals{}
	j, fake, _ := a130Journal(t, refusals)

	a130Adjust(t, j, "acct-1", AdjustmentExternal, "0", "3", "70100")
	id := a130Adopt(t, j, "acct-1", "3", "70000", "66500")
	fake.Advance(time.Minute)
	a130Adjust(t, j, "acct-1", AdjustmentUnknown, "3", "6", "")
	fake.Advance(time.Minute)
	a130Adjust(t, j, "acct-1", AdjustmentUnknown, "6", "3", "")
	fake.Advance(time.Minute)
	a130Sell(t, j, "acct-1", id, "stop", "3", "72000")

	a130NoRow(t, j, id)
	a130Only(t, refusals, OutcomeRefusalPostAdoptionDisposal)
}

// TestA130APostAdoptionDisposalBlocksASweptFlatten 는 F-3 의 구멍임: 이전 인스턴스의 flatten 매도 2주가 계좌 · 심볼로 쓸려 들어오고,
// 편입 뒤 사람이 2주를 팔면 귀속 매도 = 3(엔진) + 2(이전 flatten) = 5 = 편입 수량 → a130 이전엔 조작된 행. 뒤엔 편입 뒤 감소로 거절.
func TestA130APostAdoptionDisposalBlocksASweptFlatten(t *testing.T) {
	refusals := &a130Refusals{}
	j, fake, _ := a130Journal(t, refusals)
	ctx := context.Background()

	// 이전 인스턴스: 엔진 진입 2주 → flatten 이 2주 청산.
	buy := place(t, j, order{intentID: "i-b0", attemptID: "a-b0", orderID: "o-b0", decisionID: "d-b0", side: "BUY", quantity: "2"})
	if _, err := j.RecordFill(ctx, terminalFill(buy, "2", "70000")); err != nil {
		t.Fatal(err)
	}
	flat := place(t, j, order{intentID: "i-f0", attemptID: "a-f0", orderID: "o-f0", decisionID: "d-f0", side: "SELL", quantity: "2"})
	flattenLiquidation(t, j, "saga-0", flat)
	if _, err := j.RecordFill(ctx, terminalFill(flat, "2", "71000")); err != nil {
		t.Fatal(err)
	}
	refusals.mu.Lock()
	refusals.seen = nil // 이전 인스턴스(exit 상태 없음)의 종결은 이 시험의 대상이 아님
	refusals.mu.Unlock()

	fake.Advance(time.Minute)
	a130Adjust(t, j, "acct-1", AdjustmentExternal, "0", "5", "70100")
	id := a130Adopt(t, j, "acct-1", "5", "70000", "66500")
	fake.Advance(time.Minute)
	a130Adjust(t, j, "acct-1", AdjustmentUnknown, "5", "3", "")
	fake.Advance(time.Minute)
	a130Sell(t, j, "acct-1", id, "stop", "3", "72000")

	a130NoRow(t, j, id)
	a130Only(t, refusals, OutcomeRefusalPostAdoptionDisposal)
}

// TestA130AReadoptedPositionIsRefused(I-2): 재편입은 exit 상태의 주당 위험을 다시 씀 — 편입 기록의 관측가와 시점이 갈리므로 거절.
func TestA130AReadoptedPositionIsRefused(t *testing.T) {
	refusals := &a130Refusals{}
	j, fake, _ := a130Journal(t, refusals)
	ctx := context.Background()

	a130Adjust(t, j, "acct-1", AdjustmentExternal, "0", "3", "70100")
	id := a130Adopt(t, j, "acct-1", "3", "70000", "66500")
	fake.Advance(time.Minute)
	release := positionpolicy.Request{
		PositionID: id, ExpectedGeneration: 1, ExpectedVersion: 0,
		Action: positionpolicy.ActionRelease, Actor: positionpolicy.ActorLocalOperator,
		Reason: positionpolicy.ReasonRelease, At: fake.Now(),
	}
	if _, err := j.ApplyPositionPolicy(ctx, release); err != nil {
		t.Fatalf("release: %v", err)
	}
	fake.Advance(time.Minute)
	readopt := positionpolicy.Request{
		PositionID: id, ExpectedGeneration: 1, ExpectedVersion: 1,
		Action: positionpolicy.ActionReadopt, Actor: positionpolicy.ActorLocalOperator,
		Reason: positionpolicy.ReasonReadopt, At: fake.Now(),
		ReAdoption: &positionpolicy.ReAdoptionObservation{
			ObservedPrice: "72000", SyntheticStop: "68400", ObservedAt: fake.Now().Format(time.RFC3339Nano),
		},
	}
	if _, err := j.ApplyPositionPolicy(ctx, readopt); err != nil {
		t.Fatalf("readopt: %v", err)
	}
	fake.Advance(time.Minute)
	a130Sell(t, j, "acct-1", id, "stop", "3", "71000")

	a130NoRow(t, j, id)
	a130Only(t, refusals, OutcomeRefusalReadopted)
}

// TestA130APreAdoptionAdjustmentStillFreezes: 편입 전 조정(1→3)은 편입 수량 3 에 들어 있음 — 거절하지 않고, 숫자는 a130 이전 산식 그대로.
func TestA130APreAdoptionAdjustmentStillFreezes(t *testing.T) {
	refusals := &a130Refusals{}
	j, fake, _ := a130Journal(t, refusals)

	a130Adjust(t, j, "acct-1", AdjustmentExternal, "0", "1", "70100")
	fake.Advance(time.Minute)
	a130Adjust(t, j, "acct-1", AdjustmentUnknown, "1", "3", "")
	fake.Advance(time.Minute)
	id := a130Adopt(t, j, "acct-1", "3", "70000", "66500")
	fake.Advance(time.Minute)
	a130Sell(t, j, "acct-1", id, "stop", "3", "72000")

	got, err := j.TradeOutcomeOf(context.Background(), id)
	if err != nil {
		t.Fatalf("TradeOutcomeOf: %v — a pre-adoption adjustment is part of the adopted quantity", err)
	}
	wantPnL, wantR := a130ExpectedR(t, "210000", "216000", "3500", "3")
	if got.InitialQuantity != "3" || got.RealizedPnLAfterCosts != wantPnL || got.RealizedR != wantR {
		t.Errorf("got qty=%s pnl=%s R=%s, want 3 / %s / %s", got.InitialQuantity, got.RealizedPnLAfterCosts,
			got.RealizedR, wantPnL, wantR)
	}
	if n := len(refusals.all()); n != 0 {
		t.Errorf("a frozen close reported %d refusals", n)
	}
}

// TestA130TheSameSecondBoundaryFailsClosed: 편입과 같은 초의 증가는 편입 뒤로 세지 않음(→ 매도 6 ≠ 편입 3, 수량 불일치 거절),
// 같은 초의 감소는 편입 뒤로 셈(→ 편입 뒤 감소 거절). 어느 쪽도 행을 만들지 않음.
func TestA130TheSameSecondBoundaryFailsClosed(t *testing.T) {
	t.Run("increase", func(t *testing.T) {
		refusals := &a130Refusals{}
		j, fake, _ := a130Journal(t, refusals)
		a130Adjust(t, j, "acct-1", AdjustmentExternal, "0", "3", "70100")
		id := a130Adopt(t, j, "acct-1", "3", "70000", "66500")
		a130Adjust(t, j, "acct-1", AdjustmentUnknown, "3", "6", "")
		fake.Advance(time.Minute)
		a130Sell(t, j, "acct-1", id, "stop", "6", "72000")
		a130NoRow(t, j, id)
		a130Only(t, refusals, OutcomeRefusalSellQuantityMismatch)
	})
	t.Run("decrease", func(t *testing.T) {
		refusals := &a130Refusals{}
		j, fake, _ := a130Journal(t, refusals)
		a130Adjust(t, j, "acct-1", AdjustmentExternal, "0", "5", "70100")
		id := a130Adopt(t, j, "acct-1", "5", "70000", "66500")
		a130Adjust(t, j, "acct-1", AdjustmentUnknown, "5", "3", "")
		fake.Advance(time.Minute)
		a130Sell(t, j, "acct-1", id, "stop", "3", "72000")
		a130NoRow(t, j, id)
		a130Only(t, refusals, OutcomeRefusalPostAdoptionDisposal)
	})
}

// TestA130ARolledBackCloseReportsNothing: 훅 뒤 트랜잭션이 실패하면(여기선 Exit 훅이 ApplyExitFill 뒤 오류) 종결이 커밋되지 않으므로
// 거절도 알리지 않음(F-8 — 거짓 줄 없음).
func TestA130ARolledBackCloseReportsNothing(t *testing.T) {
	refusals := &a130Refusals{}
	path := t.TempDir() + "/journal.db"
	j := openTestJournalAt(t, path)
	failAfter := false
	boom := errors.New("a later failure in the same transaction")
	if err := j.SetApplyHooks(ApplyHooks{
		Project: ProjectPosition, Costs: costs.DefaultModel(), OutcomeRefused: refusals.sink,
		Exit: func(ctx context.Context, tx *ApplyTx, fill AppliedFill) error {
			if err := ApplyExitFill(ctx, tx, fill); err != nil {
				return err
			}
			if failAfter {
				return boom
			}
			return nil
		},
	}); err != nil {
		t.Fatal(err)
	}
	fake := j.clk.(*clock.Fake)
	a130Adjust(t, j, "acct-1", AdjustmentExternal, "0", "3", "70100")
	id := a130Adopt(t, j, "acct-1", "3", "70000", "66500")
	fake.Advance(time.Minute)
	a130Adjust(t, j, "acct-1", AdjustmentUnknown, "3", "6", "")
	fake.Advance(time.Minute) // 늘어난 묶음 → 트랜잭션 안에서는 needs_intake_price

	failAfter = true
	ctx := context.Background()
	sell := place(t, j, order{intentID: "i-stop", attemptID: "a-stop", orderID: "o-stop", decisionID: "d-stop", side: "SELL", quantity: "6"})
	state, err := j.ExitState(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if err := j.RecordExitJudgement(ctx, ExitJudgement{
		PositionID: id, ObservedPrice: "72000", HighWater: state.HighWater, Baseline: state.Baseline,
		RatchetLevel: state.RatchetLevel, ActiveRung: exitpolicy.NoRung,
		Proposal: &ExitProposal{Action: string(exitpolicy.ActionBaselineBreach), Level: RatchetNone, IntentID: "i-stop"},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := j.RecordFill(ctx, terminalFill(sell, "6", "72000")); !errors.Is(err, boom) {
		t.Fatalf("RecordFill err = %v, want the injected failure", err)
	}
	if n := len(refusals.all()); n != 0 {
		t.Fatalf("a rolled-back close reported %d refusals, want 0", n)
	}
	if p, err := j.LookupPosition(ctx, id); err != nil || p.State == PositionClosed {
		t.Fatalf("position = %+v, %v — the close must have rolled back", p, err)
	}

	failAfter = false
	if _, err := j.RecordFill(ctx, terminalFill(sell, "6", "72000")); err != nil {
		t.Fatalf("RecordFill retry: %v", err)
	}
	a130Only(t, refusals, OutcomeRefusalNeedsIntakePrice)
}

// TestA130APanickingRefusalSinkDoesNotRollBackTheClose: 싱크가 패닉해도 체결 · CLOSED · exit 완료는 커밋됨.
func TestA130APanickingRefusalSinkDoesNotRollBackTheClose(t *testing.T) {
	path := t.TempDir() + "/journal.db"
	j := openTestJournalAt(t, path)
	if err := j.SetApplyHooks(ApplyHooks{
		Project: ProjectPosition, Exit: ApplyExitFill, Costs: costs.DefaultModel(),
		OutcomeRefused: func(TradeOutcomeRefusal) { panic("sink exploded") },
	}); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	fake := j.clk.(*clock.Fake)
	a130Adjust(t, j, "acct-1", AdjustmentExternal, "0", "3", "70100")
	id := a130Adopt(t, j, "acct-1", "3", "70000", "66500")
	fake.Advance(time.Minute)
	a130Adjust(t, j, "acct-1", AdjustmentUnknown, "3", "6", "")
	fake.Advance(time.Minute) // 늘어난 묶음 → 트랜잭션 안에서는 needs_intake_price
	a130Sell(t, j, "acct-1", id, "stop", "6", "72000")

	p, err := j.LookupPosition(ctx, id)
	if err != nil || p.State != PositionClosed {
		t.Fatalf("position = %+v, %v — a refusal sink may not roll back the fill", p, err)
	}
	state, err := j.ExitState(ctx, id)
	if err != nil || !state.Completed {
		t.Fatalf("exit state = %+v, %v — the completion must still commit", state, err)
	}
}

// TestA130AnEngineEnteredRoundTripIsUnchanged 는 결정 경로 회귀: 같은 왕복이 a130 이전 산식 그대로 동결됨.
func TestA130AnEngineEnteredRoundTripIsUnchanged(t *testing.T) {
	refusals := &a130Refusals{}
	j, _, _ := a130Journal(t, refusals)
	id := roundTrip(t, j, "10", "70000", "72000")
	got := outcomeOf(t, j, id)
	wantPnL, wantR := a130ExpectedR(t, "700000", "720000", "2000", "10")
	if got.RealizedPnLAfterCosts != wantPnL || got.RealizedR != wantR || got.InitialQuantity != "10" {
		t.Errorf("got %s / %s / %s, want %s / %s / 10", got.RealizedPnLAfterCosts, got.RealizedR,
			got.InitialQuantity, wantPnL, wantR)
	}
	if n := len(refusals.all()); n != 0 {
		t.Errorf("refusals = %d, want 0", n)
	}
}

// TestA130AnUnconfiguredModelIsReportedWithItsReason: 비용 모델 없이 묶인 저널의 종결은 cost_model_unconfigured 로 보임(B1).
func TestA130AnUnconfiguredModelIsReportedWithItsReason(t *testing.T) {
	refusals := &a130Refusals{}
	j := openTestJournal(t)
	if err := j.SetApplyHooks(ApplyHooks{Project: ProjectPosition, Exit: ApplyExitFill, OutcomeRefused: refusals.sink}); err != nil {
		t.Fatal(err)
	}
	roundTrip(t, j, "10", "70000", "72000")
	a130Only(t, refusals, OutcomeRefusalCostModelUnconfigured)
}

// TestA130TheRefusalRecordHasOnlyTheWhitelistedFields 는 거절 기록 타입의 필드 집합을 못 박음.
func TestA130TheRefusalRecordHasOnlyTheWhitelistedFields(t *testing.T) {
	var fields []string
	typ := reflect.TypeOf(TradeOutcomeRefusal{})
	for i := 0; i < typ.NumField(); i++ {
		fields = append(fields, typ.Field(i).Name)
	}
	sort.Strings(fields)
	if want := []string{"InstanceSeq", "Market", "Reason", "Symbol"}; !reflect.DeepEqual(fields, want) {
		t.Errorf("TradeOutcomeRefusal fields = %v, want %v", fields, want)
	}
}

// TestA130TheReasonSetIsClosedAndTokenShaped: 사유 목록은 이 파일의 TradeOutcomeRefusalReason 상수 전부와 같고(go/parser), 모두 토큰 모양.
func TestA130TheReasonSetIsClosedAndTokenShaped(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "trade_outcome_refusal.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	declared := map[string]bool{}
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, spec := range gen.Specs {
			vs := spec.(*ast.ValueSpec)
			if id, ok := vs.Type.(*ast.Ident); ok && id.Name == "TradeOutcomeRefusalReason" {
				for _, v := range vs.Values {
					declared[strings.Trim(v.(*ast.BasicLit).Value, `"`)] = true
				}
			}
		}
	}
	listed := map[string]bool{}
	shape := regexp.MustCompile(`^[a-z_]{1,64}$`)
	for _, r := range TradeOutcomeRefusalReasons() {
		if !shape.MatchString(string(r)) || listed[string(r)] {
			t.Errorf("reason %q is not a unique token", r)
		}
		listed[string(r)] = true
	}
	if !reflect.DeepEqual(declared, listed) || len(listed) != 20 {
		t.Errorf("declared %v, listed %v", declared, listed)
	}
}

// TestA130AnUnpricedEngineSellIsReportedAsSuch: 편입 포지션의 엔진 매도 평균가가 미관측이면 수량 불일치가 아니라 sell_unpriced 로 보임(B5 의 둘째 사유).
func TestA130AnUnpricedEngineSellIsReportedAsSuch(t *testing.T) {
	refusals := &a130Refusals{}
	j, fake, _ := a130Journal(t, refusals)
	a130Adjust(t, j, "acct-1", AdjustmentExternal, "0", "3", "70100")
	id := a130Adopt(t, j, "acct-1", "3", "70000", "66500")
	fake.Advance(time.Minute)
	a130Sell(t, j, "acct-1", id, "stop", "3", "")
	a130NoRow(t, j, id)
	a130Only(t, refusals, OutcomeRefusalSellUnpriced)
}

// TestA130AFoldInBeforeOrAtTheAdoptionIsNotALot(리뷰 2차 P2-5): fold-in(0→q) 조정이 편입 관측보다 N초 앞이든 같은 초든 늘어난 묶음이 아님 —
// 엔진이 편입 수량을 팔면 a130 이전 산식 그대로 동결. 기존 고정 시계 시험이 「같은 초」 한 경우만 재던 것을 둘로 나눔.
func TestA130AFoldInBeforeOrAtTheAdoptionIsNotALot(t *testing.T) {
	for _, before := range []time.Duration{30 * time.Second, 0} {
		refusals := &a130Refusals{}
		j, fake, _ := a130Journal(t, refusals)
		a130Adjust(t, j, "acct-1", AdjustmentExternal, "0", "3", "70100")
		fake.Advance(before)
		id := a130Adopt(t, j, "acct-1", "3", "70000", "66500")
		fake.Advance(time.Minute)
		a130Sell(t, j, "acct-1", id, "stop", "3", "72000")
		got, err := j.TradeOutcomeOf(context.Background(), id)
		if err != nil {
			t.Fatalf("fold-in %v before: TradeOutcomeOf: %v", before, err)
		}
		if _, wantR := a130ExpectedR(t, "210000", "216000", "3500", "3"); got.InitialQuantity != "3" || got.RealizedR != wantR {
			t.Errorf("fold-in %v before: qty=%s R=%s, want 3 / %s", before, got.InitialQuantity, got.RealizedR, wantR)
		}
		a130Silent(t, refusals)
	}
}

// TestA130AnOverAttributedSellWithoutAnyAdjustmentIsAForeignSell(리뷰 2차 P2-6 · 3차 R3-3): 조정 없이 이전 인스턴스의 flatten 매도가
// 쓸려 들어옴 — 그 체결은 이 인스턴스 수명 밖이라 foreign_instance_sell(a130 이전엔 수량 불일치로 거절되던 것, 이제 원인이 보임).
func TestA130AnOverAttributedSellWithoutAnyAdjustmentIsAForeignSell(t *testing.T) {
	refusals := &a130Refusals{}
	j, fake, _ := a130Journal(t, refusals)
	ctx := context.Background()
	buy := place(t, j, order{intentID: "i-b0", attemptID: "a-b0", orderID: "o-b0", decisionID: "d-b0", side: "BUY", quantity: "2"})
	if _, err := j.RecordFill(ctx, terminalFill(buy, "2", "70000")); err != nil {
		t.Fatal(err)
	}
	flat := place(t, j, order{intentID: "i-f0", attemptID: "a-f0", orderID: "o-f0", decisionID: "d-f0", side: "SELL", quantity: "2"})
	flattenLiquidation(t, j, "saga-0", flat)
	if _, err := j.RecordFill(ctx, terminalFill(flat, "2", "71000")); err != nil {
		t.Fatal(err)
	}
	refusals.mu.Lock()
	refusals.seen = nil
	refusals.mu.Unlock()

	fake.Advance(time.Minute)
	a130Adjust(t, j, "acct-1", AdjustmentExternal, "0", "10", "70100")
	id := a130Adopt(t, j, "acct-1", "10", "70000", "66500")
	fake.Advance(time.Minute)
	a130Sell(t, j, "acct-1", id, "stop", "10", "72000")

	a130NoRow(t, j, id)
	a130Only(t, refusals, OutcomeRefusalForeignInstanceSell)
}

// TestA130EachReachableRefusalSiteNamesItsOwnReason(리뷰 2차 P2-6): 판정 함수를 직접 불러 자리마다 사유를 대조함 — 잘못 붙은 사유를 잡음.
// 질의 실패 자리(position_unreadable · adoption_unreadable · engine_sell_unreadable · adoption_epoch_unreadable · cost_estimate_failed ·
// empty_buy_leg · insert_failed)는 정상 원장으로 도달할 수 없어 여기서 재지 않음(design D5).
func TestA130EachReachableRefusalSiteNamesItsOwnReason(t *testing.T) {
	ctx := context.Background()
	model := costs.DefaultModel()

	t.Run("unmanaged_position", func(t *testing.T) {
		j := outcomeFixture(t)
		o := place(t, j, order{intentID: "i-x", attemptID: "a-x", orderID: "o-x", side: "BUY", quantity: "5"})
		if _, err := j.RecordFill(ctx, terminalFill(o, "5", "70000")); err != nil {
			t.Fatal(err)
		}
		p := currentPosition(t, j, o)
		s := place(t, j, order{intentID: "i-xs", attemptID: "a-xs", orderID: "o-xs", side: "SELL", quantity: "5"})
		if _, err := j.RecordFill(ctx, terminalFill(s, "5", "71000")); err != nil {
			t.Fatal(err)
		}
		if _, reason := computeTradeOutcome(ctx, journalReader{j: j}, p.ID, model, nil); reason != OutcomeRefusalUnmanagedPosition {
			t.Errorf("reason = %q", reason)
		}
	})
	t.Run("cost_model_unconfigured", func(t *testing.T) {
		j := outcomeFixture(t)
		id := roundTrip(t, j, "10", "70000", "72000")
		if _, reason := computeTradeOutcome(ctx, journalReader{j: j}, id, costs.Model{}, nil); reason != OutcomeRefusalCostModelUnconfigured {
			t.Errorf("reason = %q", reason)
		}
	})
	t.Run("no_engine_sell", func(t *testing.T) {
		// 편입했고 편입 뒤 조정이 없는데 엔진 매도가 하나도 귀속되지 않음(판정 함수는 열린 포지션에도 부를 수 있음).
		j, _, _ := a130Journal(t, nil)
		a130Adjust(t, j, "acct-1", AdjustmentExternal, "0", "3", "70100")
		id := a130Adopt(t, j, "acct-1", "3", "70000", "66500")
		if _, reason := computeTradeOutcome(ctx, journalReader{j: j}, id, model, nil); reason != OutcomeRefusalNoEngineSell {
			t.Errorf("reason = %q", reason)
		}
	})
	t.Run("round_trip_unpriceable", func(t *testing.T) {
		j := outcomeFixture(t)
		buy := place(t, j, order{intentID: "i-buy", attemptID: "a-buy", orderID: "o-buy", decisionID: "d-buy", side: "BUY", quantity: "10"})
		if _, err := j.RecordFill(ctx, terminalFill(buy, "10", "")); err != nil {
			t.Fatal(err)
		}
		p := currentPosition(t, j, buy)
		if _, err := j.OpenExitState(ctx, ExitStateSeed{PositionID: p.ID, EntryPrice: "70000", InitialStop: "68000"}); err != nil {
			t.Fatal(err)
		}
		if _, reason := computeTradeOutcome(ctx, journalReader{j: j}, p.ID, model, nil); reason != OutcomeRefusalRoundTripUnpriceable {
			t.Errorf("reason = %q", reason)
		}
	})
	t.Run("exit_state_unreadable", func(t *testing.T) {
		j, _, _ := a130Journal(t, nil)
		a130Adjust(t, j, "acct-1", AdjustmentExternal, "0", "3", "70100")
		p, err := j.CurrentPosition(ctx, "acct-1", "kr", "005930")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := j.AdoptPosition(ctx, AdoptionRequest{PositionID: p.ID, Symbol: "005930", Market: "kr",
			Quantity: "3", CostBasis: "55000", ObservedPrice: "70000", SyntheticStop: "66500",
			ObservedAt: j.nowString()}); err != nil {
			t.Fatal(err)
		}
		if _, reason := computeTradeOutcome(ctx, journalReader{j: j}, p.ID, model, nil); reason != OutcomeRefusalExitStateUnreadable {
			t.Errorf("reason = %q", reason)
		}
	})
}

// TestA130RecordFillReportsRefusalsOnlyAfterItsCommit 는 구조 핀임(F-8): RecordFill 안에서 reportOutcomeRefusals 는 정확히 한 번,
// 마지막 tx.Commit() 호출(오류면 return 하는 if 의 조건) **뒤**에 있음. 행동 시험(롤백 무보고)은 훅 오류로만 롤백을 만들 수 있어
// 「커밋 자체가 실패」 하는 갈래를 못 재므로 순서를 구조로 고정함.
func TestA130RecordFillReportsRefusalsOnlyAfterItsCommit(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "fills.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var body *ast.BlockStmt
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "RecordFill" && fn.Recv != nil {
			body = fn.Body
		}
	}
	if body == nil {
		t.Fatal("RecordFill not found")
	}
	var lastCommit, report token.Pos
	reports := 0
	ast.Inspect(body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		switch sel.Sel.Name {
		case "Commit":
			if call.Pos() > lastCommit {
				lastCommit = call.Pos()
			}
		case "reportOutcomeRefusals":
			reports++
			report = call.Pos()
		}
		return true
	})
	if reports != 1 || lastCommit == token.NoPos || report < lastCommit {
		t.Errorf("RecordFill reports %d time(s) at %v, last commit at %v — want one report after the last commit",
			reports, fset.Position(report), fset.Position(lastCommit))
	}
}

// a130LocalSell 은 엔진 제안도 flatten 도 아닌 로컬 매도(사람이 tossctl 로 낸 주문 같은)를 체결시킴.
func a130LocalSell(t *testing.T, j *Journal, tag, quantity, price string) {
	t.Helper()
	sell := place(t, j, order{intentID: "i-" + tag, attemptID: "a-" + tag, orderID: "o-" + tag,
		decisionID: "d-" + tag, side: "SELL", quantity: quantity})
	if _, err := j.RecordFill(context.Background(), terminalFill(sell, quantity, price)); err != nil {
		t.Fatalf("RecordFill(%s): %v", tag, err)
	}
}

// TestA130AnUnattributedLocalSellIsRefused(R3-3): 편입 5 → 사람이 로컬 매도 2 → 엔진 3주 매도. a130 이전엔 수량 불일치, 이제 원인
// (귀속 밖의 로컬 매도)이 보임.
func TestA130AnUnattributedLocalSellIsRefused(t *testing.T) {
	refusals := &a130Refusals{}
	j, fake, _ := a130Journal(t, refusals)
	a130Adjust(t, j, "acct-1", AdjustmentExternal, "0", "5", "70100")
	id := a130Adopt(t, j, "acct-1", "5", "70000", "66500")
	fake.Advance(time.Minute)
	a130LocalSell(t, j, "hand", "2", "71000")
	fake.Advance(time.Minute)
	a130Sell(t, j, "acct-1", id, "stop", "3", "72000")
	a130NoRow(t, j, id)
	a130Only(t, refusals, OutcomeRefusalUnattributedLocalSell)
}

// TestA130ALocalSellOffsetBySweptFlattenIsNoLongerFrozen(R3-3): 위에 이전 인스턴스의 flatten 매도 2주가 더해지면 귀속 매도 = 3 + 2 = 5 =
// 편입 수량 — a130 이전엔 다른 인스턴스의 매도 대금이 섞인 조작 행. 이제 foreign_instance_sell.
func TestA130ALocalSellOffsetBySweptFlattenIsNoLongerFrozen(t *testing.T) {
	refusals := &a130Refusals{}
	j, fake, _ := a130Journal(t, refusals)
	ctx := context.Background()
	buy := place(t, j, order{intentID: "i-b0", attemptID: "a-b0", orderID: "o-b0", decisionID: "d-b0", side: "BUY", quantity: "2"})
	if _, err := j.RecordFill(ctx, terminalFill(buy, "2", "70000")); err != nil {
		t.Fatal(err)
	}
	flat := place(t, j, order{intentID: "i-f0", attemptID: "a-f0", orderID: "o-f0", decisionID: "d-f0", side: "SELL", quantity: "2"})
	flattenLiquidation(t, j, "saga-0", flat)
	if _, err := j.RecordFill(ctx, terminalFill(flat, "2", "71000")); err != nil {
		t.Fatal(err)
	}
	refusals.mu.Lock()
	refusals.seen = nil
	refusals.mu.Unlock()

	fake.Advance(time.Minute)
	a130Adjust(t, j, "acct-1", AdjustmentExternal, "0", "5", "70100")
	id := a130Adopt(t, j, "acct-1", "5", "70000", "66500")
	fake.Advance(time.Minute)
	a130LocalSell(t, j, "hand", "2", "71000")
	fake.Advance(time.Minute)
	a130Sell(t, j, "acct-1", id, "stop", "3", "72000")
	a130NoRow(t, j, id)
	a130Only(t, refusals, OutcomeRefusalForeignInstanceSell)
}

// TestA130ALaterInstanceFlattenIsNotThisTripsSell(리뷰 4차 P2-3): 늘어난 포지션 A 가 닫힌 뒤 같은 종목의 나중 인스턴스 B 가 flatten 으로
// 청산되면, 늦게 돈 backfill 에서 A 의 매도 다리에 B 의 flatten 이 쓸려 들어옴 — 닫힌 뒤 확정된 귀속 주문이므로 foreign_instance_sell.
func TestA130ALaterInstanceFlattenIsNotThisTripsSell(t *testing.T) {
	j, fake, _ := a130Journal(t, nil)
	ctx := context.Background()
	id := a130Grown(t, j, fake)
	prices := a130Prices(t, j, id, "69000")
	if _, reason := computeTradeOutcome(ctx, journalReader{j: j}, id, costs.DefaultModel(), prices); reason != "" {
		t.Fatalf("before the later flatten: reason = %s, want a priced outcome", reason)
	}
	fake.Advance(time.Minute)
	buy := place(t, j, order{intentID: "i-b9", attemptID: "a-b9", orderID: "o-b9", decisionID: "d-b9", side: "BUY", quantity: "2"})
	if _, err := j.RecordFill(ctx, terminalFill(buy, "2", "70000")); err != nil {
		t.Fatal(err)
	}
	fake.Advance(time.Minute)
	flat := place(t, j, order{intentID: "i-f9", attemptID: "a-f9", orderID: "o-f9", decisionID: "d-f9", side: "SELL", quantity: "2"})
	flattenLiquidation(t, j, "saga-9", flat)
	if _, err := j.RecordFill(ctx, terminalFill(flat, "2", "71000")); err != nil {
		t.Fatal(err)
	}
	if _, reason := computeTradeOutcome(ctx, journalReader{j: j}, id, costs.DefaultModel(), prices); reason != OutcomeRefusalForeignInstanceSell {
		t.Errorf("after the later flatten: reason = %q, want foreign_instance_sell", reason)
	}
}

// TestA130ALocalSellBeforeTheAdoptionIsNotThisTripsDisposal(리뷰 4차 P3-8): fold-in 과 편입 사이의 로컬 매도는 편입 수량에 이미 반영됨 —
// 편입 뒤 엔진이 편입 수량을 팔면 동결.
func TestA130ALocalSellBeforeTheAdoptionIsNotThisTripsDisposal(t *testing.T) {
	refusals := &a130Refusals{}
	j, fake, _ := a130Journal(t, refusals)
	a130Adjust(t, j, "acct-1", AdjustmentExternal, "0", "5", "70100")
	fake.Advance(time.Minute)
	a130LocalSell(t, j, "hand", "2", "71000")
	fake.Advance(time.Minute)
	id := a130Adopt(t, j, "acct-1", "3", "70000", "66500")
	fake.Advance(time.Minute)
	a130Sell(t, j, "acct-1", id, "stop", "3", "72000")
	if _, err := j.TradeOutcomeOf(context.Background(), id); err != nil {
		t.Fatalf("TradeOutcomeOf: %v — a sale before the adoption is not this round trip's", err)
	}
	a130Silent(t, refusals)
}

// TestA130TheJournalCannotReachTheNetwork(리뷰 4차 P3-12): 원장 패키지는 net/http 도 공식 API client 도 의존하지 않음 — 체결 트랜잭션 안
// 네트워크 0 이 구성으로만이 아니라 의존 그래프로 막힘.
func TestA130TheJournalCannotReachTheNetwork(t *testing.T) {
	out, err := exec.Command("go", "list", "-deps", ".").Output()
	if err != nil {
		t.Fatalf("go list -deps: %v", err)
	}
	for _, dep := range strings.Fields(string(out)) {
		if dep == "net/http" || strings.HasSuffix(dep, "/internal/official") {
			t.Errorf("internal/journal depends on %s", dep)
		}
	}
}
