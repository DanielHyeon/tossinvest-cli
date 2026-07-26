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
func TestConfirmIsTheOnlyPathToATrueVerification(t *testing.T) {
	src, err := os.ReadFile("confirm.go")
	if err != nil {
		t.Fatalf("reading confirm.go: %v", err)
	}
	code := strings.Join(nonCommentLines(string(src)), "\n")
	if !strings.Contains(code, "if !interactive {") {
		t.Error("Confirm no longer refuses a non-terminal")
	}
	if !strings.Contains(code, "strings.TrimSpace(input) != m.Nonce") {
		t.Error("Verify no longer compares the typed input against the nonce")
	}
	for _, banned := range []string{"yes", "\"y\"", "force", "true ==", "Assume"} {
		if strings.Contains(strings.ToLower(code), strings.ToLower(banned)) {
			t.Errorf("confirm.go mentions %q outside a comment; there is no shortcut answer", banned)
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
