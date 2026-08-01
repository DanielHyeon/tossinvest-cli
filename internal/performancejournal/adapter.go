// Package performancejournal is the narrow bridge between the authoritative,
// SELECT-only journal projection and the rebuildable performance read model.
// It owns neither a journal writer nor a performance Store, so mapping lineage
// cannot accidentally acquire order, configuration, lane, or LIVE authority.
package performancejournal

import (
	"context"
	"errors"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/performance"
)

type source interface {
	ClosedStrategyTradeSources(context.Context, string, time.Time, time.Time) ([]journal.ClosedStrategyTradeSource, error)
}

// Reader adapts only the SELECT method exposed by journal.ReadOnly.
type Reader struct{ source source }

func New(reader *journal.ReadOnly) *Reader { return &Reader{source: reader} }

func (r *Reader) ClosedStrategyTrades(
	ctx context.Context,
	window performance.ClosedTradeWindow,
) ([]performance.Trade, error) {
	if r == nil || r.source == nil {
		return nil, errors.New("performance journal: read-only source is required")
	}
	rows, err := r.source.ClosedStrategyTradeSources(
		ctx, window.AccountRef, window.ClosedAfter, window.ClosedAtOrBefore,
	)
	if err != nil {
		return nil, err
	}
	out := make([]performance.Trade, 0, len(rows))
	for _, row := range rows {
		lineage := performance.Lineage{
			PositionID:    row.PositionID,
			CloseID:       row.CloseID,
			PolicyID:      row.PolicyID,
			PolicyVersion: row.PolicyVersion,
		}
		if exact := row.Lineage; exact != nil {
			lineage.CandidateLifeID = exact.CandidateLifeID
			lineage.ThresholdVersion = exact.ThresholdVersion
			lineage.ThresholdSetDigest = exact.ThresholdSetDigest
			lineage.EvidenceDigest = exact.EvidenceDigest
			lineage.LaneID = exact.LaneID
			lineage.LaneVersion = exact.LaneVersion
			lineage.DecisionID = exact.StrategyDecisionIdentity
			lineage.RiskIntentID = exact.RiskIntentID
			lineage.AttemptID = exact.StrategyAttemptID
			lineage.MutationAttemptID = exact.MutationAttemptID
			lineage.OrderID = exact.BrokerOrderID
			lineage.FillID = exact.FillID
			lineage.PositionID = exact.PositionID
			lineage.CloseID = exact.CloseOutcomeID
		}
		cost := ""
		if row.CostTotal != nil {
			cost = *row.CostTotal
		}
		out = append(out, performance.Trade{
			ID: row.TradeID, Lineage: lineage, Market: row.Market, Side: performance.Side(row.Side),
			DecisionAt: row.DecisionAt, DecisionPrice: row.DecisionPrice,
			EntryAt: row.EntryAt, EntryPrice: row.EntryPrice, Quantity: row.Quantity, CostTotal: cost,
			RealizedPnLAfterCosts: row.RealizedPnLAfterCosts, RealizedR: row.RealizedR, ClosedAt: row.ClosedAt,
		})
	}
	return out, nil
}

var _ performance.JournalLineageReader = (*Reader)(nil)
