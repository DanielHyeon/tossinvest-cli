package verifylive

// static_test.go asserts the properties no runtime test can observe, because the
// offending code would simply never be executed by a test.
//
// It is internal/soak/static_test.go's technique pointed at the opposite problem.
// There, the claim was "this package contains no mutation transport at all", so
// the import graph was the assertion. Here the package exists in order to mutate,
// so the claims are narrower and there are three of them:
//
//	one door        every call to a mutating Broker method lives in mutate.go,
//	                next to the confirmation, the exposure check and the evidence
//	                line. A step cannot quietly place an order.
//	no bypass       nothing in the package reads an environment variable or a
//	                flag that could stand in for a typed confirmation.
//	no real host    the test binary cannot reach Toss. That one has a runtime
//	                half too (TestGuardBlocksTheRealBroker), because the guard is
//	                only worth anything if it is actually installed.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/testenv"
)

const modulePath = "github.com/JungHoonGhae/tossinvest-cli/"

func repoRoot(t *testing.T) string {
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

// packageFiles returns this package's non-test source files.
func packageFiles(t *testing.T, includeTests bool) map[string]string {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("reading the package directory: %v", err)
	}
	out := map[string]string{}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") {
			continue
		}
		if !includeTests && strings.HasSuffix(name, "_test.go") {
			continue
		}
		if includeTests && !strings.HasSuffix(name, "_test.go") {
			continue
		}
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("reading %s: %v", name, err)
		}
		out[name] = string(data)
	}
	if len(out) == 0 {
		t.Fatal("no source files were found; the guard is not looking at the package")
	}
	return out
}

// TestEveryMutationGoesThroughMutateGo.
//
// The confirmation, the exposure cap and the evidence line are all in mutate.go's
// wrappers. A step that called r.broker.PlaceOrder directly would skip all three
// and nothing at runtime would notice, because the test suite would simply never
// run that line — which is exactly the case a source-level assertion is for.
func TestEveryMutationGoesThroughMutateGo(t *testing.T) {
	const door = "mutate.go"
	for name, src := range packageFiles(t, false) {
		if name == door || name == "verifylive.go" {
			// verifylive.go declares the Broker interface and names the methods
			// in MutationMethods; it calls none of them.
			continue
		}
		for _, method := range MutationMethods() {
			if callsMethod(t, name, src, method) {
				t.Errorf("%s calls %s. Every mutation must go through %s, where the typed confirmation, "+
					"the live-exposure cap and the evidence line are", name, method, door)
			}
		}
	}
}

// callsMethod reports a real call expression rather than a mention, by parsing
// the file. A comment that explains what a function is avoiding has to remain
// legal, and a string in an error message must not trip the guard.
func callsMethod(t *testing.T, name, src, method string) bool {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, name, src, 0)
	if err != nil {
		t.Fatalf("parsing %s: %v", name, err)
	}
	found := false
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if sel.Sel.Name == method {
			found = true
		}
		return true
	})
	return found
}

// TestNoAutomationBypassExists.
//
// tasks.md forbids an automation flag on the confirmation ("자동화 플래그 금지"),
// and an environment variable is the same thing with a different spelling. This
// checks the package reads none, and that the only place a confirmation can be
// satisfied is Confirm's typed comparison.
func TestNoAutomationBypassExists(t *testing.T) {
	banned := []string{
		"os.Getenv", "os.LookupEnv",
		"TOSSCTL_VERIFY_YES", "SKIP_CONFIRM", "AUTO_APPROVE", "NO_CONFIRM",
	}
	for name, src := range packageFiles(t, false) {
		code := strings.Join(nonCommentLines(src), "\n")
		for _, b := range banned {
			if strings.Contains(code, b) {
				t.Errorf("%s contains %q; a confirmation this tool asks for must be typed by a person "+
					"and nothing may stand in for it", name, b)
			}
		}
	}
	// os.Getpid is the one os call the package needs, and record.go is the only
	// file allowed to make it.
	for name, src := range packageFiles(t, false) {
		if name == "record.go" {
			continue
		}
		if strings.Contains(strings.Join(nonCommentLines(src), "\n"), "os.") {
			t.Errorf("%s reaches into os; only record.go needs to (the process identity and the record file)", name)
		}
	}
}

// TestConfirmIsTheOnlyPathToATrueVerification. Verify's comparison is the gate;
// a second function that returned nil for some other input would be a back door.
//
// Both gates are checked, because the batch approval is now the default one: a
// batch that accepted anything but its own nonce, or that printed its list to
// something that is not a terminal, would be the back door with extra steps.
func TestConfirmIsTheOnlyPathToATrueVerification(t *testing.T) {
	src, err := os.ReadFile("confirm.go")
	if err != nil {
		t.Fatalf("reading confirm.go: %v", err)
	}
	code := strings.Join(nonCommentLines(string(src)), "\n")
	if n := strings.Count(code, "if !interactive {"); n != 2 {
		t.Errorf("%d of the two confirmations refuse a non-terminal; both must", n)
	}
	if !strings.Contains(code, "strings.TrimSpace(input) != m.Nonce") {
		t.Error("Mutation.Verify no longer compares the typed input against the nonce")
	}
	if !strings.Contains(code, "strings.TrimSpace(input) != b.Nonce") {
		t.Error("Batch.Verify no longer compares the typed input against the nonce")
	}
	for _, banned := range []string{"yes", "\"y\"", "force", "true ==", "Assume"} {
		if strings.Contains(strings.ToLower(code), strings.ToLower(banned)) {
			t.Errorf("confirm.go mentions %q outside a comment; there is no shortcut answer", banned)
		}
	}
}

// TestTheApprovedPlanIsTheOnlyThingMutateGoActsOn.
//
// The batch model's whole safety claim is that a request the operator's list does
// not carry cannot be sent. That is one function — Runner.authorise — and every
// mutation wrapper has to go through it, either directly (the two replays) or
// through gate. A wrapper added later that called the broker without it would be a
// mutation nobody approved, and no runtime test would notice because the test suite
// would simply never execute that line.
func TestTheApprovedPlanIsTheOnlyThingMutateGoActsOn(t *testing.T) {
	src, err := os.ReadFile("mutate.go")
	if err != nil {
		t.Fatalf("reading mutate.go: %v", err)
	}
	code := strings.Join(nonCommentLines(string(src)), "\n")
	if !strings.Contains(code, "if r.plan == nil {") {
		t.Error("authorise no longer fails closed when there is no approved plan")
	}
	if !strings.Contains(code, "r.plan.Authorises(") {
		t.Error("authorise no longer consults the approved plan")
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "mutate.go", src, 0)
	if err != nil {
		t.Fatalf("parsing mutate.go: %v", err)
	}
	mutating := map[string]bool{}
	for _, m := range MutationMethods() {
		mutating[m] = true
	}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		var sends, gated bool
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			switch {
			case mutating[sel.Sel.Name]:
				sends = true
			case sel.Sel.Name == "gate", sel.Sel.Name == "authorise":
				gated = true
			}
			return true
		})
		if sends && !gated {
			t.Errorf("%s calls a mutating broker method without going through gate or authorise; it could "+
				"send a request the operator never approved", fn.Name.Name)
		}
	}
}

// TestTestBinaryInstallsTheRealHostGuard.
//
// The runtime half below only proves anything if the guard is actually installed
// for this package, and that happens once, in TestMain. Asserting it in the
// source means deleting TestMain fails a test rather than silently unguarding the
// suite.
func TestTestBinaryInstallsTheRealHostGuard(t *testing.T) {
	tests := packageFiles(t, true)
	installed := false
	for _, src := range tests {
		code := strings.Join(nonCommentLines(src), "\n")
		if strings.Contains(code, "func TestMain(") && strings.Contains(code, "testenv.Guard{") {
			installed = true
		}
	}
	if !installed {
		t.Fatal("no TestMain in this package installs internal/testenv's guard: a test could reach the live API")
	}
	// And no test may build an official client without pointing it somewhere.
	for name, src := range tests {
		code := strings.Join(nonCommentLines(src), "\n")
		if strings.Contains(code, "official.New(") && !strings.Contains(code, "WithBaseURL(") {
			t.Errorf("%s builds an official client with no WithBaseURL; tests must use httptest", name)
		}
	}
}

// TestGuardBlocksTheRealBroker is the runtime half: the transport this test
// binary is running with refuses a POST to the live order endpoint.
//
// The hostname is spelled out rather than imported, because the constant is
// unexported and because the thing being asserted is "requests to Toss are
// blocked", not "requests to whatever the client happens to be configured with".
func TestGuardBlocksTheRealBroker(t *testing.T) {
	guard, ok := http.DefaultTransport.(*testenv.Guard)
	if !ok {
		t.Fatalf("http.DefaultTransport is %T, not the testenv guard TestMain installs", http.DefaultTransport)
	}
	for _, target := range []string{
		"https://openapi.tossinvest.com/api/v1/orders",
		"https://wts-api.tossinvest.com/api/v2/orders",
	} {
		req, err := http.NewRequest(http.MethodPost, target, nil)
		if err != nil {
			t.Fatalf("building the probe request: %v", err)
		}
		// A bare Guard value, so the block is returned rather than failing this
		// test through the installed OnBlock.
		probe := &testenv.Guard{Base: guard.Base}
		if _, err := probe.RoundTrip(req); err == nil {
			t.Errorf("a POST to %s was not blocked", target)
		} else if !strings.Contains(err.Error(), "BLOCKED") {
			t.Errorf("a POST to %s failed with %v, not the mutation block", target, err)
		}
	}
}

// TestPackageDoesNotImportTheEngineOrTheWebSession.
//
// Order execution goes through the official Open API only (불변 규칙), and this
// tool has no business holding an engine, a journal or a scraped web session: it
// is a measurement harness, and every one of those would give it a second way to
// reach an account.
func TestPackageDoesNotImportTheEngineOrTheWebSession(t *testing.T) {
	root := repoRoot(t)
	_ = root
	banned := []string{
		"internal/app/engine",
		"internal/client",  // the web-session client
		"internal/hybrid",  // routes writes to whichever backend answers
		"internal/execgw",  // the engine's gateway; this tool journals nothing
		"internal/flatten", // the liquidation saga
	}
	for name, src := range packageFiles(t, false) {
		for _, b := range banned {
			if strings.Contains(src, `"`+modulePath+b+`"`) {
				t.Errorf("%s imports %s", name, b)
			}
		}
	}
}

// TestTheProcessBoundaryIsNeverJudgedByPID is the structural half of
// TestTheProcessBoundaryIsTheInstanceIDAndNotThePID (tasks.md 1.8 ①).
//
// Process.PID exists for the audit trail and for the banner. It must not be what
// any judgement reads: a PID is reused by the operating system, and a re-exec —
// which is exactly what the console's restart button does — keeps it. Both
// directions are wrong, in opposite ways: a reused PID would let a new process be
// mistaken for the registering one, and a preserved PID would stop a genuinely new
// process from finishing the measurement.
func TestTheProcessBoundaryIsNeverJudgedByPID(t *testing.T) {
	// The banner prints it and NewProcess fills it in. Everywhere else, reading it
	// is a comparison waiting to happen.
	allowed := map[string]bool{"record.go": true, "runner.go": true}
	for name, src := range packageFiles(t, false) {
		if allowed[name] {
			continue
		}
		for _, line := range nonCommentLines(src) {
			if strings.Contains(line, "process.PID") || strings.Contains(line, "Process.PID") {
				t.Errorf("%s reads the PID (%q). The process boundary is Process.InstanceID: a PID is reused, "+
					"and a re-exec keeps it", name, strings.TrimSpace(line))
			}
		}
	}
	// runner.go may print it; it may not compare it.
	for _, line := range nonCommentLines(packageFiles(t, false)["runner.go"]) {
		if !strings.Contains(line, "process.PID") {
			continue
		}
		if !strings.Contains(line, "Fprint") {
			t.Errorf("runner.go uses the PID outside an output line (%q); it is audit trail, not a judgement",
				strings.TrimSpace(line))
		}
	}
	// And the judgement itself is still spelled against the instance identifier.
	if !strings.Contains(strings.Join(nonCommentLines(packageFiles(t, false)["steps.go"]), "\n"),
		"registrar == r.process.InstanceID") {
		t.Error("steps.go no longer compares the registering instance identifier; " +
			"conditional-persist has lost its process boundary")
	}
}

// nonCommentLines drops whole-line comments. It does not parse Go: a line that is
// wrongly kept makes the guard stricter, which is the safe direction here.
func nonCommentLines(src string) []string {
	var out []string
	for _, line := range strings.Split(src, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "//") {
			continue
		}
		out = append(out, line)
	}
	return out
}
