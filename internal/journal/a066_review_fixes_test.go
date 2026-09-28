package journal

// a066 6.5 적대 리뷰 발견의 수리 시험(Manager 판정 2026-09-28).
//
//   1. 제출 재검증이 latch 를 다시 보지 않았음 — 이미 발급된 q_final 이 owner·scope·공유 bucket 이 latch 된 뒤에도
//      제출됐음(새 admission 은 같은 상태에서 거절). 사용자 결정 ⑤("노출은 제출 시점의 상태")와 같은 모양.
//   3. 공유 bucket 의 한도가 admission 마다 자기 snapshot 의 것이었음 — 더 큰 한도를 선언한 진입이 앞 진입의 한도를
//      넘겨 admit 됨. 수리: 그 bucket 의 활성 예약에 기록된 최소 한도로 cap.
//   4. 상태 snapshot 봉인이 없으면 verifyRiskBucketStateDigest 가 raw sql.ErrNoRows 를 돌려 체결 트랜잭션 전체가 되돌려졌음
//      (체결 감지 차단). 수리: 재구성 갭으로 분류해 latch + 체결 커밋.

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
)

// issuedSecondOnSharedBuckets 는 공유 bucket 에 두 진입을 발급하고(KR 두 종목, 두 번째는 원장과 맞는 snapshot) 두 요청을 돌려줌.
func issuedSecondOnSharedBuckets(t *testing.T) (*Journal, QFinalIssueRequest, QFinalIssueRequest) {
	t.Helper()
	j := openTestJournal(t)
	first, second := sharedBucketPair(t, j, riskbucket.MarketKR, "6")
	if _, err := j.RecordQFinalDecisionAndReserve(context.Background(), first); err != nil {
		t.Fatal(err)
	}
	withSnapshotUsage(t, &second.Admission, true, "50")
	intent := second.Issue.Decision.Preimage.(RiskIntent)
	intent.Quantity = "5"
	second.Issue.Decision.Preimage = intent
	if _, err := issueSecondWithFreshVersion(t, j, second); err != nil {
		t.Fatal(err)
	}
	return j, first, second
}

func TestA066SubmitRevalidationRefusesALatchTakenAfterIssuance(t *testing.T) {
	for _, tc := range []struct {
		name  string
		latch func(*testing.T, *Journal, QFinalIssueRequest, QFinalIssueRequest)
		names []string
	}{
		{"shared bucket overage on another owner", func(t *testing.T, j *Journal, first, _ QFinalIssueRequest) {
			if _, err := j.db.Exec(`UPDATE risk_bucket_reservations SET risk_overage_latched=1 WHERE decision_id=? AND bucket_dimension='sector'`, first.Issue.Decision.ID); err != nil {
				t.Fatal(err)
			}
		}, []string{"RISK_OVERAGE", "sector"}},
		{"shared bucket unknown actual on another owner", func(t *testing.T, j *Journal, first, _ QFinalIssueRequest) {
			if _, err := j.db.Exec(`UPDATE risk_bucket_reservations SET unknown_actual_latched=1 WHERE decision_id=? AND bucket_dimension='horizon'`, first.Issue.Decision.ID); err != nil {
				t.Fatal(err)
			}
		}, []string{"UNKNOWN_ACTUAL_RISK", "horizon"}},
		{"active reconcile on the entry symbol", func(t *testing.T, j *Journal, _, second QFinalIssueRequest) {
			key := second.Admission.Owner.Key
			if _, entered, err := j.EnterReconcile(context.Background(), EnterReconcileRequest{AccountRef: key.AccountID, Symbol: key.Symbol,
				Cause: ReconcileCauseQuantityMismatch, Evidence: "a066 6.5 submit revalidation probe"}); err != nil || !entered {
				t.Fatalf("enter reconcile: entered=%v err=%v", entered, err)
			}
		}, []string{"RECONCILE"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			j, first, second := issuedSecondOnSharedBuckets(t)
			tc.latch(t, j, first, second)
			required, err := j.RevalidateQFinalAdmission(context.Background(), second.Issue.Decision.ID)
			if !required || !errors.Is(err, ErrRiskBucketEntryBlocked) {
				t.Fatalf("submit after the latch: required=%v err=%v, want entry blocked", required, err)
			}
			for _, name := range tc.names {
				if !strings.Contains(err.Error(), name) {
					t.Fatalf("refusal %q does not name %q", err, name)
				}
			}
		})
	}
	// 대조군: latch 가 없으면 같은 결정은 재검증을 통과함(거절이 이 자리에서 난 것임을 보이기 위해).
	j, _, second := issuedSecondOnSharedBuckets(t)
	if required, err := j.RevalidateQFinalAdmission(context.Background(), second.Issue.Decision.ID); err != nil || !required {
		t.Fatalf("control: required=%v err=%v", required, err)
	}
}

func TestA066SharedBucketAdmissionIsCappedByTheSmallestRecordedLimit(t *testing.T) {
	j := openTestJournal(t)
	first, second := sharedBucketPair(t, j, riskbucket.MarketKR, "6")
	if _, err := j.RecordQFinalDecisionAndReserve(context.Background(), first); err != nil {
		t.Fatal(err)
	}
	// 둘째 진입은 같은 bucket 에 더 큰 한도(1000)를 선언함 — 첫 진입이 기록한 80 을 넘기면 안 됨.
	for i := range second.Admission.Admission.Buckets {
		rebindRiskBucketLimit(t, &second.Admission, i, second.Admission.Admission.Buckets[i].Key, "1000")
	}
	withSnapshotUsage(t, &second.Admission, true, "50")
	_, err := issueSecondWithFreshVersion(t, j, second)
	if !riskbucket.IsRefusal(err, riskbucket.RefusalBucketCapExhausted) || !strings.Contains(err.Error(), "80") {
		t.Fatalf("a larger declared limit must be capped by the recorded 80: err=%v", err)
	}
	if got := countRiskBucketRows(t, j, "risk_bucket_final_decisions"); got != 1 {
		t.Fatalf("decisions=%d", got)
	}
	// 대조군: 기록된 한도 안이면(원장 50 + 둘째 30 = 80) 같은 선언으로도 admit 됨.
	j2 := openTestJournal(t)
	first2, second2 := sharedBucketPair(t, j2, riskbucket.MarketKR, "6")
	if _, err := j2.RecordQFinalDecisionAndReserve(context.Background(), first2); err != nil {
		t.Fatal(err)
	}
	for i := range second2.Admission.Admission.Buckets {
		rebindRiskBucketLimit(t, &second2.Admission, i, second2.Admission.Admission.Buckets[i].Key, "1000")
	}
	withSnapshotUsage(t, &second2.Admission, true, "50")
	intent := second2.Issue.Decision.Preimage.(RiskIntent)
	intent.Quantity = "5"
	second2.Issue.Decision.Preimage = intent
	second2.Admission.Admission.QCandidate = 5
	if _, err := issueSecondWithFreshVersion(t, j2, second2); err != nil {
		t.Fatalf("control within the recorded limit: %v", err)
	}
}

func TestA066FillWithAMissingStateSealLatchesAndKeepsTheFill(t *testing.T) {
	ctx := context.Background()
	j, key, decisionID, reserved := riskBucketFillFixture(t, "missing-seal", "risk-missing-seal")
	if err := j.RegisterRiskBucketOrder(ctx, RiskBucketOrderPlan{OrderID: "risk-missing-seal", DecisionID: decisionID, OrderQuantity: 10, ReservedMinor: reserved, CreatedAt: riskFillNow}); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`DELETE FROM risk_bucket_state_snapshots WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=?`,
		key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration); err != nil {
		t.Fatalf("remove the seal: %v", err)
	}
	result, err := j.RecordFill(ctx, observation("risk-missing-seal", "2"))
	if err != nil || !result.Changed {
		t.Fatalf("the broker fill must be kept when the state seal is missing: result=%+v err=%v", result, err)
	}
	var latches int
	if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_scope_latches WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=? AND latch='REPLAY_MISMATCH'`,
		key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration).Scan(&latches); err != nil || latches != 1 {
		t.Fatalf("REPLAY_MISMATCH latches=%d err=%v", latches, err)
	}
}

// TestA066RecordedLimitCapUsesTheSmallestOfSeveral 은 기록된 한도가 여럿일 때 cap 이 **가장 작은 값**임을 고정함(변이
// "가장 큰 값"이 한 값만 기록된 시험을 통과했음 — mutation-6.5 L03). 첫 진입 한도 100(50 예약), 둘째 80(30 예약)으로
// 기록한 뒤, 셋째가 1000 을 선언해도 80 에서 막힘(원장 80 + 셋째 예약 > 80); 가장 큰 값(100)이면 셋째 10 이 들어감.
func TestA066RecordedLimitCapUsesTheSmallestOfSeveral(t *testing.T) {
	j := openTestJournal(t)
	first, second := sharedBucketPair(t, j, riskbucket.MarketKR, "6")
	for i := range first.Admission.Admission.Buckets {
		rebindRiskBucketLimit(t, &first.Admission, i, first.Admission.Admission.Buckets[i].Key, "100")
	}
	if _, err := j.RecordQFinalDecisionAndReserve(context.Background(), first); err != nil {
		t.Fatal(err)
	}
	withSnapshotUsage(t, &second.Admission, true, "50")
	intent := second.Issue.Decision.Preimage.(RiskIntent)
	intent.Quantity = "5"
	second.Issue.Decision.Preimage = intent
	if _, err := issueSecondWithFreshVersion(t, j, second); err != nil {
		t.Fatalf("second within 80: %v", err)
	}
	third := qFinalIssueFixture(t, j, "shared-c")
	third.Admission.Owner.Key.Symbol = "035420"
	rebindRiskBucket(t, &third.Admission, 4, riskbucket.BucketKey{Dimension: riskbucket.DimensionSymbol, Value: "035420", PolicyVersion: "policy-v1"})
	for i := range third.Admission.Admission.Buckets {
		rebindRiskBucketLimit(t, &third.Admission, i, third.Admission.Admission.Buckets[i].Key, "1000")
	}
	third.Admission.Admission.Policy.Price.WorstExecutableQuote = "1"
	third.Admission.Admission.QCandidate = 10
	thirdIntent := third.Issue.Decision.Preimage.(RiskIntent)
	thirdIntent.Symbol, thirdIntent.Quantity = "035420", "10"
	third.Issue.Decision.Preimage = thirdIntent
	withSnapshotUsage(t, &third.Admission, true, "80")
	_, err := issueSecondWithFreshVersion(t, j, third)
	if !riskbucket.IsRefusal(err, riskbucket.RefusalBucketCapExhausted) || !strings.Contains(err.Error(), "limit 80") {
		t.Fatalf("third entry must be capped by the smallest recorded limit 80, not 100: err=%v", err)
	}
}

// TestA066AdmissionRefusalNamesAnOwnerOverageLatch 는 규칙 함수가 owner RISK_OVERAGE 를 이름으로 말하며 막음을 고정함
// (변이 "owner overage 무시"가 살아남았음 — mutation-6.5 V09). admission 은 owner 조회보다 앞에서 이 규칙을 부름.
func TestA066AdmissionRefusalNamesAnOwnerOverageLatch(t *testing.T) {
	j := openTestJournal(t)
	seedExistingRiskReservation(t, j, "existing-owner-overage-a", "acct-1")
	first := riskBucketAdmissionFixture(t, "owner-overage-a", "acct-1", "lane-short", "campaign-1", "prospective-owner-overage", "300", "0")
	if _, err := j.CommitRiskBucketAdmission(context.Background(), first); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`UPDATE risk_bucket_owners SET risk_overage_latched=1 WHERE prospective_generation='prospective-owner-overage'`); err != nil {
		t.Fatal(err)
	}
	seedExistingRiskReservation(t, j, "existing-owner-overage-b", "acct-1")
	second := riskBucketAdmissionFixture(t, "owner-overage-b", "acct-1", "lane-short", "campaign-1", "prospective-owner-overage", "300", "50")
	_, err := j.CommitRiskBucketAdmission(context.Background(), second)
	if !errors.Is(err, ErrRiskBucketEntryBlocked) || !strings.Contains(err.Error(), "owner RISK_OVERAGE") {
		t.Fatalf("owner overage latch: err=%v", err)
	}
}

// TestA066RecordedLimitIncludesReleasedRowsThatStillCountFilledUsage 는 좁힌 재리뷰 P2 를 고정함: 부분 체결 뒤 취소된
// 예약은 held 0 으로 RELEASED 가 되지만 filled 는 원장 사용량에 계속 셈 — 그 행의 한도도 기록된 최소 한도에 들어가야 함.
// 모양은 실제 경로(주문 10 등록 → 2 체결 → CANCEL 해제: 다섯 예약 RELEASED held=0 filled>0, 재리뷰 실측)와 같게 둠.
func TestA066RecordedLimitIncludesReleasedRowsThatStillCountFilledUsage(t *testing.T) {
	j := openTestJournal(t)
	first, second := sharedBucketPair(t, j, riskbucket.MarketKR, "6")
	if _, err := j.RecordQFinalDecisionAndReserve(context.Background(), first); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`UPDATE risk_bucket_reservations SET state='RELEASED',filled_minor=held_minor,held_minor='0' WHERE decision_id=?`, first.Issue.Decision.ID); err != nil {
		t.Fatal(err)
	}
	for i := range second.Admission.Admission.Buckets {
		rebindRiskBucketLimit(t, &second.Admission, i, second.Admission.Admission.Buckets[i].Key, "1000")
	}
	refreshSnapshotUsageFromLedger(t, j, &second.Admission)
	_, err := issueSecondWithFreshVersion(t, j, second)
	if !riskbucket.IsRefusal(err, riskbucket.RefusalBucketCapExhausted) || !strings.Contains(err.Error(), "limit 80") {
		t.Fatalf("a released row with filled usage must keep its recorded limit 80: err=%v", err)
	}
}

// TestA066EntryScopeRuleBlocksOnEveryActiveReconcileRow 는 좁힌 재리뷰 P3 를 고정함: 규칙 함수는 원인 문자열이 아니라
// 활성 행의 존재로 막아야 함(원인이 빈 문자열인 옛 행에서도 — 예전 count 판정과 같게).
func TestA066EntryScopeRuleBlocksOnEveryActiveReconcileRow(t *testing.T) {
	j := openTestJournal(t)
	if _, err := j.db.Exec(`INSERT INTO reconcile_states (id, account_ref, symbol, cause, evidence, entered_at, released_at, release_cause) VALUES ('legacy-empty-cause','acct-1','005930','','legacy row','2026-03-30T00:00:00Z',NULL,NULL)`); err != nil {
		t.Fatal(err)
	}
	tx, err := j.db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	err = ensureRiskBucketEntryScopeClean(context.Background(), tx, riskBucketOwnerKey("acct-1", "p"))
	if !errors.Is(err, ErrRiskBucketEntryBlocked) || !strings.Contains(err.Error(), "RECONCILE") {
		t.Fatalf("an active reconcile row with an empty cause must block: err=%v", err)
	}
}

// TestA066SubmitRevalidationRefusesOwnerAndScopeLatches 는 좁힌 재리뷰 P3(시험 공백)를 채움: 제출 재검증의 규칙 함수
// 자리에서 owner latch 와 scope latch 가 bucket 예약 latch 없이도 막음. owner latch 는 체결 경로처럼 상태 봉인을 다시
// 찍은 뒤에 둠(그러지 않으면 앞의 상태 digest 대조가 먼저 막아 이 자리에 닿지 않음).
func TestA066SubmitRevalidationRefusesOwnerAndScopeLatches(t *testing.T) {
	for _, tc := range []struct {
		name, want string
		latch      func(*testing.T, *Journal, riskbucket.OwnerKey)
	}{
		{"owner RISK_OVERAGE after a resealed fill", "owner RISK_OVERAGE", func(t *testing.T, j *Journal, key riskbucket.OwnerKey) {
			tx, err := j.db.BeginTx(context.Background(), nil)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := tx.Exec(`UPDATE risk_bucket_owners SET risk_overage_latched=1 WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=?`,
				key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration); err != nil {
				t.Fatal(err)
			}
			if err := j.recordRiskBucketStateTx(context.Background(), tx, key, "TEST_RESEAL", "a066-6.5-reseal", "reseal", "2026-03-30T00:40:00Z"); err != nil {
				t.Fatal(err)
			}
			if err := tx.Commit(); err != nil {
				t.Fatal(err)
			}
		}},
		{"scope ORPHAN_FILL latch", "scope latch ORPHAN_FILL", func(t *testing.T, j *Journal, key riskbucket.OwnerKey) {
			if _, err := j.db.Exec(`INSERT INTO risk_bucket_scope_latches(account_ref,market,symbol,prospective_generation,latch,detail,first_seen_at,last_seen_at) VALUES(?,?,?,?,'ORPHAN_FILL','a066 6.5 probe','2026-03-30T00:40:00Z','2026-03-30T00:40:00Z')`,
				key.AccountID, string(key.Market), key.Symbol, "prior-generation"); err != nil {
				t.Fatal(err)
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			j, _, second := issuedSecondOnSharedBuckets(t)
			tc.latch(t, j, second.Admission.Owner.Key)
			required, err := j.RevalidateQFinalAdmission(context.Background(), second.Issue.Decision.ID)
			if !required || !errors.Is(err, ErrRiskBucketEntryBlocked) || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("submit under %s: required=%v err=%v", tc.name, required, err)
			}
		})
	}
}
