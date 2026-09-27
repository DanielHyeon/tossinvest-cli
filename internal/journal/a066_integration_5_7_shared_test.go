package journal

// a066 task 5.7 — 공유 bucket 에 걸친 체결 overage 전파(측정 시작점, 2026-09-28).
//
// 설계 D5: "각 transaction 은 filled + remaining HELD - monetary limit 의 positive delta 를 horizon, market, strategy, sector 와
// symbol 각각에 overage 로 저장" — bucket 의 filled·HELD 는 그 bucket 을 쓰는 **모든** 진입의 합임(admission 도 공유 bucket 을
// 원장 합으로 cap 함). 두 owner 가 공유 bucket 을 한도까지 채운 뒤 한 owner 의 실제 체결가가 예약보다 높으면 공유 bucket 의
// 합이 한도를 넘음 — 그 넘침이 overage 로 기록·latch 되는지 잼.
//
// 측정(2026-09-28, 수리 전): 네 공유 bucket 모두 사용량 128 > 한도 100 인데 RISK_OVERAGE latch 0 건 — 체결 계상이 그 owner 의
// 합만 한도와 비교했음(analysis/mutation-5.6.1/shared-overage-measure.log). 측정 로그를 계약 단언으로 바꾸어 수리의 RED 로 승격.

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
)

func TestA066SharedBucketOverageFromOneOwnersActualFillIsLatched(t *testing.T) {
	ctx := context.Background()
	// A: US AAPL, 공유 bucket 한도 100, 예약 50.
	j, keyA, decisionA, reservedA := riskBucketFillFixture(t, "shared-overage-a", "risk-shared-a")
	// B: US MSFT, 같은 horizon·market(US)·strategy·sector 를 씀, 원장 그대로의 snapshot(50 사용) → 남은 50 → 예약 50.
	seedExistingRiskReservation(t, j, "existing-shared-b", "acct-1")
	planB := riskBucketAdmissionFixture(t, "shared-b", "acct-1", "lane-b", "campaign-b", "prospective-shared-b", "100", "0")
	planB.ExistingReservationID = "existing-shared-b"
	planB.Owner.Key.Market, planB.Owner.Key.Symbol = riskbucket.MarketUS, "MSFT"
	planB.Admission.Policy.QuoteCurrency, planB.Admission.Policy.AccountCurrency = "USD", "KRW"
	rebindRiskBucket(t, &planB, 1, riskbucket.BucketKey{Dimension: riskbucket.DimensionMarket, Value: "US", PolicyVersion: "policy-v1"})
	rebindRiskBucket(t, &planB, 4, riskbucket.BucketKey{Dimension: riskbucket.DimensionSymbol, Value: "MSFT", PolicyVersion: "policy-v1"})
	refreshSnapshotUsageFromLedger(t, j, &planB)
	receiptB, err := j.CommitRiskBucketAdmission(ctx, planB)
	if err != nil || receiptB.QFinal != 10 {
		t.Fatalf("B admission: q_final=%d err=%v", receiptB.QFinal, err)
	}
	if got := contractUsage(t, j, riskbucket.DimensionStrategy, "strategy-alpha"); got != "100" {
		t.Fatalf("shared strategy usage %s, want 100 (A 50 + B 50)", got)
	}

	// A 체결 4/10, 실제 가격 12 → A filled = max(20, 12×4+1) = 48(+HELD 30) = 78; 공유 bucket 합 = 78 + 50 = 128 > 100.
	if err := j.RegisterRiskBucketOrder(ctx, RiskBucketOrderPlan{OrderID: "risk-shared-a", DecisionID: decisionA, OrderQuantity: 10, ReservedMinor: reservedA, ReservationPolicyDigest: "reservation-policy-v1", QuoteCurrency: "KRW", BaseCurrency: "KRW", CreatedAt: riskFillNow}); err != nil {
		t.Fatal(err)
	}
	if _, err := j.RecordFill(ctx, observation("risk-shared-a", "4")); err != nil {
		t.Fatal(err)
	}
	if _, err := j.completeRiskBucketFillActual(ctx, RiskBucketActualFillPlan{Owner: keyA, DecisionID: decisionA, OrderID: "risk-shared-a", CumulativeFill: 4, Actual: riskBucketActual("12", "1", "0"), ObservedAt: riskFillNow}); err != nil {
		t.Fatal(err)
	}
	shared := map[riskbucket.Dimension]string{riskbucket.DimensionHorizon: "SHORT", riskbucket.DimensionMarket: "US", riskbucket.DimensionStrategy: "strategy-alpha", riskbucket.DimensionSector: "sector-tech"}
	assertOverage := func(stage string) {
		t.Helper()
		for dimension, value := range shared {
			// 거래한 owner(A)의 예약에만 overage 28(=128-100)과 RISK_OVERAGE 가 섬 — B 의 예약에는 사본이 없음(판정은 5.6.1 공유
			// latch 차단 한 곳).
			var overage string
			var latched int
			if err := j.db.QueryRow(`SELECT overage_minor,risk_overage_latched FROM risk_bucket_reservations WHERE decision_id=? AND bucket_dimension=?`, decisionA, string(dimension)).Scan(&overage, &latched); err != nil {
				t.Fatal(err)
			}
			if overage != "28" || latched != 1 {
				t.Fatalf("%s: A's %s/%s reservation overage=%s latched=%d, want 28 and latched", stage, dimension, value, overage, latched)
			}
			var bLatched int
			if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_reservations WHERE decision_id=? AND risk_overage_latched=1`, receiptB.DecisionID).Scan(&bLatched); err != nil || bLatched != 0 {
				t.Fatalf("%s: B's reservations carry %d RISK_OVERAGE copies (err=%v); the latch belongs to the owner whose fill caused it", stage, bLatched, err)
			}
		}
		var symbolOverage string
		if err := j.db.QueryRow(`SELECT overage_minor FROM risk_bucket_reservations WHERE decision_id=? AND bucket_dimension='symbol'`, decisionA).Scan(&symbolOverage); err != nil || symbolOverage != "0" {
			t.Fatalf("%s: A's own symbol bucket (78 of 100) recorded overage %s err=%v", stage, symbolOverage, err)
		}
		var ownerLatched int
		if err := j.db.QueryRow(`SELECT risk_overage_latched FROM risk_bucket_owners WHERE prospective_generation=?`, keyA.ProspectiveGeneration).Scan(&ownerLatched); err != nil || ownerLatched != 1 {
			t.Fatalf("%s: A owner RISK_OVERAGE=%d err=%v", stage, ownerLatched, err)
		}
	}
	assertOverage("after A's actual fill")

	// 다른 owner 의 새 진입은 5.6.1 공유 latch 차단으로 막힘(원장 그대로의 snapshot 이어도).
	newEntry := func(suffix, symbol string) error {
		seedExistingRiskReservation(t, j, "existing-"+suffix, "acct-1")
		plan := riskBucketAdmissionFixture(t, suffix, "acct-1", "lane-"+suffix, "campaign-"+suffix, "prospective-"+suffix, "1000", "0")
		plan.ExistingReservationID = "existing-" + suffix
		plan.Owner.Key.Market, plan.Owner.Key.Symbol = riskbucket.MarketUS, symbol
		plan.Admission.Policy.QuoteCurrency, plan.Admission.Policy.AccountCurrency = "USD", "KRW"
		rebindRiskBucket(t, &plan, 1, riskbucket.BucketKey{Dimension: riskbucket.DimensionMarket, Value: "US", PolicyVersion: "policy-v1"})
		rebindRiskBucket(t, &plan, 4, riskbucket.BucketKey{Dimension: riskbucket.DimensionSymbol, Value: symbol, PolicyVersion: "policy-v1"})
		refreshSnapshotUsageFromLedger(t, j, &plan)
		_, err := j.CommitRiskBucketAdmission(ctx, plan)
		return err
	}
	if err := newEntry("shared-c", "GOOG"); !errors.Is(err, ErrRiskBucketEntryBlocked) {
		t.Fatalf("new entry into the over-limit shared buckets was not blocked: %v", err)
	}

	// 반증 시나리오(Manager): B 가 줄어 공유 합이 한도 아래(78)로 내려가도 A 의 latch 는 풀리지 않고 진입은 계속 막힘.
	recordConfirmedFillOrderScopeQuantity(t, j, "shared-b-intent", "shared-b-attempt", "risk-shared-b", FillSnapshotScope{
		AccountRef: "acct-1", Market: "us", TradingDay: "2026-03-30", Symbol: "MSFT", Side: "BUY"}, "10")
	bindRiskOrderAttemptDecision(t, j, "risk-shared-b", receiptB.DecisionID)
	reservedB := riskReservedMap("50")
	delete(reservedB, riskbucket.BucketKey{Dimension: riskbucket.DimensionSymbol, Value: "AAPL", PolicyVersion: "policy-v1"})
	reservedB[riskbucket.BucketKey{Dimension: riskbucket.DimensionSymbol, Value: "MSFT", PolicyVersion: "policy-v1"}] = "50"
	if err := j.RegisterRiskBucketOrder(ctx, RiskBucketOrderPlan{OrderID: "risk-shared-b", DecisionID: receiptB.DecisionID, OrderQuantity: 10, ReservedMinor: reservedB, ReservationPolicyDigest: "reservation-policy-v1", QuoteCurrency: "KRW", BaseCurrency: "KRW", CreatedAt: riskFillNow}); err != nil {
		t.Fatal(err)
	}
	keyB := planB.Owner.Key
	if release, err := j.releaseRiskBucketOrder(ctx, RiskBucketOrderRelease{Owner: keyB, DecisionID: receiptB.DecisionID, OrderID: "risk-shared-b", Reason: RiskBucketReleaseCancel, ReleasedAt: riskFillNow.Add(time.Second)}); err != nil || !release.Released {
		t.Fatalf("B release: %+v err=%v", release, err)
	}
	if got := contractUsage(t, j, riskbucket.DimensionStrategy, "strategy-alpha"); got != "78" {
		t.Fatalf("after B's release the shared strategy usage is %s, want 78 (below the limit)", got)
	}
	assertOverage("after B shrank")
	if err := newEntry("shared-d", "NVDA"); !errors.Is(err, ErrRiskBucketEntryBlocked) {
		t.Fatalf("B shrinking below the limit reopened entry: %v", err)
	}
}

// TestA066SharedUsageUnreadableKeepsTheFillAndLatches: 공유 bucket 의 다른 진입 행이 계약을 벗어나 사용량을 셀 수 없으면
// 체결·watermark 는 보존되고(버리지 않음) 이 owner 에 REPLAY_MISMATCH latch 와 FILL_UNACCOUNTED 가 남음 — "0 으로 치고
// 지나가기"는 공유 overage 를 조용히 놓치는 문임.
func TestA066SharedUsageUnreadableKeepsTheFillAndLatches(t *testing.T) {
	ctx := context.Background()
	j, _, decisionA, reservedA := riskBucketFillFixture(t, "shared-corrupt-a", "risk-shared-corrupt")
	seedExistingRiskReservation(t, j, "existing-corrupt-b", "acct-1")
	planB := riskBucketAdmissionFixture(t, "corrupt-b", "acct-1", "lane-b", "campaign-b", "prospective-corrupt-b", "100", "0")
	planB.ExistingReservationID = "existing-corrupt-b"
	planB.Owner.Key.Market, planB.Owner.Key.Symbol = riskbucket.MarketUS, "MSFT"
	planB.Admission.Policy.QuoteCurrency, planB.Admission.Policy.AccountCurrency = "USD", "KRW"
	rebindRiskBucket(t, &planB, 1, riskbucket.BucketKey{Dimension: riskbucket.DimensionMarket, Value: "US", PolicyVersion: "policy-v1"})
	rebindRiskBucket(t, &planB, 4, riskbucket.BucketKey{Dimension: riskbucket.DimensionSymbol, Value: "MSFT", PolicyVersion: "policy-v1"})
	refreshSnapshotUsageFromLedger(t, j, &planB)
	receiptB, err := j.CommitRiskBucketAdmission(ctx, planB)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`UPDATE risk_bucket_reservations SET held_minor='not-an-amount' WHERE decision_id=? AND bucket_dimension='strategy'`, receiptB.DecisionID); err != nil {
		t.Fatal(err)
	}
	if err := j.RegisterRiskBucketOrder(ctx, RiskBucketOrderPlan{OrderID: "risk-shared-corrupt", DecisionID: decisionA, OrderQuantity: 10, ReservedMinor: reservedA, ReservationPolicyDigest: "reservation-policy-v1", QuoteCurrency: "KRW", BaseCurrency: "KRW", CreatedAt: riskFillNow}); err != nil {
		t.Fatal(err)
	}
	result, err := j.RecordFill(ctx, observation("risk-shared-corrupt", "4"))
	if err != nil || !result.Changed {
		t.Fatalf("an unreadable shared bucket dropped the fill: %+v err=%v", result, err)
	}
	if fill, err := j.LookupFill(ctx, "risk-shared-corrupt"); err != nil || fill.FilledQuantity != "4" {
		t.Fatalf("watermark: %+v err=%v", fill, err)
	}
	var latch, unaccounted int
	if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_scope_latches WHERE latch='REPLAY_MISMATCH' AND symbol='AAPL'`).Scan(&latch); err != nil || latch != 1 {
		t.Fatalf("REPLAY_MISMATCH latch=%d err=%v", latch, err)
	}
	if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_events WHERE event_type='FILL_UNACCOUNTED'`).Scan(&unaccounted); err != nil || unaccounted != 1 {
		t.Fatalf("FILL_UNACCOUNTED events=%d err=%v", unaccounted, err)
	}
}
