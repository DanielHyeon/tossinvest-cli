package protection

import (
	"context"
	"errors"
	"testing"
	"time"
)

var now = time.Date(2026, 7, 31, 4, 0, 0, 0, time.UTC)

func planned() Saga {
	return Saga{
		ID: "p-1", AccountRef: "acct-1", Profile: "prod-kr", Market: "KR", Symbol: "005930",
		Generation: 1, Revision: 1, State: StatePlanned, Trigger: 70000, Quantity: 1,
		ClientOrderID: "protect:p-1:1", UpdatedAt: now,
	}
}

func TestSagaRegistrationCrashWindowsAreFailClosed(t *testing.T) {
	s, err := Transition(planned(), Event{Kind: EventBeginRegistration, At: now.Add(time.Millisecond), AttemptID: "a-1"})
	if err != nil || s.State != StateRegistering || s.AttemptID != "a-1" {
		t.Fatalf("begin = %+v, %v", s, err)
	}
	s, err = Transition(s, Event{Kind: EventMutationUnknown, At: now.Add(2 * time.Millisecond)})
	if err != nil || s.State != StateInDoubt {
		t.Fatalf("unknown = %+v, %v", s, err)
	}
	if _, err := Transition(s, Event{Kind: EventBeginRegistration, At: now.Add(3 * time.Millisecond), AttemptID: "a-2"}); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("IN_DOUBT was resubmitted: %v", err)
	}
}

func TestSagaActiveReplaceIsMonotonicAndKeepsBothIDs(t *testing.T) {
	s, _ := Transition(planned(), Event{Kind: EventBeginRegistration, At: now, AttemptID: "a-1"})
	s, _ = Transition(s, Event{Kind: EventRegistrationActive, At: now.Add(time.Millisecond), BrokerID: "old"})
	if _, err := Transition(s, Event{Kind: EventBeginReplace, At: now.Add(2 * time.Millisecond), Trigger: 69999, AttemptID: "a-2"}); !errors.Is(err, ErrWeakerProtection) {
		t.Fatalf("weaker replace = %v", err)
	}
	s, err := Transition(s, Event{Kind: EventBeginReplace, At: now.Add(3 * time.Millisecond), Trigger: 71000, Quantity: 1, AttemptID: "a-3"})
	if err != nil || s.State != StateReplacing || s.PreviousBrokerID != "old" {
		t.Fatalf("replace begin = %+v, %v", s, err)
	}
	s, err = Transition(s, Event{Kind: EventReplaceActive, At: now.Add(4 * time.Millisecond), BrokerID: "new"})
	if err != nil || s.State != StateActive || s.BrokerID != "new" || s.PreviousBrokerID != "old" || s.Generation != 2 {
		t.Fatalf("replace active = %+v, %v", s, err)
	}
}

func TestOneShareAndSellClaimsNeverOversell(t *testing.T) {
	if err := ValidateSellClaims(1, SellClaims{Protection: 1}); err != nil {
		t.Fatalf("one-share protection: %v", err)
	}
	for _, claims := range []SellClaims{
		{Protection: 1, OpenSell: 1},
		{Protection: 1, LocalReservation: 1},
		{Protection: 2},
	} {
		if !errors.Is(ValidateSellClaims(1, claims), ErrOversell) {
			t.Fatalf("claims %+v did not oversell one share", claims)
		}
	}
	if err := ValidateSellClaims(3, SellClaims{Protection: 3}); err != nil {
		t.Fatalf("three-share converged protection: %v", err)
	}
}

func TestInitialFillDeadlinesLatchProtectionGap(t *testing.T) {
	cases := []struct {
		name                string
		attemptAt, activeAt time.Time
		want                ArmStatus
	}{
		{"within sla", now.Add(time.Second), now.Add(2 * time.Second), ArmActive},
		{"late arm", now.Add(time.Second + time.Nanosecond), now.Add(2 * time.Second), ArmProtectionGap},
		{"late active", now.Add(time.Second), now.Add(2*time.Second + time.Nanosecond), ArmProtectionGap},
		{"no active", now.Add(time.Second), time.Time{}, ArmProtectionGap},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := EvaluateArm(now, tc.attemptAt, tc.activeAt); got != tc.want {
				t.Fatalf("status = %s, want %s", got, tc.want)
			}
		})
	}
}

func TestReconciliationClassifiesMissingDuplicateOrphanAndMismatch(t *testing.T) {
	local := []Saga{{ID: "p-1", Symbol: "005930", State: StateActive, BrokerID: "b-1", Quantity: 3, Trigger: 70000}}
	cases := []struct {
		name   string
		broker []BrokerProtection
		want   DiscrepancyKind
	}{
		{"missing", nil, DiscrepancyMissing},
		{"duplicate", []BrokerProtection{{ID: "b-1", Symbol: "005930", Quantity: 3, Trigger: 70000}, {ID: "b-2", Symbol: "005930", Quantity: 3, Trigger: 70000}}, DiscrepancyDuplicate},
		{"orphan", []BrokerProtection{{ID: "b-x", Symbol: "000660", Quantity: 1, Trigger: 10000}}, DiscrepancyOrphan},
		{"quantity", []BrokerProtection{{ID: "b-1", Symbol: "005930", Quantity: 2, Trigger: 70000}}, DiscrepancyQuantityMismatch},
		{"trigger", []BrokerProtection{{ID: "b-1", Symbol: "005930", Quantity: 3, Trigger: 69000}}, DiscrepancyTriggerMismatch},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Compare(local, tc.broker)
			found := false
			for _, d := range got {
				found = found || d.Kind == tc.want
			}
			if !found {
				t.Fatalf("discrepancies = %+v, want %s", got, tc.want)
			}
		})
	}
}

func TestReconciliationIgnoresCompletedLocalAndBrokerProtection(t *testing.T) {
	local := []Saga{{ID: "p-1", Symbol: "005930", State: StateTriggered, BrokerID: "b-1", Quantity: 1, Trigger: 70000}}
	broker := []BrokerProtection{{ID: "finished-elsewhere", Symbol: "000660", Quantity: 1, Trigger: 70000, Terminal: true}}
	if got := Compare(local, broker); len(got) != 0 {
		t.Fatalf("completed protection produced discrepancies: %+v", got)
	}
}

func TestFlattenDecisionRequiresTerminalCancelAndFreshSellableSnapshot(t *testing.T) {
	deadline := now.Add(2 * time.Second)
	if got := DecideFlatten(deadline, CancelObservation{Terminal: true, At: deadline}, SellableObservation{Quantity: 3, At: deadline}, 3); got != FlattenAllowed {
		t.Fatalf("decision = %s, want allowed", got)
	}
	for _, tc := range []struct {
		name string
		c    CancelObservation
		s    SellableObservation
	}{
		{"response loss", CancelObservation{}, SellableObservation{Quantity: 3, At: deadline}},
		{"trigger race", CancelObservation{Terminal: false, Triggered: true, At: deadline}, SellableObservation{Quantity: 3, At: deadline}},
		{"late cancel", CancelObservation{Terminal: true, At: deadline.Add(time.Nanosecond)}, SellableObservation{Quantity: 3, At: deadline}},
		{"late sellable", CancelObservation{Terminal: true, At: deadline}, SellableObservation{Quantity: 3, At: deadline.Add(time.Nanosecond)}},
		{"insufficient sellable", CancelObservation{Terminal: true, At: deadline}, SellableObservation{Quantity: 2, At: deadline}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := DecideFlatten(deadline, tc.c, tc.s, 3); got != FlattenInDoubt {
				t.Fatalf("decision = %s, want in doubt", got)
			}
		})
	}
}

func TestCanonicalBodyIsStableAndMarketSingleOnly(t *testing.T) {
	body := ConditionalBody{SerializerVersion: SerializerVersion, ClientOrderID: "protect:p-1:1", AccountRef: "acct-1", Market: "KR", Symbol: "005930", Side: "SELL", ConditionalType: "SINGLE", OrderType: "MARKET", TriggerSource: "LAST_TRADE", Trigger: 70000, Quantity: 1}
	one, err := body.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	two, _ := body.CanonicalJSON()
	if string(one) != string(two) {
		t.Fatalf("canonical body changed: %s / %s", one, two)
	}
	body.OrderType = "LIMIT"
	if _, err := body.CanonicalJSON(); !errors.Is(err, ErrInvalidBody) {
		t.Fatalf("LIMIT body accepted: %v", err)
	}
}

func TestRepositoryRoundTripAndOptimisticGeneration(t *testing.T) {
	db := openTestDB(t)
	repo, err := NewRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	s := planned()
	if err := repo.Insert(ctx, s); err != nil {
		t.Fatal(err)
	}
	got, err := repo.Get(ctx, s.ID)
	if err != nil || got.ID != s.ID || got.State != StatePlanned {
		t.Fatalf("Get = %+v, %v", got, err)
	}
	updated := s
	updated.State = StateRegistering
	updated.AttemptID = "a-1"
	updated.UpdatedAt = now.Add(time.Second)
	if err := repo.Update(ctx, s.Revision, updated); err != nil {
		t.Fatal(err)
	}
	if err := repo.Update(ctx, s.Revision, updated); !errors.Is(err, ErrConcurrentUpdate) {
		t.Fatalf("stale update = %v", err)
	}
}

// Compile-time only: production defines a contract, while every implementation
// in this change remains in a _test.go file.
var _ Gateway = (*fakeGateway)(nil)

type fakeGateway struct{}

func (*fakeGateway) Create(context.Context, ConditionalBody) (BrokerProtection, error) {
	return BrokerProtection{}, nil
}
func (*fakeGateway) Replace(context.Context, string, ConditionalBody) (BrokerProtection, error) {
	return BrokerProtection{}, nil
}
func (*fakeGateway) Cancel(context.Context, string) (CancelObservation, error) {
	return CancelObservation{}, nil
}
func (*fakeGateway) Get(context.Context, string) (BrokerProtection, error) {
	return BrokerProtection{}, nil
}
func (*fakeGateway) List(context.Context, string) ([]BrokerProtection, error) { return nil, nil }
