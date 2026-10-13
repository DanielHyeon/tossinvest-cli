#!/usr/bin/env python3
"""a130 의 기존 함수 편집 — 정확히 한 번 나오는 조각만 바꿈."""
import sys, pathlib
ROOT = pathlib.Path(sys.argv[1])
def sub(path, old, new):
    p = ROOT / path
    s = p.read_text()
    n = s.count(old)
    if n != 1:
        raise SystemExit(f"{path}: expected 1 occurrence, found {n}: {old[:80]!r}")
    p.write_text(s.replace(old, new))

T = "internal/journal/trade_outcomes.go"
# freezeTradeOutcomeTx
sub(T, """// It returns no error by design — see the file header. A refusal to compute is
// reported through the returned bool so the caller can say so in a log without
// being able to turn it into a rollback.
func freezeTradeOutcomeTx(ctx context.Context, tx *ApplyTx, positionID string, model costs.Model) bool {
	outcome, ok := computeTradeOutcome(ctx, tx, positionID, model)
	if !ok {
		return false
	}""", """// It returns no error by design — see the file header. A refusal to compute is
// reported through the returned reason code (a130: "" = frozen) so the caller can
// say so in a log without being able to turn it into a rollback.
func freezeTradeOutcomeTx(ctx context.Context, tx *ApplyTx, positionID string, model costs.Model) TradeOutcomeRefusalReason {
	outcome, reason := computeTradeOutcome(ctx, tx, positionID, model, nil)
	if reason != "" {
		return reason
	}""")
sub(T, """		outcome.ClosedAt, outcome.CostTotal)
	return err == nil
}""", """		outcome.ClosedAt, outcome.CostTotal)
	if err != nil {
		return OutcomeRefusalInsertFailed
	}
	return ""
}""")
# computeTradeOutcome
sub(T, """// computeTradeOutcome derives the frozen row. The bool is false when the trade
// cannot be priced honestly, which is never an error the caller may act on.
func computeTradeOutcome(ctx context.Context, r tradeOutcomeReader, positionID string,
	model costs.Model) (TradeOutcome, bool) {
	if !model.Configured() {
		// An unconfigured model prices the round trip as free, which would report
		// every trade as more profitable than it was. Recording nothing is the
		// honest outcome; the backfill can fill it in once a model exists.
		return TradeOutcome{}, false
	}""", """// computeTradeOutcome derives the frozen row. The reason is non-empty (a130: one
// of the OutcomeRefusal* codes) when the trade cannot be priced honestly, which is
// never an error the caller may act on. prices carries the broker 1-minute bars
// for an adopted position's post-adoption lots; nil (the fill transaction, and
// BackfillTradeOutcome) leaves such a position refused needs_intake_price.
func computeTradeOutcome(ctx context.Context, r tradeOutcomeReader, positionID string,
	model costs.Model, prices IntakePrices) (TradeOutcome, TradeOutcomeRefusalReason) {
	if !model.Configured() {
		// An unconfigured model prices the round trip as free, which would report
		// every trade as more profitable than it was. Recording nothing is the
		// honest outcome; the backfill can fill it in once a model exists.
		return TradeOutcome{}, OutcomeRefusalCostModelUnconfigured
	}""")
sub(T, """		  FROM positions WHERE id = ?`, positionID)
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
	}""", """		  FROM positions WHERE id = ?`, positionID)
	if err != nil {
		return TradeOutcome{}, OutcomeRefusalPositionUnreadable
	}
	if !rows.Next() {
		_ = rows.Close()
		return TradeOutcome{}, OutcomeRefusalPositionUnreadable
	}
	if err := rows.Scan(&account, &market, &symbol, &decisionID, &adoptionID,
		&openedAt, &closedAt); err != nil {
		_ = rows.Close()
		return TradeOutcome{}, OutcomeRefusalPositionUnreadable
	}""")
sub(T, """		// engine's trade.
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
	}""", """		// engine's trade.
		return TradeOutcome{}, OutcomeRefusalUnmanagedPosition
	}

	risk, level, rung, ok := frozenExitFacts(ctx, r, positionID)
	if !ok {
		return TradeOutcome{}, OutcomeRefusalExitStateUnreadable
	}

	var (
		buy, sell tradeLeg
		reason    TradeOutcomeRefusalReason
	)
	if adoptionID != "" {
		buy, sell, reason = adoptedRoundTripLegs(ctx, r, positionID, account, market, symbol, prices)
	} else if buy, sell, ok = roundTripLegs(ctx, r, account, market, symbol, decisionID); !ok {
		reason = OutcomeRefusalRoundTripUnpriceable
	}
	if reason != "" {
		return TradeOutcome{}, reason
	}
	if buy.quantity.Sign() <= 0 {
		return TradeOutcome{}, OutcomeRefusalEmptyBuyLeg
	}

	venue := costs.Market(strings.ToLower(strings.TrimSpace(market)))
	buyCost, err := model.EstimateTradeCost(ratText(buy.notional), costs.SideBuy, venue)
	if err != nil {
		return TradeOutcome{}, OutcomeRefusalCostEstimateFailed
	}
	sellCost, err := model.EstimateTradeCost(ratText(sell.notional), costs.SideSell, venue)
	if err != nil {
		return TradeOutcome{}, OutcomeRefusalCostEstimateFailed
	}""")
sub(T, """	// 실현 R = pnl ÷ (초기 위험 × 초기 수량). Both factors are frozen: the risk at
	// t0 and the quantity the position was opened with, neither of which a
	// partial take-profit moves.
	denominator""", """	// 실현 R = pnl ÷ (초기 위험 × 초기 수량). Both factors are frozen: the risk at
	// t0 and the quantity the position was opened with, neither of which a
	// partial take-profit moves. (a130: for an adopted position whose quantity
	// reconciliation raised after the adoption, buy.quantity is the adopted
	// quantity plus every priced lot — the initial quantity of this round trip,
	// which the sell leg was required to equal — so numerator and denominator
	// count the same shares; r stays the adoption-time per-share risk.)
	denominator""")
sub(T, """		ClosedAt:              closedAt,
	}, true
}""", """		ClosedAt:              closedAt,
	}, ""
}""")
# adoptedRoundTripLegs
sub(T, """// adoptedRoundTripLegs builds the synthetic buy leg and the attributed sell leg.
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
}""", """// adoptedRoundTripLegs builds the synthetic buy leg and the attributed sell leg.
//
// a130: the reason is a fixed code ("" = both legs stand). The buy leg applies
// the adoption rule (A7: the basis is the observation, not the cost basis) to
// every lot: the adopted quantity at the adoption observation, plus each
// quantity reconciliation added after the adoption at the broker's own market
// price for that moment — the close of the official 1-minute bar the
// adjustment fell in (design D1; user decision 2026-10-13, third). That price is
// a network read, so it is never fetched here: inside the fill transaction
// prices is nil and a position with such a lot is refused needs_intake_price;
// the human-run backfill (trade_outcome_backfill.go) fetches the bars outside
// any transaction and passes them in. The sell leg must equal the whole bought
// quantity, so realised R divides by the risk of the same shares the P&L
// counts. A re-adoption (the exit state's risk was rewritten), a post-adoption
// decrease (a person's disposal) and a sell outside this instance's attributed
// orders (instanceSellIntegrity) are refused. With no post-adoption increase
// the arithmetic is exactly the pre-a130 one.
func adoptedRoundTripLegs(ctx context.Context, r tradeOutcomeReader,
	positionID, account, market, symbol string, prices IntakePrices) (tradeLeg, tradeLeg, TradeOutcomeRefusalReason) {
	buy := tradeLeg{quantity: new(big.Rat), notional: new(big.Rat), priced: true}
	sell := tradeLeg{quantity: new(big.Rat), notional: new(big.Rat), priced: true}

	adopted, observed, ok := adoptedBasis(ctx, r, positionID)
	if !ok {
		return buy, sell, OutcomeRefusalAdoptionUnreadable
	}
	lots, reason := adoptionEpoch(ctx, r, positionID)
	if reason != "" {
		return buy, sell, reason
	}
	added, addedNotional, _, reason := priceLots(lots, prices)
	if reason != "" {
		return buy, sell, reason
	}
	buy.quantity = new(big.Rat).Add(adopted, added)
	buy.notional = new(big.Rat).Add(new(big.Rat).Mul(adopted, observed), addedNotional)

	orders, ok := engineSellOrders(ctx, r, positionID, account, market, symbol)
	if !ok {
		return buy, sell, OutcomeRefusalEngineSellUnreadable
	}
	if len(orders) == 0 {
		// No engine sell fill is attributable to this instance at all. Freezing
		// here would price the whole position as sold for nothing.
		return buy, sell, OutcomeRefusalNoEngineSell
	}
	if !sumSellFills(ctx, r, orders, account, market, symbol, &sell) {
		return buy, sell, OutcomeRefusalEngineSellUnreadable
	}
	if !sell.priced {
		return buy, sell, OutcomeRefusalSellUnpriced
	}
	if reason := instanceSellIntegrity(ctx, r, positionID, account, market, symbol, orders); reason != "" {
		return buy, sell, reason
	}
	if sell.quantity.Cmp(buy.quantity) != 0 {
		// Under-attributed (a person finished the disposal) or over-attributed (a
		// previous instance's flatten swept in). Either way the round trip this
		// row would describe is not the one that happened.
		return buy, sell, OutcomeRefusalSellQuantityMismatch
	}
	return buy, sell, ""
}""")
# BackfillTradeOutcome
sub(T, """	outcome, ok := computeTradeOutcome(ctx, journalReader{j: j}, id, model)
	if !ok {
		return TradeOutcome{}, fmt.Errorf(
			"%w: the round trip of %s cannot be priced from what the ledger holds", ErrInvalidRequest, id)
	}""", """	outcome, reason := computeTradeOutcome(ctx, journalReader{j: j}, id, model, nil)
	if reason != "" {
		return TradeOutcome{}, fmt.Errorf(
			"%w: the round trip of %s cannot be priced from what the ledger holds (%s)", ErrInvalidRequest, id, reason)
	}""")

A = "internal/journal/apply_hook.go"
sub(A, """	Costs costs.Model
}

// SetApplyHooks""", """	Costs costs.Model
	// OutcomeRefused receives each close whose outcome could not be frozen (a130,
	// design D3) — symbol, market, instance and a fixed reason code, never the
	// account or the position id. It is called only after the fill transaction
	// has committed (RecordFill), never for a rolled-back one; it returns
	// nothing and its panic is recovered, so it cannot fail a fill. A sink that
	// does I/O should hand it off so a blocked writer does not hold the fill
	// path. Nil, the shipped state, reads and calls nothing on the refusal path.
	OutcomeRefused func(TradeOutcomeRefusal)
}

// SetApplyHooks""")
sub(A, """	handle := &ApplyTx{tx: tx, now: fill.CommittedAt, costs: hooks.Costs}""",
       """	handle := &ApplyTx{tx: tx, now: fill.CommittedAt, costs: hooks.Costs, outcomeRefused: hooks.OutcomeRefused}""")
sub(A, """type ApplyTx struct {
	tx     *sql.Tx
	now    string
	costs  costs.Model
	closed atomic.Bool
}""", """type ApplyTx struct {
	tx     *sql.Tx
	now    string
	costs  costs.Model
	closed atomic.Bool
	// outcomeRefused is ApplyHooks.OutcomeRefused (a130). Nil collects nothing.
	outcomeRefused func(TradeOutcomeRefusal)
}""")
sub(A, """		//    data that, the position being CLOSED, cannot move any more.
		freezeTradeOutcomeTx(ctx, tx, where.PositionID, tx.costs)
	}""", """		//    data that, the position being CLOSED, cannot move any more.
		//    a130: the refusal is no longer dropped — its reason is noted on the
		//    handle and reported after the transaction commits. Noting returns
		//    nothing, so it is still unable to turn into a rollback.
		if reason := freezeTradeOutcomeTx(ctx, tx, where.PositionID, tx.costs); reason != "" {
			tx.noteOutcomeRefusal(ctx, where.PositionID, reason)
		}
	}""")

F = "internal/journal/fills.go"
sub(F, """	if locallyOwned {
		if err := j.runApplyHooks(ctx, tx, applied); err != nil {
			return res, err
		}
	}

	if err := tx.Commit(); err != nil {
		return res, fmt.Errorf("journal: committing the fill snapshot of %s: %w", orderID, err)
	}
	res.Changed = true""", """	// a130: the closes this transaction cannot freeze are collected here and
	// reported only after the commit below succeeds.
	var refusals []TradeOutcomeRefusal
	if locallyOwned {
		if err := j.runApplyHooks(withOutcomeRefusals(ctx, &refusals), tx, applied); err != nil {
			return res, err
		}
	}

	if err := tx.Commit(); err != nil {
		return res, fmt.Errorf("journal: committing the fill snapshot of %s: %w", orderID, err)
	}
	// a130: an outcome refusal is reported only for a close that committed.
	j.reportOutcomeRefusals(refusals)
	res.Changed = true""")


G = "internal/app/engine/gateway.go"
sub(G, """func bindApplyHooks(j *journal.Journal) error {
	if err := j.SetApplyHooks(journal.ApplyHooks{""", """//
// a130: the optional refused argument is the trade-outcome refusal sink
// (trade_outcome_refusal_log.go). NewContext passes the engine logger's sink;
// the package tests bind with none, which is the binding before a130.
func bindApplyHooks(j *journal.Journal, refused ...func(journal.TradeOutcomeRefusal)) error {
	var outcomeRefused func(journal.TradeOutcomeRefusal)
	if len(refused) > 0 {
		outcomeRefused = refused[0]
	}
	if err := j.SetApplyHooks(journal.ApplyHooks{""")
sub(G, """		Costs: costs.DefaultModel(),
	}); err != nil {
		return fmt.Errorf("engine: binding the fill apply hooks: %w", err)
	}""", """		Costs: costs.DefaultModel(),
		// a130: a close whose outcome cannot be frozen leaves one engine.log
		// line after its fill commits. Nil (tests) reports nothing.
		OutcomeRefused: outcomeRefused,
	}); err != nil {
		return fmt.Errorf("engine: binding the fill apply hooks: %w", err)
	}""")

E = "internal/app/engine/engine.go"
sub(E, """	if err := bindApplyHooks(jrn); err != nil {""", """	if err := bindApplyHooks(jrn, tradeOutcomeRefusalSink(opts.Logger)); err != nil {""")
print("applied")
print("applied-2")
