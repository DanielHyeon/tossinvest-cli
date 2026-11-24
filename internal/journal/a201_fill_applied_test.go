package journal

// a201_fill_applied_test.go 는 a201 의 새 해제 원인 FILL_APPLIED 의 장부 쪽 계약을 고정함(design V6 · §3-10).
//
//   - 쓰기 시점 허용 목록(ValidReconcileReleaseCause)이 FILL_APPLIED 를 받아 QM 행을 그 원인으로 풀 수 있음.
//   - 위험 버킷 broker-zero 경로는 허용 목록(RECHECK_MATCHED · ADJUSTMENT_APPLIED)을 고치지 않았으므로 FILL_APPLIED 로 풀린
//     행을 「보유 0 의 근거」로 받지 않음(released_reconciliation) — 체결 credit 해제는 보유 0 을 뜻하지 않음.
//
// 편집 전 코드에서도 컴파일되도록 원인은 상수 대신 글자로 씀.

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
)

const a201FillApplied = "FILL_APPLIED"

func TestA201FillAppliedIsAReleaseCauseThisBuildWrites(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()
	if !ValidReconcileReleaseCause(a201FillApplied) {
		t.Fatalf("%s must be a valid release cause", a201FillApplied)
	}
	if _, _, err := j.EnterReconcile(ctx, EnterReconcileRequest{
		AccountRef: "acct-1", Symbol: "005930", Cause: ReconcileCauseQuantityMismatch, Evidence: "quantity mismatch",
	}); err != nil {
		t.Fatalf("EnterReconcile: %v", err)
	}
	state, ok, err := j.ReleaseReconcile(ctx, ReleaseReconcileRequest{
		AccountRef: "acct-1", Symbol: "005930", Cause: a201FillApplied,
		Evidence:    "a fill committed after the mark read before the local read stood in for the adjustment and the reconciliation after it agreed",
		ExpectCause: ReconcileCauseQuantityMismatch,
	})
	if err != nil || !ok || state.ReleaseCause != a201FillApplied {
		t.Fatalf("ReleaseReconcile(%s) = %+v ok=%v err=%v, want the row released with that cause", a201FillApplied, state, ok, err)
	}
}

// TestA201BrokerZeroRefusesAFillAppliedRelease: 같은 모양에서 RECHECK_MATCHED 행은 받아들여지고(대조) FILL_APPLIED 행은
// released_reconciliation 으로 거절됨.
func TestA201BrokerZeroRefusesAFillAppliedRelease(t *testing.T) {
	for _, tc := range []struct {
		cause  string
		refuse bool
	}{
		{cause: ReconcileReleaseRecheckMatched, refuse: false},
		{cause: a201FillApplied, refuse: true},
	} {
		t.Run(tc.cause, func(t *testing.T) {
			j, key := seedRiskBucketOwnerLifecycle(t, riskbucket.MarketUS, "a201-"+tc.cause)
			if _, err := j.bindRiskBucketOwnerActual(context.Background(), key); err != nil {
				t.Fatal(err)
			}
			closeRiskBucketOwnerLifecycle(t, j, key, false)
			if _, err := j.db.Exec(`INSERT INTO reconcile_states(id,account_ref,symbol,cause,evidence,entered_at,released_at,release_cause) VALUES(?,?,?,'QUANTITY_MISMATCH','quantity mismatch','2026-03-30T00:40:00Z','2026-03-30T00:42:00Z',?)`,
				"a201-reconcile-"+key.ProspectiveGeneration, key.AccountID, key.Symbol, tc.cause); err != nil {
				t.Fatal(err)
			}
			result, err := j.recordRiskBucketBrokerZeroObservation(context.Background(),
				officialZeroCapabilityForTest(t, key, "a201-zero-"+key.ProspectiveGeneration, time.Date(2026, 3, 30, 0, 41, 0, 0, time.UTC)))
			if !tc.refuse {
				if err != nil || (!result.Recorded && !result.Idempotent) {
					t.Fatalf("control %s: result=%+v err=%v, want the observation recorded", tc.cause, result, err)
				}
				return
			}
			var lifecycle *RiskBucketOwnerLifecycleError
			if !errors.Is(err, ErrRiskBucketOwnerReleaseBlocked) || !errors.As(err, &lifecycle) ||
				lifecycle.BlockingField != "released_reconciliation" {
				t.Fatalf("%s: result=%+v err=%v lifecycle=%+v, want released_reconciliation", tc.cause, result, err, lifecycle)
			}
		})
	}
}
