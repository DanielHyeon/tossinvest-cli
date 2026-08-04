package engine

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
)

type strategyDispatchGateway interface {
	ObserveStrategyProtection(context.Context, string, uint64) (execgw.StrategyProtectionAuthority, error)
	ObserveStrategyEntryGate(context.Context, string, string) (execgw.StrategyEntryGateAuthority, error)
	PlaceClaimedStrategy(context.Context, execgw.StrategyPlaceRequest) (execgw.Outcome, error)
}

// strategyDispatchCycle is package-private and is not a trigger. It owns the
// only sequence that can turn an accepted sealed result into an official
// Gateway request once a separately authorized production worker invokes it.
type strategyDispatchCycle struct {
	journal            *journal.Journal
	gateway            strategyDispatchGateway
	firstLeg           *strategyFirstLegAdmissionBridge
	schedule           strategyScheduleAuthorityPair
	fx                 strategyFXAuthorityPair
	risk               strategyRiskAuthorityPair
	revalidateSchedule func(context.Context, StrategyMarket, strategyScheduleMarketAuthority) error

	owner *strategyDispatchOwnerCoordinator
}

type strategyDispatchOwnerCoordinator struct {
	mu    sync.Mutex
	owner journal.StrategyDispatchOwner
}

func newStrategyDispatchCycle(jrn *journal.Journal, gateway strategyDispatchGateway, firstLeg *strategyFirstLegAdmissionBridge,
	schedule strategyScheduleAuthorityPair, fx strategyFXAuthorityPair, risk strategyRiskAuthorityPair, owner *strategyDispatchOwnerCoordinator,
) *strategyDispatchCycle {
	if owner == nil {
		owner = &strategyDispatchOwnerCoordinator{}
	}
	return &strategyDispatchCycle{journal: jrn, gateway: gateway, firstLeg: firstLeg, schedule: schedule, fx: fx, risk: risk, owner: owner}
}

func (cycle *strategyDispatchCycle) dispatch(ctx context.Context, result strategyflow.Result) (execgw.Outcome, error) {
	accepted, refusal := validateStrategyFirstLegResult(result)
	if refusal.Code != "" {
		return execgw.Outcome{}, errors.New(refusal.Detail)
	}
	if cycle == nil || cycle.journal == nil || cycle.gateway == nil || cycle.firstLeg == nil || ctx == nil {
		return execgw.Outcome{}, errors.New("engine: strategy dispatch cycle unavailable")
	}
	market := StrategyMarket(accepted.market)
	schedule, fx := cycle.schedule.forMarket(market), cycle.fx.forMarket(market)
	if !schedule.snapshot.Ready || schedule.restore.Activation == nil || schedule.desired.Revision == 0 ||
		schedule.calendar.Version == "" || !fx.snapshot.Ready || !fx.read.valid {
		return execgw.Outcome{}, errors.New("engine: current market schedule or FX authority unavailable")
	}
	// Complete every fallible read-only boundary before q_final admission commits
	// the campaign, aggregate reservation and five bucket holds. The Gateway will
	// still re-check protection and entry authority twice around SUBMITTING.
	if _, err := strategyFirstLegPlaceIntent(accepted, result.Quantity); err != nil {
		return execgw.Outcome{}, err
	}
	protection, err := cycle.gateway.ObserveStrategyProtection(ctx, strings.ToLower(string(market)), result.Quantity)
	if err != nil {
		return execgw.Outcome{}, err
	}
	reconciliation, err := cycle.gateway.ObserveStrategyEntryGate(ctx, strings.ToLower(string(market)), accepted.result.Lineage.Symbol)
	if err != nil {
		return execgw.Outcome{}, err
	}
	owner, err := cycle.dispatchOwner(ctx)
	if err != nil {
		return execgw.Outcome{}, err
	}
	admitted := cycle.firstLeg.admit(ctx, result)
	if admitted.Code != StrategyFirstLegAdmitted {
		return execgw.Outcome{}, fmt.Errorf("engine: first-leg admission %s: %s", admitted.Code, admitted.Detail)
	}
	decision, err := cycle.journal.LookupDecision(ctx, admitted.Receipt.DecisionID)
	if err != nil || decision.Generation < 0 {
		return execgw.Outcome{}, errors.New("engine: authoritative Guardian decision generation unavailable")
	}
	// Journal generations are zero-based. Dispatch authority generations are
	// one-based so zero remains the fail-closed "unavailable" sentinel.
	guardianGeneration := uint64(decision.Generation) + 1
	riskGeneration := cycle.risk.forMarket(market).bundle.Generation()
	if riskGeneration == 0 {
		return execgw.Outcome{}, errors.New("engine: signed risk policy generation unavailable")
	}
	activationGeneration := schedule.restore.Activation.Generation()
	activationExpiresAt := schedule.restore.Activation.ExpiresAt()
	now := cycle.schedule.observedAt
	if activationGeneration == 0 || activationExpiresAt.IsZero() || now.IsZero() || !now.Before(activationExpiresAt) || cycle.revalidateSchedule == nil {
		return execgw.Outcome{}, errors.New("engine: signed activation lifetime or final revalidator unavailable")
	}
	evidence := journal.StrategyDispatchVerifiedEvidence{Market: journal.StrategyDispatchMarket(market),
		ActivationGeneration: activationGeneration, ActivationDigest: schedule.snapshot.ActivationManifestDigest,
		CalendarGeneration: schedule.desired.Revision, CalendarDigest: schedule.calendar.Version,
		ProtectionGeneration: protection.Generation(), ProtectionSerial: strconv.FormatUint(protection.Generation(), 10), ProtectionDigest: protection.Digest(),
		ReconciliationGeneration: reconciliation.Generation(), ReconciliationDigest: reconciliation.Digest(),
		RiskPolicyGeneration: riskGeneration, GuardianGeneration: guardianGeneration, BuildDigest: strategyRuntimeBuildDigest()}
	ttl := min(30*time.Second, activationExpiresAt.Sub(now))
	lease, err := cycle.journal.IssueVerifiedFirstLegStrategyDispatchLease(ctx, journal.VerifiedFirstLegStrategyDispatchLeaseRequest{
		Receipt: admitted.Receipt, Owner: owner, Evidence: evidence, TTL: ttl,
	})
	if err != nil {
		return execgw.Outcome{}, err
	}
	claimed, err := cycle.journal.ClaimStrategyDispatchLease(ctx, journal.StrategyDispatchLeaseCAS{
		LeaseID: lease.LeaseID, ExpectedRevision: lease.Revision, OwnerEpoch: owner.Epoch, FencingToken: owner.FencingToken,
	})
	if err != nil {
		return execgw.Outcome{}, err
	}
	intent, err := strategyFirstLegPlaceIntent(accepted, admitted.Receipt.QFinal)
	if err != nil {
		return execgw.Outcome{}, err
	}
	return cycle.gateway.PlaceClaimedStrategy(ctx, execgw.StrategyPlaceRequest{Intent: intent,
		Decision: execgw.GuardianDecision{ID: admitted.Receipt.DecisionID},
		Lease: journal.StrategyDispatchLeaseCAS{LeaseID: claimed.LeaseID, ExpectedRevision: claimed.Revision,
			OwnerEpoch: owner.Epoch, FencingToken: owner.FencingToken},
		FXAuthority: fx.read.evidence, IntentID: admitted.Receipt.AttemptID,
		EntryGateAuthority: reconciliation,
		FinalAuthorityCheck: func(checkCtx context.Context) error {
			return cycle.revalidateSchedule(checkCtx, market, schedule)
		}})
}

func (cycle *strategyDispatchCycle) dispatchOwner(ctx context.Context) (journal.StrategyDispatchOwner, error) {
	if cycle.owner == nil {
		return journal.StrategyDispatchOwner{}, errors.New("engine: strategy dispatch owner coordinator unavailable")
	}
	cycle.owner.mu.Lock()
	defer cycle.owner.mu.Unlock()
	if cycle.owner.owner.Epoch != 0 {
		return cycle.owner.owner, nil
	}
	owner, err := cycle.journal.AcquireStrategyDispatchOwner(ctx, "paired-kr-us-strategy-runtime")
	if err != nil {
		return journal.StrategyDispatchOwner{}, err
	}
	cycle.owner.owner = owner
	return owner, nil
}

func strategyFirstLegPlaceIntent(accepted strategyFirstLegAccepted, quantity uint64) (orderintent.PlaceIntent, error) {
	if quantity == 0 || quantity > 1<<53 {
		return orderintent.PlaceIntent{}, errors.New("engine: strategy quantity is outside exact Gateway boundary")
	}
	priceText, ok := accepted.result.ExecutionTerms.Entry().MajorDecimal()
	if !ok {
		return orderintent.PlaceIntent{}, errors.New("engine: strategy entry price is invalid")
	}
	price, err := strconv.ParseFloat(priceText, 64)
	if err != nil || price <= 0 {
		return orderintent.PlaceIntent{}, errors.New("engine: strategy entry price cannot reach Gateway")
	}
	roundTrip, normalized := journal.NormalizeDecimal(strconv.FormatFloat(price, 'f', -1, 64))
	if !normalized || roundTrip != priceText {
		return orderintent.PlaceIntent{}, errors.New("engine: strategy entry price loses precision at Gateway boundary")
	}
	return orderintent.NormalizePlace(orderintent.PlaceInput{Symbol: accepted.result.Lineage.Symbol,
		Market: strings.ToLower(string(accepted.market)), Side: "buy", OrderType: "limit",
		Quantity: float64(quantity), Price: price, CurrencyMode: accepted.currency})
}

var _ strategyDispatchGateway = (*execgw.Gateway)(nil)
