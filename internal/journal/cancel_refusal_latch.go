package journal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// ConfirmedCancelAfterLatch 는 R3 latch 뒤에 기록을 시작한 같은 대상 취소를 찾음.
func (j *Journal) ConfirmedCancelAfterLatch(ctx context.Context, latchAttemptID string) (AttemptRecord, bool, error) {
	latchAttemptID = strings.TrimSpace(latchAttemptID)
	if latchAttemptID == "" {
		return AttemptRecord{}, false, nil
	}
	row := j.db.QueryRowContext(ctx, `
		WITH l AS (
			SELECT a.id, a.target_order_id AS target, TRIM(i.account_ref) AS account,
			       LOWER(TRIM(i.market)) AS market, UPPER(TRIM(i.symbol)) AS symbol,
			       (SELECT MAX(t.id) FROM attempt_transitions t
			         WHERE t.attempt_id = a.id AND t.to_state = 'FAILED_CONFIRMED') AS latch_seq
			  FROM mutation_attempts a JOIN intents i ON i.id = a.intent_id
			 WHERE a.id = ? AND a.kind = 'CANCEL' AND a.state = 'FAILED_CONFIRMED' AND a.target_order_id <> ''
		), owners AS (
			SELECT COUNT(DISTINCT o.intent_id) AS n
			  FROM l, mutation_attempts o JOIN intents oi ON oi.id = o.intent_id
			 WHERE o.state = 'CONFIRMED' AND o.kind IN ('PLACE','AMEND') AND o.broker_order_id = l.target
			   AND TRIM(oi.account_ref) = l.account AND LOWER(TRIM(oi.market)) = l.market AND UPPER(TRIM(oi.symbol)) = l.symbol
		), candidates AS (
			SELECT c.id, l.latch_seq,
			       (SELECT MIN(t.id) FROM attempt_transitions t
			         WHERE t.attempt_id = c.id AND t.from_state = '' AND t.to_state = 'RECORDED') AS recorded_seq
			  FROM l, owners, mutation_attempts c JOIN intents ci ON ci.id = c.intent_id
			 WHERE c.kind = 'CANCEL' AND c.state = 'CONFIRMED' AND c.id <> l.id AND c.target_order_id = l.target
			   AND TRIM(ci.account_ref) = l.account AND LOWER(TRIM(ci.market)) = l.market AND UPPER(TRIM(ci.symbol)) = l.symbol
			   AND owners.n = 1
		)
		SELECT id FROM candidates WHERE recorded_seq > latch_seq ORDER BY recorded_seq LIMIT 1`, latchAttemptID)
	var id string
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return AttemptRecord{}, false, nil
		}
		return AttemptRecord{}, false, fmt.Errorf("journal: reading confirmed cancel after latch %s: %w", latchAttemptID, err)
	}
	record, err := j.LookupAttempt(ctx, id)
	if err != nil {
		return AttemptRecord{}, false, err
	}
	return record, true, nil
}

// FailedConfirmedCancelsForSymbol 은 진입 판정이 확인할 종결 취소 행을 읽음.
func (j *Journal) FailedConfirmedCancelsForSymbol(ctx context.Context, accountRef, market, symbol string) ([]AttemptRecord, error) {
	rows, err := j.db.QueryContext(ctx, attemptSelect+`
		 WHERE id IN (
			SELECT a.id FROM mutation_attempts a JOIN intents i ON i.id = a.intent_id
			 WHERE a.kind = 'CANCEL' AND a.state = 'FAILED_CONFIRMED'
			   AND TRIM(i.account_ref) = ? AND LOWER(TRIM(i.market)) = ? AND UPPER(TRIM(i.symbol)) = ?
		 ) ORDER BY settled_at, rowid`, strings.TrimSpace(accountRef), normaliseMarket(market), normaliseSymbol(symbol))
	if err != nil {
		return nil, fmt.Errorf("journal: listing failed confirmed cancels for %s: %w", symbol, err)
	}
	defer rows.Close()
	var records []AttemptRecord
	for rows.Next() {
		record, err := scanAttempt(rows)
		if err != nil {
			return nil, fmt.Errorf("journal: reading failed confirmed cancels for %s: %w", symbol, err)
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: listing failed confirmed cancels for %s: %w", symbol, err)
	}
	return records, nil
}
