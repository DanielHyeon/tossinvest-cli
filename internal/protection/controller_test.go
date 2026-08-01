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
	mu                    sync.Mutex
	scope                 Scope
	clock                 *time.Time
	createErr             error
	replaceErr            error
	cancelErr             error
	createExpiryOverride  string
	replaceExpiryOverride string
	broker                []BrokerProtection
	createDelay           time.Duration
	cancelDelay           time.Duration
	createWait            bool
	listWait              bool
	listErr               error
	cancelWait            bool
	createCalls           int
	createDL              time.Time
	listDL                time.Time
	cancelDL              time.Time
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
	if g.createExpiryOverride != "" {
		b.ExpireDate = g.createExpiryOverride
	}
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
	if g.replaceExpiryOverride != "" {
		b.ExpireDate = g.replaceExpiryOverride
	}
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
	if g.listErr != nil {
		return nil, g.listErr
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

func stageCreateCrash(t *testing.T, c *Controller, repo *Repository, clock *time.Time, state MutationState) Saga {
	t.Helper()
	ctx := context.Background()
	fillAt := *clock
	saga, err := c.PlanFill(ctx, Fill{At: fillAt, Quantity: 1, Trigger: 70000, ExpireDate: "2026-08-08"})
	if err != nil {
		t.Fatal(err)
	}
	body := bodyForSaga(saga, "2026-08-08")
	canonical, _ := body.CanonicalJSON()
	at := fillAt.Add(100 * time.Millisecond)
	*clock = at
	if err := repo.recordAttempt(ctx, MutationAttempt{ID: "crash-create", SagaID: saga.ID, Generation: saga.Generation,
		Kind: MutationCreate, State: MutationPlanned, SerializerVersion: SerializerVersion, CanonicalBody: string(canonical), CreatedAt: at, UpdatedAt: at}); err != nil {
		t.Fatal(err)
	}
	registering, err := repo.BeginRegistration(ctx, saga.ID, saga.Revision, at, "crash-create")
	if err != nil {
		t.Fatal(err)
	}
	if state == MutationDispatched || state == MutationAcknowledged {
		if err := repo.markAttempt(ctx, "crash-create", MutationPlanned, MutationDispatched, at, ""); err != nil {
			t.Fatal(err)
		}
	}
	if state == MutationAcknowledged {
		if err := repo.markAttempt(ctx, "crash-create", MutationDispatched, MutationAcknowledged, at, "broker-crash-create"); err != nil {
			t.Fatal(err)
		}
	}
	return registering
}

func stageReplaceCrash(t *testing.T, repo *Repository, active Saga, clock *time.Time, state MutationState) Saga {
	t.Helper()
	ctx := context.Background()
	body := bodyForSaga(active, "2026-08-08")
	body.Trigger = 72000
	canonical, _ := body.CanonicalJSON()
	at := clock.Add(100 * time.Millisecond)
	*clock = at
	if err := repo.recordAttempt(ctx, MutationAttempt{ID: "crash-replace", SagaID: active.ID, Generation: active.Generation,
		Kind: MutationReplace, State: MutationPlanned, SerializerVersion: SerializerVersion, CanonicalBody: string(canonical),
		TargetBrokerID: active.BrokerID, CreatedAt: at, UpdatedAt: at}); err != nil {
		t.Fatal(err)
	}
	replacing, err := repo.BeginReplace(ctx, active.ID, active.Revision, at, "crash-replace", 72000, 1)
	if err != nil {
		t.Fatal(err)
	}
	if state == MutationDispatched || state == MutationAcknowledged {
		if err := repo.markAttempt(ctx, "crash-replace", MutationPlanned, MutationDispatched, at, ""); err != nil {
			t.Fatal(err)
		}
	}
	if state == MutationAcknowledged {
		if err := repo.markAttempt(ctx, "crash-replace", MutationDispatched, MutationAcknowledged, at, "broker-crash-replace"); err != nil {
			t.Fatal(err)
		}
	}
	return replacing
}

func TestNewControllerNeverOpensUntilBoundedAuthoritativeInventoryIsClean(t *testing.T) {
	c, _, gw, _ := controllerHarness(t)
	if c.EntryAllowed() {
		t.Fatal("empty local repository opened before broker inventory proof")
	}
	if _, err := c.Reconcile(context.Background()); err != nil || !c.EntryAllowed() {
		t.Fatalf("clean reconcile err=%v entry=%v", err, c.EntryAllowed())
	}
	if gw.listDL.IsZero() {
		t.Fatal("startup clean proof used an unbounded broker list")
	}

	c2, _, gw2, _ := controllerHarness(t)
	gw2.listErr = errors.New("broker unavailable")
	if _, err := c2.Reconcile(context.Background()); err == nil || c2.EntryAllowed() {
		t.Fatalf("unavailable reconcile err=%v entry=%v", err, c2.EntryAllowed())
	}
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
	gw.broker = []BrokerProtection{
		{Scope: gw.scope, ID: active.BrokerID, ClientOrderID: stored.ClientOrderID, Quantity: 1, Trigger: 70000,
			Terminal: true, OrderSide: "SELL", OrderType: "MARKET", ConditionType: "STOP", ExpireDate: "2026-08-08"},
		{Scope: gw.scope, ID: "broker-replaced", ClientOrderID: stored.ClientOrderID,
			Quantity: 1, Trigger: 71000, OrderSide: "SELL", OrderType: "MARKET", ConditionType: "STOP", ExpireDate: "2026-08-08"},
	}
	recovered, err := c.Recover(ctx, stored.ID)
	if err != nil || recovered.State != StateActive || recovered.BrokerID != "broker-replaced" || recovered.Trigger != 71000 || !c.EntryAllowed() {
		t.Fatalf("recovered=%+v err=%v entry=%v", recovered, err, c.EntryAllowed())
	}
	restarted, _ := NewController(repo, gw, Activation{ready: true, scope: gw.scope}, func() time.Time { return *clock }, func() string { return "response-loss-restart" })
	if again, restartErr := restarted.Recover(ctx, stored.ID); restartErr != nil || again.BrokerID != "broker-replaced" || !restarted.EntryAllowed() {
		t.Fatalf("restart recovered=%+v err=%v entry=%v", again, restartErr, restarted.EntryAllowed())
	}
}

func TestCrashWindowsRecoverDurableCreateAndRepeatedReplaceIdentity(t *testing.T) {
	for _, attemptState := range []MutationState{MutationDispatched, MutationAcknowledged} {
		t.Run("create-"+string(attemptState), func(t *testing.T) {
			c, repo, gw, clock := controllerHarness(t)
			registering := stageCreateCrash(t, c, repo, clock, attemptState)
			gw.broker = []BrokerProtection{{Scope: gw.scope, ID: "broker-crash-create", ClientOrderID: registering.ClientOrderID,
				Quantity: 1, Trigger: 70000, OrderSide: "SELL", OrderType: "MARKET", ConditionType: "STOP", ExpireDate: "2026-08-08"}}
			restarted, _ := NewController(repo, gw, Activation{ready: true, scope: gw.scope}, func() time.Time { return *clock }, func() string { return "restart" })
			active, err := restarted.Recover(context.Background(), registering.ID)
			if err != nil || active.State != StateActive || active.BrokerID != "broker-crash-create" || !restarted.EntryAllowed() {
				t.Fatalf("recover=%+v err=%v entry=%v", active, err, restarted.EntryAllowed())
			}
			attempts, _ := repo.Attempts(context.Background(), registering.ID)
			if attempts[0].State != MutationAcknowledged || attempts[0].ResultBrokerID != "broker-crash-create" {
				t.Fatalf("attempt=%+v", attempts[0])
			}
			restartedAgain, _ := NewController(repo, gw, Activation{ready: true, scope: gw.scope}, func() time.Time { return *clock }, func() string { return "restart-again" })
			if discrepancies, reconcileErr := restartedAgain.Reconcile(context.Background()); reconcileErr != nil || len(discrepancies) != 0 || !restartedAgain.EntryAllowed() {
				t.Fatalf("reconcile=%+v err=%v entry=%v", discrepancies, reconcileErr, restartedAgain.EntryAllowed())
			}
		})

		t.Run("replace-"+string(attemptState), func(t *testing.T) {
			c, repo, gw, clock := controllerHarness(t)
			ctx := context.Background()
			fillAt := *clock
			saga, _ := c.PlanFill(ctx, Fill{At: fillAt, Quantity: 1, Trigger: 70000, ExpireDate: "2026-08-08"})
			one, _ := c.Register(ctx, saga.ID, "create", "2026-08-08", fillAt)
			two, _ := c.Replace(ctx, one.ID, "replace-one", 71000, 1, "2026-08-08")
			replacing := stageReplaceCrash(t, repo, two, clock, attemptState)
			gw.broker = []BrokerProtection{
				{Scope: gw.scope, ID: one.BrokerID, ClientOrderID: replacing.ClientOrderID, Quantity: 1, Trigger: 70000, Terminal: true, OrderSide: "SELL", OrderType: "MARKET", ConditionType: "STOP", ExpireDate: "2026-08-08"},
				{Scope: gw.scope, ID: two.BrokerID, ClientOrderID: replacing.ClientOrderID, Quantity: 1, Trigger: 71000, Terminal: true, OrderSide: "SELL", OrderType: "MARKET", ConditionType: "STOP", ExpireDate: "2026-08-08"},
				{Scope: gw.scope, ID: "broker-crash-replace", ClientOrderID: replacing.ClientOrderID, Quantity: 1, Trigger: 72000, OrderSide: "SELL", OrderType: "MARKET", ConditionType: "STOP", ExpireDate: "2026-08-08"},
			}
			restarted, _ := NewController(repo, gw, Activation{ready: true, scope: gw.scope}, func() time.Time { return *clock }, func() string { return "restart" })
			active, err := restarted.Recover(ctx, replacing.ID)
			if err != nil || active.State != StateActive || active.Generation != 3 || active.BrokerID != "broker-crash-replace" || !restarted.EntryAllowed() {
				t.Fatalf("recover=%+v err=%v entry=%v", active, err, restarted.EntryAllowed())
			}
			restartedAgain, _ := NewController(repo, gw, Activation{ready: true, scope: gw.scope}, func() time.Time { return *clock }, func() string { return "restart-again" })
			if discrepancies, reconcileErr := restartedAgain.Reconcile(ctx); reconcileErr != nil || len(discrepancies) != 0 || !restartedAgain.EntryAllowed() {
				t.Fatalf("reconcile=%+v err=%v entry=%v", discrepancies, reconcileErr, restartedAgain.EntryAllowed())
			}
		})
	}
}

func TestCrashWindowUnknownDispatchRequiresExactlyOneCanonicalBrokerRow(t *testing.T) {
	for _, tc := range []struct {
		name  string
		state MutationState
		count int
	}{
		{name: "planned is never claimed", state: MutationPlanned, count: 1},
		{name: "dispatched missing", state: MutationDispatched, count: 0},
		{name: "dispatched duplicate", state: MutationDispatched, count: 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, repo, gw, clock := controllerHarness(t)
			registering := stageCreateCrash(t, c, repo, clock, tc.state)
			for i := 0; i < tc.count; i++ {
				gw.broker = append(gw.broker, BrokerProtection{Scope: gw.scope, ID: fmt.Sprintf("broker-candidate-%d", i), ClientOrderID: registering.ClientOrderID,
					Quantity: 1, Trigger: 70000, OrderSide: "SELL", OrderType: "MARKET", ConditionType: "STOP", ExpireDate: "2026-08-08"})
			}
			restarted, _ := NewController(repo, gw, Activation{ready: true, scope: gw.scope}, func() time.Time { return *clock }, func() string { return "restart" })
			if _, err := restarted.Recover(context.Background(), registering.ID); !errors.Is(err, ErrMutationInDoubt) || restarted.EntryAllowed() {
				t.Fatalf("recover err=%v entry=%v", err, restarted.EntryAllowed())
			}
		})
	}
}

func TestExpiryIsExactAcrossCreateRecoverAndReconcile(t *testing.T) {
	t.Run("create response mismatch", func(t *testing.T) {
		c, _, gw, clock := controllerHarness(t)
		gw.createExpiryOverride = "2026-08-09"
		fillAt := *clock
		saga, _ := c.PlanFill(context.Background(), Fill{At: fillAt, Quantity: 1, Trigger: 70000, ExpireDate: "2026-08-08"})
		if _, err := c.Register(context.Background(), saga.ID, "create", "2026-08-08", fillAt); !errors.Is(err, ErrMutationInDoubt) || c.EntryAllowed() {
			t.Fatalf("register err=%v entry=%v", err, c.EntryAllowed())
		}
	})

	t.Run("replace response mismatch", func(t *testing.T) {
		c, _, gw, clock := controllerHarness(t)
		ctx := context.Background()
		fillAt := *clock
		saga, _ := c.PlanFill(ctx, Fill{At: fillAt, Quantity: 1, Trigger: 70000, ExpireDate: "2026-08-08"})
		active, _ := c.Register(ctx, saga.ID, "create", "2026-08-08", fillAt)
		gw.replaceExpiryOverride = "2026-08-09"
		if _, err := c.Replace(ctx, active.ID, "replace", 71000, 1, "2026-08-08"); !errors.Is(err, ErrMutationInDoubt) || c.EntryAllowed() {
			t.Fatalf("replace err=%v entry=%v", err, c.EntryAllowed())
		}
	})

	t.Run("restart inventory mismatch", func(t *testing.T) {
		c, repo, gw, clock := controllerHarness(t)
		ctx := context.Background()
		fillAt := *clock
		saga, _ := c.PlanFill(ctx, Fill{At: fillAt, Quantity: 1, Trigger: 70000, ExpireDate: "2026-08-08"})
		active, _ := c.Register(ctx, saga.ID, "create", "2026-08-08", fillAt)
		gw.broker[0].ExpireDate = "2026-08-09"
		reconcileRestart, _ := NewController(repo, gw, Activation{ready: true, scope: gw.scope}, func() time.Time { return *clock }, func() string { return "reconcile" })
		if _, err := reconcileRestart.Reconcile(ctx); !errors.Is(err, ErrMutationInDoubt) || reconcileRestart.EntryAllowed() {
			t.Fatalf("reconcile err=%v entry=%v", err, reconcileRestart.EntryAllowed())
		}
		recoverRestart, _ := NewController(repo, gw, Activation{ready: true, scope: gw.scope}, func() time.Time { return *clock }, func() string { return "recover" })
		if _, err := recoverRestart.Recover(ctx, active.ID); !errors.Is(err, ErrMutationInDoubt) || recoverRestart.EntryAllowed() {
			t.Fatalf("recover err=%v entry=%v", err, recoverRestart.EntryAllowed())
		}
	})
}

func TestRestartRecoverAndReconcileIgnoreOnlyExactRetiredReplaceLineage(t *testing.T) {
	c, repo, gw, clock := controllerHarness(t)
	ctx := context.Background()
	fillAt := *clock
	saga, _ := c.PlanFill(ctx, Fill{At: fillAt, Quantity: 1, Trigger: 70000, ExpireDate: "2026-08-08"})
	one, _ := c.Register(ctx, saga.ID, "create", "2026-08-08", fillAt)
	two, _ := c.Replace(ctx, one.ID, "replace-1", 71000, 1, "2026-08-08")
	three, _ := c.Replace(ctx, two.ID, "replace-2", 72000, 1, "2026-08-08")
	gw.broker = []BrokerProtection{
		{Scope: gw.scope, ID: one.BrokerID, ClientOrderID: three.ClientOrderID, Quantity: 1, Trigger: 70000, Terminal: true, OrderSide: "SELL", OrderType: "MARKET", ConditionType: "STOP", ExpireDate: "2026-08-08"},
		{Scope: gw.scope, ID: two.BrokerID, ClientOrderID: three.ClientOrderID, Quantity: 1, Trigger: 71000, Terminal: true, OrderSide: "SELL", OrderType: "MARKET", ConditionType: "STOP", ExpireDate: "2026-08-08"},
		{Scope: gw.scope, ID: three.BrokerID, ClientOrderID: three.ClientOrderID, Quantity: 1, Trigger: 72000, OrderSide: "SELL", OrderType: "MARKET", ConditionType: "STOP", ExpireDate: "2026-08-08"},
	}

	restarted, _ := NewController(repo, gw, Activation{ready: true, scope: gw.scope}, func() time.Time { return *clock }, func() string { return "restart" })
	if _, err := restarted.Recover(ctx, three.ID); err != nil || !restarted.EntryAllowed() {
		t.Fatalf("recover err=%v entry=%v", err, restarted.EntryAllowed())
	}
	restarted2, _ := NewController(repo, gw, Activation{ready: true, scope: gw.scope}, func() time.Time { return *clock }, func() string { return "restart-2" })
	if discrepancies, err := restarted2.Reconcile(ctx); err != nil || len(discrepancies) != 0 || !restarted2.EntryAllowed() {
		t.Fatalf("reconcile discrepancies=%+v err=%v entry=%v", discrepancies, err, restarted2.EntryAllowed())
	}

	gw.broker[0].Trigger = 69999
	restarted3, _ := NewController(repo, gw, Activation{ready: true, scope: gw.scope}, func() time.Time { return *clock }, func() string { return "restart-3" })
	if _, err := restarted3.Recover(ctx, three.ID); !errors.Is(err, ErrMutationInDoubt) || restarted3.EntryAllowed() {
		t.Fatalf("mismatched retired err=%v entry=%v", err, restarted3.EntryAllowed())
	}
	gw.broker[0].Trigger = 70000
	gw.broker = append(gw.broker, BrokerProtection{Scope: gw.scope, ID: "unrelated", ClientOrderID: three.ClientOrderID,
		Quantity: 1, Trigger: 70000, Terminal: true, OrderSide: "SELL", OrderType: "MARKET", ConditionType: "STOP", ExpireDate: "2026-08-08"})
	restarted4, _ := NewController(repo, gw, Activation{ready: true, scope: gw.scope}, func() time.Time { return *clock }, func() string { return "restart-4" })
	if _, err := restarted4.Recover(ctx, three.ID); !errors.Is(err, ErrMutationInDoubt) || restarted4.EntryAllowed() {
		t.Fatalf("unrelated retired err=%v entry=%v", err, restarted4.EntryAllowed())
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

func TestRegisterUsesPersistedFillTimeInsteadOfCallerDeadlineAuthority(t *testing.T) {
	t.Run("delayed caller cannot restart arm clock", func(t *testing.T) {
		c, _, gw, clock := controllerHarness(t)
		ctx := context.Background()
		durableFillAt := *clock
		saga, _ := c.PlanFill(ctx, Fill{At: durableFillAt, Quantity: 1, Trigger: 70000, ExpireDate: "2026-08-08"})
		*clock = clock.Add(1500 * time.Millisecond)
		if _, err := c.Register(ctx, saga.ID, "delayed", "2026-08-08", *clock); !errors.Is(err, ErrProtectionGap) || c.EntryAllowed() {
			t.Fatalf("delayed register err=%v entry=%v", err, c.EntryAllowed())
		}
		if gw.createCalls != 0 {
			t.Fatalf("delayed caller dispatched create calls=%d", gw.createCalls)
		}
	})

	t.Run("nearby caller timestamp must exactly match durable fill", func(t *testing.T) {
		c, _, gw, clock := controllerHarness(t)
		ctx := context.Background()
		durableFillAt := *clock
		saga, _ := c.PlanFill(ctx, Fill{At: durableFillAt, Quantity: 1, Trigger: 70000, ExpireDate: "2026-08-08"})
		*clock = clock.Add(100 * time.Millisecond)
		if _, err := c.Register(ctx, saga.ID, "shifted", "2026-08-08", durableFillAt.Add(100*time.Millisecond)); !errors.Is(err, ErrProtectionGap) || c.EntryAllowed() {
			t.Fatalf("shifted register err=%v entry=%v", err, c.EntryAllowed())
		}
		if gw.createCalls != 0 {
			t.Fatalf("shifted caller dispatched create calls=%d", gw.createCalls)
		}
	})
}

func TestFlattenUsesInternalStartAndSharedAbsoluteTwoSecondDeadline(t *testing.T) {
	t.Run("future caller cannot extend deadline", func(t *testing.T) {
		c, _, gw, clock := controllerHarness(t)
		ctx := context.Background()
		fillAt := *clock
		saga, _ := c.PlanFill(ctx, Fill{At: fillAt, Quantity: 1, Trigger: 70000, ExpireDate: "2026-08-08"})
		active, _ := c.Register(ctx, saga.ID, "create", "2026-08-08", fillAt)
		gw.cancelDelay = 2100 * time.Millisecond
		if _, err := c.AuthorizeFlatten(ctx, active.ID, "cancel", clock.Add(time.Second), 1); !errors.Is(err, ErrMutationInDoubt) || c.EntryAllowed() {
			t.Fatalf("future start err=%v entry=%v", err, c.EntryAllowed())
		}
	})

	t.Run("past caller cannot shorten internal operation budget", func(t *testing.T) {
		c, _, gw, clock := controllerHarness(t)
		ctx := context.Background()
		fillAt := *clock
		saga, _ := c.PlanFill(ctx, Fill{At: fillAt, Quantity: 1, Trigger: 70000, ExpireDate: "2026-08-08"})
		active, _ := c.Register(ctx, saga.ID, "create", "2026-08-08", fillAt)
		gw.cancelDelay = 500 * time.Millisecond
		wallStart := time.Now()
		if _, err := c.AuthorizeFlatten(ctx, active.ID, "cancel", clock.Add(-24*time.Hour), 1); err != nil {
			t.Fatalf("past start err=%v", err)
		}
		if gw.cancelDL.IsZero() || gw.cancelDL.After(wallStart.Add(2100*time.Millisecond)) {
			t.Fatalf("cancel broker call deadline=%v wall_start=%v", gw.cancelDL, wallStart)
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
