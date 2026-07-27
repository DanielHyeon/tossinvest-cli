package flatten_test

// adopted_test.go is the §0 review test for change adopt-external-positions
// (task 3.1): flatten covers an adopted position exactly as it covers any other
// holding.
//
// # Why this needs a test rather than an argument
//
// The change's honest statement about emergency stops (design A5) is that there
// is *no switch that stops an adopted position being sold* — "pause the exits"
// would be a §0.3 violation — and that the available instruments are the
// exclusion list, `adoption.enabled=false` for new adoptions, flatten, and
// killing the process. Two of those four are only true if flatten actually
// reaches adopted positions.
//
// It does, and the reason is structural: flatten plans from the *account's
// holdings snapshot* and never consults the ledger's protection state. The
// eligibility drift test (internal/position) already proves it cannot even name
// the two eligibility columns. This is the behavioural half — a holding whose
// journal row carries an adoption is planned, sized and sold like any other,
// with nothing in the path that could learn it was adopted and treat it
// differently.

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// TestFlattenCoversAnAdoptedPosition is design A5's load-bearing half.
func TestFlattenCoversAnAdoptedPosition(t *testing.T) {
	account := &accountFake{
		positions: []domain.Position{position("005930", "kr", 10)},
		sellable:  map[string]float64{"005930": 10},
		lower:     map[string]float64{"005930": 63000},
	}
	h := newLiqHarness(t, [][]json.RawMessage{{}}, account)
	ctx := context.Background()

	// Put the holding in the projection and adopt it, the way the reconciliation
	// driver would: an external fold, then the adoption record and its pointer.
	watermark, err := h.journal.FillWatermark(ctx, "005930")
	if err != nil {
		t.Fatal(err)
	}
	fold, err := h.journal.ApplyPositionAdjustment(ctx, journal.AdjustmentRequest{
		AccountRef: "acct-7", Market: "kr", Symbol: "005930", Kind: journal.AdjustmentExternal,
		ExpectedPrevQuantity: "0", ExpectedFillWatermark: watermark,
		NewQuantity: "10", NewAvgPrice: "55000",
		BrokerAsOf: "2026-03-30T00:29:00Z",
		Evidence:   "the account holds it and no local instance explains it",
	})
	if err != nil {
		t.Fatalf("folding the holding in: %v", err)
	}
	if _, err := h.journal.AdoptPosition(ctx, journal.AdoptionRequest{
		PositionID: fold.Position.ID, Symbol: "005930", Market: "kr", Quantity: "10",
		CostBasis: "55000", ObservedPrice: "70000", SyntheticStop: "66500",
		ObservedAt: "2026-03-30T00:30:00Z",
	}); err != nil {
		t.Fatalf("AdoptPosition: %v", err)
	}
	if _, err := h.journal.OpenAdoptedExitState(ctx, fold.Position.ID); err != nil {
		t.Fatalf("OpenAdoptedExitState: %v", err)
	}
	if p, err := h.journal.LookupPosition(ctx, fold.Position.ID); err != nil {
		t.Fatal(err)
	} else if !p.Adopted() {
		t.Fatal("the fixture did not adopt the position; the test would prove nothing")
	}

	saga := h.saga(false)
	h.sell.after = account.setFlat

	if _, err := saga.CancelAll(ctx); err != nil {
		t.Fatalf("CancelAll: %v", err)
	}
	report, err := saga.Liquidate(ctx)
	if err != nil {
		t.Fatalf("Liquidate: %v", err)
	}

	orders := h.sell.orders()
	if len(orders) != 1 {
		t.Fatalf("orders = %+v, want the adopted holding liquidated like any other", orders)
	}
	if orders[0].Symbol != "005930" || orders[0].Side != "sell" || orders[0].Quantity != 10 {
		t.Errorf("order = %+v, want a full reduce-only sell of the adopted holding", orders[0])
	}
	if report.Submitted != 1 {
		t.Errorf("submitted = %d, want 1: an adopted position is a holding, and flatten flattens "+
			"holdings — there is deliberately no switch that exempts one (design A5)", report.Submitted)
	}
}
