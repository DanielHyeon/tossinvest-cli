package journal

// a137_daily_loss_rows.go 는 Guardian 의 일일 손실 측정이 읽는 원장 행을 돌려줌(a137 design D1 · D2).
//
// # 무엇을 돌려주는가
//
// 계좌의 CLOSED **관리** 포지션(Position.ExitEligible) 중 closed_at 이 [from, to) 인 것, 그리고 행마다
//   - 동결된 성과(realized_pnl_after_costs, 없으면 nil),
//   - 조정으로 닫혔는지(마지막 조정의 new_quantity 가 decimal 0),
//   - 조정으로 닫혔다면 그 인스턴스 수명과 겹쳐 엔진 매도가 브로커에 나갔을 수 있는지
//     (발송된 SELL PLACE attempt — NOT_DISPATCHED · FAILED_CONFIRMED 제외 — 또는 PLANNED 가 아닌 보호 주문 CREATE · REPLACE).
//
// 무엇을 손실로 세는지는 여기서 정하지 않음 — internal/dailyloss 가 정함. 여기는 사실만 읽음.
//
// # 단일 연결
//
// 원장은 연결 하나(SetMaxOpenConns(1))라 행을 연 채 다른 일을 하면 모든 원장 쓰기가 막힘. 그래서 질의마다 커서를 끝까지 읽고 닫은
// 뒤 다음 질의를 함. 네트워크 · 환산은 이 함수 밖(Guardian)이고, 이 함수는 어떤 브로커도 부르지 않음.
//
// # 형식이 어긋난 closed_at
//
// [from, to) 는 RFC3339 UTC 문자열 비교인 **사전 필터**임. 형식이 어긋난 행은 그 비교에서 빠질 수 있으므로, 같은 계좌의 CLOSED 관리
// 행 중 closed_at 이 「Z 로 끝나는 RFC3339 모양」이 아닌 것(NULL · 빈 값 포함)이 하나라도 있으면 오류를 돌려줌 — 그 행이 오늘의
// 손실일 수 있음(§4.2 P3-2).

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrDailyLossLedgerMalformed 는 측정 대상이 될 수 있는 행의 closed_at 형식이 어긋났다는 뜻임.
var ErrDailyLossLedgerMalformed = errors.New("journal: a closed managed position has a malformed closed_at")

// DailyLossRow 는 측정 후보 하나임.
type DailyLossRow struct {
	PositionID           string
	Market               string
	Symbol               string
	OpenedAt             string
	ClosedAt             string
	Outcome              *string
	AdjustmentClosed     bool
	EngineSellDispatched bool
}

// 관리 여부는 이 파일이 열 이름으로 다시 쓰지 않음 — positionSelect 로 읽고 Position.ExitEligible(position.ExitEligible 한 곳)로 가름
// (eligibility_drift_test · adoption_static_test 가 그 단일 철자를 지킴).
const closedPositionsWhere = " WHERE account_ref = ? AND state = 'CLOSED'"

// DailyLossRows 는 계좌의 측정 후보 행을 돌려줌. from · to 는 사전 필터 창(호출자가 일 키 창에 여유를 더해 넘김).
func (j *Journal) DailyLossRows(ctx context.Context, accountRef string, from, to time.Time) ([]DailyLossRow, error) {
	accountRef = strings.TrimSpace(accountRef)
	if accountRef == "" || from.IsZero() || !to.After(from) {
		return nil, errors.New("journal: a daily-loss read needs an account and a window")
	}
	malformed, err := j.countMalformedClosedAt(ctx, accountRef)
	if err != nil {
		return nil, err
	}
	if malformed > 0 {
		return nil, fmt.Errorf("%w: %d row(s)", ErrDailyLossLedgerMalformed, malformed)
	}
	rows, err := j.dailyLossCandidates(ctx, accountRef, formatJournalTime(from), formatJournalTime(to))
	if err != nil {
		return nil, err
	}
	for i := range rows {
		if !rows[i].AdjustmentClosed {
			continue
		}
		dispatched, err := j.engineSellOverlaps(ctx, accountRef, rows[i])
		if err != nil {
			return nil, err
		}
		rows[i].EngineSellDispatched = dispatched
	}
	return rows, nil
}

func (j *Journal) countMalformedClosedAt(ctx context.Context, accountRef string) (int, error) {
	positions, err := j.closedPositions(ctx, closedPositionsWhere+` AND (closed_at IS NULL OR closed_at NOT GLOB
	        '[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]T[0-9][0-9]:[0-9][0-9]:[0-9][0-9]*Z')`, accountRef)
	if err != nil {
		return 0, fmt.Errorf("journal: checking closed_at shapes for the daily loss: %w", err)
	}
	n := 0
	for _, p := range positions {
		if p.ExitEligible() {
			n++
		}
	}
	return n, nil
}

// closedPositions 는 positionSelect 로 읽고 커서를 닫은 뒤 돌려줌(단일 연결 — 다음 질의 전에 행을 다 읽음).
func (j *Journal) closedPositions(ctx context.Context, where string, args ...any) ([]Position, error) {
	rows, err := j.db.QueryContext(ctx, positionSelect+where, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Position
	for rows.Next() {
		p, err := scanPosition(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (j *Journal) dailyLossCandidates(ctx context.Context, accountRef, from, to string) ([]DailyLossRow, error) {
	positions, err := j.closedPositions(ctx, closedPositionsWhere+` AND closed_at >= ? AND closed_at < ? ORDER BY closed_at, id`,
		accountRef, from, to)
	if err != nil {
		return nil, fmt.Errorf("journal: reading the daily-loss candidates: %w", err)
	}
	var out []DailyLossRow
	for _, p := range positions {
		if !p.ExitEligible() {
			continue // 관리 대상이 아님(외부 · 수동 보유) — 세지 않음
		}
		r := DailyLossRow{PositionID: p.ID, Market: p.Market, Symbol: p.Symbol, OpenedAt: p.OpenedAt, ClosedAt: p.ClosedAt}
		var outcome, lastAdjQ *string
		if err := j.db.QueryRowContext(ctx, `SELECT
		        (SELECT realized_pnl_after_costs FROM trade_outcomes WHERE position_id = ?),
		        (SELECT new_quantity FROM position_adjustments WHERE position_id = ? ORDER BY created_at DESC, rowid DESC LIMIT 1)`,
			p.ID, p.ID).Scan(&outcome, &lastAdjQ); err != nil {
			return nil, fmt.Errorf("journal: reading the daily-loss facts of a close: %w", err)
		}
		r.Outcome = outcome
		// 조정 종결: 마지막 조정이 수량을 decimal 0 으로 만듦(「0.000」 도 0). 그 조정이 포지션을 CLOSED 로 만든 것임(convergedState).
		r.AdjustmentClosed = lastAdjQ != nil && isZeroDecimal(*lastAdjQ)
		out = append(out, r)
	}
	return out, nil
}

// engineSellOverlaps 는 그 인스턴스 수명과 겹쳐 엔진 매도가 브로커에 나갔을 수 있는지 봄. 시각을 해석할 수 없으면 겹친 것으로 봄(보수).
func (j *Journal) engineSellOverlaps(ctx context.Context, accountRef string, r DailyLossRow) (bool, error) {
	opened, openedOK := time.Time{}, true
	if strings.TrimSpace(r.OpenedAt) != "" {
		var err error
		if opened, err = parseJournalTime(r.OpenedAt); err != nil {
			openedOK = false
		}
	}
	closed, err := parseJournalTime(r.ClosedAt)
	if err != nil {
		return true, nil
	}
	// overlaps: [start, end] 이 [opened, closed] 와 겹침. end 가 비면 아직 끝나지 않은 것.
	overlaps := func(start, end string) bool {
		s, err := parseJournalTime(start)
		if err != nil || !openedOK {
			return true
		}
		if s.After(closed) {
			return false
		}
		if strings.TrimSpace(end) == "" {
			return true
		}
		e, err := parseJournalTime(end)
		if err != nil {
			return true
		}
		return opened.IsZero() || !e.Before(opened)
	}

	sells, err := j.collectSpans(ctx, `
		SELECT ma.recorded_at, coalesce(ma.settled_at, '')
		  FROM mutation_attempts ma JOIN intents i ON i.id = ma.intent_id
		 WHERE i.account_ref = ? AND lower(i.market) = lower(?) AND upper(i.symbol) = upper(?)
		   AND upper(i.side) = 'SELL' AND ma.kind = 'PLACE'
		   AND ma.state NOT IN (?, ?)`, accountRef, r.Market, r.Symbol, string(StateNotDispatched), string(StateFailedConfirmed))
	if err != nil {
		return false, err
	}
	for _, span := range sells {
		if overlaps(span[0], span[1]) {
			return true, nil
		}
	}
	// 보호 주문: 끝난 saga(TRIGGERED · CLOSED)는 saga 의 마지막 갱신을 끝으로, 살아 있으면 끝이 없음.
	protections, err := j.collectSpans(ctx, `
		SELECT pa.created_at,
		       CASE WHEN s.state IN ('TRIGGERED', 'CLOSED') THEN s.updated_at ELSE '' END
		  FROM protection_mutation_attempts pa JOIN protection_sagas s ON s.saga_id = pa.saga_id
		 WHERE s.account_ref = ? AND upper(s.market) = upper(?) AND upper(s.symbol) = upper(?)
		   AND pa.kind IN ('CREATE', 'REPLACE') AND pa.state <> 'PLANNED'`, accountRef, r.Market, r.Symbol)
	if err != nil {
		return false, err
	}
	for _, span := range protections {
		if overlaps(span[0], span[1]) {
			return true, nil
		}
	}
	return false, nil
}

func (j *Journal) collectSpans(ctx context.Context, query string, args ...any) ([][2]string, error) {
	rows, err := j.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("journal: reading engine sells for the daily loss: %w", err)
	}
	defer rows.Close()
	var out [][2]string
	for rows.Next() {
		var start, end string
		if err := rows.Scan(&start, &end); err != nil {
			return nil, fmt.Errorf("journal: reading engine sells for the daily loss: %w", err)
		}
		out = append(out, [2]string{start, end})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: reading engine sells for the daily loss: %w", err)
	}
	return out, nil
}
