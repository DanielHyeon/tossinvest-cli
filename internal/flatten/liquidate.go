package flatten

// liquidate.go is phase 2 of the flatten saga: sell down every position, once the
// cancels are settled (harden-execution-base task 4.5, engine-safety "Flatten
// Saga" steps 3-5).
//
// # The sequence, and why each step is where it is
//
//	3. re-read the account until it stops moving
//	4. sell the *sellable* quantity of each symbol, as an aggressive limit
//	5. reconcile again, and repeat until nothing is left
//
// Step 3 is the one that looks like it violates §0.3 and does not. The
// stabilisation wait is not in front of an exit; it is in front of *deciding how
// much to exit*. Selling from a snapshot taken mid-settlement means selling a
// quantity the account does not have, and an oversell is a new short position —
// a bigger exposure than the one being closed. The cancels in phase 1 have
// already removed every live order without waiting for anything.
//
// # Reduce-only, by construction rather than by flag
//
// The broker has no reduce-only order type, so "reduce-only" here means the
// quantity comes from the account and never from an intention:
//
//   - the sellable quantity is read per symbol, immediately before the order;
//   - a symbol whose cancel did not settle is skipped entirely, because a live
//     sell order plus a full-size sell is exactly how a flatten overshoots;
//   - nothing in this file can produce a BUY.
//
// # Aggressive, but inside the band
//
// A limit at the exchange's own lower limit (하한가) is the most aggressive price
// that is guaranteed to be a valid tick and guaranteed to be accepted. That is
// preferred wherever the broker reports one. Where it does not, the price is the
// last trade discounted by AggressiveDiscount and rounded *down* to the market's
// tick — down, because for a sell, lower is more likely to fill.

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

// DefaultAggressiveDiscount is how far below the last trade a liquidation limit
// is placed when the broker reports no price band.
//
// Five percent: far enough to cross the spread and a few levels of book on any
// liquid name, near enough that a thin book does not print an absurd fill. Where
// a band exists the band's own floor is used instead, which is both more
// aggressive and always a valid tick.
const DefaultAggressiveDiscount = 0.05

// DefaultLiquidationRounds is how many sell-and-recheck cycles are attempted
// before the saga stops and asks for a human.
//
// Three: the first sells, the second catches what the first could not (a partial
// fill, a quantity that settled in between), and a third that still leaves a
// position means something is wrong that another loop will not fix.
const DefaultLiquidationRounds = 3

// PositionReader sweeps the account's holdings.
type PositionReader = reconcile.PositionsReader

// SellableReader reports how much of a symbol can be sold right now.
//
// This is deliberately not "how much do I hold": settlement, pledges and
// same-day buys make the two differ, and the sellable number is the only one an
// order can be built from without risking a rejection or an oversell.
type SellableReader interface {
	SellableQuantity(ctx context.Context, symbol string) (domain.SellableQuantity, error)
}

// PriceReader supplies the inputs to the aggressive limit price.
type PriceReader interface {
	PriceLimits(ctx context.Context, symbol string) (domain.PriceLimits, error)
	Prices(ctx context.Context, symbols []string) ([]domain.Quote, error)
}

// LiquidationTarget is one position the saga intends to sell.
type LiquidationTarget struct {
	Symbol   string
	Market   string
	Currency string
	// Held is the holding as the account reports it.
	Held float64
	// Sellable is the quantity that can actually be sold now.
	Sellable float64
	// Quantity is what the order will carry: the whole-share part of Sellable.
	Quantity float64
	// Remainder is the fractional part no limit order can express.
	Remainder float64
	// Price is the aggressive limit.
	Price float64
	// PriceSource explains where Price came from, for the operator.
	PriceSource string
	// State is the outcome (journal.FlattenStep*).
	State string
	// Reason and Detail explain a skip or a failure.
	Reason string
	Detail string
}

// LiquidateReport is the result of phase 2.
type LiquidateReport struct {
	SagaID string
	DryRun bool
	// Rounds is how many sell-and-recheck cycles ran.
	Rounds int
	// Targets is every position considered, across all rounds, latest state.
	Targets []LiquidationTarget
	// Submitted is how many sell orders were sent.
	Submitted int
	// Held is how many symbols were skipped to avoid an oversell.
	Held int
	// Remaining is how many positions the account still showed at the end.
	Remaining int
	Phase     string
}

// Flat reports whether the account ended with nothing left.
func (r LiquidateReport) Flat() bool { return r.Remaining == 0 && r.Held == 0 }

// Liquidate runs phase 2.
//
// It requires the cancel phase to have settled *for the symbols it sells*, not
// globally: one symbol stuck in doubt must not stop the other nine from being
// exited. A symbol that is held is reported as held, never silently dropped.
func (s *Saga) Liquidate(ctx context.Context) (LiquidateReport, error) {
	if err := s.validateLiquidation(); err != nil {
		return LiquidateReport{}, err
	}

	saga, err := s.Journal.ActiveFlatten(ctx)
	if err != nil {
		return LiquidateReport{}, fmt.Errorf("flatten: no flatten saga is running: %w", err)
	}
	report := LiquidateReport{SagaID: saga.ID, DryRun: s.DryRun, Phase: saga.Phase}

	// The gate stays latched from phase 1; re-raising it is idempotent and makes
	// Liquidate safe to call on its own.
	s.blockEntries()

	rounds := s.Rounds
	if rounds <= 0 {
		rounds = DefaultLiquidationRounds
	}

	byKey := map[string]LiquidationTarget{}
	for round := 1; round <= rounds; round++ {
		report.Rounds = round

		// 3. Stabilise. Deciding *how much* to sell from a moving account is what
		//    produces an oversell.
		if err := s.Journal.SetFlattenPhase(ctx, saga.ID, journal.FlattenPhaseStabilising,
			fmt.Sprintf("round %d", round)); err != nil {
			return report, err
		}
		snapshot, err := s.stabilise(ctx)
		if err != nil {
			return report, err
		}

		targets, err := s.planLiquidation(ctx, saga.ID, snapshot)
		if err != nil {
			return report, err
		}
		if len(targets) == 0 {
			for _, t := range byKey {
				report.Targets = append(report.Targets, t)
			}
			break
		}

		// 4. Sell.
		if err := s.Journal.SetFlattenPhase(ctx, saga.ID, journal.FlattenPhaseLiquidating,
			fmt.Sprintf("round %d: %d position(s)", round, len(targets))); err != nil {
			return report, err
		}
		for i := range targets {
			s.sell(ctx, saga.ID, round, &targets[i])
			byKey[targets[i].Market+"|"+targets[i].Symbol] = targets[i]
		}

		// 5. Recheck. A dry run stops here: it has produced the list, and running
		//    the loop again would only re-produce it.
		if s.DryRun {
			break
		}
	}

	report.Targets = collectTargets(byKey)
	for _, t := range report.Targets {
		switch t.State {
		case journal.FlattenStepDone:
			report.Submitted++
		case journal.FlattenStepHeld:
			report.Held++
		}
	}

	// Final verification: what does the account still show?
	if err := s.Journal.SetFlattenPhase(ctx, saga.ID, journal.FlattenPhaseVerifying, ""); err != nil {
		return report, err
	}
	remaining, err := s.remainingPositions(ctx)
	if err != nil {
		return report, err
	}
	report.Remaining = remaining

	phase := journal.FlattenPhaseComplete
	detail := "the account is flat"
	switch {
	case s.DryRun:
		phase = journal.FlattenPhaseCancelled
		detail = "dry run: nothing was submitted"
	case !report.Flat():
		phase = journal.FlattenPhaseStalled
		detail = fmt.Sprintf("%d position(s) remain and %d symbol(s) are held; an operator has to finish this",
			report.Remaining, report.Held)
	}
	if err := s.Journal.SetFlattenPhase(ctx, saga.ID, phase, detail); err != nil {
		return report, err
	}
	report.Phase = phase

	switch phase {
	case journal.FlattenPhaseComplete:
		s.event(obs.EventFlattenComplete, saga.ID+"|complete", map[string]any{
			"saga_id":       saga.ID,
			"rounds":        report.Rounds,
			"submitted":     report.Submitted,
			obs.FieldDetail: detail,
		})
	case journal.FlattenPhaseStalled:
		s.event(obs.EventFlattenStalled, saga.ID+"|liquidate", map[string]any{
			"saga_id":       saga.ID,
			"remaining":     report.Remaining,
			"held":          report.Held,
			obs.FieldDetail: detail,
		})
	}
	return report, nil
}

// --- planning ---------------------------------------------------------------

// PlanLiquidation reports what Liquidate would sell, without selling anything.
//
// It is what the confirmation prompt is built from, and it is a real plan rather
// than an estimate: the same snapshot, the same sellable reads, the same oversell
// guard. An operator who is shown a number that Liquidate then does not act on has
// been asked to confirm something other than what happens.
func (s *Saga) PlanLiquidation(ctx context.Context) ([]LiquidationTarget, error) {
	if err := s.validateLiquidation(); err != nil {
		return nil, err
	}
	saga, err := s.Journal.ActiveFlatten(ctx)
	if err != nil {
		return nil, fmt.Errorf("flatten: no flatten saga is running: %w", err)
	}
	snapshot, err := s.stabilise(ctx)
	if err != nil {
		return nil, err
	}
	return s.planLiquidation(ctx, saga.ID, snapshot)
}

// planLiquidation turns a stable snapshot into the orders to send.
func (s *Saga) planLiquidation(ctx context.Context, sagaID string, snapshot reconcile.Snapshot) ([]LiquidationTarget, error) {
	// The oversell guard, read fresh: a symbol whose cancel is unsettled may
	// still have a live order against it.
	held, err := s.Journal.UnsettledCancelSymbols(ctx, sagaID)
	if err != nil {
		return nil, err
	}

	var targets []LiquidationTarget
	for _, holding := range snapshot.Holdings {
		quantity := parseDecimal(holding.Quantity)
		if quantity <= 0 {
			// A zero or short position is not something this saga sells. Going
			// further negative is the opposite of flattening.
			continue
		}
		symbol := strings.ToUpper(strings.TrimSpace(holding.Symbol))
		market := holding.Market
		if market == "" {
			market = MarketOf(symbol, "")
		}
		target := LiquidationTarget{
			Symbol:   symbol,
			Market:   market,
			Currency: currencyFor(market),
			Held:     quantity,
		}

		if state, blocked := held[symbol]; blocked {
			target.State = journal.FlattenStepHeld
			target.Reason = string(execgw.ReasonBrokerOutcomeUnknown)
			target.Detail = fmt.Sprintf(
				"the cancel for %s is %s; selling the full holding while an order may still be live "+
					"would oversell, so this symbol waits for an operator", symbol, state)
			targets = append(targets, target)
			continue
		}

		if err := s.priceAndSize(ctx, &target); err != nil {
			target.State = journal.FlattenStepHeld
			target.Reason = string(execgw.ReasonBalanceUnavailable)
			target.Detail = err.Error()
		}
		targets = append(targets, target)
	}

	sort.Slice(targets, func(i, j int) bool { return targets[i].Symbol < targets[j].Symbol })
	return targets, nil
}

// priceAndSize fills in the sellable quantity and the aggressive limit price.
func (s *Saga) priceAndSize(ctx context.Context, target *LiquidationTarget) error {
	sellable, err := s.Sellable.SellableQuantity(ctx, target.Symbol)
	if err != nil {
		return fmt.Errorf("the sellable quantity of %s could not be read, so no order can be sized: %w",
			target.Symbol, err)
	}
	// The account is the authority, and the smaller of the two numbers is the
	// one that cannot oversell.
	target.Sellable = math.Min(sellable.Quantity, target.Held)
	if target.Sellable <= 0 {
		target.State = journal.FlattenStepHeld
		target.Reason = string(execgw.ReasonBrokerOutcomeUnknown)
		target.Detail = fmt.Sprintf("%s reports 0 sellable of %v held (unsettled or pledged)",
			target.Symbol, target.Held)
		return nil
	}

	whole := math.Floor(target.Sellable + 1e-9)
	target.Quantity = whole
	target.Remainder = target.Sellable - whole
	if target.Quantity <= 0 {
		// A purely fractional position. A limit order cannot express it (the
		// official path takes fractional orders as amount-based market buys
		// only), so it is reported rather than silently left behind.
		target.State = journal.FlattenStepHeld
		target.Reason = string(execgw.ReasonUnsupportedOrderType)
		target.Detail = fmt.Sprintf(
			"%s holds only a fractional %v, which no limit order can express; an operator has to exit it",
			target.Symbol, target.Sellable)
		return nil
	}

	price, source, err := s.aggressivePrice(ctx, target.Symbol, target.Market)
	if err != nil {
		return err
	}
	target.Price = price
	target.PriceSource = source
	return nil
}

// aggressivePrice returns the most aggressive sell limit that is still valid.
func (s *Saga) aggressivePrice(ctx context.Context, symbol, market string) (float64, string, error) {
	if s.Prices != nil {
		if limits, err := s.Prices.PriceLimits(ctx, symbol); err == nil && limits.LowerLimit > 0 {
			// The exchange's own floor: maximally aggressive, always a valid tick,
			// and never rejected for being outside the band.
			return limits.LowerLimit, "exchange lower limit", nil
		}

		quotes, err := s.Prices.Prices(ctx, []string{symbol})
		if err != nil {
			return 0, "", fmt.Errorf("no price for %s, so no limit can be set: %w", symbol, err)
		}
		for _, q := range quotes {
			if !strings.EqualFold(strings.TrimSpace(q.Symbol), symbol) || q.Last <= 0 {
				continue
			}
			discount := s.AggressiveDiscount
			if discount <= 0 {
				discount = DefaultAggressiveDiscount
			}
			// Rounded down: for a sell, lower is more likely to fill, and rounding
			// up could land outside the tick grid or above the last trade.
			return roundDownToTick(q.Last*(1-discount), market), "last trade less the aggressive discount", nil
		}
	}
	return 0, "", fmt.Errorf("no usable price for %s; a liquidation order cannot be priced", symbol)
}

// roundDownToTick snaps a price to a valid tick for the market.
//
// KR tick sizes are stepped by price band; the table below is the KRX schedule.
// US equities tick at a cent above $1. Getting this wrong means a rejected order
// during an emergency exit, which is why it is a table and not a guess.
func roundDownToTick(price float64, market string) float64 {
	if price <= 0 {
		return 0
	}
	if strings.EqualFold(market, "kr") {
		tick := krTick(price)
		return math.Floor(price/tick) * tick
	}
	// US: cents.
	return math.Floor(price*100) / 100
}

// krTick is the KRX tick schedule for equities.
func krTick(price float64) float64 {
	switch {
	case price < 2000:
		return 1
	case price < 5000:
		return 5
	case price < 20000:
		return 10
	case price < 50000:
		return 50
	case price < 200000:
		return 100
	case price < 500000:
		return 500
	default:
		return 1000
	}
}

// --- selling ----------------------------------------------------------------

// sell submits one reduce-only liquidation order.
//
// round is part of the step's identity, so a later round can sell a remainder the
// first one did not clear — while a *resumed* run inside the same round finds the
// step already recorded and does not submit a second order for it.
func (s *Saga) sell(ctx context.Context, sagaID string, round int, target *LiquidationTarget) {
	if target.State == journal.FlattenStepHeld {
		s.recordLiquidationStep(ctx, sagaID, round, *target)
		return
	}
	if target.Quantity <= 0 || target.Price <= 0 {
		target.State = journal.FlattenStepHeld
		if target.Detail == "" {
			target.Detail = "the order could not be sized or priced"
		}
		s.recordLiquidationStep(ctx, sagaID, round, *target)
		return
	}

	step := s.recordLiquidationStep(ctx, sagaID, round, *target)
	if s.DryRun {
		target.State = journal.FlattenStepPending
		target.Detail = fmt.Sprintf("dry run: would sell %s of %s at %s (%s)",
			decimalString(target.Quantity), target.Symbol, decimalString(target.Price), target.PriceSource)
		return
	}
	// A step already settled by an earlier round or an earlier process must not
	// produce a second order: this broker has no idempotency key.
	if step.State == journal.FlattenStepDone || step.State == journal.FlattenStepInDoubt {
		target.State = step.State
		target.Detail = "already submitted by an earlier round"
		return
	}

	intent := orderintent.PlaceIntent{
		Symbol:       target.Symbol,
		Market:       target.Market,
		Side:         "sell",
		OrderType:    "limit",
		Quantity:     target.Quantity,
		Price:        target.Price,
		CurrencyMode: target.Currency,
	}
	// A liquidation sell is an order creation, so its decision carries an
	// idempotency key — and it is still RISK_REDUCING, so no limit snapshot
	// travels with it and none is applied (§0.3).
	decision, err := s.decisionFor(ctx, execgw.ReductionRequest{
		Kind:        journal.KindPlace,
		Market:      target.Market,
		Symbol:      target.Symbol,
		Side:        "SELL",
		MaxQuantity: target.Quantity,
		Reason:      s.reductionReason(fmt.Sprintf("liquidation round %d", round)),
	})
	if err != nil {
		// Nothing was submitted. The step stays failed rather than in doubt: no
		// decision was recorded, so no order could have been sent under one.
		target.State = journal.FlattenStepFailed
		target.Reason = string(execgw.ReasonGuardianMissing)
		target.Detail = "the liquidation decision could not be recorded: " + err.Error()
		if step.ID != 0 {
			_ = s.Journal.UpdateFlattenStep(ctx, step.ID, target.State, "", "",
				target.Reason, target.Detail)
		}
		return
	}

	out, err := s.Gateway.Place(ctx, execgw.PlaceRequest{
		Intent:   intent,
		Decision: decision,
	})

	switch out.State {
	case journal.StateConfirmed:
		target.State = journal.FlattenStepDone
		target.Detail = fmt.Sprintf("sell %s @ %s submitted as %s",
			decimalString(target.Quantity), decimalString(target.Price), out.BrokerOrderID)
	case journal.StateInDoubt, journal.StateUnresolvedInDoubt:
		target.State = journal.FlattenStepInDoubt
		target.Detail = out.Detail
	default:
		target.State = journal.FlattenStepFailed
		target.Detail = out.Detail
		if target.Detail == "" && err != nil {
			target.Detail = err.Error()
		}
	}
	target.Reason = string(out.Reason)

	if step.ID != 0 {
		_ = s.Journal.UpdateFlattenStep(ctx, step.ID, target.State,
			out.IntentID, out.AttemptID, target.Reason, target.Detail)
	}

	s.logf(obs.EventFlattenProgress,
		"saga_id", sagaID,
		obs.FieldSymbol, target.Symbol,
		obs.FieldQuantity, target.Quantity,
		obs.FieldPrice, target.Price,
		"price_source", target.PriceSource,
		obs.FieldToState, target.State)
}

func (s *Saga) recordLiquidationStep(ctx context.Context, sagaID string, round int, target LiquidationTarget) journal.FlattenStep {
	state := target.State
	if state == "" {
		state = journal.FlattenStepPending
	}
	step, err := s.Journal.AddFlattenStep(ctx, journal.FlattenStep{
		SagaID: sagaID,
		Kind:   journal.FlattenStepLiquidate,
		Market: target.Market,
		Symbol: target.Symbol,
		// The round is the step's discriminator: a remainder left by round one is
		// a different order from round one's, and both belong in the record.
		TargetOrderID: fmt.Sprintf("round-%d", round),
		Side:          "SELL",
		Quantity:      decimalString(target.Quantity),
		Price:         decimalString(target.Price),
		Currency:      target.Currency,
		State:         state,
	})
	if err != nil {
		return journal.FlattenStep{}
	}
	if step.State != state && (state == journal.FlattenStepHeld) {
		_ = s.Journal.UpdateFlattenStep(ctx, step.ID, state, "", "", target.Reason, target.Detail)
	}
	return step
}

// --- account reads ----------------------------------------------------------

// stabilise reads the account until two readings agree.
//
// It gives up after the configured attempts rather than looping forever: an
// account that never stops moving is one where somebody else is trading, and the
// right response is to act on the latest reading and report it, not to wait.
func (s *Saga) stabilise(ctx context.Context) (reconcile.Snapshot, error) {
	collector := &reconcile.Collector{
		Orders:     s.Orders,
		Positions:  s.Positions,
		Balance:    s.Balance,
		Currencies: s.currencies(),
		AccountRef: s.AccountRef,
		Clock:      s.clock(),
		MaxPages:   s.MaxPages,
	}
	stabiliser := &reconcile.Stabiliser{MinInterval: s.StabiliseInterval}

	attempts := s.StabiliseAttempts
	if attempts <= 0 {
		attempts = 4
	}

	var last reconcile.Snapshot
	for i := 0; i < attempts; i++ {
		snapshot, err := collector.Collect(ctx)
		if err != nil {
			return reconcile.Snapshot{}, fmt.Errorf("flatten: reading the account before liquidating: %w", err)
		}
		last = snapshot
		if result := stabiliser.Offer(snapshot); result.Stable {
			return snapshot, nil
		}
		if i < attempts-1 {
			interval := s.StabiliseInterval
			if interval <= 0 {
				interval = reconcile.DefaultStabilisationInterval
			}
			if err := s.clock().Sleep(ctx, interval); err != nil {
				return last, nil
			}
		}
	}
	// Unstable: act on the latest reading. The sellable-quantity read happens per
	// symbol immediately before each order anyway, which is the number that
	// actually bounds the size.
	s.logf(obs.EventFlattenProgress,
		obs.FieldDetail, "the account did not stabilise; liquidating from the latest reading")
	return last, nil
}

// remainingPositions counts the long positions the account still shows.
func (s *Saga) remainingPositions(ctx context.Context) (int, error) {
	positions, err := s.Positions.Positions(ctx)
	if err != nil {
		return 0, fmt.Errorf("flatten: re-reading the account after liquidating: %w", err)
	}
	remaining := 0
	for _, p := range positions {
		if p.Quantity > 0 {
			remaining++
		}
	}
	return remaining, nil
}

func (s *Saga) validateLiquidation() error {
	switch {
	case s.Journal == nil:
		return fmt.Errorf("%w: a journal is required", ErrNotConfigured)
	case s.Gateway == nil && !s.DryRun:
		return fmt.Errorf("%w: a gateway is required to liquidate anything", ErrNotConfigured)
	case s.Orders == nil:
		return fmt.Errorf("%w: an order pager is required for the account snapshot", ErrNotConfigured)
	case s.Positions == nil:
		return fmt.Errorf("%w: a positions reader is required — there is nothing to liquidate without one",
			ErrNotConfigured)
	case s.Balance == nil:
		return fmt.Errorf("%w: a buying-power reader is required for the account snapshot", ErrNotConfigured)
	case s.Sellable == nil:
		return fmt.Errorf("%w: a sellable-quantity reader is required; sizing a sell from the holding "+
			"alone can oversell", ErrNotConfigured)
	case strings.TrimSpace(s.AccountRef) == "":
		return fmt.Errorf("%w: an account reference is required", ErrNotConfigured)
	}
	return nil
}

func (s *Saga) currencies() []string {
	if len(s.Currencies) > 0 {
		return s.Currencies
	}
	return []string{"KRW", "USD"}
}

func currencyFor(market string) string {
	if strings.EqualFold(market, "kr") {
		return "KRW"
	}
	return "USD"
}

func collectTargets(byKey map[string]LiquidationTarget) []LiquidationTarget {
	out := make([]LiquidationTarget, 0, len(byKey))
	for _, t := range byKey {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Symbol < out[j].Symbol })
	return out
}
