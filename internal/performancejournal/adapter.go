// Package performancejournal is the narrow bridge between the authoritative,
// SELECT-only journal projection and the rebuildable performance read model.
// It owns neither a journal writer nor a performance Store, so mapping lineage
// cannot accidentally acquire order, configuration, lane, or LIVE authority.
package performancejournal

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/performance"
)

type source interface {
	ClosedStrategyTradeSources(context.Context, string, time.Time, time.Time) ([]journal.ClosedStrategyTradeSource, error)
}

type campaignSource interface {
	PositionCampaignLineage(context.Context, string) (journal.PositionCampaignLineageRead, error)
	PositionCampaign(context.Context, string) (journal.PositionCampaignRecord, error)
}

var ErrCampaignLineageConflict = errors.New("performance journal: campaign lineage conflict")

// Reader adapts only the SELECT method exposed by journal.ReadOnly.
type Reader struct{ source source }

func New(reader *journal.ReadOnly) *Reader { return &Reader{source: reader} }

// AttributionRebuild maps only facts available from the query-only journal
// surface. Current journal reads cannot expose campaign legs, staged fill
// deltas, fee/tax components or persisted FX, so those remain explicitly
// missing and no numeric attribution is manufactured.
func (r *Reader) AttributionRebuild(
	ctx context.Context,
	window performance.AttributionEvidenceWindow,
	rebuildID string,
	calculatedAt time.Time,
) (performance.AttributionRebuild, error) {
	if r == nil || r.source == nil {
		return performance.AttributionRebuild{}, errors.New("performance journal: read-only source is required")
	}
	rows, err := r.source.ClosedStrategyTradeSources(
		ctx, window.AccountRef, window.ClosedAfter, window.ClosedAtOrBefore,
	)
	if err != nil {
		return performance.AttributionRebuild{}, err
	}
	out := performance.AttributionRebuild{ID: strings.TrimSpace(rebuildID), AccountRef: strings.TrimSpace(window.AccountRef), CalculatedAt: calculatedAt.UTC()}
	campaigns, hasCampaigns := r.source.(campaignSource)
	for _, sourceRow := range rows {
		lineage := performance.CompositeLineage{
			Market: strings.ToUpper(strings.TrimSpace(sourceRow.Market)), PositionID: sourceRow.PositionID,
			CloseID: sourceRow.CloseID, PolicyID: sourceRow.PolicyID, PolicyVersion: sourceRow.PolicyVersion,
		}
		if exact := sourceRow.Lineage; exact != nil {
			lineage.CandidateID = exact.CandidateLifeID
			lineage.LaneID, lineage.LaneVersion = exact.LaneID, exact.LaneVersion
			lineage.DecisionID, lineage.AttemptID = exact.StrategyDecisionIdentity, exact.StrategyAttemptID
			lineage.OrderID, lineage.FillID = exact.BrokerOrderID, exact.FillID
			lineage.PositionID, lineage.CloseID = exact.PositionID, exact.CloseOutcomeID
		}
		if hasCampaigns {
			lineage, err = enrichCampaignLineage(ctx, campaigns, out.AccountRef, sourceRow, lineage)
			if err != nil {
				return performance.AttributionRebuild{}, err
			}
		}
		missingLineage := missingCompositeLineage(lineage)
		row := performance.NewUnavailableAttribution(performance.AttributionKey{
			Market: lineage.Market, Ticker: lineage.Ticker, LaneID: lineage.LaneID, LaneVersion: lineage.LaneVersion,
			CampaignID: lineage.CampaignID, LegID: lineage.LegID, PositionID: lineage.PositionID,
			PolicyID: lineage.PolicyID, PolicyVersion: lineage.PolicyVersion,
		}, missingLineage, []string{"close_fill_deltas", "entry_fill_deltas", "fees", "fx", "taxes"}, "", "")
		row.ObservedLineage = []performance.CompositeLineage{lineage}
		out.Unavailable = append(out.Unavailable, row)
	}
	return out, nil
}

func enrichCampaignLineage(ctx context.Context, campaigns campaignSource, accountRef string,
	sourceRow journal.ClosedStrategyTradeSource, lineage performance.CompositeLineage,
) (performance.CompositeLineage, error) {
	binding, err := campaigns.PositionCampaignLineage(ctx, sourceRow.PositionID)
	if err != nil {
		return performance.CompositeLineage{}, err
	}
	if binding.Status != journal.PositionCampaignLineageKnown {
		return lineage, nil
	}
	if binding.PositionID != sourceRow.PositionID || binding.AccountRef != accountRef ||
		!strings.EqualFold(binding.Market, lineage.Market) || strings.TrimSpace(binding.Symbol) == "" ||
		binding.PositionGeneration <= 0 || strings.TrimSpace(binding.CampaignID) == "" {
		return performance.CompositeLineage{}, ErrCampaignLineageConflict
	}
	campaign, err := campaigns.PositionCampaign(ctx, binding.CampaignID)
	if err != nil {
		return performance.CompositeLineage{}, err
	}
	if campaign.ID != binding.CampaignID || campaign.AccountRef != accountRef ||
		!strings.EqualFold(campaign.Market, lineage.Market) || campaign.Symbol != binding.Symbol ||
		campaign.ActualPositionGeneration != binding.PositionGeneration ||
		!exactOrEmpty(lineage.LaneID, campaign.LaneID) || !exactOrEmpty(lineage.LaneVersion, campaign.LaneVersion) ||
		!exactOrEmpty(lineage.DecisionID, campaign.DecisionID) {
		return performance.CompositeLineage{}, ErrCampaignLineageConflict
	}
	lineage.CampaignID = campaign.ID
	lineage.Ticker = strings.TrimSpace(campaign.Symbol)
	if lineage.LaneID == "" {
		lineage.LaneID = campaign.LaneID
	}
	if lineage.LaneVersion == "" {
		lineage.LaneVersion = campaign.LaneVersion
	}
	if lineage.DecisionID == "" {
		lineage.DecisionID = campaign.DecisionID
	}
	return lineage, nil
}

func exactOrEmpty(current, authoritative string) bool {
	return strings.TrimSpace(current) == "" || strings.TrimSpace(current) == strings.TrimSpace(authoritative)
}

func missingCompositeLineage(lineage performance.CompositeLineage) []string {
	fields := []struct{ name, value string }{
		{"candidate_id", lineage.CandidateID}, {"lane_id", lineage.LaneID}, {"lane_version", lineage.LaneVersion},
		{"campaign_id", lineage.CampaignID}, {"leg_id", lineage.LegID}, {"decision_id", lineage.DecisionID},
		{"attempt_id", lineage.AttemptID}, {"order_id", lineage.OrderID}, {"fill_id", lineage.FillID},
		{"position_id", lineage.PositionID}, {"close_id", lineage.CloseID}, {"close_leg_id", lineage.CloseLegID},
		{"policy_id", lineage.PolicyID}, {"policy_version", lineage.PolicyVersion},
	}
	missing := make([]string, 0)
	for _, field := range fields {
		if strings.TrimSpace(field.value) == "" {
			missing = append(missing, field.name)
		}
	}
	sort.Strings(missing)
	return missing
}

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
