package performance

import (
	"context"
	"fmt"
	"time"
)

// JournalLineageReader is the dormant a047 handoff. Implementations must build
// Trade values only from persisted identifier joins; the contract has no symbol,
// time-window, or nearest-neighbour lookup method with which to guess a link.
//
// The journal schema/adapter is intentionally not implemented in a049 until the
// a045 and a047 migrations have landed and the real next schema version is known.
type JournalLineageReader interface {
	ClosedStrategyTrades(context.Context, ClosedTradeWindow) ([]Trade, error)
}

type ClosedTradeWindow struct {
	ClosedAfter      time.Time
	ClosedAtOrBefore time.Time
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
	trades, err := reader.ClosedStrategyTrades(ctx, window)
	if err != nil {
		return nil, fmt.Errorf("performance: reading exact journal lineage: %w", err)
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
