package costs

// costs_test.go is the port of StockOS `tests/test_costs.py` (4 cases).
//
// # What "port" means here, and what changed
//
// The *structure* of each case is preserved — which quantity is asserted, and
// what property of the cost model it pins. The *numbers* are not: StockOS's
// assertions are written against the KIS rate sheet, and trade-analytics says
// KIS 수치·`KIS_*` 명명은 이식하지 않는다(SHALL NOT). Every assertion below is
// therefore rewritten against either (a) a rate the test itself injects, so the
// expected value is arithmetic rather than a copied constant, or (b) the TossOS
// conservative placeholders in costs.go, which are `[미검증]` over-estimates
// awaiting 2b 실측.
//
// Case mapping (StockOS → TossOS):
//
//	test_buy_cost_has_commission_but_no_transaction_tax        → TestBuyCostHasCommissionButNoTransactionTax
//	test_sell_cost_includes_commission_and_transaction_tax     → TestSellCostIncludesCommissionAndTransactionTax
//	test_break_even_sell_price_is_above_buy_price_after_costs  → TestBreakEvenSellPriceIsAboveBuyPriceAfterCosts
//	test_net_pnl_subtracts_round_trip_costs                    → TestNetPnLSubtractsRoundTripCosts

import (
	"math"
	"strconv"
	"strings"
	"testing"
)

// amount reads a decimal string the way a caller would, so a malformed result
// fails the test instead of silently comparing as zero.
func amount(t *testing.T, s string) float64 {
	t.Helper()
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		t.Fatalf("result %q is not a decimal: %v", s, err)
	}
	return v
}

func closeTo(t *testing.T, got float64, want float64, tol float64, what string) {
	t.Helper()
	if math.Abs(got-want) > tol {
		t.Fatalf("%s = %v, want %v (±%v)", what, got, want, tol)
	}
}

// modelWith builds a model from overrides and fails the test if the gate
// refuses them — a test that silently fell back to defaults would assert
// nothing.
func modelWith(t *testing.T, overrides map[string]string) Model {
	t.Helper()
	m, err := NewModel(overrides)
	if err != nil {
		t.Fatalf("NewModel(%v): %v", overrides, err)
	}
	return m
}

func TestBuyCostHasCommissionButNoTransactionTax(t *testing.T) {
	// Injected rate, so the expected 100 is arithmetic (100,000 × 0.001) and
	// not a transcribed KIS number.
	m := modelWith(t, map[string]string{KeyKRBuyCommissionRate: "0.001"})

	cost, err := m.EstimateTradeCost("100000", SideBuy, MarketKR)
	if err != nil {
		t.Fatalf("EstimateTradeCost: %v", err)
	}

	closeTo(t, amount(t, cost.Commission), 100, 1e-9, "buy commission")
	if got := amount(t, cost.Tax); got != 0 {
		t.Fatalf("buy tax = %v, want 0 — the transaction tax is sell-side only", got)
	}
	closeTo(t, amount(t, cost.Total), 100, 1e-9, "buy total")
}

func TestSellCostIncludesCommissionAndTransactionTax(t *testing.T) {
	m := modelWith(t, map[string]string{
		KeyKRSellCommissionRate: "0.001",
		KeyKRSellTaxRate:        "0.002",
	})

	cost, err := m.EstimateTradeCost("100000", SideSell, MarketKR)
	if err != nil {
		t.Fatalf("EstimateTradeCost: %v", err)
	}

	closeTo(t, amount(t, cost.Commission), 100, 1e-9, "sell commission")
	closeTo(t, amount(t, cost.Tax), 200, 1e-9, "sell tax")
	closeTo(t, amount(t, cost.Total), 300, 1e-9, "sell total")
}

func TestBreakEvenSellPriceIsAboveBuyPriceAfterCosts(t *testing.T) {
	m := DefaultModel()

	price, err := m.BreakEvenSellPrice("10000", "10", MarketKR)
	if err != nil {
		t.Fatalf("BreakEvenSellPrice: %v", err)
	}
	if amount(t, price) <= 10000 {
		t.Fatalf("break-even %s is not above the buy price; costs would be free", price)
	}

	// The defining property: selling at exactly the break-even price nets zero.
	// This is what makes it "실질 본전" rather than a padded guess.
	pnl, err := m.NetPnL("10000", price, "10", MarketKR)
	if err != nil {
		t.Fatalf("NetPnL: %v", err)
	}
	closeTo(t, amount(t, pnl), 0, 1e-6, "net P&L at break-even")
}

func TestNetPnLSubtractsRoundTripCosts(t *testing.T) {
	m := DefaultModel()

	pnl, err := m.NetPnL("10000", "10100", "10", MarketKR)
	if err != nil {
		t.Fatalf("NetPnL: %v", err)
	}
	// Gross is 1,000; the round trip must cost something, so the net is
	// strictly less. The bound is the gross rather than a KIS-derived figure.
	if got := amount(t, pnl); got >= 1000 {
		t.Fatalf("net P&L = %v, want < 1000 (gross) — round-trip costs were not subtracted", got)
	}
}

// --- input validation (the part StockOS gets from Python's ValueError) -------

func TestAmountsAreRefusedRatherThanGuessed(t *testing.T) {
	m := DefaultModel()

	cases := []struct {
		name     string
		run      func() (string, error)
		contains string
	}{
		{"empty notional", func() (string, error) {
			c, err := m.EstimateTradeCost("", SideBuy, MarketKR)
			return c.Total, err
		}, "notional"},
		{"negative notional", func() (string, error) {
			c, err := m.EstimateTradeCost("-1", SideBuy, MarketKR)
			return c.Total, err
		}, "notional"},
		{"non-numeric notional", func() (string, error) {
			c, err := m.EstimateTradeCost("free", SideBuy, MarketKR)
			return c.Total, err
		}, "notional"},
		{"zero buy price", func() (string, error) {
			return m.BreakEvenSellPrice("0", "10", MarketKR)
		}, "buy price"},
		{"zero quantity", func() (string, error) {
			return m.BreakEvenSellPrice("10000", "0", MarketKR)
		}, "quantity"},
		{"unknown market", func() (string, error) {
			c, err := m.EstimateTradeCost("100", SideBuy, Market("jp"))
			return c.Total, err
		}, "market"},
		{"unknown side", func() (string, error) {
			c, err := m.EstimateTradeCost("100", Side("SHORT"), MarketKR)
			return c.Total, err
		}, "side"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.run()
			if err == nil {
				t.Fatalf("got %q with no error; an unusable input must refuse, not guess", got)
			}
			if !strings.Contains(err.Error(), tc.contains) {
				t.Fatalf("error %q does not name the offending input (%q)", err, tc.contains)
			}
		})
	}
}
