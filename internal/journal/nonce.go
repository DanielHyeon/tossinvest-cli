package journal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// nonce.go is the durable one-shot store (engine-safety "결정 nonce의 durable
// 저장", task 1.7).
//
// # Why the journal and not a map
//
// An in-memory set forgets on restart, and a decision snapshot outlives a
// process: the row is still on disk, still unexpired, and a restarted engine
// with an empty set would submit it again. The store therefore *is* the
// database, and the primary key on spent_nonces is the guarantee.
//
// # Why consumption is not its own write
//
// It is folded into the dispatch transition (durability.go). A separate write
// would leave a window where the dispatch is recorded and the consumption is
// not, and a crash inside that window produces exactly the state this whole
// contract exists to prevent: an authorisation that can be used twice.
//
// # The retention invariant
//
// A consumption record must outlive the longest decision TTL: no decision may
// survive the record of its own consumption, or pruning would silently re-arm
// it. PruneSpentNonces refuses a retention shorter than the longest TTL any
// persisted decision was issued with, rather than trusting the caller to have
// read this comment.

// ErrNonceSpent means a decision's one-shot nonce was already consumed. The
// attempt that consumed it is the record of when.
var ErrNonceSpent = errors.New("journal: decision nonce has already been spent")

// consumeDecisionNonce spends the nonce of the decision an attempt is bound to,
// inside the caller's transaction.
//
// An attempt with no decision consumes nothing: the P1-shaped paths that predate
// the decision contract have no nonce, and refusing them here would be a refusal
// with no authorisation to point at. The gateway refuses those on its own terms,
// which is where the reason code belongs.
func consumeDecisionNonce(ctx context.Context, tx *sql.Tx, attemptID, now string) error {
	var (
		nonce      string
		decisionID string
	)
	err := tx.QueryRowContext(ctx,
		`SELECT d.nonce, d.id FROM mutation_attempts a
		 JOIN decisions d ON d.id = a.decision_id
		 WHERE a.id = ?`, attemptID).Scan(&nonce, &decisionID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("journal: reading the decision nonce for attempt %s: %w", attemptID, err)
	}

	// INSERT OR IGNORE plus the affected-row count, rather than catching a
	// constraint error: the outcome is then a value this code decides on, not a
	// driver's error string.
	res, err := tx.ExecContext(ctx,
		`INSERT OR IGNORE INTO spent_nonces (nonce, decision_id, consumed_at) VALUES (?,?,?)`,
		nonce, decisionID, now)
	if err != nil {
		return fmt.Errorf("journal: spending the nonce of decision %s: %w", decisionID, err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("journal: spending the nonce of decision %s: %w", decisionID, err)
	}
	if affected != 1 {
		return fmt.Errorf("%w: decision %s", ErrNonceSpent, decisionID)
	}
	return nil
}

// NonceSpent reports whether a nonce has already been consumed.
//
// It is advisory: the authoritative check is the insert inside the dispatch
// transaction, and a caller that acted on this answer alone would have a
// check-then-act race. Its purpose is to refuse an obvious re-submission before
// an attempt is recorded for it.
func (j *Journal) NonceSpent(ctx context.Context, nonce string) (bool, error) {
	if nonce == "" {
		return false, nil
	}
	var consumedAt string
	err := j.db.QueryRowContext(ctx,
		"SELECT consumed_at FROM spent_nonces WHERE nonce = ?", nonce).Scan(&consumedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("journal: reading the spent nonce: %w", err)
	}
	return true, nil
}

// MaxDecisionTTL returns the longest validity any persisted decision was issued
// with. It is the floor the retention has to clear.
func (j *Journal) MaxDecisionTTL(ctx context.Context) (time.Duration, error) {
	rows, err := j.db.QueryContext(ctx, "SELECT issued_at, expires_at FROM decisions")
	if err != nil {
		return 0, fmt.Errorf("journal: measuring the longest decision TTL: %w", err)
	}
	defer rows.Close()

	var longest time.Duration
	for rows.Next() {
		var issuedAt, expiresAt string
		if err := rows.Scan(&issuedAt, &expiresAt); err != nil {
			return 0, fmt.Errorf("journal: measuring the longest decision TTL: %w", err)
		}
		issued, err := parseJournalTime(issuedAt)
		if err != nil {
			return 0, err
		}
		expires, err := parseJournalTime(expiresAt)
		if err != nil {
			return 0, err
		}
		if ttl := expires.Sub(issued); ttl > longest {
			longest = ttl
		}
	}
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("journal: measuring the longest decision TTL: %w", err)
	}
	return longest, nil
}

// ErrRetentionTooShort means a prune would have deleted consumption records
// while decisions issued with a longer TTL can still exist.
var ErrRetentionTooShort = errors.New("journal: nonce retention is shorter than the longest decision TTL")

// PruneSpentNonces deletes consumption records older than retention and reports
// how many it removed.
//
// It refuses, and deletes nothing, when retention is shorter than the longest
// TTL any decision on disk was issued with. That is the invariant made
// executable: a decision whose consumption record has been pruned is a decision
// that can be spent again, and "the caller passes a sensible retention" is not a
// property anything checks at 3am.
func (j *Journal) PruneSpentNonces(ctx context.Context, now time.Time, retention time.Duration) (int, error) {
	if retention <= 0 {
		return 0, fmt.Errorf("%w: retention must be positive", ErrRetentionTooShort)
	}
	longest, err := j.MaxDecisionTTL(ctx)
	if err != nil {
		return 0, err
	}
	if retention < longest {
		return 0, fmt.Errorf("%w: retention %s, longest decision TTL %s", ErrRetentionTooShort,
			retention, longest)
	}
	cutoff := formatJournalTime(now.Add(-retention))
	res, err := j.db.ExecContext(ctx, "DELETE FROM spent_nonces WHERE consumed_at < ?", cutoff)
	if err != nil {
		return 0, fmt.Errorf("journal: pruning spent nonces: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("journal: pruning spent nonces: %w", err)
	}
	return int(affected), nil
}
