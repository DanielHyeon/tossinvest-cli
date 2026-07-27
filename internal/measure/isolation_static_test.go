package measure

import (
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"
	"testing"
)

// isolation_static_test.go is change add-net-rr-measurement tasks 2.5b and 2.8,
// enforced by layout rather than by care.
//
// Both rules are of a kind a behavioural test cannot hold. A test can show that
// today nobody registers the reconstruction job as a supervised loop; it cannot
// show that tomorrow nobody will, and the edit that did it would compile and pass
// everything. So the rules are scans, in the style of
// internal/journal/adoption_static_test.go: they fail if the import appears at
// all, and a second writer of that dependency cannot arrive without somebody
// editing this file and justifying it in review.
//
// # 2.5b — the reconstruction job is not a supervised loop
//
// engine.Runtime's supervisor stops every loop when one returns (§0.3), so a
// measurement fault would take the exit observer and the fill detector with it.
// Its degradation threshold escalates to ENTRY_BLOCKED, and SupervisedLoop.Trigger
// is validated against journal's closed enumeration — so a measurement loop would
// have to borrow RECONCILE_CYCLE_FAILURE or FILL_DETECTION_CYCLE_FAILURE and
// misattribute its own fault as one of those. The package cannot see
// internal/app/engine at all, so it cannot construct a SupervisedLoop.
//
// # 2.8 — the degradation counter is not in the journal's failure domain
//
// internal/measure/degrade must not be able to reach a database handle. The
// behavioural proof is internal/journal/observation_degradation_test.go, which
// fills a real journal and shows the count landing anyway; this is the structural
// half, and it is why the counter is a package rather than a type in this one.

// allowedImports is the dependency budget, per file. Standard library imports are
// unrestricted; these are the internal packages each file may name.
//
// Adding an entry is a design decision. internal/app/engine must never appear.
var allowedImports = map[string][]string{
	"reconstruct.go": {
		// Reading decisions and writing observations is the job.
		"github.com/JungHoonGhae/tossinvest-cli/internal/journal",
		// The loss counter, which lives in its own package precisely so that it
		// carries no journal dependency of its own.
		"github.com/JungHoonGhae/tossinvest-cli/internal/measure/degrade",
	},
	// The harness (task 5.6). Its budget is the whole isolation guarantee: the
	// chain it evaluates and the cost model that chain needs, and nothing that
	// could place an order, write a row or reach an account. Note the absence of
	// internal/journal in particular — the harness runs thousands of grid points
	// and must not need a database to do it.
	"counterfactual.go": {
		"github.com/JungHoonGhae/tossinvest-cli/internal/costs",
		"github.com/JungHoonGhae/tossinvest-cli/internal/risk",
		"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc",
	},
	"population.go": {
		"github.com/JungHoonGhae/tossinvest-cli/internal/costs",
		"github.com/JungHoonGhae/tossinvest-cli/internal/risk",
	},
}

// harnessFiles are the files task 5.6's guarantee is about, checked against a
// deny-list as well as their budgets. The budget already excludes these by
// omission; naming them is what makes the failure message say *why*.
var harnessFiles = []string{"counterfactual.go", "population.go"}

// forbiddenInHarness must never appear in a harness file, whatever anyone's
// budget says. Each one is a way the analysis path could reach the execution path.
var forbiddenInHarness = []string{
	// Would let the harness write to the ledger.
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal",
	// Would let it issue a decision or submit an order.
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw",
	// Would let it reach a real broker endpoint.
	"github.com/JungHoonGhae/tossinvest-cli/internal/official",
	"github.com/JungHoonGhae/tossinvest-cli/internal/client",
	// Would let it read credentials or a session.
	"github.com/JungHoonGhae/tossinvest-cli/internal/auth",
	// Network, at all. The harness is arithmetic over values.
	"net/http",
}

// degradeBudget is the counter package's, and it is empty on purpose: task 2.8's
// independence is the claim that no database handle is reachable from that code.
var degradeBudget = map[string][]string{
	"degrade.go": {},
}

// forbiddenEverywhere are packages no file in this package may import, whatever
// the budget above says.
var forbiddenEverywhere = []string{
	// The supervisor. Task 2.5b: a measurement job registered here would kill the
	// exit observer on its own fault and escalate to ENTRY_BLOCKED under somebody
	// else's trigger.
	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine",
}

func TestPackageImportsStayInsideTheBudget(t *testing.T) {
	scanImports(t, ".", allowedImports)
}

// TestTheCounterPackageReachesNoDatabase is task 2.8's structural half.
func TestTheCounterPackageReachesNoDatabase(t *testing.T) {
	scanImports(t, "degrade", degradeBudget)
}

// TestTheHarnessCannotReachTheExecutionPath is task 5.6.
//
// trade-analytics requires that the measurement "주문을 만들지 않고, 원장에 쓰지
// 않고, 실계좌에 접근하지 않는다" (SHALL NOT). A behavioural test can show that
// today's harness does none of those; it cannot show that a later edit will not,
// and the edit that did it would be one import line. So the guarantee is that the
// capability is absent — the harness cannot name a journal, a gateway, a broker
// client, a credential store, or net/http, so there is no code path from it to any
// of them.
//
// It also cannot reach them transitively through what it does import: internal/risk
// is a pure function over values (execgw's observation_scope_test.go pins that it
// imports no storage), and internal/costs is arithmetic over rates.
func TestTheHarnessCannotReachTheExecutionPath(t *testing.T) {
	fset := token.NewFileSet()
	for _, name := range harnessFiles {
		file, err := parser.ParseFile(fset, name, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parsing %s: %v", name, err)
		}
		for _, spec := range file.Imports {
			path, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				t.Fatal(err)
			}
			for _, bad := range forbiddenInHarness {
				if path == bad || strings.HasPrefix(path, bad+"/") {
					t.Errorf("%s imports %s. The counterfactual harness must not be able to "+
						"place an order, write to the ledger, or reach a real account — the "+
						"measurement is arithmetic over values", name, path)
				}
			}
		}
	}
}

func scanImports(t *testing.T, dir string, budgets map[string][]string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading %s: %v", dir, err)
	}
	seen := 0
	fset := token.NewFileSet()
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		seen++
		file, err := parser.ParseFile(fset, dir+"/"+name, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parsing %s: %v", name, err)
		}
		budget, listed := budgets[name]
		if !listed {
			t.Errorf("%s is not in the dependency budget; a new file here needs its "+
				"budget stated before it can carry an import", name)
			continue
		}
		for _, spec := range file.Imports {
			path, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				t.Fatalf("%s: unquoting %s: %v", name, spec.Path.Value, err)
			}
			if !strings.HasPrefix(path, "github.com/JungHoonGhae/tossinvest-cli/") {
				continue // standard library and vendored deps are unrestricted
			}
			for _, forbidden := range forbiddenEverywhere {
				if path == forbidden || strings.HasPrefix(path, forbidden+"/") {
					t.Errorf("%s imports %s. The reconstruction job must not be registrable as a "+
						"SupervisedLoop: the supervisor's return semantics stop every other loop "+
						"(§0.3) and its degradation threshold escalates to ENTRY_BLOCKED under a "+
						"trigger this job would have to borrow", name, path)
				}
			}
			if !allowed(budget, path) {
				t.Errorf("%s imports %s, which is outside its budget %v", name, path, budget)
			}
		}
	}
	if seen == 0 {
		t.Fatalf("scanned no sources in %s; the path is wrong and this test proves nothing", dir)
	}
}

func allowed(budget []string, path string) bool {
	for _, p := range budget {
		if p == path {
			return true
		}
	}
	return false
}

// TestTheRuntimeCannotSeeThisPackage closes the other direction. The scan above
// stops this package reaching the supervisor; on its own that leaves the
// supervisor free to reach *here* — engine.Runtime could import measure and wrap
// Run in a SupervisedLoop without this package changing at all.
//
// The test lives here rather than in internal/app/engine because it is this
// change's invariant, and somebody adding the loop would be reading this file's
// header to find out why they should not.
func TestTheRuntimeCannotSeeThisPackage(t *testing.T) {
	const engineDir = "../app/engine"
	const self = "github.com/JungHoonGhae/tossinvest-cli/internal/measure"

	entries, err := os.ReadDir(engineDir)
	if err != nil {
		t.Fatalf("reading %s: %v", engineDir, err)
	}
	fset := token.NewFileSet()
	scanned := 0
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") {
			continue
		}
		scanned++
		file, err := parser.ParseFile(fset, engineDir+"/"+name, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parsing %s: %v", name, err)
		}
		for _, spec := range file.Imports {
			path, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				t.Fatal(err)
			}
			if path == self || strings.HasPrefix(path, self+"/") {
				t.Errorf("internal/app/engine/%s imports %s. The reconstruction job runs from a "+
					"separate process or schedule boundary; wrapping it in a SupervisedLoop makes "+
					"a measurement fault stop the exit observer and the fill detector", name, self)
			}
		}
	}
	if scanned == 0 {
		t.Fatal("scanned no engine sources; the path is wrong and this test proves nothing")
	}
}

// TestNoModeTriggerIsNamedAnywhere is the other half of 2.5b (SHALL NOT — 그 실패는
// 어떤 운영 모드 트리거에도 사상되어서는 안 되며 기존 트리거 이름을 재사용해서도 안 된다).
//
// The import scan stops this package building a SupervisedLoop. This stops it
// naming a trigger by its string value, which is the way the rule could be broken
// without importing anything: a caller elsewhere could read a trigger constant out
// of this package and register the job themselves.
func TestNoModeTriggerIsNamedAnywhere(t *testing.T) {
	// journal's closed enumeration, spelled out. Written as literals on purpose:
	// importing them to compare against would be this package naming them.
	triggers := []string{
		"DAILY_LOSS_LIMIT_REACHED",
		"BROKER_AUTH_REJECTED",
		"CRITICAL_ALERT_UNDELIVERED",
		"EXIT_OBSERVATION_OUTAGE",
		"RECONCILE_CYCLE_FAILURE",
		"FILL_DETECTION_CYCLE_FAILURE",
	}
	for _, dir := range []string{".", "degrade"} {
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("reading %s: %v", dir, err)
		}
		for _, e := range entries {
			name := e.Name()
			if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
				continue
			}
			source, err := os.ReadFile(dir + "/" + name)
			if err != nil {
				t.Fatalf("reading %s: %v", name, err)
			}
			text := string(source)
			for _, trigger := range triggers {
				// The file header explains why the job is not a supervised loop
				// and names the failure mode; a comment mentioning a trigger is
				// the documentation working, not the rule breaking. Only code
				// counts.
				if namesInCode(text, trigger) {
					t.Errorf("%s/%s names the operating-mode trigger %q. This job's failure must "+
						"not be attributable to any of them — borrowing one would report a "+
						"measurement fault as a reconciliation outage", dir, name, trigger)
				}
			}
		}
	}
}

// namesInCode reports whether the needle appears outside a comment.
func namesInCode(source, needle string) bool {
	for _, line := range strings.Split(source, "\n") {
		code, _, _ := strings.Cut(line, "//")
		if strings.Contains(code, needle) {
			return true
		}
	}
	return false
}
