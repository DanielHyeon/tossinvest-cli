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
// 거절함 — 재시도하지 않음: 같은 wave 의 공유 bucket 2번째 시장은 그 wave 에서 거절되고 다음 wave 에서 원장에서 새로 모은
// snapshot 으로 성립함(진입 경로의 1주기 지연, fail-closed). 많게 주장하는 것은 보수 방향이라 받음.

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
)

// refuseStaleBucketUsage 는 공유 bucket 사용량 대조의 **유일한** 판정 규칙임. 두 admission 트랜잭션
// (CommitRiskBucketAdmission, commitFreshRiskBucketAdmissionTx)이 부름.
//
// 원장을 읽지 못하거나 원장 행이 잘못되면 거절함(진입 경로에서 "모름"은 "막힘"). 공유 bucket 에 latch 된 사용량이
// 있으면 진입 차단(적대 리뷰 5.6.1: owner 범위만 보던 ensureRiskBucketEntryScopeClean 은 다른 종목의 latch 를 못 봄).
func refuseStaleBucketUsage(ctx context.Context, tx *sql.Tx, account string, buckets []riskbucket.BucketSnapshot, caps []riskbucket.BucketCap) error {
	if len(caps) != len(buckets) {
		return fmt.Errorf("%w: bucket caps do not align with snapshots", ErrRiskBucketSnapshotMismatch)
	}
	for i, bucket := range buckets {
		usage, err := riskbucket.ReadJournalBucketUsage(ctx, tx, account, bucket.Key.Dimension, bucket.Key.Value)
		if err != nil {
			return fmt.Errorf("%w: %s bucket %q ledger usage unreadable: %v", ErrRiskBucketSnapshotMismatch, bucket.Key.Dimension, bucket.Key.Value, err)
		}
		// latch 된 사용량(UNKNOWN_ACTUAL_RISK·RISK_OVERAGE)은 원장 합이 실제 노출을 **다 담지 못한다**는 표시임 —
		// UNKNOWN 이면 filled 는 이전 하한일 뿐이고 실제 가격이 더 높을 수 있음. 그 bucket 에 새 노출을 더하면 합이 cap 안이어도
		// 실제로는 넘을 수 있으므로 거절함(설계 D5: latch 는 모든 적용 bucket 의 신규 노출을 막음; 생산 snapshot reader 도
		// 같은 행에서 거절함). 재수집으로 풀리지 않으므로 stale 이 아니라 진입 차단으로 돌려줌.
		if err := latchedUsageRefusal(bucket.Key.Dimension, bucket.Key.Value, usage); err != nil {
			return err
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
		// a066 6.5: 공유 bucket 의 한도는 진입마다 자기 snapshot 이 들고 오는 값이라, 더 큰 한도를 선언한 진입이 앞 진입의
		// 한도를 넘길 수 있었음. 그 bucket 의 활성(HELD·FILLED) 예약이 기록한 snapshot 한도 중 **가장 작은 값**으로 cap 함
		// (보수 방향). 한도 하나로의 단일화는 상류 매니페스트 검증의 몫(잔여, 정책 불변성 잔여와 같은 가족). 한도 모집단은 filled 가
		// 남은 RELEASED 행과 **떠난 행을 포함한다**(a126 D3) — 사용량 합에서 빠진 행도 자기가 기록한 한도로 계속 cap 함.
		if caps[i].Key != bucket.Key {
			return fmt.Errorf("%w: %s bucket cap does not align with its snapshot", ErrRiskBucketSnapshotMismatch, bucket.Key.Dimension)
		}
		recorded, found, err := smallestRecordedBucketLimit(ctx, tx, account, bucket.Key.Dimension, bucket.Key.Value)
		if err != nil {
			return err
		}
		if !found {
			continue
		}
		after, ok := sumMinor(ledger.String(), caps[i].ReservationAtFinal)
		if !ok {
			return fmt.Errorf("%w: %s bucket %q reservation is not an amount", ErrRiskBucketSnapshotMismatch, bucket.Key.Dimension, bucket.Key.Value)
		}
		if after.Cmp(recorded) > 0 {
			return &riskbucket.RefusalError{Code: riskbucket.RefusalBucketCapExhausted, Field: string(bucket.Key.Dimension),
				Cause: fmt.Errorf("%s bucket %q: ledger %s + reservation %s exceeds the smallest recorded limit %s",
					bucket.Key.Dimension, bucket.Key.Value, ledger, caps[i].ReservationAtFinal, recorded)}
		}
	}
	return nil
}

// smallestRecordedBucketLimit 은 계좌 bucket(dimension, value)의 예약 중 원장 사용량에 드는 것(HELD·FILLED, 그리고 부분
// 체결 뒤 해제되어 filled 가 남은 RELEASED — 좁힌 재리뷰 P2)이 기록한 snapshot 한도 중 가장 작은 값을 돌려줌. 활성 예약이 없으면 found=false(첫 진입은 자기 snapshot 한도만 받음). 기록된 한도가 금액이 아니면 거절.
func smallestRecordedBucketLimit(ctx context.Context, tx *sql.Tx, account string, dimension riskbucket.Dimension, value string) (*big.Int, bool, error) {
	rows, err := tx.QueryContext(ctx, `SELECT DISTINCT s.limit_minor FROM risk_bucket_reservations r
		JOIN risk_bucket_snapshots s ON s.snapshot_id=r.snapshot_id AND s.bucket_dimension=r.bucket_dimension AND s.bucket_value=r.bucket_value
		WHERE r.account_ref=? AND r.bucket_dimension=? AND r.bucket_value=? AND (r.state IN ('HELD','FILLED') OR r.filled_minor<>'0')`, account, string(dimension), value)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()
	var smallest *big.Int
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, false, err
		}
		limit, ok := new(big.Int).SetString(raw, 10)
		if !ok || limit.Sign() < 0 {
			return nil, false, fmt.Errorf("%w: %s bucket %q recorded limit %q is not an amount", ErrRiskBucketSnapshotMismatch, dimension, value, raw)
		}
		if smallest == nil || limit.Cmp(smallest) < 0 {
			smallest = limit
		}
	}
	if err := rows.Err(); err != nil {
		return nil, false, err
	}
	return smallest, smallest != nil, nil
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

// latchedUsageRefusal 은 bucket 원장 사용량에 latch 가 있으면 진입을 막는 **단일 규칙**임 — admission 의 대조
// (refuseStaleBucketUsage)와 제출 재검증(RevalidateQFinalAdmission)이 같은 함수를 부름(a066 6.5). 거절은 bucket 과
// latch 종류를 이름으로 말함.
func latchedUsageRefusal(dimension riskbucket.Dimension, value string, usage riskbucket.JournalBucketUsage) error {
	if !usage.Latched {
		return nil
	}
	var kinds []string
	if usage.OverageLatched {
		kinds = append(kinds, string(riskbucket.LatchRiskOverage))
	}
	if usage.UnknownLatched {
		kinds = append(kinds, string(riskbucket.LatchUnknownActualRisk))
	}
	return fmt.Errorf("%w: %s bucket %q carries latched usage (%s)", ErrRiskBucketEntryBlocked, dimension, value, strings.Join(kinds, ","))
}
