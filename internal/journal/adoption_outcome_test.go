package journal

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

// adoption_outcome_test.go is the freeze branch for an adopted position (change
// adopt-external-positions task 1.6; design A7).
//
// The three rules being pinned, in the order they matter:
//
//  1. an engine-liquidated adopted position *does* get a row. Before this change
//     the freeze returned early on an empty entry decision, so a position the
//     engine protected and sold left no trace in any aggregate.
//  2. the basis is `observed_price` and never `cost_basis` — numerator and
//     denominator have to share a t0.
//  3. an empty or incomplete sell leg produces *no row*, never a fabricated
//     loss.

// adoptedHolding folds an external holding into the projection, adopts it and
// opens its exit state. It returns the position id.
func adoptedHolding(t *testing.T, j *Journal, quantity, costBasis, observed, stop string) string {
	t.Helper()
	ctx := context.Background()

	watermark, err := j.FillWatermark(ctx, "005930")
	if err != nil {
		t.Fatal(err)
	}
	fold, err := j.ApplyPositionAdjustment(ctx, AdjustmentRequest{
		AccountRef: "acct-1", Market: "kr", Symbol: "005930", Kind: AdjustmentExternal,
		ExpectedPrevQuantity: "0", ExpectedFillWatermark: watermark,
		NewQuantity: quantity, NewAvgPrice: costBasis,
		BrokerAsOf: "2026-03-30T00:29:00Z",
		Evidence:   "the account holds it and no local instance explains it",
	})
	if err != nil {
		t.Fatalf("folding the holding in: %v", err)
	}
	if _, err := j.AdoptPosition(ctx, AdoptionRequest{
		PositionID: fold.Position.ID, Symbol: "005930", Market: "kr",
		Quantity: quantity, CostBasis: costBasis,
		ObservedPrice: observed, SyntheticStop: stop,
		ObservedAt: "2026-03-30T00:30:00Z",
	}); err != nil {
		t.Fatalf("AdoptPosition: %v", err)
	}
	if _, err := j.OpenAdoptedExitState(ctx, fold.Position.ID); err != nil {
		t.Fatalf("OpenAdoptedExitState: %v", err)
	}
	return fold.Position.ID
}

// exitLoopSell places a sell the exit policy proposed: the judgement names the
// intent, which is the declared reference the freeze attributes the fill by.
func exitLoopSell(t *testing.T, j *Journal, positionID, intentID, attemptID, orderID,
	quantity, price string) {
	t.Helper()
	ctx := context.Background()

	sell := place(t, j, order{
		intentID: intentID, attemptID: attemptID, orderID: orderID,
		decisionID: "d-" + intentID, side: "SELL", quantity: quantity,
	})
	state, err := j.ExitState(ctx, positionID)
	if err != nil {
		t.Fatal(err)
	}
	if err := j.RecordExitJudgement(ctx, ExitJudgement{
		PositionID: positionID, ObservedPrice: price,
		HighWater: state.HighWater, Baseline: state.Baseline,
		RatchetLevel: state.RatchetLevel, ActiveRung: exitpolicy.NoRung,
		Proposal: &ExitProposal{
			Action: string(exitpolicy.ActionBaselineBreach), Level: RatchetNone, IntentID: intentID,
		},
	}); err != nil {
		t.Fatalf("RecordExitJudgement: %v", err)
	}
	if _, err := j.RecordFill(ctx, terminalFill(sell, quantity, price)); err != nil {
		t.Fatalf("RecordFill(%s): %v", orderID, err)
	}
}

// flattenLiquidation records the saga and the LIQUIDATE step that names the
// sell intent. That step is the declared reference the freeze attributes a
// flatten fill by — there is no time-window matching anywhere in the chain.
func flattenLiquidation(t *testing.T, j *Journal, sagaID string, o order) {
	t.Helper()
	ctx := context.Background()
	if _, err := j.StartFlatten(ctx, FlattenSaga{
		ID: sagaID, AccountRef: o.account, Reason: "test", Operator: "operator",
	}); err != nil {
		t.Fatalf("StartFlatten: %v", err)
	}
	step, err := j.AddFlattenStep(ctx, FlattenStep{
		SagaID: sagaID, Kind: FlattenStepLiquidate, Market: o.market, Symbol: o.symbol,
		Side: "SELL", Quantity: o.quantity,
	})
	if err != nil {
		t.Fatalf("AddFlattenStep: %v", err)
	}
	if err := j.UpdateFlattenStep(ctx, step.ID, FlattenStepDone,
		o.intentID, o.attemptID, "", "liquidated"); err != nil {
		t.Fatalf("UpdateFlattenStep: %v", err)
	}
}

// TestAnAdoptedPositionTheEngineLiquidatedFreezesAnOutcome is the headline: the
// row exists, and its denominator is the synthetic risk.
func TestAnAdoptedPositionTheEngineLiquidatedFreezesAnOutcome(t *testing.T) {
	j := outcomeFixture(t)
	// Bought long ago at 55,000; adopted at 70,000 with a 5 % synthetic stop.
	id := adoptedHolding(t, j, "10", "55000", "70000", "66500")

	exitLoopSell(t, j, id, "i-exit", "a-exit", "o-exit", "10", "72000")

	got := outcomeOf(t, j, id)
	if got.InitialQuantity != "10" {
		t.Errorf("initial quantity = %s, want the adopted quantity", got.InitialQuantity)
	}
	if got.InitialRisk != "3500" {
		t.Errorf("initial risk = %s, want the synthetic 70000 − 66500", got.InitialRisk)
	}

	// The basis is the observation, not the cost basis. Against 70,000 the gross
	// is 10 × 2,000 = 20,000; against the 55,000 cost basis it would have been
	// 170,000, and dividing *that* by one day's risk would be an R multiple of
	// nothing.
	net, _ := ratOf(got.RealizedPnLAfterCosts).Float64()
	if net <= 0 || net >= 20000 {
		t.Errorf("pnl = %s, want a bit under the 20000 gross measured from the observation "+
			"(the cost basis must not appear in the formula at all)", got.RealizedPnLAfterCosts)
	}
	// Realised R = pnl ÷ (3500 × 10) = a bit under 0.571.
	r, _ := ratOf(got.RealizedR).Float64()
	if r <= 0 || r >= 0.5715 {
		t.Errorf("realised R = %s, want a bit under 20000 ÷ 35000", got.RealizedR)
	}
}

// TestAdoptedOutcomeIgnoresTheCostBasisEntirely is the same rule stated so it
// cannot be satisfied by accident: two positions identical except for what the
// broker said they cost freeze the same numbers.
func TestAdoptedOutcomeIgnoresTheCostBasisEntirely(t *testing.T) {
	cheap := outcomeFixture(t)
	dear := outcomeFixture(t)

	cheapID := adoptedHolding(t, cheap, "10", "55000", "70000", "66500")
	dearID := adoptedHolding(t, dear, "10", "90000", "70000", "66500")
	exitLoopSell(t, cheap, cheapID, "i-exit", "a-exit", "o-exit", "10", "72000")
	exitLoopSell(t, dear, dearID, "i-exit", "a-exit", "o-exit", "10", "72000")

	a, b := outcomeOf(t, cheap, cheapID), outcomeOf(t, dear, dearID)
	if a.RealizedPnLAfterCosts != b.RealizedPnLAfterCosts || a.RealizedR != b.RealizedR {
		t.Errorf("a 55000 cost basis froze %s/%sR and a 90000 one froze %s/%sR; cost_basis is "+
			"record-only and must not reach the arithmetic",
			a.RealizedPnLAfterCosts, a.RealizedR, b.RealizedPnLAfterCosts, b.RealizedR)
	}
}

// TestAdoptedOutcomeIncludesFlattenSells is the round-4 finding: attributing
// only through the exit-event chain leaves a flatten-closed position with an
// empty sell leg, and an empty sell leg against a whole buy leg records a
// fabricated total loss.
func TestAdoptedOutcomeIncludesFlattenSells(t *testing.T) {
	j := outcomeFixture(t)
	ctx := context.Background()
	id := adoptedHolding(t, j, "10", "55000", "70000", "66500")

	sell := place(t, j, order{
		intentID: "i-flat", attemptID: "a-flat", orderID: "o-flat",
		decisionID: "d-flat", side: "SELL", quantity: "10",
	})
	flattenLiquidation(t, j, "saga-1", sell)
	if _, err := j.RecordFill(ctx, terminalFill(sell, "10", "72000")); err != nil {
		t.Fatalf("RecordFill: %v", err)
	}

	got := outcomeOf(t, j, id)
	net, _ := ratOf(got.RealizedPnLAfterCosts).Float64()
	if net <= 0 {
		t.Errorf("pnl = %s; a flatten liquidation is a real fill and belongs in the number, and "+
			"pricing it as an empty sell leg would report a total loss", got.RealizedPnLAfterCosts)
	}
}

// TestAnAdoptedPositionWithNoEngineSellFreezesNothing is the hard stop. The
// position went to zero and the exit state is completed, but nothing the engine
// did explains where the shares went — so there is no row, rather than one
// claiming the position was sold for nothing.
func TestAnAdoptedPositionWithNoEngineSellFreezesNothing(t *testing.T) {
	j := outcomeFixture(t)
	ctx := context.Background()
	id := adoptedHolding(t, j, "10", "55000", "70000", "66500")

	// Somebody sold it by hand; reconciliation converges the projection to zero.
	watermark, err := j.FillWatermark(ctx, "005930")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := j.ApplyPositionAdjustment(ctx, AdjustmentRequest{
		AccountRef: "acct-1", Market: "kr", Symbol: "005930", Kind: AdjustmentExternal,
		ExpectedPrevQuantity: "10", ExpectedFillWatermark: watermark, NewQuantity: "0",
		BrokerAsOf: "2026-03-30T01:00:00Z", Evidence: "the account no longer holds it",
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := j.TradeOutcomeOf(ctx, id); !errors.Is(err, ErrTradeOutcomeNotFound) {
		t.Fatalf("TradeOutcomeOf: %v, want ErrTradeOutcomeNotFound", err)
	}
	// And the backfill must refuse for the same reason, rather than being the
	// back door that writes the fabricated row.
	if _, err := j.BackfillTradeOutcome(ctx, id, costs.DefaultModel()); err == nil {
		t.Error("BackfillTradeOutcome wrote a row for a disposal the engine cannot account for")
	}
}

// TestAPartlyEngineSoldAdoptedPositionFreezesNothing is the round-5 crossed
// case: the engine took 40 % and a person disposed of the rest. Matching a
// partial sell leg against a whole synthetic buy leg would report a loss the
// trade did not make, so the default is no row.
func TestAPartlyEngineSoldAdoptedPositionFreezesNothing(t *testing.T) {
	j := outcomeFixture(t)
	ctx := context.Background()
	id := adoptedHolding(t, j, "10", "55000", "70000", "66500")

	exitLoopSell(t, j, id, "i-tp", "a-tp", "o-tp", "4", "72000")

	watermark, err := j.FillWatermark(ctx, "005930")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := j.ApplyPositionAdjustment(ctx, AdjustmentRequest{
		AccountRef: "acct-1", Market: "kr", Symbol: "005930", Kind: AdjustmentExternal,
		ExpectedPrevQuantity: "6", ExpectedFillWatermark: watermark, NewQuantity: "0",
		BrokerAsOf: "2026-03-30T01:00:00Z", Evidence: "the owner sold the remainder",
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := j.TradeOutcomeOf(ctx, id); !errors.Is(err, ErrTradeOutcomeNotFound) {
		t.Fatalf("TradeOutcomeOf: %v, want ErrTradeOutcomeNotFound", err)
	}
	if _, err := j.BackfillTradeOutcome(ctx, id, costs.DefaultModel()); err == nil {
		t.Error("BackfillTradeOutcome priced a 40 % disposal against a 100 % buy leg")
	}
}

// TestAdoptedOutcomeSumsEveryEngineSell keeps the partial path honest: a
// take-profit and the final liquidation are both this instance's, and the row is
// frozen from the two together.
func TestAdoptedOutcomeSumsEveryEngineSell(t *testing.T) {
	j := outcomeFixture(t)
	id := adoptedHolding(t, j, "10", "55000", "70000", "66500")

	exitLoopSell(t, j, id, "i-tp", "a-tp", "o-tp", "4", "73500")
	exitLoopSell(t, j, id, "i-exit", "a-exit", "o-exit", "6", "71000")

	got := outcomeOf(t, j, id)
	if got.InitialQuantity != "10" {
		t.Errorf("initial quantity = %s, want the adopted quantity", got.InitialQuantity)
	}
	// Gross = 4 × 3500 + 6 × 1000 = 20,000, measured from the observation.
	gross := new(big.Rat).SetInt64(20000)
	net := ratOf(got.RealizedPnLAfterCosts)
	if net.Sign() <= 0 || net.Cmp(gross) >= 0 {
		t.Errorf("pnl = %s, want a bit under the 20000 gross of both sells together",
			got.RealizedPnLAfterCosts)
	}
}
