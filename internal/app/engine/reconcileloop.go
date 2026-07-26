package engine

// reconcileloop.go is the reconciliation driver loop (change
// adopt-external-positions task 2.1; reconciliation "대사 구동 루프", design A6).
//
// # What was missing, and why it had to be built here
//
// internal/reconcile has had every piece of a reconciliation for two changes:
// the snapshot collector, the stabiliser, the comparer, the external-holding
// ingest, the quantity convergence and the mismatch tracker. What it has never
// had is a *caller*. `IngestExternalPositions`, `ConvergeQuantities` and
// `Tracker.Observe` had no production caller at all, which meant the projection
// was never folded to the account's holdings, `ADJUSTMENT_APPLIED` was
// unreachable, and the RECONCILE state machine only ever moved from the fill
// path's own refusals.
//
// Adoption cannot be built on top of that. It needs a periodic, stabilised view
// of what the account holds — so this file supplies the driver, and adoption is
// the last step of its cycle rather than a loop of its own.
//
// # One cycle
//
//	 1. collect a full snapshot (open orders paginated → holdings → balances)
//	 2. wait the stabilisation interval and collect a second one
//	 3. offer both to the Stabiliser; an unstable account ends the cycle silently
//	 4. compare the stable snapshot against the position projection
//	 5. fold external holdings in (Ingestor) and converge quantity mismatches
//	    (Converger)
//	 6. drive `Tracker.Observe` — the block/release state machine, and the
//	    precondition of the exit loop's RECONCILE confirmed-floor cap
//	 7. judge adoption candidates (adoption.go)
//
// Steps 5 and 6 are in that order because the design fixes it: the tracker
// observes the diff the comparison produced, and what the fold wrote shows up in
// the *next* cycle's re-read — which is exactly what makes an automatic release
// evidence-based rather than coincidental.
//
// # The rate budget (§0.4)
//
// Per 60-second period, in the steady state (open orders fitting one page, one
// currency):
//
//	2 collections × (1 open-order page + 1 holdings + 1 balance) = 6 requests
//	+ 1 batched price request, only when there is at least one adoption candidate
//
// So 6–7 requests per minute, **0.10–0.12 req/s**. The price read is one call for
// every candidate together and never one per symbol (design A6): a fan-out that
// scaled with the portfolio would make the budget a function of a number nobody
// bounded.
//
// The ceiling is the pagination bound. `Collector.MaxPages` (50) caps the
// open-order walk, so the worst case is 2 × (50 + 1 + currencies) + 1 requests
// per period — 103 with one currency, **1.7 req/s** — and reaching it means the
// account has more than fifty pages of resting orders, which is a state the
// engine cannot produce on its own.
//
// # The execution predicate
//
// `AutomationStatus.Verified`. The loop writes adjustments and, with adoption on,
// opens protection states that place sell orders without a human present — which
// is what the automation gate is the master switch for. Interlock clause 6 makes
// a verified gate unreachable in this build, so in production this refuses to be
// constructed until 2c. That is the design, not a gap.

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

// DefaultReconcilePeriod is how often the loop runs (design A6).
//
// Sixty seconds, which satisfies reconciliation's "재대사 최소 간격 30초" with a
// factor of two to spare — this loop *is* the periodic reconciliation that
// requirement describes.
const DefaultReconcilePeriod = 60 * time.Second

// DefaultAdoptionStaleness is how old the stable snapshot may be when a holding
// is judged for adoption (exit-policy: 관측 staleness 10초 이내).
const DefaultAdoptionStaleness = 10 * time.Second

// DefaultAdoptionPriceStaleness is how old the price observation may be at the
// moment the adoption transaction runs (exit-policy: staleness 15초 초과 시
// 편입을 연기한다).
//
// It is execgw's own QueryPrice freshness bound. A synthetic stop frozen from a
// stale price is a stop that may already be above the market, which makes the
// first observation a liquidation — the exact failure manage-forward exists to
// avoid.
const DefaultAdoptionPriceStaleness = 15 * time.Second

// ErrReconcileDriverUnavailable means the loop could not be assembled.
var ErrReconcileDriverUnavailable = errors.New(
	"engine: the reconciliation driver loop needs a verified automation gate, a snapshot collector, " +
		"the mismatch tracker and the journal that records its adjustments")

// ReconcileDriverOptions is the loop's wiring.
type ReconcileDriverOptions struct {
	// Journal supplies the position projection and records every adjustment and
	// adoption. Required.
	Journal *journal.Journal
	// Collector takes the account snapshot. Required.
	Collector *reconcile.Collector
	// Tracker is the block/release state machine this loop drives. Required.
	Tracker *reconcile.Tracker
	// Ingest folds external holdings in; Converge writes a quantity mismatch to
	// the account's own number. Both required — a loop with neither observes a
	// disagreement it can never settle.
	Ingest   *reconcile.Ingestor
	Converge *reconcile.Converger

	// Prices is the batched market read the adoption observation uses. Required
	// only when adoption is enabled.
	Prices PriceReader
	// Retrier runs the price read under the matrix and stamps its freshness.
	Retrier *execgw.Retrier

	// Alerts receives the operator alerts. Optional only so the loop can be unit
	// tested; an engine without one reconciles silently.
	Alerts ExitAlerter
	// Log receives the structured lines. Optional.
	Log *obs.Logger

	// AccountRef scopes everything. Required.
	AccountRef string
	// Clock drives the period, the stabilisation wait and every staleness
	// judgement. Defaults to the system clock.
	Clock clock.Clock
	// Period is the loop period. Zero takes DefaultReconcilePeriod.
	Period time.Duration
	// StabilisationInterval is the wait between a cycle's two collections. Zero
	// takes reconcile.DefaultStabilisationInterval.
	StabilisationInterval time.Duration
	// SnapshotStaleness bounds how old a stable snapshot may be when a holding is
	// judged for adoption. Zero takes DefaultAdoptionStaleness.
	SnapshotStaleness time.Duration
	// PriceStaleness bounds the price observation at the moment of the adoption
	// transaction. Zero takes DefaultAdoptionPriceStaleness.
	PriceStaleness time.Duration

	// Adoption is the feature's settings. The zero value is off, and off is a
	// complete configuration: the loop still reconciles and still alerts on
	// unmanaged holdings.
	Adoption config.Adoption
	// DefaultMarket is the venue a holding that names none is recorded under.
	DefaultMarket string
}

// ReconcileCycle is what one pass did. Every field is a count rather than a
// judgement, because the operator surface is the alerts and the log line.
type ReconcileCycle struct {
	// Collected is how many snapshots were taken (0, 1 or 2).
	Collected int
	// Stable reports that the two agreed and the cycle proceeded past step 3.
	Stable bool
	// Folded and Converged are what reached the projection.
	Folded    int
	Converged int
	// Closed counts managed positions found to have gone to zero outside the
	// engine.
	Closed int
	// Blocked and Released are the tracker's verdict.
	Blocked  int
	Released int
	// Unmanaged counts holdings confirmed to be outside exit management, and
	// therefore alerted on. It excludes transition states.
	Unmanaged int
	// Adopted counts holdings taken into exit management this cycle.
	Adopted int
	// Deferred counts adoption candidates held back — a stale observation, a
	// price the read did not answer for, or a refused transaction.
	Deferred int
	// Err is the first failure. A cycle that fails is retried next period; the
	// loop does not stop.
	Err error
}

// ReconcileDriver is the loop.
type ReconcileDriver struct {
	opts ReconcileDriverOptions
	clk  clock.Clock

	// stabiliser decides when the account has stopped moving. It lives for the
	// whole loop rather than per cycle so that a period whose two collections
	// disagree cannot be rescued by looking twice more.
	stabiliser reconcile.Stabiliser

	// ingest is a copy of the injected Ingestor with its alert routed through
	// this loop, so that a fold followed by an adoption in the same cycle
	// produces the adoption event and not both (design A4).
	ingest reconcile.Ingestor

	// unmanaged latches the "confirmed outside exit management" alert per
	// position, so a holding nobody adopts is reported once rather than every
	// minute for the rest of the day.
	unmanaged map[string]bool
	// grown latches the external-increase alert per position.
	grown map[string]bool
}

// NewReconcileDriver validates the wiring.
func NewReconcileDriver(opts ReconcileDriverOptions) (*ReconcileDriver, error) {
	switch {
	case opts.Journal == nil:
		return nil, fmt.Errorf("%w: no journal", ErrReconcileDriverUnavailable)
	case opts.Collector == nil:
		return nil, fmt.Errorf("%w: no snapshot collector", ErrReconcileDriverUnavailable)
	case opts.Tracker == nil:
		return nil, fmt.Errorf("%w: no mismatch tracker", ErrReconcileDriverUnavailable)
	case opts.Ingest == nil || opts.Converge == nil:
		return nil, fmt.Errorf("%w: no adjustment writers", ErrReconcileDriverUnavailable)
	case strings.TrimSpace(opts.AccountRef) == "":
		return nil, fmt.Errorf("%w: the loop is scoped to one account; none was named",
			ErrReconcileDriverUnavailable)
	case opts.Adoption.Enabled && opts.Prices == nil:
		return nil, fmt.Errorf("%w: adoption is on and there is no price read; the synthetic t0 is "+
			"an observation and cannot be invented", ErrReconcileDriverUnavailable)
	}

	d := &ReconcileDriver{
		opts:      opts,
		clk:       opts.Clock,
		unmanaged: map[string]bool{},
		grown:     map[string]bool{},
	}
	if d.clk == nil {
		d.clk = clock.System()
	}
	d.stabiliser = reconcile.Stabiliser{MinInterval: opts.StabilisationInterval}
	// The fold's alert is routed through this loop rather than raised inside the
	// ingest, because whether a newly folded holding is "unmanaged" is not known
	// until the adoption judgement at the end of the cycle. Copying the struct
	// leaves the caller's Ingestor untouched; it holds no state.
	d.ingest = *opts.Ingest
	d.ingest.Alert = nil
	return d, nil
}

// ReconcileDriver assembles the loop for this engine.
//
// It refuses on an unverified automation gate, for the reason the exit observer
// does: the loop writes to the ledger and, with adoption on, starts protecting
// positions with real sell orders. An engine whose master switch is off must not
// be doing either unattended.
func (c *Context) ReconcileDriver(opts ReconcileDriverOptions) (*ReconcileDriver, error) {
	if c == nil {
		return nil, ErrReconcileDriverUnavailable
	}
	if !c.Automation.Verified {
		return nil, fmt.Errorf("%w: the automation gate is not verified", ErrReconcileDriverUnavailable)
	}
	opts.Journal = c.Journal
	opts.Tracker = c.Reconcile
	opts.Ingest = c.Ingest
	opts.Converge = c.Converge
	opts.Retrier = c.Retrier
	opts.AccountRef = c.AccountRef
	opts.Adoption = c.Config.Engine.Adoption
	if opts.Prices == nil {
		opts.Prices = c.Official
	}
	if opts.Alerts == nil && c.Notifier != nil {
		opts.Alerts = c.Notifier
	}
	if opts.Log == nil {
		opts.Log = c.Log
	}
	return NewReconcileDriver(opts)
}

// Run drives the loop until the context ends.
//
// A failed cycle is logged and the loop continues. The alternative — returning —
// would make one unreadable snapshot stop the engine reconciling for the rest of
// the day, and an engine that has stopped comparing itself against the account
// is the state reconciliation exists to prevent.
func (d *ReconcileDriver) Run(ctx context.Context) error {
	period := d.opts.Period
	if period <= 0 {
		period = DefaultReconcilePeriod
	}
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		cycle := d.RunOnce(ctx)
		d.logCycle(cycle)
		if err := d.clk.Sleep(ctx, period); err != nil {
			return err
		}
	}
}

// RunOnce performs one cycle.
func (d *ReconcileDriver) RunOnce(ctx context.Context) ReconcileCycle {
	var cycle ReconcileCycle

	snapshot, ok := d.stabilise(ctx, &cycle)
	if !ok {
		return cycle
	}
	cycle.Stable = true

	local, err := reconcile.LocalStateFromJournal(ctx, d.opts.Journal, d.opts.AccountRef)
	if err != nil {
		cycle.Err = fmt.Errorf("engine: reading the local state for reconciliation: %w", err)
		return cycle
	}
	diff := reconcile.Comparer{}.Compare(snapshot, local)

	// The fold and the convergence are the two writers. Their failures are
	// recorded and do not stop the cycle: the tracker still has to see the diff,
	// because a disagreement nobody observed is a disagreement that blocks
	// nothing.
	folded, err := d.ingest.IngestExternalPositions(ctx, diff)
	if err != nil && cycle.Err == nil {
		cycle.Err = err
	}
	cycle.Folded = len(folded.Folded)

	converged, err := d.opts.Converge.ConvergeQuantities(ctx, diff)
	if err != nil && cycle.Err == nil {
		cycle.Err = err
	}
	cycle.Converged = len(converged.Converged)
	cycle.Closed = converged.Closed

	outcome, err := d.opts.Tracker.Observe(ctx, diff)
	if err != nil && cycle.Err == nil {
		cycle.Err = err
	}
	cycle.Blocked = len(outcome.Added)
	cycle.Released = len(outcome.Cleared)

	d.judgeHoldings(ctx, snapshot, &cycle)
	return cycle
}

// stabilise takes the cycle's two snapshots and reports the stable one.
//
// Two collections, at least the stabilisation interval apart, is the whole of
// reconciliation's "안정화는 최소 간격(기본 2초)을 둔 연속 2회 동일 스냅샷으로
// 판정한다". An account that is still moving ends the cycle here — silently,
// because "the owner is trading right now" is a transition state and not a
// finding (design A4).
func (d *ReconcileDriver) stabilise(ctx context.Context, cycle *ReconcileCycle) (reconcile.Snapshot, bool) {
	interval := d.opts.StabilisationInterval
	if interval <= 0 {
		interval = reconcile.DefaultStabilisationInterval
	}

	first, err := d.opts.Collector.Collect(ctx)
	if err != nil {
		cycle.Err = err
		return reconcile.Snapshot{}, false
	}
	cycle.Collected++
	d.stabiliser.Offer(first)

	if err := d.clk.Sleep(ctx, interval); err != nil {
		cycle.Err = err
		return reconcile.Snapshot{}, false
	}

	second, err := d.opts.Collector.Collect(ctx)
	if err != nil {
		cycle.Err = err
		return reconcile.Snapshot{}, false
	}
	cycle.Collected++
	result := d.stabiliser.Offer(second)
	if !result.Stable {
		// Deliberately not reset: a streak that survives into the next period is
		// evidence, and throwing it away would make a busy account permanently
		// unadoptable rather than merely slow to converge.
		return reconcile.Snapshot{}, false
	}
	// Acted on, so the next cycle starts from scratch rather than inheriting its
	// own corroboration (Stabiliser.Reset's contract).
	d.stabiliser.Reset()
	return second, true
}

func (d *ReconcileDriver) logCycle(cycle ReconcileCycle) {
	if d.opts.Log == nil {
		return
	}
	if cycle.Err != nil {
		d.opts.Log.Error(obs.EventReconcileMismatch, cycle.Err,
			obs.FieldAccount, d.opts.AccountRef,
			obs.FieldDetail, "the reconciliation cycle failed; it will be retried next period")
		return
	}
	d.opts.Log.Event(obs.EventReconcileClean,
		obs.FieldAccount, d.opts.AccountRef,
		"stable", cycle.Stable,
		"folded", cycle.Folded,
		"converged", cycle.Converged,
		"blocked", cycle.Blocked,
		"released", cycle.Released,
		"unmanaged", cycle.Unmanaged,
		"adopted", cycle.Adopted,
		"deferred", cycle.Deferred)
}

func (d *ReconcileDriver) alert(ctx context.Context, e obs.Event) {
	if d.opts.Alerts == nil {
		return
	}
	if err := d.opts.Alerts.Notify(ctx, e); err != nil && d.opts.Log != nil {
		d.opts.Log.Error(obs.EventAlertUndelivered, err,
			obs.FieldAccount, d.opts.AccountRef, "alert", string(e.Type))
	}
}
