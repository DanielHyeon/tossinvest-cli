package journal

// replay.go is the durable bookkeeping the idempotent replay procedure needs
// (extend-execution-contract tasks 2.1/2.2/2.4, design D3).
//
// # What is here, and what deliberately is not
//
// The replay *door* is on the gateway. The journal holds no HTTP client and the
// resolution procedure holds no mutator (P1 invariant), so nothing in this file
// sends anything. What the journal owns is the counting — how many times a
// stored body has been resent, and when — because that is the only bound on
// replay that survives a crash. An in-memory counter would reset in exactly the
// situation replay exists for: the process that lost a response restarting.
//
// # Why the count is taken before the request rather than after
//
// A replay is bounded by a cap (default 2). If the count were written after the
// response, a crash between "sent" and "counted" would hand back a free replay,
// and a crash loop would hand back an unbounded number of them. Counting first
// makes the crash case conservative: at worst an attempt loses a replay it never
// used, which costs nothing — the query fallback still runs — while the opposite
// mistake spends the broker's ten-minute key window on an unbounded loop.
//
// The one answer that does not consume the cap is `409 request-in-progress`
// ("동일 주문 키에 대해 처리 중인 요청이 있습니다" — openapi): the original request
// is still being processed, so the replay learned nothing, and it is the most
// common answer of all. RefundReplay gives the count back and deliberately
// leaves last_replay_at alone — the minimum interval between replays still
// applies, because the thing to do about "still processing" is to wait.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// Reason codes written by the replay procedure. Stable strings: they are stored
// in reason_code and read by alerting and by the Phase 2 ledger.
const (
	// ReasonReplayRecovered: the attempt's identity was recovered by resending
	// the stored body under the stored idempotency key. The broker returned the
	// original order's result rather than creating a second order (openapi).
	ReasonReplayRecovered = "in_doubt_replay_recovered"
	// ReasonReplayKeyConflict: the broker answered a replay with
	// `422 idempotency-key-conflict` (openapi), or echoed a key that is not the
	// one this attempt sent. Either says the key does not name our order — and
	// nothing at all about whether the original order exists. It is therefore a
	// park, never a FAILED_CONFIRMED.
	ReasonReplayKeyConflict = "in_doubt_replay_key_conflict"
	// ReasonObservationContaminated: absence could not be judged because another
	// mutation on the same symbol was dispatched inside the observation window,
	// which invalidates the balance/holding cross-check.
	ReasonObservationContaminated = "in_doubt_observation_contaminated"
)

var (
	// ErrReplayCapReached means the attempt has already used its replay
	// allowance. The caller falls back to the query procedure.
	ErrReplayCapReached = errors.New("journal: the attempt has used its replay allowance")
	// ErrReplayNotInDoubt means the attempt is not in the only state a replay is
	// defined for. Checked inside the transaction, so a concurrent settle cannot
	// slip a replay through behind it.
	ErrReplayNotInDoubt = errors.New("journal: only an IN_DOUBT attempt can be replayed")
)

// ReplayState is an attempt's replay bookkeeping.
type ReplayState struct {
	// Count is how many replays have been counted against the attempt.
	Count int
	// LastAt is when the last one was counted, RFC3339 UTC, "" before the first.
	LastAt string
}

// MarkReplayStarted counts one replay *before* it is sent, refusing when the cap
// is already spent.
//
// The cap is checked and the counter advanced in one BEGIN IMMEDIATE
// transaction, so two callers cannot both read "one replay left" and both send
// one. The state check is in the same transaction for the same reason.
func (j *Journal) MarkReplayStarted(ctx context.Context, attemptID string, max int) (ReplayState, error) {
	if max <= 0 {
		return ReplayState{}, fmt.Errorf("%w: a replay cap of %d authorises nothing", ErrInvalidRequest, max)
	}
	if strings.TrimSpace(attemptID) == "" {
		return ReplayState{}, fmt.Errorf("%w: attempt id is required", ErrInvalidRequest)
	}
	now := j.nowString()

	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return ReplayState{}, fmt.Errorf("journal: starting the replay transaction for %s: %w", attemptID, err)
	}
	defer tx.Rollback()

	var (
		state  AttemptState
		count  sql.NullInt64
		lastAt sql.NullString
	)
	err = tx.QueryRowContext(ctx,
		`SELECT state, replay_count, last_replay_at FROM mutation_attempts WHERE id = ?`,
		attemptID).Scan(&state, &count, &lastAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ReplayState{}, fmt.Errorf("%w: %s", ErrAttemptNotFound, attemptID)
	}
	if err != nil {
		return ReplayState{}, fmt.Errorf("journal: reading the replay state of %s: %w", attemptID, err)
	}

	current := ReplayState{Count: int(count.Int64), LastAt: lastAt.String}
	if state != StateInDoubt {
		return current, fmt.Errorf("%w: attempt %s is %s", ErrReplayNotInDoubt, attemptID, state)
	}
	if current.Count >= max {
		return current, fmt.Errorf("%w: attempt %s has been replayed %d time(s) and the cap is %d",
			ErrReplayCapReached, attemptID, current.Count, max)
	}

	next := current.Count + 1
	res, err := tx.ExecContext(ctx,
		`UPDATE mutation_attempts SET replay_count = ?, last_replay_at = ?
		  WHERE id = ? AND state = ? AND coalesce(replay_count, 0) = ?`,
		next, now, attemptID, string(StateInDoubt), current.Count)
	if err != nil {
		return current, fmt.Errorf("journal: counting a replay for %s: %w", attemptID, err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return current, fmt.Errorf("journal: counting a replay for %s: %w", attemptID, err)
	}
	if affected != 1 {
		return current, fmt.Errorf("journal: the replay counter for %s moved under us", attemptID)
	}
	if err := tx.Commit(); err != nil {
		return current, fmt.Errorf("journal: committing the replay count for %s: %w", attemptID, err)
	}
	return ReplayState{Count: next, LastAt: now}, nil
}

// RefundReplay gives back one counted replay.
//
// It exists for exactly one answer — `409 request-in-progress` (openapi) — where
// the broker is still processing the original request and the replay therefore
// established nothing. last_replay_at is left where it is on purpose: the reply
// to "still processing" is to wait, and the minimum interval is what enforces
// that.
func (j *Journal) RefundReplay(ctx context.Context, attemptID string) (ReplayState, error) {
	if _, err := j.db.ExecContext(ctx,
		`UPDATE mutation_attempts SET replay_count = max(coalesce(replay_count, 0) - 1, 0) WHERE id = ?`,
		attemptID); err != nil {
		return ReplayState{}, fmt.Errorf("journal: refunding a replay for %s: %w", attemptID, err)
	}
	var (
		count  sql.NullInt64
		lastAt sql.NullString
	)
	err := j.db.QueryRowContext(ctx,
		`SELECT replay_count, last_replay_at FROM mutation_attempts WHERE id = ?`, attemptID).
		Scan(&count, &lastAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ReplayState{}, fmt.Errorf("%w: %s", ErrAttemptNotFound, attemptID)
	}
	if err != nil {
		return ReplayState{}, fmt.Errorf("journal: reading the replay state of %s: %w", attemptID, err)
	}
	return ReplayState{Count: int(count.Int64), LastAt: lastAt.String}, nil
}

// --- query fallback support (task 2.4) --------------------------------------

// MutationsDispatchedSince returns the attempts on one market and symbol whose
// dispatch began at or after `since`, excluding one attempt id.
//
// It answers the question the absence judgement depends on: "could anything
// other than this attempt have moved the numbers I am about to compare?" The
// balance and holding cross-check compares the account now against a baseline
// captured before this attempt was dispatched, so any *other* mutation on the
// same symbol dispatched inside that window makes the comparison meaningless —
// and an automatic FAILED_CONFIRMED built on a meaningless comparison is the one
// error that leaves a live order untracked (order-execution: "관측 창 동안 같은
// 심볼에 다른 mutation이 전송되었다면 … 자동 FAILED_CONFIRMED는 금지된다").
//
// NOT_DISPATCHED attempts are excluded: the dispatch transition commits before
// the call, so a refusal raised between the two carries a dispatch timestamp
// while provably having sent nothing. Timestamps are compared as text, which is
// ordering-correct because every stamp this journal writes is RFC3339 UTC.
func (j *Journal) MutationsDispatchedSince(ctx context.Context,
	market, symbol, since, exceptAttemptID string,
) ([]AttemptRecord, error) {
	rows, err := j.db.QueryContext(ctx, attemptSelect+`
		 WHERE dispatch_started_at IS NOT NULL
		   AND dispatch_started_at >= ?
		   AND state != ?
		   AND id != ?
		   AND intent_id IN (
		         SELECT id FROM intents
		          WHERE upper(symbol) = upper(?) AND lower(market) = lower(?))
		 ORDER BY dispatch_started_at, rowid`,
		since, string(StateNotDispatched), exceptAttemptID, symbol, market)
	if err != nil {
		return nil, fmt.Errorf("journal: listing mutations on %s dispatched since %s: %w", symbol, since, err)
	}
	defer rows.Close()

	var out []AttemptRecord
	for rows.Next() {
		rec, err := scanAttempt(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: listing mutations on %s dispatched since %s: %w", symbol, since, err)
	}
	return out, nil
}
