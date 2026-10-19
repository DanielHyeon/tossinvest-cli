package journal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
)

// StrategyProtection is the durable admission snapshot, not current configuration.
type StrategyProtection struct {
	Legacy             bool
	Risk               RiskIntent
	Policy             *StrategyExitPolicySnapshot
	PositionGeneration int64
}

// StrategyProtection reads the decision and its binding from one database snapshot.
func (j *Journal) StrategyProtection(ctx context.Context, decisionID, campaignID string) (StrategyProtection, error) {
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return StrategyProtection{}, err
	}
	defer tx.Rollback()
	return strategyProtectionInTx(ctx, tx, decisionID, campaignID)
}

func strategyProtectionInTx(ctx context.Context, tx *sql.Tx, decisionID, campaignID string) (StrategyProtection, error) {
	var marker int
	if err := tx.QueryRowContext(ctx, `SELECT strategy_binding_version FROM decisions WHERE id=?`, decisionID).Scan(&marker); err != nil {
		return StrategyProtection{}, fmt.Errorf("strategy protection: decision marker: %w", err)
	}
	decision, err := scanDecision(tx.QueryRowContext(ctx, decisionSelect+` WHERE id=?`, decisionID))
	if err != nil {
		return StrategyProtection{}, err
	}
	preimage, err := ParsePreimage(decision.PreimageKind, decision.RiskPreimage)
	risk, ok := preimage.(RiskIntent)
	if err != nil || !ok || HashPreimage(decision.RiskPreimage) != decision.RiskHash {
		return StrategyProtection{}, errors.New("strategy protection: invalid risk intent")
	}
	var identity, boundCampaign, account, market, symbol string
	err = tx.QueryRowContext(ctx, `SELECT entry_decision_identity,campaign_id,account_ref,market,symbol FROM strategy_first_leg_bindings WHERE decision_id=?`, decisionID).Scan(&identity, &boundCampaign, &account, &market, &symbol)
	if errors.Is(err, sql.ErrNoRows) && marker == 0 {
		return StrategyProtection{Legacy: true, Risk: risk}, nil
	}
	if err != nil {
		return StrategyProtection{}, fmt.Errorf("strategy protection: binding: %w", err)
	}
	var lineage StrategyDecisionLineage
	lineage.DecisionIdentity = identity
	var created string
	err = tx.QueryRowContext(ctx, `SELECT candidate_life_id,market,symbol,threshold_version,threshold_set_digest,evidence_digest,COALESCE(consumed_evidence_snapshot_id,''),COALESCE(consumed_evidence_snapshot_digest,''),lane_id,lane_version,lane_source_digest,lane_constants_digest,entry_price,stop_price,target_price,quantity,policy_version,settings_digest,decision_payload,decision_payload_digest,activation_manifest_digest,created_at FROM strategy_decision_lineage WHERE entry_decision_identity=?`, identity).Scan(
		&lineage.CandidateLifeID, &lineage.Market, &lineage.Symbol, &lineage.ThresholdVersion, &lineage.ThresholdSetDigest, &lineage.EvidenceDigest, &lineage.ConsumedEvidenceSnapshotID, &lineage.ConsumedEvidenceSnapshotDigest, &lineage.LaneID, &lineage.LaneVersion, &lineage.LaneSourceDigest, &lineage.LaneConstantsDigest, &lineage.EntryPrice, &lineage.StopPrice, &lineage.TargetPrice, &lineage.Quantity, &lineage.PolicyVersion, &lineage.SettingsDigest, &lineage.DecisionPayload, &lineage.DecisionPayloadDigest, &lineage.ActivationManifestDigest, &created)
	if err != nil {
		return StrategyProtection{}, fmt.Errorf("strategy protection: lineage: %w", err)
	}
	lineage.CreatedAt, err = parseJournalTime(created)
	if err != nil {
		return StrategyProtection{}, err
	}
	schema, versioned, err := strategyDecisionPayloadSchema(lineage.DecisionPayload)
	if err != nil || (marker == 4) != (versioned && schema == strategyflowRiskBindingSchemaVersionV4) || (marker != 0 && marker != 4) {
		return StrategyProtection{}, errors.New("strategy protection: marker mismatch")
	}
	if marker == 0 {
		return StrategyProtection{Legacy: true, Risk: risk}, nil
	}
	if boundCampaign != campaignID || account != decision.AccountRef || account != risk.AccountRef || !strings.EqualFold(market, risk.Market) || symbol != risk.Symbol {
		return StrategyProtection{}, errors.New("strategy protection: binding scope mismatch")
	}
	if err := verifyStrategyRiskBinding(decision, lineage); err != nil {
		return StrategyProtection{}, err
	}
	var campaignAccount, campaignMarket, campaignSymbol, campaignDecision string
	var expectedGeneration int64
	if err := tx.QueryRowContext(ctx, `SELECT account_ref,market,symbol,decision_id,expected_position_generation FROM position_campaigns WHERE id=?`, campaignID).Scan(&campaignAccount, &campaignMarket, &campaignSymbol, &campaignDecision, &expectedGeneration); err != nil {
		return StrategyProtection{}, err
	}
	if campaignAccount != account || !strings.EqualFold(campaignMarket, market) || campaignSymbol != symbol || campaignDecision != decisionID {
		return StrategyProtection{}, errors.New("strategy protection: campaign scope mismatch")
	}
	payload, err := decodeStrategyflowRiskBinding(lineage.DecisionPayload)
	if err != nil {
		return StrategyProtection{}, err
	}
	projection, err := strategyflow.VerifyAcceptedProjection(string(payload.Strategyflow))
	if err != nil {
		return StrategyProtection{}, err
	}
	generation := projection.Lineage().PositionGeneration
	if generation == 0 || generation > math.MaxInt64 || expectedGeneration < 0 || expectedGeneration == math.MaxInt64 || generation != uint64(expectedGeneration)+1 {
		return StrategyProtection{}, errors.New("strategy protection: admitted generation mismatch")
	}
	return StrategyProtection{Risk: risk, Policy: payload.ExitPolicy, PositionGeneration: int64(generation)}, nil
}

// StrategyProtectionForPosition validates a position against its admission.
// The campaign id is returned on assessment failure when it could be located,
// so a caller can attempt the separately scoped durable reconciliation latch.
func (j *Journal) StrategyProtectionForPosition(ctx context.Context, p Position) (StrategyProtection, string, error) {
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return StrategyProtection{}, "", err
	}
	defer tx.Rollback()
	var campaign string
	err = tx.QueryRowContext(ctx, `SELECT id FROM position_campaigns WHERE decision_id=? AND account_ref=? AND lower(market)=lower(?) AND symbol=?`, p.EntryDecisionID, p.AccountRef, p.Market, p.Symbol).Scan(&campaign)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return StrategyProtection{}, "", err
	}
	protection, err := strategyProtectionInTx(ctx, tx, p.EntryDecisionID, campaign)
	if err != nil {
		return StrategyProtection{}, campaign, err
	}
	if protection.Legacy {
		return protection, campaign, nil
	}
	stored, err := scanPosition(tx.QueryRowContext(ctx, positionSelect+` WHERE id=?`, p.ID))
	if err != nil {
		return protection, campaign, err
	}
	var generation sql.NullInt64
	if err := tx.QueryRowContext(ctx, `SELECT actual_position_generation FROM position_campaigns WHERE id=?`, campaign).Scan(&generation); err != nil {
		return protection, campaign, err
	}
	if stored.State == PositionClosed || stored.AccountRef != p.AccountRef || !strings.EqualFold(stored.Market, p.Market) || stored.Symbol != p.Symbol || stored.InstanceSeq != p.InstanceSeq || stored.EntryDecisionID != p.EntryDecisionID || p.InstanceSeq != protection.PositionGeneration {
		return StrategyProtection{}, campaign, errors.New("strategy protection: position identity mismatch")
	}
	if !generation.Valid || generation.Int64 != p.InstanceSeq {
		return protection, campaign, errors.New("strategy protection: campaign generation mismatch")
	}
	return protection, campaign, nil
}

func requireStrategyProtectionPositionInTx(ctx context.Context, tx *sql.Tx, campaignID string, protection StrategyProtection) error {
	if protection.Legacy {
		return nil
	}
	var account, market, symbol, decision string
	var generation sql.NullInt64
	var expectedGeneration, expectedVersion int64
	if err := tx.QueryRowContext(ctx, `SELECT account_ref,market,symbol,decision_id,actual_position_generation,expected_position_generation,expected_position_version FROM position_campaigns WHERE id=?`, campaignID).Scan(&account, &market, &symbol, &decision, &generation, &expectedGeneration, &expectedVersion); err != nil {
		return err
	}
	p, err := scanPosition(tx.QueryRowContext(ctx, positionSelect+` WHERE account_ref=? AND lower(market)=lower(?) AND symbol=? ORDER BY instance_seq DESC LIMIT 1`, account, market, symbol))
	if errors.Is(err, ErrPositionNotFound) {
		if generation.Valid || expectedGeneration != 0 || expectedVersion != 0 {
			return errors.New("strategy protection: missing expected position")
		}
		return nil
	}
	if err != nil {
		return err
	}
	if p.State == PositionClosed && !generation.Valid {
		var currentVersion int64
		if err := tx.QueryRowContext(ctx, `SELECT version FROM position_projection_versions WHERE position_id=?`, p.ID).Scan(&currentVersion); err != nil {
			return fmt.Errorf("strategy protection: missing position version: %w", err)
		}
		if p.InstanceSeq != expectedGeneration || currentVersion != expectedVersion {
			return errors.New("strategy protection: prefill position generation changed")
		}
		return nil
	}
	if p.State == PositionClosed || !generation.Valid || generation.Int64 != p.InstanceSeq || p.EntryDecisionID != decision || p.InstanceSeq != protection.PositionGeneration {
		return errors.New("strategy protection: current position mismatch")
	}
	result, err := scanExitStateResult(tx.QueryRowContext(ctx, exitStateSelect+` WHERE position_id=?`, p.ID))
	if err != nil {
		return err
	}
	s := result.State
	if result.Corruption != nil || s.Completed || protection.Policy == nil || s.PositionGeneration != p.InstanceSeq || s.PolicyKind != protection.Policy.Kind || s.PolicyIdentity != protection.Policy.Identity || s.EntryPrice != protection.Risk.EntryPrice || s.InitialStop != protection.Risk.StopPrice {
		return errors.New("strategy protection: exit snapshot mismatch")
	}
	return nil
}
