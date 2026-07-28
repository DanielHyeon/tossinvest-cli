package candidate

// consumer_guard_test.go is section 5: the reverse direction of isolation_test.go.
//
// # Why the existing four are not this
//
// isolation_test.go's four tests all run forwards — they ask what internal/candidate
// imports, and refuse a path from here to an order. Nothing asks the other question:
// who reads the verdict, and can that reader place an order. This file is that
// question.
//
// # Why the unit is the file and the symbol, not the package
//
// A package allowlist fails open in two places, and both of them are in this
// repository today.
//
//	internal/candidatesrc imports internal/candidate and internal/official, and
//	internal/official carries PlaceOrder. Any allowlist that lets a wiring package
//	read the verdict has to contain the one package that can call both.
//
//	cmd/tossctl is a single Go package that already imports internal/execgw,
//	internal/orderintent, internal/trading and internal/app/engine. A new file in
//	it that reads Chase.Passed() and submits an order adds zero import edges, so a
//	package-level check is green while the exact thing D7 forbids is happening.
//
// So the allowlist is a list of files, and the second half is an intersection: a
// file that may name the verdict may not also name an order verb. That check keeps
// working inside a package whose import graph already contains everything.
//
// # Why now, with no thresholds anywhere
//
// Chase.Passed() is unreachable today — two of the three vetoes have no approved
// threshold, so no candidate can have all three measured and clear — and this change
// does not give it one. That is exactly why the guard goes in now. Written after a
// threshold is approved, its allowlist would be seeded with whatever consumers
// existed at that moment, and every one of them would be approved by the act of
// writing the list down rather than by anybody deciding. The list below is short
// because the answer today is "three files render it and nothing else looks at it",
// and that is a fact worth freezing while it is still true.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

// verdictSymbols are the names that carry a chase judgement out of this package.
//
// AssessChase and the two tally constructors are here as well as the types, because
// a consumer that produced its own verdict rather than reading one would be doing
// the same thing by a different route.
var verdictSymbols = map[string]bool{
	"Chase": true, "Verdict": true, "VetoTally": true,
	"AssessChase": true, "TallyVetoes": true, "TallyVerdicts": true,
}

// verdictReaders is every file outside this package that may name one of them.
//
// It was produced by searching, not by remembering: three production files render
// the verdict and their tests read it back. internal/candidatesrc is deliberately
// NOT here — it imports this package for Row and Source and has never named a
// verdict, and the day it does, this test fails and somebody has to say why the
// package that can also reach PlaceOrder needs to see a judgement.
var verdictReaders = map[string]string{
	"cmd/tossctl/candidate.go":                     "renders the scan report's veto block",
	"cmd/tossctl/console.go":                       "wires the /signals seam to candidate.Assess",
	"internal/console/signals.go":                  "renders the /signals screen",
	"cmd/tossctl/candidate_test.go":                "asserts the scan report",
	"cmd/tossctl/candidate_review_test.go":         "asserts the report's notes and reducer use",
	"cmd/tossctl/consolesignals_test.go":           "asserts the /signals seam's own tallies",
	"cmd/tossctl/tally_alarm_surface_test.go":      "asserts the tally alarm at the scan output",
	"internal/console/signals_test.go":             "asserts the /signals screen",
	"internal/console/signals_newlylisted_test.go": "asserts the new-entrant marker",
	"internal/console/tally_alarm_test.go":         "asserts the tally alarm on the page",
}

// orderVerbs are the package selectors that can submit, amend, cancel or liquidate.
//
// They are the roots isolation_test.go's forbidden table names, expressed as things
// a file says rather than as things a package imports — which is the only form that
// still discriminates inside cmd/tossctl.
var orderVerbs = map[string][]string{
	"execgw":      nil, // every exported call on the execution gateway
	"orderintent": nil,
	"trading":     nil,
	"flatten":     nil,
	"official":    {"Place", "Cancel", "Modify"},
}

// goFilesUnder walks the repository for .go files, skipping this package (which
// declares the symbols) and anything vendored.
func goFilesUnder(t *testing.T, root string) []string {
	t.Helper()
	var out []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			switch {
			case rel == ".": // keep walking
			case strings.HasPrefix(filepath.Base(rel), "."),
				rel == "vendor", rel == "internal/candidate":
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(rel, ".go") {
			out = append(out, rel)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking %s: %v", root, err)
	}
	return out
}

// candidateImportName is the identifier a file refers to this package by, and
// whether it imports it at all. Aliased imports are handled because an alias is the
// cheapest way past a guard that matches the literal word "candidate".
func candidateImportName(file *ast.File, module string) (string, bool) {
	for _, spec := range file.Imports {
		path := strings.Trim(spec.Path.Value, `"`)
		if path != module+"/internal/candidate" {
			continue
		}
		if spec.Name != nil {
			return spec.Name.Name, spec.Name.Name != "_"
		}
		return "candidate", true
	}
	return "", false
}

// namesVerdict reports whether a parsed file reads a chase judgement.
func namesVerdict(file *ast.File, pkg string) bool {
	found := false
	ast.Inspect(file, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if id, ok := sel.X.(*ast.Ident); ok && id.Name == pkg && verdictSymbols[sel.Sel.Name] {
			found = true
		}
		// And the verdict's own predicate, wherever the value came from. A helper
		// that returns a Chase makes the type name disappear from the call site.
		if sel.Sel.Name == "Passed" {
			found = true
		}
		return true
	})
	return found
}

// namesAnOrderVerb reports whether a parsed file can say place, amend, cancel or
// liquidate, and which selector said it.
func namesAnOrderVerb(file *ast.File) (string, bool) {
	var found string
	ast.Inspect(file, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		id, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}
		prefixes, listed := orderVerbs[id.Name]
		if !listed {
			return true
		}
		if prefixes == nil {
			found = id.Name + "." + sel.Sel.Name
			return false
		}
		for _, p := range prefixes {
			if strings.HasPrefix(sel.Sel.Name, p) {
				found = id.Name + "." + sel.Sel.Name
				return false
			}
		}
		return true
	})
	return found, found != ""
}

// TestOnlyTheListedFilesCanNameTheChaseVerdict is task 5.1.
func TestOnlyTheListedFilesCanNameTheChaseVerdict(t *testing.T) {
	root := moduleRoot(t)
	module := modulePath(t, root)
	files := goFilesUnder(t, root)
	if len(files) < 50 {
		t.Fatalf("the walk found %d Go files; it is reading the wrong tree and this guard "+
			"proves nothing", len(files))
	}

	fset := token.NewFileSet()
	seen := map[string]bool{}
	for _, rel := range files {
		parsed, err := parser.ParseFile(fset, filepath.Join(root, rel), nil,
			parser.SkipObjectResolution)
		if err != nil {
			t.Fatalf("parsing %s: %v", rel, err)
		}
		pkg, imports := candidateImportName(parsed, module)
		if !imports {
			continue
		}
		if !namesVerdict(parsed, pkg) {
			continue
		}
		seen[rel] = true
		if _, allowed := verdictReaders[rel]; !allowed {
			t.Errorf("%s names a chase verdict and is not in this test's list.\n\n"+
				"Chase.Passed() is the answer to \"may this candidate be bought\", and the "+
				"list exists so that gaining a reader is a decision somebody made rather "+
				"than a diff nobody noticed. Add the file with the reason it needs the "+
				"verdict — and if that file can also reach an order path, the next test "+
				"will say so.", rel)
		}
	}
	for rel := range verdictReaders {
		if !seen[rel] {
			t.Errorf("the list names %s and it no longer reads a verdict; a permission "+
				"nobody uses is one nobody re-examines", rel)
		}
	}
}

// TestNoFileThatReadsTheVerdictCanAlsoPlaceAnOrder is task 5.2.
//
// This is the half a package allowlist cannot do. cmd/tossctl is one package and
// already imports the execution gateway, so the import graph says nothing about
// whether the file rendering the veto block could submit an order — only the file's
// own text does.
func TestNoFileThatReadsTheVerdictCanAlsoPlaceAnOrder(t *testing.T) {
	root := moduleRoot(t)
	fset := token.NewFileSet()
	for rel, why := range verdictReaders {
		parsed, err := parser.ParseFile(fset, filepath.Join(root, rel), nil,
			parser.SkipObjectResolution)
		if err != nil {
			t.Fatalf("parsing %s: %v", rel, err)
		}
		if verb, bad := namesAnOrderVerb(parsed); bad {
			t.Errorf("%s reads the chase verdict (%s) and also names %s.\n\n"+
				"spec candidate-discovery: 후보에서 주문으로 가는 코드가 컴파일되지 않아야 한다. "+
				"One file that can do both is one edit away from doing both in sequence, and "+
				"the import graph will not change when it does.", rel, why, verb)
		}
	}
}

// TestTheConsumerGuardFiresOnAFileThatNamesAnOrderVerb is the positive control.
//
// Every static check in this repository can fail in one direction only. A detector
// that had quietly stopped matching would forbid nothing while reporting success,
// and this file's whole claim is a set of things it did not find.
//
// cmd/tossctl/order.go is the fixture on purpose: it is the order command, it is not
// in the reader list, and if it ever stops naming an order verb the fixture has to
// be replaced rather than the guard trusted.
func TestTheConsumerGuardFiresOnAFileThatNamesAnOrderVerb(t *testing.T) {
	root := moduleRoot(t)
	const fixture = "cmd/tossctl/order.go"
	if _, listed := verdictReaders[fixture]; listed {
		t.Fatalf("%s is now a verdict reader, so it cannot serve as the positive control", fixture)
	}
	parsed, err := parser.ParseFile(token.NewFileSet(), filepath.Join(root, fixture), nil,
		parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parsing %s: %v", fixture, err)
	}
	verb, bad := namesAnOrderVerb(parsed)
	if !bad {
		t.Fatalf("the order-verb detector found nothing in %s, which is the order command. "+
			"TestNoFileThatReadsTheVerdictCanAlsoPlaceAnOrder is passing because it is "+
			"matching nothing", fixture)
	}
	t.Logf("positive control: %s names %s", fixture, verb)
}

// TestTheVerdictDetectorFiresOnAFileThatReadsOne is the other positive control: the
// list is only meaningful if the thing it lists is detectable.
func TestTheVerdictDetectorFiresOnAFileThatReadsOne(t *testing.T) {
	root := moduleRoot(t)
	module := modulePath(t, root)
	const fixture = "internal/console/signals.go"
	parsed, err := parser.ParseFile(token.NewFileSet(), filepath.Join(root, fixture), nil,
		parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parsing %s: %v", fixture, err)
	}
	pkg, imports := candidateImportName(parsed, module)
	if !imports {
		t.Fatalf("%s no longer imports this package; the detector's import step is matching "+
			"nothing", fixture)
	}
	if !namesVerdict(parsed, pkg) {
		t.Fatalf("the verdict detector found nothing in %s, which renders the verdict on "+
			"every row of the signals screen", fixture)
	}
}
