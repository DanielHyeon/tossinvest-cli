package journal

// trade_outcomes_test.go is task 8.1: what a completed trade froze, and what
// nothing afterwards is allowed to change about it.

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
)

// outcomeFixture is a journal with both apply hooks bound and a cost model, which
// is the wiring an engine runs with since task 8.1.
func outcomeFixture(t *testing.T) *Journal {
	t.Helper()
	j := openTestJournal(t)
	if err := j.SetApplyHooks(ApplyHooks{
		Project: ProjectPosition, Exit: ApplyExitFill, Costs: costs.DefaultModel(),
	}); err != nil {
		t.Fatalf("SetApplyHooks: %v", err)
	}
	return j
}

// roundTrip opens a position, runs it through the exit state, sells all of it and
// returns the position id.
func roundTrip(t *testing.T, j *Journal, quantity, buyPrice, sellPrice string) string {
	t.Helper()
	ctx := context.Background()

	buy := place(t, j, order{
		intentID: "i-buy", attemptID: "a-buy", orderID: "o-buy",
		decisionID: "d-buy", side: "BUY", quantity: quantity,
	})
	if _, err := j.RecordFill(ctx, terminalFill(buy, quantity, buyPrice)); err != nil {
		t.Fatalf("RecordFill(buy): %v", err)
	}
	p := currentPosition(t, j, buy)
	if _, err := j.OpenExitState(ctx, ExitStateSeed{
		PositionID: p.ID, EntryPrice: buyPrice, InitialStop: "68000",
	}); err != nil {
		t.Fatalf("OpenExitState: %v", err)
	}

	sell := place(t, j, order{
		intentID: "i-sell", attemptID: "a-sell", orderID: "o-sell",
		decisionID: "d-sell", side: "SELL", quantity: quantity,
	})
	if _, err := j.RecordFill(ctx, terminalFill(sell, quantity, sellPrice)); err != nil {
		t.Fatalf("RecordFill(sell): %v", err)
	}
	return p.ID
}

func outcomeOf(t *testing.T, j *Journal, positionID string) TradeOutcome {
	t.Helper()
	o, err := j.TradeOutcomeOf(context.Background(), positionID)
	if err != nil {
		t.Fatalf("TradeOutcomeOf: %v", err)
	}
	return o
}

// TestTheOutcomeIsFrozenInTheClosingTransaction is the scenario: CLOSED and the
// record arrive together, or neither does.
func TestTheOutcomeIsFrozenInTheClosingTransaction(t *testing.T) {
	j := outcomeFixture(t)
	id := roundTrip(t, j, "10", "70000", "72000")

	got := outcomeOf(t, j, id)
	if got.InitialQuantity != "10" {
		t.Errorf("initial quantity = %s, want the quantity bought", got.InitialQuantity)
	}
	if got.InitialRisk != "2000" {
		t.Errorf("initial risk = %s, want the frozen entry − stop", got.InitialRisk)
	}
	if got.ClosedAt == "" {
		t.Error("the outcome carries the close it was frozen at")
	}
	// 20,000 gross on the round trip, less both legs' costs, so strictly between
	// zero and the gross. The exact figure is the cost model's and is tested
	// there; what is tested here is that costs were deducted at all.
	if !strings.HasPrefix(got.RealizedPnLAfterCosts, "1") {
		t.Errorf("pnl = %s, want a bit under the 20000 gross", got.RealizedPnLAfterCosts)
	}
	gross, _ := ratOf("20000").Float64()
	net, _ := ratOf(got.RealizedPnLAfterCosts).Float64()
	if net >= gross {
		t.Errorf("pnl = %s, which is not below the %v gross; the costs were not deducted",
			got.RealizedPnLAfterCosts, gross)
	}
}

// TestRealizedRIsNotThePriceR is the naming separation trade-analytics demands.
// After a partial the two numbers differ, and the row must carry the realised
// one: total, net of costs, over the frozen denominator.
func TestRealizedRIsNotThePriceR(t *testing.T) {
	j := outcomeFixture(t)
	ctx := context.Background()

	buy := place(t, j, order{
		intentID: "i-buy", attemptID: "a-buy", orderID: "o-buy",
		decisionID: "d-buy", side: "BUY", quantity: "10",
	})
	if _, err := j.RecordFill(ctx, terminalFill(buy, "10", "70000")); err != nil {
		t.Fatalf("RecordFill(buy): %v", err)
	}
	p := currentPosition(t, j, buy)
	if _, err := j.OpenExitState(ctx, ExitStateSeed{
		PositionID: p.ID, EntryPrice: "70000", InitialStop: "68000",
	}); err != nil {
		t.Fatalf("OpenExitState: %v", err)
	}

	// 4 taken at +1R, the rest at break-even. Price R at the close is 0; realised
	// R is positive, because 4 shares were sold 2000 higher.
	first := place(t, j, order{
		intentID: "i-tp", attemptID: "a-tp", orderID: "o-tp",
		decisionID: "d-tp", side: "SELL", quantity: "4",
	})
	if _, err := j.RecordFill(ctx, terminalFill(first, "4", "72000")); err != nil {
		t.Fatalf("RecordFill(partial): %v", err)
	}
	rest := place(t, j, order{
		intentID: "i-rest", attemptID: "a-rest", orderID: "o-rest",
		decisionID: "d-rest", side: "SELL", quantity: "6",
	})
	if _, err := j.RecordFill(ctx, terminalFill(rest, "6", "70000")); err != nil {
		t.Fatalf("RecordFill(rest): %v", err)
	}

	got := outcomeOf(t, j, p.ID)
	if got.InitialQuantity != "10" {
		t.Fatalf("initial quantity = %s; the denominator is the quantity bought, not what was left",
			got.InitialQuantity)
	}
	r, _ := ratOf(got.RealizedR).Float64()
	// pnl ÷ (2000 × 10) = pnl ÷ 20000. The gross is 8000, so R is well under 1
	// even though the price R at the take-profit was exactly 1.
	if r <= 0 || r >= 1 {
		t.Errorf("realized R = %s, want a positive fraction; the price R at the partial was 1.0", got.RealizedR)
	}
}

// TestTheReachedExitStageIsCopiedAtTheClose: one policy per position, so exactly
// one of the two stage columns is meaningful.
func TestTheReachedExitStageIsCopiedAtTheClose(t *testing.T) {
	j := outcomeFixture(t)
	ctx := context.Background()

	buy := place(t, j, order{
		intentID: "i-buy", attemptID: "a-buy", orderID: "o-buy",
		decisionID: "d-buy", side: "BUY", quantity: "10",
	})
	if _, err := j.RecordFill(ctx, terminalFill(buy, "10", "70000")); err != nil {
		t.Fatal(err)
	}
	p := currentPosition(t, j, buy)
	if _, err := j.OpenExitState(ctx, ExitStateSeed{
		PositionID: p.ID, EntryPrice: "70000", InitialStop: "68000",
	}); err != nil {
		t.Fatal(err)
	}
	if err := j.RecordExitJudgement(ctx, ExitJudgement{
		PositionID: p.ID, HighWater: "72400", Baseline: "70600",
		RatchetLevel: RatchetPartialLock, ActiveRung: -1,
	}); err != nil {
		t.Fatal(err)
	}

	sell := place(t, j, order{
		intentID: "i-sell", attemptID: "a-sell", orderID: "o-sell",
		decisionID: "d-sell", side: "SELL", quantity: "10",
	})
	if _, err := j.RecordFill(ctx, terminalFill(sell, "10", "70600")); err != nil {
		t.Fatal(err)
	}

	got := outcomeOf(t, j, p.ID)
	if got.ExitRatchetLevel != RatchetPartialLock {
		t.Errorf("exit level = %q, want the level reached", got.ExitRatchetLevel)
	}
	if got.ExitRung != -1 {
		t.Errorf("exit rung = %d, want none under RATCHET", got.ExitRung)
	}
}

// TestAnAnalyticsFailureDoesNotRollBackTheClose is the isolation requirement.
// A journal with no cost model cannot price the round trip; the position must
// close anyway.
func TestAnAnalyticsFailureDoesNotRollBackTheClose(t *testing.T) {
	j := openTestJournal(t)
	if err := j.SetApplyHooks(ApplyHooks{Project: ProjectPosition, Exit: ApplyExitFill}); err != nil {
		t.Fatalf("SetApplyHooks: %v", err)
	}
	id := roundTrip(t, j, "10", "70000", "72000")

	p, err := j.LookupPosition(context.Background(), id)
	if err != nil {
		t.Fatalf("LookupPosition: %v", err)
	}
	if p.State != PositionClosed {
		t.Fatalf("position = %s, want CLOSED — analytics may not roll back a fill", p.State)
	}
	state, err := j.ExitState(context.Background(), id)
	if err != nil {
		t.Fatalf("ExitState: %v", err)
	}
	if !state.Completed {
		t.Error("the exit state must still have completed")
	}
	if _, err := j.TradeOutcomeOf(context.Background(), id); !errors.Is(err, ErrTradeOutcomeNotFound) {
		t.Errorf("err = %v, want no outcome rather than a wrong one", err)
	}
}

// TestABackfillRecoversTheGapAndRefusesToRewriteIt is the other half: the gap
// left above is recoverable, because CLOSED is terminal and the fills it reads
// cannot move any more — and a backfill that could overwrite a frozen row would
// be the recomputation the freeze exists to prevent.
func TestABackfillRecoversTheGapAndRefusesToRewriteIt(t *testing.T) {
	j := openTestJournal(t)
	if err := j.SetApplyHooks(ApplyHooks{Project: ProjectPosition, Exit: ApplyExitFill}); err != nil {
		t.Fatalf("SetApplyHooks: %v", err)
	}
	ctx := context.Background()
	id := roundTrip(t, j, "10", "70000", "72000")

	filled, err := j.BackfillTradeOutcome(ctx, id, costs.DefaultModel())
	if err != nil {
		t.Fatalf("BackfillTradeOutcome: %v", err)
	}
	if filled.InitialQuantity != "10" {
		t.Errorf("backfilled = %+v, want the same numbers the close would have frozen", filled)
	}
	if _, err := j.BackfillTradeOutcome(ctx, id, costs.DefaultModel()); !errors.Is(err, ErrTradeOutcomeExists) {
		t.Fatalf("err = %v, want ErrTradeOutcomeExists — a frozen row is not rewritten", err)
	}
}

// TestAnExternalPositionHasNoOutcome: no entry decision, no initial risk, no R.
// A round trip somebody else made is not this engine's trade.
func TestAnExternalPositionHasNoOutcome(t *testing.T) {
	j := outcomeFixture(t)
	ctx := context.Background()

	o := place(t, j, order{intentID: "i-x", attemptID: "a-x", orderID: "o-x", side: "BUY", quantity: "5"})
	if _, err := j.RecordFill(ctx, terminalFill(o, "5", "70000")); err != nil {
		t.Fatal(err)
	}
	p := currentPosition(t, j, o)
	sell := place(t, j, order{intentID: "i-xs", attemptID: "a-xs", orderID: "o-xs", side: "SELL", quantity: "5"})
	if _, err := j.RecordFill(ctx, terminalFill(sell, "5", "71000")); err != nil {
		t.Fatal(err)
	}
	if _, err := j.TradeOutcomeOf(ctx, p.ID); !errors.Is(err, ErrTradeOutcomeNotFound) {
		t.Errorf("err = %v, want no outcome for a position with no entry decision", err)
	}
}

// TestAnUnpricedFillLeavesNoOutcome keeps the direction honest: an unobserved
// average price makes the basis unknown, and an unknown basis reported as zero
// would claim a profit nobody made.
func TestAnUnpricedFillLeavesNoOutcome(t *testing.T) {
	j := outcomeFixture(t)
	ctx := context.Background()

	buy := place(t, j, order{
		intentID: "i-buy", attemptID: "a-buy", orderID: "o-buy",
		decisionID: "d-buy", side: "BUY", quantity: "10",
	})
	if _, err := j.RecordFill(ctx, terminalFill(buy, "10", "")); err != nil {
		t.Fatal(err)
	}
	p := currentPosition(t, j, buy)
	if _, err := j.OpenExitState(ctx, ExitStateSeed{
		PositionID: p.ID, EntryPrice: "70000", InitialStop: "68000",
	}); err != nil {
		t.Fatal(err)
	}
	sell := place(t, j, order{
		intentID: "i-sell", attemptID: "a-sell", orderID: "o-sell",
		decisionID: "d-sell", side: "SELL", quantity: "10",
	})
	if _, err := j.RecordFill(ctx, terminalFill(sell, "10", "72000")); err != nil {
		t.Fatal(err)
	}
	if _, err := j.TradeOutcomeOf(ctx, p.ID); !errors.Is(err, ErrTradeOutcomeNotFound) {
		t.Errorf("err = %v, want no outcome when the basis is unknown", err)
	}
}

// --- aggregates ------------------------------------------------------------------

// TestAggregatesAreDerivedFromTheRows is the SHALL: the summary is computed on
// read and nothing stores it.
func TestAggregatesAreDerivedFromTheRows(t *testing.T) {
	agg := AggregateTradeOutcomes([]TradeOutcome{
		{RealizedPnLAfterCosts: "300", RealizedR: "1.5"},
		{RealizedPnLAfterCosts: "-100", RealizedR: "-0.5"},
		{RealizedPnLAfterCosts: "-100", RealizedR: "-0.5"},
		{RealizedPnLAfterCosts: "100", RealizedR: "0.5"},
	})
	if agg.Trades != 4 || agg.Wins != 2 || agg.Losses != 2 {
		t.Fatalf("counts = %+v", agg)
	}
	if agg.WinRate != "0.5" {
		t.Errorf("win rate = %s, want 0.5", agg.WinRate)
	}
	if agg.GrossProfit != "400" || agg.GrossLoss != "200" {
		t.Errorf("gross = %s/%s, want 400/200", agg.GrossProfit, agg.GrossLoss)
	}
	if agg.ProfitFactor != "2" {
		t.Errorf("profit factor = %s, want 2", agg.ProfitFactor)
	}
	if agg.NetPnL != "200" {
		t.Errorf("net = %s, want 200", agg.NetPnL)
	}
	// The running net is 300, 200, 100, 200: the peak is 300 and the trough 100.
	if agg.MaxDrawdown != "200" {
		t.Errorf("max drawdown = %s, want 200", agg.MaxDrawdown)
	}
	if agg.SumRealizedR != "1" {
		t.Errorf("sum R = %s, want 1", agg.SumRealizedR)
	}
}

// TestAProfitFactorWithNoLossesIsUnmeasuredRatherThanInfinite: dividing by zero
// is not "infinitely good".
func TestAProfitFactorWithNoLossesIsUnmeasuredRatherThanInfinite(t *testing.T) {
	agg := AggregateTradeOutcomes([]TradeOutcome{{RealizedPnLAfterCosts: "100", RealizedR: "1"}})
	if agg.ProfitFactor != "" {
		t.Errorf("profit factor = %q, want empty", agg.ProfitFactor)
	}
	if agg.WinRate != "1" {
		t.Errorf("win rate = %q, want 1", agg.WinRate)
	}
}

func TestAggregatesOfNothingAreNotDivisions(t *testing.T) {
	agg := AggregateTradeOutcomes(nil)
	if agg.Trades != 0 || agg.WinRate != "" || agg.ProfitFactor != "" {
		t.Errorf("agg = %+v, want an empty summary and no division", agg)
	}
}

// --- retention ---------------------------------------------------------------------

// TestRetentionDeletesOnlyWhatIsPastTheHorizon is the 180-day policy. The sweep
// is an ordinary delete on the indexed close column and takes no other lock,
// which is what keeps it out of the order path's way.
func TestRetentionDeletesOnlyWhatIsPastTheHorizon(t *testing.T) {
	j := outcomeFixture(t)
	ctx := context.Background()
	id := roundTrip(t, j, "10", "70000", "72000")
	if _, err := j.TradeOutcomeOf(ctx, id); err != nil {
		t.Fatalf("the fixture must have frozen an outcome: %v", err)
	}

	// The fixture's clock is 2026-03-30; a horizon before that keeps the row.
	kept, err := j.PruneTradeOutcomes(ctx, time.Date(2026, 3, 29, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("PruneTradeOutcomes: %v", err)
	}
	if kept != 0 {
		t.Fatalf("pruned %d rows inside the horizon, want 0", kept)
	}
	if _, err := j.TradeOutcomeOf(ctx, id); err != nil {
		t.Fatalf("the row must survive: %v", err)
	}

	gone, err := j.PruneTradeOutcomes(ctx, time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("PruneTradeOutcomes: %v", err)
	}
	if gone != 1 {
		t.Fatalf("pruned %d rows past the horizon, want 1", gone)
	}
	if _, err := j.TradeOutcomeOf(ctx, id); !errors.Is(err, ErrTradeOutcomeNotFound) {
		t.Errorf("err = %v, want the row gone", err)
	}
}

func TestTheRetentionHorizonIsAHundredAndEightyDays(t *testing.T) {
	if TradeOutcomeRetention != 180*24*time.Hour {
		t.Fatalf("retention = %v, want the 180 days D7 specifies", TradeOutcomeRetention)
	}
}

// TestTradeOutcomesAreScopedToTheAccount: a multi-account journal must not
// aggregate somebody else's trades into this account's win rate.
func TestTradeOutcomesAreScopedToTheAccount(t *testing.T) {
	j := outcomeFixture(t)
	ctx := context.Background()
	roundTrip(t, j, "10", "70000", "72000")

	mine, err := j.TradeOutcomes(ctx, "acct-1")
	if err != nil {
		t.Fatalf("TradeOutcomes: %v", err)
	}
	if len(mine) != 1 {
		t.Fatalf("outcomes = %+v, want the one trade", mine)
	}
	theirs, err := j.TradeOutcomes(ctx, "acct-other")
	if err != nil {
		t.Fatalf("TradeOutcomes: %v", err)
	}
	if len(theirs) != 0 {
		t.Errorf("another account's outcomes = %+v, want none", theirs)
	}
}
