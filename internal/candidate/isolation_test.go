package candidate

// isolation_test.go is task 6.1: the spec requirement "발굴은 주문 경로에 도달할 수
// 없어야 한다", stated as something that fails rather than as something written
// down.
//
// # Why a test and not a comment
//
// This package is where strategy lanes will attach, and the natural mistake at
// that moment is one line from a candidate to an order (D7). A comment saying not
// to does not fail; a review that catches it once does not catch it next quarter.
// The requirement's own words are that the code must not compile — so what is
// checked is the import graph, because an import is what makes a type nameable
// and a type is what makes the line compile.
//
// # Why the graph is walked transitively
//
// A check against this package's *direct* imports passes while
// `candidate → obs → execgw` exists, and that is the shape a real regression
// takes: nobody imports the execution gateway into a discovery package, they
// import a helper that happens to reach it. So the walk follows every
// module-internal edge out of this package until it either exhausts the graph or
// arrives somewhere forbidden, and it prints the chain it took.
//
// The walk stops at the module boundary. An external module cannot import this
// module's internal packages, so nothing outside can be a route back to an order
// path — and reading the whole dependency tree of modernc.org/sqlite would be a
// slow way to learn that.
//
// # Why the source is parsed rather than `go list` run
//
// Same technique as fsguard_drift_test.go and internal/console/static_test.go:
// the claim is about what is written, so what is read is what is written. Build
// constraints are deliberately ignored — fsprobe_linux.go and fsprobe_other.go
// are never compiled together, and a forbidden import behind any tag is a
// forbidden import. Ignoring the tags makes the guard stricter, which is the safe
// direction for it.

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// forbidden is every package internal/candidate may not reach, and why.
//
// These are the *roots* of order capability. The walk is transitive, so a package
// that reaches one of these is caught by the chain rather than by being listed:
// internal/reconcile and internal/filldetect arrive through internal/execgw,
// internal/soak and internal/ops through the clients, internal/position through
// internal/riskcalc. Listing roots keeps the table short enough to be argued
// with, and the failure message names the whole path anyway.
var forbidden = map[string]string{
	// --- placing, amending, cancelling, liquidating -----------------------------
	"internal/execgw":  "the execution gateway: order submission, cancellation and amendment",
	"internal/trading": "the trading-policy service and the conditional-order mutations",
	"internal/orderintent": "PlaceIntent and AmendIntent — the order request itself. " +
		"D7's \"주문 인터페이스를 타입으로도 볼 수 없게\" is this entry",
	"internal/orderlineage": "order lineage ids, which exist only because orders do",
	"internal/flatten":      "flatten-all: the emergency liquidation path",

	// --- the clients that carry the order verbs ----------------------------------
	// internal/candidatesrc exists precisely so that these can be reached from
	// somewhere that is not here: one client per backend, and every one of them
	// carries PlaceOrder, CancelOrder and the conditional-order mutations
	// alongside the rankings this change reads.
	"internal/official": "the Open API client — PlaceOrder, CancelOrder, ModifyOrder",
	"internal/client":   "the WTS web-session client",
	"internal/hybrid":   "the router that sends a mutation to whichever backend answers",
	"internal/candidatesrc": "the wiring package that adapts those clients to Source. " +
		"It imports this package; the dependency runs one way and only one way",

	// --- live verification ---------------------------------------------------------
	"internal/verifylive": "the live-verification runner, which places real orders against a real account",

	// --- the engine's order paths ---------------------------------------------------
	"internal/app/engine": "the engine loop",
	"internal/filldetect": "fill detection, which reaches the exit path and therefore stop-loss immediacy",
	"internal/reconcile":  "reconciliation between the account and the ledger",

	// --- stop-loss, take-profit, sizing ---------------------------------------------
	// The spec forbids the sizing and stop interfaces in the same breath as the
	// order ones: "발굴 패키지는 주문·손절·사이징 인터페이스를 참조해서는 안 되며".
	"internal/exitpolicy": "stop-loss and take-profit policy",
	"internal/risk":       "position sizing",
	"internal/riskcalc":   "the sizing arithmetic",

	// --- the ledger -------------------------------------------------------------------
	// D2 chose to copy the ledger's filesystem guard into fsguard.go rather than
	// import it, for exactly this reason, and paid for the copy with
	// fsguard_drift_test.go. That price is only worth paying if this entry holds.
	"internal/journal": "the intent journal — the order, fill and position ledger",

	// --- the order value objects -------------------------------------------------------
	// domain declares Order, ConditionalOrder and MutationResult. It is also the
	// package source.go declines to take its Row type from, because
	// domain.RankingItem has already turned every absent decimal into 0.0 by the
	// time it exists — the distinction this package is built on.
	"internal/domain": "Order, ConditionalOrder and MutationResult, and decimals that have " +
		"already lost the difference between absent and zero",
}

// modulePath reads the module line out of go.mod, rather than assuming it. A
// renamed module would otherwise make every import look external and the guard
// would pass by checking nothing.
func modulePath(t *testing.T, root string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatalf("reading go.mod: %v", err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		if rest, ok := strings.CutPrefix(strings.TrimSpace(line), "module "); ok {
			return strings.TrimSpace(rest)
		}
	}
	t.Fatal("go.mod declares no module path")
	return ""
}

// moduleRoot is the repository root, two levels up from this package.
func moduleRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolving the repository root: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("%s does not look like the repository root: %v", root, err)
	}
	return root
}

// packageImports returns the module-internal import paths of one package,
// relative to the module (so "internal/execgw", not the full path).
//
// tests selects which half of the package is read: false is what the binary
// compiles, true is the _test.go files.
func packageImports(t *testing.T, root, module, pkg string, tests bool) []string {
	t.Helper()
	dir := filepath.Join(root, filepath.FromSlash(pkg))
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading the package directory %s: %v", pkg, err)
	}

	seen := map[string]bool{}
	var out []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") {
			continue
		}
		if strings.HasSuffix(name, "_test.go") != tests {
			continue
		}
		path := filepath.Join(dir, name)
		file, err := parser.ParseFile(token.NewFileSet(), path, nil,
			parser.ImportsOnly|parser.SkipObjectResolution)
		if err != nil {
			t.Fatalf("parsing %s: %v", path, err)
		}
		for _, spec := range file.Imports {
			imported, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				t.Fatalf("%s has an unreadable import %s", path, spec.Path.Value)
			}
			rel, ok := strings.CutPrefix(imported, module+"/")
			if !ok || seen[rel] {
				continue
			}
			seen[rel] = true
			out = append(out, rel)
		}
	}
	sort.Strings(out)
	return out
}

// reachForbidden walks the module-internal import graph breadth-first from the
// given roots and returns the first chain that arrives somewhere forbidden.
//
// Breadth-first so the reported chain is a shortest one: the point of printing it
// is that somebody has to find the edge to delete, and the shortest path is the
// one with the fewest candidates for it.
func reachForbidden(t *testing.T, root, module string, from []string, rootsAreTests bool) ([]string, string) {
	t.Helper()

	via := map[string]string{}
	queue := make([]string, 0, len(from))
	visited := map[string]bool{}
	for _, pkg := range from {
		if visited[pkg] {
			continue
		}
		visited[pkg] = true
		queue = append(queue, pkg)
		if why, bad := forbidden[pkg]; bad {
			return []string{pkg}, why
		}
	}

	// The starting packages themselves are read with the caller's choice of half;
	// everything downstream is production code, because that is what a compiled
	// dependency is.
	half := rootsAreTests
	for depth := 0; len(queue) > 0; depth++ {
		var next []string
		for _, pkg := range queue {
			for _, imported := range packageImports(t, root, module, pkg, half && depth == 0) {
				if visited[imported] {
					continue
				}
				visited[imported] = true
				via[imported] = pkg
				if why, bad := forbidden[imported]; bad {
					return chainTo(via, imported), why
				}
				next = append(next, imported)
			}
		}
		queue = next
	}
	return nil, ""
}

// chainTo rebuilds the path from the walk's origin to pkg.
func chainTo(via map[string]string, pkg string) []string {
	chain := []string{pkg}
	for {
		parent, ok := via[pkg]
		if !ok {
			break
		}
		chain = append([]string{parent}, chain...)
		pkg = parent
	}
	return chain
}

const discoveryPkg = "internal/candidate"

// TestDiscoveryCannotReachAnOrderPath is the spec scenario "주문 타입은 이 패키지에서
// 이름조차 없다".
//
// It walks what the binary compiles. A failure prints the whole chain, because
// the edge to delete is rarely the last one — a report that only named the
// forbidden package would send the reader looking for an import that is not
// there.
func TestDiscoveryCannotReachAnOrderPath(t *testing.T) {
	root := moduleRoot(t)
	module := modulePath(t, root)

	// The guard has to be looking at something. A walk that read no files would
	// pass silently, which is the failure mode every static test in this
	// repository guards against first.
	if imports := packageImports(t, root, module, discoveryPkg, false); len(imports) == 0 {
		t.Fatalf("%s appears to import nothing from this module; the walk is reading no files "+
			"(it should at least reach internal/clock)", discoveryPkg)
	}

	chain, why := reachForbidden(t, root, module, []string{discoveryPkg}, false)
	if chain != nil {
		t.Errorf("discovery reaches a forbidden package:\n\n\t%s\n\n%s is %s.\n\n"+
			"spec candidate-discovery: 발굴 패키지는 주문·손절·사이징 인터페이스를 참조해서는 안 되며, "+
			"후보에서 주문으로 가는 코드가 컴파일되지 않아야 한다. The market-data clients are "+
			"reached from internal/candidatesrc, which imports this package rather than the "+
			"other way round.",
			strings.Join(chain, "\n\t  → "), chain[len(chain)-1], why)
	}
}

// TestTheDiscoveryTestsCannotReachAnOrderPathEither.
//
// A _test.go file is not compiled into tossctl, so it is outside the requirement
// as written — and it is inside the safety invariant that matters, because a test
// binary that can name PlaceOrder is a test that can place one. §0.1 of
// docs/WORKFLOW.md forbids an unapproved live order side effect during
// development, and this package's tests run on every `make test`.
func TestTheDiscoveryTestsCannotReachAnOrderPathEither(t *testing.T) {
	root := moduleRoot(t)
	module := modulePath(t, root)

	if imports := packageImports(t, root, module, discoveryPkg, true); len(imports) == 0 {
		t.Fatalf("no module-internal imports were read from %s's tests; the walk is reading "+
			"no files", discoveryPkg)
	}

	chain, why := reachForbidden(t, root, module, []string{discoveryPkg}, true)
	if chain != nil {
		t.Errorf("a discovery test reaches a forbidden package:\n\n\t%s\n\n%s is %s.\n\n"+
			"A test that can name an order is a test that can place one.",
			strings.Join(chain, "\n\t  → "), chain[len(chain)-1], why)
	}
}

// TestTheImportWalkerFindsAForbiddenPathWhenThereIsOne is the guard's own guard.
//
// Every static test in this repository can fail in one direction only —
// forbidding something is easy, and a guard that has quietly stopped reading
// files forbids nothing while reporting success. So the walker is pointed at a
// package that is known to reach a forbidden one, and required to find it.
//
// internal/obs is the fixture on purpose: it is not itself in the table, and it
// imports internal/execgw. A one-hop check against internal/candidate's direct
// imports would pass on a `candidate → obs` edge, and the whole reason this walk
// is transitive is that this is the shape a real regression takes.
func TestTheImportWalkerFindsAForbiddenPathWhenThereIsOne(t *testing.T) {
	root := moduleRoot(t)
	module := modulePath(t, root)

	const fixture = "internal/obs"
	if _, listed := forbidden[fixture]; listed {
		t.Fatalf("%s is now in the forbidden table, so it can no longer serve as the "+
			"positive control; pick another package that reaches one transitively", fixture)
	}
	chain, why := reachForbidden(t, root, module, []string{fixture}, false)
	if chain == nil {
		t.Fatalf("the walker found no forbidden package from %s, which imports internal/execgw. "+
			"The walk is not following edges, and every other test in this file is passing "+
			"because it is reading nothing", fixture)
	}
	if len(chain) < 2 {
		t.Errorf("the chain %v has no intermediate hop; the positive control is meant to prove "+
			"the walk is transitive", chain)
	}
	t.Logf("positive control: %s (%s)", strings.Join(chain, " → "), why)
}

// TestDiscoveryDependsOnNothingItHasNotArguedFor.
//
// The forbidden table names what is known to be dangerous today. This one asks
// the question the table cannot: is there a new edge at all? A package that grows
// a dependency nobody discussed is how the table goes out of date, and a diff to
// this list is where the discussion happens.
//
// internal/clock is the whole allowance, and it is the one D5 requires: the time
// axis is this package's subject, so the clock has to be injectable or no test
// can make time pass.
func TestDiscoveryDependsOnNothingItHasNotArguedFor(t *testing.T) {
	root := moduleRoot(t)
	module := modulePath(t, root)

	allowed := map[string]string{
		"internal/clock": "D5 — the injected clock. Nothing here calls time.Now()",
	}
	// The tests reach one package the production code does not: enginelock, which
	// is how TestAScanningStoreDoesNotBlockTheEngineFromStarting asks the engine
	// to take its lock while a scan is running (D2). It reaches internal/binstamp
	// and internal/runlock and nothing else.
	allowedInTests := map[string]string{
		"internal/enginelock": "D2 — the test that proves a scan does not block `tossctl engine run`",
		"internal/binstamp":   "reached through enginelock",
		"internal/runlock":    "reached through enginelock",
	}

	for _, tests := range []bool{false, true} {
		seen := map[string]bool{discoveryPkg: true}
		queue := []string{discoveryPkg}
		for depth := 0; len(queue) > 0; depth++ {
			var next []string
			for _, pkg := range queue {
				for _, imported := range packageImports(t, root, module, pkg, tests && depth == 0) {
					if seen[imported] {
						continue
					}
					seen[imported] = true
					next = append(next, imported)
					if _, ok := allowed[imported]; ok {
						continue
					}
					if _, ok := allowedInTests[imported]; tests && ok {
						continue
					}
					where := "the package"
					if tests {
						where = "the package's tests"
					}
					t.Errorf("%s now depends on %s, which is not in this test's allowlist. "+
						"Add it here with the reason, or do not add the import — a dependency "+
						"nobody argued for is how the forbidden table goes out of date",
						where, imported)
				}
			}
			queue = next
		}
		if !seen["internal/clock"] {
			t.Error("internal/clock is no longer reachable; either the walk is broken or the " +
				"injected clock is gone, and D5 needs it to exist")
		}
	}
}
