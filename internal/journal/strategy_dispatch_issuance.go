package journal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// StrategyDispatchVerifiedEvidence is audit metadata collected from sealed
// production authorities. It cannot create a lease by itself: issuance derives
// the entire order lineage from the immutable first-leg rows and requires the
// current journal-owned owner fence and all six real HELD reservations.
type StrategyDispatchVerifiedEvidence struct {
	Market StrategyDispatchMarket

	ActivationGeneration uint64
	ActivationDigest     string

	CalendarGeneration uint64
	CalendarDigest     string

	ProtectionGeneration uint64
	ProtectionSerial     string
	ProtectionDigest     string

	ReconciliationGeneration uint64
	ReconciliationDigest     string

	RiskPolicyGeneration uint64
	GuardianGeneration   uint64
	BuildDigest          string
}

type VerifiedFirstLegStrategyDispatchLeaseRequest struct {
	Receipt  QFinalCampaignFirstLegReceipt
	Owner    StrategyDispatchOwner
	Evidence StrategyDispatchVerifiedEvidence
	TTL      time.Duration
}

type derivedFirstLegDispatch struct {
	plan                       StrategyDispatchLeasePlan
	activationDigest           string
	riskDigest, guardianDigest string
	guardianGeneration         uint64
	decisionExpiresAt          time.Time
}

// IssueVerifiedFirstLegStrategyDispatchLease is the sole production mint. It
// commits the current market authority and its first ISSUED lease atomically.
// The older caller-plan APIs remain dormant.
func (j *Journal) IssueVerifiedFirstLegStrategyDispatchLease(ctx context.Context, request VerifiedFirstLegStrategyDispatchLeaseRequest) (StrategyDispatchLease, error) {
	if j == nil || j.db == nil || ctx == nil || request.TTL <= 0 || request.TTL > time.Minute {
		return StrategyDispatchLease{}, fmt.Errorf("%w: invalid verified dispatch issuance", ErrInvalidRequest)
	}
	if err := validateStrategyDispatchVerifiedEvidence(request.Evidence); err != nil {
		return StrategyDispatchLease{}, err
	}
	if request.Owner.Epoch == 0 || !validStrategyDispatchIdentity(request.Owner.FencingToken) {
		return StrategyDispatchLease{}, fmt.Errorf("%w: invalid verified dispatch owner", ErrInvalidRequest)
	}
	now := j.clk.Now().UTC()
	if now.IsZero() {
		return StrategyDispatchLease{}, fmt.Errorf("%w: verified dispatch time unavailable", ErrInvalidRequest)
	}
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return StrategyDispatchLease{}, fmt.Errorf("journal: begin verified dispatch issuance: %w", err)
	}
	defer tx.Rollback()
	if err := requireCurrentStrategyDispatchOwner(ctx, tx, request.Owner.Epoch, request.Owner.FencingToken); err != nil {
		return StrategyDispatchLease{}, err
	}
	derived, err := deriveFirstLegDispatchTx(ctx, tx, request.Receipt)
	if err != nil {
		return StrategyDispatchLease{}, err
	}
	if derived.plan.Market != request.Evidence.Market {
		return StrategyDispatchLease{}, fmt.Errorf("%w: verified dispatch market mismatch", ErrStrategyDispatchLeaseUnavailable)
	}
	expires := now.Add(request.TTL)
	if derived.decisionExpiresAt.Before(expires) {
		expires = derived.decisionExpiresAt
	}
	if !expires.After(now) {
		return StrategyDispatchLease{}, fmt.Errorf("%w: Guardian authority expired before dispatch issuance", ErrStrategyDispatchLeaseUnavailable)
	}
	authority, err := commitDerivedStrategyDispatchMarketAuthorityTx(ctx, tx, derived, request.Evidence, now)
	if err != nil {
		return StrategyDispatchLease{}, err
	}
	plan := derived.plan
	plan.OwnerEpoch, plan.FencingToken = request.Owner.Epoch, request.Owner.FencingToken
	plan.AuthorityRevision, plan.AuthorityDigest = authority.Revision, authority.RecordDigest
	plan.IssuedAt, plan.ExpiresAt = now, expires
	plan.LeaseID = "strategy-lease:" + digestParts(plan.GuardianDecisionID, plan.OperationID, plan.RiskReservationID)[:40]
	leaseDigest := strategyDispatchLeaseIssuanceDigest(plan)
	if prior, priorErr := loadStrategyDispatchLease(ctx, tx, plan.LeaseID); priorErr == nil {
		if prior.LeaseDigest != leaseDigest || prior.StrategyDispatchLeasePlan != plan || prior.State != StrategyDispatchLeaseIssued || prior.Revision != 1 {
			return StrategyDispatchLease{}, fmt.Errorf("%w: divergent verified dispatch replay", ErrStrategyDispatchLeaseConsumed)
		}
		return prior, nil
	} else if !errors.Is(priorErr, sql.ErrNoRows) {
		return StrategyDispatchLease{}, priorErr
	}
	nowText := formatJournalTime(now)
	_, err = tx.ExecContext(ctx, `INSERT INTO strategy_dispatch_leases(
		lease_id,operation_id,account_ref,market,symbol,candidate_id,evidence_digest,router_id,router_version,
		lane_id,lane_version,campaign_id,leg_id,risk_reservation_id,guardian_decision_id,owner_epoch,fencing_token,
		authority_revision,authority_digest,issued_at,expires_at,state,disposition,revision,transport_started_at,
		refusal_code,outcome_code,broker_order_id,query_digest,outcome_observed_at,lease_digest,created_at,updated_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,'ISSUED','RESERVED',1,NULL,'','','','',NULL,?,?,?)`,
		plan.LeaseID, plan.OperationID, plan.AccountRef, plan.Market, plan.Symbol, plan.CandidateID, plan.EvidenceDigest,
		plan.RouterID, plan.RouterVersion, plan.LaneID, plan.LaneVersion, plan.CampaignID, plan.LegID,
		plan.RiskReservationID, plan.GuardianDecisionID, plan.OwnerEpoch, plan.FencingToken, plan.AuthorityRevision,
		plan.AuthorityDigest, formatJournalTime(plan.IssuedAt), formatJournalTime(plan.ExpiresAt), leaseDigest, nowText, nowText)
	if err != nil {
		return StrategyDispatchLease{}, fmt.Errorf("journal: inserting verified strategy dispatch lease: %w", err)
	}
	lease, err := loadStrategyDispatchLease(ctx, tx, plan.LeaseID)
	if err != nil {
		return StrategyDispatchLease{}, err
	}
	if err := tx.Commit(); err != nil {
		return StrategyDispatchLease{}, fmt.Errorf("journal: commit verified dispatch issuance: %w", err)
	}
	return lease, nil
}

func deriveFirstLegDispatchTx(ctx context.Context, tx *sql.Tx, receipt QFinalCampaignFirstLegReceipt) (derivedFirstLegDispatch, error) {
	var out derivedFirstLegDispatch
	var market, expires string
	var generation int
	err := tx.QueryRowContext(ctx, `SELECT b.account_ref,b.market,b.symbol,b.candidate_id,b.evidence_digest,
		b.router_id,b.router_version,b.lane_id,b.lane_version,b.campaign_id,b.leg_plan_id,
		b.aggregate_reservation_id,b.decision_id,a.client_order_id,a.activation_manifest_digest,
		q.snapshot_set_digest,d.generation,d.risk_hash,d.expires_at
		FROM strategy_first_leg_bindings b
		JOIN strategy_attempt_lineage a ON a.attempt_id=b.attempt_id
		JOIN risk_bucket_final_decisions q ON q.decision_id=b.decision_id
		JOIN decisions d ON d.id=b.decision_id
		WHERE b.decision_id=? AND b.aggregate_reservation_id=? AND b.campaign_id=? AND b.attempt_id=?
		  AND b.leg_plan_id=? AND b.leg_sequence=? AND b.q_final=?`, receipt.DecisionID,
		receipt.AggregateReservationID, receipt.CampaignID, receipt.AttemptID, receipt.FirstLegPlanID,
		receipt.LegSequence, receipt.QFinal).Scan(&out.plan.AccountRef, &market, &out.plan.Symbol,
		&out.plan.CandidateID, &out.plan.EvidenceDigest, &out.plan.RouterID, &out.plan.RouterVersion,
		&out.plan.LaneID, &out.plan.LaneVersion, &out.plan.CampaignID, &out.plan.LegID,
		&out.plan.RiskReservationID, &out.plan.GuardianDecisionID, &out.plan.OperationID,
		&out.activationDigest, &out.riskDigest, &generation, &out.guardianDigest, &expires)
	if errors.Is(err, sql.ErrNoRows) {
		return derivedFirstLegDispatch{}, fmt.Errorf("%w: exact first-leg receipt unavailable", ErrStrategyDispatchLeaseUnavailable)
	}
	if err != nil {
		return derivedFirstLegDispatch{}, fmt.Errorf("journal: deriving first-leg dispatch: %w", err)
	}
	out.plan.Market = StrategyDispatchMarket(market)
	out.guardianGeneration = uint64(generation) + 1
	out.decisionExpiresAt, err = parseJournalTime(expires)
	if err != nil {
		return derivedFirstLegDispatch{}, err
	}
	if receipt.Market != market || receipt.Symbol != out.plan.Symbol || receipt.RouterID != out.plan.RouterID ||
		receipt.RouterVersion != out.plan.RouterVersion || out.plan.OperationID == "" || out.activationDigest == "" ||
		(out.plan.Market != StrategyDispatchMarketKR && out.plan.Market != StrategyDispatchMarketUS) {
		return derivedFirstLegDispatch{}, fmt.Errorf("%w: first-leg receipt scope changed", ErrStrategyDispatchLeaseUnavailable)
	}
	var aggregateHeld, monetaryHeld, dimensions int
	if err := tx.QueryRowContext(ctx, `SELECT
		(SELECT count(*) FROM risk_reservations WHERE id=? AND decision_id=? AND account_ref=? AND state='HELD'),
		(SELECT count(*) FROM risk_bucket_reservations WHERE decision_id=? AND existing_reservation_id=?
		 AND account_ref=? AND market=? AND symbol=? AND state='HELD' AND held_minor=reserved_minor),
		(SELECT count(DISTINCT bucket_dimension) FROM risk_bucket_reservations WHERE decision_id=?
		 AND existing_reservation_id=? AND state='HELD' AND held_minor=reserved_minor
		 AND bucket_dimension IN ('horizon','market','strategy','sector','symbol'))`,
		out.plan.RiskReservationID, out.plan.GuardianDecisionID, out.plan.AccountRef,
		out.plan.GuardianDecisionID, out.plan.RiskReservationID, out.plan.AccountRef, out.plan.Market, out.plan.Symbol,
		out.plan.GuardianDecisionID, out.plan.RiskReservationID).Scan(&aggregateHeld, &monetaryHeld, &dimensions); err != nil {
		return derivedFirstLegDispatch{}, err
	}
	if aggregateHeld != 1 || monetaryHeld != 5 || dimensions != 5 {
		return derivedFirstLegDispatch{}, fmt.Errorf("%w: exact first-leg holds unavailable", ErrStrategyDispatchLeaseUnavailable)
	}
	return out, nil
}

func commitDerivedStrategyDispatchMarketAuthorityTx(ctx context.Context, tx *sql.Tx, derived derivedFirstLegDispatch,
	evidence StrategyDispatchVerifiedEvidence, now time.Time,
) (StrategyDispatchMarketAuthority, error) {
	result := StrategyDispatchMarketAuthority{AccountRef: derived.plan.AccountRef, Market: derived.plan.Market, Symbol: derived.plan.Symbol,
		ActivationGeneration: evidence.ActivationGeneration, ActivationDigest: derived.activationDigest,
		CalendarGeneration: evidence.CalendarGeneration, ProtectionGeneration: evidence.ProtectionGeneration,
		ProtectionSerial: evidence.ProtectionSerial, ProtectionDigest: evidence.ProtectionDigest,
		ReconciliationGeneration: evidence.ReconciliationGeneration, RiskPolicyGeneration: evidence.RiskPolicyGeneration,
		RiskPolicyDigest: derived.riskDigest, GuardianGeneration: derived.guardianGeneration,
		GuardianDigest: derived.guardianDigest, BuildDigest: evidence.BuildDigest, UpdatedAt: now}
	if evidence.ActivationDigest != result.ActivationDigest || evidence.GuardianGeneration != result.GuardianGeneration {
		return StrategyDispatchMarketAuthority{}, fmt.Errorf("%w: activation or Guardian authority mismatch", ErrStrategyDispatchLeaseUnavailable)
	}
	result.RecordDigest = digestParts("strategy-dispatch-market-authority:v1", result.AccountRef, string(result.Market), result.Symbol,
		strconv.FormatUint(result.ActivationGeneration, 10), result.ActivationDigest,
		strconv.FormatUint(result.CalendarGeneration, 10), evidence.CalendarDigest,
		strconv.FormatUint(result.ProtectionGeneration, 10), result.ProtectionSerial, result.ProtectionDigest,
		strconv.FormatUint(result.ReconciliationGeneration, 10), evidence.ReconciliationDigest,
		strconv.FormatUint(result.RiskPolicyGeneration, 10), result.RiskPolicyDigest,
		strconv.FormatUint(result.GuardianGeneration, 10), result.GuardianDigest, result.BuildDigest)
	authorityID := "strategy-authority:" + digestParts(result.AccountRef, string(result.Market), result.Symbol)[:40]
	var priorDigest string
	var priorRevision uint64
	err := tx.QueryRowContext(ctx, `SELECT record_digest,revision FROM strategy_dispatch_market_authorities
		WHERE account_ref=? AND market=? AND symbol=?`, result.AccountRef, result.Market, result.Symbol).Scan(&priorDigest, &priorRevision)
	if err == nil && priorDigest == result.RecordDigest {
		result.Revision = priorRevision
		return result, nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return StrategyDispatchMarketAuthority{}, err
	}
	if errors.Is(err, sql.ErrNoRows) {
		result.Revision = 1
		_, err = tx.ExecContext(ctx, `INSERT INTO strategy_dispatch_market_authorities(
			authority_id,account_ref,market,symbol,activation_generation,activation_digest,calendar_generation,
			protection_generation,protection_serial,protection_digest,reconciliation_generation,risk_policy_generation,
			risk_policy_digest,guardian_generation,guardian_digest,build_digest,revision,record_digest,updated_at)
			VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, authorityID, result.AccountRef, result.Market, result.Symbol,
			result.ActivationGeneration, result.ActivationDigest, result.CalendarGeneration, result.ProtectionGeneration,
			result.ProtectionSerial, result.ProtectionDigest, result.ReconciliationGeneration, result.RiskPolicyGeneration,
			result.RiskPolicyDigest, result.GuardianGeneration, result.GuardianDigest, result.BuildDigest, result.Revision,
			result.RecordDigest, formatJournalTime(now))
	} else {
		result.Revision = priorRevision + 1
		_, err = tx.ExecContext(ctx, `UPDATE strategy_dispatch_market_authorities SET
			activation_generation=?,activation_digest=?,calendar_generation=?,protection_generation=?,protection_serial=?,
			protection_digest=?,reconciliation_generation=?,risk_policy_generation=?,risk_policy_digest=?,guardian_generation=?,
			guardian_digest=?,build_digest=?,revision=?,record_digest=?,updated_at=?
			WHERE account_ref=? AND market=? AND symbol=? AND revision=?`, result.ActivationGeneration, result.ActivationDigest,
			result.CalendarGeneration, result.ProtectionGeneration, result.ProtectionSerial, result.ProtectionDigest,
			result.ReconciliationGeneration, result.RiskPolicyGeneration, result.RiskPolicyDigest, result.GuardianGeneration,
			result.GuardianDigest, result.BuildDigest, result.Revision, result.RecordDigest, formatJournalTime(now),
			result.AccountRef, result.Market, result.Symbol, priorRevision)
	}
	if err != nil {
		return StrategyDispatchMarketAuthority{}, fmt.Errorf("journal: committing derived market authority: %w", err)
	}
	return result, nil
}

func validateStrategyDispatchVerifiedEvidence(e StrategyDispatchVerifiedEvidence) error {
	if e.Market != StrategyDispatchMarketKR && e.Market != StrategyDispatchMarketUS ||
		e.ActivationGeneration == 0 || e.CalendarGeneration == 0 || e.ProtectionGeneration == 0 || e.ReconciliationGeneration == 0 ||
		e.RiskPolicyGeneration == 0 || e.GuardianGeneration == 0 || !validStrategyDispatchIdentity(e.ProtectionSerial) {
		return fmt.Errorf("%w: incomplete verified dispatch evidence", ErrInvalidRequest)
	}
	for _, digest := range []string{e.ActivationDigest, e.CalendarDigest, e.ProtectionDigest, e.ReconciliationDigest, e.BuildDigest} {
		if len(digest) != 71 || !strings.HasPrefix(digest, "sha256:") || strings.Trim(digest[7:], "0123456789abcdef") != "" {
			return fmt.Errorf("%w: invalid verified dispatch digest", ErrInvalidRequest)
		}
	}
	return nil
}

func strategyDispatchLeaseIssuanceDigest(plan StrategyDispatchLeasePlan) string {
	return digestParts("strategy-dispatch-lease:v1", plan.LeaseID, plan.OperationID, plan.AccountRef, string(plan.Market), plan.Symbol,
		plan.CandidateID, plan.EvidenceDigest, plan.RouterID, plan.RouterVersion, plan.LaneID, plan.LaneVersion,
		plan.CampaignID, plan.LegID, plan.RiskReservationID, plan.GuardianDecisionID,
		strconv.FormatUint(plan.OwnerEpoch, 10), plan.FencingToken, strconv.FormatUint(plan.AuthorityRevision, 10),
		plan.AuthorityDigest, formatJournalTime(plan.IssuedAt), formatJournalTime(plan.ExpiresAt))
}
