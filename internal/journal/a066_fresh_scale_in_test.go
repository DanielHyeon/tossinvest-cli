package journal

// a066 6.1 (B)(1) — BTM 행 뮤테이션이 살아남은 자리: commitFreshRiskBucketAdmissionTx B15(`if ownerReused { … }`).
// q_final 발급·전략 첫 leg 가 같은 owner 에 scale-in 할 때 상태 digest 대조와 bucket 신원 대조를 건너뛰어도 스위트가
// 초록이었음(CommitRiskBucketAdmission 쪽 같은 가드는 TestRiskBucketSameOwnerScaleInRejectsBucketKeyOrPolicyVersionChange 가
// 잼 — 발급 경로에는 없었음). 이 시험은 발급 경로에서 두 가드를 **각자의 문장으로** 고정하고, 대조군(표류 없음)이
// 같은 자리를 지나 성립함을 함께 봄 — 거절이 다른 가드에서 난 것이 아님을 보이기 위해.

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
)

// freshScaleInPair 는 같은 owner(prospective·lane·campaign)로 q_final 을 한 번 발급한 뒤, 같은 owner 의 두 번째 발급
// 요청을 원장과 맞는 snapshot 으로 돌려줌.
func freshScaleInPair(t *testing.T, suffix string) (*Journal, QFinalIssueRequest) {
	t.Helper()
	j := openTestJournal(t)
	first := qFinalIssueFixture(t, j, suffix+"-a")
	if _, err := j.RecordQFinalDecisionAndReserve(context.Background(), first); err != nil {
		t.Fatalf("first issuance: %v", err)
	}
	second := qFinalIssueFixture(t, j, suffix+"-b")
	second.Admission.Owner = first.Admission.Owner
	refreshSnapshotUsageFromLedger(t, j, &second.Admission)
	return j, second
}

func TestA066FreshScaleInReusesTheOwnerOnlyAfterStateAndBucketIdentityMatch(t *testing.T) {
	t.Run("control", func(t *testing.T) {
		j, second := freshScaleInPair(t, "scale-ok")
		result, err := issueSecondWithFreshVersion(t, j, second)
		if err != nil || result.Admission.QFinal == 0 {
			t.Fatalf("same-owner scale-in without drift must be admitted: q_final=%d err=%v", result.Admission.QFinal, err)
		}
		var owners int
		if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_owners WHERE released_at IS NULL`).Scan(&owners); err != nil || owners != 1 {
			t.Fatalf("scale-in must reuse the one owner: owners=%d err=%v", owners, err)
		}
	})
	t.Run("bucket identity drift", func(t *testing.T) {
		j, second := freshScaleInPair(t, "scale-key")
		rebindRiskBucket(t, &second.Admission, 2, riskbucket.BucketKey{Dimension: riskbucket.DimensionStrategy, Value: "strategy-beta", PolicyVersion: "policy-v2"})
		_, err := issueSecondWithFreshVersion(t, j, second)
		if !errors.Is(err, ErrRiskBucketSnapshotMismatch) || !strings.Contains(err.Error(), "scale-in bucket identity") {
			t.Fatalf("scale-in bucket key drift on the issuance path: err=%v", err)
		}
		assertSingleQFinalDecision(t, j)
	})
	t.Run("state digest drift", func(t *testing.T) {
		j, second := freshScaleInPair(t, "scale-digest")
		// 원장 행을 기록된 상태 snapshot 과 어긋나게 만듦 — 재구성한 상태 digest 가 마지막 기록과 달라짐.
		if _, err := j.db.Exec(`UPDATE risk_bucket_reservations SET held_minor='49' WHERE bucket_dimension='horizon'`); err != nil {
			t.Fatal(err)
		}
		refreshSnapshotUsageFromLedger(t, j, &second.Admission)
		_, err := issueSecondWithFreshVersion(t, j, second)
		if !errors.Is(err, ErrRiskBucketReplayMismatch) || !strings.Contains(err.Error(), "state digest") {
			t.Fatalf("scale-in over a drifted owner state on the issuance path: err=%v", err)
		}
		assertSingleQFinalDecision(t, j)
	})
}

func assertSingleQFinalDecision(t *testing.T, j *Journal) {
	t.Helper()
	if got := countRiskBucketRows(t, j, "risk_bucket_final_decisions"); got != 1 {
		t.Fatalf("a refused scale-in wrote a q_final decision: decisions=%d", got)
	}
}
