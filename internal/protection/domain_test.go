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
		ID: "p-1", AccountRef: "acct-1", Profile: "prod-kr", Market: MarketKR, Symbol: "005930",
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

func TestSellClaimsAreOverflowSafeAtInt64Boundary(t *testing.T) {
	const max = int64(^uint64(0) >> 1)
	if err := ValidateSellClaims(max, SellClaims{Protection: max}); err != nil {
		t.Fatalf("exact max claim: %v", err)
	}
	for _, claims := range []SellClaims{
		{Protection: max, OpenSell: 1},
		{Protection: max - 1, OpenSell: 1, LocalReservation: 1},
	} {
		if !errors.Is(ValidateSellClaims(max, claims), ErrOversell) {
			t.Fatalf("overflowing claims %+v were accepted", claims)
		}
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
	scope := Scope{AccountRef: "acct-1", Profile: "prod-kr", Market: MarketKR, Symbol: "005930"}
	localSaga := planned()
	localSaga.State, localSaga.AttemptID, localSaga.BrokerID, localSaga.Quantity = StateActive, "a-1", "b-1", 3
	local := []Saga{localSaga}
	cases := []struct {
		name   string
		broker []BrokerProtection
		want   DiscrepancyKind
	}{
		{"missing", nil, DiscrepancyMissing},
		{"duplicate", []BrokerProtection{{Scope: scope, ID: "b-1", Quantity: 3, Trigger: 70000}, {Scope: scope, ID: "b-2", Quantity: 3, Trigger: 70000}}, DiscrepancyDuplicate},
		{"orphan", []BrokerProtection{{Scope: scope, ID: "b-x", Quantity: 1, Trigger: 10000}}, DiscrepancyOrphan},
		{"quantity", []BrokerProtection{{Scope: scope, ID: "b-1", Quantity: 2, Trigger: 70000}}, DiscrepancyQuantityMismatch},
		{"trigger", []BrokerProtection{{Scope: scope, ID: "b-1", Quantity: 3, Trigger: 69000}}, DiscrepancyTriggerMismatch},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Compare(scope, local, tc.broker)
			if err != nil {
				t.Fatal(err)
			}
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
	scope := Scope{AccountRef: "acct-1", Profile: "prod-kr", Market: MarketKR, Symbol: "005930"}
	localSaga := planned()
	localSaga.State, localSaga.AttemptID, localSaga.BrokerID = StateTriggered, "a-1", "b-1"
	broker := []BrokerProtection{{Scope: scope, ID: "finished-elsewhere", Quantity: 1, Trigger: 70000, Terminal: true}}
	got, err := Compare(scope, []Saga{localSaga}, broker)
	if err != nil || len(got) != 0 {
		t.Fatalf("completed protection produced discrepancies: %+v", got)
	}
}

func TestReconciliationRejectsMixedScopeAndDuplicateBrokerIdentity(t *testing.T) {
	scope := Scope{AccountRef: "acct-1", Profile: "prod-kr", Market: MarketKR, Symbol: "005930"}
	active := planned()
	active.State, active.AttemptID, active.BrokerID = StateActive, "a-1", "b-1"
	for _, tc := range []struct {
		name   string
		mutate func(*Scope)
	}{
		{"account", func(s *Scope) { s.AccountRef = "acct-2" }},
		{"profile", func(s *Scope) { s.Profile = "paper" }},
		{"market", func(s *Scope) { s.Market = MarketUS }},
		{"symbol", func(s *Scope) { s.Symbol = "000660" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mixed := scope
			tc.mutate(&mixed)
			if _, err := Compare(scope, []Saga{active}, []BrokerProtection{{Scope: mixed, ID: "b-1", Quantity: 1, Trigger: 70000}}); !errors.Is(err, ErrMixedScope) {
				t.Fatalf("mixed scope error = %v", err)
			}
		})
	}
	duplicate := []BrokerProtection{
		{Scope: scope, ID: "same", Quantity: 1, Trigger: 70000},
		{Scope: scope, ID: "same", Quantity: 1, Trigger: 70000, Terminal: true},
	}
	if _, err := Compare(scope, []Saga{active}, duplicate); !errors.Is(err, ErrDuplicateBrokerID) {
		t.Fatalf("duplicate broker identity error = %v", err)
	}
}

func TestFlattenDecisionRequiresTerminalCancelAndFreshSellableSnapshot(t *testing.T) {
	scope := Scope{AccountRef: "acct-1", Profile: "prod-kr", Market: MarketKR, Symbol: "005930"}
	target := FlattenScope{Scope: scope, BrokerID: "b-1"}
	deadline := now.Add(2 * time.Second)
	if got := DecideFlatten(now, deadline, target, CancelObservation{Scope: scope, BrokerID: "b-1", Terminal: true, At: now.Add(time.Second)}, SellableObservation{Scope: scope, BrokerID: "b-1", Quantity: 3, At: deadline}, 3); got != FlattenAllowed {
		t.Fatalf("decision = %s, want allowed", got)
	}
	for _, tc := range []struct {
		name string
		c    CancelObservation
		s    SellableObservation
	}{
		{"response loss", CancelObservation{}, SellableObservation{Scope: scope, BrokerID: "b-1", Quantity: 3, At: deadline}},
		{"trigger race", CancelObservation{Scope: scope, BrokerID: "b-1", Terminal: false, Triggered: true, At: now.Add(time.Second)}, SellableObservation{Scope: scope, BrokerID: "b-1", Quantity: 3, At: deadline}},
		{"late cancel", CancelObservation{Scope: scope, BrokerID: "b-1", Terminal: true, At: deadline.Add(time.Nanosecond)}, SellableObservation{Scope: scope, BrokerID: "b-1", Quantity: 3, At: deadline}},
		{"late sellable", CancelObservation{Scope: scope, BrokerID: "b-1", Terminal: true, At: deadline}, SellableObservation{Scope: scope, BrokerID: "b-1", Quantity: 3, At: deadline.Add(time.Nanosecond)}},
		{"insufficient sellable", CancelObservation{Scope: scope, BrokerID: "b-1", Terminal: true, At: now.Add(time.Second)}, SellableObservation{Scope: scope, BrokerID: "b-1", Quantity: 2, At: deadline}},
		{"reversed observation", CancelObservation{Scope: scope, BrokerID: "b-1", Terminal: true, At: deadline}, SellableObservation{Scope: scope, BrokerID: "b-1", Quantity: 3, At: now.Add(time.Second)}},
		{"wrong broker", CancelObservation{Scope: scope, BrokerID: "other", Terminal: true, At: now.Add(time.Second)}, SellableObservation{Scope: scope, BrokerID: "b-1", Quantity: 3, At: deadline}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := DecideFlatten(now, deadline, target, tc.c, tc.s, 3); got != FlattenInDoubt {
				t.Fatalf("decision = %s, want in doubt", got)
			}
		})
	}
	if got := DecideFlatten(now, deadline.Add(time.Nanosecond), target, CancelObservation{Scope: scope, BrokerID: "b-1", Terminal: true, At: now.Add(time.Second)}, SellableObservation{Scope: scope, BrokerID: "b-1", Quantity: 3, At: deadline}, 3); got != FlattenInDoubt {
		t.Fatalf("window beyond two seconds = %s", got)
	}
}

func TestSagaValidateEnforcesStateSpecificFields(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*Saga)
	}{
		{"planned attempt", func(s *Saga) { s.AttemptID = "unexpected" }},
		{"planned broker", func(s *Saga) { s.BrokerID = "unexpected" }},
		{"registering missing attempt", func(s *Saga) { s.State = StateRegistering }},
		{"active missing broker", func(s *Saga) { s.State, s.AttemptID = StateActive, "a-1" }},
		{"replacing missing pending", func(s *Saga) {
			s.State, s.AttemptID, s.BrokerID, s.PreviousBrokerID = StateReplacing, "a-2", "b-1", "b-1"
		}},
		{"reconcile missing reason", func(s *Saga) { s.State = StateReconcile }},
		{"in doubt missing attempt", func(s *Saga) { s.State, s.ReconcileReason = StateInDoubt, "unknown" }},
		{"zero revision", func(s *Saga) { s.Revision = 0 }},
		{"unknown market", func(s *Saga) { s.Market = Market("EU") }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := planned()
			tc.mutate(&s)
			if !errors.Is(s.Validate(), ErrInvalidSaga) {
				t.Fatalf("invalid saga accepted: %+v", s)
			}
		})
	}
}

func TestTransitionValidatesItsOutput(t *testing.T) {
	s := planned()
	s, _ = Transition(s, Event{Kind: EventBeginRegistration, At: now.Add(time.Millisecond), AttemptID: "a-1"})
	s, _ = Transition(s, Event{Kind: EventRegistrationActive, At: now.Add(2 * time.Millisecond), BrokerID: "b-1"})
	s, _ = Transition(s, Event{Kind: EventBeginReplace, At: now.Add(3 * time.Millisecond), AttemptID: "a-2", Trigger: 71000})
	if _, err := Transition(s, Event{Kind: EventTriggerObserved, At: now.Add(4 * time.Millisecond)}); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("replace trigger produced invalid state: %v", err)
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
	updated, err := Transition(s, Event{Kind: EventBeginRegistration, At: now.Add(time.Second), AttemptID: "a-1"})
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.Update(ctx, s.Revision, updated); err != nil {
		t.Fatal(err)
	}
	if err := repo.Update(ctx, s.Revision, updated); !errors.Is(err, ErrConcurrentUpdate) {
		t.Fatalf("stale update = %v", err)
	}
}

func TestRepositoryRejectsIdentityMutationAndStateJump(t *testing.T) {
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
	mutated, _ := Transition(s, Event{Kind: EventBeginRegistration, At: now.Add(time.Second), AttemptID: "a-1"})
	mutated.AccountRef = "acct-2"
	if err := repo.Update(ctx, s.Revision, mutated); !errors.Is(err, ErrImmutableIdentity) {
		t.Fatalf("identity mutation = %v", err)
	}
	jump := s
	jump.State, jump.AttemptID, jump.BrokerID, jump.UpdatedAt = StateActive, "a-1", "b-1", now.Add(time.Second)
	if err := repo.Update(ctx, s.Revision, jump); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("state jump = %v", err)
	}
	changed, _ := Transition(s, Event{Kind: EventBeginRegistration, At: now.Add(time.Second), AttemptID: "a-1"})
	changed.Trigger++
	if err := repo.Update(ctx, s.Revision, changed); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("trigger changed outside replacement = %v", err)
	}
	changed, _ = Transition(s, Event{Kind: EventBeginRegistration, At: now.Add(time.Second), AttemptID: "a-1"})
	changed.Revision++
	if err := repo.Update(ctx, s.Revision, changed); !errors.Is(err, ErrConcurrentUpdate) {
		t.Fatalf("input revision mismatch = %v", err)
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
func (*fakeGateway) List(context.Context, Scope) ([]BrokerProtection, error) { return nil, nil }
