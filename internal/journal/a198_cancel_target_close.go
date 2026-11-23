package journal

// a198_cancel_target_close.go 는 세션 중 IN_DOUBT 로 남은 엔진 취소를 「대상 주문의 원장 종결 기록」으로 닫는 자리임(a198 D10).
//
// # 왜 있나
//
// 취소가 5xx·전송 불명으로 IN_DOUBT 가 되면 같은 종목의 미종결 가드가 그 종목의 손절을 막고, 취소는 기동만 해소했음 — 대상 주문이
// 원장에서 끝났어도 재시작까지 손절이 보류됐음(구현 리뷰 r2 의 회귀). 대상이 끝났으면 그 취소가 성립했는지와 무관하게 「그 종목에
// 그 주문이 살아 있을 수 있다」는 불확실성이 사라짐 — 미종결 가드가 지키던 것이 그것임.
//
// # 지키는 것
//
//   - 브로커를 부르지 않음 — 원장만 읽고 씀. 판독과 전이를 한 트랜잭션(BEGIN IMMEDIATE)에 둠.
//   - 종결의 정의는 길 치우기와 같음(confirmedOrderTerminalMatch). 단 fail_closed 가 선 종결 행은 닫지 않음(증거 모순).
//   - 범위는 대상의 소유 PLACE/AMEND intent 의 계좌·시장·거래일·종목·방향 — 취소 intent 의 거래일을 쓰지 않음(전일 주문).
//     소유 intent 가 거래일을 건너 하나가 아니면 닫지 않음(번호 재사용).
//   - 판정 표는 기동 해소 resolveCancel 의 종결 갈래와 같음(취소로 끝남 → CONFIRMED, 체결·거부로 끝남 → FAILED_CONFIRMED).
//     REPLACED(잔량이 후계에 삶)·비종결·판정 불능은 아무것도 하지 않음(park 하지 않음 — 세션 park 는 계정 진입 래치를 검).
//   - 쓰기는 Attempt.transition 이 ResolveConfirmed/ResolveFailed 로 쓰는 것과 같은 행·열·사유(시험이 동치로 묶음).

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/brokerstate"
)

// CancelTargetClose 는 닫기 한 번의 결과임. Closed 가 거짓이면 attempt 는 바뀌지 않았음.
type CancelTargetClose struct {
	Closed bool
	State  AttemptState
	Reason string
}

// CancelVerdictForClosedTarget 은 대상 주문의 종결 state 로 취소 attempt 의 종결을 정함 — resolveCancel(amend_indoubt.go)의
// CANCELLED* → CONFIRMED, FILLED·REJECTED* → FAILED_CONFIRMED 와 같은 표. 그 밖은 닫지 않음(ok=false).
func CancelVerdictForClosedTarget(state string) (AttemptState, string, bool) {
	switch brokerstate.State(state) {
	case brokerstate.StateCancelled, brokerstate.StateCancelledPartiallyFilled:
		return StateConfirmed, ReasonResolvedFound, true
	case brokerstate.StateFilled, brokerstate.StateRejected, brokerstate.StateRejectedPartiallyFilled:
		return StateFailedConfirmed, ReasonResolvedAbsent, true
	default:
		return "", "", false
	}
}

// cancelTargetOwners 는 대상 번호의 소유 후보(접수 확정 PLACE/AMEND)를 취소 intent 의 계좌·시장·종목·방향으로 거름.
// 거래일은 걸지 않음 — 소유 intent 의 거래일이 종결 행 판독의 범위임.
const cancelTargetOwners = allFillSnapshotsCTE + `,
	owners AS (
		SELECT a.broker_order_id AS order_id, a.settled_at AS ownership_at, i.id AS intent_id,
		       i.account_ref, i.market, i.trading_day, i.symbol, i.side
		  FROM mutation_attempts a JOIN intents i ON i.id = a.intent_id
		 WHERE a.state = 'CONFIRMED' AND a.kind IN ('PLACE','AMEND') AND a.broker_order_id = ?
		   AND TRIM(i.account_ref) = ? AND LOWER(TRIM(i.market)) = ?
		   AND UPPER(TRIM(i.symbol)) = ? AND UPPER(TRIM(i.side)) = ?
	)`

// CloseInDoubtCancelOnTargetClose 는 IN_DOUBT 취소 attempt 하나를 대상 주문의 원장 종결 기록으로 닫음(a198 D10).
// source 는 닫을 취소를 낸 생산자(intents.source) — 엔진은 자기 취소만 닫음(flatten 의 saga 는 자기 취소를 자기가 해소).
// 닫을 대상이 아니거나 증거가 모자라면 오류 없이 Closed=false. 오류는 원장 읽기·쓰기 실패뿐이고, 그때 아무것도 바뀌지 않음.
func (j *Journal) CloseInDoubtCancelOnTargetClose(ctx context.Context, attemptID, source string) (CancelTargetClose, error) {
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return CancelTargetClose{}, fmt.Errorf("journal: starting the cancel close for %s: %w", attemptID, err)
	}
	defer tx.Rollback()

	var state, kind, target, account, market, symbol, side, producer string
	err = tx.QueryRowContext(ctx, `
		SELECT a.state, a.kind, a.target_order_id, TRIM(i.account_ref), LOWER(TRIM(i.market)),
		       UPPER(TRIM(i.symbol)), UPPER(TRIM(i.side)), TRIM(i.source)
		  FROM mutation_attempts a JOIN intents i ON i.id = a.intent_id
		 WHERE a.id = ?`, attemptID).Scan(&state, &kind, &target, &account, &market, &symbol, &side, &producer)
	if errors.Is(err, sql.ErrNoRows) {
		return CancelTargetClose{}, fmt.Errorf("%w: %s", ErrAttemptNotFound, attemptID)
	}
	if err != nil {
		return CancelTargetClose{}, fmt.Errorf("journal: reading cancel %s: %w", attemptID, err)
	}
	// 대상 고르기: IN_DOUBT 인 취소, 대상 번호 있음, 그 생산자가 낸 것 — 그 밖은 손대지 않음.
	if AttemptState(state) != StateInDoubt || MutationKind(kind) != KindCancel ||
		strings.TrimSpace(target) == "" || producer != strings.TrimSpace(source) {
		return CancelTargetClose{}, nil
	}
	args := []any{target, account, market, symbol, side}

	// 소유 유일성: 같은 범위에서 그 번호의 소유 intent 가 거래일을 건너 정확히 하나여야 함(번호 재사용이면 닫지 않음).
	var owners int
	if err := tx.QueryRowContext(ctx, cancelTargetOwners+`
		SELECT COUNT(DISTINCT intent_id) FROM owners`, args...).Scan(&owners); err != nil {
		return CancelTargetClose{}, fmt.Errorf("journal: reading the owner of order %s: %w", target, err)
	}
	if owners != 1 {
		return CancelTargetClose{}, nil
	}

	// 종결 행: 길 치우기와 같은 일치 조건. 하나라도 fail_closed 이거나 판정이 하나로 모이지 않으면 닫지 않음.
	rows, err := tx.QueryContext(ctx, cancelTargetOwners+`
		SELECT f.state, f.fail_closed FROM owners c JOIN all_fill_snapshots f ON `+confirmedOrderTerminalMatch, args...)
	if err != nil {
		return CancelTargetClose{}, fmt.Errorf("journal: reading the closing record of order %s: %w", target, err)
	}
	var verdict CancelTargetClose
	var closedAs string
	seen := false
	for rows.Next() {
		var fillState string
		var failClosed int
		if err := rows.Scan(&fillState, &failClosed); err != nil {
			rows.Close()
			return CancelTargetClose{}, fmt.Errorf("journal: reading the closing record of order %s: %w", target, err)
		}
		to, reason, ok := CancelVerdictForClosedTarget(fillState)
		if failClosed != 0 || !ok || (seen && (to != verdict.State || reason != verdict.Reason)) {
			rows.Close()
			return CancelTargetClose{}, nil
		}
		verdict, closedAs, seen = CancelTargetClose{Closed: true, State: to, Reason: reason}, fillState, true
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return CancelTargetClose{}, fmt.Errorf("journal: reading the closing record of order %s: %w", target, err)
	}
	rows.Close()
	if !seen {
		return CancelTargetClose{}, nil
	}

	// 전이 쓰기 — Attempt.transition 이 ResolveConfirmed/ResolveFailed 로 쓰는 것과 같은 열(state·settled_at·번호·사유·상세)과
	// 전이 행, 그리고 FAILED_CONFIRMED 면 같은 트랜잭션의 예약 해제.
	if err := ValidateTransition(StateInDoubt, verdict.State); err != nil {
		return CancelTargetClose{}, err
	}
	now := j.nowString()
	set, vals := "state = ?, settled_at = ?, reason_code = ?, detail = ?", []any{string(verdict.State), now, verdict.Reason}
	var detail string
	if verdict.State == StateConfirmed {
		detail = fmt.Sprintf("order %s is %s in the ledger's closing record, so the cancel is settled without a broker read", target, closedAs)
		set = "state = ?, settled_at = ?, broker_order_id = ?, reason_code = ?, detail = ?"
		vals = []any{string(verdict.State), now, target, verdict.Reason}
	} else {
		detail = fmt.Sprintf("order %s is %s in the ledger's closing record, so the cancel did not take effect", target, closedAs)
	}
	vals = append(vals, detail, attemptID, string(StateInDoubt))
	res, err := tx.ExecContext(ctx, "UPDATE mutation_attempts SET "+set+" WHERE id = ? AND state = ?", vals...)
	if err != nil {
		return CancelTargetClose{}, fmt.Errorf("journal: closing cancel %s: %w", attemptID, err)
	}
	if n, err := res.RowsAffected(); err != nil || n != 1 {
		// 같은 트랜잭션에서 읽은 IN_DOUBT 가 사라졌다면 다른 손이 먼저 닫은 것 — 무변으로 끝냄.
		if err != nil {
			return CancelTargetClose{}, fmt.Errorf("journal: closing cancel %s: %w", attemptID, err)
		}
		return CancelTargetClose{}, nil
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO attempt_transitions (attempt_id, from_state, to_state, at, reason_code, detail)
		 VALUES (?,?,?,?,?,?)`,
		attemptID, string(StateInDoubt), string(verdict.State), now, verdict.Reason, detail); err != nil {
		return CancelTargetClose{}, fmt.Errorf("journal: recording the IN_DOUBT->%s transition for %s: %w", verdict.State, attemptID, err)
	}
	if releasesReservations(verdict.State) {
		if _, err := releaseReservationsForAttempt(ctx, tx, attemptID, ReleaseReasonBrokerTerminal,
			fmt.Sprintf("attempt %s reached %s", attemptID, verdict.State), now); err != nil {
			return CancelTargetClose{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return CancelTargetClose{}, fmt.Errorf("journal: committing the cancel close for %s: %w", attemptID, err)
	}
	return verdict, nil
}
