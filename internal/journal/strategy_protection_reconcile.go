package journal

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/positioncampaign"
)

// ReconcileStrategyProtection blocks additional exposure without quarantining an exit.
func (j *Journal) ReconcileStrategyProtection(ctx context.Context, campaignID string, p Position) error {
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var account, market, symbol, decision, state string
	var generation sql.NullInt64
	var version int64
	if err := tx.QueryRowContext(ctx, `SELECT account_ref,market,symbol,decision_id,actual_position_generation,state,version FROM position_campaigns WHERE id=?`, campaignID).Scan(&account, &market, &symbol, &decision, &generation, &state, &version); err != nil {
		return err
	}
	if state == string(positioncampaign.CampaignClosed) || !generation.Valid || generation.Int64 != p.InstanceSeq || account != p.AccountRef || !strings.EqualFold(market, p.Market) || symbol != p.Symbol || decision != p.EntryDecisionID {
		return fmt.Errorf("strategy protection: campaign/position scope mismatch")
	}
	stored, err := scanPosition(tx.QueryRowContext(ctx, positionSelect+` WHERE id=?`, p.ID))
	if err != nil {
		return err
	}
	if stored.State == PositionClosed || stored.AccountRef != account || !strings.EqualFold(stored.Market, market) || stored.Symbol != symbol || stored.InstanceSeq != generation.Int64 || stored.EntryDecisionID != decision {
		return fmt.Errorf("strategy protection: position identity mismatch")
	}
	key := "protection-" + strconv.FormatInt(generation.Int64, 10)
	digest := digestParts(campaignID, account, market, symbol, decision, p.ID, strconv.FormatInt(generation.Int64, 10))
	if _, found, err := campaignCommandResult(ctx, tx, campaignID, string(positioncampaign.CommandProtectionReconcile), key, digest); err != nil {
		return err
	} else if found {
		return tx.Commit()
	}
	transition, err := positioncampaign.TransitionCampaign(positioncampaign.CampaignState(state), positioncampaign.CampaignLineageMismatch)
	if positioncampaign.CampaignState(state) == positioncampaign.CampaignReconcile {
		transition, err = positioncampaign.TransitionCampaign(positioncampaign.CampaignReconcile, positioncampaign.CampaignEvidenceIncomplete)
	}
	if err != nil {
		return err
	}
	now := j.nowString()
	result, err := tx.ExecContext(ctx, `UPDATE position_campaigns SET state=?,entry_blocked=1,version=version+1,updated_at=? WHERE id=? AND version=? AND state<>'CLOSED'`, string(transition.State), now, campaignID, version)
	if err != nil {
		return err
	}
	if rows, err := result.RowsAffected(); err != nil || rows != 1 {
		return ErrCampaignVersionConflict
	}
	version++
	if err := insertCampaignCommand(ctx, tx, campaignID, string(positioncampaign.CommandProtectionReconcile), key, digest, version, version, now); err != nil {
		return err
	}
	if err := insertCampaignEvent(ctx, tx, campaignID, version, version, 0, "", "PROTECTION_RECONCILE", string(positioncampaign.CommandProtectionReconcile), key, digest, transition.State, "", generation.Int64, "", "", now); err != nil {
		return err
	}
	return tx.Commit()
}
