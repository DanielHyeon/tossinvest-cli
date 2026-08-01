package performance

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// JournalLineageReader is the dormant a047 handoff. Implementations must build
// Trade values only from persisted identifier joins; the contract has no symbol,
// time-window, or nearest-neighbour lookup method with which to guess a link.
//
// The concrete bridge lives in internal/performancejournal so this package's
// dependency closure cannot acquire journal-adjacent policy/write authority.
type JournalLineageReader interface {
	ClosedStrategyTrades(context.Context, ClosedTradeWindow) ([]Trade, error)
}

type ClosedTradeWindow struct {
	// AccountRef is selected by the server/session wiring. It is not accepted
	// from an arbitrary UI field, and prevents a multi-account journal from
	// silently combining two owners' lane results.
	AccountRef       string
	ClosedAfter      time.Time
	ClosedAtOrBefore time.Time
}

func (w ClosedTradeWindow) validate() error {
	if strings.TrimSpace(w.AccountRef) == "" {
		return errors.New("performance: journal account scope is required")
	}
	if w.ClosedAfter.IsZero() || w.ClosedAtOrBefore.IsZero() || !w.ClosedAfter.Before(w.ClosedAtOrBefore) {
		return errors.New("performance: closed trade window must have ordered non-zero bounds")
	}
	return nil
}

// CollectClosedStrategyTrades persists journal-derived lineage and measurements
// from observations the caller already owns. existingObservations is a value,
// not a reader capability: this boundary cannot turn a refresh into an API poll.
func (s *Store) CollectClosedStrategyTrades(
	ctx context.Context,
	reader JournalLineageReader,
	window ClosedTradeWindow,
	existingObservations map[string][]Observation,
	calculatedAt time.Time,
) ([]Snapshot, error) {
	if reader == nil {
		return nil, fmt.Errorf("performance: journal lineage reader is required")
	}
	if err := window.validate(); err != nil {
		return nil, err
	}
	trades, err := reader.ClosedStrategyTrades(ctx, window)
	if err != nil {
		return nil, fmt.Errorf("performance: reading exact journal lineage: %w", err)
	}
	if err := s.bindJournalAccount(ctx, window.AccountRef); err != nil {
		return nil, err
	}
	snapshots := make([]Snapshot, 0, len(trades))
	for _, trade := range trades {
		snapshot, err := s.Collect(ctx, trade, existingObservations[trade.Lineage.PositionID], calculatedAt)
		if err != nil {
			return snapshots, err
		}
		snapshots = append(snapshots, snapshot)
	}
	return snapshots, nil
}

func (s *Store) bindJournalAccount(ctx context.Context, accountRef string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("performance: starting journal account binding: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO performance_scope(singleton,account_ref) VALUES(1,?)`, strings.TrimSpace(accountRef)); err != nil {
		return fmt.Errorf("performance: binding journal account: %w", err)
	}
	var bound string
	if err := tx.QueryRowContext(ctx, `SELECT account_ref FROM performance_scope WHERE singleton=1`).Scan(&bound); err != nil {
		return fmt.Errorf("performance: reading journal account binding: %w", err)
	}
	if bound != strings.TrimSpace(accountRef) {
		return fmt.Errorf("performance: store is bound to a different journal account")
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("performance: committing journal account binding: %w", err)
	}
	return nil
}
