package journal

// exit_proposal_release.go 는 a094 R3 의 발의 해제 판정임 — 「종결된 attempt 가 그것이 무장한 발의를 푼다」.
//
// # 왜 있나
//
// 무장된 발의(exit 상태의 발의 컬럼)는 두 번째 매도를 막는 실제 방벽임(armExitProposalTx 가 두 번째를 거절함).
// 그 발의를 실행하려던 attempt 가 접수되지 않음으로 끝났는데 발의가 무장된 채 남으면, 청산 정책은 손절 조건에서도 빈 전이를
// 돌려줘 그 포지션은 영구히 무보호가 됨(2026-08-07 475150 · 080220). 반대로 주문이 살아 있을 수 있는데 발의를 비우면 다음
// 관측이 그 위에 두 번째 매도를 얹음. 그래서 해제는 **attempt 상태에 선 한 판정**으로만 함.
//
// # 판정 하나 (정본 order-execution 「종결된 mutation attempt 는 그것이 무장한 보호 발의를 풀어야 한다」)
//
//   - 해제 대상은 그 intent 의 attempt 가 **전부** FAILED_CONFIRMED · NOT_DISPATCHED 이거나 **하나도 없을** 때뿐임.
//   - UNRESOLVED_IN_DOUBT(park)는 해제하지 않음 — 원 주문의 존재가 미지임. 운영자 해소 뒤에만.
//   - 늦게 도착한 해제가 그 사이 무장된 다른 발의를 지우지 않게, 현재 무장된 발의의 intent 가 기대 intent 와 같을 때만 비움.
//   - 판정 읽기와 해제 쓰기는 한 트랜잭션임 — 읽은 뒤 attempt 가 바뀌어 해제가 틀리는 창이 없음.
//
// 세션 중 제출(submit) · 기동 따라잡기 · 운영자 해동 명령 · 청소가 모두 이 파일의 분류기를 부름(판정을 둘로 두지 않음).

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// ExitIntentVerdict 는 발의 intent 의 attempt 들을 해제 관점에서 분류한 값임. 우선순위는 위에서 아래임 —
// 하나라도 살아 있을 수 있는 attempt 가 있으면 그것이 답임.
type ExitIntentVerdict string

const (
	// ExitIntentLive 는 기록 · 전송 시작 · 접수 · 모호 상태의 attempt 가 있음 — 주문이 살아 있을 수 있음.
	ExitIntentLive ExitIntentVerdict = "LIVE"
	// ExitIntentParked 는 판정 불능으로 park 된 attempt 가 있음 — 운영자 해소 전에는 풀지 않음.
	ExitIntentParked ExitIntentVerdict = "PARKED"
	// ExitIntentAwaitingClose 는 접수 확정 주문 중 종결 체결 기록이 없는 것이 있음 — 매도가 살아 있을 수 있음.
	ExitIntentAwaitingClose ExitIntentVerdict = "AWAITING_CLOSE"
	// ExitIntentOrdersClosed 는 접수 확정 주문이 있고 전부 종결 증거가 있으며 나머지는 비수용 종결임.
	ExitIntentOrdersClosed ExitIntentVerdict = "ORDERS_CLOSED"
	// ExitIntentUnaccepted 는 attempt 가 전부 FAILED_CONFIRMED · NOT_DISPATCHED 임 — 입증된 비수용.
	ExitIntentUnaccepted ExitIntentVerdict = "UNACCEPTED"
	// ExitIntentNoAttempts 는 attempt 가 하나도 없음 — 전송 전 기록이 없으므로 아무것도 나가지 않았음.
	ExitIntentNoAttempts ExitIntentVerdict = "NO_ATTEMPTS"
)

// releasesUnaccepted 는 해제 판정 함수의 조건임 — 입증된 비수용이거나 attempt 가 없음.
func (v ExitIntentVerdict) releasesUnaccepted() bool {
	return v == ExitIntentUnaccepted || v == ExitIntentNoAttempts
}

// IntentOrderAwaitingClose 는 발의 intent 의 접수 확정 주문 중 종결 체결 기록이 아직 없는 것 하나임.
type IntentOrderAwaitingClose struct {
	AttemptID  string
	OrderID    string // 빈 값이면 번호가 기록되지 않은 접수 확정 — 종결을 확인할 방법이 없음
	AccountRef string
	Market     string
	Symbol     string
	Side       string
}

// ExitIntentAttempts 는 발의 intent 의 attempt 판정과 알림에 필요한 사실임.
type ExitIntentAttempts struct {
	Verdict ExitIntentVerdict
	// Parked 는 판정 불능으로 park 된 attempt 들임(park 원인 알림이 이름을 댐).
	Parked []AttemptRecord
	// Awaiting 은 종결 증거를 기다리는 접수 확정 주문들임.
	Awaiting []IntentOrderAwaitingClose
}

// ExitIntentAttempts 는 intent 하나의 attempt 를 읽어 분류함. 원장만 읽음(브로커 호출 0).
func (j *Journal) ExitIntentAttempts(ctx context.Context, intentID string) (ExitIntentAttempts, error) {
	return exitIntentAttemptsQ(ctx, j.db, intentID)
}

// ReleaseUnacceptedExitProposal 은 해제 판정 함수임(a094 D−2.5 · Q4-1) — 세션 중 제출 · 기동 따라잡기 · 운영자 해동이 공유함.
//
// 발의 intent 의 attempt 가 전부 입증된 비수용이거나 하나도 없고, 현재 무장된 발의의 intent 가 intentID 와 같을 때만 발의를
// 비움. 판정이 해제를 허락하지 않거나 intent 가 다르면 아무것도 바꾸지 않고 released=false 를 돌려줌. 이미 해제된 발의는 멱등.
func (j *Journal) ReleaseUnacceptedExitProposal(ctx context.Context, positionID, intentID string,
	resolution ProposalResolution) (ExitIntentVerdict, bool, error) {
	return j.releaseExitProposal(ctx, positionID, intentID, resolution, ExitIntentVerdict.releasesUnaccepted)
}

// ReleaseClearedExitProposal 은 청소(clearTheSymbol)의 해제임 — 위 판정에 더해, 접수 확정 주문이 전부 종결 증거를 가진
// 경우(ExitIntentOrdersClosed)도 PROPOSAL_CANCELLED 로 풂. 매도의 취소 접수는 치움이 아니라는 규칙(a094 D−4.3)이 여기서
// 선다 — 종결 체결 기록이 올 때까지 발의는 무장된 채 남음.
func (j *Journal) ReleaseClearedExitProposal(ctx context.Context, positionID, intentID string) (ExitIntentVerdict, bool, error) {
	return j.releaseExitProposal(ctx, positionID, intentID, ProposalCancelled, func(v ExitIntentVerdict) bool {
		return v.releasesUnaccepted() || v == ExitIntentOrdersClosed
	})
}

func (j *Journal) releaseExitProposal(ctx context.Context, positionID, intentID string,
	resolution ProposalResolution, allows func(ExitIntentVerdict) bool) (ExitIntentVerdict, bool, error) {
	id := strings.TrimSpace(positionID)
	intentID = strings.TrimSpace(intentID)
	if intentID == "" {
		return "", false, fmt.Errorf("%w: releasing the proposal of %s needs the intent it armed", ErrInvalidRequest, id)
	}
	action, err := proposalResolutionAction(resolution)
	if err != nil {
		return "", false, err
	}
	tx, err := j.db.BeginTx(ctx, nil) // BEGIN IMMEDIATE — 판정 읽기와 해제 쓰기가 한 트랜잭션
	if err != nil {
		return "", false, fmt.Errorf("journal: releasing the proposal of %s: %w", id, err)
	}
	defer tx.Rollback()

	facts, err := exitIntentAttemptsQ(ctx, tx, intentID)
	if err != nil {
		return "", false, err
	}
	if !allows(facts.Verdict) {
		return facts.Verdict, false, nil
	}
	released, err := clearExitProposalTx(ctx, tx, id, intentID, action, j.nowString())
	if err != nil {
		return facts.Verdict, false, err
	}
	if !released {
		return facts.Verdict, false, nil
	}
	if err := tx.Commit(); err != nil {
		return facts.Verdict, false, fmt.Errorf("journal: releasing the proposal of %s: %w", id, err)
	}
	return facts.Verdict, true, nil
}

// ArmedExitProposal 은 무장된 발의 하나임(기동 따라잡기의 순회 대상).
type ArmedExitProposal struct {
	PositionID string
	IntentID   string
	Action     string
	Level      string
}

// ArmedExitProposals 는 무장된 발의가 있는 exit 상태를 전부 읽음(계정 한정, 원장 읽기뿐). 가드 컬럼을 이름으로 읽는
// 질의는 apply_hook.go 에 둠(가드 컬럼 이름은 그 파일과 DDL · 관리 목록에만 나온다 — TestGuardedExitColumns…).
func (j *Journal) ArmedExitProposals(ctx context.Context, accountRef string) ([]ArmedExitProposal, error) {
	return armedExitProposalsQ(ctx, j.db, accountRef)
}

// ConfirmedCancelOf 는 엔진이 낸 취소 중 그 주문을 대상으로 접수 확정된 첫 attempt 를 돌려줌(계정 · 시장 · 종목 한정).
// found=false 면 그런 취소가 없음. 청소의 재취소 금지와 종결 증거 대기 알림이 씀.
func (j *Journal) ConfirmedCancelOf(ctx context.Context, accountRef, market, symbol, orderID string) (AttemptRecord, bool, error) {
	orderID = strings.TrimSpace(orderID)
	if orderID == "" {
		return AttemptRecord{}, false, nil
	}
	row := j.db.QueryRowContext(ctx, `
		SELECT a.id FROM mutation_attempts a JOIN intents i ON i.id = a.intent_id
		 WHERE a.kind = 'CANCEL' AND a.state = ? AND a.target_order_id = ?
		   AND TRIM(i.account_ref) = ? AND LOWER(TRIM(i.market)) = ? AND UPPER(TRIM(i.symbol)) = ?
		 ORDER BY a.settled_at, a.rowid LIMIT 1`,
		string(StateConfirmed), orderID, strings.TrimSpace(accountRef), normaliseMarket(market), normaliseSymbol(symbol))
	var id string
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return AttemptRecord{}, false, nil
		}
		return AttemptRecord{}, false, fmt.Errorf("journal: reading the engine cancel of order %s: %w", orderID, err)
	}
	rec, err := j.LookupAttempt(ctx, id)
	if err != nil {
		return AttemptRecord{}, false, err
	}
	return rec, true, nil
}

// queryer 는 *sql.DB 와 *sql.Tx 가 공유하는 읽기 면임.
type queryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// exitIntentAttemptsQ 는 분류기 본체임 — 해제 두 입구와 읽기 입구가 모두 이것을 부름.
func exitIntentAttemptsQ(ctx context.Context, q queryer, intentID string) (ExitIntentAttempts, error) {
	intentID = strings.TrimSpace(intentID)
	rows, err := q.QueryContext(ctx, attemptSelect+` WHERE intent_id = ? ORDER BY recorded_at, rowid`, intentID)
	if err != nil {
		return ExitIntentAttempts{}, fmt.Errorf("journal: reading the attempts of intent %s: %w", intentID, err)
	}
	var attempts []AttemptRecord
	for rows.Next() {
		rec, err := scanAttempt(rows)
		if err != nil {
			rows.Close()
			return ExitIntentAttempts{}, fmt.Errorf("journal: reading the attempts of intent %s: %w", intentID, err)
		}
		attempts = append(attempts, rec)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return ExitIntentAttempts{}, fmt.Errorf("journal: reading the attempts of intent %s: %w", intentID, err)
	}
	rows.Close()

	var out ExitIntentAttempts
	if len(attempts) == 0 {
		out.Verdict = ExitIntentNoAttempts
		return out, nil
	}
	live, confirmed := false, false
	for _, rec := range attempts {
		switch rec.State {
		case StateFailedConfirmed, StateNotDispatched:
		case StateUnresolvedInDoubt:
			out.Parked = append(out.Parked, rec)
		case StateConfirmed:
			confirmed = true
		default:
			// RECORDED · DISPATCH_STARTED · ACKED · IN_DOUBT, 그리고 이 판본이 모르는 상태 — 살아 있을 수 있음으로 다룸.
			live = true
		}
	}
	if confirmed {
		awaiting, err := intentOrdersAwaitingCloseQ(ctx, q, intentID)
		if err != nil {
			return ExitIntentAttempts{}, err
		}
		out.Awaiting = awaiting
	}
	switch {
	case live:
		out.Verdict = ExitIntentLive
	case len(out.Parked) > 0:
		out.Verdict = ExitIntentParked
	case len(out.Awaiting) > 0:
		out.Verdict = ExitIntentAwaitingClose
	case confirmed:
		out.Verdict = ExitIntentOrdersClosed
	default:
		out.Verdict = ExitIntentUnaccepted
	}
	return out, nil
}

// intentOrdersAwaitingCloseQ 는 intent 의 접수 확정 PLACE · AMEND 주문 중 종결 체결 기록이 없는 것을 돌려줌.
//
// 종결 증거는 미체결 목록(LiveOrdersForSymbol)과 **같은 술어**(confirmedOrderTerminalEvidence)로 봄 — 다만 소유 intent 의
// 유일성 필터를 두지 않음. 소유가 모호해 미체결 목록에서 빠진 주문도 여기서는 그 intent 의 주문 번호로 종결을 확인함
// (a094 D−4.3-2 — 목록 부재를 「없음」으로 읽지 않음). 번호 없는 접수 확정은 종결을 확인할 수 없으므로 기다림으로 다룸.
func intentOrdersAwaitingCloseQ(ctx context.Context, q queryer, intentID string) ([]IntentOrderAwaitingClose, error) {
	rows, err := q.QueryContext(ctx, allFillSnapshotsCTE+`,
		intent_orders AS (
			SELECT a.id AS attempt_id, a.broker_order_id AS order_id, a.settled_at AS ownership_at,
			       i.id AS intent_id, i.account_ref, i.market, i.trading_day, i.symbol, i.side
			  FROM mutation_attempts a JOIN intents i ON i.id = a.intent_id
			 WHERE a.intent_id = ? AND a.state = ? AND a.kind IN ('PLACE','AMEND')
		)
		SELECT c.attempt_id, c.order_id, c.account_ref, c.market, c.symbol, c.side
		  FROM intent_orders c
		 WHERE c.order_id = '' OR NOT `+confirmedOrderTerminalEvidence+`
		 ORDER BY c.ownership_at, c.attempt_id`, strings.TrimSpace(intentID), string(StateConfirmed))
	if err != nil {
		return nil, fmt.Errorf("journal: reading the closing evidence of intent %s: %w", intentID, err)
	}
	defer rows.Close()
	var out []IntentOrderAwaitingClose
	for rows.Next() {
		var o IntentOrderAwaitingClose
		if err := rows.Scan(&o.AttemptID, &o.OrderID, &o.AccountRef, &o.Market, &o.Symbol, &o.Side); err != nil {
			return nil, fmt.Errorf("journal: reading the closing evidence of intent %s: %w", intentID, err)
		}
		out = append(out, o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: reading the closing evidence of intent %s: %w", intentID, err)
	}
	return out, nil
}

// confirmedOrderTerminalEvidence 는 접수 확정 주문 c 에 종결 체결 기록이 있는지의 술어임 — 미체결 목록과 발의 해제가 공유함.
// c 는 order_id · ownership_at · intent_id · account_ref · market · trading_day · symbol · side 를 가져야 하고, 질의는
// allFillSnapshotsCTE 를 앞에 둬야 함.
const confirmedOrderTerminalEvidence = `EXISTS (
		   SELECT 1 FROM all_fill_snapshots f
		    WHERE f.order_id = c.order_id AND f.terminal = 1
		      AND c.ownership_at < f.committed_at
		      AND UPPER(TRIM(f.symbol)) = UPPER(TRIM(c.symbol))
		      AND UPPER(TRIM(f.market)) = UPPER(TRIM(c.market))
		      AND (TRIM(f.side) = '' OR UPPER(TRIM(f.side)) = UPPER(TRIM(c.side)))
		      AND ((TRIM(f.account_ref) = TRIM(c.account_ref)
		            AND TRIM(f.trading_day) = TRIM(c.trading_day))
		           OR (TRIM(f.account_ref) = '' AND TRIM(f.trading_day) = ''
		               AND TRIM(f.side) = '' AND EXISTS (
		                 SELECT 1 FROM legacy_snapshot_owner owner
		                  WHERE owner.order_id = f.order_id AND owner.intent_id = c.intent_id)))
		 )`

// proposalResolutionAction 은 해제 방식을 exit_events 행동으로 바꿈.
func proposalResolutionAction(resolution ProposalResolution) (string, error) {
	switch resolution {
	case ProposalRefused:
		return ExitEventProposalRefused, nil
	case ProposalCancelled:
		return ExitEventProposalCancelled, nil
	default:
		return "", fmt.Errorf("%w: %q is neither %s nor %s", ErrInvalidRequest,
			string(resolution), ProposalRefused, ProposalCancelled)
	}
}
