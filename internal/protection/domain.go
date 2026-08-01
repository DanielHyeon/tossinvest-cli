// Package protection contains the dormant, broker-neutral domain contract for
// resident protection orders. It has no official client adapter and is not
// imported by the engine. In particular, this package cannot make a broker
// request or change execgw.ProfileProtection from UNWIRED.
package protection

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"
)

const SerializerVersion = 1

var (
	ErrInvalidSaga          = errors.New("protection: invalid saga")
	ErrInvalidTransition    = errors.New("protection: invalid saga transition")
	ErrWeakerProtection     = errors.New("protection: replacement would weaken protection")
	ErrOversell             = errors.New("protection: sell claims exceed the holding")
	ErrInvalidBody          = errors.New("protection: invalid conditional body")
	ErrConcurrentUpdate     = errors.New("protection: concurrent saga update")
	ErrMixedScope           = errors.New("protection: mixed reconciliation scope")
	ErrDuplicateBrokerID    = errors.New("protection: duplicate broker identity")
	ErrFlattenAuthorization = errors.New("protection: flatten authorization is absent, expired, mismatched, or spent")
)

type Market string

const (
	MarketKR Market = "KR"
	MarketUS Market = "US"
)

type Scope struct {
	AccountRef string
	Profile    string
	Market     Market
	Symbol     string
}

func (s Scope) Validate() error {
	if s.AccountRef == "" || s.AccountRef != strings.TrimSpace(s.AccountRef) || s.Profile == "" || s.Profile != strings.TrimSpace(s.Profile) || s.Symbol == "" || s.Symbol != strings.TrimSpace(s.Symbol) || (s.Market != MarketKR && s.Market != MarketUS) {
		return fmt.Errorf("%w: account/profile/market/symbol is invalid", ErrMixedScope)
	}
	return nil
}

func (s Scope) equal(other Scope) bool { return s == other }

type State string

const (
	StatePlanned     State = "PLANNED"
	StateRegistering State = "REGISTERING"
	StateActive      State = "ACTIVE"
	StateReplacing   State = "REPLACING"
	StateTriggered   State = "TRIGGERED"
	StateClosed      State = "CLOSED"
	StateReconcile   State = "RECONCILE"
	StateInDoubt     State = "IN_DOUBT"
)

type Saga struct {
	ID               string
	AccountRef       string
	Profile          string
	Market           Market
	Symbol           string
	Generation       int64
	Revision         int64
	State            State
	Trigger          int64
	Quantity         int64
	PendingTrigger   int64
	PendingQuantity  int64
	ClientOrderID    string
	AttemptID        string
	BrokerID         string
	PreviousBrokerID string
	ReconcileReason  string
	UpdatedAt        time.Time
}

type EventKind string

const (
	EventBeginRegistration  EventKind = "BEGIN_REGISTRATION"
	EventRegistrationActive EventKind = "REGISTRATION_ACTIVE"
	EventMutationUnknown    EventKind = "MUTATION_UNKNOWN"
	EventBeginReplace       EventKind = "BEGIN_REPLACE"
	EventReplaceActive      EventKind = "REPLACE_ACTIVE"
	EventTriggerObserved    EventKind = "TRIGGER_OBSERVED"
	EventClose              EventKind = "CLOSE"
	EventDiscrepancy        EventKind = "DISCREPANCY"
)

type Event struct {
	Kind      EventKind
	At        time.Time
	AttemptID string
	BrokerID  string
	Trigger   int64
	Quantity  int64
	Reason    string
}

func Transition(in Saga, event Event) (Saga, error) {
	if err := in.Validate(); err != nil {
		return in, err
	}
	if event.At.IsZero() || event.At.Before(in.UpdatedAt) {
		return in, fmt.Errorf("%w: event time is absent or moves backward", ErrInvalidTransition)
	}
	out := in
	out.UpdatedAt = event.At

	switch event.Kind {
	case EventBeginRegistration:
		if in.State != StatePlanned || strings.TrimSpace(event.AttemptID) == "" || event.BrokerID != "" {
			return in, invalidTransition(in.State, event.Kind)
		}
		out.State, out.AttemptID = StateRegistering, event.AttemptID
	case EventRegistrationActive:
		if in.State != StateRegistering || event.AttemptID != in.AttemptID || strings.TrimSpace(event.BrokerID) == "" {
			return in, invalidTransition(in.State, event.Kind)
		}
		out.State, out.BrokerID = StateActive, event.BrokerID
	case EventMutationUnknown:
		if (in.State != StateRegistering && in.State != StateReplacing) || event.AttemptID != in.AttemptID || event.BrokerID != "" {
			return in, invalidTransition(in.State, event.Kind)
		}
		out.State = StateInDoubt
		out.ReconcileReason = "MUTATION_RESULT_UNKNOWN"
	case EventBeginReplace:
		if in.State != StateActive || strings.TrimSpace(event.AttemptID) == "" || event.AttemptID == in.AttemptID || event.BrokerID != "" {
			return in, invalidTransition(in.State, event.Kind)
		}
		if event.Trigger < in.Trigger {
			return in, fmt.Errorf("%w: %d is below active trigger %d", ErrWeakerProtection, event.Trigger, in.Trigger)
		}
		quantity := event.Quantity
		if quantity == 0 {
			quantity = in.Quantity
		}
		if quantity < 1 {
			return in, fmt.Errorf("%w: replacement quantity must be positive", ErrInvalidTransition)
		}
		out.State = StateReplacing
		out.AttemptID = event.AttemptID
		out.PreviousBrokerID = in.BrokerID
		out.PendingTrigger = event.Trigger
		out.PendingQuantity = quantity
	case EventReplaceActive:
		if in.State != StateReplacing || event.AttemptID != in.AttemptID || strings.TrimSpace(event.BrokerID) == "" {
			return in, invalidTransition(in.State, event.Kind)
		}
		out.State = StateActive
		out.BrokerID = event.BrokerID
		out.Trigger = in.PendingTrigger
		out.Quantity = in.PendingQuantity
		out.PendingTrigger = 0
		out.PendingQuantity = 0
		out.Generation++
	case EventTriggerObserved:
		if in.State != StateActive || event.BrokerID != in.BrokerID || event.AttemptID != "" {
			return in, invalidTransition(in.State, event.Kind)
		}
		out.State = StateTriggered
	case EventClose:
		if (in.State != StateTriggered && in.State != StateReconcile) || event.BrokerID != in.BrokerID || event.AttemptID != "" {
			return in, invalidTransition(in.State, event.Kind)
		}
		out.State = StateClosed
	case EventDiscrepancy:
		if in.State == StateClosed || strings.TrimSpace(event.Reason) == "" {
			return in, invalidTransition(in.State, event.Kind)
		}
		out.State = StateReconcile
		out.ReconcileReason = event.Reason
	default:
		return in, invalidTransition(in.State, event.Kind)
	}
	if err := out.Validate(); err != nil {
		return in, fmt.Errorf("%w: transition output: %v", ErrInvalidTransition, err)
	}
	return out, nil
}

func invalidTransition(state State, event EventKind) error {
	return fmt.Errorf("%w: %s from %s", ErrInvalidTransition, event, state)
}

func (s Saga) Validate() error {
	switch {
	case s.ID == "" || s.ID != strings.TrimSpace(s.ID), s.AccountRef == "" || s.AccountRef != strings.TrimSpace(s.AccountRef), s.Profile == "" || s.Profile != strings.TrimSpace(s.Profile):
		return fmt.Errorf("%w: identity is incomplete", ErrInvalidSaga)
	case (s.Market != MarketKR && s.Market != MarketUS), s.Symbol == "" || s.Symbol != strings.TrimSpace(s.Symbol):
		return fmt.Errorf("%w: instrument is incomplete", ErrInvalidSaga)
	case s.Generation < 1, s.Revision < 1, s.Trigger < 1, s.Quantity < 1:
		return fmt.Errorf("%w: generation/trigger/quantity must be positive", ErrInvalidSaga)
	case s.ClientOrderID == "" || s.ClientOrderID != strings.TrimSpace(s.ClientOrderID):
		return fmt.Errorf("%w: client order identity is absent", ErrInvalidSaga)
	case !validState(s.State):
		return fmt.Errorf("%w: unknown state %q", ErrInvalidSaga, s.State)
	case s.UpdatedAt.IsZero():
		return fmt.Errorf("%w: updated time is absent", ErrInvalidSaga)
	}
	if err := s.validateStateFields(); err != nil {
		return err
	}
	return nil
}

func (s Saga) validateStateFields() error {
	nonempty := func(value string) bool { return value != "" && value == strings.TrimSpace(value) }
	noPending := s.PendingTrigger == 0 && s.PendingQuantity == 0
	switch s.State {
	case StatePlanned:
		if s.AttemptID != "" || s.BrokerID != "" || s.PreviousBrokerID != "" || s.ReconcileReason != "" || !noPending {
			return fmt.Errorf("%w: PLANNED carries mutation state", ErrInvalidSaga)
		}
	case StateRegistering:
		if !nonempty(s.AttemptID) || s.BrokerID != "" || s.PreviousBrokerID != "" || s.ReconcileReason != "" || !noPending {
			return fmt.Errorf("%w: REGISTERING fields are inconsistent", ErrInvalidSaga)
		}
	case StateActive:
		if !nonempty(s.AttemptID) || !nonempty(s.BrokerID) || s.ReconcileReason != "" || !noPending {
			return fmt.Errorf("%w: ACTIVE fields are inconsistent", ErrInvalidSaga)
		}
	case StateReplacing:
		if !nonempty(s.AttemptID) || !nonempty(s.BrokerID) || !nonempty(s.PreviousBrokerID) || s.PreviousBrokerID != s.BrokerID || s.ReconcileReason != "" || s.PendingTrigger < s.Trigger || s.PendingQuantity < 1 {
			return fmt.Errorf("%w: REPLACING fields are inconsistent", ErrInvalidSaga)
		}
	case StateTriggered:
		if !nonempty(s.BrokerID) || s.ReconcileReason != "" || !noPending {
			return fmt.Errorf("%w: TRIGGERED fields are inconsistent", ErrInvalidSaga)
		}
	case StateClosed:
		if !noPending {
			return fmt.Errorf("%w: CLOSED carries pending replacement", ErrInvalidSaga)
		}
	case StateReconcile:
		if !nonempty(s.ReconcileReason) {
			return fmt.Errorf("%w: RECONCILE reason is absent", ErrInvalidSaga)
		}
	case StateInDoubt:
		if !nonempty(s.AttemptID) || !nonempty(s.ReconcileReason) {
			return fmt.Errorf("%w: IN_DOUBT mutation identity/reason is absent", ErrInvalidSaga)
		}
	}
	return nil
}

func validState(state State) bool {
	switch state {
	case StatePlanned, StateRegistering, StateActive, StateReplacing, StateTriggered, StateClosed, StateReconcile, StateInDoubt:
		return true
	default:
		return false
	}
}

type SellClaims struct {
	Protection       int64
	OpenSell         int64
	LocalReservation int64
}

func ValidateSellClaims(holding int64, claims SellClaims) error {
	if holding < 0 || claims.Protection < 0 || claims.OpenSell < 0 || claims.LocalReservation < 0 {
		return fmt.Errorf("%w: negative quantity", ErrOversell)
	}
	remaining := holding
	for _, claim := range []int64{claims.Protection, claims.OpenSell, claims.LocalReservation} {
		if claim > remaining {
			return fmt.Errorf("%w: holding=%d protection=%d open_sell=%d local=%d", ErrOversell, holding, claims.Protection, claims.OpenSell, claims.LocalReservation)
		}
		remaining -= claim
	}
	return nil
}

type ArmStatus string

const (
	ArmActive        ArmStatus = "ACTIVE"
	ArmProtectionGap ArmStatus = "PROTECTION_GAP"
)

const (
	RegistrationArmDeadline    = time.Second
	RegistrationActiveDeadline = 2 * time.Second
)

func EvaluateArm(fillAt, attemptAt, activeAt time.Time) ArmStatus {
	if fillAt.IsZero() || attemptAt.IsZero() || activeAt.IsZero() || attemptAt.Before(fillAt) || activeAt.Before(attemptAt) {
		return ArmProtectionGap
	}
	if attemptAt.Sub(fillAt) > RegistrationArmDeadline || activeAt.Sub(fillAt) > RegistrationActiveDeadline {
		return ArmProtectionGap
	}
	return ArmActive
}

// ConditionalBody is a pure wire-schema candidate. CanonicalJSON validates and
// serializes it; no production adapter in this change can submit the bytes.
type ConditionalBody struct {
	SerializerVersion int    `json:"serializer_version"`
	ClientOrderID     string `json:"client_order_id"`
	AccountRef        string `json:"account_ref"`
	Market            string `json:"market"`
	Symbol            string `json:"symbol"`
	Side              string `json:"side"`
	ConditionalType   string `json:"conditional_type"`
	OrderType         string `json:"order_type"`
	TriggerSource     string `json:"trigger_source"`
	Trigger           int64  `json:"trigger"`
	Quantity          int64  `json:"quantity"`
}

func (b ConditionalBody) CanonicalJSON() ([]byte, error) {
	if b.SerializerVersion != SerializerVersion || strings.TrimSpace(b.ClientOrderID) == "" || strings.TrimSpace(b.AccountRef) == "" || strings.TrimSpace(b.Symbol) == "" || (b.Market != "KR" && b.Market != "US") || b.Side != "SELL" || b.ConditionalType != "SINGLE" || b.OrderType != "MARKET" || b.TriggerSource != "LAST_TRADE" || b.Trigger < 1 || b.Quantity < 1 {
		return nil, ErrInvalidBody
	}
	return json.Marshal(b)
}

// Gateway is intentionally only a contract. The dormant change has no
// non-test implementation, so it cannot reach an official or unofficial API.
type Gateway interface {
	Create(context.Context, ConditionalBody) (BrokerProtection, error)
	Replace(context.Context, string, ConditionalBody) (BrokerProtection, error)
	Cancel(context.Context, string) (CancelObservation, error)
	Get(context.Context, string) (BrokerProtection, error)
	List(context.Context, Scope) ([]BrokerProtection, error)
}

type BrokerProtection struct {
	Scope    Scope
	ID       string
	Quantity int64
	Trigger  int64
	Terminal bool
}

type CancelObservation struct {
	Scope     Scope
	BrokerID  string
	Terminal  bool
	Triggered bool
	At        time.Time
}

type SellableObservation struct {
	Scope    Scope
	BrokerID string
	Quantity int64
	At       time.Time
}

type FlattenDecision string

const (
	FlattenAllowed FlattenDecision = "ALLOWED"
	FlattenInDoubt FlattenDecision = "IN_DOUBT"
)

type FlattenScope struct {
	Scope    Scope
	BrokerID string
}

type flattenPermit struct {
	consumed  atomic.Bool
	issuedAt  time.Time
	expiresAt time.Time
	target    FlattenScope
	required  int64
	clock     func() time.Time
}

// FlattenAuthorization is opaque and copy-safe: every copy shares the same
// one-shot consumption state. It carries no broker mutation method.
type FlattenAuthorization struct{ permit *flattenPermit }

func (a FlattenAuthorization) Consume(target FlattenScope, required int64) error {
	permit := a.permit
	if permit == nil || permit.clock == nil || target != permit.target || required != permit.required {
		return ErrFlattenAuthorization
	}
	decisionAt := permit.clock()
	if decisionAt.IsZero() || decisionAt.Before(permit.issuedAt) || decisionAt.After(permit.expiresAt) {
		return ErrFlattenAuthorization
	}
	if !permit.consumed.CompareAndSwap(false, true) {
		return ErrFlattenAuthorization
	}
	return nil
}

func DecideFlatten(start, deadline time.Time, target FlattenScope, cancel CancelObservation, sellable SellableObservation, required int64) (FlattenDecision, FlattenAuthorization) {
	return decideFlatten(start, deadline, time.Now(), target, cancel, sellable, required, time.Now)
}

func decideFlatten(start, deadline, decisionAt time.Time, target FlattenScope, cancel CancelObservation, sellable SellableObservation, required int64, clock func() time.Time) (FlattenDecision, FlattenAuthorization) {
	if start.IsZero() || deadline.IsZero() || decisionAt.IsZero() || start.After(deadline) || deadline.Sub(start) > 2*time.Second || decisionAt.Before(start) || decisionAt.After(deadline) || target.Scope.Validate() != nil || target.BrokerID == "" || required < 1 || !cancel.Terminal || cancel.Triggered || cancel.At.IsZero() || sellable.At.IsZero() || cancel.At.Before(start) || sellable.At.Before(cancel.At) || decisionAt.Before(sellable.At) || cancel.At.After(deadline) || sellable.At.After(deadline) || !cancel.Scope.equal(target.Scope) || !sellable.Scope.equal(target.Scope) || cancel.BrokerID != target.BrokerID || sellable.BrokerID != target.BrokerID || sellable.Quantity < required {
		return FlattenInDoubt, FlattenAuthorization{}
	}
	return FlattenAllowed, FlattenAuthorization{permit: &flattenPermit{issuedAt: decisionAt, expiresAt: deadline, target: target, required: required, clock: clock}}
}
