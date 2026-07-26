package console

// static_test.go asserts the properties a runtime test cannot see, because the
// code that would violate them is never executed by a test until the day it is.
//
// It is internal/verifylive/static_test.go's technique aimed at this package's own
// claims, and there are five:
//
//	one door        every route goes through the session gate, and every
//	                state-changing route also goes through the CSRF gate. A handler
//	                registered without one would be a page — or an approval —
//	                reachable by anybody who can open a socket on this machine.
//	borrowed judge  the nonce comparison is verifylive.Batch.Verify's. A second
//	                comparison here could drift into accepting what the terminal
//	                would refuse.
//	loopback        the address is spelled once, and it is 127.0.0.1.
//	no bypass       nothing reads an environment variable, and nothing can preset
//	                the session token.
//	no broker       this package cannot reach an account except through the
//	                StartVerify it is handed.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

const modulePath = "github.com/JungHoonGhae/tossinvest-cli/"

// packageFiles returns this package's non-test source files.
func packageFiles(t *testing.T) map[string]string {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("reading the package directory: %v", err)
	}
	out := map[string]string{}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
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

func parseFile(t *testing.T, name, src string) *ast.File {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), name, src, 0)
	if err != nil {
		t.Fatalf("parsing %s: %v", name, err)
	}
	return file
}

// route is one mux registration, as the source declares it.
type route struct {
	Path      string
	Session   bool
	CSRFGated bool
}

// registeredRoutes reads Console.routes out of the source.
func registeredRoutes(t *testing.T) []route {
	t.Helper()
	src := packageFiles(t)["console.go"]
	if src == "" {
		t.Fatal("console.go is missing; the routing table cannot be checked")
	}
	file := parseFile(t, "console.go", src)

	var routes []route
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "HandleFunc" || len(call.Args) != 2 {
			return true
		}
		lit, ok := call.Args[0].(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			t.Errorf("a route is registered with a non-literal path: %v", call.Args[0])
			return true
		}
		r := route{Path: strings.Trim(lit.Value, `"`)}
		if outer, ok := call.Args[1].(*ast.CallExpr); ok {
			if fn, ok := outer.Fun.(*ast.SelectorExpr); ok && fn.Sel.Name == "session0" {
				r.Session = true
			}
		}
		ast.Inspect(call.Args[1], func(inner ast.Node) bool {
			if c, ok := inner.(*ast.CallExpr); ok {
				if fn, ok := c.Fun.(*ast.SelectorExpr); ok && fn.Sel.Name == "mutating" {
					r.CSRFGated = true
				}
			}
			return true
		})
		routes = append(routes, r)
		return true
	})
	if len(routes) == 0 {
		t.Fatal("no routes were found; the guard is not reading the routing table")
	}
	return routes
}

// TestEveryRouteGoesThroughTheSessionGate.
//
// The session token is the console's whole authentication: possession of the
// terminal that printed it. A handler registered without session0 in front of it
// would be reachable by anything on this machine that can open a socket, which on
// a developer's laptop includes every browser tab and every agent.
func TestEveryRouteGoesThroughTheSessionGate(t *testing.T) {
	routes := registeredRoutes(t)
	for _, r := range routes {
		if !r.Session {
			t.Errorf("%s is registered without session0; it is reachable without the session token", r.Path)
		}
	}
	if len(routes) < 9 {
		t.Errorf("only %d route(s) were read; the guard is not seeing the whole table", len(routes))
	}
}

// TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate.
//
// The approval flow is the only state change this console has, and every part of
// it must be a POST carrying the CSRF token. A read route must NOT be behind that
// gate, because a dashboard that only answers POSTs is a dashboard nobody can
// open — the two lists are asserted against each other so neither can drift.
func TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate(t *testing.T) {
	stateChanging := map[string]bool{
		"/verify/start":   true,
		"/verify/approve": true,
		"/verify/abort":   true,
		// The two restarts (task 1.8). Neither touches an account, but both are
		// acts rather than readings, and a page that could be made to restart the
		// console by embedding an image would be a denial of service with a nice
		// interface.
		"/restart":      true,
		"/soak/restart": true,
	}
	seen := map[string]bool{}
	for _, r := range registeredRoutes(t) {
		seen[r.Path] = true
		switch {
		case stateChanging[r.Path] && !r.CSRFGated:
			t.Errorf("%s changes state but is not behind the CSRF gate", r.Path)
		case !stateChanging[r.Path] && r.CSRFGated:
			t.Errorf("%s is a read route behind the CSRF gate; it would be unopenable", r.Path)
		}
	}
	for path := range stateChanging {
		if !seen[path] {
			t.Errorf("%s is not registered at all", path)
		}
	}
}

// TestTheNonceIsJudgedByVerifylive.
//
// The equivalence claim in tasks.md 1.6 is that the web approval is the same act
// as the terminal one. It can only stay true if the same code decides: a
// comparison written here would be a second definition of "correct", and the two
// would drift the first time either changed.
func TestTheNonceIsJudgedByVerifylive(t *testing.T) {
	src := packageFiles(t)["pages.go"]
	code := strings.Join(nonCommentLines(src), "\n")
	if !strings.Contains(code, "view.Batch.Verify(") {
		t.Error("pages.go no longer asks verifylive.Batch to judge the typed nonce")
	}
	for _, banned := range []string{"== view.Batch.Nonce", "Nonce ==", "== nonce"} {
		if strings.Contains(code, banned) {
			t.Errorf("pages.go compares the nonce itself (%q); verifylive.Batch.Verify is the only judge", banned)
		}
	}
	// And nothing in the package invents an approval.
	for name, fileSrc := range packageFiles(t) {
		body := strings.Join(nonCommentLines(fileSrc), "\n")
		for _, banned := range []string{"AUTO_APPROVE", "SKIP_CONFIRM", "NO_CONFIRM", "TOSSCTL_CONSOLE_TOKEN"} {
			if strings.Contains(body, banned) {
				t.Errorf("%s contains %q; the approval must be typed by a person", name, banned)
			}
		}
	}
}

// TestTheRedoSetIsReadFromTheRecordAndNeverFromTheRequest.
//
// The re-measurement is the one place this console decides which live requests a
// run may make (task 1.7). That decision is verifylive.RedoSet's, taken against
// the evidence file; a step id that could arrive in a form field would be a way to
// aim a second live order at a step that already passed.
//
// The runtime half is TestTheRedoSetComesFromTheRecordAndNotFromTheForm. This is
// the half that survives somebody adding a "convenient" parameter later.
func TestTheRedoSetIsReadFromTheRecordAndNeverFromTheRequest(t *testing.T) {
	code := strings.Join(nonCommentLines(packageFiles(t)["pages.go"]), "\n")
	if !strings.Contains(code, "c.redoSet()") {
		t.Error("handleStart no longer asks the record which steps may be re-measured")
	}
	for _, banned := range []string{
		"verifylive.StepID(", // a string from anywhere becoming a step id
		"r.PostForm[", "r.Form[", "r.URL.Query().Get(\"step",
	} {
		if strings.Contains(code, banned) {
			t.Errorf("pages.go contains %q; the redo set comes from the evidence record, not the request", banned)
		}
	}

	data := strings.Join(nonCommentLines(packageFiles(t)["data.go"]), "\n")
	if !strings.Contains(data, "verifylive.RedoSet(") {
		t.Error("data.go no longer uses verifylive.RedoSet; the console must not define its own redo rule")
	}
}

// TestTheMarketHoursAdvisoryCannotBlockAnything.
//
// tasks.md 1.7 ②: "advisory만, 하드 차단 금지 — 주문 접수 창은 [미측정]". The
// advisory is rendered and never consulted, so no handler may branch on it.
func TestTheMarketHoursAdvisoryCannotBlockAnything(t *testing.T) {
	// data.go reads the clock into the snapshot; templates.go is the markup that
	// renders it. Everywhere else, a mention would be a branch.
	rendering := map[string]bool{"data.go": true, "templates.go": true}
	for name, src := range packageFiles(t) {
		if rendering[name] {
			continue
		}
		code := strings.Join(nonCommentLines(src), "\n")
		for _, banned := range []string{"KRSessionAdvisory(", ".Outside"} {
			if strings.Contains(code, banned) {
				t.Errorf("%s reads the market-hours advisory in Go (%q). It is rendered by the template and "+
					"consulted by nothing: the window in which the broker accepts an order is unmeasured, so a "+
					"refusal here would assert something nobody has observed", name, banned)
			}
		}
	}
}

// TestTheAddressIsSpelledOnceAndItIsLoopback.
func TestTheAddressIsSpelledOnceAndItIsLoopback(t *testing.T) {
	for name, src := range packageFiles(t) {
		code := strings.Join(nonCommentLines(src), "\n")
		for _, banned := range []string{`"0.0.0.0`, `"[::]`, `net.Listen("tcp", addr`, `net.Listen("tcp", host`} {
			if strings.Contains(code, banned) {
				t.Errorf("%s contains %q; this console binds 127.0.0.1 and nothing else", name, banned)
			}
		}
	}
	code := strings.Join(nonCommentLines(packageFiles(t)["console.go"]), "\n")
	if !strings.Contains(code, `"127.0.0.1:%d"`) {
		t.Error("console.go no longer spells the loopback address it binds")
	}
	if !strings.Contains(code, "IsLoopback()") {
		t.Error("Serve no longer checks that the listener it was handed is loopback")
	}
}

// TestNothingCanPresetTheSessionOrCSRFToken.
//
// A settable token is a non-interactive approval path with extra steps: whoever
// could set it could open the console, submit the form and answer the nonce it
// prints. Both are minted in New from crypto/rand and are unexported.
func TestNothingCanPresetTheSessionOrCSRFToken(t *testing.T) {
	for name, src := range packageFiles(t) {
		code := strings.Join(nonCommentLines(src), "\n")
		if strings.Contains(code, "os.Getenv") || strings.Contains(code, "os.LookupEnv") {
			t.Errorf("%s reads the environment; nothing about this console's authentication may come from it", name)
		}
	}
	// Options is the whole configuration surface. If it grew a token field, this
	// catches it.
	file := parseFile(t, "console.go", packageFiles(t)["console.go"])
	ast.Inspect(file, func(n ast.Node) bool {
		spec, ok := n.(*ast.TypeSpec)
		if !ok || spec.Name.Name != "Options" {
			return true
		}
		st, ok := spec.Type.(*ast.StructType)
		if !ok {
			return true
		}
		for _, field := range st.Fields.List {
			for _, ident := range field.Names {
				lowered := strings.ToLower(ident.Name)
				for _, banned := range []string{"token", "session", "csrf", "nonce", "approve", "host", "bind", "address", "interface"} {
					if strings.Contains(lowered, banned) {
						t.Errorf("console.Options has a %s field; the console's authentication and its "+
							"interface are not configurable", ident.Name)
					}
				}
			}
		}
		return false
	})
}

// TestTheConsoleHoldsNoBrokerOfItsOwn.
//
// StartVerify is the console's only route to an account, and it is supplied by
// cmd/tossctl. An import of the Open API client here would be a second route —
// one with no plan authorisation, no exposure cap and no evidence line in front
// of it.
func TestTheConsoleHoldsNoBrokerOfItsOwn(t *testing.T) {
	banned := []string{
		"internal/official", // the Open API client
		"internal/client",   // the web-session client
		"internal/hybrid",   // routes writes to whichever backend answers
		"internal/execgw",
		"internal/flatten",
		"internal/app/engine",
		"internal/trading",
		"internal/orderintent",
	}
	for name, src := range packageFiles(t) {
		for _, b := range banned {
			if strings.Contains(src, `"`+modulePath+b+`"`) {
				t.Errorf("%s imports %s; the console reaches an account only through the StartVerify it is given", name, b)
			}
		}
	}
}

// TestTheConsoleStartsNoProcessOfItsOwn.
//
// Task 1.8 puts two restarts on the dashboard, and neither of them is implemented
// here: Relaunch and RestartSoak are functions cmd/tossctl supplies, and this
// package only decides whether the person asking has cleared both gates. That is
// what lets the whole restart surface be tested with httptest and no fork — and it
// is what keeps "the console runs nothing" true after the buttons exist.
func TestTheConsoleStartsNoProcessOfItsOwn(t *testing.T) {
	for name, src := range packageFiles(t) {
		if strings.Contains(src, `"os/exec"`) || strings.Contains(src, `"syscall"`) {
			t.Errorf("%s imports a process-spawning package; Relaunch and RestartSoak are seams the caller "+
				"fills, so that nothing here forks and no test has to", name)
		}
		code := strings.Join(nonCommentLines(src), "\n")
		for _, banned := range []string{"os.StartProcess", "exec.Command", "syscall.Exec", "os.Process"} {
			if strings.Contains(code, banned) {
				t.Errorf("%s contains %q; the console asks for a restart, it does not perform one", name, banned)
			}
		}
	}
}

// TestTheConsoleWritesNothingButTheEvidenceItsRunnerWrites.
//
// Everything this package reads is a local file, and it reads them to render a
// page. A write from here would be a state change nobody approved — and the two
// legitimate writers are both the caller's: verifylive's recorder, and the handoff
// store behind console.Handoff (task 1.8 ①), which owns the 0600 file, the
// single-use rule and the window so that this package owns none of them.
func TestTheConsoleWritesNothingButTheEvidenceItsRunnerWrites(t *testing.T) {
	for name, src := range packageFiles(t) {
		code := strings.Join(nonCommentLines(src), "\n")
		for _, banned := range []string{"os.WriteFile", "os.Create", "os.OpenFile", "os.Remove", "os.MkdirAll", "OpenRecorder("} {
			if strings.Contains(code, banned) {
				t.Errorf("%s contains %q; the console renders local files, it does not write them", name, banned)
			}
		}
	}
}

// nonCommentLines drops whole-line comments. It does not parse Go: a line wrongly
// kept makes the guard stricter, which is the safe direction here.
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
