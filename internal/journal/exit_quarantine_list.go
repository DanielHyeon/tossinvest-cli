package journal

// exit_quarantine_list.go is the read side of change a079: the list an operator
// console shows before deciding whether to lift a quarantine.
//
// It is a new file rather than another method on account_views.go because the
// audience is different. accountActiveQuarantines answers "is this position's
// protection line trustworthy" for a dashboard row; this answers "what exactly
// would I be releasing, and what does the stored snapshot still protect".

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitquarantine"
)

// ActiveExitSnapshotQuarantines lists every unreleased quarantine that belongs
// to a position's *current* generation, with the protection line its stored
// snapshot still holds.
//
// # The generation predicate
//
// The active index is on (position_id, position_generation), so a position that
// was released and re-adopted keeps the old instance's quarantine row on disk
// forever — unreleased, because nothing releases the quarantine of a generation
// that no longer exists. The engine's judgement gate asks about the current
// generation only (exitloop.go: ActiveExitSnapshotQuarantine(ctx, p.ID,
// p.InstanceSeq)); offering a release for any other row would hand the operator
// a button that changes nothing they can observe.
//
// # Why the protection line is best-effort
//
// A quarantine whose reason is stored_snapshot_corrupt has, by definition, a
// snapshot that may not decode. That is not a reason to refuse the whole list —
// the operator still needs to see the row in order to act on it — so a snapshot
// that cannot be read yields an explicit unknown reason rather than an error or,
// worse, a blank that reads as "no protection".
func (j *Journal) ActiveExitSnapshotQuarantines(ctx context.Context) ([]exitquarantine.Row, error) {
	rows, err := j.db.QueryContext(ctx, `SELECT q.position_id,p.market,p.symbol,
		 q.position_generation,q.quarantine_version,q.reason,q.evidence,q.quarantined_at,
		 e.effective_snapshot_json
		 FROM exit_snapshot_quarantines q
		 JOIN positions p ON p.id = q.position_id
		 LEFT JOIN exit_states e ON e.position_id = q.position_id
		 WHERE q.released_at IS NULL AND q.position_generation = p.instance_seq
		 ORDER BY q.quarantined_at, q.position_id`)
	if err != nil {
		return nil, fmt.Errorf("journal: listing active exit snapshot quarantines: %w", err)
	}
	defer rows.Close()

	var out []exitquarantine.Row
	for rows.Next() {
		var (
			row      exitquarantine.Row
			snapshot sql.NullString
		)
		if err := rows.Scan(&row.PositionID, &row.Market, &row.Symbol, &row.Generation,
			&row.Version, &row.Reason, &row.Evidence, &row.QuarantinedAt, &snapshot); err != nil {
			return nil, fmt.Errorf("journal: reading an active exit snapshot quarantine: %w", err)
		}
		row.Protection, row.ProtectionUnknown = quarantineProtectionLine(snapshot)
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: listing active exit snapshot quarantines: %w", err)
	}
	return out, nil
}

// quarantineProtectionLine reports the protection the stored effective snapshot
// keeps, or why it cannot be shown. It never returns both, and never returns
// neither.
func quarantineProtectionLine(raw sql.NullString) (protection, unknown string) {
	if !raw.Valid {
		return "", "no_effective_snapshot"
	}
	stored, err := decodeStoredSnapshot(raw.String)
	if err != nil {
		return "", "unreadable_effective_snapshot"
	}
	if stored.Line.CurrentProtection == "" {
		return "", "snapshot_has_no_protection"
	}
	return stored.Line.CurrentProtection, ""
}
