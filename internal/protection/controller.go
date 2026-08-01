package protection

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

var (
	ErrProtectionInactive = errors.New("protection: activation is OFF")
	ErrProtectionGap      = errors.New("protection: broker protection did not become ACTIVE inside the deadline")
	ErrMutationInDoubt    = errors.New("protection: broker mutation result is in doubt")
)

type Readiness string

const (
	ReadinessUnwired Readiness = "UNWIRED"
	ReadinessWired   Readiness = "WIRED"
)

// AssessReadiness is display/admission evidence, not an activation capability.
// A valid signed profile is necessary, but the caller must also prove the saga
// wiring exists. The shipped execgw constant remains UNWIRED and this function
// has no path that changes it.
func AssessReadiness(verified attest.VerifiedProtectionCapability, requested attest.ProtectionScope, sagaWired bool) Readiness {
	if !sagaWired || !sameAttestedScope(verified.Scope(), requested) {
		return ReadinessUnwired
	}
	capability := verified.Capability()
	if !capability.Persistence.SurvivesProcessExit || !capability.Persistence.SurvivesRestart ||
		!capability.Reservation.ReservesSellableQuantity || !capability.Idempotency.Create ||
		!capability.Idempotency.ClientOrderID || !capability.Replace.ContinuousCoverage ||
		!capability.Replace.NewIdentifierRecorded ||
		(capability.Replace.Mode != attest.ReplaceAtomic && capability.Replace.Mode != attest.ReplaceContinuous) {
		return ReadinessUnwired
	}
	return ReadinessWired
}

func sameAttestedScope(left, right attest.ProtectionScope) bool {
	if left.AccountRef != right.AccountRef || left.Profile != right.Profile || left.Market != right.Market ||
		left.Session != right.Session || left.ConditionalType != right.ConditionalType || left.OrderType != right.OrderType ||
		left.TriggerSource != right.TriggerSource || left.Quantity != right.Quantity || len(left.Tools) != len(right.Tools) {
		return false
	}
	for tool, build := range left.Tools {
		if right.Tools[tool] != build {
			return false
		}
	}
	return true
}

// Activation is deliberately opaque. This change ships no exported minter,
// provisioning reader, config key, or command route. Consequently the public
// constructor below fails closed in production until a separately reviewed
// operator-approval change supplies an activation capability.
type Activation struct {
	ready bool
	scope Scope
}

type SellableReader interface {
	Sellable(context.Context, Scope, string) (SellableObservation, error)
}

type ExecutionGateway interface {
	Gateway
	SellableReader
}

type Controller struct {
	repository *Repository
	gateway    ExecutionGateway
	scope      Scope
	now        func() time.Time
	newID      func() string
	entryOpen  atomic.Bool
}

func NewController(repository *Repository, gateway ExecutionGateway, activation Activation, now func() time.Time, newID func() string) (*Controller, error) {
	if repository == nil || gateway == nil || !activation.ready || activation.scope.Validate() != nil {
		return nil, ErrProtectionInactive
	}
	if now == nil || newID == nil {
		return nil, errors.New("protection: clock and identity source are required")
	}
	c := &Controller{repository: repository, gateway: gateway, scope: activation.scope, now: now, newID: newID}
	c.entryOpen.Store(true)
	return c, nil
}

func (c *Controller) EntryAllowed() bool { return c != nil && c.entryOpen.Load() }

type Fill struct {
	At         time.Time
	Quantity   int64
	Trigger    int64
	ExpireDate string
}

func (c *Controller) PlanFill(ctx context.Context, fill Fill) (Saga, error) {
	if c == nil || fill.At.IsZero() || fill.Quantity < 1 || fill.Trigger < 1 || !validExpireDate(fill.ExpireDate) {
		return Saga{}, ErrInvalidSaga
	}
	id := c.newID()
	clientID := c.newID()
	saga := Saga{
		ID: id, AccountRef: c.scope.AccountRef, Profile: c.scope.Profile, Market: c.scope.Market,
		Symbol: c.scope.Symbol, Generation: 1, Revision: 1, State: StatePlanned,
		Trigger: fill.Trigger, Quantity: fill.Quantity, ClientOrderID: clientID, UpdatedAt: fill.At,
	}
	if err := c.repository.Insert(ctx, saga); err != nil {
		return Saga{}, err
	}
	return saga, nil
}

func (c *Controller) Register(ctx context.Context, sagaID, attemptID, expireDate string, fillAt time.Time) (Saga, error) {
	saga, err := c.repository.Get(ctx, sagaID)
	if err != nil {
		return Saga{}, err
	}
	now := c.now()
	if now.Before(fillAt) || now.Sub(fillAt) > RegistrationArmDeadline {
		c.entryOpen.Store(false)
		_, _ = c.repository.MarkDiscrepancy(ctx, saga.ID, saga.Revision, now, "PROTECTION_GAP_ARM_DEADLINE")
		return Saga{}, ErrProtectionGap
	}
	body := bodyForSaga(saga, expireDate)
	canonical, err := body.CanonicalJSON()
	if err != nil {
		return Saga{}, err
	}
	attempt := MutationAttempt{
		ID: attemptID, SagaID: saga.ID, Generation: saga.Generation, Kind: MutationCreate,
		State: MutationPlanned, SerializerVersion: SerializerVersion, CanonicalBody: string(canonical),
		CreatedAt: now, UpdatedAt: now,
	}
	if err := c.repository.recordAttempt(ctx, attempt); err != nil {
		return Saga{}, err
	}
	saga, err = c.repository.BeginRegistration(ctx, saga.ID, saga.Revision, now, attemptID)
	if err != nil {
		return Saga{}, err
	}
	if err := c.repository.markAttempt(ctx, attemptID, MutationPlanned, MutationDispatched, now, ""); err != nil {
		return Saga{}, err
	}
	broker, callErr := c.gateway.Create(ctx, body)
	settledAt := c.now()
	if callErr != nil || !exactBrokerResult(saga, broker) {
		c.entryOpen.Store(false)
		_ = c.repository.markAttempt(ctx, attemptID, MutationDispatched, MutationInDoubt, settledAt, broker.ID)
		_, _ = c.repository.MarkMutationUnknown(ctx, saga.ID, saga.Revision, settledAt, attemptID)
		return Saga{}, fmt.Errorf("%w: create: %v", ErrMutationInDoubt, callErr)
	}
	if err := c.repository.markAttempt(ctx, attemptID, MutationDispatched, MutationAcknowledged, settledAt, broker.ID); err != nil {
		c.entryOpen.Store(false)
		return Saga{}, err
	}
	saga, err = c.repository.MarkRegistrationActive(ctx, saga.ID, saga.Revision, settledAt, attemptID, broker.ID)
	if err != nil {
		c.entryOpen.Store(false)
		return Saga{}, err
	}
	if EvaluateArm(fillAt, now, settledAt) != ArmActive {
		c.entryOpen.Store(false)
		_, _ = c.repository.MarkDiscrepancy(ctx, saga.ID, saga.Revision, settledAt, "PROTECTION_GAP_ACTIVE_DEADLINE")
		return Saga{}, ErrProtectionGap
	}
	return saga, nil
}

func (c *Controller) Replace(ctx context.Context, sagaID, attemptID string, trigger, quantity int64, expireDate string) (Saga, error) {
	saga, err := c.repository.Get(ctx, sagaID)
	if err != nil {
		return Saga{}, err
	}
	if trigger < saga.Trigger {
		return Saga{}, fmt.Errorf("%w: %d is below active trigger %d", ErrWeakerProtection, trigger, saga.Trigger)
	}
	if quantity < 1 {
		return Saga{}, ErrInvalidBody
	}
	body := bodyForSaga(saga, expireDate)
	body.Trigger, body.Quantity = trigger, quantity
	canonical, err := body.CanonicalJSON()
	if err != nil {
		return Saga{}, err
	}
	now := c.now()
	if err := c.repository.recordAttempt(ctx, MutationAttempt{
		ID: attemptID, SagaID: saga.ID, Generation: saga.Generation, Kind: MutationReplace,
		State: MutationPlanned, SerializerVersion: SerializerVersion, CanonicalBody: string(canonical),
		TargetBrokerID: saga.BrokerID, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		return Saga{}, err
	}
	saga, err = c.repository.BeginReplace(ctx, saga.ID, saga.Revision, now, attemptID, trigger, quantity)
	if err != nil {
		return Saga{}, err
	}
	if err := c.repository.markAttempt(ctx, attemptID, MutationPlanned, MutationDispatched, now, ""); err != nil {
		return Saga{}, err
	}
	broker, callErr := c.gateway.Replace(ctx, saga.BrokerID, body)
	settledAt := c.now()
	if callErr != nil || !exactPendingBrokerResult(saga, broker) {
		c.entryOpen.Store(false)
		_ = c.repository.markAttempt(ctx, attemptID, MutationDispatched, MutationInDoubt, settledAt, broker.ID)
		_, _ = c.repository.MarkMutationUnknown(ctx, saga.ID, saga.Revision, settledAt, attemptID)
		return Saga{}, fmt.Errorf("%w: replace: %v", ErrMutationInDoubt, callErr)
	}
	if err := c.repository.markAttempt(ctx, attemptID, MutationDispatched, MutationAcknowledged, settledAt, broker.ID); err != nil {
		c.entryOpen.Store(false)
		return Saga{}, err
	}
	return c.repository.MarkReplaceActive(ctx, saga.ID, saga.Revision, settledAt, attemptID, broker.ID)
}

// ReplaceFromExitSnapshot is the typed a041 bridge. The broker trigger comes
// only from the immutable snapshot's current effective protection; callers
// cannot independently recompute or substitute a second exit line.
func (c *Controller) ReplaceFromExitSnapshot(ctx context.Context, sagaID, attemptID string, snapshot exitpolicy.ExitLineSnapshot, quantity int64, expireDate string) (Saga, error) {
	trigger, err := TriggerFromExitSnapshot(snapshot)
	if err != nil {
		return Saga{}, err
	}
	return c.Replace(ctx, sagaID, attemptID, trigger, quantity, expireDate)
}

func TriggerFromExitSnapshot(snapshot exitpolicy.ExitLineSnapshot) (int64, error) {
	if snapshot.SnapshotID == "" || snapshot.DecisionID == "" || snapshot.InputDigest == "" || snapshot.PositionID == "" || snapshot.PositionGeneration < 1 || snapshot.ObservationID == "" {
		return 0, fmt.Errorf("%w: exit snapshot identity is incomplete", ErrInvalidBody)
	}
	trigger, err := strconv.ParseInt(snapshot.CurrentProtection, 10, 64)
	if err != nil || trigger < 1 || strconv.FormatInt(trigger, 10) != snapshot.CurrentProtection {
		return 0, fmt.Errorf("%w: current protection is not a canonical integer", ErrInvalidBody)
	}
	return trigger, nil
}

func (c *Controller) Reconcile(ctx context.Context) ([]Discrepancy, error) {
	local, err := c.repository.List(ctx, c.scope)
	if err != nil {
		c.entryOpen.Store(false)
		return nil, err
	}
	broker, err := c.gateway.List(ctx, c.scope)
	if err != nil {
		c.entryOpen.Store(false)
		return nil, err
	}
	discrepancies, err := Compare(c.scope, local, broker)
	if err != nil {
		c.entryOpen.Store(false)
		return nil, err
	}
	if len(discrepancies) == 0 {
		return nil, nil
	}
	c.entryOpen.Store(false)
	now := c.now()
	for _, discrepancy := range discrepancies {
		if discrepancy.SagaID == "" {
			continue
		}
		for _, saga := range local {
			if saga.ID == discrepancy.SagaID && saga.State != StateClosed {
				_, _ = c.repository.MarkDiscrepancy(ctx, saga.ID, saga.Revision, now, string(discrepancy.Kind))
			}
		}
	}
	return discrepancies, nil
}

func (c *Controller) Recover(ctx context.Context, sagaID string) (Saga, error) {
	saga, err := c.repository.Get(ctx, sagaID)
	if err != nil {
		return Saga{}, err
	}
	broker, err := c.gateway.List(ctx, c.scope)
	if err != nil {
		c.entryOpen.Store(false)
		return Saga{}, err
	}
	var matched *BrokerProtection
	for i := range broker {
		candidate := &broker[i]
		if (saga.BrokerID != "" && candidate.ID == saga.BrokerID) || candidate.ClientOrderID == saga.ClientOrderID {
			if matched != nil {
				c.entryOpen.Store(false)
				return Saga{}, ErrDuplicateBrokerID
			}
			matched = candidate
		}
	}
	now := c.now()
	if matched == nil {
		c.entryOpen.Store(false)
		return c.repository.MarkDiscrepancy(ctx, saga.ID, saga.Revision, now, string(DiscrepancyMissing))
	}
	if matched.Triggered {
		c.entryOpen.Store(false)
		return c.repository.RecoverTriggered(ctx, saga.ID, saga.Revision, now, matched.ID)
	}
	if matched.Terminal {
		return c.repository.RecoverClosed(ctx, saga.ID, saga.Revision, now, matched.ID)
	}
	return c.repository.RecoverActive(ctx, saga.ID, saga.Revision, now, matched.ID, matched.Trigger, matched.Quantity)
}

func (c *Controller) AuthorizeFlatten(ctx context.Context, sagaID, attemptID string, start time.Time, required int64) (FlattenAuthorization, error) {
	saga, err := c.repository.Get(ctx, sagaID)
	if err != nil {
		return FlattenAuthorization{}, err
	}
	now := c.now()
	canonical := `{"broker_id":` + strconv.Quote(saga.BrokerID) + `}`
	if err := c.repository.recordAttempt(ctx, MutationAttempt{
		ID: attemptID, SagaID: saga.ID, Generation: saga.Generation, Kind: MutationCancel,
		State: MutationPlanned, SerializerVersion: SerializerVersion, CanonicalBody: canonical,
		TargetBrokerID: saga.BrokerID, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		return FlattenAuthorization{}, err
	}
	saga, err = c.repository.BeginCancel(ctx, saga.ID, saga.Revision, now, attemptID, saga.BrokerID)
	if err != nil {
		return FlattenAuthorization{}, err
	}
	if err := c.repository.markAttempt(ctx, attemptID, MutationPlanned, MutationDispatched, now, ""); err != nil {
		return FlattenAuthorization{}, err
	}
	cancel, callErr := c.gateway.Cancel(ctx, saga.BrokerID)
	if callErr != nil || !cancel.Terminal || cancel.Triggered {
		c.entryOpen.Store(false)
		at := c.now()
		_ = c.repository.markAttempt(ctx, attemptID, MutationDispatched, MutationInDoubt, at, saga.BrokerID)
		_, _ = c.repository.MarkMutationUnknown(ctx, saga.ID, saga.Revision, at, attemptID)
		return FlattenAuthorization{}, ErrMutationInDoubt
	}
	sellable, err := c.gateway.Sellable(ctx, c.scope, saga.BrokerID)
	decisionAt := c.now()
	deadline := start.Add(2 * time.Second)
	decision, authorization := decideFlatten(start, deadline, decisionAt,
		FlattenScope{Scope: c.scope, BrokerID: saga.BrokerID}, cancel, sellable, required, c.now)
	if err != nil || decision != FlattenAllowed {
		c.entryOpen.Store(false)
		_ = c.repository.markAttempt(ctx, attemptID, MutationDispatched, MutationInDoubt, decisionAt, saga.BrokerID)
		_, _ = c.repository.MarkMutationUnknown(ctx, saga.ID, saga.Revision, decisionAt, attemptID)
		return FlattenAuthorization{}, ErrMutationInDoubt
	}
	if err := c.repository.markAttempt(ctx, attemptID, MutationDispatched, MutationClosed, decisionAt, saga.BrokerID); err != nil {
		return FlattenAuthorization{}, err
	}
	if _, err := c.repository.MarkCancelClosed(ctx, saga.ID, saga.Revision, decisionAt, attemptID, saga.BrokerID); err != nil {
		return FlattenAuthorization{}, err
	}
	return authorization, nil
}

func DesiredProtectionQuantity(holding, openSell, localReservation int64) (int64, error) {
	if holding < 1 || openSell < 0 || localReservation < 0 || openSell > holding || localReservation > holding-openSell {
		return 0, ErrOversell
	}
	return holding - openSell - localReservation, nil
}

func bodyForSaga(saga Saga, expireDate string) ConditionalBody {
	return ConditionalBody{
		SerializerVersion: SerializerVersion, ClientOrderID: saga.ClientOrderID,
		AccountRef: saga.AccountRef, Market: string(saga.Market), Symbol: saga.Symbol, Side: "SELL",
		ConditionalType: "SINGLE", OrderType: "MARKET", TriggerSource: "LAST_TRADE",
		Trigger: saga.Trigger, Quantity: saga.Quantity, ExpireDate: expireDate,
	}
}

func exactBrokerResult(saga Saga, broker BrokerProtection) bool {
	return broker.ID != "" && !broker.Terminal && !broker.Triggered && broker.Scope.equal(Scope{
		AccountRef: saga.AccountRef, Profile: saga.Profile, Market: saga.Market, Symbol: saga.Symbol,
	}) && broker.Quantity == saga.Quantity && broker.Trigger == saga.Trigger &&
		broker.ClientOrderID == saga.ClientOrderID
}

func exactPendingBrokerResult(saga Saga, broker BrokerProtection) bool {
	copy := saga
	copy.Trigger, copy.Quantity = saga.PendingTrigger, saga.PendingQuantity
	return exactBrokerResult(copy, broker)
}
