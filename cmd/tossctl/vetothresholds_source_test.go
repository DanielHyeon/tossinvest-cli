package main

// vetothresholds_source_test.go is section 6: one source for the veto thresholds,
// pinned structurally.
//
// A value test would not do it. Two literals agree until somebody edits one, and the
// day they stop agreeing is the day a value is introduced — which is exactly when
// nobody is looking at the other one. So what is checked is that there is only one
// place to edit.
//
// The precedent is candidate_review_test.go's guard on the tallies, and the reason
// is the one written into internal/candidate/watch.go: a fourth shadow band appeared
// on the scan output and not on /signals, because two surfaces assembled the same
// thing separately.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/candidate"
)

// thresholdSource is the one file allowed to spell the composite literal.
const thresholdSource = "vetothresholds.go"

// TestOnlyOneFileInThisCommandBuildsTheVetoThresholds is task 6.2.
func TestOnlyOneFileInThisCommandBuildsTheVetoThresholds(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("reading the command directory: %v", err)
	}
	fset := token.NewFileSet()
	scanned, found := 0, map[string]int{}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		scanned++
		file, parseErr := parser.ParseFile(fset, name, nil, parser.SkipObjectResolution)
		if parseErr != nil {
			t.Fatalf("parsing %s: %v", name, parseErr)
		}
		ast.Inspect(file, func(n ast.Node) bool {
			lit, ok := n.(*ast.CompositeLit)
			if !ok {
				return true
			}
			sel, ok := lit.Type.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "VetoThresholds" {
				return true
			}
			found[name]++
			return true
		})
	}
	if scanned == 0 {
		t.Fatal("no production sources were scanned; this guard is reading nothing")
	}
	if found[thresholdSource] == 0 {
		t.Errorf("%s builds no VetoThresholds; either the constructor moved or this guard "+
			"is matching nothing, and in both cases the check below proves nothing",
			thresholdSource)
	}
	for name, n := range found {
		if name == thresholdSource {
			continue
		}
		t.Errorf("%s builds a candidate.VetoThresholds literal (%d of them). Two literals "+
			"agree until one is edited, and then `tossctl candidate scan` and /signals "+
			"report different verdicts about the same store at the same instant — with "+
			"nothing failing, because each screen renders its own numbers correctly. "+
			"Call %s's constructor instead.", name, n, thresholdSource)
	}
}

// TestTheTwoSurfacesApplyTheSameThresholds is the value half.
//
// Structural guards go stale in one direction — they can stop matching. This asks
// the question directly, and it is cheap because the constructor is a function both
// call sites use.
func TestTheTwoSurfacesApplyTheSameThresholds(t *testing.T) {
	scan := candidateCycleOptions(&rootOptions{}, candidate.MarketKR, nil, nil, nil).Thresholds
	one := candidateVetoThresholds()
	if scan != one {
		t.Errorf("the scan cycle applies %+v and the constructor returns %+v", scan, one)
	}
	// And the values are still the ones D18 approved and did not approve: one
	// threshold, and two deliberately absent.
	if one.NearHighDistancePct != candidate.DefaultNearHighThresholdPct {
		t.Errorf("near_high = %q, want the contract's %q",
			one.NearHighDistancePct, candidate.DefaultNearHighThresholdPct)
	}
	if one.SeenLatePercentilePct != "" || one.ExtendedGainPct != "" {
		t.Errorf("seen_late = %q and extended = %q; this change introduces no threshold, and "+
			"a value here would be a policy number with no source, no market and no "+
			"verification state — the thing design D6 forbids",
			one.SeenLatePercentilePct, one.ExtendedGainPct)
	}
}

// TestTheSignalsSeamReadsTheSameConstructor closes the path the structural test
// above cannot see: console.go could stop building a literal by calling something
// else entirely.
func TestTheSignalsSeamReadsTheSameConstructor(t *testing.T) {
	const path = "console.go"
	src, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	file, err := parser.ParseFile(token.NewFileSet(), path, src, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parsing %s: %v", path, err)
	}
	calls := 0
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if id, ok := call.Fun.(*ast.Ident); ok && id.Name == "candidateVetoThresholds" {
			calls++
		}
		return true
	})
	if calls == 0 {
		t.Errorf("%s never calls candidateVetoThresholds; the /signals seam is getting its "+
			"thresholds from somewhere else, and the two surfaces can then diverge without "+
			"a literal to find", path)
	}
}
