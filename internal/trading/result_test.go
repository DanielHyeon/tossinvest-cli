package trading

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
)

// MutationResult is the shape every order-mutation surface returns, so the
// engine packages that will consume it (journal, execution gateway) must be
// able to name it without importing the CLI's trading policy module. It now
// lives in internal/domain; internal/trading keeps a type alias so upstream
// callers (client, official, hybrid, ops, output, cmd) compile unchanged.
//
// This test is the contract: alias, not copy.
func TestMutationResultIsDomainAlias(t *testing.T) {
	t.Parallel()

	if got, want := reflect.TypeOf(MutationResult{}), reflect.TypeOf(domain.MutationResult{}); got != want {
		t.Fatalf("trading.MutationResult must be the same type as domain.MutationResult; got %v, want %v", got, want)
	}

	// A converted copy would still pass the reflect check if the fields matched,
	// so assign in both directions without conversion: only a true alias
	// compiles.
	var fromTrading domain.MutationResult = MutationResult{Kind: "place"}
	var fromDomain MutationResult = fromTrading
	if fromDomain.Kind != "place" {
		t.Fatalf("round-trip lost data: %+v", fromDomain)
	}
}

// TestMutationResultJSONContractUnchanged pins the wire field names and their
// omitempty behaviour. `tossctl order place --output json` and the MCP order
// tools serialize this struct, so a renamed or dropped tag is a breaking change
// for agents and scripts, not an internal detail.
func TestMutationResultJSONContractUnchanged(t *testing.T) {
	t.Parallel()

	full := MutationResult{
		Kind:                  "amend",
		Status:                "accepted",
		OrderID:               "o-new",
		OriginalOrderID:       "o-old",
		CurrentOrderID:        "o-new",
		Symbol:                "005930",
		Market:                "kr",
		Quantity:              3,
		FilledQuantity:        1,
		Price:                 70000,
		AverageExecutionPrice: 69950,
		OrderDate:             "2026-07-26",
		Warnings:              []string{"lineage cache not updated"},
	}
	const wantFull = `{"kind":"amend","status":"accepted","order_id":"o-new","original_order_id":"o-old","current_order_id":"o-new","symbol":"005930","market":"kr","quantity":3,"filled_quantity":1,"price":70000,"average_execution_price":69950,"order_date":"2026-07-26","warnings":["lineage cache not updated"]}`

	got, err := json.Marshal(full)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(got) != wantFull {
		t.Fatalf("json contract changed:\n got: %s\nwant: %s", got, wantFull)
	}

	// Zero value: only the two non-omitempty fields survive.
	got, err = json.Marshal(MutationResult{})
	if err != nil {
		t.Fatalf("marshal zero: %v", err)
	}
	if string(got) != `{"kind":"","status":""}` {
		t.Fatalf("zero-value json contract changed: %s", got)
	}
}
