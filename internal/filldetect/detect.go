// Package filldetect is the engine's authority on what has actually filled.
//
// # Why polling is the authority and SSE is not
//
// Toss publishes fills two ways. The SSE channel is fast but is a browser-session
// feature: it needs WTS cookies, it can be dropped by the server at any time, and
// its frames carry no state — only "something about your orders changed". The
// official Open API is slower but is the same source the account itself is
// computed from, and it keeps working when the browser session is long gone.
//
// So the loop here polls the official API and treats it as the truth
// (fill-detection: "폴링이 체결 감지의 권위"), and the SSE consumer next door
// (hints.go) is allowed to do exactly one thing: make the next poll happen sooner.
//
// # The three reads, every cycle
//
//  1. the open-order list, walked to its last page
//  2. every order we still track that is *not* in that list, read by id
//  3. the account's holdings
//
// (2) is the one that is easy to leave out and expensive to miss: a fully filled
// order leaves the open list, so a poller that only reads the open list watches
// orders vanish and cannot tell a fill from a cancellation. (1) has to be walked
// to the end for the same reason — an order on page two that is reported as absent
// is a live position the engine does not know it has.
//
// # What "detected" means
//
// A fill is detected when it is *durably committed locally*, not when it is seen.
// That is also where the SLO measures to: from the instant the broker made the
// fill observable to the instant the local commit returned (fill-detection: the
// measurement point). Everything in between — the poll interval, the retry budget,
// the journal fsync — is inside the number, which is the only way the number means
// anything operationally.
package filldetect

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/brokerstate"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
)

// OrderReader reads one order's raw JSON by id.
//
// Raw rather than domain.Order: the derivation needs canceledAt and the SLO needs
// execution.filledAt, and the upstream adapter drops both (issues.md 2026-07-26).
// execgw.OfficialOrders satisfies this.
type OrderReader interface {
	OrderRaw(ctx context.Context, orderID string) (json.RawMessage, error)
}

// PositionsReader sweeps the account's holdings.
type PositionsReader interface {
	Positions(ctx context.Context) ([]domain.Position, error)
}

// BuyingPowerReader reads the cash buying power of one currency.
//
// It is optional and, when wired, is read on its own slower cadence rather than
// every cycle. Cash moves only on an order, a fill or a transfer, so polling it at
// the fill-detection rate would spend the rate-limit budget (§0.4) on a number
// that barely changes — the reconciliation snapshot (task 3.4) is its natural
// owner. Wiring it here as well is what lets a deployment without the reconciler
// still satisfy the spec's "잔고·매수가능금액" read.
type BuyingPowerReader interface {
	BuyingPower(ctx context.Context, currency string) (float64, error)
}

// TrackedOrder is an order the detector must keep reading even after it has left
// the open list.
type TrackedOrder struct {
	OrderID    string
	AccountRef string
	Symbol     string
	Market     string
	TradingDay string
	Side       string
	// Lineage is what the journal knows about this order's replacement chain. It
	// is what stops an amended order being reported as cancelled.
	Lineage brokerstate.Lineage
}

// TrackedSource names those orders. The journal-backed implementation is
// journal.Journal.TrackedFillOrders (task 3.2).
type TrackedSource interface {
	// SelectedAccountRef is the detector instance's account scope. The official
	// order payload does not repeat it, so it has to enter the canonical identity
	// from the adapter that selected the account.
	SelectedAccountRef() string
	TrackedOrders(ctx context.Context) ([]TrackedOrder, error)
}

// Ledger applies a cumulative snapshot durably and idempotently (task 3.2).
type Ledger interface {
	Apply(ctx context.Context, snap Snapshot) (Applied, error)
}

// Config tunes the loop. Zero fields take the conservative defaults.
type Config struct {
	// PollInterval is the wait between cycles in Run.
	PollInterval time.Duration
	// MaxPages bounds one open-list walk (cursor-loop defence).
	MaxPages int
	// BalanceInterval is the minimum wait between buying-power reads.
	BalanceInterval time.Duration
	// Currencies are the buying-power currencies to read. Empty means every
	// currency the wired reader is asked for by the engine, which is none by
	// default — the engine names them.
	Currencies []string
	// SLO is the freshness objective.
	SLO SLO
}

// DefaultConfig is the provisional default. The poll interval is the 3s the retry
// matrix budgets for (retry-matrix.md §3); the real numbers come from
// verify-execution-capability.
func DefaultConfig() Config {
	return Config{
		PollInterval:    3 * time.Second,
		MaxPages:        50,
		BalanceInterval: 15 * time.Second,
		SLO:             DefaultSLO(),
	}
}

func (c Config) withDefaults() Config {
	d := DefaultConfig()
	if c.PollInterval <= 0 {
		c.PollInterval = d.PollInterval
	}
	if c.MaxPages <= 0 {
		c.MaxPages = d.MaxPages
	}
	if c.BalanceInterval <= 0 {
		c.BalanceInterval = d.BalanceInterval
	}
	c.SLO = c.SLO.withDefaults()
	return c
}

// Detector is the polling loop.
type Detector struct {
	// Orders walks the broker's open-order list. Required.
	Orders execgw.OrderPager
	// Order reads a single order by id. Required.
	Order OrderReader
	// Positions sweeps the account. Required.
	Positions PositionsReader
	// Balance reads buying power on its own slower cadence. Optional.
	Balance BuyingPowerReader
	// Tracked names the orders that must be followed by id. Required.
	Tracked TrackedSource
	// Ledger applies the snapshots durably. Required.
	Ledger Ledger
	// Clock drives the interval and the SLO measurement. Defaults to
	// clock.System().
	Clock clock.Clock
	// Gate is stamped fresh by every successful cycle and blocked while the SLO
	// is violated past its grace period. Optional.
	Gate   *execgw.EntryGate
	Config Config

	// poll serialises cycles. Hints (task 3.3) can ask for a cycle at any time,
	// and two cycles interleaving would let an older snapshot overwrite a newer
	// one in the ledger.
	poll sync.Mutex

	once   sync.Once
	slo    *sloTracker
	outage *outageTracker

	mu sync.Mutex
	// lastPollAt is the previous cycle's start. It is the conservative
	// broker-visible instant for a fill the broker did not timestamp.
	lastPollAt time.Time
	// lastBalanceAt is when buying power was last read, for its slower cadence.
	lastBalanceAt time.Time
	blocked       bool
}

// Snapshot is one observation of a lineage node's cumulative fill state.
//
// There is no per-fill identifier in this API (design D4), so a "fill" is never an
// event we receive — it is the difference between two of these.
type Snapshot struct {
	OrderID    string
	AccountRef string
	Symbol     string
	Market     string
	TradingDay string
	Side       string
	// Derived is the priority-table verdict on the payload.
	Derived brokerstate.Derived
	// RawStatus, Canceled and CanceledAt are the derivation's inputs, kept on the
	// snapshot so a ledger can re-derive with lineage the poller did not have.
	RawStatus  string
	Canceled   bool
	CanceledAt *time.Time
	// Quantity is the ordered quantity.
	Quantity float64
	// FilledQuantity is the broker's cumulative filled quantity.
	FilledQuantity float64
	// AveragePrice is the broker's average fill price, kept as the decimal
	// string it arrived as. Averages are replaced wholesale, never accumulated.
	AveragePrice string
	// FilledAmount is `execution.filledAmount`, kept as the decimal string it
	// arrived as. It exists so an amount-only restatement at an unchanged
	// cumulative quantity is detectable at all (design D7).
	FilledAmount string
	// BrokerVisibleAt is when the broker made this fill state observable. It is
	// execution.filledAt when the payload carries one; the caller substitutes the
	// previous poll's instant when it does not.
	BrokerVisibleAt time.Time
	// HasBrokerTime reports whether BrokerVisibleAt came from the payload.
	HasBrokerTime bool
	// ObservedAt is when this cycle read the payload.
	ObservedAt time.Time
}

// Applied is the ledger's verdict on one snapshot.
type Applied struct {
	OrderID string
	// Delta is the newly filled quantity, always >= 0.
	Delta float64
	// Changed reports whether the snapshot moved local state at all.
	Changed bool
	// Corrected marks an EXECUTION_CORRECTION: the broker restated an execution
	// it had already reported, with no change in cumulative quantity.
	Corrected bool
	// FailClosed marks a snapshot that contradicts what was already observed.
	FailClosed bool
	Reason     execgw.ReasonCode
	Detail     string
	// CommittedAt is when the durable write returned. It is the far end of the
	// SLO measurement.
	CommittedAt time.Time
}

// Cycle is what one poll did.
type Cycle struct {
	StartedAt  time.Time
	FinishedAt time.Time
	// OpenOrders is how many open orders the completed walk returned.
	OpenOrders int
	// TrackedReads is how many single-order reads were needed for orders that
	// had left the open list.
	TrackedReads int
	// Positions is the size of the account sweep.
	Positions int
	// ReadBuyingPower reports whether this cycle also refreshed buying power.
	ReadBuyingPower bool
	// Applied is the ledger verdict per snapshot, in observation order.
	Applied []Applied
	// Fills counts the positive deltas.
	Fills int
	// Latencies are the SLO samples this cycle produced.
	Latencies []time.Duration
	// FailedClosed lists the order ids whose snapshot could not be reconciled.
	FailedClosed []string
}

// PollOnce runs one full cycle.
//
// It returns an error for anything that makes the picture incomplete — a failed
// read, an unparseable order — and deliberately does not stamp query freshness in
// that case, so the entry gate blocks on staleness without the detector having to
// decide that itself.
func (d *Detector) PollOnce(ctx context.Context) (Cycle, error) {
	d.poll.Lock()
	defer d.poll.Unlock()
	return d.pollLocked(ctx)
}

func (d *Detector) pollLocked(ctx context.Context) (Cycle, error) {
	d.ensure()
	if err := d.validate(); err != nil {
		return Cycle{}, err
	}
	cfg := d.Config.withDefaults()
	clk := d.clk()
	started := clk.Now()
	cycle := Cycle{StartedAt: started}

	snaps, err := d.collect(ctx, cfg, started, &cycle)
	if err != nil {
		d.outage.failure(clk.Now(), err)
		cycle.FinishedAt = clk.Now()
		return cycle, err
	}

	positions, err := d.Positions.Positions(ctx)
	if err != nil {
		d.outage.failure(clk.Now(), err)
		cycle.FinishedAt = clk.Now()
		return cycle, fmt.Errorf("filldetect: sweeping the account: %w", err)
	}
	cycle.Positions = len(positions)

	readBalance, err := d.sweepBalance(ctx, cfg, started)
	if err != nil {
		d.outage.failure(clk.Now(), err)
		cycle.FinishedAt = clk.Now()
		return cycle, err
	}
	cycle.ReadBuyingPower = readBalance

	for _, snap := range snaps {
		applied, err := d.Ledger.Apply(ctx, snap)
		if err != nil {
			d.outage.failure(clk.Now(), err)
			cycle.FinishedAt = clk.Now()
			return cycle, fmt.Errorf("filldetect: committing the %s snapshot: %w", snap.OrderID, err)
		}
		cycle.Applied = append(cycle.Applied, applied)
		if applied.FailClosed {
			cycle.FailedClosed = append(cycle.FailedClosed, applied.OrderID)
			d.blockSymbol(snap, applied)
			continue
		}
		if applied.Delta > 0 {
			cycle.Fills++
			committed := applied.CommittedAt
			if committed.IsZero() {
				committed = clk.Now()
			}
			latency := committed.Sub(snap.BrokerVisibleAt)
			if latency < 0 {
				// A broker timestamp ahead of our clock is skew, not a negative
				// delay. Zero is the honest floor.
				latency = 0
			}
			cycle.Latencies = append(cycle.Latencies, latency)
			d.slo.observe(committed, latency)
		}
	}

	now := clk.Now()
	d.outage.success(now)
	d.stampFreshness(readBalance)
	d.evaluateSLO(now)

	d.mu.Lock()
	d.lastPollAt = started
	d.mu.Unlock()

	cycle.FinishedAt = now
	return cycle, nil
}

// collect performs the two order reads and turns them into snapshots.
//
// The order is fixed: the open list first (so the tracked set can skip anything it
// already covered), then the leftovers by id.
func (d *Detector) collect(ctx context.Context, cfg Config, started time.Time, cycle *Cycle) ([]Snapshot, error) {
	raws, err := execgw.ScanOrders(ctx, d.Orders, execgw.OrderQuery{Status: statusOpen}, cfg.MaxPages)
	if err != nil {
		return nil, fmt.Errorf("filldetect: walking the open-order list: %w", err)
	}
	cycle.OpenOrders = len(raws)

	tracked, err := d.Tracked.TrackedOrders(ctx)
	if err != nil {
		return nil, fmt.Errorf("filldetect: listing the tracked orders: %w", err)
	}
	accountRef := strings.TrimSpace(d.Tracked.SelectedAccountRef())
	if accountRef == "" {
		return nil, errors.New("filldetect: the tracked-order source has no selected account")
	}

	trackedByKey := make(map[canonicalOrderKey]TrackedOrder, len(tracked))
	trackedByID := make(map[string][]canonicalOrderKey, len(tracked))
	trackedKeys := make([]canonicalOrderKey, 0, len(tracked))
	for _, order := range tracked {
		key, keyErr := trackedOrderKey(order, accountRef)
		if keyErr != nil {
			return nil, keyErr
		}
		if _, duplicate := trackedByKey[key]; duplicate {
			return nil, fmt.Errorf("filldetect: duplicate tracked order in canonical scope %s", key)
		}
		trackedByKey[key] = order
		trackedByID[key.OrderID] = append(trackedByID[key.OrderID], key)
		trackedKeys = append(trackedKeys, key)
	}

	fallback := d.brokerVisibleFallback(started)
	seen := make(map[canonicalOrderKey]bool, len(raws))
	snaps := make([]Snapshot, 0, len(raws)+len(tracked))

	for _, raw := range raws {
		snap, err := parseSnapshot(raw, started, fallback)
		if err != nil {
			// An unreadable record is an unknown, not an absence. Reporting the
			// list as complete without it would be a lie about a live account.
			return nil, fmt.Errorf("filldetect: reading an order from the open list: %w", err)
		}
		snap.AccountRef = accountRef
		key, complete := snapshotOrderKey(snap)
		lineage := brokerstate.Lineage{}
		if complete {
			if order, locallyTracked := trackedByKey[key]; locallyTracked {
				lineage = order.Lineage
				seen[key] = true
			}
		} else if len(trackedByID[snap.OrderID]) > 0 {
			return nil, fmt.Errorf(
				"filldetect: open order %q has incomplete canonical scope and could match a locally tracked order",
				snap.OrderID)
		}
		snap.Derived = brokerstate.Derive(viewOf(snap, lineage))
		snaps = append(snaps, snap)
	}

	for _, key := range trackedKeys {
		if seen[key] {
			continue
		}
		order := trackedByKey[key]
		raw, err := d.Order.OrderRaw(ctx, order.OrderID)
		if err != nil {
			return nil, fmt.Errorf("filldetect: reading order %s: %w", order.OrderID, err)
		}
		cycle.TrackedReads++
		snap, err := parseSnapshot(raw, started, fallback)
		if err != nil {
			return nil, fmt.Errorf("filldetect: reading order %s: %w", order.OrderID, err)
		}
		// Trimmed only to judge emptiness: the payload's identifier is kept
		// verbatim when there is one, and the tracked id fills in when there is
		// not (order-execution "브로커 식별자의 opaque 취급").
		if strings.TrimSpace(snap.OrderID) == "" {
			snap.OrderID = order.OrderID
		}
		snap.AccountRef = accountRef
		actual, complete := snapshotOrderKey(snap)
		if !complete {
			return nil, fmt.Errorf(
				"filldetect: tracked order %q response has incomplete canonical scope", order.OrderID)
		}
		if actual != key {
			return nil, fmt.Errorf(
				"filldetect: tracked order %q response scope %s does not match requested scope %s",
				order.OrderID, actual, key)
		}
		snap.Derived = brokerstate.Derive(viewOf(snap, order.Lineage))
		snaps = append(snaps, snap)
	}
	return snaps, nil
}

type canonicalOrderKey struct {
	AccountRef string
	Market     string
	TradingDay string
	Symbol     string
	Side       string
	OrderID    string
}

func (k canonicalOrderKey) String() string {
	return fmt.Sprintf("account=%q market=%q day=%q symbol=%q side=%q order=%q",
		k.AccountRef, k.Market, k.TradingDay, k.Symbol, k.Side, k.OrderID)
}

func trackedOrderKey(order TrackedOrder, selectedAccount string) (canonicalOrderKey, error) {
	key := canonicalOrderKey{
		AccountRef: strings.TrimSpace(order.AccountRef),
		Market:     strings.ToLower(strings.TrimSpace(order.Market)),
		TradingDay: strings.TrimSpace(order.TradingDay),
		Symbol:     strings.ToUpper(strings.TrimSpace(order.Symbol)),
		Side:       strings.ToUpper(strings.TrimSpace(order.Side)),
		OrderID:    order.OrderID,
	}
	if strings.TrimSpace(key.OrderID) == "" || key.AccountRef == "" || key.Market == "" ||
		key.TradingDay == "" || key.Symbol == "" || key.Side == "" {
		return canonicalOrderKey{}, fmt.Errorf(
			"filldetect: tracked order %q has incomplete canonical scope", order.OrderID)
	}
	if key.AccountRef != selectedAccount {
		return canonicalOrderKey{}, fmt.Errorf(
			"filldetect: tracked order %q belongs to account %q, selected account is %q",
			order.OrderID, key.AccountRef, selectedAccount)
	}
	return key, nil
}

func snapshotOrderKey(snap Snapshot) (canonicalOrderKey, bool) {
	key := canonicalOrderKey{
		AccountRef: strings.TrimSpace(snap.AccountRef),
		Market:     strings.ToLower(strings.TrimSpace(snap.Market)),
		TradingDay: strings.TrimSpace(snap.TradingDay),
		Symbol:     strings.ToUpper(strings.TrimSpace(snap.Symbol)),
		Side:       strings.ToUpper(strings.TrimSpace(snap.Side)),
		OrderID:    snap.OrderID,
	}
	complete := strings.TrimSpace(key.OrderID) != "" && key.AccountRef != "" && key.Market != "" &&
		key.TradingDay != "" && key.Symbol != "" && key.Side != ""
	return key, complete
}

// Run polls until ctx is done.
//
// A failed cycle is logged into the outage state and then retried on the next
// tick: an outage is a reason to keep looking, not a reason to stop watching a
// live account.
func (d *Detector) Run(ctx context.Context) error {
	cfg := d.Config.withDefaults()
	clk := d.clk()
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		_, _ = d.PollOnce(ctx)
		if err := clk.Sleep(ctx, cfg.PollInterval); err != nil {
			return err
		}
	}
}

// Health is the current detection health, for status output and alerting.
type Health struct {
	Outage Outage
	SLO    SLOStatus
	// EntryBlocked reports whether the detector currently holds the entry gate
	// shut.
	EntryBlocked bool
}

// Health reports the current state without polling.
func (d *Detector) Health() Health {
	d.ensure()
	d.mu.Lock()
	blocked := d.blocked
	d.mu.Unlock()
	return Health{
		Outage:       d.outage.status(),
		SLO:          d.slo.status(d.clk().Now()),
		EntryBlocked: blocked,
	}
}

// --- internals --------------------------------------------------------------

const statusOpen = "OPEN"

func (d *Detector) clk() clock.Clock {
	if d.Clock == nil {
		return clock.System()
	}
	return d.Clock
}

func (d *Detector) ensure() {
	d.once.Do(func() {
		cfg := d.Config.withDefaults()
		d.slo = newSLOTracker(cfg.SLO)
		d.outage = &outageTracker{}
	})
}

func (d *Detector) validate() error {
	switch {
	case d.Orders == nil:
		return errors.New("filldetect: an order pager is required — the open list is the authority")
	case d.Order == nil:
		return errors.New("filldetect: a single-order reader is required — a filled order leaves the open list")
	case d.Positions == nil:
		return errors.New("filldetect: a positions reader is required")
	case d.Tracked == nil:
		return errors.New("filldetect: a tracked-order source is required")
	case d.Ledger == nil:
		return errors.New("filldetect: a ledger is required — an uncommitted fill is an undetected fill")
	}
	return nil
}

// brokerVisibleFallback is the instant used for a fill the broker did not
// timestamp: the previous cycle's start, i.e. the earliest moment the fill could
// have become visible without us seeing it. That makes the measured latency an
// upper bound, which is the safe direction for a threshold that blocks entries.
func (d *Detector) brokerVisibleFallback(started time.Time) time.Time {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.lastPollAt.IsZero() {
		return started
	}
	return d.lastPollAt
}

// sweepBalance reads buying power when its own interval has elapsed, reporting
// whether it did.
func (d *Detector) sweepBalance(ctx context.Context, cfg Config, started time.Time) (bool, error) {
	if d.Balance == nil || len(cfg.Currencies) == 0 {
		return false, nil
	}
	d.mu.Lock()
	last := d.lastBalanceAt
	d.mu.Unlock()
	if !last.IsZero() && started.Sub(last) < cfg.BalanceInterval {
		return false, nil
	}
	for _, currency := range cfg.Currencies {
		if _, err := d.Balance.BuyingPower(ctx, currency); err != nil {
			return false, fmt.Errorf("filldetect: reading the %s buying power: %w", currency, err)
		}
	}
	d.mu.Lock()
	d.lastBalanceAt = started
	d.mu.Unlock()
	return true, nil
}

// stampFreshness tells the entry gate the reads it gates on are current. Only a
// complete cycle does this: a partial read is not a fresh read.
func (d *Detector) stampFreshness(readBalance bool) {
	if d.Gate == nil {
		return
	}
	d.Gate.RecordSuccess(execgw.QueryOpenOrders)
	d.Gate.RecordSuccess(execgw.QueryHoldings)
	if readBalance {
		d.Gate.RecordSuccess(execgw.QueryBuyingPower)
	}
}

// evaluateSLO latches or releases the entry block.
//
// Unlike the auth latch, this one clears itself. The condition it describes —
// "detection is behind" — is disproved by the next fast fill, and keeping an
// engine shut after it has recovered would trade one failure for another.
func (d *Detector) evaluateSLO(now time.Time) {
	status := d.slo.status(now)
	block := status.Violated && status.ViolatedFor >= d.Config.withDefaults().SLO.Grace

	d.mu.Lock()
	was := d.blocked
	d.blocked = block
	d.mu.Unlock()

	if d.Gate == nil {
		return
	}
	switch {
	case block && !was:
		d.Gate.Block(execgw.ReasonFillDetectionSLO, fmt.Sprintf(
			"fill detection p%d is %s against a %s objective and has been for %s",
			int(status.Percentile*100), status.Observed.Round(time.Millisecond),
			status.Target, status.ViolatedFor.Round(time.Second)))
	case !block && was:
		d.Gate.Clear(execgw.ReasonFillDetectionSLO)
	}
}

// blockSymbol latches the gate for a snapshot the ledger refused.
//
// The latch is account-wide because the entry gate has no symbol dimension yet
// (the gateway's per-symbol check reads the journal, not this gate). Blocking more
// than the spec's symbol scope is the conservative direction; narrowing it is
// wiring work for task 4.2, and is recorded in issues.md.
func (d *Detector) blockSymbol(snap Snapshot, applied Applied) {
	if d.Gate == nil {
		return
	}
	reason := applied.Reason
	if reason == "" {
		reason = execgw.ReasonBrokerStateUnknown
	}
	detail := applied.Detail
	if detail == "" {
		detail = fmt.Sprintf("order %s on %s could not be reconciled with what was already observed",
			snap.OrderID, snap.Symbol)
	}
	d.Gate.Block(reason, detail)
}

func viewOf(s Snapshot, lineage brokerstate.Lineage) brokerstate.OrderView {
	return brokerstate.OrderView{
		OrderID:        s.OrderID,
		RawStatus:      s.RawStatus,
		Canceled:       s.Canceled,
		CanceledAt:     s.CanceledAt,
		Quantity:       s.Quantity,
		FilledQuantity: s.FilledQuantity,
		Lineage:        lineage,
	}
}
