package journal

// a066 task 5.7 — 부분·교체·선행 주문 늦은 체결의 crash·재시도 통합 시험(공유 bucket 에 의존하지 않는 부분).
//
// 증명할 것(tasks.md 5.7): Position, fill watermark, 비례 HELD 이전, filled=max(transfer, actual), 모든 bucket 의
// overage/latch 가 **정확히 한 번** 커밋되고, 그동안 위험 감소 경로(매도 체결)는 막히지 않음.
// 기존 Wave 1C/1D 시험은 각 성질을 따로 쟀음. 여기서는 한 흐름 안에서 crash 뒤 재시도와 중복 관측을 끼워 넣어
// "한 번" 을 원장 행 수·사용량·이벤트 수로 직접 셈.

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
)

// riskBucketLedgerCounts 는 "정확히 한 번" 을 재는 원장 행 수임.
type riskBucketLedgerCounts struct {
	fills, allocations, events, latches int
}

func readRiskBucketLedgerCounts(t *testing.T, j *Journal) riskBucketLedgerCounts {
	t.Helper()
	return riskBucketLedgerCounts{
		fills:       countRiskBucketRows(t, j, "risk_bucket_fills"),
		allocations: countRiskBucketRows(t, j, "risk_bucket_fill_allocations"),
		events:      countRiskBucketRows(t, j, "risk_bucket_events"),
		latches:     countRiskBucketRows(t, j, "risk_bucket_scope_latches"),
	}
}

func positionQuantity(t *testing.T, j *Journal, id string) string {
	t.Helper()
	var quantity string
	if err := j.db.QueryRow(`SELECT quantity FROM positions WHERE id=?`, id).Scan(&quantity); err != nil {
		t.Fatal(err)
	}
	return quantity
}

// TestA066PartialFillCrashThenRetryCommitsEverythingExactlyOnce: Position 투영 hook 이 한 번 실패(crash)하면 체결
// snapshot·sidecar·Position 이 모두 되돌려지고, 같은 관측을 다시 넣으면 전부 한 번 커밋되며, 같은 관측을 또 넣어도
// (재시도·중복) 아무것도 더 쓰이지 않음.
func TestA066PartialFillCrashThenRetryCommitsEverythingExactlyOnce(t *testing.T) {
	ctx := context.Background()
	j, key, decisionID, reserved := riskBucketFillFixture(t, "crash-retry", "risk-crash-retry")
	insertPosition(t, j, "crash-retry-position", nil)
	if err := j.RegisterRiskBucketOrder(ctx, RiskBucketOrderPlan{OrderID: "risk-crash-retry", DecisionID: decisionID, OrderQuantity: 10, ReservedMinor: reserved, ReservationPolicyDigest: "reservation-policy-v1", QuoteCurrency: "KRW", BaseCurrency: "KRW", CreatedAt: riskFillNow}); err != nil {
		t.Fatal(err)
	}
	crash := true
	boom := errors.New("synthetic crash inside the fill transaction")
	if err := j.SetApplyHooks(ApplyHooks{Project: func(ctx context.Context, tx *ApplyTx, fill AppliedFill) error {
		if crash {
			return boom
		}
		_, err := tx.Exec(ctx, `UPDATE positions SET quantity=?,state='OPEN' WHERE id='crash-retry-position'`, fill.CumulativeQuantity)
		return err
	}}); err != nil {
		t.Fatal(err)
	}
	before := readRiskBucketLedgerCounts(t, j)

	// 1) crash: 아무것도 남지 않음.
	if _, err := j.RecordFill(ctx, observation("risk-crash-retry", "4")); !errors.Is(err, boom) {
		t.Fatalf("crash fill error=%v", err)
	}
	if got := readRiskBucketLedgerCounts(t, j); got != before {
		t.Fatalf("crash left ledger rows: before=%+v after=%+v", before, got)
	}
	if _, err := j.LookupFill(ctx, "risk-crash-retry"); !errors.Is(err, ErrFillNotFound) {
		t.Fatalf("crash left the fill watermark: %v", err)
	}
	if q := positionQuantity(t, j, "crash-retry-position"); q != "10" {
		t.Fatalf("crash moved the position: %s", q)
	}
	state, err := j.ReadRiskBucketState(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	assertRiskUsage(t, state, "50", "0", false, false)

	// 2) 재시도: 한 번 커밋(4/10 체결 → HELD 50 의 비례 이전 20, 실제 증거 없음 → UNKNOWN latch).
	crash = false
	result, err := j.RecordFill(ctx, observation("risk-crash-retry", "4"))
	if err != nil || !result.Changed || result.Delta != "4" {
		t.Fatalf("retry after crash: %+v err=%v", result, err)
	}
	afterRetry := readRiskBucketLedgerCounts(t, j)
	if afterRetry.fills != before.fills+1 || afterRetry.allocations != before.allocations+5 {
		t.Fatalf("retry did not commit exactly one fill and five allocations: before=%+v after=%+v", before, afterRetry)
	}
	if fill, err := j.LookupFill(ctx, "risk-crash-retry"); err != nil || fill.FilledQuantity != "4" {
		t.Fatalf("watermark after retry: %+v err=%v", fill, err)
	}
	if q := positionQuantity(t, j, "crash-retry-position"); q != "4" {
		t.Fatalf("position after retry: %s", q)
	}
	state, err = j.ReadRiskBucketState(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	assertRiskUsage(t, state, "30", "20", false, true)

	// 3) 같은 관측 재전달(재시도·중복): 아무것도 더 쓰이지 않음.
	duplicate, err := j.RecordFill(ctx, observation("risk-crash-retry", "4"))
	if err != nil || duplicate.Changed || duplicate.Delta != "0" {
		t.Fatalf("duplicate observation changed the ledger: %+v err=%v", duplicate, err)
	}
	if got := readRiskBucketLedgerCounts(t, j); got != afterRetry {
		t.Fatalf("duplicate observation wrote rows: %+v → %+v", afterRetry, got)
	}
	stateAgain, err := j.ReadRiskBucketState(ctx, key)
	if err != nil || stateAgain.Digest != state.Digest {
		t.Fatalf("duplicate observation moved the risk state digest: %s → %s err=%v", state.Digest, stateAgain.Digest, err)
	}

	// 4) 실제 증거 보완은 filled=max(transfer, actual) 로 한 번만(12×4+1 fee = 49 > 이전 20), 같은 증거 재전달은 중복.
	actual := riskBucketActual("12", "1", "0")
	completed, err := j.completeRiskBucketFillActual(ctx, RiskBucketActualFillPlan{Owner: key, DecisionID: decisionID, OrderID: "risk-crash-retry", CumulativeFill: 4, Actual: actual, ObservedAt: riskFillNow})
	if err != nil || !completed.ActualEvidenceCompleted {
		t.Fatalf("actual evidence: %+v err=%v", completed, err)
	}
	state, err = j.ReadRiskBucketState(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	assertRiskUsage(t, state, "30", "48", false, false)
	again, err := j.completeRiskBucketFillActual(ctx, RiskBucketActualFillPlan{Owner: key, DecisionID: decisionID, OrderID: "risk-crash-retry", CumulativeFill: 4, Actual: actual, ObservedAt: riskFillNow})
	if err != nil || !again.Duplicate {
		t.Fatalf("repeated actual evidence was not a duplicate: %+v err=%v", again, err)
	}
	stateAgain, err = j.ReadRiskBucketState(ctx, key)
	if err != nil || stateAgain.Digest != state.Digest {
		t.Fatalf("repeated actual evidence moved the state: err=%v", err)
	}
}

// TestA066LateFillOverageLatchesEveryBucketOnceAndTheExitStaysOpen: 교체 뒤 선행 주문의 늦은 체결과 취소 뒤 후속 주문의
// 늦은 체결이 HELD 를 넘기면 다섯 bucket 모두에 overage 가 한 번 latch 되고(같은 관측 재전달은 아무것도 더 안 씀),
// 그 latch 아래에서도 같은 종목의 매도(위험 감소) 체결은 거절·지연 없이 기록됨. 신규 진입은 막힘.
func TestA066LateFillOverageLatchesEveryBucketOnceAndTheExitStaysOpen(t *testing.T) {
	ctx := context.Background()
	j, key, decisionID, reserved := riskBucketFillFixture(t, "overage-once", "risk-overage-parent")
	if err := j.RegisterRiskBucketOrder(ctx, RiskBucketOrderPlan{OrderID: "risk-overage-parent", DecisionID: decisionID, OrderQuantity: 10, ReservedMinor: reserved, ReservationPolicyDigest: "reservation-policy-v1", QuoteCurrency: "KRW", BaseCurrency: "KRW", CreatedAt: riskFillNow}); err != nil {
		t.Fatal(err)
	}
	if _, err := j.RecordFill(ctx, observation("risk-overage-parent", "4")); err != nil {
		t.Fatal(err)
	}
	recordConfirmedFillOrderScopeQuantity(t, j, "risk-overage-child-intent", "risk-overage-child-attempt", "risk-overage-child", FillSnapshotScope{
		AccountRef: "acct-1", Market: "us", TradingDay: "2026-03-30", Symbol: "AAPL", Side: "BUY",
	}, "6")
	bindRiskOrderAttemptDecision(t, j, "risk-overage-child", decisionID)
	if err := j.RegisterRiskBucketOrder(ctx, RiskBucketOrderPlan{OrderID: "risk-overage-child", DecisionID: decisionID, PredecessorOrderID: "risk-overage-parent", OrderQuantity: 6, ReservedMinor: riskReservedMap("30"), ReservationPolicyDigest: "reservation-policy-v1", QuoteCurrency: "KRW", BaseCurrency: "KRW", CreatedAt: riskFillNow.Add(time.Second)}); err != nil {
		t.Fatal(err)
	}
	child := observation("risk-overage-child", "3")
	child.Quantity = "6"
	if _, err := j.RecordFill(ctx, child); err != nil {
		t.Fatal(err)
	}
	if _, err := j.RecordFill(ctx, observation("risk-overage-parent", "5")); err != nil { // 선행 주문의 늦은 체결
		t.Fatal(err)
	}
	if _, err := j.releaseRiskBucketOrder(ctx, RiskBucketOrderRelease{Owner: key, DecisionID: decisionID, OrderID: "risk-overage-child", Reason: RiskBucketReleaseCancel, ReleasedAt: riskFillNow.Add(2 * time.Second)}); err != nil {
		t.Fatal(err)
	}
	child.FilledQuantity = "4" // 취소 뒤 후속 주문의 늦은 체결 — 남은 HELD 가 없어 overage
	if _, err := j.RecordFill(ctx, child); err != nil {
		t.Fatalf("late successor fill was dropped: %v", err)
	}
	state, err := j.ReadRiskBucketState(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	assertRiskUsage(t, state, "0", "45", true, true)
	// 한도(100) 안이라 overage_minor 는 0 이지만, 해제된 HELD 로는 늦은 체결을 덮지 못하므로 RISK_OVERAGE latch 가 다섯
	// bucket 전부에 섬(설계 D5: "released HELD is insufficient → latch overage").
	var latchedReservations int
	if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_reservations WHERE decision_id=? AND risk_overage_latched=1`, decisionID).Scan(&latchedReservations); err != nil || latchedReservations != 5 {
		t.Fatalf("overage latched on %d of 5 buckets err=%v", latchedReservations, err)
	}
	once := readRiskBucketLedgerCounts(t, j)
	// 같은 늦은 체결들을 다시 전달해도 아무것도 더 쓰이지 않음.
	for _, obs := range []FillObservation{child, observation("risk-overage-parent", "5")} {
		if result, err := j.RecordFill(ctx, obs); err != nil || result.Changed {
			t.Fatalf("redelivered late fill %s changed the ledger: %+v err=%v", obs.OrderID, result, err)
		}
	}
	if got := readRiskBucketLedgerCounts(t, j); got != once {
		t.Fatalf("redelivery wrote rows: %+v → %+v", once, got)
	}
	again, err := j.ReadRiskBucketState(ctx, key)
	if err != nil || again.Digest != state.Digest {
		t.Fatalf("redelivery moved the risk state: err=%v", err)
	}

	// 위험 감소: 같은 계좌·종목의 매도 체결은 latch 아래에서도 거절·실패 없이 기록되고 bucket 계상을 거치지 않음.
	recordConfirmedFillOrderScope(t, j, "overage-exit-intent", "overage-exit-attempt", "overage-exit", FillSnapshotScope{AccountRef: "acct-1", Market: "us", TradingDay: "2026-03-30", Symbol: "AAPL", Side: "SELL"})
	exit := observation("overage-exit", "9")
	exit.Side, exit.Quantity = "SELL", "9"
	if result, err := j.RecordFill(ctx, exit); err != nil || !result.Changed || result.FailClosed {
		t.Fatalf("exit fill under overage latch: %+v err=%v", result, err)
	}
	if got := readRiskBucketLedgerCounts(t, j); got != once {
		t.Fatalf("exit fill touched risk bucket accounting: %+v → %+v", once, got)
	}

	// 신규 진입은 latch 로 막힘(같은 owner 범위).
	seedExistingRiskReservation(t, j, "existing-overage-next", "acct-1")
	next := riskBucketAdmissionFixture(t, "overage-next", "acct-1", "lane-short", "campaign-1", key.ProspectiveGeneration, "1000", "45")
	next.Owner.Key = key
	next.Admission.Policy.QuoteCurrency = "USD"
	rebindRiskBucket(t, &next, 1, riskbucket.BucketKey{Dimension: riskbucket.DimensionMarket, Value: string(riskbucket.MarketUS), PolicyVersion: "policy-v1"})
	rebindRiskBucket(t, &next, 4, riskbucket.BucketKey{Dimension: riskbucket.DimensionSymbol, Value: "AAPL", PolicyVersion: "policy-v1"})
	refreshSnapshotUsageFromLedger(t, j, &next)
	if _, err := j.CommitRiskBucketAdmission(ctx, next); !errors.Is(err, ErrRiskBucketEntryBlocked) {
		t.Fatalf("new exposure under an overage latch was not blocked: %v", err)
	}
}
