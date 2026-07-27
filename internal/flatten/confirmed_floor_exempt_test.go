package flatten_test

import (
	"context"
	"encoding/json"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
)

// confirmed_floor_exempt_test.go is the §0.3 regression for
// extend-execution-contract task 4.2.
//
// The change introduces a confirmed-floor quantity
// (riskcalc.ConfirmedFloorQuantity) that caps what an *automated* path may sell
// while the account is in RECONCILE. The spec exempts the manual flatten saga:
// 수동 flatten은 자체 신선 조회를 수행하므로 이 규칙의 대상이 아니다(§0.3 —
// P1 동작 보존).
//
// The exemption is not a preference. The floor is zero whenever a snapshot is
// absent or stale, and "we cannot read the account right now" is exactly the
// situation in which a human hits flatten. Applying it here would mean the
// safety mechanism, not the market, decided the account could not be exited.
//
// Two things are asserted, because either alone is weak:
//
//  1. the rule cannot reach this package at all — no production file here
//     imports internal/riskcalc, so no call site can appear by accident;
//  2. the saga still sizes each order from its own fresh per-symbol read, with
//     no snapshot of any kind supplied. TestLiquidationSizesFromTheSellableQuantity
//     next door is the fuller characterisation; this one states the property in
//     the terms of the rule it is exempt from.

const flattenModulePath = "github.com/JungHoonGhae/tossinvest-cli/"

// TestFlattenCannotReachTheConfirmedFloorRule is the structural half. A
// behavioural test can only show that today's code path does not call it; this
// shows the call is unspellable without adding an import somebody has to
// justify in review.
//
// It checks *direct* imports, not the transitive graph: internal/flatten
// legitimately reaches riskcalc through internal/journal (the reservation
// ledger's arithmetic), and forbidding that would forbid the journal itself.
// What must not exist is a call from this package.
func TestFlattenCannotReachTheConfirmedFloorRule(t *testing.T) {
	t.Parallel()

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("reading the package directory: %v", err)
	}
	fset := token.NewFileSet()
	checked := 0
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, name, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parsing %s: %v", name, err)
		}
		checked++
		for _, imp := range f.Imports {
			path := strings.Trim(imp.Path.Value, `"`)
			if path == flattenModulePath+"internal/riskcalc" {
				t.Errorf("%s imports internal/riskcalc: the confirmed-floor rule is for automated "+
					"exits, and gating a manual flatten on a snapshot that has just been declared "+
					"unreadable traps the account in the position it is trying to leave (§0.3)", name)
			}
		}
	}
	// A positive control: an empty walk would make the assertion above pass for
	// the wrong reason.
	if checked == 0 {
		t.Fatal("no production files were inspected; the walk is broken")
	}
	if _, err := os.Stat(filepath.Join(".", "liquidate.go")); err != nil {
		t.Fatalf("expected liquidate.go in this package: %v", err)
	}
}

// TestFlattenSizesFromItsOwnFreshReadNotAConfirmedFloor is the behavioural
// half, stated in the rule's own terms: the saga is handed no snapshot and no
// staleness bound, and it still produces an order for the full sellable
// quantity. Under the confirmed-floor rule the same situation — no snapshot —
// would authorise nothing.
func TestFlattenSizesFromItsOwnFreshReadNotAConfirmedFloor(t *testing.T) {
	account := &accountFake{
		positions: []domain.Position{position("005930", "kr", 10)},
		sellable:  map[string]float64{"005930": 7},
		lower:     map[string]float64{"005930": 63000},
	}
	h := newLiqHarness(t, [][]json.RawMessage{{}}, account)
	ctx := context.Background()
	saga := h.saga(false)
	h.sell.after = account.setFlat

	if _, err := saga.CancelAll(ctx); err != nil {
		t.Fatalf("CancelAll: %v", err)
	}
	if _, err := saga.Liquidate(ctx); err != nil {
		t.Fatalf("Liquidate: %v", err)
	}

	orders := h.sell.orders()
	if len(orders) != 1 {
		t.Fatalf("orders = %+v, want one — a confirmed floor of 0 would have produced none", orders)
	}
	if orders[0].Quantity != 7 {
		t.Errorf("quantity = %v, want the 7 the saga read for itself", orders[0].Quantity)
	}
}
