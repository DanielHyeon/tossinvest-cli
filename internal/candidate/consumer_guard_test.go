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
	"Chase": true, "Verdict": true, "ApprovedCandidate": true, "VetoTally": true,
	"AssessChase": true, "AssessApprovedCandidate": true,
	"TallyVetoes": true, "TallyVerdicts": true,
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
	"cmd/tossctl/httpapi_reader.go":                "projects the same read-only assessed verdict for the private operator API",
	"internal/console/signals.go":                  "renders the /signals screen",
	"cmd/tossctl/candidate_test.go":                "asserts the scan report",
	"cmd/tossctl/candidate_review_test.go":         "asserts the report's notes and reducer use",
	"cmd/tossctl/consolesignals_test.go":           "asserts the /signals seam's own tallies",
	"cmd/tossctl/tally_alarm_surface_test.go":      "asserts the tally alarm at the scan output",
	"internal/console/signals_test.go":             "asserts the /signals screen",
	"internal/console/signals_newlylisted_test.go": "asserts the new-entrant marker",
	"internal/console/tally_alarm_test.go":         "asserts the tally alarm on the page",
	"internal/strategy/approved.go":                "a047 value-only handoff into a pure lane; it has no order authority",
	"internal/strategyengine/lane_test.go":         "asserts a synthetic approved-candidate lane contract without production authority",
}

// orderVerbs are the package selectors that can submit, amend, cancel or liquidate,
// keyed by the package's own name and mapped to the selector prefixes that matter.
// A nil list means every exported call on that package.
//
// They are the roots isolation_test.go's forbidden table names, expressed as things
// a file says rather than as things a package imports — which is the only form that
// still discriminates inside cmd/tossctl.
//
// # official.New, and what a selector scan cannot see
//
// The first four entries name packages whose whole purpose is the order path, so any
// call into them is the thing being forbidden. internal/official is not like that:
// it is the read client as well, and this repository's discovery adapters are built
// on its Rankings method. So only some of its surface is an order verb.
//
// The three write prefixes were the original list and they match nothing in this
// tree, which is worth stating rather than leaving as a comfortable silence: there
// is no package-level official.Place. The writes are methods — Client.PlaceOrder,
// Client.CreateConditionalOrder, Client.CancelConditionalOrder
// (internal/official/conditional_writes.go) — and by the time a file has a *Client
// in hand the package name is gone from the call site. `client.CreateConditionalOrder(…)`
// is invisible to any scan of this shape, and cmd/tossctl/candidate.go used to
// construct exactly that client.
//
// What is visible is the construction. `official.New` is the one line that turns
// credentials into a value carrying every write the backend has, so that is what a
// verdict-reading file may not say. The write prefixes stay beside it: they cost
// nothing and they would catch a package-level helper the day somebody adds one.
var orderVerbs = map[string][]string{
	"execgw":      nil, // every exported call on the execution gateway
	"orderintent": nil,
	"trading":     nil,
	"flatten":     nil,
	"official":    {"Place", "Cancel", "Modify", "New"},
}

// orderPackagePaths maps each key above onto the import path it names, so that an
// aliased import is resolved rather than matched by spelling.
//
// The candidate side has done this since it was written (candidateImportName) and
// this side did not, which left `import eg ".../internal/execgw"` followed by
// `eg.Submit(…)` passing — one line, no import-graph change, in a package whose
// import graph already contains the execution gateway.
var orderPackagePaths = map[string]string{
	"execgw":      "/internal/execgw",
	"orderintent": "/internal/orderintent",
	"trading":     "/internal/trading",
	"flatten":     "/internal/flatten",
	"official":    "/internal/official",
}

// orderVerbNames is what this file calls each order package by, and the prefixes
// that matter for it.
//
// It starts from the canonical names — a file naming `trading.Something` without
// importing it would not compile, but a guard that only trusts the import list is a
// guard that stops working the day the resolution has a hole in it — and then adds
// whatever local name each import actually binds.
func orderVerbNames(file *ast.File, module string) map[string][]string {
	out := make(map[string][]string, len(orderVerbs))
	for name, prefixes := range orderVerbs {
		out[name] = prefixes
	}
	for _, spec := range file.Imports {
		path := strings.Trim(spec.Path.Value, `"`)
		for key, suffix := range orderPackagePaths {
			if path != module+suffix {
				continue
			}
			local := key
			if spec.Name != nil {
				local = spec.Name.Name
			}
			if local == "_" || local == "." {
				// A blank import binds no identifier and a dot import binds every
				// exported name with no qualifier at all — neither is reachable by a
				// selector scan. namesAnOrderVerb refuses both outright rather than
				// mapping them onto a name that would then match nothing.
				continue
			}
			out[local] = orderVerbs[key]
		}
	}
	return out
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

// unqualifiedVerdictSelectors are the ones a file can say without naming this
// package at all.
//
// `Passed` is the predicate and `Chase` is the field on Verdict — `v.Chase.Passed()`
// and `v.Chase.SeenLate` are how both screens reach the judgement, and neither of
// them requires an import: a helper in the same package that returns a
// candidate.Verdict makes every type name vanish from the call site, and design D6
// case 2 is exactly that shape inside cmd/tossctl.
//
// # Why it is two names and not all of verdictSymbols
//
// `Verdict` and `VetoTally` stay in the qualified list only, and that is a measured
// decision rather than a shortcut. Scanning `Verdict` unqualified matches
// cmd/tossctl/verify.go:363 (`o.Verdict`, a live-verification step's outcome),
// internal/console/data.go:318 (`s.Verdict.Terminal()`, an order state) and three
// console tests — none of which has ever seen a chase judgement. Adding five files
// to the allowlist to catch a spelling collision would make the list longer than the
// set of files anybody has thought about, and an allowlist nobody can read is one
// nobody re-examines.
//
// Nothing escapes through the gap. To hold a chase judgement without importing this
// package a file has to get one from a helper, and to use it, it has to select
// `.Chase` or `.Passed` — `.SeenLate`, `.Extended` and `.NearHigh` are reachable only
// through `.Chase`, and VetoTally's own pass count is spelled `.Passed`. Naming
// either type outright still requires the import, which the qualified scan sees.
//
// `Sighting` and `SeenLate` were in this set for one draft and are not: they are the
// measurements a judgement is made *from*, so a fixture that builds one is not a
// consumer of the verdict.
var unqualifiedVerdictSelectors = map[string]bool{
	"Passed": true, "Chase": true,
}

// namesVerdict reports whether a parsed file reads a chase judgement.
//
// `pkg` is what this package is imported as, or "" when it is not imported at all —
// which does not end the question. The qualified form (`candidate.Chase`) needs the
// import; the unqualified one does not, and the unqualified one is the case the
// guard was written for.
func namesVerdict(file *ast.File, pkg string) bool {
	found := false
	ast.Inspect(file, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if id, ok := sel.X.(*ast.Ident); ok && pkg != "" && id.Name == pkg &&
			verdictSymbols[sel.Sel.Name] {
			found = true
		}
		// And the verdict's own predicate and fields, wherever the value came from.
		if unqualifiedVerdictSelectors[sel.Sel.Name] {
			found = true
		}
		return true
	})
	return found
}

// namesAnOrderVerb reports whether a parsed file can say place, amend, cancel or
// liquidate, and which selector said it.
//
// The package names are resolved from the file's own imports rather than matched by
// spelling, for candidateImportName's reason one direction over: an alias is the
// cheapest way past a guard that compares literal words, and this side did not
// resolve them.
func namesAnOrderVerb(file *ast.File, module string) (string, bool) {
	// A dot import binds the package's exported names with no qualifier, so nothing
	// below could ever see them. It is refused as the package having been named.
	for _, spec := range file.Imports {
		if spec.Name == nil || spec.Name.Name != "." {
			continue
		}
		path := strings.Trim(spec.Path.Value, `"`)
		for key, suffix := range orderPackagePaths {
			if path == module+suffix {
				return key + " (dot-imported, so its calls carry no qualifier at all)", true
			}
		}
	}

	names := orderVerbNames(file, module)
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
		prefixes, listed := names[id.Name]
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
		// The import is looked up and the scan runs either way.
		//
		// It used to gate the whole scan: a file that did not import this package was
		// skipped outright. Go imports are per file and cmd/tossctl is one package, so
		// a new file there that takes a candidate.Chase from a helper beside it and
		// calls .Passed() needs no import of its own — design D6 case 2, which this
		// guard was written for, walked straight through the gate meant to catch it.
		// So the import now decides only whether the qualified form can be resolved.
		pkg, _ := candidateImportName(parsed, module)
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
	module := modulePath(t, root)
	fset := token.NewFileSet()
	for rel, why := range verdictReaders {
		parsed, err := parser.ParseFile(fset, filepath.Join(root, rel), nil,
			parser.SkipObjectResolution)
		if err != nil {
			t.Fatalf("parsing %s: %v", rel, err)
		}
		if verb, bad := namesAnOrderVerb(parsed, module); bad {
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
	module := modulePath(t, root)
	const fixture = "cmd/tossctl/order.go"
	if _, listed := verdictReaders[fixture]; listed {
		t.Fatalf("%s is now a verdict reader, so it cannot serve as the positive control", fixture)
	}
	parsed, err := parser.ParseFile(token.NewFileSet(), filepath.Join(root, fixture), nil,
		parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parsing %s: %v", fixture, err)
	}
	verb, bad := namesAnOrderVerb(parsed, module)
	if !bad {
		t.Fatalf("the order-verb detector found nothing in %s, which is the order command. "+
			"TestNoFileThatReadsTheVerdictCanAlsoPlaceAnOrder is passing because it is "+
			"matching nothing", fixture)
	}
	t.Logf("positive control: %s names %s", fixture, verb)
}

// TestTheOrderVerbDetectorSeesTheFormsThatUsedToWalkPast is the positive control for
// the three holes F5 measured, each one written out as the source it would be.
//
// A guard's claim is a set of things it did not find, so the only evidence that it
// finds anything is a fixture it must reject. These three were verified green — that
// is, forbidden and undetected — before this test existed.
func TestTheOrderVerbDetectorSeesTheFormsThatUsedToWalkPast(t *testing.T) {
	module := modulePath(t, moduleRoot(t))
	for _, tc := range []struct {
		name string
		src  string
	}{
		{
			// (a) The client, not the package. Every write on the official backend is
			// a method on *official.Client, so `client.CreateConditionalOrder(…)`
			// carries no package name at all. What is visible is the constructor.
			name: "constructing the official client",
			src: `package main
import "` + module + `/internal/official"
func f(creds official.Credentials, token string) {
	client := official.New(creds, token)
	_ = client
}`,
		},
		{
			// (c) An alias. The candidate side has resolved these since it was
			// written; this side compared spellings.
			name: "an aliased execution gateway",
			src: `package main
import eg "` + module + `/internal/execgw"
func f(g *eg.Gateway) { _ = g }`,
		},
		{
			name: "a dot-imported execution gateway",
			src: `package main
import . "` + module + `/internal/execgw"
func f() {}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			parsed, err := parser.ParseFile(token.NewFileSet(), "fixture.go", tc.src,
				parser.SkipObjectResolution)
			if err != nil {
				t.Fatalf("parsing the fixture: %v", err)
			}
			verb, bad := namesAnOrderVerb(parsed, module)
			if !bad {
				t.Errorf("the detector found nothing in:\n%s\n\nEach of these was verified to "+
					"pass the guard while doing the thing the guard forbids", tc.src)
			}
			t.Logf("detected: %s", verb)
		})
	}
}

// TestTheVerdictDetectorSeesAReaderThatNeverImportsThisPackage is D6 case 2 as a
// fixture: cmd/tossctl is one Go package, so a new file in it can take a
// candidate.Chase from a helper beside it and never write the import.
func TestTheVerdictDetectorSeesAReaderThatNeverImportsThisPackage(t *testing.T) {
	const src = `package main
func f() {
	v := latestVerdict()
	if v.Chase.Passed() {
		submit(v)
	}
}`
	parsed, err := parser.ParseFile(token.NewFileSet(), "fixture.go", src,
		parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parsing the fixture: %v", err)
	}
	// "" is what candidateImportName reports for a file that does not import this
	// package, and the scan has to run anyway — that is the whole finding.
	if !namesVerdict(parsed, "") {
		t.Error("the verdict detector found nothing in a file that reads Chase.Passed() " +
			"without importing this package. That file needs zero import edges, which is " +
			"why the unit of this guard is the file's text rather than its import list")
	}
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
