package journal

// trade_outcomes.go freezes what a completed trade actually did (add-core-domain
// task 8.1; trade-analytics "성과 원시 지표").
//
// # Frozen, and what that word is doing
//
// The row is written in the same transaction as the fill that closed the
// position, and nothing rewrites it afterwards. The reason is the denominator:
// realised R is `pnl ÷ (initial risk × initial quantity)`, and both of those are
// values that only exist while the position does. A later job recomputing R
// would be dividing by whatever the ledger holds *then*, which is not the number
// the decision was judged by. Aggregates (win rate, profit factor, drawdown) are
// therefore derived on read from these rows and never stored (SHALL).
//
// # Two different numbers called R
//
//	가격 R   (entry − observed) ÷ (entry − stop)  — per share, gross. The exit
//	         policy's probe. internal/exitpolicy computes it and it is not here.
//	실현 R   pnl_after_costs ÷ (initial_risk × initial_quantity) — total, net.
//
// After a partial take-profit the two disagree, permanently, and trade-analytics
// forbids giving them one name (SHALL). The column is `realized_r`; the other
// one has no column at all.
//
// # How the legs are attributed, without a heuristic
//
// The hard part is "which fills belong to this position instance". There is no
// order→instance column (D7 has none, and issues.md task 6.4 rejected adding
// one in this change), and attributing by timestamp is the guesswork the ledger
// rules forbid. What is used instead is `fill_events.id`, which is a monotone
// append sequence:
//
//	low  = the smallest fill_events.id belonging to an order of this instance's
//	       *entry decision*
//	legs = every local fill of this account/market/symbol with id ≥ low
//
// That is exact rather than approximate. Position instances of one symbol are
// disjoint by construction — a new instance opens only once the previous one is
// CLOSED, and CLOSED is terminal — so every fill of the previous instance has a
// smaller id than this instance's first entry fill. And there is no upper bound
// to get wrong, because this runs *inside* the closing transaction: no later
// fill exists yet.
//
// External fills are excluded for free: a fill no local intent claims does not
// join, and the projection did not move for it either.
//
// # Analytics must not be able to break the order path
//
// trade-analytics: 분석 작업의 실패·지연이 체결 반영·청산 처리를 지연시키거나
// 실행 상태를 되돌리…게 해서는 안 된다 (SHALL NOT). So every failure in here is
// swallowed: the position still closes, the fill still commits, and the missing
// row is recoverable by [Journal.BackfillTradeOutcome] from data that cannot
// move any more. The alternative — returning the error — would roll back a fill
// the broker has already reported, which is the one thing task 6.1 established
// must never happen for a *projection* failure and is no more acceptable for an
// analytics one.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
)

// schemaV15 records the exact total transaction cost already deducted from the
// frozen P&L. It is intentionally nullable with no default and no UPDATE:
// historical rows did not measure this value and must remain distinguishable
// from a genuinely cost-free trade. The trigger freezes every authority byte
// while leaving the documented retention DELETE path available.
const schemaV15 = `
ALTER TABLE trade_outcomes ADD COLUMN cost_total TEXT;
CREATE TRIGGER trade_outcomes_no_update
BEFORE UPDATE ON trade_outcomes
BEGIN SELECT RAISE(ABORT,'trade outcome is immutable'); END;
`

// ErrTradeOutcomeExists means the position already has a frozen outcome. It is
// what makes a backfill unable to rewrite history.
var ErrTradeOutcomeExists = errors.New("journal: the position already has a frozen trade outcome")

// ErrTradeOutcomeNotFound means no outcome has been frozen for that position.
var ErrTradeOutcomeNotFound = errors.New("journal: no trade outcome for that position")

// TradeOutcome is one completed trade's frozen record.
type TradeOutcome struct {
	PositionID string
	// RealizedPnLAfterCosts is signed: a loss is negative. Both legs' costs are
	// deducted (internal/costs).
	RealizedPnLAfterCosts string
	// RealizedR is the *realised* R multiple, which is not the exit policy's
	// price R. See the file header.
	RealizedR string
	// InitialRisk and InitialQuantity are the frozen denominator's two factors.
	InitialRisk     string
	InitialQuantity string
	// CostTotal is nil for historical rows that predate schema v15. A newly
	// frozen or backfilled row always stores buyCost.Total + sellCost.Total,
	// exactly the amount already deducted from RealizedPnLAfterCosts.
	CostTotal *string
	// HeldSeconds is opened_at → closed_at, or 0 when either is unknown.
	HeldSeconds int64
	// ExitRatchetLevel is the level reached under RATCHET; ExitRung the rung
	// index under LADDER, or -1.
	ExitRatchetLevel string
	ExitRung         int
	ClosedAt         string
}

// tradeLeg is one order's contribution to one side of the round trip.
type tradeLeg struct {
	quantity *big.Rat
	notional *big.Rat
	// priced is false when the broker reported a quantity with no price. The
	// cost basis is then unknown and no honest P&L can be stated.
	priced bool
}

// freezeTradeOutcomeTx writes the outcome of a position that has just closed.
//
// It returns no error by design — see the file header. A refusal to compute is
// reported through the returned bool so the caller can say so in a log without
// being able to turn it into a rollback.
func freezeTradeOutcomeTx(ctx context.Context, tx *ApplyTx, positionID string, model costs.Model) bool {
	outcome, ok := computeTradeOutcome(ctx, tx, positionID, model)
	if !ok {
		return false
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO trade_outcomes
		  (position_id, realized_pnl_after_costs, realized_r, initial_risk,
		   initial_quantity, held_seconds, exit_ratchet_level, exit_rung, closed_at,
		   cost_total)
		VALUES (?,?,?,?,?,?,?,?,?,?)`,
		outcome.PositionID, outcome.RealizedPnLAfterCosts, outcome.RealizedR,
		outcome.InitialRisk, outcome.InitialQuantity, outcome.HeldSeconds,
		nullableString(outcome.ExitRatchetLevel), nullableRung(outcome.ExitRung),
		outcome.ClosedAt, outcome.CostTotal)
	return err == nil
}

// tradeOutcomeReader is the read surface the computation needs, satisfied both
// by the apply handle (inside the closing transaction) and by the journal (a
// backfill afterwards).
type tradeOutcomeReader interface {
	Query(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// journalReader adapts *Journal to tradeOutcomeReader for the backfill path.
type journalReader struct{ j *Journal }

func (r journalReader) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return r.j.db.QueryContext(ctx, query, args...)
}

// computeTradeOutcome derives the frozen row. The bool is false when the trade
// cannot be priced honestly, which is never an error the caller may act on.
func computeTradeOutcome(ctx context.Context, r tradeOutcomeReader, positionID string,
	model costs.Model) (TradeOutcome, bool) {
	if !model.Configured() {
		// An unconfigured model prices the round trip as free, which would report
		// every trade as more profitable than it was. Recording nothing is the
		// honest outcome; the backfill can fill it in once a model exists.
		return TradeOutcome{}, false
	}

	var (
		account, market, symbol, decisionID, adoptionID string
		openedAt, closedAt                              string
	)
	rows, err := r.Query(ctx, `
		SELECT account_ref, market, symbol, coalesce(entry_decision_id,''),
		       coalesce(adoption_id,''), coalesce(opened_at,''), coalesce(closed_at,'')
		  FROM positions WHERE id = ?`, positionID)
	if err != nil {
		return TradeOutcome{}, false
	}
	if !rows.Next() {
		_ = rows.Close()
		return TradeOutcome{}, false
	}
	if err := rows.Scan(&account, &market, &symbol, &decisionID, &adoptionID,
		&openedAt, &closedAt); err != nil {
		_ = rows.Close()
		return TradeOutcome{}, false
	}
	_ = rows.Close()
	if decisionID == "" && adoptionID == "" {
		// Neither record justifies the position: no stop, no initial risk, no R to
		// express anything in. An unmanaged position's round trip is not this
		// engine's trade.
		return TradeOutcome{}, false
	}

	risk, level, rung, ok := frozenExitFacts(ctx, r, positionID)
	if !ok {
		return TradeOutcome{}, false
	}

	var buy, sell tradeLeg
	if adoptionID != "" {
		buy, sell, ok = adoptedRoundTripLegs(ctx, r, positionID, account, market, symbol)
	} else {
		buy, sell, ok = roundTripLegs(ctx, r, account, market, symbol, decisionID)
	}
	if !ok || buy.quantity.Sign() <= 0 {
		return TradeOutcome{}, false
	}

	venue := costs.Market(strings.ToLower(strings.TrimSpace(market)))
	buyCost, err := model.EstimateTradeCost(ratText(buy.notional), costs.SideBuy, venue)
	if err != nil {
		return TradeOutcome{}, false
	}
	sellCost, err := model.EstimateTradeCost(ratText(sell.notional), costs.SideSell, venue)
	if err != nil {
		return TradeOutcome{}, false
	}
	pnl := new(big.Rat).Sub(sell.notional, buy.notional)
	pnl.Sub(pnl, ratOf(buyCost.Total))
	pnl.Sub(pnl, ratOf(sellCost.Total))
	totalCost := ratText(new(big.Rat).Add(ratOf(buyCost.Total), ratOf(sellCost.Total)))

	// 실현 R = pnl ÷ (초기 위험 × 초기 수량). Both factors are frozen: the risk at
	// t0 and the quantity the position was opened with, neither of which a
	// partial take-profit moves.
	denominator := new(big.Rat).Mul(ratOf(risk), buy.quantity)
	realizedR := "0"
	if denominator.Sign() != 0 {
		realizedR = ratText(new(big.Rat).Quo(pnl, denominator))
	}

	return TradeOutcome{
		PositionID:            positionID,
		RealizedPnLAfterCosts: ratText(pnl),
		RealizedR:             realizedR,
		InitialRisk:           risk,
		InitialQuantity:       ratText(buy.quantity),
		CostTotal:             &totalCost,
		HeldSeconds:           heldSeconds(openedAt, closedAt),
		ExitRatchetLevel:      level,
		ExitRung:              rung,
		ClosedAt:              closedAt,
	}, true
}

// frozenExitFacts reads the exit-state values the outcome copies at close.
//
// It reads none of the four columns the apply point owns, which is what keeps
// this file outside apply_hook.go: the initial risk, the reached level and the
// reached rung are all judgement-path state.
func frozenExitFacts(ctx context.Context, r tradeOutcomeReader, positionID string) (
	risk, level string, rung int, ok bool) {
	rows, err := r.Query(ctx, `
		SELECT initial_risk, policy_kind, ratchet_level, active_rung
		  FROM exit_states WHERE position_id = ?`, positionID)
	if err != nil {
		return "", "", 0, false
	}
	defer rows.Close()
	if !rows.Next() {
		return "", "", 0, false
	}
	var (
		kind      string
		levelText string
		rungValue sql.NullInt64
	)
	if err := rows.Scan(&risk, &kind, &levelText, &rungValue); err != nil {
		return "", "", 0, false
	}
	rung = -1
	if rungValue.Valid {
		rung = int(rungValue.Int64)
	}
	// One policy per position, so exactly one of the two reached-stage columns
	// is meaningful and the other is stored NULL rather than as a default that
	// reads like a real value.
	if kind == ExitPolicyLadder {
		return risk, "", rung, true
	}
	return risk, levelText, -1, true
}

// roundTripLegs sums the marginal notional of every local fill belonging to this
// position instance, split by side.
func roundTripLegs(ctx context.Context, r tradeOutcomeReader,
	account, market, symbol, decisionID string) (tradeLeg, tradeLeg, bool) {
	buy := tradeLeg{quantity: new(big.Rat), notional: new(big.Rat), priced: true}
	sell := tradeLeg{quantity: new(big.Rat), notional: new(big.Rat), priced: true}

	var low sql.NullInt64
	bound, err := r.Query(ctx, `
		WITH order_scope_count AS (
			SELECT broker_order_id, count(*) scope_count FROM (
				SELECT DISTINCT a.broker_order_id, TRIM(i.account_ref), LOWER(TRIM(i.market)),
				       TRIM(i.trading_day), UPPER(TRIM(i.symbol)), UPPER(TRIM(i.side))
				  FROM mutation_attempts a JOIN intents i ON i.id = a.intent_id
				 WHERE a.broker_order_id <> ''
			) GROUP BY broker_order_id
		)
		SELECT MIN(f.id)
		  FROM fill_events f
		  JOIN mutation_attempts a ON a.broker_order_id = f.order_id
		  JOIN intents i ON i.id = a.intent_id
		  JOIN order_scope_count c ON c.broker_order_id = a.broker_order_id
		 WHERE a.decision_id = ? AND (
			(TRIM(f.account_ref) = TRIM(i.account_ref)
			 AND LOWER(TRIM(f.market)) = LOWER(TRIM(i.market))
			 AND TRIM(f.trading_day) = TRIM(i.trading_day)
			 AND UPPER(TRIM(f.symbol)) = UPPER(TRIM(i.symbol))
			 AND UPPER(TRIM(f.side)) = UPPER(TRIM(i.side)))
			OR (TRIM(f.account_ref) = '' AND TRIM(f.trading_day) = '' AND TRIM(f.side) = ''
			    AND c.scope_count = 1
			    AND (TRIM(f.market) = '' OR LOWER(TRIM(f.market)) = LOWER(TRIM(i.market)))
			    AND UPPER(TRIM(f.symbol)) = UPPER(TRIM(i.symbol)))
		 )`, decisionID)
	if err != nil {
		return buy, sell, false
	}
	if bound.Next() {
		_ = bound.Scan(&low)
	}
	_ = bound.Close()
	if !low.Valid {
		// The entry decision produced no fill event, so the instance was not
		// opened by it and the window has no floor to stand on.
		return buy, sell, false
	}

	// One row per fill event, attributed through its complete canonical scope.
	// Legacy rows participate only when the order id has exactly one intent
	// scope, so a reused id cannot move a later position's frozen P&L.
	rows, err := r.Query(ctx, `
		WITH order_owner AS (
			SELECT DISTINCT a.broker_order_id order_id, TRIM(i.account_ref) account_ref,
			       LOWER(TRIM(i.market)) market, TRIM(i.trading_day) trading_day,
			       UPPER(TRIM(i.symbol)) symbol, UPPER(TRIM(i.side)) side
			  FROM mutation_attempts a JOIN intents i ON i.id = a.intent_id
			 WHERE a.broker_order_id <> ''
		), order_scope_count AS (
			SELECT order_id, count(*) scope_count FROM order_owner GROUP BY order_id
		), attributed AS (
			SELECT DISTINCT f.id, f.order_id, o.account_ref, o.market, o.trading_day,
			       o.symbol, o.side, f.cumulative_quantity, f.average_price
			  FROM fill_events f
			  JOIN order_owner o ON o.order_id = f.order_id
			  JOIN order_scope_count c ON c.order_id = f.order_id
			 WHERE (
				(TRIM(f.account_ref) = o.account_ref AND LOWER(TRIM(f.market)) = o.market
				 AND TRIM(f.trading_day) = o.trading_day AND UPPER(TRIM(f.symbol)) = o.symbol
				 AND UPPER(TRIM(f.side)) = o.side)
				OR (TRIM(f.account_ref) = '' AND TRIM(f.trading_day) = '' AND TRIM(f.side) = ''
				    AND c.scope_count = 1
				    AND (TRIM(f.market) = '' OR LOWER(TRIM(f.market)) = o.market)
				    AND UPPER(TRIM(f.symbol)) = o.symbol)
			 )
		)
		SELECT order_id, account_ref, market, trading_day, symbol, side,
		       cumulative_quantity, average_price
		  FROM attributed
		 WHERE id >= ? AND account_ref = ? AND symbol = ? AND market = ?
		 ORDER BY account_ref, market, trading_day, symbol, side, order_id, id`,
		low.Int64, strings.TrimSpace(account), normaliseSymbol(symbol), normaliseMarket(market))
	if err != nil {
		return buy, sell, false
	}
	defer rows.Close()

	var (
		currentOrder string
		prev         = new(big.Rat)
	)
	for rows.Next() {
		var (
			orderID, eventAccount, eventMarket, tradingDay, eventSymbol string
			side, cumulative, average                                   string
		)
		if err := rows.Scan(&orderID, &eventAccount, &eventMarket, &tradingDay,
			&eventSymbol, &side, &cumulative, &average); err != nil {
			return buy, sell, false
		}
		orderKey := strings.Join([]string{eventAccount, eventMarket, tradingDay,
			eventSymbol, side, orderID}, "\x00")
		if orderKey != currentOrder {
			currentOrder, prev = orderKey, new(big.Rat)
		}

		leg := &buy
		if strings.EqualFold(strings.TrimSpace(side), "SELL") {
			leg = &sell
		}
		filled, ok := new(big.Rat).SetString(orZero(cumulative))
		if !ok {
			return buy, sell, false
		}
		if strings.TrimSpace(average) == "" {
			// 미관측 평균가: the contribution is genuinely unknown and calling it
			// zero would understate the basis, which is the fail-open direction
			// task 6.1 refused for the projection and refuses here too.
			leg.priced = false
			continue
		}
		price, ok := new(big.Rat).SetString(strings.TrimSpace(average))
		if !ok {
			return buy, sell, false
		}
		contribution := new(big.Rat).Mul(filled, price)
		leg.notional.Add(leg.notional, new(big.Rat).Sub(contribution, prev))
		prev = contribution

		if leg == &buy {
			// The initial quantity is what was bought, summed as the cumulative
			// quantity of each entry order rather than as a running delta: the
			// last observation of an order is its total.
			buy.quantity = filled
		}
	}
	if err := rows.Err(); err != nil {
		return buy, sell, false
	}
	if !buy.priced || !sell.priced {
		return buy, sell, false
	}
	return buy, sell, true
}

// --- the adopted round trip (change adopt-external-positions, design A7) ---------
//
// An adopted position has no entry fill, so `roundTripLegs`' whole method — find
// the smallest fill id belonging to the entry decision's orders, take everything
// after it — has no floor to stand on. Both legs are therefore built differently,
// and each difference is a decision worth stating.
//
// # The buy leg is synthesised, from the observation and not from the cost basis
//
// quantity = `position_adoptions.quantity`, basis = `observed_price`, and
// `cost_basis` appears in neither (SHALL NOT — design A7, round 3). The reason is
// epoch consistency: the denominator of realised R is the *synthetic* initial
// risk, which is `observed_price − synthetic_stop`. A numerator measured from the
// original purchase price would be dividing a whole holding period's P&L by the
// risk of one day, and the resulting number would be an R multiple of nothing.
// `cost_basis` stays on the record for display and for the 2b fee measurement,
// and is excluded from the arithmetic here.
//
// Both legs' costs are still deducted, exactly as an engine-entered round trip's
// are. That is not double-counting a purchase the engine did not make: it is the
// same round-trip cost the exit policy's BREAKEVEN level is composed from
// (internal/exitpolicy composes break-even as entry plus the round trip), so a
// position liquidated at its break-even baseline reports realised R of about
// zero rather than of the entry cost.
//
// # The sell leg is attributed by explicit reference, per proposer
//
// Two chains, both declared columns, no time-window matching (position-ledger's
// prohibition):
//
//	exit loop  exit_events.position_id → proposed_intent_id → mutation_attempts
//	           → broker_order_id → fill_events. Instance-exact by construction.
//	flatten    flatten_sagas.account_ref → flatten_steps(LIQUIDATE, symbol,
//	           market) → intent_id → mutation_attempts → broker_order_id →
//	           fill_events.
//
// The flatten chain is included because excluding it was the round-4 finding: a
// position that flatten closed would otherwise freeze with an *empty* sell leg,
// and an empty sell leg priced against a full buy leg records a fabricated total
// loss. Flatten's liquidation is a real fill and belongs in the number.
//
// # What makes the attribution safe rather than merely plausible
//
// The flatten chain names a symbol and not an instance, so in principle a
// previous instance's flatten sells could be swept in. That is caught rather than
// ignored: the freeze requires the attributed sell quantity to *equal* the
// adopted quantity, so over-attribution overshoots and under-attribution
// undershoots, and either one produces no row at all.
//
// The same rule implements the round-5 decision about the crossed case. If the
// engine sold part of the position and a person disposed of the rest outside it,
// the engine-attributed quantity is short of the adopted quantity and no outcome
// is frozen — matching a partial sell leg against a whole synthetic buy leg would
// report a loss the trade did not make. No row is the honest answer; the
// ADJUSTMENT_CLOSED event is where that story is recorded instead.

// adoptedRoundTripLegs builds the synthetic buy leg and the attributed sell leg.
func adoptedRoundTripLegs(ctx context.Context, r tradeOutcomeReader,
	positionID, account, market, symbol string) (tradeLeg, tradeLeg, bool) {
	buy := tradeLeg{quantity: new(big.Rat), notional: new(big.Rat), priced: true}
	sell := tradeLeg{quantity: new(big.Rat), notional: new(big.Rat), priced: true}

	adopted, observed, ok := adoptedBasis(ctx, r, positionID)
	if !ok {
		return buy, sell, false
	}
	buy.quantity = adopted
	buy.notional = new(big.Rat).Mul(adopted, observed)

	orders, ok := engineSellOrders(ctx, r, positionID, account, market, symbol)
	if !ok {
		return buy, sell, false
	}
	if len(orders) == 0 {
		// No engine sell fill is attributable to this instance at all. Freezing
		// here would price the whole position as sold for nothing.
		return buy, sell, false
	}
	if !sumSellFills(ctx, r, orders, account, market, symbol, &sell) {
		return buy, sell, false
	}
	if !sell.priced || sell.quantity.Cmp(adopted) != 0 {
		// Under-attributed (a person finished the disposal) or over-attributed (a
		// previous instance's flatten swept in). Either way the round trip this
		// row would describe is not the one that happened.
		return buy, sell, false
	}
	return buy, sell, true
}

// adoptedBasis reads the synthetic t0: the adopted quantity and the observation
// the baseline was built from.
func adoptedBasis(ctx context.Context, r tradeOutcomeReader, positionID string) (
	quantity, observed *big.Rat, ok bool) {
	rows, err := r.Query(ctx, `
		SELECT a.quantity, a.observed_price
		  FROM position_adoptions a
		  JOIN positions p ON p.adoption_id = a.id
		 WHERE p.id = ?`, positionID)
	if err != nil {
		return nil, nil, false
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, nil, false
	}
	var quantityText, priceText string
	if err := rows.Scan(&quantityText, &priceText); err != nil {
		return nil, nil, false
	}
	q, qok := new(big.Rat).SetString(strings.TrimSpace(quantityText))
	p, pok := new(big.Rat).SetString(strings.TrimSpace(priceText))
	if !qok || !pok || q.Sign() <= 0 || p.Sign() <= 0 {
		return nil, nil, false
	}
	return q, p, true
}

// engineSellOrders collects the broker orders this instance's sells were placed
// under, per proposer, through declared reference columns only.
func engineSellOrders(ctx context.Context, r tradeOutcomeReader,
	positionID, account, market, symbol string) ([]FillSnapshotScope, bool) {
	rows, err := r.Query(ctx, `
		SELECT DISTINCT a.broker_order_id, TRIM(i.account_ref), LOWER(TRIM(i.market)),
		       TRIM(i.trading_day), UPPER(TRIM(i.symbol)), UPPER(TRIM(i.side))
		  FROM exit_events e
		  JOIN mutation_attempts a ON a.intent_id = e.proposed_intent_id
		  JOIN intents i ON i.id = a.intent_id
		 WHERE e.position_id = ?
		   AND coalesce(e.proposed_intent_id, '') <> ''
		   AND coalesce(a.broker_order_id, '') <> ''
		UNION
		SELECT DISTINCT a.broker_order_id, TRIM(i.account_ref), LOWER(TRIM(i.market)),
		       TRIM(i.trading_day), UPPER(TRIM(i.symbol)), UPPER(TRIM(i.side))
		  FROM flatten_steps s
		  JOIN flatten_sagas g ON g.id = s.saga_id
		  JOIN mutation_attempts a ON a.intent_id = s.intent_id
		  JOIN intents i ON i.id = a.intent_id
		 WHERE g.account_ref = ?
		   AND s.kind = ?
		   AND upper(trim(s.symbol)) = ?
		   AND (lower(trim(s.market)) = ? OR trim(s.market) = '')
		   AND coalesce(s.intent_id, '') <> ''
		   AND coalesce(a.broker_order_id, '') <> ''`,
		positionID, account, FlattenStepLiquidate, normaliseSymbol(symbol), normaliseMarket(market))
	if err != nil {
		return nil, false
	}
	defer rows.Close()

	var out []FillSnapshotScope
	for rows.Next() {
		var scope FillSnapshotScope
		if err := rows.Scan(&scope.OrderID, &scope.AccountRef, &scope.Market,
			&scope.TradingDay, &scope.Symbol, &scope.Side); err != nil {
			return nil, false
		}
		out = append(out, scope)
	}
	if err := rows.Err(); err != nil {
		return nil, false
	}
	return out, true
}

// sumSellFills adds the marginal notional of every sell fill of the named
// orders.
//
// The per-order marginal arithmetic is roundTripLegs': `fill_events` carries
// cumulative quantities, so an order's contribution is the change in
// `cumulative × average`, and adding the cumulative values directly would count
// every earlier observation again.
func sumSellFills(ctx context.Context, r tradeOutcomeReader, orders []FillSnapshotScope,
	account, market, symbol string, sell *tradeLeg) bool {
	values := strings.TrimSuffix(strings.Repeat("(?,?,?,?,?,?),", len(orders)), ",")
	args := make([]any, 0, len(orders)*6+3)
	for _, scope := range orders {
		scope = canonicalFillSnapshotScope(scope)
		args = append(args, scope.AccountRef, scope.Market, scope.TradingDay,
			scope.Symbol, scope.Side, scope.OrderID)
	}
	args = append(args, strings.TrimSpace(account), normaliseSymbol(symbol), normaliseMarket(market))

	rows, err := r.Query(ctx, `
		WITH requested(account_ref, market, trading_day, symbol, side, order_id) AS (
			VALUES `+values+`
		), order_owner AS (
			SELECT DISTINCT a.broker_order_id order_id, TRIM(i.account_ref) account_ref,
			       LOWER(TRIM(i.market)) market, TRIM(i.trading_day) trading_day,
			       UPPER(TRIM(i.symbol)) symbol, UPPER(TRIM(i.side)) side
			  FROM mutation_attempts a JOIN intents i ON i.id = a.intent_id
			 WHERE a.broker_order_id <> ''
		), order_scope_count AS (
			SELECT order_id, count(*) scope_count FROM order_owner GROUP BY order_id
		)
		SELECT f.order_id, q.account_ref, q.market, q.trading_day, q.symbol, q.side,
		       f.cumulative_quantity, f.average_price
		  FROM requested q
		  JOIN order_scope_count c ON c.order_id = q.order_id
		  JOIN fill_events f ON f.order_id = q.order_id AND (
			(TRIM(f.account_ref) = q.account_ref AND LOWER(TRIM(f.market)) = q.market
			 AND TRIM(f.trading_day) = q.trading_day AND UPPER(TRIM(f.symbol)) = q.symbol
			 AND UPPER(TRIM(f.side)) = q.side)
			OR (TRIM(f.account_ref) = '' AND TRIM(f.trading_day) = '' AND TRIM(f.side) = ''
			    AND c.scope_count = 1
			    AND (TRIM(f.market) = '' OR LOWER(TRIM(f.market)) = q.market)
			    AND UPPER(TRIM(f.symbol)) = q.symbol)
		  )
		 WHERE q.account_ref = ? AND q.symbol = ? AND q.market = ?
		 ORDER BY q.account_ref, q.market, q.trading_day, q.symbol, q.side, q.order_id, f.id`,
		args...)
	if err != nil {
		return false
	}
	defer rows.Close()

	var (
		currentOrder string
		prev         = new(big.Rat)
		// perOrder holds each order's latest cumulative quantity. The rows are
		// ordered by (order, fill id), so the last one seen for an order is its
		// total — and summing the totals is the disposal's quantity, whereas
		// summing every observation would count each partial fill again.
		perOrder = map[string]*big.Rat{}
	)
	for rows.Next() {
		var (
			orderID, eventAccount, eventMarket, tradingDay, eventSymbol string
			side, cumulative, average                                   string
		)
		if err := rows.Scan(&orderID, &eventAccount, &eventMarket, &tradingDay,
			&eventSymbol, &side, &cumulative, &average); err != nil {
			return false
		}
		if !strings.EqualFold(strings.TrimSpace(side), "SELL") {
			// A buy through one of these chains would be a scale-in the exit policy
			// never proposes; it is not part of the disposal either way.
			continue
		}
		orderKey := strings.Join([]string{eventAccount, eventMarket, tradingDay,
			eventSymbol, side, orderID}, "\x00")
		if orderKey != currentOrder {
			currentOrder, prev = orderKey, new(big.Rat)
		}
		filled, ok := new(big.Rat).SetString(orZero(cumulative))
		if !ok {
			return false
		}
		perOrder[orderKey] = filled
		if strings.TrimSpace(average) == "" {
			// An unpriced sell makes the disposal's proceeds unknown. Calling it
			// zero would overstate the loss, which is the fail-open direction the
			// engine-entered path refuses too.
			sell.priced = false
			continue
		}
		price, ok := new(big.Rat).SetString(strings.TrimSpace(average))
		if !ok {
			return false
		}
		contribution := new(big.Rat).Mul(filled, price)
		sell.notional.Add(sell.notional, new(big.Rat).Sub(contribution, prev))
		prev = contribution
	}
	if err := rows.Err(); err != nil {
		return false
	}
	for _, filled := range perOrder {
		sell.quantity.Add(sell.quantity, filled)
	}
	return true
}

func heldSeconds(openedAt, closedAt string) int64 {
	opened, err := time.Parse(time.RFC3339, strings.TrimSpace(openedAt))
	if err != nil {
		return 0
	}
	closed, err := time.Parse(time.RFC3339, strings.TrimSpace(closedAt))
	if err != nil {
		return 0
	}
	if closed.Before(opened) {
		return 0
	}
	return int64(closed.Sub(opened).Seconds())
}

func ratOf(s string) *big.Rat {
	r, ok := new(big.Rat).SetString(strings.TrimSpace(s))
	if !ok {
		return new(big.Rat)
	}
	return r
}

// ratText renders a rational at a fixed scale.
//
// Twelve places, half-away-from-zero, the same convention internal/position uses
// for an average price and for the same reason: a division may not terminate,
// and the alternative to a documented scale is a value whose spelling depends on
// the arithmetic that produced it.
func ratText(r *big.Rat) string {
	if r.IsInt() {
		return r.Num().String()
	}
	out := strings.TrimRight(r.FloatString(12), "0")
	return strings.TrimSuffix(out, ".")
}

// --- reads --------------------------------------------------------------------

// TradeOutcomeOf returns one position's frozen outcome.
func (j *Journal) TradeOutcomeOf(ctx context.Context, positionID string) (TradeOutcome, error) {
	rows, err := j.db.QueryContext(ctx, `
		SELECT position_id, realized_pnl_after_costs, realized_r, initial_risk,
		       initial_quantity, coalesce(held_seconds, 0), coalesce(exit_ratchet_level, ''),
		       exit_rung, closed_at, cost_total
		  FROM trade_outcomes WHERE position_id = ?`, strings.TrimSpace(positionID))
	if err != nil {
		return TradeOutcome{}, fmt.Errorf("journal: reading the outcome of %s: %w", positionID, err)
	}
	defer rows.Close()
	if !rows.Next() {
		return TradeOutcome{}, fmt.Errorf("%w: %s", ErrTradeOutcomeNotFound, positionID)
	}
	return scanTradeOutcome(rows)
}

// TradeOutcomes returns the frozen outcomes of one account, oldest close first.
func (j *Journal) TradeOutcomes(ctx context.Context, accountRef string) ([]TradeOutcome, error) {
	rows, err := j.db.QueryContext(ctx, `
		SELECT o.position_id, o.realized_pnl_after_costs, o.realized_r, o.initial_risk,
		       o.initial_quantity, coalesce(o.held_seconds, 0),
		       coalesce(o.exit_ratchet_level, ''), o.exit_rung, o.closed_at, o.cost_total
		  FROM trade_outcomes o
		  JOIN positions p ON p.id = o.position_id
		 WHERE p.account_ref = ?
		 ORDER BY o.closed_at, o.rowid`, strings.TrimSpace(accountRef))
	if err != nil {
		return nil, fmt.Errorf("journal: listing the outcomes of %s: %w", accountRef, err)
	}
	defer rows.Close()

	var out []TradeOutcome
	for rows.Next() {
		outcome, err := scanTradeOutcome(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, outcome)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: listing the outcomes of %s: %w", accountRef, err)
	}
	return out, nil
}

func scanTradeOutcome(rows *sql.Rows) (TradeOutcome, error) {
	var (
		o    TradeOutcome
		rung sql.NullInt64
		cost sql.NullString
	)
	if err := rows.Scan(&o.PositionID, &o.RealizedPnLAfterCosts, &o.RealizedR,
		&o.InitialRisk, &o.InitialQuantity, &o.HeldSeconds, &o.ExitRatchetLevel,
		&rung, &o.ClosedAt, &cost); err != nil {
		return TradeOutcome{}, fmt.Errorf("journal: reading a trade outcome: %w", err)
	}
	o.ExitRung = -1
	if rung.Valid {
		o.ExitRung = int(rung.Int64)
	}
	if cost.Valid {
		o.CostTotal = &cost.String
	}
	return o, nil
}

// BackfillTradeOutcome writes the outcome of a closed position that has none.
//
// It exists because the freeze inside the closing transaction is allowed to give
// up silently (analytics may not break the order path), and a gap left that way
// is recoverable: CLOSED is terminal, so the fills the computation reads cannot
// move any more and the backfilled numbers are the numbers the close would have
// produced.
//
// It refuses to overwrite. "이후 비동기 작업이 원시 행을 다시 읽거나 갱신하지
// 않는다" is the freeze, and a backfill that could rewrite one would be the
// recomputation that requirement forbids.
func (j *Journal) BackfillTradeOutcome(ctx context.Context, positionID string,
	model costs.Model) (TradeOutcome, error) {
	id := strings.TrimSpace(positionID)
	if _, err := j.TradeOutcomeOf(ctx, id); err == nil {
		return TradeOutcome{}, fmt.Errorf("%w: %s", ErrTradeOutcomeExists, id)
	} else if !errors.Is(err, ErrTradeOutcomeNotFound) {
		return TradeOutcome{}, err
	}

	outcome, ok := computeTradeOutcome(ctx, journalReader{j: j}, id, model)
	if !ok {
		return TradeOutcome{}, fmt.Errorf(
			"%w: the round trip of %s cannot be priced from what the ledger holds", ErrInvalidRequest, id)
	}
	if _, err := j.db.ExecContext(ctx, `
		INSERT INTO trade_outcomes
		  (position_id, realized_pnl_after_costs, realized_r, initial_risk,
		   initial_quantity, held_seconds, exit_ratchet_level, exit_rung, closed_at,
		   cost_total)
		VALUES (?,?,?,?,?,?,?,?,?,?)`,
		outcome.PositionID, outcome.RealizedPnLAfterCosts, outcome.RealizedR,
		outcome.InitialRisk, outcome.InitialQuantity, outcome.HeldSeconds,
		nullableString(outcome.ExitRatchetLevel), nullableRung(outcome.ExitRung),
		outcome.ClosedAt, outcome.CostTotal); err != nil {
		if isUniqueViolation(err) {
			return TradeOutcome{}, fmt.Errorf("%w: %s", ErrTradeOutcomeExists, id)
		}
		return TradeOutcome{}, fmt.Errorf("journal: backfilling the outcome of %s: %w", id, err)
	}
	return outcome, nil
}

// --- aggregates, derived on read ------------------------------------------------

// TradeAggregates is the summary trade-analytics asks to be *derived* rather
// than stored (SHALL — 집계는 원시 기록에서 파생 계산한다).
type TradeAggregates struct {
	Trades int
	Wins   int
	Losses int
	// WinRate is wins ÷ trades, "" when there are no trades.
	WinRate string
	// ProfitFactor is gross profit ÷ gross loss. "" when nothing lost, because
	// dividing by zero is not "infinitely good" — it is "not yet measurable".
	ProfitFactor string
	// GrossProfit and GrossLoss are magnitudes.
	GrossProfit string
	GrossLoss   string
	// NetPnL is the signed sum.
	NetPnL string
	// MaxDrawdown is the largest peak-to-trough fall of the running net P&L, a
	// magnitude. It is over the *sequence of closes*, which is the only ordering
	// these rows carry — an intra-trade drawdown would need MFE/MAE, which
	// trade-analytics puts out of scope for want of a time series.
	MaxDrawdown string
	// SumRealizedR is the total realised R.
	SumRealizedR string
}

// AggregateTradeOutcomes derives the summary from the frozen rows.
func AggregateTradeOutcomes(outcomes []TradeOutcome) TradeAggregates {
	agg := TradeAggregates{Trades: len(outcomes)}
	var (
		profit  = new(big.Rat)
		loss    = new(big.Rat)
		net     = new(big.Rat)
		peak    = new(big.Rat)
		drawdwn = new(big.Rat)
		sumR    = new(big.Rat)
	)
	for _, o := range outcomes {
		pnl := ratOf(o.RealizedPnLAfterCosts)
		switch pnl.Sign() {
		case 1:
			agg.Wins++
			profit.Add(profit, pnl)
		case -1:
			agg.Losses++
			loss.Sub(loss, pnl) // a magnitude
		}
		net.Add(net, pnl)
		if net.Cmp(peak) > 0 {
			peak = new(big.Rat).Set(net)
		}
		if fall := new(big.Rat).Sub(peak, net); fall.Cmp(drawdwn) > 0 {
			drawdwn = fall
		}
		sumR.Add(sumR, ratOf(o.RealizedR))
	}

	agg.GrossProfit = ratText(profit)
	agg.GrossLoss = ratText(loss)
	agg.NetPnL = ratText(net)
	agg.MaxDrawdown = ratText(drawdwn)
	agg.SumRealizedR = ratText(sumR)
	if agg.Trades > 0 {
		agg.WinRate = ratText(new(big.Rat).SetFrac64(int64(agg.Wins), int64(agg.Trades)))
	}
	if loss.Sign() > 0 {
		agg.ProfitFactor = ratText(new(big.Rat).Quo(profit, loss))
	}
	return agg
}

// --- retention -------------------------------------------------------------------

// TradeOutcomeRetention is how long a frozen outcome is kept (D7: 보존 180일).
const TradeOutcomeRetention = 180 * 24 * time.Hour

// PruneTradeOutcomes deletes outcomes closed before the cutoff and reports how
// many rows went.
//
// It is a plain delete on an indexed column and it takes no other lock, which is
// what "주문 경로 트랜잭션과 경쟁하지 않는 비동기 작업" means in practice. It is
// not called from anywhere in the fill path, and it cannot be: the sweep is the
// caller's own goroutine (engine.Context.RunAnalyticsRetention).
func (j *Journal) PruneTradeOutcomes(ctx context.Context, before time.Time) (int64, error) {
	res, err := j.db.ExecContext(ctx,
		`DELETE FROM trade_outcomes WHERE closed_at < ?`, formatJournalTime(before.UTC()))
	if err != nil {
		return 0, fmt.Errorf("journal: pruning trade outcomes: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("journal: pruning trade outcomes: %w", err)
	}
	return n, nil
}
