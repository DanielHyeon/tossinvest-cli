package journal

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

// a138_open_positions_test.go 는 최대 보유 포지션 수(risk-management 「최대 보유 포지션 수 한도」)의 원장 판정을 고정함.
//
// 세는 것(사용자 결정 2026-10-15 D1 = B): 계좌의 살아 있는 포지션(엔진 진입 · 편입 · 대사가 접어 넣은 외부 보유) ∪ HELD 진입 예약의 종목 ∪
// 이번 결정의 종목 — 서로 다른 (시장, 종목). 그 수가 상한을 **넘으면** 거부. 이미 든 종목의 진입은 수를 늘리지 않음(D2 = 허용).
// 시계 · 노출 한도 · 하루 매수 상한은 넉넉하게 두어 이 시험에서는 개수 상한만 묶임.

// a138Journal 은 예약 시험 시계(2026-03-30 00:30 UTC)에 포지션 투영 훅을 단 원장임 — 체결이 포지션 행을 만듦.
func a138Journal(t *testing.T) *Journal {
	t.Helper()
	j, _ := openReservationJournal(t)
	if err := j.SetApplyHooks(ApplyHooks{Project: ProjectPosition}); err != nil {
		t.Fatalf("SetApplyHooks: %v", err)
	}
	return j
}

// a138Decision 은 종목을 고를 수 있는 진입 결정을 기록함(recordEntryDecision 은 005930 고정).
func a138Decision(t *testing.T, j *Journal, id, market, symbol string) Decision {
	t.Helper()
	issued := j.Now()
	dec, err := j.RecordDecision(context.Background(), DecisionRequest{
		ID: id, AccountRef: "acct-1", SafetyClass: SafetyClassExposureRaising, Kind: KindPlace,
		Preimage: RiskIntent{
			AccountRef: "acct-1", Market: market, Symbol: symbol, Side: "BUY",
			Quantity: "1", EntryPrice: "100", StopPrice: "90", PolicyVersion: "test-1",
		},
		LimitsJSON: `{"max_notional":"1000000"}`,
		Nonce:      "nonce-" + id,
		IssuedAt:   issued,
		ExpiresAt:  issued.Add(10 * time.Minute),
	})
	if err != nil {
		t.Fatalf("RecordDecision(%s): %v", id, err)
	}
	return dec
}

// a138Request 는 개수 상한 max 를 실은 OPEN_EXPOSURE 예약 요청임.
func a138Request(t *testing.T, j *Journal, id string, max int) ReserveRequest {
	t.Helper()
	req := exposureReserve(j, id, "acct-1", "100", "0", "1000000000", mustVersion(t, j, "acct-1"))
	req.OpenPositionLimit = &OpenPositionLimit{Max: max}
	return req
}

// a138Hold 는 그 종목의 진입을 발급 · 예약해 HELD 로 둠(= 아직 포지션이 안 된 진행 중 진입).
func a138Hold(t *testing.T, j *Journal, id, market, symbol string, max int) {
	t.Helper()
	a138Decision(t, j, id, market, symbol)
	if _, err := j.Reserve(context.Background(), a138Request(t, j, id, max)); err != nil {
		t.Fatalf("Reserve(%s %s/%s): %v", id, market, symbol, err)
	}
}

func a138Try(t *testing.T, j *Journal, id, market, symbol string, max int) error {
	t.Helper()
	a138Decision(t, j, id, market, symbol)
	_, err := j.Reserve(context.Background(), a138Request(t, j, id, max))
	return err
}

func a138RequireRefused(t *testing.T, j *Journal, id, market, symbol string, max int) {
	t.Helper()
	err := a138Try(t, j, id, market, symbol, max)
	if !errors.Is(err, ErrReservationLimitExceeded) || !strings.Contains(err.Error(), "MAX_OPEN_POSITIONS") {
		t.Fatalf("Reserve(%s %s/%s) must be refused by the open position maximum, got %v", id, market, symbol, err)
	}
	rows, err := j.ReservationsForDecision(context.Background(), id)
	if err != nil {
		t.Fatalf("ReservationsForDecision: %v", err)
	}
	if len(rows) != 0 {
		t.Fatalf("a refused reservation must write nothing, found %+v", rows)
	}
}

// a138EnginePosition 은 엔진 진입 주문이 체결되어 entry_decision_id 를 가진 포지션을 만듦(실제 Prepare · RecordFill 경로).
func a138EnginePosition(t *testing.T, j *Journal, n int, symbol string) {
	t.Helper()
	o := place(t, j, order{
		intentID: fmt.Sprintf("i-%d", n), attemptID: fmt.Sprintf("a-%d", n), orderID: fmt.Sprintf("o-%d", n),
		decisionID: fmt.Sprintf("d-eng-%d", n), symbol: symbol,
	})
	if _, err := j.RecordFill(context.Background(), terminalFill(o, "10", "70000")); err != nil {
		t.Fatalf("RecordFill(%s): %v", symbol, err)
	}
}

// a138ExternalPosition 은 대사가 계좌 보유를 접어 넣은 외부 포지션(진입 결정 · 편입 없음)을 만듦 — 생산 경로 ApplyPositionAdjustment.
func a138ExternalPosition(t *testing.T, j *Journal, market, symbol string) Position {
	t.Helper()
	ctx := context.Background()
	watermark, err := j.FillWatermark(ctx, symbol)
	if err != nil {
		t.Fatal(err)
	}
	res, err := j.ApplyPositionAdjustment(ctx, AdjustmentRequest{
		AccountRef: "acct-1", Market: market, Symbol: symbol, Kind: AdjustmentExternal,
		ExpectedPrevQuantity: "0", ExpectedFillWatermark: watermark, NewQuantity: "5",
		BrokerAsOf: "2026-03-30T00:29:00Z", Evidence: "holdings snapshot reported 5",
	})
	if err != nil {
		t.Fatalf("ApplyPositionAdjustment(%s): %v", symbol, err)
	}
	if !res.OpenedInstance || res.Position.ExitEligible() {
		t.Fatalf("fixture: %s must be an unmanaged external position, got %+v", symbol, res)
	}
	return res.Position
}

// a138AdoptedPosition 은 외부 포지션을 편입해 adoption_id 를 가진 관리 포지션으로 만듦.
func a138AdoptedPosition(t *testing.T, j *Journal, market, symbol string) {
	t.Helper()
	p := a138ExternalPosition(t, j, market, symbol)
	if _, err := j.AdoptPosition(context.Background(), AdoptionRequest{
		PositionID: p.ID, Symbol: symbol, Market: market, Quantity: "5", CostBasis: "100",
		ObservedPrice: "100", SyntheticStop: "95", ObservedAt: "2026-03-30T00:30:00Z",
	}); err != nil {
		t.Fatalf("AdoptPosition(%s): %v", symbol, err)
	}
}

var a138Symbols = []string{"000001", "000002", "000003", "000004", "000005", "000006", "000007"}

func TestA138TheSeventhSymbolIsRefused(t *testing.T) {
	j := a138Journal(t)
	for i, symbol := range a138Symbols[:6] {
		a138Hold(t, j, fmt.Sprintf("d-%d", i), "kr", symbol, 6)
	}
	a138RequireRefused(t, j, "d-7", "kr", a138Symbols[6], 6)
	// 상한은 허용되는 최대 개수 — 6 번째는 통과했고(위), 상한 7 이면 7 번째도 통과.
	if err := a138Try(t, j, "d-7b", "kr", a138Symbols[6], 7); err != nil {
		t.Fatalf("with a maximum of 7 the seventh symbol fits, got %v", err)
	}
}

func TestA138AnEntryOnAHeldSymbolIsAllowedAtTheMaximum(t *testing.T) {
	j := a138Journal(t)
	for i, symbol := range a138Symbols[:6] {
		a138Hold(t, j, fmt.Sprintf("d-%d", i), "kr", symbol, 6)
	}
	// 이미 집합에 있는 종목(HELD 예약)의 진입 — 수가 늘지 않음(D2 = 허용).
	if err := a138Try(t, j, "d-again", "kr", a138Symbols[2], 6); err != nil {
		t.Fatalf("a scale-in on a held symbol must not be refused by the maximum, got %v", err)
	}
	// 살아 있는 포지션의 종목도 같음.
	k := a138Journal(t)
	for i, symbol := range a138Symbols[:6] {
		a138EnginePosition(t, k, i, symbol)
	}
	if err := a138Try(t, k, "d-scale", "kr", a138Symbols[0], 6); err != nil {
		t.Fatalf("a scale-in on a live position must not be refused by the maximum, got %v", err)
	}
	a138RequireRefused(t, k, "d-new", "kr", a138Symbols[6], 6)
}

func TestA138EveryKindOfLivePositionCounts(t *testing.T) {
	j := a138Journal(t)
	for i, symbol := range a138Symbols[:4] {
		a138EnginePosition(t, j, i, symbol)
	}
	a138AdoptedPosition(t, j, "kr", a138Symbols[4])
	a138ExternalPosition(t, j, "kr", a138Symbols[5])
	// 엔진 4 + 편입 1 + 외부 1 = 6 — 새 종목은 일곱 번째.
	a138RequireRefused(t, j, "d-new", "kr", a138Symbols[6], 6)
	if err := a138Try(t, j, "d-ok", "kr", a138Symbols[6], 7); err != nil {
		t.Fatalf("with a maximum of 7 the entry fits, got %v", err)
	}
}

func TestA138PositionsAndHoldsOfOneSymbolCountOnceAndClosedOnesNotAtAll(t *testing.T) {
	j := a138Journal(t)
	ctx := context.Background()
	// 같은 종목의 포지션 + HELD 예약 = 하나.
	a138EnginePosition(t, j, 0, a138Symbols[0])
	a138Hold(t, j, "d-same", "kr", a138Symbols[0], 6)
	// 같은 종목이라도 시장이 다르면 다른 칸.
	a138Hold(t, j, "d-us", "us", a138Symbols[0], 6)
	// 닫힌 포지션은 세지 않음 — 외부 보유를 계좌 값 0 으로 수렴시켜 CLOSED 로 만듦.
	ext := a138ExternalPosition(t, j, "kr", a138Symbols[1])
	watermark, err := j.FillWatermark(ctx, a138Symbols[1])
	if err != nil {
		t.Fatal(err)
	}
	res, err := j.ApplyPositionAdjustment(ctx, AdjustmentRequest{
		AccountRef: "acct-1", Market: "kr", Symbol: a138Symbols[1], Kind: AdjustmentExternal,
		ExpectedPrevQuantity: ext.Quantity, ExpectedFillWatermark: watermark, NewQuantity: "0",
		BrokerAsOf: "2026-03-30T00:29:30Z", Evidence: "holdings snapshot reported 0",
	})
	if err != nil || res.Position.State != PositionClosed {
		t.Fatalf("fixture: the external position must close, got %+v, %v", res.Position, err)
	}
	// 지금 셈: kr/000001 · us/000001 = 2. 상한 3 이면 새 종목 하나 통과, 그다음 하나는 거부.
	if err := a138Try(t, j, "d-third", "kr", a138Symbols[2], 3); err != nil {
		t.Fatalf("two distinct live symbols plus one new must fit a maximum of 3, got %v", err)
	}
	a138RequireRefused(t, j, "d-fourth", "kr", a138Symbols[3], 3)
}

func TestA138ConcurrentEntriesCannotBothTakeTheLastSlot(t *testing.T) {
	j := a138Journal(t)
	ctx := context.Background()
	for i, symbol := range a138Symbols[:5] {
		a138Hold(t, j, fmt.Sprintf("d-%d", i), "kr", symbol, 6)
	}
	a138Decision(t, j, "d-x", "kr", "100001")
	a138Decision(t, j, "d-y", "kr", "100002")
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		okCount int
		errs    []error
	)
	for _, id := range []string{"d-x", "d-y"} {
		wg.Add(1)
		go func(decision string) {
			defer wg.Done()
			_, err := j.ReserveWithRecollection(ctx, func(ctx context.Context, attempt int) (ReserveRequest, error) {
				version, err := j.ReservationVersion(ctx, "acct-1")
				if err != nil {
					return ReserveRequest{}, err
				}
				req := exposureReserve(j, decision, "acct-1", "100", "0", "1000000000", version)
				req.OpenPositionLimit = &OpenPositionLimit{Max: 6}
				return req, nil
			}, RecollectPolicy{})
			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				okCount++
				return
			}
			errs = append(errs, err)
		}(id)
	}
	wg.Wait()
	// 진 쪽이 재수집(ErrSnapshotSuperseded)을 거쳤든 아니든 마지막 거부는 개수 상한이어야 함 — 판정이 없으면 둘 다 통과함.
	if okCount != 1 || len(errs) != 1 || !errors.Is(errs[0], ErrReservationLimitExceeded) ||
		!strings.Contains(errs[0].Error(), "MAX_OPEN_POSITIONS") {
		t.Fatalf("exactly one of two concurrent new symbols may take the sixth slot, ok=%d errs=%v", okCount, errs)
	}
}

func TestA138AnExposureRaisingReservationWithoutTheMaximumIsRefused(t *testing.T) {
	j := a138Journal(t)
	a138Decision(t, j, "d-bare", "kr", a138Symbols[0])
	req := a138Request(t, j, "d-bare", 6)
	req.OpenPositionLimit = nil
	if _, err := j.Reserve(context.Background(), req); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("an exposure-raising reservation without the maximum must be refused, got %v", err)
	}
	for _, bad := range []int{0, -1} {
		req.OpenPositionLimit = &OpenPositionLimit{Max: bad}
		if _, err := j.Reserve(context.Background(), req); !errors.Is(err, ErrInvalidRequest) {
			t.Fatalf("maximum %d must be refused as an invalid request, got %v", bad, err)
		}
	}
	if rows, _ := j.ReservationsForDecision(context.Background(), "d-bare"); len(rows) != 0 {
		t.Fatalf("a refused reservation must write nothing, found %+v", rows)
	}
}

func TestA138AnUncountableHoldFailsClosed(t *testing.T) {
	j := a138Journal(t)
	a138Hold(t, j, "d-bad", "kr", a138Symbols[0], 6)
	// HELD 예약을 가진 결정의 preimage 가 해석 불가 — 그 종목을 셀 수 없으므로 0 으로 두지 않고 판정 실패.
	if _, err := j.db.ExecContext(context.Background(),
		"UPDATE decisions SET risk_preimage = 'not-json' WHERE id = 'd-bad'"); err != nil {
		t.Fatal(err)
	}
	err := a138Try(t, j, "d-after", "kr", a138Symbols[1], 6)
	if err == nil || errors.Is(err, ErrReservationLimitExceeded) {
		t.Fatalf("an uncountable hold must refuse the entry as a failed count, not pass or read as a limit, got %v", err)
	}
	if rows, _ := j.ReservationsForDecision(context.Background(), "d-after"); len(rows) != 0 {
		t.Fatalf("a refused reservation must write nothing, found %+v", rows)
	}
}

// freeze 리뷰 P1-1: 종결 체결이 예약을 해제한 뒤 투영이 거절되면 포지션 행 없이 종목 범위 RECONCILE 만 남음 — 그 종목도 한 칸.
func TestA138AReconciledSymbolWithoutAPositionRowCounts(t *testing.T) {
	ctx := context.Background()
	for _, market := range []string{"", "KR"} {
		t.Run("scope market "+market, func(t *testing.T) {
			j := a138Journal(t)
			for i, symbol := range a138Symbols[:5] {
				a138Hold(t, j, fmt.Sprintf("d-%d", i), "kr", symbol, 6)
			}
			if _, _, err := j.EnterReconcile(ctx, EnterReconcileRequest{
				AccountRef: "acct-1", Symbol: "777777", ScopeMarket: market, Cause: ReconcileCauseAttributionFailed,
				Evidence: "position projection refused the fill",
			}); err != nil {
				t.Fatalf("EnterReconcile: %v", err)
			}
			a138RequireRefused(t, j, "d-new", "kr", a138Symbols[5], 6)
		})
	}
	// RECONCILE 종목이 이미 S 에 있으면(같은 종목의 HELD 예약) 두 번 세지 않음.
	j := a138Journal(t)
	for i, symbol := range a138Symbols[:5] {
		a138Hold(t, j, fmt.Sprintf("d-%d", i), "kr", symbol, 6)
	}
	if _, _, err := j.EnterReconcile(ctx, EnterReconcileRequest{
		AccountRef: "acct-1", Symbol: a138Symbols[0], Cause: ReconcileCauseQuantityMismatch, Evidence: "mismatch",
	}); err != nil {
		t.Fatalf("EnterReconcile: %v", err)
	}
	if err := a138Try(t, j, "d-sixth", "kr", a138Symbols[5], 6); err != nil {
		t.Fatalf("a RECONCILE scope on a symbol already counted must not count twice, got %v", err)
	}
	// 계좌 전체 RECONCILE 은 종목이 없어 개수와 무관(진입은 latch 가 막음).
}

// freeze 리뷰 P1-2: Gateway 는 종류와 무관하게 HELD 예약이 있으면 제출을 인가하므로 DAILY_LOSS 만 예약한 결정의 종목도 셈.
// 그리고 해제된(RELEASED) 예약은 세지 않음.
func TestA138EveryHeldReservationKindCountsAndReleasedOnesDoNot(t *testing.T) {
	ctx := context.Background()
	j := a138Journal(t)
	for i, symbol := range a138Symbols[:5] {
		a138Hold(t, j, fmt.Sprintf("d-%d", i), "kr", symbol, 6)
	}
	a138Decision(t, j, "d-loss", "kr", "555555")
	req := ReserveRequest{
		DecisionID: "d-loss", AccountRef: "acct-1", SnapshotAsOf: j.Now(), ObservedVersion: mustVersion(t, j, "acct-1"),
		SnapshotUsage:     []AggregateAmount{{Kind: ReservationKindDailyLoss, Amount: "0", Currency: "KRW"}},
		Limits:            []AggregateAmount{{Kind: ReservationKindDailyLoss, Amount: "1000000000", Currency: "KRW"}},
		Reservations:      []ReservationRequest{{ID: "res-d-loss", Kind: ReservationKindDailyLoss, Amount: "1", Currency: "KRW", TradingDay: "2026-03-30"}},
		DailyBuyLimit:     &DailyBuyLimit{Amount: "1000000000000", Currency: "KRW", ManualBuy: a136ZeroManualBuy(j.Now())},
		OpenPositionLimit: &OpenPositionLimit{Max: 6},
	}
	if _, err := j.Reserve(ctx, req); err != nil {
		t.Fatalf("Reserve(DAILY_LOSS only): %v", err)
	}
	a138RequireRefused(t, j, "d-new", "kr", a138Symbols[5], 6)

	// 운영자 해제로 RELEASED 가 된 예약의 종목은 빠짐.
	k := a138Journal(t)
	for i, symbol := range a138Symbols[:6] {
		a138Hold(t, k, fmt.Sprintf("d-%d", i), "kr", symbol, 6)
	}
	if _, err := k.db.ExecContext(ctx,
		"UPDATE risk_reservations SET state = 'RELEASED', released_at = ?, release_reason = 'OPERATOR' WHERE decision_id = 'd-0'",
		k.nowString()); err != nil {
		t.Fatal(err)
	}
	if err := a138Try(t, k, "d-after", "kr", a138Symbols[6], 6); err != nil {
		t.Fatalf("a released reservation must not hold a slot, got %v", err)
	}
}

// 변이 M16 이 살아남은 자리를 못 박음: 원장은 노출을 올리지 않는 결정의 예약도 받으므로(checkDecisionReservable 는 종류를 묻지 않음),
// ② 의 safety_class 거름이 없으면 그런 결정의 종목이 칸을 차지함. Gateway 는 EXPOSURE_RAISING 결정의 예약만 매수 인가로 쓰므로 세지 않는 것이 맞음.
func TestA138AHeldReservationOfANonEntryDecisionTakesNoSlot(t *testing.T) {
	ctx := context.Background()
	j := a138Journal(t)
	for i, symbol := range a138Symbols[:5] {
		a138Hold(t, j, fmt.Sprintf("d-%d", i), "kr", symbol, 6)
	}
	issued := j.Now()
	if _, err := j.RecordDecision(ctx, DecisionRequest{
		ID: "d-reduce", AccountRef: "acct-1", SafetyClass: SafetyClassRiskReducing, Kind: KindPlace,
		Preimage: ReductionIntent{AccountRef: "acct-1", Market: "kr", Symbol: "444444", Side: "SELL", MaxQuantity: "1", Reason: "test"},
		Nonce:    "nonce-d-reduce", IssuedAt: issued, ExpiresAt: issued.Add(10 * time.Minute),
	}); err != nil {
		t.Fatalf("RecordDecision(RISK_REDUCING): %v", err)
	}
	req := exposureReserve(j, "d-reduce", "acct-1", "1", "0", "1000000000", mustVersion(t, j, "acct-1"))
	if _, err := j.Reserve(ctx, req); err != nil {
		t.Fatalf("fixture: the journal accepts a reservation of a non-entry decision: %v", err)
	}
	if err := a138Try(t, j, "d-sixth", "kr", a138Symbols[5], 6); err != nil {
		t.Fatalf("a non-entry decision's hold must not take a slot, got %v", err)
	}
}

// 구현 리뷰 P3-1: ④ 의 맞춤 규칙을 갈래마다 못 박음 — 해제된 상태 · 시장 범위가 이미 든 종목과 겹침 · 이번 종목 자신 · 다른 시장의 같은 종목.
func TestA138ReconcileScopesMatchExactlyWhatTheyCover(t *testing.T) {
	ctx := context.Background()
	enter := func(t *testing.T, j *Journal, symbol, market string) {
		t.Helper()
		if _, _, err := j.EnterReconcile(ctx, EnterReconcileRequest{AccountRef: "acct-1", Symbol: symbol, ScopeMarket: market,
			Cause: ReconcileCauseQuantityMismatch, Evidence: "mismatch"}); err != nil {
			t.Fatalf("EnterReconcile(%s %s): %v", market, symbol, err)
		}
	}
	five := func(t *testing.T, j *Journal) {
		t.Helper()
		for i, symbol := range a138Symbols[:5] {
			a138Hold(t, j, fmt.Sprintf("d-%d", i), "kr", symbol, 6)
		}
	}

	t.Run("released RECONCILE takes no slot", func(t *testing.T) {
		j := a138Journal(t)
		five(t, j)
		enter(t, j, "777777", "")
		if _, _, err := j.ReleaseReconcile(ctx, ReleaseReconcileRequest{AccountRef: "acct-1", Symbol: "777777",
			Cause: ReconcileReleaseRecheckMatched, Evidence: "recheck matched"}); err != nil {
			t.Fatalf("ReleaseReconcile: %v", err)
		}
		if err := a138Try(t, j, "d-sixth", "kr", a138Symbols[5], 6); err != nil {
			t.Fatalf("a released RECONCILE must not hold a slot, got %v", err)
		}
	})
	t.Run("KR scope over a held kr symbol counts once", func(t *testing.T) {
		j := a138Journal(t)
		five(t, j)
		enter(t, j, a138Symbols[0], "KR") // 시장 범위 'KR' 과 결정 시장 'kr' 가 같은 칸이어야 함(정규화).
		if err := a138Try(t, j, "d-sixth", "kr", a138Symbols[5], 6); err != nil {
			t.Fatalf("a KR-scoped RECONCILE on a held kr symbol must not count twice, got %v", err)
		}
	})
	t.Run("this symbol's own RECONCILE is this slot", func(t *testing.T) {
		for _, market := range []string{"", "KR"} {
			j := a138Journal(t)
			five(t, j)
			enter(t, j, a138Symbols[5], market)
			if err := a138Try(t, j, "d-sixth", "kr", a138Symbols[5], 6); err != nil {
				t.Fatalf("scope %q: the entry's own RECONCILE symbol is the entry's slot, not another one, got %v", market, err)
			}
		}
	})
	t.Run("KR scope beside a held us symbol is another slot", func(t *testing.T) {
		j := a138Journal(t)
		for i, symbol := range a138Symbols[:4] {
			a138Hold(t, j, fmt.Sprintf("d-%d", i), "kr", symbol, 6)
		}
		a138Hold(t, j, "d-us", "us", "888888", 6)
		enter(t, j, "888888", "KR") // (kr, 888888) 은 (us, 888888) 과 다른 칸.
		a138RequireRefused(t, j, "d-new", "kr", a138Symbols[5], 6)
	})
}
