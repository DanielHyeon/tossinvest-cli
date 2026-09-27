package journal

// a066 6.x 하위 발견(2026-09-28, 6.1 (B)(3) 저장 출구 census 가 드러냄): applyRiskBucketOwnerBindingInTx 는 BUY 체결의
// 증분(fill.Delta)을 해석하지 못하면 owner 결속도 latch 도 없이 nil 을 돌려줬음 — 함수 머리 주석("semantic gaps latch
// entry for this owner")과 반대로, 등록된 위험 주문의 체결이 계상 흔적 없이 지나감. 이 시험은 그 갭이 같은 트랜잭션에서
// 형제 의미 갭(다중 활성 scope · 결속 거절)과 같은 방향(REPLAY_MISMATCH scope latch + unknown-actual latch)으로
// 막히고, 체결 자체와 exit 훅은 막히지 않음을 고정함.

import (
	"context"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

func TestA066OwnerBindUnreadableFillDeltaLatchesEntryWithoutBlockingTheFill(t *testing.T) {
	for _, delta := range []string{"not-a-quantity", "-1"} {
		t.Run(delta, func(t *testing.T) {
			j, key := seedRiskBucketOwnerLifecycle(t, riskbucket.MarketUS, "delta-gap")
			var decisionID string
			if err := j.db.QueryRow(`SELECT decision_id FROM risk_bucket_final_decisions WHERE owner_prospective_generation=?`, key.ProspectiveGeneration).Scan(&decisionID); err != nil {
				t.Fatal(err)
			}
			recordConfirmedFillOrderScope(t, j, "delta-gap-intent", "delta-gap-attempt", "delta-gap-order", FillSnapshotScope{
				AccountRef: key.AccountID, Market: "us", TradingDay: "2026-03-30", Symbol: key.Symbol, Side: "BUY",
			})
			bindRiskOrderAttemptDecision(t, j, "delta-gap-order", decisionID)
			if err := j.RegisterRiskBucketOrder(context.Background(), RiskBucketOrderPlan{
				OrderID: "delta-gap-order", DecisionID: decisionID, OrderQuantity: 10,
				ReservedMinor: riskReservedMap("50"), CreatedAt: riskFillNow,
			}); err != nil {
				t.Fatal(err)
			}
			exitRan := false
			if err := j.SetApplyHooks(ApplyHooks{
				Campaign: func(context.Context, *ApplyTx, AppliedFill) error { return nil },
				Exit: func(context.Context, *ApplyTx, AppliedFill) error {
					exitRan = true
					return nil
				},
			}); err != nil {
				t.Fatal(err)
			}
			tx, err := j.db.BeginTx(context.Background(), nil)
			if err != nil {
				t.Fatal(err)
			}
			fill := AppliedFill{OrderID: "delta-gap-order", AccountRef: key.AccountID, Market: "us", Symbol: key.Symbol,
				Side: "BUY", Delta: delta, CumulativeQuantity: "1", TradingDay: "2026-03-30", CommittedAt: "2026-03-30T00:32:00Z"}
			if err := j.runApplyHooks(context.Background(), tx, fill); err != nil {
				_ = tx.Rollback()
				t.Fatalf("an unreadable fill delta must not reject the broker fill: %v", err)
			}
			if err := tx.Commit(); err != nil {
				t.Fatal(err)
			}
			if !exitRan {
				t.Fatal("the delta gap blocked the exit hook")
			}
			var ownerLatched, scopeLatches, unaccounted, bound int
			if err := j.db.QueryRow(`SELECT unknown_actual_latched, actual_generation IS NOT NULL FROM risk_bucket_owners WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=?`,
				key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration).Scan(&ownerLatched, &bound); err != nil {
				t.Fatal(err)
			}
			if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_scope_latches WHERE account_ref=? AND market=? AND symbol=?`,
				key.AccountID, string(key.Market), key.Symbol).Scan(&scopeLatches); err != nil {
				t.Fatal(err)
			}
			if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_events WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=? AND event_type='FILL_UNACCOUNTED'`,
				key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration).Scan(&unaccounted); err != nil {
				t.Fatal(err)
			}
			if ownerLatched != 1 || scopeLatches == 0 || unaccounted != 1 || bound != 0 {
				t.Fatalf("delta %q: owner unknown_actual_latched=%d scope latches=%d FILL_UNACCOUNTED=%d bound=%d, want 1/>0/1/0",
					delta, ownerLatched, scopeLatches, unaccounted, bound)
			}
		})
	}
}

// FuzzA066CampaignQuantityIsComparable 은 applyRiskBucketOwnerBindingInTx 의 CompareDecimal 오류 갈래가 닿지 않는다는
// 전제를 못 박음: campaignQuantity 가 받아들인 값(riskcalc.CanonicalDecimal 의 출력, 음수 아님)은 CompareDecimal 이
// 언제나 해석함. 전제가 깨지면(riskcalc 가 바뀌면) 그 갈래가 살아나고 — 그때도 latch 쪽이므로 안전 방향임.
func FuzzA066CampaignQuantityIsComparable(f *testing.F) {
	for _, seed := range []string{"0", "1", "0.5", "00012.3400", "-1", "-0", "abc", "", "1e3", ".5", "5.", "+3", "1_000"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, raw string) {
		delta, err := campaignQuantity(raw)
		if err != nil {
			return
		}
		if _, err := riskcalc.CompareDecimal(delta, "0"); err != nil {
			t.Fatalf("campaignQuantity(%q)=%q is not comparable: %v", raw, delta, err)
		}
	})
}

// TestA066OwnerBindZeroDeltaIsANoOpNotAGap 는 수리가 거부할 정상 입력이 없음을 고정함: 증분 0(같은 누적량 재관측)은
// 의미 갭이 아니므로 latch 도 결속도 없이 지나가야 함(fail-closed 는 무엇을 거부하는지 말해야 함).
func TestA066OwnerBindZeroDeltaIsANoOpNotAGap(t *testing.T) {
	j, key := seedRiskBucketOwnerLifecycle(t, riskbucket.MarketUS, "delta-zero")
	var decisionID string
	if err := j.db.QueryRow(`SELECT decision_id FROM risk_bucket_final_decisions WHERE owner_prospective_generation=?`, key.ProspectiveGeneration).Scan(&decisionID); err != nil {
		t.Fatal(err)
	}
	recordConfirmedFillOrderScope(t, j, "delta-zero-intent", "delta-zero-attempt", "delta-zero-order", FillSnapshotScope{
		AccountRef: key.AccountID, Market: "us", TradingDay: "2026-03-30", Symbol: key.Symbol, Side: "BUY",
	})
	bindRiskOrderAttemptDecision(t, j, "delta-zero-order", decisionID)
	if err := j.RegisterRiskBucketOrder(context.Background(), RiskBucketOrderPlan{
		OrderID: "delta-zero-order", DecisionID: decisionID, OrderQuantity: 10,
		ReservedMinor: riskReservedMap("50"), CreatedAt: riskFillNow,
	}); err != nil {
		t.Fatal(err)
	}
	tx, err := j.db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	fill := AppliedFill{OrderID: "delta-zero-order", AccountRef: key.AccountID, Market: "us", Symbol: key.Symbol,
		Side: "BUY", Delta: "0", CumulativeQuantity: "1", TradingDay: "2026-03-30", CommittedAt: "2026-03-30T00:32:00Z"}
	if err := j.applyRiskBucketOwnerBindingInTx(context.Background(), tx, fill); err != nil {
		_ = tx.Rollback()
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	var ownerLatched, scopeLatches int
	if err := j.db.QueryRow(`SELECT unknown_actual_latched FROM risk_bucket_owners WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=?`,
		key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration).Scan(&ownerLatched); err != nil {
		t.Fatal(err)
	}
	if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_scope_latches WHERE account_ref=? AND market=? AND symbol=?`,
		key.AccountID, string(key.Market), key.Symbol).Scan(&scopeLatches); err != nil {
		t.Fatal(err)
	}
	if ownerLatched != 0 || scopeLatches != 0 {
		t.Fatalf("zero delta latched: owner=%d scope latches=%d", ownerLatched, scopeLatches)
	}
}
