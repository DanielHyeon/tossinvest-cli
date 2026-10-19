package journal

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

// ErrContinuationUnavailable means the requested current campaign cannot be proven.
var ErrContinuationUnavailable = errors.New("journal: continuation snapshot unavailable")

type ContinuationPlanAvailability string

const FullPlanUnavailable ContinuationPlanAvailability = "FULL_PLAN_PREIMAGE_UNAVAILABLE"

type ContinuationSnapshotRequest struct {
	AccountRef, Market, Symbol, CampaignID string
	ExpectedGeneration                     int64
}

type ContinuationSnapshot struct {
	Campaign          PositionCampaignRecord
	Position          Position
	ProjectionVersion int64
	Legs              []CampaignLegRecord
	Watermarks        []CampaignOrderRecord
	PlanAvailability  ContinuationPlanAvailability
}

// ReadCampaignContinuationSnapshot returns stored facts, never entry authority.
func (r *ReadOnly) ReadCampaignContinuationSnapshot(ctx context.Context, req ContinuationSnapshotRequest) (ContinuationSnapshot, error) {
	if r == nil || r.db == nil || ctx == nil || strings.TrimSpace(req.AccountRef) == "" || strings.TrimSpace(req.CampaignID) == "" || strings.TrimSpace(req.Symbol) == "" || req.ExpectedGeneration <= 0 {
		return ContinuationSnapshot{}, ErrInvalidRequest
	}
	req.Market = normaliseMarket(req.Market)
	req.Symbol = normaliseSymbol(req.Symbol)
	if req.Market != "kr" && req.Market != "us" {
		return ContinuationSnapshot{}, ErrInvalidRequest
	}
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return ContinuationSnapshot{}, err
	}
	defer tx.Rollback()
	result, err := readCampaignContinuationSnapshot(ctx, tx, req)
	if err != nil {
		return ContinuationSnapshot{}, err
	}
	if err = tx.Commit(); err != nil {
		return ContinuationSnapshot{}, err
	}
	return result, nil
}

func readCampaignContinuationSnapshot(ctx context.Context, tx *sql.Tx, req ContinuationSnapshotRequest) (ContinuationSnapshot, error) {
	unavailable := func() (ContinuationSnapshot, error) { return ContinuationSnapshot{}, ErrContinuationUnavailable }
	c, err := positionCampaignQuery(ctx, tx, req.CampaignID)
	if errors.Is(err, ErrCampaignNotFound) {
		return unavailable()
	}
	if err != nil {
		return ContinuationSnapshot{}, err
	}
	if c.AccountRef != req.AccountRef || normaliseMarket(c.Market) != req.Market || c.Symbol != req.Symbol || (c.State == "CLOSED" || c.State == "PLANNED") || c.ActualPositionGeneration != req.ExpectedGeneration {
		return unavailable()
	}
	p, err := scanPosition(tx.QueryRowContext(ctx, positionSelect+` WHERE account_ref=? AND market=? AND symbol=? ORDER BY instance_seq DESC LIMIT 1`, req.AccountRef, req.Market, req.Symbol))
	if errors.Is(err, ErrPositionNotFound) {
		return unavailable()
	}
	if err != nil {
		return ContinuationSnapshot{}, err
	}
	if p.InstanceSeq != req.ExpectedGeneration || p.State != "OPEN" || strings.TrimSpace(p.EntryDecisionID) == "" || p.EntryDecisionID != c.DecisionID {
		return unavailable()
	}
	var version int64
	err = tx.QueryRowContext(ctx, `SELECT version FROM position_projection_versions WHERE position_id=? AND account_ref=? AND market=? AND symbol=? AND generation=? AND state=?`, p.ID, p.AccountRef, p.Market, p.Symbol, p.InstanceSeq, p.State).Scan(&version)
	if errors.Is(err, sql.ErrNoRows) {
		return unavailable()
	}
	if err != nil {
		return ContinuationSnapshot{}, err
	}
	if version <= 0 {
		return unavailable()
	}
	result := ContinuationSnapshot{Campaign: c, Position: p, ProjectionVersion: version, PlanAvailability: FullPlanUnavailable, Legs: []CampaignLegRecord{}, Watermarks: []CampaignOrderRecord{}}
	legs, err := tx.QueryContext(ctx, `SELECT l.campaign_id,l.sequence,l.plan_id,coalesce(l.intent_id,''),l.requested_quantity,l.filled_quantity,l.residual_quantity,l.state,l.version,c.version,l.created_at,l.updated_at FROM campaign_legs l JOIN position_campaigns c ON c.id=l.campaign_id WHERE l.campaign_id=? ORDER BY l.sequence`, c.ID)
	if err != nil {
		return ContinuationSnapshot{}, err
	}
	for legs.Next() {
		v, e := scanCampaignLeg(legs)
		if e != nil {
			legs.Close()
			return ContinuationSnapshot{}, e
		}
		result.Legs = append(result.Legs, v)
	}
	err = legs.Err()
	legs.Close()
	if err != nil {
		return ContinuationSnapshot{}, err
	}
	if len(result.Legs) == 0 {
		return unavailable()
	}
	var foreignScope int
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM campaign_order_watermarks WHERE campaign_id=? AND (account_ref<>? OR market<>? OR symbol<>?))`, c.ID, c.AccountRef, c.Market, c.Symbol).Scan(&foreignScope); err != nil {
		return ContinuationSnapshot{}, err
	}
	if foreignScope != 0 {
		return unavailable()
	}
	orders, err := tx.QueryContext(ctx, `SELECT w.campaign_id,w.leg_sequence,w.order_id,w.intent_id,w.attempt_id,coalesce(w.predecessor_order_id,''),w.carry_baseline,w.requested_cap,w.cumulative_filled,w.remaining_quantity,w.terminal,w.lineage_ambiguous,coalesce(w.last_observation_id,''),c.version,w.created_at,w.updated_at FROM campaign_order_watermarks w JOIN position_campaigns c ON c.id=w.campaign_id WHERE w.campaign_id=? ORDER BY w.leg_sequence,w.order_id`, c.ID)
	if err != nil {
		return ContinuationSnapshot{}, err
	}
	known := make(map[int64]bool, len(result.Legs))
	for _, l := range result.Legs {
		known[l.Sequence] = true
	}
	for orders.Next() {
		v, e := scanCampaignOrder(orders)
		if e != nil {
			orders.Close()
			return ContinuationSnapshot{}, e
		}
		if !known[v.LegSequence] {
			orders.Close()
			return unavailable()
		}
		result.Watermarks = append(result.Watermarks, v)
	}
	err = orders.Err()
	orders.Close()
	if err != nil {
		return ContinuationSnapshot{}, err
	}
	return result, nil
}
