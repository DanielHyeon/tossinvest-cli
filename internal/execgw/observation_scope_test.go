package execgw_test

import (
	"context"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// observation_scope_test.go is change add-net-rr-measurement task 4.3: the
// recording is wired to entries and to nothing else.
//
// Two boundaries, each with a different failure if crossed.
//
//	IssueReduction   shares the evaluateChain seam. Wiring it would produce a row
//	                 per exit whose every measured column is empty — an exit carries
//	                 no stop and no target — and those rows would sit in the same
//	                 table a threshold study reads, diluting it with entries that
//	                 were never entries.
//	risk.Evaluate    is a pure function over values. A journal dependency inside it
//	                 would mean the counterfactual harness could not run the chain
//	                 without a database, which is the isolation trade-analytics
//	                 requires of the analysis path.

// TestAReductionIsNotObserved is the behavioural half.
func TestAReductionIsNotObserved(t *testing.T) {
	observer := &recordingObserver{}
	rig := newGuardian(t, func(o *execgw.RiskGuardianOptions) { o.Observer = observer })
	ctx := context.Background()

	account := guardianAccount()
	account.HeldQuantity = "10"
	if _, err := rig.guardian.IssueReduction(ctx, execgw.ReductionIssuance{
		Intent: risk.Intent{
			Market: guardianIntent().Market, Symbol: "005930",
			Side: risk.SideSell, Quantity: "10",
		},
		Account: account,
		Reason:  "stop breach",
	}); err != nil {
		t.Fatalf("the reduction must be issued: %v", err)
	}
	if rows := observer.all(); len(rows) != 0 {
		t.Fatalf("a reduction produced %d observations: %+v. An exit has no stop and no "+
			"target, so the row would be empty where the measurement is", len(rows), rows)
	}

	// A *refused* reduction is equally unobserved: the seam is shared, so the
	// refusal path is where an accidental wiring would show up first.
	oversold := guardianAccount()
	oversold.HeldQuantity = "1"
	if _, err := rig.guardian.IssueReduction(ctx, execgw.ReductionIssuance{
		Intent: risk.Intent{
			Market: guardianIntent().Market, Symbol: "005930",
			Side: risk.SideSell, Quantity: "10",
		},
		Account: oversold,
		Reason:  "stop breach",
	}); err == nil {
		t.Fatal("selling more than the holding must be refused")
	}
	if rows := observer.all(); len(rows) != 0 {
		t.Fatalf("a refused reduction produced %d observations: %+v", len(rows), rows)
	}
}

// TestOnlyEntriesReachTheObservationTable is the same claim through the real sink,
// so it holds for what is on disk rather than for what an in-memory double saw.
func TestOnlyEntriesReachTheObservationTable(t *testing.T) {
	rig := newGuardian(t, nil)
	ctx := context.Background()
	observer := execgw.NewAsyncObserver(execgw.AsyncObserverOptions{Sink: rig.journal})
	guardian := newGuardianOn(t, rig, observer)

	account := guardianAccount()
	account.HeldQuantity = "10"
	if _, err := guardian.IssueReduction(ctx, execgw.ReductionIssuance{
		Intent: risk.Intent{
			Market: guardianIntent().Market, Symbol: "005930",
			Side: risk.SideSell, Quantity: "10",
		},
		Account: account,
		Reason:  "take profit",
	}); err != nil {
		t.Fatalf("the reduction: %v", err)
	}
	observer.Close()

	rows, err := rig.journal.EntryObservations(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Fatalf("the observation table holds %d rows after only a reduction: %+v", len(rows), rows)
	}
}

// TestTheRiskPackageStaysFreeOfTheJournal is the structural half, and it is the
// one a behavioural test cannot give. internal/risk being a pure function over
// values is what lets the counterfactual harness push thousands of grid points
// through the real chain with no database at all; the day it imports the journal,
// that harness needs one, and "the analysis path does not touch the ledger" stops
// being true by construction.
func TestTheRiskPackageStaysFreeOfTheJournal(t *testing.T) {
	const riskDir = "../risk"
	// internal/costs is deliberately absent from this list: the chain genuinely
	// needs a cost model to find break-even, and that model is a value the caller
	// passes in rather than a store the chain reaches into.
	forbidden := []string{
		"github.com/JungHoonGhae/tossinvest-cli/internal/journal",
		"github.com/JungHoonGhae/tossinvest-cli/internal/execgw",
	}

	entries, err := os.ReadDir(riskDir)
	if err != nil {
		t.Fatalf("reading %s: %v", riskDir, err)
	}
	fset := token.NewFileSet()
	scanned := 0
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		scanned++
		file, err := parser.ParseFile(fset, riskDir+"/"+name, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parsing %s: %v", name, err)
		}
		for _, spec := range file.Imports {
			path, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				t.Fatal(err)
			}
			for _, bad := range forbidden {
				if path == bad || strings.HasPrefix(path, bad+"/") {
					t.Errorf("internal/risk/%s imports %s. The chain is a pure function over "+
						"values; a storage dependency in it would put a database behind every "+
						"counterfactual grid point and behind every rung's unit test", name, path)
				}
			}
		}
	}
	if scanned == 0 {
		t.Fatal("scanned no risk sources; the path is wrong and this test proves nothing")
	}
}

// TestTheObservationIsBuiltOutsideTheChain: the recording is assembled from the
// chain's *output*, not by the chain. Evaluate returns the same verdict whether or
// not anybody is observing, which is why the harness can call it directly.
func TestTheObservationIsBuiltOutsideTheChain(t *testing.T) {
	in := risk.Input{
		Now:    fixedNow,
		Intent: guardianIntent(),
		Account: risk.AccountState{
			Mode:              risk.ModeNormal,
			AllowedSymbols:    []string{"005930"},
			CashAvailable:     riskcalc.Money{Amount: "5000000", Currency: "KRW"},
			OpenExposure:      riskcalc.Money{Amount: "0", Currency: "KRW"},
			DailyRealizedLoss: riskcalc.Money{Amount: "0", Currency: "KRW"},
			AccountEquity:     riskcalc.Money{Amount: "10000000", Currency: "KRW"},
		},
		Policy: guardianPolicy(),
		Costs:  costs.DefaultModel(),
	}
	in.Intent.AccountRef = "acct-7"

	// No journal, no Guardian, no observer: just the chain and the values.
	first := risk.Evaluate(in)
	second := risk.Evaluate(in)
	if first != second {
		t.Fatalf("the chain is not a pure function of its input: %+v then %+v", first, second)
	}
	if !first.Allowed {
		t.Fatalf("the fixture must be allowed: %s %s", first.Reason, first.Detail)
	}

	// And the measurement is likewise computable from values alone, which is what
	// the harness will do with it.
	ratios := risk.MeasureEntry(in.Costs, in.Intent.Market,
		in.Intent.LimitPrice, in.Intent.StopPrice, in.Intent.TargetPrice)
	if ratios.GrossRewardRisk == "" || ratios.NetRewardRisk == "" {
		t.Errorf("the ratios must be computable without any storage: %+v", ratios)
	}
	_ = journal.CostScopeFeeTaxOnly // the scope is a constant, not a lookup
}
