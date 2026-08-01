package protection

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

var (
	ErrProtectionInactive = errors.New("protection: activation is OFF")
	ErrProtectionGap      = errors.New("protection: broker protection did not become ACTIVE inside the deadline")
	ErrMutationInDoubt    = errors.New("protection: broker mutation result is in doubt")
	ErrProtectionGone     = errors.New("protection: broker protection disappeared or closed without a trigger")
)

const brokerRecoveryDeadline = 2 * time.Second

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
	opMu       sync.Mutex
	confirmed  map[string]activeConfirmation
}

type activeConfirmation struct {
	revision   int64
	generation int64
	brokerID   string
}

func NewController(repository *Repository, gateway ExecutionGateway, activation Activation, now func() time.Time, newID func() string) (*Controller, error) {
	if repository == nil || gateway == nil || !activation.ready || activation.scope.Validate() != nil {
		return nil, ErrProtectionInactive
	}
	if now == nil || newID == nil {
		return nil, errors.New("protection: clock and identity source are required")
	}
	c := &Controller{repository: repository, gateway: gateway, scope: activation.scope, now: now, newID: newID, confirmed: make(map[string]activeConfirmation)}
	// Construction never proves broker inventory. Even an empty local database
	// may coexist with an orphaned broker order, so only a bounded authoritative
	// reconciliation/recovery may open this latch.
	c.entryOpen.Store(false)
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
	if c == nil {
		return Saga{}, ErrProtectionInactive
	}
	c.opMu.Lock()
	defer c.opMu.Unlock()
	c.closeEntry("")
	if fill.At.IsZero() || fill.Quantity < 1 || fill.Trigger < 1 || !validExpireDate(fill.ExpireDate) {
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
	c.opMu.Lock()
	defer c.opMu.Unlock()
	c.closeEntry(sagaID)
	saga, err := c.repository.Get(ctx, sagaID)
	if err != nil {
		return Saga{}, err
	}
	durableFillAt := saga.UpdatedAt
	now := c.now()
	if !fillAt.Equal(durableFillAt) || now.Before(durableFillAt) || now.Sub(durableFillAt) > RegistrationArmDeadline {
		c.entryOpen.Store(false)
		_, _ = c.repository.MarkDiscrepancy(ctx, saga.ID, saga.Revision, now, "PROTECTION_GAP_ARM_DEADLINE")
		return Saga{}, ErrProtectionGap
	}
	if saga.State != StatePlanned {
		return Saga{}, fmt.Errorf("%w: registration requires PLANNED", ErrInvalidTransition)
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
	brokerCtx, cancel, err := boundedBrokerContext(ctx, durableFillAt.Add(RegistrationActiveDeadline).Sub(c.now()))
	if err != nil {
		at := c.now()
		_ = c.repository.markAttempt(ctx, attemptID, MutationDispatched, MutationInDoubt, at, "")
		_, _ = c.repository.MarkMutationUnknown(ctx, saga.ID, saga.Revision, at, attemptID)
		return Saga{}, ErrProtectionGap
	}
	defer cancel()
	broker, callErr := c.gateway.Create(brokerCtx, body)
	settledAt := c.now()
	if callErr != nil || brokerCtx.Err() != nil || !exactBrokerBodyResult(c.scope, body, broker) {
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
	if EvaluateArm(durableFillAt, now, settledAt) != ArmActive {
		c.entryOpen.Store(false)
		_, _ = c.repository.MarkDiscrepancy(ctx, saga.ID, saga.Revision, settledAt, "PROTECTION_GAP_ACTIVE_DEADLINE")
		return Saga{}, ErrProtectionGap
	}
	if err := c.confirmActive(ctx, saga); err != nil {
		return Saga{}, err
	}
	return saga, nil
}

func (c *Controller) Replace(ctx context.Context, sagaID, attemptID string, trigger, quantity int64, expireDate string) (Saga, error) {
	c.opMu.Lock()
	defer c.opMu.Unlock()
	c.closeEntry(sagaID)
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
	brokerCtx, cancel, err := boundedBrokerContext(ctx, brokerRecoveryDeadline)
	if err != nil {
		return Saga{}, ErrMutationInDoubt
	}
	defer cancel()
	broker, callErr := c.gateway.Replace(brokerCtx, saga.BrokerID, body)
	settledAt := c.now()
	if callErr != nil || brokerCtx.Err() != nil || !exactBrokerBodyResult(c.scope, body, broker) {
		c.entryOpen.Store(false)
		_ = c.repository.markAttempt(ctx, attemptID, MutationDispatched, MutationInDoubt, settledAt, broker.ID)
		_, _ = c.repository.MarkMutationUnknown(ctx, saga.ID, saga.Revision, settledAt, attemptID)
		return Saga{}, fmt.Errorf("%w: replace: %v", ErrMutationInDoubt, callErr)
	}
	if err := c.repository.markAttempt(ctx, attemptID, MutationDispatched, MutationAcknowledged, settledAt, broker.ID); err != nil {
		c.entryOpen.Store(false)
		return Saga{}, err
	}
	active, err := c.repository.MarkReplaceActive(ctx, saga.ID, saga.Revision, settledAt, attemptID, broker.ID)
	if err != nil {
		return Saga{}, err
	}
	if err := c.confirmActive(ctx, active); err != nil {
		return Saga{}, err
	}
	return active, nil
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
	c.opMu.Lock()
	defer c.opMu.Unlock()
	c.closeEntry("")
	local, err := c.repository.List(ctx, c.scope)
	if err != nil {
		c.entryOpen.Store(false)
		return nil, err
	}
	brokerCtx, cancel, deadlineErr := boundedBrokerContext(ctx, brokerRecoveryDeadline)
	if deadlineErr != nil {
		return nil, ErrMutationInDoubt
	}
	defer cancel()
	broker, err := c.gateway.List(brokerCtx, c.scope)
	if err != nil {
		c.entryOpen.Store(false)
		return nil, err
	}
	if brokerCtx.Err() != nil {
		return nil, ErrMutationInDoubt
	}
	filtered := broker
	matchedBySaga := make(map[string]BrokerProtection)
	for _, saga := range local {
		if saga.State != StateActive {
			continue
		}
		expected, expectationErr := c.brokerExpectation(ctx, saga)
		if expectationErr != nil {
			_, _ = c.repository.MarkDiscrepancy(ctx, saga.ID, saga.Revision, c.now(), "RECONCILE_DURABLE_LINEAGE_INVALID")
			return nil, ErrMutationInDoubt
		}
		matched, next, selectErr := selectExpectedBroker(c.scope, saga.ClientOrderID, expected, filtered)
		if selectErr != nil || matched.Terminal || matched.Triggered {
			_, _ = c.repository.MarkDiscrepancy(ctx, saga.ID, saga.Revision, c.now(), "RECONCILE_IDENTITY_MISMATCH")
			return nil, ErrMutationInDoubt
		}
		matchedBySaga[saga.ID] = matched
		filtered = next
	}
	discrepancies, err := Compare(c.scope, local, filtered)
	if err != nil {
		c.entryOpen.Store(false)
		return nil, err
	}
	if len(discrepancies) == 0 {
		for _, saga := range local {
			if saga.State != StateActive {
				continue
			}
			matched, ok := matchedBySaga[saga.ID]
			if !ok {
				_, _ = c.repository.MarkDiscrepancy(ctx, saga.ID, saga.Revision, c.now(), "RECONCILE_IDENTITY_MISMATCH")
				return nil, ErrMutationInDoubt
			}
			c.confirmed[saga.ID] = activeConfirmation{revision: saga.Revision, generation: saga.Generation, brokerID: matched.ID}
		}
		if err := c.refreshEntryLatch(ctx); err != nil {
			return nil, err
		}
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
	c.opMu.Lock()
	defer c.opMu.Unlock()
	c.closeEntry(sagaID)
	saga, err := c.repository.Get(ctx, sagaID)
	if err != nil {
		return Saga{}, err
	}
	brokerCtx, cancel, deadlineErr := boundedBrokerContext(ctx, brokerRecoveryDeadline)
	if deadlineErr != nil {
		return Saga{}, ErrMutationInDoubt
	}
	defer cancel()
	broker, err := c.gateway.List(brokerCtx, c.scope)
	if err != nil {
		c.entryOpen.Store(false)
		return Saga{}, fmt.Errorf("%w: recovery list: %v", ErrMutationInDoubt, err)
	}
	if brokerCtx.Err() != nil {
		return Saga{}, ErrMutationInDoubt
	}
	expected, err := c.brokerExpectation(ctx, saga)
	if err != nil {
		return Saga{}, fmt.Errorf("%w: durable lineage: %v", ErrMutationInDoubt, err)
	}
	matched, filtered, selectErr := selectExpectedBroker(c.scope, saga.ClientOrderID, expected, broker)
	now := c.now()
	if selectErr != nil {
		reason := "RECOVERY_IDENTITY_MISMATCH"
		unknownDispatch := expected.pending != nil && (expected.pending.State == MutationDispatched || expected.pending.State == MutationInDoubt)
		if errors.Is(selectErr, ErrProtectionGone) && !unknownDispatch {
			reason = string(DiscrepancyMissing)
		}
		_, markErr := c.repository.MarkDiscrepancy(ctx, saga.ID, saga.Revision, now, reason)
		if markErr != nil {
			return Saga{}, fmt.Errorf("%w: missing protection and discrepancy write failed: %v", ErrMutationInDoubt, markErr)
		}
		if errors.Is(selectErr, ErrProtectionGone) && !unknownDispatch {
			return Saga{}, ErrProtectionGone
		}
		return Saga{}, ErrMutationInDoubt
	}
	// Recovery observed the whole bounded inventory. Any unrelated non-terminal
	// row is an orphan and therefore prevents the entry latch from reopening.
	for _, candidate := range filtered {
		if candidate.ID != matched.ID && !candidate.Terminal {
			_, _ = c.repository.MarkDiscrepancy(ctx, saga.ID, saga.Revision, now, "RECOVERY_ORPHAN")
			return Saga{}, ErrMutationInDoubt
		}
	}
	if matched.Triggered {
		c.entryOpen.Store(false)
		return c.repository.RecoverTriggered(ctx, saga.ID, saga.Revision, now, matched.ID)
	}
	if matched.Terminal {
		_, markErr := c.repository.MarkDiscrepancy(ctx, saga.ID, saga.Revision, now, "TERMINAL_WITHOUT_TRIGGER")
		if markErr != nil {
			return Saga{}, fmt.Errorf("%w: terminal protection and discrepancy write failed: %v", ErrMutationInDoubt, markErr)
		}
		return Saga{}, ErrProtectionGone
	}
	if saga.State == StateActive {
		if err := c.confirmActive(ctx, saga); err != nil {
			return Saga{}, err
		}
		return saga, nil
	}
	if expected.pending != nil {
		switch expected.pending.State {
		case MutationDispatched, MutationInDoubt:
			if err := c.repository.markAttempt(ctx, expected.pending.ID, expected.pending.State, MutationAcknowledged, now, matched.ID); err != nil {
				return Saga{}, fmt.Errorf("%w: persist recovered broker identity: %v", ErrMutationInDoubt, err)
			}
		case MutationAcknowledged:
			if expected.pending.ResultBrokerID != matched.ID {
				return Saga{}, ErrMutationInDoubt
			}
		default:
			return Saga{}, ErrMutationInDoubt
		}
	}
	active, err := c.repository.RecoverActive(ctx, saga.ID, saga.Revision, now, matched.ID, matched.Trigger, matched.Quantity)
	if err != nil {
		return Saga{}, err
	}
	if err := c.confirmActive(ctx, active); err != nil {
		return Saga{}, err
	}
	return active, nil
}

func (c *Controller) AuthorizeFlatten(ctx context.Context, sagaID, attemptID string, _ time.Time, required int64) (FlattenAuthorization, error) {
	c.opMu.Lock()
	defer c.opMu.Unlock()
	c.closeEntry(sagaID)
	operationStart := c.now()
	saga, err := c.repository.Get(ctx, sagaID)
	if err != nil {
		return FlattenAuthorization{}, err
	}
	target, err := c.brokerTargetForSaga(ctx, saga)
	if err != nil {
		return FlattenAuthorization{}, fmt.Errorf("%w: cancel target lineage: %v", ErrMutationInDoubt, err)
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
	brokerCtx, cancelBroker, deadlineErr := boundedBrokerContext(ctx, operationStart.Add(2*time.Second).Sub(c.now()))
	if deadlineErr != nil {
		at := c.now()
		_ = c.repository.markAttempt(ctx, attemptID, MutationDispatched, MutationInDoubt, at, saga.BrokerID)
		_, _ = c.repository.MarkMutationUnknown(ctx, saga.ID, saga.Revision, at, attemptID)
		return FlattenAuthorization{}, ErrMutationInDoubt
	}
	defer cancelBroker()
	cancel, callErr := c.gateway.Cancel(brokerCtx, target)
	if callErr != nil || brokerCtx.Err() != nil || !cancel.Terminal || cancel.Triggered || cancel.Scope != c.scope || cancel.BrokerID != saga.BrokerID || cancel.ClientOrderID != saga.ClientOrderID {
		c.entryOpen.Store(false)
		at := c.now()
		_ = c.repository.markAttempt(ctx, attemptID, MutationDispatched, MutationInDoubt, at, saga.BrokerID)
		_, _ = c.repository.MarkMutationUnknown(ctx, saga.ID, saga.Revision, at, attemptID)
		return FlattenAuthorization{}, ErrMutationInDoubt
	}
	sellable, err := c.gateway.Sellable(brokerCtx, c.scope, saga.BrokerID)
	decisionAt := c.now()
	deadline := operationStart.Add(2 * time.Second)
	decision, authorization := decideFlatten(operationStart, deadline, decisionAt,
		FlattenScope{Scope: c.scope, BrokerID: saga.BrokerID}, cancel, sellable, required, c.now)
	if err != nil || brokerCtx.Err() != nil || decision != FlattenAllowed {
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

func (c *Controller) closeEntry(sagaID string) {
	c.entryOpen.Store(false)
	if sagaID == "" {
		clear(c.confirmed)
		return
	}
	delete(c.confirmed, sagaID)
}

func (c *Controller) confirmActive(ctx context.Context, saga Saga) error {
	if saga.State != StateActive || saga.BrokerID == "" {
		c.entryOpen.Store(false)
		return ErrMutationInDoubt
	}
	c.confirmed[saga.ID] = activeConfirmation{revision: saga.Revision, generation: saga.Generation, brokerID: saga.BrokerID}
	return c.refreshEntryLatch(ctx)
}

func (c *Controller) refreshEntryLatch(ctx context.Context) error {
	rows, err := c.repository.List(ctx, c.scope)
	if err != nil {
		c.entryOpen.Store(false)
		return fmt.Errorf("protection: refresh entry latch: %w", err)
	}
	for _, saga := range rows {
		confirmed, ok := c.confirmed[saga.ID]
		if !ok || saga.State != StateActive || confirmed.revision != saga.Revision || confirmed.generation != saga.Generation || confirmed.brokerID != saga.BrokerID {
			c.entryOpen.Store(false)
			return nil
		}
	}
	c.entryOpen.Store(true)
	return nil
}

type brokerExpectation struct {
	id      string
	body    ConditionalBody
	retired []RetiredBrokerTarget
	pending *MutationAttempt
}

func (c *Controller) brokerTargetForSaga(ctx context.Context, saga Saga) (BrokerTarget, error) {
	expected, err := c.brokerExpectation(ctx, saga)
	if err != nil || expected.id == "" {
		return BrokerTarget{}, ErrInvalidSaga
	}
	return BrokerTarget{Scope: c.scope, BrokerID: expected.id, ClientOrderID: saga.ClientOrderID,
		Trigger: expected.body.Trigger, Quantity: expected.body.Quantity, ExpireDate: expected.body.ExpireDate,
		Retired: expected.retired}, nil
}

// brokerExpectation reconstructs the current identity and its exact retired
// predecessors exclusively from durable mutation attempts. A DISPATCHED or
// IN_DOUBT request supplies an exact body but no trusted result ID; an
// ACKNOWLEDGED request supplies the durable result ID even when the subsequent
// saga commit was interrupted.
func (c *Controller) brokerExpectation(ctx context.Context, saga Saga) (brokerExpectation, error) {
	attempts, err := c.repository.Attempts(ctx, saga.ID)
	if err != nil {
		return brokerExpectation{}, err
	}
	type acknowledged struct {
		target string
		body   ConditionalBody
		kind   MutationKind
		gen    int64
	}
	byResult := make(map[string]acknowledged)
	var sagaAttempt *MutationAttempt
	for i := range attempts {
		a := &attempts[i]
		if a.Kind != MutationCreate && a.Kind != MutationReplace {
			continue
		}
		body, bodyErr := attemptConditionalBody(*a, saga)
		if bodyErr != nil {
			return brokerExpectation{}, bodyErr
		}
		if a.State == MutationAcknowledged {
			if a.ResultBrokerID == "" {
				return brokerExpectation{}, ErrInvalidSaga
			}
			if _, duplicate := byResult[a.ResultBrokerID]; duplicate {
				return brokerExpectation{}, ErrDuplicateBrokerID
			}
			byResult[a.ResultBrokerID] = acknowledged{target: a.TargetBrokerID, body: body, kind: a.Kind, gen: a.Generation}
		}
		if a.ID == saga.AttemptID {
			copy := *a
			sagaAttempt = &copy
		}
	}

	recoveringAttempt := sagaAttempt != nil && ((sagaAttempt.Kind == MutationCreate && saga.BrokerID == "" &&
		(saga.State == StateRegistering || saga.State == StateInDoubt || saga.State == StateReconcile)) ||
		(sagaAttempt.Kind == MutationReplace && saga.PendingTrigger > 0 &&
			(saga.State == StateReplacing || saga.State == StateInDoubt || saga.State == StateReconcile)))
	expected := brokerExpectation{id: saga.BrokerID}
	chainID := saga.BrokerID
	chainGeneration := saga.Generation
	if recoveringAttempt {
		if sagaAttempt.State == MutationPlanned || (sagaAttempt.State != MutationDispatched && sagaAttempt.State != MutationInDoubt && sagaAttempt.State != MutationAcknowledged) {
			return brokerExpectation{}, ErrInvalidSaga
		}
		body, bodyErr := attemptConditionalBody(*sagaAttempt, saga)
		if bodyErr != nil {
			return brokerExpectation{}, bodyErr
		}
		expected.pending = sagaAttempt
		expected.body = body
		if sagaAttempt.State == MutationAcknowledged {
			if sagaAttempt.ResultBrokerID == "" {
				return brokerExpectation{}, ErrInvalidSaga
			}
			expected.id = sagaAttempt.ResultBrokerID
		} else {
			expected.id = ""
		}
		if sagaAttempt.Kind == MutationCreate && (sagaAttempt.Generation != 1 || saga.Generation != 1 || saga.BrokerID != "" ||
			body.Trigger != saga.Trigger || body.Quantity != saga.Quantity) {
			return brokerExpectation{}, ErrInvalidSaga
		}
		if sagaAttempt.Kind == MutationCreate {
			if expected.id == "" {
				chainID, chainGeneration = "", 0
			} else {
				chainID, chainGeneration = expected.id, 1
			}
		}
		if sagaAttempt.Kind == MutationReplace && (sagaAttempt.Generation != saga.Generation || sagaAttempt.TargetBrokerID != saga.BrokerID ||
			saga.PreviousBrokerID != sagaAttempt.TargetBrokerID || body.Trigger != saga.PendingTrigger || body.Quantity != saga.PendingQuantity) {
			return brokerExpectation{}, ErrInvalidSaga
		}
		if sagaAttempt.Kind == MutationReplace {
			if expected.id == "" {
				chainID, chainGeneration = sagaAttempt.TargetBrokerID, saga.Generation
			} else {
				chainID, chainGeneration = expected.id, saga.Generation+1
			}
		}
	} else {
		current, ok := byResult[saga.BrokerID]
		if !ok {
			return brokerExpectation{}, ErrInvalidSaga
		}
		expected.body = current.body
		if current.body.Trigger != saga.Trigger || current.body.Quantity != saga.Quantity {
			return brokerExpectation{}, ErrInvalidSaga
		}
	}

	seen := make(map[string]bool)
	for chainID != "" {
		if seen[chainID] {
			return brokerExpectation{}, ErrDuplicateBrokerID
		}
		seen[chainID] = true
		current, ok := byResult[chainID]
		if !ok {
			return brokerExpectation{}, ErrInvalidSaga
		}
		if chainGeneration == 1 {
			if current.kind != MutationCreate || current.gen != 1 || current.target != "" {
				return brokerExpectation{}, ErrInvalidSaga
			}
		} else if current.kind != MutationReplace || current.gen != chainGeneration-1 || current.target == "" {
			return brokerExpectation{}, ErrInvalidSaga
		}
		retireCurrent := expected.id == "" || chainID != expected.id
		if retireCurrent {
			expected.retired = append(expected.retired, RetiredBrokerTarget{BrokerID: chainID,
				ClientOrderID: saga.ClientOrderID, Trigger: current.body.Trigger, Quantity: current.body.Quantity,
				ExpireDate: current.body.ExpireDate})
		}
		if expected.id != "" && chainID == expected.id && saga.Generation > 1 && saga.PreviousBrokerID != current.target {
			return brokerExpectation{}, ErrInvalidSaga
		}
		chainID = current.target
		chainGeneration--
	}
	if chainGeneration != 0 || len(seen) != len(byResult) {
		return brokerExpectation{}, ErrInvalidSaga
	}
	return expected, nil
}

func attemptConditionalBody(attempt MutationAttempt, saga Saga) (ConditionalBody, error) {
	if attempt.SerializerVersion != SerializerVersion {
		return ConditionalBody{}, ErrInvalidBody
	}
	decoder := json.NewDecoder(bytes.NewBufferString(attempt.CanonicalBody))
	decoder.DisallowUnknownFields()
	var body ConditionalBody
	if err := decoder.Decode(&body); err != nil || decoder.More() {
		return ConditionalBody{}, ErrInvalidBody
	}
	canonical, err := body.CanonicalJSON()
	if err != nil || string(canonical) != attempt.CanonicalBody || body.ClientOrderID != saga.ClientOrderID ||
		body.AccountRef != saga.AccountRef || body.Market != string(saga.Market) || body.Symbol != saga.Symbol {
		return ConditionalBody{}, ErrInvalidBody
	}
	return body, nil
}

func exactExpectedBroker(scope Scope, clientID string, expected brokerExpectation, broker BrokerProtection) bool {
	idMatches := expected.id == "" || broker.ID == expected.id
	return broker.ID != "" && idMatches && broker.Scope.equal(scope) && broker.ClientOrderID == clientID &&
		broker.Trigger == expected.body.Trigger && broker.Quantity == expected.body.Quantity &&
		broker.ExpireDate == expected.body.ExpireDate && broker.OrderSide == "SELL" && broker.OrderType == "MARKET" &&
		broker.ConditionType == "STOP"
}

func exactRetiredBroker(scope Scope, expected RetiredBrokerTarget, broker BrokerProtection) bool {
	return broker.Terminal && !broker.Triggered && broker.Scope.equal(scope) && broker.ID == expected.BrokerID &&
		broker.ClientOrderID == expected.ClientOrderID && broker.Trigger == expected.Trigger && broker.Quantity == expected.Quantity &&
		broker.ExpireDate == expected.ExpireDate && broker.OrderSide == "SELL" && broker.OrderType == "MARKET" && broker.ConditionType == "STOP"
}

func selectExpectedBroker(scope Scope, clientID string, expected brokerExpectation, broker []BrokerProtection) (BrokerProtection, []BrokerProtection, error) {
	retired := make(map[string]RetiredBrokerTarget, len(expected.retired))
	for _, item := range expected.retired {
		retired[item.BrokerID] = item
	}
	filtered := make([]BrokerProtection, 0, len(broker))
	var matched *BrokerProtection
	for i := range broker {
		candidate := broker[i]
		if old, ok := retired[candidate.ID]; ok {
			if !exactRetiredBroker(scope, old, candidate) {
				return BrokerProtection{}, nil, ErrMutationInDoubt
			}
			continue
		}
		filtered = append(filtered, candidate)
		isCurrent := (expected.id != "" && candidate.ID == expected.id) ||
			(expected.id == "" && candidate.ClientOrderID == clientID && candidate.Trigger == expected.body.Trigger && candidate.Quantity == expected.body.Quantity)
		if isCurrent {
			if matched != nil || !exactExpectedBroker(scope, clientID, expected, candidate) {
				return BrokerProtection{}, nil, ErrDuplicateBrokerID
			}
			copy := candidate
			matched = &copy
			continue
		}
		if candidate.ClientOrderID == clientID {
			return BrokerProtection{}, nil, ErrMutationInDoubt
		}
	}
	if matched == nil {
		return BrokerProtection{}, filtered, ErrProtectionGone
	}
	return *matched, filtered, nil
}

func boundedBrokerContext(parent context.Context, duration time.Duration) (context.Context, context.CancelFunc, error) {
	if parent == nil || duration <= 0 {
		return nil, nil, ErrMutationInDoubt
	}
	ctx, cancel := context.WithTimeout(parent, duration)
	return ctx, cancel, nil
}

func exactBrokerBodyResult(scope Scope, body ConditionalBody, broker BrokerProtection) bool {
	return !broker.Terminal && !broker.Triggered && broker.ID != "" && broker.Scope.equal(scope) &&
		broker.ClientOrderID == body.ClientOrderID && broker.Trigger == body.Trigger && broker.Quantity == body.Quantity &&
		broker.ExpireDate == body.ExpireDate && broker.OrderSide == "SELL" && broker.OrderType == "MARKET" && broker.ConditionType == "STOP"
}
