package engine

// tracer.go is the tracer slice (add-core-domain task 8.2, design D8): the
// smallest complete pass through the engine's own machinery — one allowlisted
// symbol, a LIMIT order, the minimum quantity — from entry issuance to the exit
// policy closing the position.
//
// # What it is for, and what it is not
//
// It exists so that the first real trade this system ever makes is a trade whose
// every step has already been executed together, in order, against something.
// Here that something is an httptest broker (tracer_test.go); the *live* run is
// the verify track's, and it is not this change's to perform (D8: 실전 실행은
// verify 트랙).
//
// It is not a strategy and it takes no signal input. The entry it proposes is
// the caller's parameters, spelled out and bounded; nothing in this file decides
// what to buy.
//
// # Why the production entry still refuses while protection is unwired
//
// Every order it places goes through execgw.Gateway with a GuardianDecision, and
// the engine profile only holds a Guardian once the startup interlock verifies
// the automation gate. The runtime can now be constructed while
// ProtectionReady is UNWIRED, but the gateway refuses the tracer's
// exposure-raising entry from the mutation's own shape. Adding a second gate
// here would be a second place to get the answer wrong.
//
// What the tracer *does* add is a set of bounds that hold whatever the gate
// says, because "the gate is off" is not a reason to accept an unbounded
// parameter set in a file whose whole purpose is to be pointed at a live account
// one day:
//
//	one symbol            a tracer that could take a list is a strategy runner
//	LIMIT only            a market order has no exposure valuation to bound
//	an explicit quantity  and a notional ceiling that is checked against it
//	a freshness bound     a price older than it is not a price to trade on
//	abort criteria        cycles and wall time, both required and both finite
//
// # It refuses to start on a non-empty account
//
// The tracer measures one trade. On an account that already holds something, the
// aggregate it would be sized against is not the one this file can compute
// (open exposure needs a price for every held symbol), and the exit policy's
// working set would contain positions the tracer did not open. Both are reasons
// to stop rather than to guess, and the refusal is cheap: a tracer is run on a
// flat account by definition.

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// ErrTracerRefused is the class of every refusal to start or continue a tracer
// run. It is one class because the operator's next action is the same for all of
// them: read the detail, change a parameter or the account, run it again.
var ErrTracerRefused = errors.New("engine: the tracer refused")

// EntryIssuer mints the EXPOSURE_RAISING decision the tracer's entry travels
// under. *execgw.RiskGuardian satisfies it.
type EntryIssuer interface {
	IssueEntry(ctx context.Context, req execgw.EntryIssuance) (execgw.Issued, error)
}

// GateReads are the four reads the entry gate ages (execgw.DefaultStaleness).
// OfficialReads satisfies it.
//
// The tracer needs all four and not only the price, because the gate is
// fail-closed on every one of them: a required query that has never succeeded
// is "infinitely stale" and blocks the entry. Establishing that freshness is
// part of what a tracer is *for* — it is the first thing a live run would have
// to do, and doing it here means the run either has a complete, fresh picture
// or does not start.
type GateReads interface {
	PriceReader
	Orders(ctx context.Context, filter official.OrdersFilter) ([]domain.Order, error)
	BuyingPower(ctx context.Context, currency string) (domain.BuyingPower, error)
	Holdings(ctx context.Context, symbol string) ([]domain.Position, error)
}

// TracerParams is the whole parameter surface, and every field is required.
//
// There are no defaults on purpose. A default quantity or a default ceiling is a
// number nobody chose appearing in a run whose entire value is that somebody
// chose every number in it.
type TracerParams struct {
	// Symbol is the one allowlisted symbol the run trades.
	Symbol string
	// Market is "kr" or "us".
	Market string
	// Quantity is the whole number of units to buy — the minimum the venue
	// admits, for a run whose purpose is evidence rather than profit.
	Quantity string
	// LimitPrice is the entry limit. MARKET is not offered: an automated entry
	// with no price has no exposure valuation to size or reserve against
	// (riskcalc.CheckAutomatedEntry).
	LimitPrice string
	// StopPrice is the protective stop, which becomes the exit policy's t0
	// baseline. TargetPrice is the profit target the chain measures reward:risk
	// against. Both are required for the same reason an ordinary entry needs
	// them: No Stop = No Trade.
	StopPrice   string
	TargetPrice string
	// NotionalCeiling bounds quantity × limit price. It is checked here as well
	// as by the Guardian chain, and that duplication is deliberate: this is the
	// number the operator of *this run* set, and it must be able to be smaller
	// than the account's policy without editing the account's policy.
	NotionalCeiling string
	// Freshness is how old an observed price may be. A tracer that traded on a
	// stale quote would be measuring the ledger, not the broker.
	Freshness time.Duration
	// MaxCycles and MaxDuration are the abort criteria. Both are required and
	// both are finite: a run that could not end is not an experiment.
	MaxCycles   int
	MaxDuration time.Duration
}

// Validate refuses a parameter set that is not a tracer's.
func (p TracerParams) Validate() error {
	refuse := func(format string, args ...any) error {
		return fmt.Errorf("%w: %s", ErrTracerRefused, fmt.Sprintf(format, args...))
	}
	if strings.TrimSpace(p.Symbol) == "" {
		return refuse("a tracer trades one named symbol and none was given")
	}
	if strings.Contains(p.Symbol, ",") || strings.Contains(p.Symbol, " ") {
		return refuse("%q looks like more than one symbol; a tracer trades exactly one", p.Symbol)
	}
	if _, err := clock.ParseMarket(p.Market); err != nil {
		return refuse("market %q is not one this build trades", p.Market)
	}
	quantity, ok := new(big.Rat).SetString(strings.TrimSpace(p.Quantity))
	if !ok || quantity.Sign() <= 0 || !quantity.IsInt() {
		return refuse("quantity %q is not a positive whole number of units", p.Quantity)
	}
	limit, ok := new(big.Rat).SetString(strings.TrimSpace(p.LimitPrice))
	if !ok || limit.Sign() <= 0 {
		return refuse("a tracer entry is LIMIT only and %q is not a price", p.LimitPrice)
	}
	stop, ok := new(big.Rat).SetString(strings.TrimSpace(p.StopPrice))
	if !ok || stop.Sign() <= 0 {
		return refuse("stop %q is not a price; No Stop = No Trade", p.StopPrice)
	}
	if stop.Cmp(limit) >= 0 {
		return refuse("stop %s is not below the entry %s", p.StopPrice, p.LimitPrice)
	}
	target, ok := new(big.Rat).SetString(strings.TrimSpace(p.TargetPrice))
	if !ok || target.Cmp(limit) <= 0 {
		return refuse("target %q is not above the entry %s", p.TargetPrice, p.LimitPrice)
	}
	ceiling, ok := new(big.Rat).SetString(strings.TrimSpace(p.NotionalCeiling))
	if !ok || ceiling.Sign() <= 0 {
		return refuse("a tracer needs a notional ceiling and %q is not one", p.NotionalCeiling)
	}
	if notional := new(big.Rat).Mul(quantity, limit); notional.Cmp(ceiling) > 0 {
		return refuse("%s × %s exceeds the notional ceiling %s",
			p.Quantity, p.LimitPrice, p.NotionalCeiling)
	}
	if p.Freshness <= 0 {
		return refuse("a tracer needs a freshness bound; a price with no age limit is not an observation")
	}
	if p.MaxCycles <= 0 || p.MaxDuration <= 0 {
		return refuse("a tracer needs both abort criteria: cycles (%d) and wall time (%s)",
			p.MaxCycles, p.MaxDuration)
	}
	return nil
}

// Notional is quantity × limit price, as a decimal string.
func (p TracerParams) Notional() string {
	quantity, ok := new(big.Rat).SetString(strings.TrimSpace(p.Quantity))
	if !ok {
		return "0"
	}
	limit, ok := new(big.Rat).SetString(strings.TrimSpace(p.LimitPrice))
	if !ok {
		return "0"
	}
	return trimRational(new(big.Rat).Mul(quantity, limit))
}

// TracerOptions is the machinery one run drives. Every field is required, and
// they are named rather than reached out of a Context for the reason
// NewOrderPath is: a caller has to ask for this by name.
type TracerOptions struct {
	Journal    *journal.Journal
	Issuer     EntryIssuer
	Submit     ExitSubmitter
	Observer   *ExitObserver
	Retrier    *execgw.Retrier
	Reads      GateReads
	Clock      clock.Clock
	AccountRef string
	Params     TracerParams
}

// TracerReport is what one run did.
type TracerReport struct {
	// EntryIntentID and EntryOrderID name the order the run opened with.
	EntryIntentID string
	EntryOrderID  string
	// PositionID is the instance the entry fill produced, empty when nothing
	// filled before the run ended.
	PositionID string
	// Cycles is how many observation cycles ran.
	Cycles int
	// Proposals is how many exit proposals reached the broker.
	Proposals int
	// Closed reports that the exit policy completed: the position reached CLOSED
	// and its state is finished. That is the run's success condition.
	Closed bool
	// Aborted names the criterion that ended the run early, empty when it ended
	// on its own.
	Aborted string
	// Outcome is the frozen trade outcome, when there is one.
	Outcome *journal.TradeOutcome
}

// Tracer runs one parameterised pass.
type Tracer struct {
	opts TracerOptions
}

// NewTracer validates everything before anything can be submitted.
func NewTracer(opts TracerOptions) (*Tracer, error) {
	if err := opts.Params.Validate(); err != nil {
		return nil, err
	}
	switch {
	case opts.Journal == nil:
		return nil, fmt.Errorf("%w: a tracer needs the journal that records it", ErrTracerRefused)
	case opts.Issuer == nil:
		return nil, fmt.Errorf("%w: a tracer needs an entry issuer", ErrTracerRefused)
	case opts.Submit == nil:
		return nil, fmt.Errorf("%w: a tracer needs the execution gateway", ErrTracerRefused)
	case opts.Observer == nil:
		return nil, fmt.Errorf("%w: a tracer needs the exit observation loop; "+
			"an entry with nothing watching it is the one thing this must never produce", ErrTracerRefused)
	case opts.Retrier == nil || opts.Reads == nil:
		return nil, fmt.Errorf("%w: a tracer needs the broker reads and the retrier that ages them",
			ErrTracerRefused)
	case strings.TrimSpace(opts.AccountRef) == "":
		return nil, fmt.Errorf("%w: a tracer is scoped to one account", ErrTracerRefused)
	}
	if opts.Clock == nil {
		opts.Clock = clock.System()
	}
	return &Tracer{opts: opts}, nil
}

// Run performs the pass: check the account is flat, observe a fresh price, issue
// and submit the entry, then drive the exit observation loop until the position
// closes or an abort criterion trips.
//
// It never places a sell itself. Every exit order in the run is the exit
// policy's, submitted through the same path a production engine would use —
// which is the whole point of the slice: what is being traced is the machinery,
// not a script that imitates it.
func (t *Tracer) Run(ctx context.Context) (TracerReport, error) {
	var report TracerReport
	p := t.opts.Params

	if err := t.requireFlatAccount(ctx); err != nil {
		return report, err
	}
	if err := t.freshenGateInputs(ctx); err != nil {
		return report, err
	}
	if err := t.submitEntry(ctx, &report); err != nil {
		return report, err
	}

	deadline := t.opts.Clock.Now().Add(p.MaxDuration)
	for report.Cycles < p.MaxCycles {
		if err := ctx.Err(); err != nil {
			report.Aborted = "the run's context ended"
			return report, err
		}
		if !t.opts.Clock.Now().Before(deadline) {
			report.Aborted = fmt.Sprintf("the wall-time budget of %s ran out", p.MaxDuration)
			break
		}
		cycle := t.opts.Observer.ObserveOnce(ctx)
		report.Cycles++
		report.Proposals += cycle.Proposed
		if cycle.Err != nil {
			// An observation failure is a hold, not an abort: that is the exit
			// policy's own rule, and a tracer that stopped on one would report a
			// transient read failure as a machinery failure.
			continue
		}
		if done, err := t.settled(ctx, &report); err != nil {
			return report, err
		} else if done {
			report.Closed = true
			break
		}
		if err := t.opts.Clock.Sleep(ctx, t.opts.Observer.Interval()); err != nil {
			report.Aborted = "the run's context ended"
			return report, err
		}
	}
	if !report.Closed && report.Aborted == "" {
		report.Aborted = fmt.Sprintf("the cycle budget of %d ran out", p.MaxCycles)
	}
	t.attachOutcome(ctx, &report)
	return report, nil
}

// requireFlatAccount refuses a run on an account that already holds something.
func (t *Tracer) requireFlatAccount(ctx context.Context) error {
	positions, err := t.opts.Journal.Positions(ctx, t.opts.AccountRef)
	if err != nil {
		return err
	}
	for _, held := range positions {
		if held.State == journal.PositionClosed || isZeroQuantity(held.Quantity) {
			continue
		}
		return fmt.Errorf(
			"%w: the account already holds %s of %s; a tracer measures one trade and cannot size "+
				"against an exposure it did not open", ErrTracerRefused, held.Quantity, held.Symbol)
	}
	return nil
}

// freshenGateInputs performs the four required reads through the Retrier and
// refuses a picture the run cannot trade on.
//
// Every read goes through the Retrier and that is the point of them: a success
// stamps the query's freshness on the entry gate, and the gate refuses an entry
// while any required query is stale — including one that has never succeeded at
// all. A tracer that skipped this would be refused by the gateway a moment
// later, with a reason pointing at the gate rather than at the run.
//
// The price is the one whose *value* is used, and the freshness bound is checked
// against the round trip: a quote that took longer to arrive than the parameter
// admits is not a quote this run may trade on.
func (t *Tracer) freshenGateInputs(ctx context.Context) error {
	symbol := strings.ToUpper(strings.TrimSpace(t.opts.Params.Symbol))
	currency := "KRW"
	if strings.EqualFold(strings.TrimSpace(t.opts.Params.Market), "us") {
		currency = "USD"
	}

	if err := t.opts.Retrier.Query(ctx, execgw.QueryOpenOrders, func(ctx context.Context) error {
		_, err := t.opts.Reads.Orders(ctx, official.OrdersFilter{})
		return err
	}); err != nil {
		return fmt.Errorf("%w: the open-order list could not be read: %v", ErrTracerRefused, err)
	}
	if err := t.opts.Retrier.Query(ctx, execgw.QueryBuyingPower, func(ctx context.Context) error {
		_, err := t.opts.Reads.BuyingPower(ctx, currency)
		return err
	}); err != nil {
		return fmt.Errorf("%w: the buying power could not be read: %v", ErrTracerRefused, err)
	}
	if err := t.opts.Retrier.Query(ctx, execgw.QueryHoldings, func(ctx context.Context) error {
		_, err := t.opts.Reads.Holdings(ctx, symbol)
		return err
	}); err != nil {
		return fmt.Errorf("%w: the holdings could not be read: %v", ErrTracerRefused, err)
	}

	asked := t.opts.Clock.Now()
	var last string
	if err := t.opts.Retrier.Query(ctx, execgw.QueryPrice, func(ctx context.Context) error {
		quotes, err := t.opts.Reads.Prices(ctx, []string{symbol})
		if err != nil {
			return err
		}
		for _, q := range quotes {
			if strings.EqualFold(strings.TrimSpace(q.Symbol), symbol) && q.Last > 0 {
				last = decimalOf(q.Last)
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("%w: the price of %s could not be read: %v", ErrTracerRefused, symbol, err)
	}
	if last == "" {
		return fmt.Errorf("%w: the broker returned no last trade for %s", ErrTracerRefused, symbol)
	}
	if age := t.opts.Clock.Now().Sub(asked); age > t.opts.Params.Freshness {
		return fmt.Errorf("%w: the price of %s took %s to arrive, past the freshness bound of %s",
			ErrTracerRefused, symbol, age, t.opts.Params.Freshness)
	}
	return nil
}

// submitEntry issues and places the one entry of the run.
func (t *Tracer) submitEntry(ctx context.Context, report *TracerReport) error {
	p := t.opts.Params
	market, err := clock.ParseMarket(p.Market)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrTracerRefused, err)
	}
	symbol := strings.ToUpper(strings.TrimSpace(p.Symbol))

	issued, err := t.opts.Issuer.IssueEntry(ctx, execgw.EntryIssuance{
		Intent: risk.Intent{
			AccountRef: t.opts.AccountRef,
			Market:     costs.Market(market),
			Symbol:     symbol,
			Side:       risk.SideBuy,
			Quantity:   strings.TrimSpace(p.Quantity),
			LimitPrice: strings.TrimSpace(p.LimitPrice),
			StopPrice:  strings.TrimSpace(p.StopPrice),
			// The target is the chain's reward:risk input; it is not a resting
			// order and this build places none.
			TargetPrice: strings.TrimSpace(p.TargetPrice),
		},
		Account: t.accountState(),
		Collect: t.collectExposure,
	})
	if err != nil {
		return fmt.Errorf("%w: the entry was not authorised: %v", ErrTracerRefused, err)
	}

	quantity, err := floatOf("tracer quantity", p.Quantity)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrTracerRefused, err)
	}
	limit, err := floatOf("tracer limit price", p.LimitPrice)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrTracerRefused, err)
	}
	out, err := t.opts.Submit.Place(ctx, execgw.PlaceRequest{
		Intent: orderintent.PlaceIntent{
			Symbol: symbol, Market: string(market), Side: "buy", OrderType: "limit",
			Quantity: quantity, Price: limit, CurrencyMode: currencyFor(string(market)),
		},
		Decision: issued.Decision,
	})
	report.EntryIntentID = out.IntentID
	report.EntryOrderID = out.BrokerOrderID
	if out.State != journal.StateConfirmed {
		detail := out.Detail
		if detail == "" && err != nil {
			detail = err.Error()
		}
		return fmt.Errorf("%w: the entry did not confirm (%s): %s", ErrTracerRefused, out.State, detail)
	}
	return nil
}

// accountState is what the chain measures the tracer's entry against.
//
// The aggregates are zero because requireFlatAccount has just established that
// they are: nothing is held, so open exposure is nothing. They are stated rather
// than assumed, and if that precondition is ever relaxed this is the function
// that has to learn to compute them.
func (t *Tracer) accountState() risk.AccountState {
	currency := "KRW"
	if strings.EqualFold(strings.TrimSpace(t.opts.Params.Market), "us") {
		currency = "USD"
	}
	return risk.AccountState{
		Mode:              risk.ModeNormal,
		AllowedSymbols:    []string{strings.ToUpper(strings.TrimSpace(t.opts.Params.Symbol))},
		CashAvailable:     riskcalc.Money{Amount: t.opts.Params.NotionalCeiling, Currency: currency},
		OpenExposure:      riskcalc.Money{Amount: "0", Currency: currency},
		DailyRealizedLoss: riskcalc.Money{Amount: "0", Currency: currency},
		AccountEquity:     riskcalc.Money{Amount: t.opts.Params.NotionalCeiling, Currency: currency},
	}
}

// collectExposure supplies the reservation snapshot the atomic issuance takes.
func (t *Tracer) collectExposure(ctx context.Context, _ int) (execgw.ExposureSnapshot, error) {
	version, err := t.opts.Journal.ReservationVersion(ctx, t.opts.AccountRef)
	if err != nil {
		return execgw.ExposureSnapshot{}, err
	}
	currency := "KRW"
	if strings.EqualFold(strings.TrimSpace(t.opts.Params.Market), "us") {
		currency = "USD"
	}
	return execgw.ExposureSnapshot{
		AsOf:         t.opts.Clock.Now(),
		Version:      version,
		OpenExposure: riskcalc.Money{Amount: "0", Currency: currency},
	}, nil
}

// settled reports whether the exit policy has finished with the position.
func (t *Tracer) settled(ctx context.Context, report *TracerReport) (bool, error) {
	if report.PositionID == "" {
		held, err := t.opts.Journal.CurrentPosition(ctx, t.opts.AccountRef,
			t.opts.Params.Market, t.opts.Params.Symbol)
		if errors.Is(err, journal.ErrPositionNotFound) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		report.PositionID = held.ID
	}
	state, err := t.opts.Journal.ExitState(ctx, report.PositionID)
	if errors.Is(err, journal.ErrExitStateNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return state.Completed, nil
}

// attachOutcome reads the frozen outcome, when the run produced one. Its absence
// is not an error: analytics is isolated from the order path and a run that
// closed a position without one is still a run that closed a position.
func (t *Tracer) attachOutcome(ctx context.Context, report *TracerReport) {
	if report.PositionID == "" {
		return
	}
	outcome, err := t.opts.Journal.TradeOutcomeOf(ctx, report.PositionID)
	if err != nil {
		return
	}
	report.Outcome = &outcome
}

// trimRational renders a rational without trailing zeros.
func trimRational(r *big.Rat) string {
	if r.IsInt() {
		return r.Num().String()
	}
	out := strings.TrimRight(r.FloatString(12), "0")
	return strings.TrimSuffix(out, ".")
}
