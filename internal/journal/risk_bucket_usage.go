package journal

// 이 파일은 a066 task 5.6.1 F1 의 수리임: admission 이 호출자가 넘긴 bucket snapshot 의 사용량을 원장과 대조함.
//
// 왜 필요한가. 생산 첫 leg loader(internal/app/engine/strategy_account_first_leg_authority.go)는 bucket snapshot 을 주기
// 앞에서 한 번 모으고, 예약 버전(ObservedVersion)만 발급 시점 collect 에서 새로 읽음. 그래서 같은 주기의 두 번째 진입은
// 최신 버전과 **첫 진입 이전** 사용량을 함께 들고 오고, 계좌 전체 예약 버전 검사를 통과해 공유 cap 을 넘겼음
// (2026-09-27 실측: 한도 80 에 사용량 100/110).
//
// 규칙은 상태가 사는 자리(admission 트랜잭션)에 섬: 같은 BEGIN IMMEDIATE 안에서 bucket 마다 원장 사용량을 다시 셈 —
// 생산 snapshot reader 와 **같은 함수**(riskbucket.ReadJournalBucketUsage)로. snapshot 이 원장보다 적게 주장하면 stale 로
// 거절함(ErrSnapshotStale 을 감싸 기존 재수집 루프가 다시 모음). 많게 주장하는 것은 보수 방향이라 받음.

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
)

// refuseStaleBucketUsage 는 공유 bucket 사용량 대조의 **유일한** 판정 규칙임. 두 admission 트랜잭션
// (CommitRiskBucketAdmission, commitFreshRiskBucketAdmissionTx)이 부름.
//
// 원장을 읽지 못하거나 원장 행이 잘못되면 거절함(진입 경로에서 "모름"은 "막힘"). latch 는 여기서 판정하지 않음 —
// latch 의 진입 차단은 ensureRiskBucketEntryScopeClean 의 몫이고, 여기서 막으면 차단 범위가 조용히 넓어짐.
func refuseStaleBucketUsage(ctx context.Context, tx *sql.Tx, account string, buckets []riskbucket.BucketSnapshot) error {
	for _, bucket := range buckets {
		usage, err := riskbucket.ReadJournalBucketUsage(ctx, tx, account, bucket.Key.Dimension, bucket.Key.Value)
		if err != nil {
			return fmt.Errorf("%w: %s bucket %q ledger usage unreadable: %v", ErrRiskBucketSnapshotMismatch, bucket.Key.Dimension, bucket.Key.Value, err)
		}
		ledger, ok := sumMinor(usage.FilledMinor, usage.HeldMinor)
		if !ok {
			return fmt.Errorf("%w: %s bucket %q ledger usage is not an amount", ErrRiskBucketSnapshotMismatch, bucket.Key.Dimension, bucket.Key.Value)
		}
		claimed, ok := sumMinor(bucket.FilledMinor, bucket.HeldMinor)
		if !ok {
			return fmt.Errorf("%w: %s bucket %q snapshot usage is not an amount", ErrRiskBucketSnapshotMismatch, bucket.Key.Dimension, bucket.Key.Value)
		}
		if claimed.Cmp(ledger) < 0 {
			return &riskbucket.RefusalError{Code: riskbucket.RefusalBucketUsageStale, Field: string(bucket.Key.Dimension),
				Cause: fmt.Errorf("%w: %s bucket %q snapshot claims %s used, the ledger holds %s", ErrRiskBucketUsageStale,
					bucket.Key.Dimension, bucket.Key.Value, claimed, ledger)}
		}
	}
	return nil
}

// sumMinor 는 음이 아닌 minor 단위 정수 둘의 합임.
func sumMinor(a, b string) (*big.Int, bool) {
	x, okA := new(big.Int).SetString(a, 10)
	y, okB := new(big.Int).SetString(b, 10)
	if !okA || !okB || x.Sign() < 0 || y.Sign() < 0 {
		return nil, false
	}
	return x.Add(x, y), true
}
