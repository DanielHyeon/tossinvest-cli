package journal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// StrategyDispatchPreTransportRefusalReason is record-only classification from
// a Gateway validation that failed before a durable transport-start marker.
// It never participates in authority validation.
type StrategyDispatchPreTransportRefusalReason string

const (
	StrategyDispatchPreTransportDecisionRefused      StrategyDispatchPreTransportRefusalReason = "GATEWAY_DECISION_REFUSED"
	StrategyDispatchPreTransportProtectionRefused    StrategyDispatchPreTransportRefusalReason = "GATEWAY_PROTECTION_REFUSED"
	StrategyDispatchPreTransportReservationRefused   StrategyDispatchPreTransportRefusalReason = "GATEWAY_RESERVATION_REFUSED"
	StrategyDispatchPreTransportAccountBaseFXRefused StrategyDispatchPreTransportRefusalReason = "GATEWAY_ACCOUNT_BASE_FX_REFUSED"
	StrategyDispatchPreTransportPolicyRefused        StrategyDispatchPreTransportRefusalReason = "GATEWAY_POLICY_REFUSED"
)

// StrategyDispatchPreTransportRefusalRequest repeats the exact durable lease
// plan observed by the caller. Binding is comparison data only: the transaction
// reloads the lease, current owner, first-leg binding and all six holds.
type StrategyDispatchPreTransportRefusalRequest struct {
	Lease   StrategyDispatchLeaseCAS
	Binding StrategyDispatchLeasePlan
	Reason  StrategyDispatchPreTransportRefusalReason
}

// RefuseClaimedStrategyDispatchPreTransport consumes a provably not-sent
// CLAIMED lease. It has no transport, broker, activation, recovery or lease
// issuance capability.
func (j *Journal) RefuseClaimedStrategyDispatchPreTransport(ctx context.Context, request StrategyDispatchPreTransportRefusalRequest) (StrategyDispatchLease, error) {
	if j == nil || j.db == nil || !validStrategyDispatchPreTransportRefusalRequest(request) {
		return StrategyDispatchLease{}, fmt.Errorf("%w: invalid strategy dispatch pre-transport refusal", ErrInvalidRequest)
	}
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return StrategyDispatchLease{}, fmt.Errorf("journal: begin strategy dispatch pre-transport refusal: %w", err)
	}
	defer tx.Rollback()
	terminal, err := j.refuseClaimedStrategyDispatchPreTransportTx(ctx, tx, nil, request, "", j.clk.Now().UTC())
	if err != nil {
		return StrategyDispatchLease{}, err
	}
	if err := tx.Commit(); err != nil {
		return StrategyDispatchLease{}, fmt.Errorf("journal: commit strategy dispatch pre-transport refusal: %w", err)
	}
	return terminal, nil
}

// RefuseClaimedStrategyPreTransport is the post-Prepare refusal path. The core
// NOT_DISPATCHED transition and exact lease/aggregate/five-bucket release share
// one transaction, so restart cannot observe a closed core with a live claim.
func (a *Attempt) RefuseClaimedStrategyPreTransport(ctx context.Context,
	request StrategyDispatchPreTransportRefusalRequest, detail string,
) (StrategyDispatchLease, error) {
	if a == nil || a.j == nil || a.j.db == nil || !validStrategyDispatchPreTransportRefusalRequest(request) {
		return StrategyDispatchLease{}, fmt.Errorf("%w: invalid prepared strategy dispatch pre-transport refusal", ErrInvalidRequest)
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	tx, err := a.j.db.BeginTx(ctx, nil)
	if err != nil {
		return StrategyDispatchLease{}, fmt.Errorf("journal: begin prepared strategy dispatch pre-transport refusal: %w", err)
	}
	defer tx.Rollback()
	terminal, err := a.j.refuseClaimedStrategyDispatchPreTransportTx(ctx, tx, a, request, detail, a.j.clk.Now().UTC())
	if err != nil {
		return StrategyDispatchLease{}, err
	}
	if err := tx.Commit(); err != nil {
		return StrategyDispatchLease{}, fmt.Errorf("journal: commit prepared strategy dispatch pre-transport refusal: %w", err)
	}
	a.state = StateNotDispatched
	return terminal, nil
}

func (j *Journal) refuseClaimedStrategyDispatchPreTransportTx(ctx context.Context, tx *sql.Tx, attempt *Attempt,
	request StrategyDispatchPreTransportRefusalRequest, detail string, now time.Time,
) (StrategyDispatchLease, error) {
	if err := requireCurrentStrategyDispatchOwner(ctx, tx, request.Lease.OwnerEpoch, request.Lease.FencingToken); err != nil {
		return StrategyDispatchLease{}, err
	}
	lease, err := loadStrategyDispatchLease(ctx, tx, request.Lease.LeaseID)
	if errors.Is(err, sql.ErrNoRows) {
		return StrategyDispatchLease{}, ErrStrategyDispatchLeaseUnavailable
	}
	if err != nil {
		return StrategyDispatchLease{}, err
	}
	if lease.State != StrategyDispatchLeaseClaimed || lease.Disposition != StrategyDispatchReservationReserved ||
		!lease.TransportStartedAt.IsZero() || lease.Revision != request.Lease.ExpectedRevision {
		return StrategyDispatchLease{}, ErrStrategyDispatchLeaseConsumed
	}
	boundPlan, err := proveStrategyDispatchPreTransportLeasePlan(ctx, tx, request.Binding)
	if err != nil {
		return StrategyDispatchLease{}, err
	}
	if !boundPlan {
		return StrategyDispatchLease{}, fmt.Errorf("%w: pre-transport durable lease plan mismatch", ErrStrategyDispatchLeaseUnavailable)
	}
	if err := proveClaimedStrategyDispatchPreTransportAuthority(ctx, tx, lease); err != nil {
		return StrategyDispatchLease{}, err
	}
	if attempt != nil {
		authority, err := loadStrategyDispatchAttemptAuthority(ctx, tx, lease, attempt.id)
		if err != nil {
			return StrategyDispatchLease{}, err
		}
		if authority.state != string(StateRecorded) || authority.brokerOrderID != "" || authority.dispatchStartedAt != "" ||
			authority.settledAt != "" {
			return StrategyDispatchLease{}, fmt.Errorf("%w: prepared strategy core is not RECORDED", ErrStrategyDispatchLeaseConsumed)
		}
		if _, err := transitionStrategyAttemptTx(ctx, tx, attempt, StateNotDispatched, "", string(request.Reason),
			detail, true, now); err != nil {
			return StrategyDispatchLease{}, err
		}
	}
	terminal, err := refuseClaimedStrategyDispatchSubmittingTx(ctx, tx, lease, string(request.Reason), now)
	if err != nil {
		return StrategyDispatchLease{}, err
	}
	return terminal, nil
}

func validStrategyDispatchPreTransportRefusalRequest(request StrategyDispatchPreTransportRefusalRequest) bool {
	validReason := false
	switch request.Reason {
	case StrategyDispatchPreTransportDecisionRefused,
		StrategyDispatchPreTransportProtectionRefused,
		StrategyDispatchPreTransportReservationRefused,
		StrategyDispatchPreTransportAccountBaseFXRefused,
		StrategyDispatchPreTransportPolicyRefused:
		validReason = true
	}
	binding := request.Binding
	if !validReason || !validStrategyDispatchIdentity(request.Lease.LeaseID) || request.Lease.ExpectedRevision == 0 ||
		request.Lease.OwnerEpoch == 0 || !validStrategyDispatchIdentity(request.Lease.FencingToken) ||
		binding.LeaseID != request.Lease.LeaseID || binding.OwnerEpoch != request.Lease.OwnerEpoch ||
		binding.FencingToken != request.Lease.FencingToken || binding.Market != StrategyDispatchMarketKR && binding.Market != StrategyDispatchMarketUS ||
		binding.AuthorityRevision == 0 || binding.IssuedAt.IsZero() || binding.ExpiresAt.IsZero() || !binding.IssuedAt.Before(binding.ExpiresAt) {
		return false
	}
	for _, value := range []string{
		binding.OperationID, binding.AccountRef, binding.Symbol, binding.CandidateID, binding.EvidenceDigest,
		binding.RouterID, binding.RouterVersion, binding.LaneID, binding.LaneVersion, binding.CampaignID,
		binding.LegID, binding.RiskReservationID, binding.GuardianDecisionID, binding.AuthorityDigest,
	} {
		if !validStrategyDispatchIdentity(value) {
			return false
		}
	}
	return true
}

func proveStrategyDispatchPreTransportLeasePlan(ctx context.Context, tx *sql.Tx, plan StrategyDispatchLeasePlan) (bool, error) {
	return strategyDispatchClaimExists(ctx, tx, `SELECT EXISTS(
		SELECT 1 FROM strategy_dispatch_leases
		WHERE lease_id=? AND operation_id=? AND account_ref=? AND market=? AND symbol=?
		  AND candidate_id=? AND evidence_digest=? AND router_id=? AND router_version=?
		  AND lane_id=? AND lane_version=? AND campaign_id=? AND leg_id=?
		  AND risk_reservation_id=? AND guardian_decision_id=?
		  AND owner_epoch=? AND fencing_token=? AND authority_revision=? AND authority_digest=?
		  AND issued_at=? AND expires_at=?)`,
		plan.LeaseID, plan.OperationID, plan.AccountRef, plan.Market, plan.Symbol,
		plan.CandidateID, plan.EvidenceDigest, plan.RouterID, plan.RouterVersion,
		plan.LaneID, plan.LaneVersion, plan.CampaignID, plan.LegID,
		plan.RiskReservationID, plan.GuardianDecisionID, plan.OwnerEpoch, plan.FencingToken,
		plan.AuthorityRevision, plan.AuthorityDigest, formatJournalTime(plan.IssuedAt), formatJournalTime(plan.ExpiresAt))
}

func proveClaimedStrategyDispatchPreTransportAuthority(ctx context.Context, tx *sql.Tx, lease StrategyDispatchLease) error {
	var prospectiveToken string
	err := tx.QueryRowContext(ctx, `SELECT binding.prospective_token
		FROM strategy_first_leg_bindings binding
		JOIN risk_bucket_final_decisions q ON q.decision_id=binding.decision_id
		JOIN strategy_attempt_lineage attempt ON attempt.attempt_id=binding.attempt_id
		WHERE binding.decision_id=? AND binding.aggregate_reservation_id=?
		  AND binding.account_ref=? AND binding.market=? AND binding.symbol=?
		  AND binding.candidate_id=? AND binding.evidence_digest=?
		  AND binding.router_id=? AND binding.router_version=?
		  AND binding.lane_id=? AND binding.lane_version=?
		  AND binding.campaign_id=? AND binding.leg_plan_id=?
		  AND q.existing_reservation_id=binding.aggregate_reservation_id
		  AND q.account_ref=binding.account_ref AND q.market=binding.market AND q.symbol=binding.symbol
		  AND q.owner_lane_id=binding.lane_id AND q.owner_campaign_id=binding.campaign_id
		  AND q.owner_prospective_generation=binding.prospective_token
		  AND q.q_final=binding.q_final
		  AND attempt.risk_intent_id=binding.decision_id
		  AND attempt.guardian_decision_id=binding.decision_id
		  AND attempt.account_ref=binding.account_ref
		  AND attempt.entry_decision_identity=binding.entry_decision_identity
		  AND attempt.client_order_id=? AND attempt.state='PLANNED'`,
		lease.GuardianDecisionID, lease.RiskReservationID, lease.AccountRef, lease.Market, lease.Symbol,
		lease.CandidateID, lease.EvidenceDigest, lease.RouterID, lease.RouterVersion, lease.LaneID, lease.LaneVersion,
		lease.CampaignID, lease.LegID, lease.OperationID).Scan(&prospectiveToken)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w: pre-transport first-leg binding mismatch", ErrStrategyDispatchLeaseUnavailable)
	}
	if err != nil {
		return fmt.Errorf("journal: proving pre-transport first-leg binding: %w", err)
	}
	var aggregateTotal, aggregateNormalizable int
	if err := tx.QueryRowContext(ctx, `SELECT count(*),
		COALESCE(sum(CASE WHEN state IN ('HELD','RELEASED') THEN 1 ELSE 0 END),0)
		FROM risk_reservations WHERE id=? AND decision_id=? AND account_ref=?`,
		lease.RiskReservationID, lease.GuardianDecisionID, lease.AccountRef).
		Scan(&aggregateTotal, &aggregateNormalizable); err != nil {
		return fmt.Errorf("journal: proving pre-transport aggregate hold: %w", err)
	}
	var buckets, dimensions, normalizable int
	if err := tx.QueryRowContext(ctx, `SELECT count(*),count(DISTINCT bucket_dimension),
		COALESCE(sum(CASE WHEN
			(state='HELD' AND held_minor=reserved_minor AND filled_minor='0' AND overage_minor='0') OR
			(state='RELEASED' AND held_minor='0' AND filled_minor='0' AND overage_minor='0')
		THEN 1 ELSE 0 END),0)
		FROM risk_bucket_reservations
		WHERE decision_id=? AND existing_reservation_id=? AND account_ref=? AND market=? AND symbol=?
		  AND owner_prospective_generation=?
		  AND bucket_dimension IN ('horizon','market','strategy','sector','symbol')`,
		lease.GuardianDecisionID, lease.RiskReservationID, lease.AccountRef, lease.Market, lease.Symbol, prospectiveToken).
		Scan(&buckets, &dimensions, &normalizable); err != nil {
		return fmt.Errorf("journal: proving pre-transport bucket holds: %w", err)
	}
	var mappedOrders int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM risk_bucket_orders WHERE decision_id=?`,
		lease.GuardianDecisionID).Scan(&mappedOrders); err != nil {
		return fmt.Errorf("journal: proving absence of a pre-transport risk order: %w", err)
	}
	if aggregateTotal != 1 || aggregateNormalizable != 1 || buckets != 5 || dimensions != 5 ||
		normalizable != 5 || mappedOrders != 0 {
		return fmt.Errorf("%w: pre-transport normalization proof aggregate=%d/%d buckets=%d dimensions=%d normalizable=%d mapped_orders=%d want=1/1/5/5/5/0",
			ErrStrategyDispatchLeaseUnavailable, aggregateTotal, aggregateNormalizable, buckets, dimensions, normalizable, mappedOrders)
	}
	return nil
}
