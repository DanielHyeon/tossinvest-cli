package protection

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

type controllerGateway struct {
	mu          sync.Mutex
	scope       Scope
	clock       *time.Time
	createErr   error
	replaceErr  error
	cancelErr   error
	broker      []BrokerProtection
	createDelay time.Duration
	cancelDelay time.Duration
}

func (g *controllerGateway) Create(_ context.Context, body ConditionalBody) (BrokerProtection, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	*g.clock = g.clock.Add(g.createDelay)
	if g.createErr != nil {
		return BrokerProtection{}, g.createErr
	}
	b := BrokerProtection{Scope: g.scope, ID: "broker-create", ClientOrderID: body.ClientOrderID, Quantity: body.Quantity, Trigger: body.Trigger}
	g.broker = []BrokerProtection{b}
	return b, nil
}

func (g *controllerGateway) Replace(_ context.Context, id string, body ConditionalBody) (BrokerProtection, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.replaceErr != nil {
		return BrokerProtection{}, g.replaceErr
	}
	b := BrokerProtection{Scope: g.scope, ID: id + "-replacement", ClientOrderID: body.ClientOrderID, Quantity: body.Quantity, Trigger: body.Trigger}
	g.broker = []BrokerProtection{b}
	return b, nil
}

func (g *controllerGateway) Cancel(_ context.Context, id string) (CancelObservation, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	*g.clock = g.clock.Add(g.cancelDelay)
	if g.cancelErr != nil {
		return CancelObservation{}, g.cancelErr
	}
	for i := range g.broker {
		if g.broker[i].ID == id {
			g.broker[i].Terminal = true
		}
	}
	return CancelObservation{Scope: g.scope, BrokerID: id, Terminal: true, At: *g.clock}, nil
}

func (g *controllerGateway) Get(_ context.Context, id string) (BrokerProtection, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	for _, item := range g.broker {
		if item.ID == id {
			return item, nil
		}
	}
	return BrokerProtection{}, errors.New("missing")
}

func (g *controllerGateway) List(_ context.Context, scope Scope) ([]BrokerProtection, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if scope != g.scope {
		return nil, ErrMixedScope
	}
	return append([]BrokerProtection(nil), g.broker...), nil
}

func (g *controllerGateway) Sellable(_ context.Context, scope Scope, brokerID string) (SellableObservation, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	*g.clock = g.clock.Add(100 * time.Millisecond)
	return SellableObservation{Scope: scope, BrokerID: brokerID, Quantity: 10, At: *g.clock}, nil
}

func controllerHarness(t *testing.T) (*Controller, *Repository, *controllerGateway, *time.Time) {
	t.Helper()
	repo, err := NewRepository(openTestDB(t))
	if err != nil {
		t.Fatal(err)
	}
	scope := Scope{AccountRef: "acct-1", Profile: "primary", Market: MarketKR, Symbol: "005930"}
	clock := time.Date(2026, 8, 1, 1, 0, 0, 0, time.UTC)
	gw := &controllerGateway{scope: scope, clock: &clock, createDelay: 200 * time.Millisecond}
	sequence := 0
	c, err := NewController(repo, gw, Activation{ready: true, scope: scope}, func() time.Time { return clock }, func() string {
		sequence++
		return fmt.Sprintf("id-%d", sequence)
	})
	if err != nil {
		t.Fatal(err)
	}
	return c, repo, gw, &clock
}

func TestControllerPersistsCreateBeforeDispatchAndArmsWithinDeadline(t *testing.T) {
	c, repo, _, clock := controllerHarness(t)
	ctx := context.Background()
	fillAt := *clock
	saga, err := c.PlanFill(ctx, Fill{At: fillAt, Quantity: 1, Trigger: 70000, ExpireDate: "2026-08-08"})
	if err != nil {
		t.Fatal(err)
	}
	active, err := c.Register(ctx, saga.ID, "attempt-create", "2026-08-08", fillAt)
	if err != nil {
		t.Fatal(err)
	}
	if active.State != StateActive || active.Quantity != 1 || !c.EntryAllowed() {
		t.Fatalf("active=%+v entry=%v", active, c.EntryAllowed())
	}
	attempts, err := repo.Attempts(ctx, saga.ID)
	if err != nil || len(attempts) != 1 || attempts[0].State != MutationAcknowledged || attempts[0].ResultBrokerID == "" {
		t.Fatalf("attempts=%+v err=%v", attempts, err)
	}
	if attempts[0].CanonicalBody == "" || attempts[0].CreatedAt.After(active.UpdatedAt) {
		t.Fatalf("attempt/body order=%+v active=%+v", attempts[0], active)
	}
}

func TestControllerCreateResponseLossLatchesEntryAndRecoversByClientIdentity(t *testing.T) {
	c, repo, gw, clock := controllerHarness(t)
	ctx := context.Background()
	fillAt := *clock
	saga, _ := c.PlanFill(ctx, Fill{At: fillAt, Quantity: 3, Trigger: 70000, ExpireDate: "2026-08-08"})
	gw.createErr = errors.New("response lost")
	if _, err := c.Register(ctx, saga.ID, "attempt-lost", "2026-08-08", fillAt); !errors.Is(err, ErrMutationInDoubt) {
		t.Fatalf("register err=%v", err)
	}
	if c.EntryAllowed() {
		t.Fatal("entry remained open after uncertain create")
	}
	stored, _ := repo.Get(ctx, saga.ID)
	gw.broker = []BrokerProtection{{Scope: gw.scope, ID: "broker-recovered", ClientOrderID: stored.ClientOrderID, Quantity: 3, Trigger: 70000}}
	recovered, err := c.Recover(ctx, saga.ID)
	if err != nil || recovered.State != StateActive || recovered.BrokerID != "broker-recovered" {
		t.Fatalf("recovered=%+v err=%v", recovered, err)
	}
}

func TestControllerReplaceIsMonotonicAndRecordsNewIdentifier(t *testing.T) {
	c, repo, _, clock := controllerHarness(t)
	ctx := context.Background()
	fillAt := *clock
	saga, _ := c.PlanFill(ctx, Fill{At: fillAt, Quantity: 1, Trigger: 70000, ExpireDate: "2026-08-08"})
	active, _ := c.Register(ctx, saga.ID, "attempt-create", "2026-08-08", fillAt)
	if _, err := c.Replace(ctx, active.ID, "attempt-weak", 69000, 1, "2026-08-08"); !errors.Is(err, ErrWeakerProtection) {
		t.Fatalf("weak replacement err=%v", err)
	}
	updated, err := c.Replace(ctx, active.ID, "attempt-raise", 71000, 1, "2026-08-08")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Trigger != 71000 || updated.PreviousBrokerID != "broker-create" || updated.BrokerID == updated.PreviousBrokerID {
		t.Fatalf("updated=%+v", updated)
	}
	attempts, _ := repo.Attempts(ctx, saga.ID)
	if len(attempts) != 2 || attempts[1].ResultBrokerID != updated.BrokerID {
		t.Fatalf("attempts=%+v", attempts)
	}
}

func TestControllerFlattenIsCancelFirstBoundedAndOneShot(t *testing.T) {
	c, _, gw, clock := controllerHarness(t)
	ctx := context.Background()
	fillAt := *clock
	saga, _ := c.PlanFill(ctx, Fill{At: fillAt, Quantity: 1, Trigger: 70000, ExpireDate: "2026-08-08"})
	active, _ := c.Register(ctx, saga.ID, "attempt-create", "2026-08-08", fillAt)
	start := *clock
	gw.cancelDelay = 500 * time.Millisecond
	authorization, err := c.AuthorizeFlatten(ctx, active.ID, "attempt-cancel", start, 1)
	if err != nil {
		t.Fatal(err)
	}
	target := FlattenScope{Scope: gw.scope, BrokerID: active.BrokerID}
	if err := authorization.Consume(target, 1); err != nil {
		t.Fatal(err)
	}
	if err := authorization.Consume(target, 1); !errors.Is(err, ErrFlattenAuthorization) {
		t.Fatalf("second consume=%v", err)
	}
}

func TestControllerArmAndFlattenDeadlineFailuresStayFailClosed(t *testing.T) {
	t.Run("arm", func(t *testing.T) {
		c, _, _, clock := controllerHarness(t)
		ctx := context.Background()
		fillAt := *clock
		saga, _ := c.PlanFill(ctx, Fill{At: fillAt, Quantity: 1, Trigger: 70000, ExpireDate: "2026-08-08"})
		*clock = clock.Add(1001 * time.Millisecond)
		if _, err := c.Register(ctx, saga.ID, "attempt-late", "2026-08-08", fillAt); !errors.Is(err, ErrProtectionGap) {
			t.Fatalf("late arm=%v", err)
		}
		if c.EntryAllowed() {
			t.Fatal("entry open after arm gap")
		}
	})
	t.Run("flatten", func(t *testing.T) {
		c, _, gw, clock := controllerHarness(t)
		ctx := context.Background()
		fillAt := *clock
		saga, _ := c.PlanFill(ctx, Fill{At: fillAt, Quantity: 1, Trigger: 70000, ExpireDate: "2026-08-08"})
		active, _ := c.Register(ctx, saga.ID, "attempt-create", "2026-08-08", fillAt)
		start := *clock
		gw.cancelDelay = 2 * time.Second
		if _, err := c.AuthorizeFlatten(ctx, active.ID, "attempt-cancel", start, 1); !errors.Is(err, ErrMutationInDoubt) {
			t.Fatalf("deadline err=%v", err)
		}
		if c.EntryAllowed() {
			t.Fatal("entry open after ambiguous flatten")
		}
	})
}

func TestDesiredProtectionQuantityConvergesOneShareAndRejectsOversell(t *testing.T) {
	if got, err := DesiredProtectionQuantity(1, 0, 0); err != nil || got != 1 {
		t.Fatalf("one share=%d err=%v", got, err)
	}
	if got, err := DesiredProtectionQuantity(3, 1, 0); err != nil || got != 2 {
		t.Fatalf("partial exit=%d err=%v", got, err)
	}
	if _, err := DesiredProtectionQuantity(3, 2, 2); !errors.Is(err, ErrOversell) {
		t.Fatalf("oversell=%v", err)
	}
}

func TestA041SnapshotIsTheOnlyReplaceTriggerAndCannotWeaken(t *testing.T) {
	c, _, _, clock := controllerHarness(t)
	ctx := context.Background()
	fillAt := *clock
	saga, _ := c.PlanFill(ctx, Fill{At: fillAt, Quantity: 1, Trigger: 70000, ExpireDate: "2026-08-08"})
	active, _ := c.Register(ctx, saga.ID, "attempt-create", "2026-08-08", fillAt)
	snapshot := exitpolicy.ExitLineSnapshot{
		SnapshotID: "snap-1", DecisionID: "decision-1", InputDigest: "digest-1",
		PositionID: "position-1", PositionGeneration: 1, ObservationID: "observation-1",
		CurrentProtection: "71000",
	}
	updated, err := c.ReplaceFromExitSnapshot(ctx, active.ID, "attempt-snapshot", snapshot, 1, "2026-08-08")
	if err != nil || updated.Trigger != 71000 {
		t.Fatalf("snapshot replace=%+v err=%v", updated, err)
	}
	snapshot.CurrentProtection = "70000"
	if _, err := c.ReplaceFromExitSnapshot(ctx, updated.ID, "attempt-weaker", snapshot, 1, "2026-08-08"); !errors.Is(err, ErrWeakerProtection) {
		t.Fatalf("weaker snapshot=%v", err)
	}
	snapshot.CurrentProtection = "71000.0"
	if _, err := TriggerFromExitSnapshot(snapshot); !errors.Is(err, ErrInvalidBody) {
		t.Fatalf("noncanonical snapshot=%v", err)
	}
}

func TestPublicControllerConstructorCannotActivateFromZeroValue(t *testing.T) {
	repo, err := NewRepository(openTestDB(t))
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now
	gw := &controllerGateway{scope: Scope{AccountRef: "a", Profile: "p", Market: MarketKR, Symbol: "s"}, clock: func() *time.Time { v := now(); return &v }()}
	if _, err := NewController(repo, gw, Activation{}, now, func() string { return "id" }); !errors.Is(err, ErrProtectionInactive) {
		t.Fatalf("zero activation=%v", err)
	}
}
