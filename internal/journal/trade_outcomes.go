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
		   initial_quantity, held_seconds, exit_ratchet_level, exit_rung, closed_at)
		VALUES (?,?,?,?,?,?,?,?,?)`,
		outcome.PositionID, outcome.RealizedPnLAfterCosts, outcome.RealizedR,
		outcome.InitialRisk, outcome.InitialQuantity, outcome.HeldSeconds,
		nullableString(outcome.ExitRatchetLevel), nullableRung(outcome.ExitRung),
		outcome.ClosedAt)
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
		account, market, symbol, decisionID string
		openedAt, closedAt                  string
	)
	rows, err := r.Query(ctx, `
		SELECT account_ref, market, symbol, coalesce(entry_decision_id,''),
		       coalesce(opened_at,''), coalesce(closed_at,'')
		  FROM positions WHERE id = ?`, positionID)
	if err != nil {
		return TradeOutcome{}, false
	}
	if !rows.Next() {
		_ = rows.Close()
		return TradeOutcome{}, false
	}
	if err := rows.Scan(&account, &market, &symbol, &decisionID, &openedAt, &closedAt); err != nil {
		_ = rows.Close()
		return TradeOutcome{}, false
	}
	_ = rows.Close()
	if decisionID == "" {
		// No entry decision: no stop, no initial risk, no R to express anything
		// in. An external position's round trip is not this engine's trade.
		return TradeOutcome{}, false
	}

	risk, level, rung, ok := frozenExitFacts(ctx, r, positionID)
	if !ok {
		return TradeOutcome{}, false
	}

	buy, sell, ok := roundTripLegs(ctx, r, account, market, symbol, decisionID)
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
		SELECT MIN(f.id)
		  FROM fill_events f
		  JOIN mutation_attempts a ON a.broker_order_id = f.order_id
		 WHERE a.decision_id = ?`, decisionID)
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

	// One row per fill event. The attempt join can match more than one row when
	// an order was amended, so the side is taken from the most recent attempt
	// that named the order — the same rule exitContextForFill uses.
	rows, err := r.Query(ctx, `
		SELECT f.order_id, f.cumulative_quantity, f.average_price,
		       (SELECT i.side FROM mutation_attempts a JOIN intents i ON i.id = a.intent_id
		         WHERE a.broker_order_id = f.order_id AND i.account_ref = ?
		         ORDER BY a.recorded_at DESC, a.rowid DESC LIMIT 1)
		  FROM fill_events f
		 WHERE f.id >= ? AND f.symbol = ? AND (f.market = ? OR f.market = '')
		 ORDER BY f.order_id, f.id`,
		account, low.Int64, normaliseSymbol(symbol), normaliseMarket(market))
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
			orderID, cumulative, average string
			side                         sql.NullString
		)
		if err := rows.Scan(&orderID, &cumulative, &average, &side); err != nil {
			return buy, sell, false
		}
		if !side.Valid {
			// A fill no local intent claims. It is somebody else's order on the
			// same symbol; the projection ignored it and so does this.
			continue
		}
		if orderID != currentOrder {
			currentOrder, prev = orderID, new(big.Rat)
		}

		leg := &buy
		if strings.EqualFold(strings.TrimSpace(side.String), "SELL") {
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
		       exit_rung, closed_at
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
		       coalesce(o.exit_ratchet_level, ''), o.exit_rung, o.closed_at
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
	)
	if err := rows.Scan(&o.PositionID, &o.RealizedPnLAfterCosts, &o.RealizedR,
		&o.InitialRisk, &o.InitialQuantity, &o.HeldSeconds, &o.ExitRatchetLevel,
		&rung, &o.ClosedAt); err != nil {
		return TradeOutcome{}, fmt.Errorf("journal: reading a trade outcome: %w", err)
	}
	o.ExitRung = -1
	if rung.Valid {
		o.ExitRung = int(rung.Int64)
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
		   initial_quantity, held_seconds, exit_ratchet_level, exit_rung, closed_at)
		VALUES (?,?,?,?,?,?,?,?,?)`,
		outcome.PositionID, outcome.RealizedPnLAfterCosts, outcome.RealizedR,
		outcome.InitialRisk, outcome.InitialQuantity, outcome.HeldSeconds,
		nullableString(outcome.ExitRatchetLevel), nullableRung(outcome.ExitRung),
		outcome.ClosedAt); err != nil {
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
