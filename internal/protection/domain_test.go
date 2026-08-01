package protection

import (
	"context"
	"errors"
	"strings"
	"sync"
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
	s, err = Transition(s, Event{Kind: EventMutationUnknown, At: now.Add(2 * time.Millisecond), AttemptID: "a-1"})
	if err != nil || s.State != StateInDoubt {
		t.Fatalf("unknown = %+v, %v", s, err)
	}
	if _, err := Transition(s, Event{Kind: EventBeginRegistration, At: now.Add(3 * time.Millisecond), AttemptID: "a-2"}); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("IN_DOUBT was resubmitted: %v", err)
	}
}

func TestTransitionBindsMutationAttemptAndBrokerLineage(t *testing.T) {
	registering, err := Transition(planned(), Event{Kind: EventBeginRegistration, At: now.Add(time.Millisecond), AttemptID: "a-1"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Transition(registering, Event{Kind: EventRegistrationActive, At: now.Add(2 * time.Millisecond), AttemptID: "a-forged", BrokerID: "b-forged"}); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("forged registration lineage = %v", err)
	}
	active, err := Transition(registering, Event{Kind: EventRegistrationActive, At: now.Add(2 * time.Millisecond), AttemptID: "a-1", BrokerID: "b-1"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Transition(active, Event{Kind: EventTriggerObserved, At: now.Add(3 * time.Millisecond), BrokerID: "b-2"}); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("forged trigger lineage = %v", err)
	}
	triggered, err := Transition(active, Event{Kind: EventTriggerObserved, At: now.Add(3 * time.Millisecond), BrokerID: "b-1"})
	if err != nil || triggered.State != StateTriggered || triggered.BrokerID != "b-1" {
		t.Fatalf("valid trigger lineage = %+v, %v", triggered, err)
	}
}

func TestSaga_ActiveReplace_MonotonicAndKeepsBothIDs(t *testing.T) {
	s, _ := Transition(planned(), Event{Kind: EventBeginRegistration, At: now, AttemptID: "a-1"})
	s, _ = Transition(s, Event{Kind: EventRegistrationActive, At: now.Add(time.Millisecond), AttemptID: "a-1", BrokerID: "old"})
	if _, err := Transition(s, Event{Kind: EventBeginReplace, At: now.Add(2 * time.Millisecond), Trigger: 69999, AttemptID: "a-2"}); !errors.Is(err, ErrWeakerProtection) {
		t.Fatalf("weaker replace = %v", err)
	}
	s, err := Transition(s, Event{Kind: EventBeginReplace, At: now.Add(3 * time.Millisecond), Trigger: 71000, Quantity: 1, AttemptID: "a-3"})
	if err != nil || s.State != StateReplacing || s.PreviousBrokerID != "old" {
		t.Fatalf("replace begin = %+v, %v", s, err)
	}
	s, err = Transition(s, Event{Kind: EventReplaceActive, At: now.Add(4 * time.Millisecond), AttemptID: "a-3", BrokerID: "new"})
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
	decisionAt := deadline
	got, permit := decideFlatten(now, deadline, decisionAt, target, CancelObservation{Scope: scope, BrokerID: "b-1", Terminal: true, At: now.Add(time.Second)}, SellableObservation{Scope: scope, BrokerID: "b-1", Quantity: 3, At: deadline}, 3, func() time.Time { return deadline })
	if got != FlattenAllowed {
		t.Fatalf("decision = %s, want allowed", got)
	}
	permitCopy := permit
	if err := permit.Consume(target, 3); err != nil {
		t.Fatalf("consume exact permit: %v", err)
	}
	if err := permitCopy.Consume(target, 3); !errors.Is(err, ErrFlattenAuthorization) {
		t.Fatalf("copied permit replay = %v", err)
	}
	for _, tc := range []struct {
		name     string
		target   FlattenScope
		quantity int64
	}{
		{"wrong account", FlattenScope{Scope: Scope{AccountRef: "acct-2", Profile: "prod-kr", Market: MarketKR, Symbol: "005930"}, BrokerID: "b-1"}, 3},
		{"wrong broker", FlattenScope{Scope: scope, BrokerID: "b-2"}, 3},
		{"wrong quantity", target, 2},
	} {
		t.Run("permit "+tc.name, func(t *testing.T) {
			_, scopedPermit := decideFlatten(now, deadline, deadline, target,
				CancelObservation{Scope: scope, BrokerID: "b-1", Terminal: true, At: now.Add(time.Second)},
				SellableObservation{Scope: scope, BrokerID: "b-1", Quantity: 3, At: deadline}, 3,
				func() time.Time { return deadline })
			if err := scopedPermit.Consume(tc.target, tc.quantity); !errors.Is(err, ErrFlattenAuthorization) {
				t.Fatalf("consume = %v, want ErrFlattenAuthorization", err)
			}
		})
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
			if got, _ := decideFlatten(now, deadline, deadline, target, tc.c, tc.s, 3, func() time.Time { return deadline }); got != FlattenInDoubt {
				t.Fatalf("decision = %s, want in doubt", got)
			}
		})
	}
	if got, _ := decideFlatten(now, deadline.Add(time.Nanosecond), deadline, target, CancelObservation{Scope: scope, BrokerID: "b-1", Terminal: true, At: now.Add(time.Second)}, SellableObservation{Scope: scope, BrokerID: "b-1", Quantity: 3, At: deadline}, 3, func() time.Time { return deadline }); got != FlattenInDoubt {
		t.Fatalf("window beyond two seconds = %s", got)
	}
	for _, tc := range []struct {
		name       string
		decisionAt time.Time
		cancelAt   time.Time
		sellableAt time.Time
	}{
		{"decision after deadline", deadline.Add(time.Nanosecond), now.Add(time.Second), deadline},
		{"decision clock rollback", now.Add(time.Second), now.Add(1500 * time.Millisecond), now.Add(1600 * time.Millisecond)},
		{"cancel before start", deadline, now.Add(-time.Nanosecond), deadline},
		{"sellable after decision", now.Add(1500 * time.Millisecond), now.Add(time.Second), now.Add(1600 * time.Millisecond)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := decideFlatten(now, deadline, tc.decisionAt, target,
				CancelObservation{Scope: scope, BrokerID: "b-1", Terminal: true, At: tc.cancelAt},
				SellableObservation{Scope: scope, BrokerID: "b-1", Quantity: 3, At: tc.sellableAt}, 3,
				func() time.Time { return tc.decisionAt })
			if got != FlattenInDoubt {
				t.Fatalf("decision = %s", got)
			}
		})
	}
	_, expired := decideFlatten(now, deadline, deadline, target,
		CancelObservation{Scope: scope, BrokerID: "b-1", Terminal: true, At: now.Add(time.Second)},
		SellableObservation{Scope: scope, BrokerID: "b-1", Quantity: 3, At: deadline}, 3, func() time.Time { return now.Add(time.Hour) })
	if err := expired.Consume(target, 3); !errors.Is(err, ErrFlattenAuthorization) {
		t.Fatalf("one-hour replay = %v", err)
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
	s, _ = Transition(s, Event{Kind: EventRegistrationActive, At: now.Add(2 * time.Millisecond), AttemptID: "a-1", BrokerID: "b-1"})
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
	updated, err := repo.BeginRegistration(ctx, s.ID, s.Revision, now.Add(time.Second), "a-1")
	if err != nil {
		t.Fatal(err)
	}
	if updated.State != StateRegistering || updated.AttemptID != "a-1" {
		t.Fatalf("updated = %+v", updated)
	}
	retried, err := repo.BeginRegistration(ctx, s.ID, s.Revision, now.Add(time.Second), "a-1")
	if err != nil || retried.Revision != updated.Revision {
		t.Fatalf("idempotent retry = %+v, %v", retried, err)
	}
}

func TestRepositoryMigratesV1RowsBeforeEventIdentityUse(t *testing.T) {
	db := openTestDB(t)
	if _, err := db.Exec(legacyProtectionSchema()); err != nil {
		t.Fatal(err)
	}
	repo, err := NewRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	s := planned()
	if err := repo.Insert(ctx, s); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.BeginRegistration(ctx, s.ID, s.Revision, now.Add(time.Second), "a-1"); err != nil {
		t.Fatalf("apply after v1 migration: %v", err)
	}
}

func TestRepositoryRejectsUnverifiableV1EventLineage(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*Saga)
	}{
		{"active revision one", func(s *Saga) { s.State, s.AttemptID, s.BrokerID = StateActive, "a-legacy", "b-legacy" }},
		{"registering revision one", func(s *Saga) { s.State, s.AttemptID = StateRegistering, "a-legacy" }},
		{"planned revision greater than one", func(s *Saga) { s.Revision = 2 }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := openTestDB(t)
			if _, err := db.Exec(legacyProtectionSchema()); err != nil {
				t.Fatal(err)
			}
			s := planned()
			tc.mutate(&s)
			if err := s.Validate(); err != nil {
				t.Fatalf("legacy fixture invalid: %v", err)
			}
			if _, err := db.Exec(`INSERT INTO protection_sagas (
 saga_id,account_ref,profile,market,symbol,generation,revision,state,trigger,quantity,
 pending_trigger,pending_quantity,client_order_id,attempt_id,broker_id,previous_broker_id,reconcile_reason,updated_at
) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, sagaValues(s)...); err != nil {
				t.Fatal(err)
			}
			if _, err := NewRepository(db); err == nil {
				t.Fatal("unverifiable legacy lineage was accepted")
			}
		})
	}
}

func legacyProtectionSchema() string {
	return strings.Replace(schemaDDL, ",\n  last_event_kind TEXT NOT NULL DEFAULT '',\n  last_event_fingerprint TEXT NOT NULL DEFAULT ''", "", 1)
}

func TestRepositoryRejectsNonPlannedInsertAndForgedLineage(t *testing.T) {
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
	registering, _ := Transition(s, Event{Kind: EventBeginRegistration, At: now.Add(time.Second), AttemptID: "a-1"})
	registering.ID = "p-registering"
	registering.ClientOrderID = "protect:p-registering:1"
	if err := repo.Insert(ctx, registering); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("REGISTERING insert = %v", err)
	}
	registered, err := repo.BeginRegistration(ctx, s.ID, s.Revision, now.Add(time.Second), "a-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.MarkRegistrationActive(ctx, s.ID, registered.Revision, now.Add(2*time.Second), "a-forged", "b-forged"); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("forged registration result = %v", err)
	}
}

func TestRepositoryEventSpecificMethodsPreserveLineage(t *testing.T) {
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
	s, err = repo.BeginRegistration(ctx, s.ID, s.Revision, now.Add(time.Second), "a-1")
	if err != nil {
		t.Fatal(err)
	}
	s, err = repo.MarkRegistrationActive(ctx, s.ID, s.Revision, now.Add(2*time.Second), "a-1", "b-1")
	if err != nil {
		t.Fatal(err)
	}
	s, err = repo.BeginReplace(ctx, s.ID, s.Revision, now.Add(3*time.Second), "a-2", 71000, 1)
	if err != nil {
		t.Fatal(err)
	}
	s, err = repo.MarkReplaceActive(ctx, s.ID, s.Revision, now.Add(4*time.Second), "a-2", "b-2")
	if err != nil {
		t.Fatal(err)
	}
	s, err = repo.MarkTriggerObserved(ctx, s.ID, s.Revision, now.Add(5*time.Second), "b-2")
	if err != nil {
		t.Fatal(err)
	}
	s, err = repo.Close(ctx, s.ID, s.Revision, now.Add(6*time.Second), "b-2")
	if err != nil || s.State != StateClosed || s.Revision != 7 {
		t.Fatalf("closed saga = %+v, %v", s, err)
	}

	unknown := planned()
	unknown.ID, unknown.ClientOrderID = "p-unknown", "protect:p-unknown:1"
	if err := repo.Insert(ctx, unknown); err != nil {
		t.Fatal(err)
	}
	unknown, err = repo.BeginRegistration(ctx, unknown.ID, unknown.Revision, now.Add(time.Second), "a-u")
	if err != nil {
		t.Fatal(err)
	}
	unknown, err = repo.MarkMutationUnknown(ctx, unknown.ID, unknown.Revision, now.Add(2*time.Second), "a-u")
	if err != nil || unknown.State != StateInDoubt {
		t.Fatalf("unknown saga = %+v, %v", unknown, err)
	}
	unknown, err = repo.MarkDiscrepancy(ctx, unknown.ID, unknown.Revision, now.Add(3*time.Second), "MANUAL_RECONCILE")
	if err != nil || unknown.State != StateReconcile {
		t.Fatalf("reconcile saga = %+v, %v", unknown, err)
	}
}

func TestRepositoryCASUsesRealConcurrentConnections(t *testing.T) {
	dbA, dbB := openConcurrentTestDBs(t)
	repoA, err := NewRepository(dbA)
	if err != nil {
		t.Fatal(err)
	}
	repoB, err := NewRepository(dbB)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := repoA.Insert(ctx, planned()); err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for index, attemptID := range []string{"a-1", "a-2"} {
		attemptID, repo := attemptID, repoA
		if index == 1 {
			repo = repoB
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, applyErr := repo.BeginRegistration(ctx, "p-1", 1, now.Add(time.Second), attemptID)
			results <- applyErr
		}()
	}
	close(start)
	wg.Wait()
	close(results)
	var succeeded, conflicted int
	for result := range results {
		switch {
		case result == nil:
			succeeded++
		case errors.Is(result, ErrConcurrentUpdate):
			conflicted++
		default:
			t.Fatalf("unexpected concurrent result: %v", result)
		}
	}
	if succeeded != 1 || conflicted != 1 {
		t.Fatalf("success=%d conflict=%d", succeeded, conflicted)
	}
	got, err := repoA.Get(ctx, "p-1")
	if err != nil || got.Revision != 2 || (got.AttemptID != "a-1" && got.AttemptID != "a-2") {
		t.Fatalf("stored = %+v, %v", got, err)
	}
}

func TestRepositoryConcurrentSameEventIsIdempotent(t *testing.T) {
	dbA, dbB := openConcurrentTestDBs(t)
	repoA, err := NewRepository(dbA)
	if err != nil {
		t.Fatal(err)
	}
	repoB, err := NewRepository(dbB)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := repoA.Insert(ctx, planned()); err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		repo := repoA
		if i == 1 {
			repo = repoB
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, applyErr := repo.BeginRegistration(ctx, "p-1", 1, now.Add(time.Second), "a-1")
			results <- applyErr
		}()
	}
	close(start)
	wg.Wait()
	close(results)
	for result := range results {
		if result != nil {
			t.Fatalf("idempotent concurrent result: %v", result)
		}
	}
	got, err := repoA.Get(ctx, "p-1")
	if err != nil || got.Revision != 2 || got.AttemptID != "a-1" {
		t.Fatalf("stored = %+v, %v", got, err)
	}
}

func TestRepositoryStaleRetryCannotMasqueradeAsDifferentEventKind(t *testing.T) {
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
	s, err = repo.BeginRegistration(ctx, s.ID, s.Revision, now.Add(time.Second), "a-1")
	if err != nil {
		t.Fatal(err)
	}
	s, err = repo.MarkRegistrationActive(ctx, s.ID, s.Revision, now.Add(2*time.Second), "a-1", "b-1")
	if err != nil {
		t.Fatal(err)
	}
	s, err = repo.BeginReplace(ctx, s.ID, s.Revision, now.Add(3*time.Second), "a-2", 71000, 1)
	if err != nil {
		t.Fatal(err)
	}
	beforeReplaceActiveRevision := s.Revision
	replaceActiveAt := now.Add(4 * time.Second)
	s, err = repo.MarkReplaceActive(ctx, s.ID, s.Revision, replaceActiveAt, "a-2", "b-2")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.MarkRegistrationActive(ctx, s.ID, beforeReplaceActiveRevision, replaceActiveAt, "a-2", "b-2"); !errors.Is(err, ErrConcurrentUpdate) {
		t.Fatalf("cross-event stale retry = %v, want ErrConcurrentUpdate", err)
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
