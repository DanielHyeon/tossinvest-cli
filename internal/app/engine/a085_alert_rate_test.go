package engine_test

// a085 §0.4: naming a stock costs no broker request.
//
// The name arrives in the holdings response the reconciliation loop already
// takes. If it ever started costing a request, the budget the loop documents —
// 6-7 requests per 60s period — would become a function of the portfolio size,
// which is the fan-out design A6 exists to forbid.

import (
	"testing"
)

func TestNamingAStockCostsNoExtraRequest(t *testing.T) {
	h := newDriverHarness(t, nil)
	h.holds("005930", "10", "55000", 70000)
	h.holdings.items[0].Name = "삼성전자"

	before := h.holdings.calls
	if cycle := h.cycle(); cycle.Err != nil {
		t.Fatalf("cycle: %v", cycle.Err)
	}
	// Two collections per cycle is the stabilisation rule, and it is what the
	// budget already accounts for. The name rides along in both.
	if got := h.holdings.calls - before; got != 2 {
		t.Fatalf("holdings reads = %d, want the 2 the stabilisation already spends", got)
	}
}

func TestTheCycleLearnsTheNameFromTheSnapshotItAlreadyTook(t *testing.T) {
	h := newDriverHarness(t, nil)
	h.holds("042660", "2", "89500", 91800)
	h.holdings.items[0].Name = "한화오션"

	if cycle := h.cycle(); cycle.Err != nil {
		t.Fatalf("cycle: %v", cycle.Err)
	}
	if got := h.names.Label("042660"); got != "한화오션(042660)" {
		t.Fatalf("Label = %q, want 한화오션(042660): the reconciliation snapshot is where the "+
			"engine learns what a holding is called", got)
	}
}
