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
	"time"
)

const SerializerVersion = 1

var (
	ErrInvalidSaga       = errors.New("protection: invalid saga")
	ErrInvalidTransition = errors.New("protection: invalid saga transition")
	ErrWeakerProtection  = errors.New("protection: replacement would weaken protection")
	ErrOversell          = errors.New("protection: sell claims exceed the holding")
	ErrInvalidBody       = errors.New("protection: invalid conditional body")
	ErrConcurrentUpdate  = errors.New("protection: concurrent saga update")
)

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
	Market           string
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
		if in.State != StatePlanned || strings.TrimSpace(event.AttemptID) == "" {
			return in, invalidTransition(in.State, event.Kind)
		}
		out.State, out.AttemptID = StateRegistering, event.AttemptID
	case EventRegistrationActive:
		if in.State != StateRegistering || strings.TrimSpace(event.BrokerID) == "" {
			return in, invalidTransition(in.State, event.Kind)
		}
		out.State, out.BrokerID = StateActive, event.BrokerID
	case EventMutationUnknown:
		if in.State != StateRegistering && in.State != StateReplacing {
			return in, invalidTransition(in.State, event.Kind)
		}
		out.State = StateInDoubt
		out.ReconcileReason = "MUTATION_RESULT_UNKNOWN"
	case EventBeginReplace:
		if in.State != StateActive || strings.TrimSpace(event.AttemptID) == "" {
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
		if in.State != StateReplacing || strings.TrimSpace(event.BrokerID) == "" {
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
		if in.State != StateActive && in.State != StateReplacing {
			return in, invalidTransition(in.State, event.Kind)
		}
		out.State = StateTriggered
	case EventClose:
		if in.State != StateTriggered && in.State != StateReconcile {
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
	return out, nil
}

func invalidTransition(state State, event EventKind) error {
	return fmt.Errorf("%w: %s from %s", ErrInvalidTransition, event, state)
}

func (s Saga) Validate() error {
	switch {
	case strings.TrimSpace(s.ID) == "", strings.TrimSpace(s.AccountRef) == "", strings.TrimSpace(s.Profile) == "":
		return fmt.Errorf("%w: identity is incomplete", ErrInvalidSaga)
	case strings.TrimSpace(s.Market) == "", strings.TrimSpace(s.Symbol) == "":
		return fmt.Errorf("%w: instrument is incomplete", ErrInvalidSaga)
	case s.Generation < 1, s.Trigger < 1, s.Quantity < 1:
		return fmt.Errorf("%w: generation/trigger/quantity must be positive", ErrInvalidSaga)
	case strings.TrimSpace(s.ClientOrderID) == "":
		return fmt.Errorf("%w: client order identity is absent", ErrInvalidSaga)
	case !validState(s.State):
		return fmt.Errorf("%w: unknown state %q", ErrInvalidSaga, s.State)
	case s.UpdatedAt.IsZero():
		return fmt.Errorf("%w: updated time is absent", ErrInvalidSaga)
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
	if claims.Protection+claims.OpenSell+claims.LocalReservation > holding {
		return fmt.Errorf("%w: holding=%d protection=%d open_sell=%d local=%d", ErrOversell, holding, claims.Protection, claims.OpenSell, claims.LocalReservation)
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
	List(context.Context, string) ([]BrokerProtection, error)
}

type BrokerProtection struct {
	ID       string
	Symbol   string
	Quantity int64
	Trigger  int64
	Terminal bool
}

type CancelObservation struct {
	Terminal  bool
	Triggered bool
	At        time.Time
}

type SellableObservation struct {
	Quantity int64
	At       time.Time
}

type FlattenDecision string

const (
	FlattenAllowed FlattenDecision = "ALLOWED"
	FlattenInDoubt FlattenDecision = "IN_DOUBT"
)

func DecideFlatten(deadline time.Time, cancel CancelObservation, sellable SellableObservation, required int64) FlattenDecision {
	if deadline.IsZero() || required < 1 || !cancel.Terminal || cancel.Triggered || cancel.At.IsZero() || sellable.At.IsZero() || cancel.At.After(deadline) || sellable.At.After(deadline) || sellable.Quantity < required {
		return FlattenInDoubt
	}
	return FlattenAllowed
}
