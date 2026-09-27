package journal

// 이 파일은 a066 task 5.5 의 horizon×market 진입 손실 잠금(entry loss lock)임.
//
// 잠금은 신규 노출 증가(EXPOSURE_RAISING)만 막음. 판정 규칙은 refuseEntryUnderLossLock
// 하나이고, 그것을 부르는 자리는 셋임 — 두 admission 트랜잭션(CommitRiskBucketAdmission,
// commitFreshRiskBucketAdmissionTx)과 Gateway 가 제출 직전에 부르는 RevalidateQFinalAdmission.
// 앞의 둘은 잠금 읽기와 결정 기록을 한 트랜잭션에 묶어 TOCTOU 를 막고, 마지막 하나는 잠금
// 전에 발급되어 잠금 뒤에 제출되는 결정을 거절함(사용자 결정 ⑤, 2026-09-25).
//
// 손절·비상 청산·대사·체결 감지 경로는 이 파일의 어떤 함수도 부르지 않음.
//
// 완화(relaxation)는 없음. 승인 흐름이 사용자 결정 대기 중이라 API 모양 자체가 그 답에
// 종속됨 — 여기에는 자리만 남김(아래 주석). 이 로트의 생산 호출자는 0 임(dormant).

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
)

// schemaV33 은 잠금 테이블임(추가 전용, 기존 테이블 무변경).
//
//go:embed risk_bucket_entry_loss_lock_v33.sql
var schemaV33 string

// ErrRiskBucketEntryLossLocked 는 잠긴 horizon×market 범위의 신규 진입 거절임.
// ErrRiskBucketEntryBlocked 를 감싸므로 기존의 "진입 차단" 분류를 쓰는 호출자도 그대로 막힘.
var ErrRiskBucketEntryLossLocked = fmt.Errorf("%w: entry loss lock active", ErrRiskBucketEntryBlocked)

// EntryLossLock 은 계좌×시장×horizon 하나의 진입 손실 잠금 기록임.
type EntryLossLock struct {
	// Seq 는 원장이 정하는 신원임.
	Seq        int64
	AccountRef string
	Market     riskbucket.Market
	Horizon    riskbucket.Horizon
	// Cause 는 잠근 이유임. 열거는 트리거를 만드는 로트의 몫이라 여기서는 비어 있지 않음만 요구함.
	Cause       string
	ActivatedAt time.Time
}

// validEntryLossLockScope 는 잠금 범위(계좌·시장·horizon)가 이 스키마가 받는 값인지 판정함.
// 스키마의 CHECK 와 같은 값이어야 함 — 갈라지면 Go 는 통과시키고 SQLite 가 거절함.
//
// 계좌 앞뒤 공백은 거절함(정규화하지 않음): 결정·예약은 계좌를 trim 해서 쓰므로(decision.go·reservations.go)
// " acct-7 " 로 기록된 잠금은 "acct-7" 진입과 끝내 일치하지 않는 **효력 없는 잠금**이 됨. 잠갔다고 믿게
// 두느니 활성화를 실패시킴(적대 리뷰 codex P1, 2026-09-27).
func validEntryLossLockScope(account string, market riskbucket.Market, horizon riskbucket.Horizon) bool {
	return account != "" && account == strings.TrimSpace(account) &&
		(market == riskbucket.MarketKR || market == riskbucket.MarketUS) &&
		(horizon == riskbucket.HorizonShort || horizon == riskbucket.HorizonMedium)
}

// ActivateEntryLossLock 은 잠금을 기록하고 **지금 효력 있는** 잠금을 돌려줌.
//
// 보수 방향이라 즉시 영속함. 같은 범위에 이미 잠금이 있으면 아무것도 쓰지 않고 그 첫 잠금과
// changed=false 를 돌려줌 — 반복 트리거가 첫 원인을 덮지 않고, 동시 활성화는 순서와 무관하게
// 잠김으로 끝남(보수 쪽 승리).
func (j *Journal) ActivateEntryLossLock(ctx context.Context, lock EntryLossLock) (EntryLossLock, bool, error) {
	if j == nil || j.db == nil {
		return EntryLossLock{}, false, fmt.Errorf("%w: journal unavailable", ErrInvalidRequest)
	}
	if !validEntryLossLockScope(lock.AccountRef, lock.Market, lock.Horizon) ||
		strings.TrimSpace(lock.Cause) == "" || lock.ActivatedAt.IsZero() {
		return EntryLossLock{}, false, fmt.Errorf("%w: entry loss lock needs an account, KR/US, SHORT/MEDIUM, a cause and a time", ErrInvalidRequest)
	}
	tx, err := j.db.BeginTx(ctx, nil) // BEGIN IMMEDIATE (DSN _txlock=immediate)
	if err != nil {
		return EntryLossLock{}, false, fmt.Errorf("journal: begin entry loss lock: %w", err)
	}
	defer tx.Rollback()
	existing, found, err := activeEntryLossLock(ctx, tx, lock.AccountRef, lock.Market, lock.Horizon)
	if err != nil {
		return EntryLossLock{}, false, err
	}
	if found {
		return existing, false, nil
	}
	result, err := tx.ExecContext(ctx, `INSERT INTO risk_bucket_entry_loss_locks(account_ref,market,horizon,cause,activated_at)
		VALUES(?,?,?,?,?)`, lock.AccountRef, string(lock.Market), string(lock.Horizon), lock.Cause, formatJournalTime(lock.ActivatedAt))
	if err != nil {
		return EntryLossLock{}, false, fmt.Errorf("journal: recording entry loss lock: %w", err)
	}
	seq, err := result.LastInsertId()
	if err != nil {
		return EntryLossLock{}, false, fmt.Errorf("journal: reading entry loss lock identity: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return EntryLossLock{}, false, fmt.Errorf("journal: commit entry loss lock: %w", err)
	}
	lock.Seq = seq
	// 돌려주는 시각은 저장된 값(초 단위, formatJournalTime)과 같아야 함 — 다시 읽은 기록과 어긋나면 audit 값이 둘이 됨.
	lock.ActivatedAt = lock.ActivatedAt.UTC().Truncate(time.Second)
	return lock, true, nil
}

// 완화(relaxation) 자리: 사람 승인·audit 되는 해제는 사용자 결정 뒤에 이 테이블 옆의 별도
// append-only 기록으로 들어옴. 그때 activeEntryLossLock 의 "효력 있음"은 "해제 기록이 없음"으로
// 바뀜. 지금은 기록된 잠금 전부가 효력 있음.

// activeEntryLossLock 은 범위 하나의 효력 있는 잠금을 읽음.
func activeEntryLossLock(ctx context.Context, q riskBucketQueryer, account string, market riskbucket.Market, horizon riskbucket.Horizon) (EntryLossLock, bool, error) {
	var lock EntryLossLock
	var marketText, horizonText, activatedAt string
	err := q.QueryRowContext(ctx, `SELECT lock_seq,account_ref,market,horizon,cause,activated_at
		FROM risk_bucket_entry_loss_locks WHERE account_ref=? AND market=? AND horizon=?
		ORDER BY lock_seq LIMIT 1`, account, string(market), string(horizon)).
		Scan(&lock.Seq, &lock.AccountRef, &marketText, &horizonText, &lock.Cause, &activatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return EntryLossLock{}, false, nil
	}
	if err != nil {
		return EntryLossLock{}, false, fmt.Errorf("journal: reading entry loss lock: %w", err)
	}
	parsed, err := time.Parse(time.RFC3339Nano, activatedAt)
	if err != nil {
		return EntryLossLock{}, false, fmt.Errorf("journal: entry loss lock activated_at: %w", err)
	}
	lock.Market, lock.Horizon, lock.ActivatedAt = riskbucket.Market(marketText), riskbucket.Horizon(horizonText), parsed.UTC()
	return lock, true, nil
}

// refuseEntryUnderLossLock 은 진입 손실 잠금의 **유일한** 판정 규칙임.
//
// 범위가 이 스키마 밖의 값이거나 잠금을 읽지 못하면 거절함 — 진입 경로에서 "모름"은 "막힘"임.
// 잠금이 있으면 ErrRiskBucketEntryLossLocked 로 거절함.
func refuseEntryUnderLossLock(ctx context.Context, q riskBucketQueryer, account string, market riskbucket.Market, horizon riskbucket.Horizon) error {
	if !validEntryLossLockScope(account, market, horizon) {
		return fmt.Errorf("%w: entry loss lock scope %q/%q/%q is not one this build judges", ErrRiskBucketEntryBlocked, account, market, horizon)
	}
	lock, found, err := activeEntryLossLock(ctx, q, account, market, horizon)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrRiskBucketEntryBlocked, err)
	}
	if found {
		// 타입 있는 거절: riskbucket 거절 코드(ENTRY_LOSS_LOCK_ACTIVE)와 journal 센티널을 함께 실음.
		// 호출자는 문구가 아니라 errors.Is(ErrRiskBucketEntryLossLocked) 로 가름.
		return &riskbucket.RefusalError{Code: riskbucket.RefusalEntryLossLockActive, Field: "horizon×market",
			Cause: fmt.Errorf("%w: %s/%s since %s (lock %d: %s)", ErrRiskBucketEntryLossLocked,
				lock.Market, lock.Horizon, formatJournalTime(lock.ActivatedAt), lock.Seq, lock.Cause)}
	}
	return nil
}

// admissionHorizon 은 admission 이 예약을 쓰는 바로 그 horizon cap 의 값임.
// 예약 행(risk_bucket_reservations)이 decision.Caps 의 key 로 쓰이므로 같은 원천을 읽음.
func admissionHorizon(decision riskbucket.AdmissionDecision) riskbucket.Horizon {
	for _, cap := range decision.Caps {
		if cap.Key.Dimension == riskbucket.DimensionHorizon {
			return riskbucket.Horizon(cap.Key.Value)
		}
	}
	return ""
}

// decisionHorizon 은 이미 기록된 q_final 결정의 horizon 예약 값임(Gateway 재검증용).
func decisionHorizon(ctx context.Context, q riskBucketQueryer, decisionID string) (riskbucket.Horizon, error) {
	var value string
	if err := q.QueryRowContext(ctx, `SELECT bucket_value FROM risk_bucket_reservations WHERE decision_id=? AND bucket_dimension=?`,
		decisionID, string(riskbucket.DimensionHorizon)).Scan(&value); err != nil {
		return "", fmt.Errorf("%w: q_final horizon reservation: %v", ErrRiskBucketReplayMismatch, err)
	}
	return riskbucket.Horizon(value), nil
}
