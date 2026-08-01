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
	createWait  bool
	listWait    bool
	cancelWait  bool
	createCalls int
	createDL    time.Time
	listDL      time.Time
	cancelDL    time.Time
}

func (g *controllerGateway) Create(ctx context.Context, body ConditionalBody) (BrokerProtection, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.createCalls++
	g.createDL, _ = ctx.Deadline()
	if g.createWait {
		<-ctx.Done()
		return BrokerProtection{}, ctx.Err()
	}
	*g.clock = g.clock.Add(g.createDelay)
	if g.createErr != nil {
		return BrokerProtection{}, g.createErr
	}
	b := BrokerProtection{Scope: g.scope, ID: "broker-create", ClientOrderID: body.ClientOrderID, Quantity: body.Quantity, Trigger: body.Trigger, OrderSide: "SELL", OrderType: "MARKET", ConditionType: "STOP", ExpireDate: body.ExpireDate}
	g.broker = []BrokerProtection{b}
	return b, nil
}

func (g *controllerGateway) Replace(_ context.Context, id string, body ConditionalBody) (BrokerProtection, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.replaceErr != nil {
		return BrokerProtection{}, g.replaceErr
	}
	b := BrokerProtection{Scope: g.scope, ID: id + "-replacement", ClientOrderID: body.ClientOrderID, Quantity: body.Quantity, Trigger: body.Trigger, OrderSide: "SELL", OrderType: "MARKET", ConditionType: "STOP", ExpireDate: body.ExpireDate}
	g.broker = []BrokerProtection{b}
	return b, nil
}

func (g *controllerGateway) Cancel(ctx context.Context, target BrokerTarget) (CancelObservation, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.cancelDL, _ = ctx.Deadline()
	if g.cancelWait {
		<-ctx.Done()
		return CancelObservation{}, ctx.Err()
	}
	*g.clock = g.clock.Add(g.cancelDelay)
	if g.cancelErr != nil {
		return CancelObservation{}, g.cancelErr
	}
	for i := range g.broker {
		if g.broker[i].ID == target.BrokerID {
			g.broker[i].Terminal = true
		}
	}
	return CancelObservation{Scope: g.scope, BrokerID: target.BrokerID, ClientOrderID: target.ClientOrderID, Terminal: true, At: *g.clock}, nil
}

func (g *controllerGateway) Get(_ context.Context, target BrokerTarget) (BrokerProtection, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	for _, item := range g.broker {
		if item.ID == target.BrokerID {
			return item, nil
		}
	}
	return BrokerProtection{}, errors.New("missing")
}

func (g *controllerGateway) List(ctx context.Context, scope Scope) ([]BrokerProtection, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.listDL, _ = ctx.Deadline()
	if g.listWait {
		<-ctx.Done()
		return nil, ctx.Err()
	}
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
	gw.broker = []BrokerProtection{{Scope: gw.scope, ID: "broker-recovered", ClientOrderID: stored.ClientOrderID, Quantity: 3, Trigger: 70000, OrderSide: "SELL", OrderType: "MARKET", ConditionType: "STOP", ExpireDate: "2026-08-08"}}
	recovered, err := c.Recover(ctx, saga.ID)
	if err != nil || recovered.State != StateActive || recovered.BrokerID != "broker-recovered" {
		t.Fatalf("recovered=%+v err=%v", recovered, err)
	}
}

func TestRestartKeepsEntryClosedUntilEverySagaIsExactlyRecovered(t *testing.T) {
	c, repo, gw, clock := controllerHarness(t)
	ctx := context.Background()
	fillAt := *clock
	saga, _ := c.PlanFill(ctx, Fill{At: fillAt, Quantity: 1, Trigger: 70000, ExpireDate: "2026-08-08"})
	active, _ := c.Register(ctx, saga.ID, "attempt-create", "2026-08-08", fillAt)

	restarted, err := NewController(repo, gw, Activation{ready: true, scope: gw.scope}, func() time.Time { return *clock }, func() string { return "restart-id" })
	if err != nil {
		t.Fatal(err)
	}
	if restarted.EntryAllowed() {
		t.Fatal("restart opened entry before broker recovery")
	}
	recovered, err := restarted.Recover(ctx, active.ID)
	if err != nil || recovered.State != StateActive || !restarted.EntryAllowed() {
		t.Fatalf("recovered=%+v err=%v entry=%v", recovered, err, restarted.EntryAllowed())
	}
}

func TestRegisterErrorsAndBrokerWaitStayFailClosedWithDeadline(t *testing.T) {
	t.Run("missing saga", func(t *testing.T) {
		c, _, _, _ := controllerHarness(t)
		if _, err := c.Register(context.Background(), "missing", "attempt", "2026-08-08", time.Now()); err == nil || c.EntryAllowed() {
			t.Fatalf("err=%v entry=%v", err, c.EntryAllowed())
		}
	})
	t.Run("invalid body", func(t *testing.T) {
		c, _, _, clock := controllerHarness(t)
		saga, _ := c.PlanFill(context.Background(), Fill{At: *clock, Quantity: 1, Trigger: 70000, ExpireDate: "2026-08-08"})
		if _, err := c.Register(context.Background(), saga.ID, "attempt", "bad-date", *clock); err == nil || c.EntryAllowed() {
			t.Fatalf("err=%v entry=%v", err, c.EntryAllowed())
		}
	})
	t.Run("broker wait", func(t *testing.T) {
		c, _, gw, clock := controllerHarness(t)
		saga, _ := c.PlanFill(context.Background(), Fill{At: *clock, Quantity: 1, Trigger: 70000, ExpireDate: "2026-08-08"})
		gw.createWait = true
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		if _, err := c.Register(ctx, saga.ID, "attempt", "2026-08-08", *clock); !errors.Is(err, ErrMutationInDoubt) || c.EntryAllowed() {
			t.Fatalf("err=%v entry=%v", err, c.EntryAllowed())
		}
		if gw.createDL.IsZero() {
			t.Fatal("create received no deadline")
		}
	})
}

func TestRecoverMissingOrTerminalUntriggeredStaysReconcileAndClosed(t *testing.T) {
	for _, terminal := range []bool{false, true} {
		t.Run(fmt.Sprintf("terminal-%v", terminal), func(t *testing.T) {
			c, repo, gw, clock := controllerHarness(t)
			ctx := context.Background()
			saga, _ := c.PlanFill(ctx, Fill{At: *clock, Quantity: 1, Trigger: 70000, ExpireDate: "2026-08-08"})
			active, _ := c.Register(ctx, saga.ID, "attempt", "2026-08-08", *clock)
			if terminal {
				gw.broker[0].Terminal = true
			} else {
				gw.broker = nil
			}
			if _, err := c.Recover(ctx, active.ID); !errors.Is(err, ErrProtectionGone) || c.EntryAllowed() {
				t.Fatalf("recover err=%v entry=%v", err, c.EntryAllowed())
			}
			stored, _ := repo.Get(ctx, active.ID)
			if stored.State != StateReconcile {
				t.Fatalf("state=%s", stored.State)
			}
		})
	}
}

func TestRecoverAndFlattenBrokerWaitsAreBoundedAndStayClosed(t *testing.T) {
	t.Run("recover", func(t *testing.T) {
		c, _, gw, clock := controllerHarness(t)
		ctx := context.Background()
		saga, _ := c.PlanFill(ctx, Fill{At: *clock, Quantity: 1, Trigger: 70000, ExpireDate: "2026-08-08"})
		active, _ := c.Register(ctx, saga.ID, "attempt", "2026-08-08", *clock)
		gw.listWait = true
		waitCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
		defer cancel()
		if _, err := c.Recover(waitCtx, active.ID); !errors.Is(err, ErrMutationInDoubt) || c.EntryAllowed() {
			t.Fatalf("recover err=%v entry=%v", err, c.EntryAllowed())
		}
		if gw.listDL.IsZero() {
			t.Fatal("recovery list received no deadline")
		}
	})
	t.Run("flatten cancel", func(t *testing.T) {
		c, _, gw, clock := controllerHarness(t)
		ctx := context.Background()
		saga, _ := c.PlanFill(ctx, Fill{At: *clock, Quantity: 1, Trigger: 70000, ExpireDate: "2026-08-08"})
		active, _ := c.Register(ctx, saga.ID, "attempt", "2026-08-08", *clock)
		gw.cancelWait = true
		waitCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
		defer cancel()
		if _, err := c.AuthorizeFlatten(waitCtx, active.ID, "cancel-attempt", *clock, 1); !errors.Is(err, ErrMutationInDoubt) || c.EntryAllowed() {
			t.Fatalf("flatten err=%v entry=%v", err, c.EntryAllowed())
		}
		if gw.cancelDL.IsZero() {
			t.Fatal("cancel received no deadline")
		}
	})
}

func TestConcurrentRegisterDispatchesOnceAndLoserCannotOpenEntry(t *testing.T) {
	c, _, gw, clock := controllerHarness(t)
	ctx := context.Background()
	fillAt := *clock
	saga, _ := c.PlanFill(ctx, Fill{At: fillAt, Quantity: 1, Trigger: 70000, ExpireDate: "2026-08-08"})
	start := make(chan struct{})
	errs := make(chan error, 2)
	for _, attempt := range []string{"attempt-a", "attempt-b"} {
		go func(attempt string) {
			<-start
			_, err := c.Register(ctx, saga.ID, attempt, "2026-08-08", fillAt)
			errs <- err
		}(attempt)
	}
	close(start)
	err1, err2 := <-errs, <-errs
	if (err1 == nil) == (err2 == nil) || gw.createCalls != 1 || c.EntryAllowed() {
		t.Fatalf("errs=(%v,%v) calls=%d entry=%v", err1, err2, gw.createCalls, c.EntryAllowed())
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

func TestReplaceResponseLossRecoversOnlyTheExactNewIdentity(t *testing.T) {
	c, repo, gw, clock := controllerHarness(t)
	ctx := context.Background()
	fillAt := *clock
	saga, _ := c.PlanFill(ctx, Fill{At: fillAt, Quantity: 1, Trigger: 70000, ExpireDate: "2026-08-08"})
	active, _ := c.Register(ctx, saga.ID, "attempt-create", "2026-08-08", fillAt)
	gw.replaceErr = errors.New("response lost")
	if _, err := c.Replace(ctx, active.ID, "attempt-replace", 71000, 1, "2026-08-08"); !errors.Is(err, ErrMutationInDoubt) {
		t.Fatalf("replace=%v", err)
	}
	stored, _ := repo.Get(ctx, active.ID)
	gw.replaceErr = nil
	gw.broker = []BrokerProtection{{Scope: gw.scope, ID: "broker-replaced", ClientOrderID: stored.ClientOrderID,
		Quantity: 1, Trigger: 71000, OrderSide: "SELL", OrderType: "MARKET", ConditionType: "STOP", ExpireDate: "2026-08-08"}}
	recovered, err := c.Recover(ctx, stored.ID)
	if err != nil || recovered.State != StateActive || recovered.BrokerID != "broker-replaced" || recovered.Trigger != 71000 || !c.EntryAllowed() {
		t.Fatalf("recovered=%+v err=%v entry=%v", recovered, err, c.EntryAllowed())
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
