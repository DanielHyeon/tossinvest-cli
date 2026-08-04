package journal

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

const (
	krWeeklyLaneID = "kr_weekly_disclosure_value_v1"
	usWeeklyLaneID = "us_weekly_disclosure_value_v1"
)

func cloneWeeklyFirstLegBinding(binding *WeeklyFirstLegReservationBinding) *WeeklyFirstLegReservationBinding {
	if binding == nil {
		return nil
	}
	copy := *binding
	return &copy
}

func validateWeeklyFirstLegBindingShape(laneID, market, campaignID string, binding *WeeklyFirstLegReservationBinding) error {
	wantWeekly := (market == "KR" && laneID == krWeeklyLaneID) || (market == "US" && laneID == usWeeklyLaneID)
	if !wantWeekly {
		if binding != nil {
			return ErrWeeklyReservationConflict
		}
		return nil
	}
	if binding == nil || binding.PlannedOrdinal != 1 || binding.ScopeVersion == 0 {
		return ErrWeeklyReservationConflict
	}
	for _, value := range []string{campaignID, binding.ReservationID, binding.StableWeek, binding.RequestDigest, binding.RecordDigest,
		binding.CalendarGeneration, binding.CalendarDigest} {
		if strings.TrimSpace(value) != value || value == "" || len(value) > 256 {
			return ErrWeeklyReservationConflict
		}
	}
	return nil
}

func validateWeeklyFirstLegBindingTx(ctx context.Context, tx *sql.Tx, prepared preparedQFinalFirstLeg) error {
	if prepared.weekly == nil {
		return nil
	}
	binding := prepared.weekly
	var count int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM strategy_weekly_market_reservations r
		JOIN strategy_weekly_reservation_scopes s ON s.campaign_id=r.campaign_id AND s.market=r.market
		WHERE r.reservation_id=? AND r.campaign_id=? AND r.market=? AND r.stable_week=?
		AND r.planned_ordinal=? AND r.status='ACTIVE' AND r.scope_version=? AND s.version=r.scope_version
		AND s.positive_leg_count=0 AND r.request_digest=? AND r.record_digest=?
		AND r.calendar_generation=? AND r.calendar_digest=?`, binding.ReservationID, prepared.campaign.CampaignID,
		strings.ToUpper(prepared.strategyPlan.Lineage.Market), binding.StableWeek, binding.PlannedOrdinal, binding.ScopeVersion,
		binding.RequestDigest, binding.RecordDigest, binding.CalendarGeneration, binding.CalendarDigest).Scan(&count); err != nil {
		return err
	}
	if count != 1 {
		return ErrWeeklyReservationConflict
	}
	return nil
}

func insertWeeklyFirstLegBindingTx(ctx context.Context, tx *sql.Tx, prepared preparedQFinalFirstLeg, now string) error {
	if prepared.weekly == nil {
		return nil
	}
	binding := prepared.weekly
	recordDigest := digestParts("strategy-weekly-first-leg-binding:v1", prepared.decision.ID, binding.ReservationID,
		prepared.campaign.CampaignID, strings.ToUpper(prepared.strategyPlan.Lineage.Market), binding.StableWeek,
		fmt.Sprint(binding.PlannedOrdinal), fmt.Sprint(binding.ScopeVersion), binding.RequestDigest, binding.RecordDigest,
		binding.CalendarGeneration, binding.CalendarDigest, prepared.firstLegDigest)
	_, err := tx.ExecContext(ctx, `INSERT INTO strategy_weekly_first_leg_bindings(decision_id,reservation_id,campaign_id,market,stable_week,
		planned_ordinal,scope_version,request_digest,reservation_record_digest,calendar_generation,calendar_digest,binding_record_digest,created_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`, prepared.decision.ID, binding.ReservationID, prepared.campaign.CampaignID,
		strings.ToUpper(prepared.strategyPlan.Lineage.Market), binding.StableWeek, binding.PlannedOrdinal, binding.ScopeVersion,
		binding.RequestDigest, binding.RecordDigest, binding.CalendarGeneration, binding.CalendarDigest, recordDigest, now)
	if err != nil {
		return fmt.Errorf("journal: insert exact weekly first-leg binding: %w", err)
	}
	return nil
}

func verifyWeeklyFirstLegReplayTx(ctx context.Context, tx *sql.Tx, prepared preparedQFinalFirstLeg) error {
	if prepared.weekly == nil {
		var count int
		if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM strategy_weekly_first_leg_bindings WHERE decision_id=?`, prepared.decision.ID).Scan(&count); err != nil {
			return err
		}
		if count != 0 {
			return fmt.Errorf("%w: unexpected weekly first-leg authority", ErrRiskBucketReplayMismatch)
		}
		return nil
	}
	binding := prepared.weekly
	var count int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM strategy_weekly_first_leg_bindings w
		JOIN strategy_weekly_market_reservations r ON r.reservation_id=w.reservation_id
		WHERE w.decision_id=? AND w.reservation_id=? AND w.campaign_id=? AND w.market=? AND w.stable_week=?
		AND w.planned_ordinal=? AND w.scope_version=? AND w.request_digest=? AND w.reservation_record_digest=?
		AND w.calendar_generation=? AND w.calendar_digest=? AND r.status='ACTIVE'`, prepared.decision.ID,
		binding.ReservationID, prepared.campaign.CampaignID, strings.ToUpper(prepared.strategyPlan.Lineage.Market), binding.StableWeek,
		binding.PlannedOrdinal, binding.ScopeVersion, binding.RequestDigest, binding.RecordDigest, binding.CalendarGeneration, binding.CalendarDigest).Scan(&count); err != nil {
		return err
	}
	if count != 1 {
		return fmt.Errorf("%w: weekly first-leg replay authority", ErrRiskBucketReplayMismatch)
	}
	return nil
}
