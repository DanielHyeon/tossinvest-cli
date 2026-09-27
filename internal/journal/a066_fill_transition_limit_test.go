package journal

// a066 6.1 (B)(2) — loadRiskBucketFillTransition B12: 한 owner 의 여러 결정이 같은 bucket 에 다른 snapshot 한도를 들고
// 있으면 체결 계상의 한도는 **가장 작은 값**(보수 방향)임. 이 분기는 어떤 시험도 실행하지 않았음 — 기존 scale-in 시험은
// 모두 뒤 결정의 한도가 크거나 같았음. 순서(owner_sequence)대로 100 → 80 을 주면 80 이, 80 → 100 을 주면 여전히 80 이
// 되어야 함(앞 결정이 작으면 뒤의 큰 값이 한도를 넓히지 못함).

import (
	"context"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
)

func TestA066FillTransitionUsesTheSmallestDecisionLimitPerBucket(t *testing.T) {
	for _, tc := range []struct {
		name, firstLimit, scaleInLimit string
	}{
		{"later decision tighter", "100", "80"},
		{"earlier decision tighter", "80", "100"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			suffix := "limit-min-" + tc.firstLimit + "-" + tc.scaleInLimit
			j := openTestJournal(t)
			key, firstDecision, firstReserved := seedRiskBucketFillFixtureWithLimit(t, j, suffix, "risk-"+suffix, tc.firstLimit)
			if err := j.RegisterRiskBucketOrder(ctx, RiskBucketOrderPlan{OrderID: "risk-" + suffix, DecisionID: firstDecision, OrderQuantity: 10, ReservedMinor: firstReserved, CreatedAt: riskFillNow}); err != nil {
				t.Fatal(err)
			}
			commitRiskBucketScaleIn(t, j, key, suffix+"-second", tc.scaleInLimit, "50")
			tx, err := j.db.BeginTx(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback()
			order, found, err := riskBucketOrderByID(ctx, tx, key, firstDecision, "risk-"+suffix)
			if err != nil || !found {
				t.Fatalf("order record found=%v err=%v", found, err)
			}
			state, _, err := loadRiskBucketFillTransition(ctx, tx, order, "1", nil)
			if err != nil {
				t.Fatal(err)
			}
			if len(state.Buckets) != 5 {
				t.Fatalf("buckets=%d", len(state.Buckets))
			}
			for bucket, usage := range state.Buckets {
				if usage.LimitMinor != "80" {
					t.Fatalf("%s limit=%s, want the smaller decision limit 80", bucket.Dimension, usage.LimitMinor)
				}
			}
		})
	}
}

// seedRiskBucketFillFixtureWithLimit 는 seedRiskBucketFillFixture 와 같되 첫 결정의 bucket 한도를 정함.
func seedRiskBucketFillFixtureWithLimit(t *testing.T, j *Journal, suffix, orderID, limit string) (riskbucket.OwnerKey, string, map[riskbucket.BucketKey]string) {
	t.Helper()
	seedExistingRiskReservation(t, j, "existing-fill-"+suffix, "acct-1")
	plan := riskBucketAdmissionFixture(t, "fill-"+suffix, "acct-1", "lane-short", "campaign-1", "prospective-fill-"+suffix, limit, "0")
	plan.ExistingReservationID = "existing-fill-" + suffix
	plan.Owner.Key.Market = riskbucket.MarketUS
	plan.Owner.Key.Symbol = "AAPL"
	plan.Admission.Policy.QuoteCurrency = "USD"
	plan.Admission.Policy.AccountCurrency = "KRW"
	rebindRiskBucket(t, &plan, 1, riskbucket.BucketKey{Dimension: riskbucket.DimensionMarket, Value: string(riskbucket.MarketUS), PolicyVersion: "policy-v1"})
	rebindRiskBucket(t, &plan, 4, riskbucket.BucketKey{Dimension: riskbucket.DimensionSymbol, Value: "AAPL", PolicyVersion: "policy-v1"})
	receipt, err := j.CommitRiskBucketAdmission(context.Background(), plan)
	if err != nil {
		t.Fatal(err)
	}
	recordConfirmedFillOrder(t, j, "risk-intent-"+suffix, "risk-attempt-"+suffix, orderID)
	bindRiskOrderAttemptDecision(t, j, orderID, receipt.DecisionID)
	return plan.Owner.Key, receipt.DecisionID, riskReservedMap("50")
}

// TestA066RevalidateReportsNotRequiredForNonQFinalDecisions 는 RevalidateQFinalAdmission B3·B4 를 고정함: q_final 표식이
// 없는 결정은 "q_final 권위가 필요 없음"(false, nil)으로 답하고, 그 뒤의 bucket·잠금 판정을 하지 않음. Gateway 의
// checkReservation 은 이 답을 받으면 기존 집계 예약 검사만으로 진행함 — a066 이전 결정과 위험 감소 결정이 a066 판정에
// 끌려 들어가지 않는다는 경계(토글 OFF = upstream 동일).
func TestA066RevalidateReportsNotRequiredForNonQFinalDecisions(t *testing.T) {
	ctx := context.Background()
	// B3 를 끄면 RiskIntent 가 아닌 결정은 영값 RiskIntent 로 B4 에 닿고, 빈 정책 버전은 q_final 표식이 아니므로 같은
	// 답(false, nil)이 남(변이 "B3 skipped" 생존 — mutation-6.1/revalidate-ledger.tsv). 그 백스톱의 전제를 못 박음.
	if _, _, required := splitQFinalPolicyVersion(""); required {
		t.Fatal("an empty policy version must not carry q_final authority (B3's backstop)")
	}
	for _, tc := range []struct {
		name    string
		request func(*testing.T, *Journal) DecisionRequest
	}{
		// B3: RiskIntent 가 아닌 preimage(위험 감소 결정).
		{"reduction preimage", func(t *testing.T, _ *Journal) DecisionRequest { return reductionRequest(t) }},
		// B4: RiskIntent 이지만 q_final 정책 버전 표식이 없음(a066 이전 진입 결정).
		{"legacy risk intent", func(t *testing.T, _ *Journal) DecisionRequest { return riskRequest(t) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			j := openTestJournal(t)
			decision, err := j.RecordDecision(ctx, tc.request(t, j))
			if err != nil {
				t.Fatal(err)
			}
			// 잠금을 켜 두어도 필요 없음 판정이 먼저라 잠금 판정까지 가지 않음.
			for _, market := range []riskbucket.Market{riskbucket.MarketKR, riskbucket.MarketUS} {
				for _, horizon := range []riskbucket.Horizon{riskbucket.HorizonShort, riskbucket.HorizonMedium} {
					if _, _, err := j.ActivateEntryLossLock(ctx, EntryLossLock{AccountRef: "acct-1", Market: market, Horizon: horizon, Cause: "a066 not-required probe", ActivatedAt: riskFillNow}); err != nil {
						t.Fatal(err)
					}
				}
			}
			required, err := j.RevalidateQFinalAdmission(ctx, decision.ID)
			if err != nil || required {
				t.Fatalf("non-q_final decision: required=%v err=%v, want false/nil", required, err)
			}
		})
	}
}
