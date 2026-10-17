package journal

import (
	"context"
	"fmt"
	"time"
)

// RecordAlert 는 발송 임차 없이 알림을 기록하고, 재알림 창(remindAfter)이 지난 정착 행은 다시 무장함(a092 C3).
//
// EnqueueAlert 와의 차이는 재알림 창 하나임 — EnqueueAlert 는 0 을 넘겨 남의 정착 행을 재무장하지 않고, 이 입구는
// 기록자가 창을 정함(0 허용). ClaimAlertForDelivery 와의 차이는 임차 하나임 — 이 입구는 발송하지 않으므로 임차를
// 잡으면 배달 실행자가 그 임차가 끝날 때까지 행을 못 집음.
//
// 비시험 호출자는 알림기의 기록 전용 입구 하나여야 함(K7 핀). 그 입구가 알림기의 배제 잠금 아래에서 부르므로
// 운영자 승인의 셈~해제 사이에 이 경로로 PENDING 행이 끼어들지 못함 — 잠금 밖에서 이것을 부르면 그 배제가 깨짐.
//
// 반환: 행 id, 발송 빚 여부(새 행 · 재무장 · 미전달 PENDING 이면 true).
func (j *Journal) RecordAlert(ctx context.Context, a Alert, remindAfter time.Duration) (int64, bool, error) {
	key, err := alertKey(a)
	if err != nil {
		return 0, false, err
	}
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, false, fmt.Errorf("journal: recording alert %s: %w", key, err)
	}
	defer tx.Rollback()

	// 알림 하나에 트랜잭션 하나 — 기록과 재무장 판정은 recordAlertTx 가 모두 함, 임차 획득은 없음.
	id, owed, err := j.recordAlertTx(ctx, tx, a, remindAfter)
	if err != nil {
		return 0, false, err
	}
	if err := tx.Commit(); err != nil {
		return 0, false, fmt.Errorf("journal: committing alert %s: %w", key, err)
	}
	return id, owed, nil
}
