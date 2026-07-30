package engine

// exitwiring.go builds the four production seams the exit observation loop asks
// for and nothing else in the engine had a constructor for yet (task 7.4/7.5
// succession; issues.md, task 3.2's producer table).
//
// Three of them are adapters that exist so internal/app/engine takes the
// dependency rather than the package that would otherwise have to:
//
//	notifierAlerter   obs.Notifier → reconcile.Alerter (issues.md, task 6.3)
//	reconcileFloor    the RECONCILE confirmed floor, per symbol (task 7.5)
//	OfficialReads     the price read the loop makes
//
// The fourth is the Retrier, and it is the one with teeth: it is where a 401/403
// becomes a durable ENTRY_BLOCKED rather than an in-memory latch, and it is what
// stamps the price query's freshness. Before this file the engine constructed
// neither it nor the Notifier at all — the escalation producers task 3.2 wired
// had no caller, which issues.md recorded as "엔진 프로필에는 아직 Retrier도
// Notifier도 구성하는 곳이 없다".

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// newRetrier builds the query retrier with every escalation field populated.
//
// The three fields below the gate are the ones task 3.2 left for the wiring, and
// leaving any of them empty silently downgrades the producer:
//
//	Escalate/AccountRef  a 401/403 latches this process and nothing else, so a
//	                     restart — the first thing an operator does when they see
//	                     a credential error — lifts the block
//	Announcer            the transition happens and nobody is told
func newRetrier(j *journal.Journal, gate *execgw.EntryGate, accountRef string,
	notifier *obs.Notifier, retryAfter execgw.RetryAfterSource, clk clock.Clock) *execgw.Retrier {
	return &execgw.Retrier{
		Policy:     execgw.DefaultRetryPolicy(),
		Clock:      clk,
		Gate:       gate,
		RetryAfter: retryAfter,
		Escalate:   j,
		AccountRef: accountRef,
		Announcer:  notifier,
	}
}

// newNotifier builds the alert path.
//
// # What an absent publisher means
//
// Publisher is nil unless the caller supplies one, and a nil publisher makes
// every *critical* alert undeliverable: the outbox row is written and stays
// PENDING, the entry gate latches, and sustained failure escalates to
// ENTRY_BLOCKED. That is the specified direction rather than an accident
// (risk-management: critical 알림 outbox 전달 실패 지속 → ENTRY_BLOCKED) — an
// engine that cannot tell its operator that a position is unprotected should not
// be opening new ones. It costs nothing today because interlock clause 6 already
// blocks the gate, and it is recorded in issues.md so the first operator to
// configure a transport knows why.
func newNotifier(j *journal.Journal, gate *execgw.EntryGate, accountRef string,
	log *obs.Logger, publisher obs.Publisher, clk clock.Clock) *obs.Notifier {
	return &obs.Notifier{
		Log:        log,
		Publisher:  publisher,
		Journal:    j,
		Gate:       gate,
		AccountRef: accountRef,
		Clock:      clk,
	}
}

// notifierAlerter adapts the Notifier to reconcile.Alerter.
//
// internal/reconcile declares the interface rather than importing internal/obs
// so the reconciliation loop does not choose the grade or the transport; this is
// where those are chosen. Normal, not critical: a person trading their own
// account by hand is not a malfunction, and a critical grade would make an
// undelivered alert about it block entries (issues.md, task 6.3).
type notifierAlerter struct{ notifier *obs.Notifier }

func (a notifierAlerter) ExternalPositionFound(ctx context.Context,
	alert reconcile.ExternalPositionAlert) error {
	if a.notifier == nil {
		return nil
	}
	return a.notifier.Notify(ctx, obs.Event{
		Type: obs.EventExitPositionUnmanaged,
		Key: string(obs.EventExitPositionUnmanaged) + "|" +
			strings.TrimSpace(alert.PositionID),
		Title: alert.Symbol + " was folded in from the account and will not be managed",
		Body: "the account holds it and no local instance explains it, so it was recorded with no " +
			"entry decision: there is no stop to build a baseline from and the exit policy leaves it alone",
		Fields: map[string]any{
			obs.FieldAccount:  alert.AccountRef,
			obs.FieldSymbol:   alert.Symbol,
			obs.FieldQuantity: alert.Quantity,
			"market":          alert.Market,
			"position_id":     alert.PositionID,
			"broker_as_of":    alert.BrokerAsOf,
			"exit_eligible":   alert.ExitEligible,
		},
	})
}

// ManagedPositionClosedExternally is the second alert the reconciliation loop
// raises (change adopt-external-positions, design A7).
//
// Critical, unlike the fold above, and the difference is what the operator has
// to do about it. A holding the engine will not manage is a fact about the
// account; a position the engine *was* managing and no longer holds is the
// engine's own protection ending in a way it did not choose — and the outcome
// row it would normally freeze does not exist, so the trade is missing from
// every aggregate until somebody accounts for it by hand.
func (a notifierAlerter) ManagedPositionClosedExternally(ctx context.Context,
	alert reconcile.ManagedCloseAlert) error {
	if a.notifier == nil {
		return nil
	}
	how := "the engine opened it"
	if alert.Adopted {
		how = "the engine had adopted it"
	}
	return a.notifier.Notify(ctx, obs.Event{
		Type: obs.EventExitPositionClosedExternally,
		Key: string(obs.EventExitPositionClosedExternally) + "|" +
			strings.TrimSpace(alert.PositionID),
		Title: alert.Symbol + " was closed outside the engine while the exit policy was managing it",
		Body: "the account no longer holds it and no local fill explains that, so the exit state was " +
			"completed with an ADJUSTMENT_CLOSED event. No trade outcome was frozen: the sell happened " +
			"where the engine cannot see it, so there is no sell leg to price and recording one would " +
			"report the position as a total loss",
		Fields: map[string]any{
			obs.FieldAccount:  alert.AccountRef,
			obs.FieldSymbol:   alert.Symbol,
			obs.FieldQuantity: alert.PrevQuantity,
			"market":          alert.Market,
			"position_id":     alert.PositionID,
			"broker_as_of":    alert.BrokerAsOf,
			"adopted":         alert.Adopted,
			obs.FieldDetail:   how,
		},
	})
}

// reconcileFloor computes the RECONCILE confirmed floor for one symbol.
//
// # When it applies
//
// Only while the tracker holds a block covering the symbol. Outside RECONCILE
// there is no disagreement to bound a liquidation by, and asking anyway would
// put two broker round trips in front of every exit — including the ones §0.3
// says must not be delayed.
//
// # Why it reads the broker again rather than reusing the reconcile snapshot
//
// Because the floor's own definition is about freshness (riskcalc: 신선 =
// 스냅샷 나이 ≤ staleness 한계 · 스냅샷 부재 시 0), and the snapshot the
// comparison ran on is by then up to a reconcile interval old. A floor computed
// from a stale read is a floor of zero by rule, and the honest way to get a
// non-zero one is to read again. Both reads go through the Retrier so their
// failures are classified and budgeted like every other query.
//
// Manual flatten is not a caller of any of this, and internal/flatten still
// reads its own sellable quantity per symbol immediately before each order
// (riskcalc/confirmed_floor.go, §0.3).
type reconcileFloor struct {
	official OfficialReads
	tracker  *reconcile.Tracker
	journal  *journal.Journal
	retrier  *execgw.Retrier
	clk      clock.Clock
	account  string
}

func (f *reconcileFloor) ConfirmedFloor(ctx context.Context, market, symbol string) (
	riskcalc.ConfirmedFloor, bool, error) {
	if f == nil || f.tracker == nil || f.official == nil {
		return riskcalc.ConfirmedFloor{}, false, nil
	}
	if f.tracker.EntryAllowed(market, symbol) == nil && !f.tracker.Permanent() {
		// No block covers this symbol, so the engine and the account are not in
		// disagreement about it and there is nothing to bound.
		return riskcalc.ConfirmedFloor{}, false, nil
	}

	var (
		holdings *riskcalc.QuantitySnapshot
		sellable *riskcalc.QuantitySnapshot
	)
	if err := f.retrier.Query(ctx, execgw.QueryHoldings, func(ctx context.Context) error {
		positions, err := f.official.Holdings(ctx, symbol)
		if err != nil {
			return err
		}
		for _, p := range positions {
			if !strings.EqualFold(strings.TrimSpace(p.Symbol), strings.TrimSpace(symbol)) {
				continue
			}
			holdings = &riskcalc.QuantitySnapshot{
				Quantity: decimalOf(p.Quantity), AsOf: f.clk.Now(),
			}
		}
		if holdings == nil {
			// The account reports no holding of this symbol. That is a quantity,
			// not a missing read, and it produces a floor of zero honestly.
			holdings = &riskcalc.QuantitySnapshot{Quantity: "0", AsOf: f.clk.Now()}
		}
		return nil
	}); err != nil {
		return riskcalc.ConfirmedFloor{}, true, fmt.Errorf(
			"engine: reading the holding of %s for the confirmed floor: %w", symbol, err)
	}

	if err := f.retrier.Query(ctx, execgw.QueryHoldings, func(ctx context.Context) error {
		q, err := f.official.SellableQuantity(ctx, symbol)
		if err != nil {
			return err
		}
		sellable = &riskcalc.QuantitySnapshot{Quantity: decimalOf(q.Quantity), AsOf: f.clk.Now()}
		return nil
	}); err != nil {
		return riskcalc.ConfirmedFloor{}, true, fmt.Errorf(
			"engine: reading the sellable quantity of %s for the confirmed floor: %w", symbol, err)
	}

	local, err := f.localOpenSells(ctx, market, symbol)
	if err != nil {
		return riskcalc.ConfirmedFloor{}, true, err
	}
	floor, err := riskcalc.ConfirmedFloorQuantity(riskcalc.ConfirmedFloorInputs{
		Now:                   f.clk.Now(),
		Symbol:                symbol,
		Holdings:              holdings,
		Sellable:              sellable,
		LocalOpenSellQuantity: local,
	})
	if err != nil {
		return riskcalc.ConfirmedFloor{}, true, err
	}
	return floor, true, nil
}

// localOpenSells sums the resting local sells of one symbol.
//
// It is a subtrahend and never an addend, which is riskcalc's first conservative
// rule: quantity a live order has already committed is quantity a second order
// must not be authorised to sell again.
func (f *reconcileFloor) localOpenSells(ctx context.Context, market, symbol string) (string, error) {
	live, err := f.journal.LiveOrdersForSymbol(ctx, f.account, market, symbol)
	if err != nil {
		return "", fmt.Errorf("engine: reading the resting sells of %s: %w", symbol, err)
	}
	total := "0"
	for _, order := range live {
		if !strings.EqualFold(strings.TrimSpace(order.Side), "SELL") {
			continue
		}
		total, err = riskcalc.AddDecimal(total, orZeroQuantity(order.Quantity))
		if err != nil {
			return "", fmt.Errorf("engine: summing the resting sells of %s: %w", symbol, err)
		}
	}
	return total, nil
}

func orZeroQuantity(q string) string {
	if strings.TrimSpace(q) == "" {
		return "0"
	}
	return strings.TrimSpace(q)
}

// ErrExitObserverUnavailable means the exit observation loop could not be
// assembled from this engine's wiring.
//
// It is returned rather than swallowed because the two reasons are operationally
// different and both matter: the automation gate is off (so nothing may place an
// order unattended, which is §0.2 working), or the injected Guardian cannot issue
// reductions (which is a wiring bug).
var ErrExitObserverUnavailable = errors.New(
	"engine: the exit observation loop needs a verified automation gate and a Guardian that can issue " +
		"reduce-only exits")

// ExitObserver assembles the exit observation loop for this engine.
//
// # Why the gate has to be verified
//
// The loop places orders without a human in the loop, which is exactly what the
// automation gate is the master switch for (config.AutomationGate). Running it
// on a gate-off engine would be unattended order placement with the switch off.
// ProtectionReady is enforced at the gateway for exposure-raising mutations,
// not here at runtime startup. That keeps reduce-only exits available even while
// broker-resident protective execution is UNWIRED.
//
// # Deference
//
// SLO is left nil because this build constructs no fill detector: there is no
// production polling loop to defer to yet, and a deference source that always
// answers "not behind" would be a claim rather than a measurement. The ordering
// itself is fixed in exitloop.go and covered by its tests; wiring the real
// detector is the change that introduces one.
func (c *Context) ExitObserver(opts ExitObserverOptions) (*ExitObserver, error) {
	if c == nil {
		return nil, ErrExitObserverUnavailable
	}
	if !c.Automation.Verified {
		return nil, fmt.Errorf("%w: the automation gate is not verified", ErrExitObserverUnavailable)
	}
	issuer, ok := c.Guardian.(ReductionIssuer)
	if !ok {
		return nil, fmt.Errorf("%w: the injected Guardian does not issue reductions",
			ErrExitObserverUnavailable)
	}
	opts.Journal = c.Journal
	opts.Prices = c.Official
	opts.Retrier = c.Retrier
	opts.Issuer = issuer
	opts.Submit = c.Gateway
	opts.AccountRef = c.AccountRef
	opts.CommonPolicy = c.Config.Engine.ExitPolicy.CommonPolicy
	if opts.Alerts == nil && c.Notifier != nil {
		opts.Alerts = c.Notifier
	}
	if opts.Floor == nil {
		opts.Floor = c.exitFloor
	}
	return NewExitObserver(opts)
}

// --- analytics retention (task 8.1) -----------------------------------------------

// DefaultRetentionSweepInterval is how often the retention sweep runs.
//
// Daily. The horizon is 180 days, so the interval only decides how long a row
// outlives it, and a sweep that ran more often would be spending broker-free
// I/O to make a boundary that nobody measures slightly sharper.
const DefaultRetentionSweepInterval = 24 * time.Hour

// RunAnalyticsRetention prunes trade outcomes past their horizon until the
// context ends.
//
// # Everything about this function is the isolation requirement
//
// trade-analytics: 성과 집계와 보존 기간 정리는 주문 실행 경로에서 격리되어야
// 한다 (SHALL), and 분석 작업의 실패·지연이 … 운영 모드 강화를 유발해서는 안 된다
// (SHALL NOT). So:
//
//   - it is its own goroutine, and the order path never waits on it;
//   - a failed sweep is logged and the loop continues — it returns only the
//     context's error, so a caller cannot accidentally treat a pruning failure
//     as an engine failure;
//   - it raises no alert and touches no operating mode. The escalation triggers
//     are a closed enumeration (journal.AutomaticTriggers) and this is not one
//     of them; a retention failure that blocked entries would be an analytics
//     job deciding whether the engine may trade.
func (c *Context) RunAnalyticsRetention(ctx context.Context, clk clock.Clock,
	interval time.Duration) error {
	if interval <= 0 {
		interval = DefaultRetentionSweepInterval
	}
	if clk == nil {
		clk = clock.System()
	}
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		c.sweepAnalyticsRetention(ctx, clk)
		if err := clk.Sleep(ctx, interval); err != nil {
			return err
		}
	}
}

// sweepAnalyticsRetention performs one prune. It reports nothing: the count is
// a log line and the failure is a log line, and neither is a value the order
// path may branch on.
func (c *Context) sweepAnalyticsRetention(ctx context.Context, clk clock.Clock) {
	if c == nil || c.Journal == nil {
		return
	}
	pruned, err := c.Journal.PruneTradeOutcomes(ctx, clk.Now().Add(-journal.TradeOutcomeRetention))
	switch {
	case err != nil:
		// Logged, never escalated. The rows stay, the next sweep tries again, and
		// the account's ability to trade is not a function of whether a delete
		// statement succeeded.
		if c.Log != nil {
			c.Log.Error(obs.EventEngineHeartbeat, err,
				obs.FieldAccount, c.AccountRef,
				obs.FieldDetail, "the trade-outcome retention sweep failed; it will be retried")
		}
	case pruned > 0 && c.Log != nil:
		c.Log.Event(obs.EventEngineHeartbeat,
			obs.FieldAccount, c.AccountRef,
			"pruned_trade_outcomes", pruned,
			obs.FieldDetail, "trade outcomes past the 180-day horizon were removed")
	}
}
