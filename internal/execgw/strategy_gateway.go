package execgw

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/officialfx"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
)

// StrategyPlaceRequest is the paired KR/US first-leg Gateway request. Market,
// lineage, q_final and reservations are not repeated here: the Gateway reloads
// them from the decision and exact durable lease. FXAuthority is opaque and
// cannot be reconstructed from the auditable decision JSON.
type StrategyPlaceRequest struct {
	Intent      orderintent.PlaceIntent
	Decision    GuardianDecision
	Lease       journal.StrategyDispatchLeaseCAS
	FXAuthority officialfx.Evidence
	Baseline    *Baseline
	IntentID    string
	// FinalAuthorityCheck re-reads the signed scheduler authority at the last
	// no-byte-sent boundary. Production supplies a source-backed closure; nil is
	// a fail-closed request, never an implied approval.
	FinalAuthorityCheck func(context.Context) error
	// EntryGateAuthority is opaque and can only be observed from this Gateway;
	// it prevents an allowed/blocked/allowed transition from reusing a lease.
	EntryGateAuthority StrategyEntryGateAuthority

	testAccountBaseFX    risk.AccountBaseFX
	testAccountBaseFXSet bool
}

type strategyDispatchCapability struct {
	lease                journal.StrategyDispatchLeaseCAS
	fxAuthority          officialfx.Evidence
	testAccountBaseFX    risk.AccountBaseFX
	testAccountBaseFXSet bool
	finalAuthorityCheck  func(context.Context) error
	entryGateAuthority   StrategyEntryGateAuthority
}

type strategyPreTransportClaim struct {
	lease journal.StrategyDispatchLease
}

const strategyPreTransportSettlementTimeout = 15 * time.Second

// PlaceClaimedStrategy enters the existing Place/submit path with the exact
// claimed-lease and opaque-FX capability attached. It is not a second broker
// transport and it cannot mint either authority.
func (g *Gateway) PlaceClaimedStrategy(ctx context.Context, req StrategyPlaceRequest) (Outcome, error) {
	capability := &strategyDispatchCapability{
		lease: req.Lease, fxAuthority: req.FXAuthority,
		testAccountBaseFX: req.testAccountBaseFX, testAccountBaseFXSet: req.testAccountBaseFXSet,
		finalAuthorityCheck: req.FinalAuthorityCheck,
		entryGateAuthority:  req.EntryGateAuthority,
	}
	return g.place(ctx, PlaceRequest{
		Intent: req.Intent, Decision: req.Decision, Baseline: req.Baseline, IntentID: req.IntentID,
	}, capability)
}

func (g *Gateway) checkClaimedStrategyLease(ctx context.Context, decision journal.Decision, clientOrderID string, plan mutationPlan) *RejectedError {
	lease, rejected := g.loadClaimedStrategyLeaseBinding(ctx, decision, clientOrderID, plan)
	if rejected != nil {
		return rejected
	}
	reservations, err := g.journal.ReservationsForDecision(ctx, decision.ID)
	if err != nil {
		return reject(ReasonStrategyDispatchFenced, "strategy reservation binding cannot be read: %v", err)
	}
	boundHeld := false
	for _, reservation := range reservations {
		if reservation.ID == lease.RiskReservationID && reservation.DecisionID == decision.ID && reservation.Held() {
			boundHeld = true
			break
		}
	}
	if !boundHeld {
		return reject(ReasonStrategyDispatchFenced, "strategy lease does not bind the decision's exact HELD aggregate reservation")
	}
	return nil
}

func (g *Gateway) loadClaimedStrategyLeaseBinding(ctx context.Context, decision journal.Decision, clientOrderID string, plan mutationPlan) (journal.StrategyDispatchLease, *RejectedError) {
	if g == nil || g.journal == nil || plan.strategy == nil {
		return journal.StrategyDispatchLease{}, reject(ReasonStrategyDispatchAuthorityMissing, "strategy dispatch lease authority is unavailable")
	}
	var lease journal.StrategyDispatchLease
	var err error
	if g.loadStrategyDispatchLease != nil {
		lease, err = g.loadStrategyDispatchLease(ctx, plan.strategy.lease.LeaseID)
	} else {
		lease, err = g.journal.LookupStrategyDispatchLease(ctx, plan.strategy.lease.LeaseID)
	}
	if err != nil {
		return journal.StrategyDispatchLease{}, strategySubmittingRefusal(err)
	}
	if rejected := checkStrategyLeaseBinding(lease, decision, clientOrderID, plan); rejected != nil {
		return journal.StrategyDispatchLease{}, rejected
	}
	if lease.State != journal.StrategyDispatchLeaseClaimed || lease.Disposition != journal.StrategyDispatchReservationReserved ||
		lease.Revision != plan.strategy.lease.ExpectedRevision || !lease.TransportStartedAt.IsZero() {
		return journal.StrategyDispatchLease{}, reject(ReasonStrategyDispatchFenced, "strategy dispatch lease is not the exact pre-transport CLAIMED revision")
	}
	return lease, nil
}

func (g *Gateway) refuseClaimedStrategyPreTransport(ctx context.Context, claim strategyPreTransportClaim, rejected *RejectedError) error {
	if g == nil || g.journal == nil || claim.lease.LeaseID == "" {
		return journal.ErrStrategyDispatchLeaseUnavailable
	}
	settlementCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), strategyPreTransportSettlementTimeout)
	defer cancel()
	request := strategyPreTransportRefusalRequest(claim, rejected)
	var err error
	if g.refuseStrategyPreTransport != nil {
		_, err = g.refuseStrategyPreTransport(settlementCtx, request)
	} else {
		_, err = g.journal.RefuseClaimedStrategyDispatchPreTransport(settlementCtx, request)
	}
	if !errors.Is(err, journal.ErrStrategyDispatchLeaseConsumed) {
		return err
	}
	current, lookupErr := g.journal.LookupStrategyDispatchLease(settlementCtx, claim.lease.LeaseID)
	if lookupErr == nil && current.State == journal.StrategyDispatchLeaseRefused &&
		current.Disposition == journal.StrategyDispatchReservationReleased {
		return nil
	}
	if lookupErr != nil {
		return errors.Join(err, lookupErr)
	}
	return err
}

func strategyPreTransportRefusalRequest(claim strategyPreTransportClaim,
	rejected *RejectedError,
) journal.StrategyDispatchPreTransportRefusalRequest {
	return journal.StrategyDispatchPreTransportRefusalRequest{
		Lease: journal.StrategyDispatchLeaseCAS{
			LeaseID:          claim.lease.LeaseID,
			ExpectedRevision: claim.lease.Revision,
			OwnerEpoch:       claim.lease.OwnerEpoch,
			FencingToken:     claim.lease.FencingToken,
		},
		Binding: claim.lease.StrategyDispatchLeasePlan,
		Reason:  strategyPreTransportReason(rejected),
	}
}

func (g *Gateway) refusePreparedClaimedStrategyPreTransport(ctx context.Context, attempt *journal.Attempt,
	claim strategyPreTransportClaim, rejected *RejectedError,
) error {
	if g == nil || g.journal == nil || attempt == nil || claim.lease.LeaseID == "" {
		return journal.ErrStrategyDispatchLeaseUnavailable
	}
	settlementCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), strategyPreTransportSettlementTimeout)
	defer cancel()
	request := strategyPreTransportRefusalRequest(claim, rejected)
	// The seam observes legacy Gateway refusal tests whose fixture does not own
	// real v26 claim rows. Production always uses the atomic Attempt method.
	if g.refuseStrategyPreTransport != nil {
		if _, err := g.refuseStrategyPreTransport(settlementCtx, request); err != nil {
			return err
		}
		return attempt.Settle(settlementCtx, journal.StateNotDispatched, string(request.Reason), rejected.Detail)
	}
	_, err := attempt.RefuseClaimedStrategyPreTransport(settlementCtx, request, rejected.Detail)
	return err
}

func strategyPreTransportReason(rejected *RejectedError) journal.StrategyDispatchPreTransportRefusalReason {
	if rejected == nil {
		return journal.StrategyDispatchPreTransportPolicyRefused
	}
	switch rejected.Reason {
	case ReasonProtectionNotWired:
		return journal.StrategyDispatchPreTransportProtectionRefused
	case ReasonGuardianReservationMissing, ReasonGuardianRiskBucketMismatch, ReasonStrategyDispatchFenced:
		return journal.StrategyDispatchPreTransportReservationRefused
	case ReasonAccountBaseFXMismatch:
		return journal.StrategyDispatchPreTransportAccountBaseFXRefused
	case ReasonGuardianMissing, ReasonGuardianExpired, ReasonGuardianIntentMismatch,
		ReasonGuardianNonceReused, ReasonGuardianLimitsUnset, ReasonGuardianLimitExceeded,
		ReasonGuardianDecisionTampered, ReasonGuardianKeyMismatch, ReasonGuardianClassMismatch:
		return journal.StrategyDispatchPreTransportDecisionRefused
	default:
		return journal.StrategyDispatchPreTransportPolicyRefused
	}
}

func joinStrategyPreTransportRefusal(original error, releaseErr error) error {
	if releaseErr == nil {
		return original
	}
	return errors.Join(original, fmt.Errorf("execgw: terminalizing exact pre-transport strategy claim: %w", releaseErr))
}

func checkStrategyLeaseBinding(lease journal.StrategyDispatchLease, decision journal.Decision, clientOrderID string, plan mutationPlan) *RejectedError {
	request := plan.strategy
	if request == nil || plan.kind != journal.KindPlace || !plan.raisesExposure || plan.side != "BUY" ||
		lease.LeaseID != request.lease.LeaseID || lease.OwnerEpoch != request.lease.OwnerEpoch || lease.FencingToken != request.lease.FencingToken ||
		lease.GuardianDecisionID != decision.ID || lease.OperationID != clientOrderID || lease.AccountRef != decision.AccountRef ||
		!strings.EqualFold(string(lease.Market), plan.market) || lease.Symbol != plan.symbol || lease.RiskReservationID == "" {
		return reject(ReasonStrategyDispatchFenced,
			"strategy lease is not exactly bound to this exposure-raising decision, client order, account, market, symbol and reservation")
	}
	return nil
}

func strategySubmittingRefusal(err error) *RejectedError {
	switch {
	case errors.Is(err, journal.ErrStrategyDispatchFenced),
		errors.Is(err, journal.ErrStrategyDispatchLeaseConsumed),
		errors.Is(err, journal.ErrStrategyDispatchLeaseUnavailable):
		return reject(ReasonStrategyDispatchFenced, "strategy dispatch lease is stale, consumed, unavailable or owner-fenced: %v", err)
	default:
		return reject(ReasonStrategyDispatchFenced, "strategy dispatch SUBMITTING fence failed closed: %v", err)
	}
}
