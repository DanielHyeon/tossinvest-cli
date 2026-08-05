package reconcile_test

// a085: the broker's display name rides along, and the comparison never sees it.
//
// Holding is the type a quantity mismatch is judged against. Adding a display
// field to it is only safe while nothing in the judgement reads it, so that is
// what this file pins.

import (
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

func TestNamesDoNotAffectTheComparison(t *testing.T) {
	local := reconcile.LocalState{
		AccountRef: "acct-7",
		Positions:  map[string]string{"042660": "2"},
	}
	named := reconcile.Snapshot{
		AccountRef: "acct-7",
		Holdings: []reconcile.Holding{
			{Symbol: "042660", Quantity: "2", Market: "kr", Name: "한화오션"},
		},
	}
	unnamed := named
	unnamed.Holdings = []reconcile.Holding{
		{Symbol: "042660", Quantity: "2", Market: "kr"},
	}

	withName := reconcile.Comparer{}.Compare(named, local)
	withoutName := reconcile.Comparer{}.Compare(unnamed, local)

	if withName.BlocksEntry() || withoutName.BlocksEntry() {
		t.Fatalf("agreeing quantities must not block: %+v / %+v", withName, withoutName)
	}
	if withName.Matched != withoutName.Matched {
		t.Errorf("matched = %d with a name, %d without; the name is display evidence and the "+
			"comparison must not be able to see it", withName.Matched, withoutName.Matched)
	}
}

func TestADisagreementIsUnchangedByTheName(t *testing.T) {
	local := reconcile.LocalState{
		AccountRef: "acct-7",
		Positions:  map[string]string{"042660": "2"},
	}
	snap := reconcile.Snapshot{
		AccountRef: "acct-7",
		Holdings: []reconcile.Holding{
			{Symbol: "042660", Quantity: "0", Market: "kr", Name: "한화오션"},
		},
	}
	diff := reconcile.Comparer{}.Compare(snap, local)
	if len(diff.Quantities) != 1 {
		t.Fatalf("quantities = %+v, want the one mismatch", diff.Quantities)
	}
	if diff.Quantities[0].Symbol != "042660" || diff.Quantities[0].Broker != "0" {
		t.Errorf("mismatch = %+v, want the account's number unchanged by the name", diff.Quantities[0])
	}
}
