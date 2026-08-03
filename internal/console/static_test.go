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
//	borrowed judge  the approval window is verifylive.Batch's answer, and no typed
//	                string is read back here. A second judgement in this package
//	                could drift into accepting what the terminal would refuse.
//	loopback        the address is spelled once, and it is 127.0.0.1.
//	no bypass       nothing reads an environment variable, and nothing can preset
//	                the session token.
//	no broker       this package cannot reach an account except through the
//	                StartVerify it is handed.

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"sort"
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
	// ReadOnly reports the `readOnly` wrapper on the handler chain: GET and HEAD
	// only, 405 for everything else (console.go).
	//
	// It exists because the record used to carry no method at all, and a guard
	// asked whether a route is a read could therefore only look at CSRFGated —
	// which turns "it is a GET" into "it is not protected", and grants an
	// exception on the strength of the route being unprotected. With that
	// reasoning a POST to the exempted path is served on a session cookie alone.
	ReadOnly bool
	// File is where the registration was found, so a failure names the file the
	// reviewer has to open.
	File string
}

// registrarNames are the two mux methods that put a handler on the table.
var registrarNames = map[string]bool{"HandleFunc": true, "Handle": true}

// registeredRoutes reads Console.routes out of the source — out of every file in
// the package, not out of one.
//
// It parsed console.go alone until change console-operator-overview. A route
// registered anywhere else was invisible to all four route guards at once: no
// account verb was checked against its path, no CSRF gate was demanded of it, and
// the read-route list never noticed it existed. That is worse than an absent
// guard, because the guards were still green — the failure this prevents is a new
// screen carrying an act into the table with every test still passing.
//
// Handle is recognised alongside HandleFunc, and a registration is no longer
// skipped for carrying an unexpected number of arguments (task 1.2): both were
// ways for a route to leave the table without anything saying so.
//
// # Two ways out of the table that are refused rather than followed
//
// The scan reads a registration whose registrar is spelled at the call site. It
// therefore has to refuse the two shapes that hide the call site from it, because
// following them would mean writing a second, partial Go evaluator inside a test:
//
//	the registrar as a value    `register := mux.HandleFunc` and then
//	                            `register("POST /verify/order/cancel", h)`. The
//	                            call's Fun is an *ast.Ident, so the scan skipped
//	                            it in silence and every one of the five route
//	                            guards passed: no session gate demanded, no CSRF
//	                            gate demanded, a method pattern unnoticed and two
//	                            account verbs unread. At runtime that route
//	                            answered an UNAUTHENTICATED POST with 200 while
//	                            /dashboard answered 403.
//	a mounted subtree           `mux.Handle("/x/", subRouter)` hands every path
//	                            beneath /x/ to a table registered somewhere this
//	                            scan cannot see. The registration itself looks
//	                            fine; what it serves is invisible.
//
// Both fail loudly. A guard that follows values is a guard whose reader has to
// believe it followed them all.
func registeredRoutes(t *testing.T) []route {
	t.Helper()
	files := packageFiles(t)
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)

	var routes []route
	for _, name := range names {
		file := parseFile(t, name, files[name])
		// Selectors that appear as a call's Fun. Anything naming a registrar and
		// NOT in here is the registrar taken as a value. ast.Inspect is
		// pre-order, so a CallExpr is always recorded before its own Fun is
		// visited.
		called := map[ast.Expr]bool{}
		ast.Inspect(file, func(n ast.Node) bool {
			if sel, ok := n.(*ast.SelectorExpr); ok && registrarNames[sel.Sel.Name] && !called[sel] {
				t.Errorf("%s takes %s as a value rather than calling it. The route table is read out "+
					"of the call site, so a registration made through a variable is invisible to every "+
					"guard in this file at once — session gate, CSRF gate, method pattern and account "+
					"verb — and the route still answers requests", name, sel.Sel.Name)
			}
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || !registrarNames[sel.Sel.Name] {
				return true
			}
			called[sel] = true
			if len(call.Args) == 0 {
				t.Errorf("%s registers a route with no path at all", name)
				return true
			}
			lit, ok := call.Args[0].(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				t.Errorf("%s registers a route with a non-literal path: %v", name, call.Args[0])
				return true
			}
			r := route{Path: routePathLiteral(lit.Value), File: name}
			if strings.HasSuffix(r.Path, "/") && r.Path != "/" {
				t.Errorf("%s registers the subtree pattern %q. Everything beneath it is served by a "+
					"handler this scan cannot look inside, so the routes that actually answer are "+
					"registered somewhere no guard in this file reads", name, r.Path)
			}
			for _, arg := range call.Args[1:] {
				if opaqueHandler(arg) {
					t.Errorf("%s registers %q with a handler this scan cannot see through (%T). A route "+
						"whose handler is a value is a route whose gates cannot be read off the "+
						"registration", name, r.Path, arg)
				}
				if outer, ok := arg.(*ast.CallExpr); ok {
					if fn, ok := outer.Fun.(*ast.SelectorExpr); ok && fn.Sel.Name == "session0" {
						r.Session = true
					}
				}
				ast.Inspect(arg, func(inner ast.Node) bool {
					if c, ok := inner.(*ast.CallExpr); ok {
						if fn, ok := c.Fun.(*ast.SelectorExpr); ok {
							switch fn.Sel.Name {
							case "mutating":
								r.CSRFGated = true
							case "readOnly":
								r.ReadOnly = true
							}
						}
					}
					return true
				})
			}
			routes = append(routes, r)
			return true
		})
	}
	if len(routes) == 0 {
		t.Fatal("no routes were found; the guard is not reading the routing table")
	}
	return routes
}

// routePathLiteral is the path inside a Go string literal, in either quotation
// form.
//
// It trimmed only `"`. A raw string literal — `mux.HandleFunc(`+"`"+`/x`+"`"+`, …)` —
// therefore carried its backticks into the route table, and the guard that
// noticed was TestNoRouteIsRegisteredWithAMethodPattern, which reported a method
// pattern. Neither the defect nor the file it named was the real one, and a
// failure message that sends a reader to the wrong place is worse than a silent
// one: they go and look, find nothing, and learn to distrust the guard.
func routePathLiteral(value string) string {
	return strings.Trim(value, "\"`")
}

// TestTheRoutePathIsReadOutOfEitherQuotationForm.
func TestTheRoutePathIsReadOutOfEitherQuotationForm(t *testing.T) {
	for _, tc := range []struct{ literal, want string }{
		{`"/dashboard"`, "/dashboard"},
		{"`/dashboard`", "/dashboard"},
		{`"GET /dashboard"`, "GET /dashboard"}, // still caught as a method pattern
	} {
		if got := routePathLiteral(tc.literal); got != tc.want {
			t.Errorf("routePathLiteral(%s) = %q, want %q; a path that keeps its quotes matches nothing "+
				"in this file and fails under whichever guard compares strings first", tc.literal,
				got, tc.want)
		}
	}
}

// opaqueHandler reports a handler argument whose gates cannot be read off the
// registration.
//
// What the scan can see through is a chain of gate wrappers ending in a method
// value on the console (c.session0(c.mutating(c.handleX))) or a literal function.
// An identifier is not one of those: it is a value assigned elsewhere, and a
// *http.ServeMux is a legal value for it.
func opaqueHandler(expr ast.Expr) bool {
	switch v := expr.(type) {
	case *ast.CallExpr:
		for _, arg := range v.Args {
			if _, ok := arg.(*ast.BasicLit); ok {
				continue // a wrapper's own literal argument is not the handler
			}
			if opaqueHandler(arg) {
				return true
			}
		}
		return false
	case *ast.SelectorExpr, *ast.FuncLit:
		return false
	case *ast.ParenExpr:
		return opaqueHandler(v.X)
	default:
		return true
	}
}

// TestEveryRouteGoesThroughTheSessionGate.
//
// The session token is the console's whole authentication: possession of the
// terminal that printed it. A handler registered without session0 in front of it
// would be reachable by anything on this machine that can open a socket, which on
// a developer's laptop includes every browser tab and every agent.
func TestEveryRouteGoesThroughTheSessionGate(t *testing.T) {
	routes := registeredRoutes(t)
	public := map[string]bool{
		"/healthz": true, // fixed response only; container probes need no credential
		"/login":   true, // the authentication endpoint itself
	}
	for _, r := range routes {
		if !r.Session && !public[r.Path] {
			t.Errorf("%s is registered without session0; it is reachable without the session token", r.Path)
		}
	}
	// The floor is the canary for "the extractor stopped parsing", so it follows
	// the real number rather than sitting at some historical low. The twenty,
	// enumerated so the number and the list cannot drift apart the way they did
	// once already (the list added to sixteen while the assertion said
	// seventeen — the settings SCREEN was missing from it, and a comment that
	// does not add up is a comment the next reader stops checking against):
	//
	//	7  the verification console: /, /verify, /verify/start, /verify/approve,
	//	   /verify/abort, /report, /report.json
	//	2  the dashboard screens (add-operator-dashboard): /positions, /history
	//	1  the settings screen (console-adoption-controls): /settings
	//	3  its three adoption edits: /settings/save, /settings/include, and
	//	   /settings/exclude (console-excludes-in-one-click)
	//	2  its two Guardian-limit edits: /settings/limits and
	//	   /settings/limits/preset (console-sets-guardian-limits)
	//	2  the engine's process control (add-engine-runtime): /engine/start,
	//	   /engine/stop
	//	2  the restarts: /restart, /soak/restart
	//	1  the overview (console-operator-overview): /dashboard
	//	1  the orders screen (console-orders-screen): /orders
	//	1  the discovery screen (add-candidate-discovery): /signals
	//
	// A floor below the truth would let a scanner that read only console.go's
	// first half go on passing.
	// The authenticated remote surface adds /login, /logout and the fixed
	// credential-free /healthz probe.
	if len(routes) < 27 {
		t.Errorf("only %d route(s) were read; the guard is not seeing the whole table", len(routes))
	}
}

// TestNoRouteIsRegisteredWithAMethodPattern.
//
// Go 1.22's `HandleFunc("GET /dashboard", …)` is legal and it is the natural way
// to say what this screen is. It must not be used here, because the extractor
// above reads the literal as the path: the route table would then hold
// "GET /dashboard" and every path-keyed guard would go quietly wrong at once —
// the read-route list reports the screen unregistered, the account-verb scan
// searches a string that is no longer a path, and the CSRF pairing compares
// against names nothing will ever match.
//
// Constraining the method is worth having; it arrives with the change that
// teaches the extractor to split the pattern (console-orders-screen). Until
// then this fails loudly instead of leaving four guards looking at the wrong
// strings.
func TestNoRouteIsRegisteredWithAMethodPattern(t *testing.T) {
	for _, r := range registeredRoutes(t) {
		if strings.ContainsAny(r.Path, " \t") || !strings.HasPrefix(r.Path, "/") {
			t.Errorf("%s registers %q, which is a method pattern rather than a path; the route "+
				"table's extractor reads the literal as the path, so every path comparison in this "+
				"file silently stops matching", r.File, r.Path)
		}
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
		// The engine's process control (add-engine-runtime task 2.1). Same
		// reasoning one step up: neither touches an account — the engine's own
		// startup interlock decides whether the process it starts may trade — but
		// both are acts, and starting a trading engine from an embedded image is
		// not a thing a page should be able to do.
		"/engine/start": true,
		"/engine/stop":  true,
		// The adoption-settings edits (console-adoption-controls task 3.2,
		// console-excludes-in-one-click task 1.19). The only thing any of them
		// writes is the engine.adoption config block through the injected seam —
		// no journal, no broker, no account — but a config that outlives the
		// console is exactly what CSRF must gate.
		"/settings/save":    true,
		"/settings/include": true,
		"/settings/exclude": true,
		// The Guardian-limit edits (console-sets-guardian-limits). They write the
		// five ceilings and the currency inside engine.automation_gate and cannot
		// write the gate's own switch: the seam they save through takes a type with
		// no field for it. A page that could lower an exposure ceiling — or a
		// forged request that could — is still an act.
		"/settings/limits":        true,
		"/settings/limits/preset": true,
		// The operating toggles (console-owns-the-operating-toggles). The trading
		// policy decides whether the engine may submit anything at all, and the
		// gate decides whether it runs a loop. Both were `config.json` hand-edits
		// until this change, which is to say both were changes with no CSRF, no
		// validation and no audit line — the console path is the one that has all
		// three. A forged request that turned the gate on would be exactly the act
		// this list exists to keep behind the token.
		"/settings/trading":                     true,
		"/settings/gate":                        true,
		"/settings/autostart":                   true,
		"/settings/system-update/download":      true,
		"/settings/system-update/install":       true,
		"/optimization/exit-policy":             true,
		"/optimization/exit-protection/preview": true,
		"/optimization/exit-protection/apply":   true,
		"/position-management/preview":          true,
		"/position-management/apply":            true,
		"/openapi/login/save":                   true,
		"/logout":                               true,
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

// TestTheApprovalWindowIsJudgedByVerifylive.
//
// The console approves with a click (operator-console: 검증 배치 승인의 형식), so the
// one thing left to judge is the window — and it is judged by the batch itself. A
// clock rule written here would be a second definition of "too late" that drifts
// from the terminal's the first time either changed.
//
// The nonce assertions are the other half: nothing in this package may read one
// back or compare one, because a typed string is not what approves anything here.
func TestTheApprovalWindowIsJudgedByVerifylive(t *testing.T) {
	src := packageFiles(t)["pages.go"]
	code := strings.Join(nonCommentLines(src), "\n")
	if !strings.Contains(code, "view.Batch.Expired(") {
		t.Error("pages.go no longer asks verifylive.Batch whether the approval window has closed")
	}
	for _, banned := range []string{
		"== view.Batch.Nonce", "Nonce ==", "== nonce", `PostFormValue("nonce")`, "ExpiresAt",
	} {
		if strings.Contains(code, banned) {
			t.Errorf("pages.go judges the approval itself (%q); verifylive.Batch is the only judge", banned)
		}
	}
	// And nothing in the package invents an approval.
	for name, fileSrc := range packageFiles(t) {
		body := strings.Join(nonCommentLines(fileSrc), "\n")
		for _, banned := range []string{"AUTO_APPROVE", "SKIP_CONFIRM", "NO_CONFIRM", "TOSSCTL_CONSOLE_TOKEN"} {
			if strings.Contains(body, banned) {
				t.Errorf("%s contains %q; the approval must be a person's own act", name, banned)
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
	if !strings.Contains(code, "c.redoSet(market)") {
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
	//
	// The dashboard files (holdings.go, portfolio.go, portfolio_pages.go,
	// templates_portfolio.go) are deliberately NOT on this list. The positions
	// screen has an honest freshness signal of its own — the broker cache's
	// timestamp — and adding the market-hours advisory to it would be one more
	// place the "advisory only" rule has to keep being true. If a later change
	// renders it there, that file joins this map and the reviewer is the one who
	// decides it is a render and not a branch.
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

// --- the dashboard (change add-operator-dashboard, task 2.3) -----------------------

// consoleStateChanging is the complete list of routes that are allowed to change
// anything, transcribed from the operator-console spec: 콘솔의 상태변경 행위는
// 검증 실행 제어(시작·승인·중단), 프로세스 기동·정지(자기 재시작·soak 재시작·
// 엔진 시작/정지), 편입 설정 편집(편입 설정 저장·종목 편입 지정·종목 제외 지정),
// 그리고 Guardian 한도 편집(한도 프리셋 적용·개별 한도 저장) 뿐이다(SHALL — 계좌
// 무접촉; 편입 설정 편집의 대상은 engine.adoption config 블록만이고, 한도 편집의
// 대상은 engine.automation_gate의 다섯 한도와 한도 통화뿐이다 — enabled와 kill
// switch는 콘솔 밖이다).
//
// It is the same set TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate uses,
// named separately here because the two tests ask different questions of it: one
// asks whether these routes are gated, the other whether anything outside them
// even looks like an act.
var consoleStateChanging = []string{
	"/verify/start", "/verify/approve", "/verify/abort", "/restart", "/soak/restart",
	"/engine/start", "/engine/stop", "/settings/save", "/settings/include",
	"/settings/exclude", "/settings/limits", "/settings/limits/preset",
	"/settings/trading", "/settings/gate", "/settings/autostart",
	"/settings/system-update/download",
	"/settings/system-update/install",
	"/openapi/login/save",
}

// consoleGateWriters is the exact route that may spell "gate" (change
// console-owns-the-operating-toggles).
//
// The ban on that word was written when the console could not touch the switch,
// and it did real work: it caught a /settings/limits that quietly grew gate
// vocabulary. Now one route legitimately writes `engine.automation_gate.enabled`,
// and the spec's answer is not to delete the word from the list — it is to name
// the one path, and to require the name rather than forbid it:
//
//	"게이트 스위치 라우트는 존재하며 상태변경 목록에 열거된다(SHALL — 이름을
//	 숨기지 않는다: 무엇을 하는 라우트인지 이름이 말해야 감사가 읽힌다)."
//
// A route that turned the engine loose under a name like /settings/operating
// would pass the verb ban and defeat the audit at the same time. So the
// exception is byte-for-byte on the whole path, with none of the softening
// consoleAccountReads spells out above — a prefix match would readmit
// /settings/gate/force, ToLower would readmit /settings/Gate as a second route,
// and a trailing-slash trim would make it a subtree pattern.
var consoleGateWriters = map[string]bool{"/settings/gate": true}

// --- the verbs, spelled once and shared by both surfaces --------------------------
//
// This file guards two surfaces — the route table and the injected capabilities —
// and each used to keep its own private verb list. They drifted, and the drift was
// not visible from either side: "flatten" was on the route list and missing from
// the capability list, so a seam whose only method was Flatten — liquidating the
// whole account — went through the capability walk under its own name while a
// route called /flatten would have failed on sight.
//
// The shared set is now spelled once and each surface adds what only it can mean.

// sharedAccountVerbs name a request against the account. Neither a path nor a
// method may be spelled with one.
var sharedAccountVerbs = []string{
	"order", "sell", "buy", "cancel", "modify", "amend", "flatten",
}

// routeOnlyAccountVerbs name a way to reach the account that only a URL can be
// accused of.
//
// They are deliberately NOT checked on methods and types, because this package
// legitimately declares GateLimitsReader (a read of a configured ceiling),
// AdoptionSettings (the operator-console spec requires that seam by name) and
// Handoff (the console's own single-use session token). None of the three is a
// request against an account; a *path* carrying any of those words would be.
var routeOnlyAccountVerbs = []string{
	"gate", "credential", "secret", "token", "adopt", "enroll",
}

// methodOnlyMutationVerbs are the spellings a method or a type reaches for that a
// path would not: a URL says /orders, a method says PlaceOrder.
var methodOnlyMutationVerbs = []string{
	"place", "create", "delete", "update", "submit", "transfer", "withdraw", "conditional",
}

// accountVerbs is what the route table is held to.
var accountVerbs = append(append([]string{}, sharedAccountVerbs...), routeOnlyAccountVerbs...)

// consoleAccountReads are the exact route paths that READ the account's order
// record rather than acting on it (change console-orders-screen, task 3.3).
//
// The account-verb ban is a loose string test and looseness is what makes it
// useful: /orders/cancel and /order-place are caught by the same line. Reading a
// record is not what that ban aims at, so this is the hole — and it is a set of
// whole paths compared byte for byte.
//
// Every softer comparison has a name and a specific failure:
//
//	prefix match      /orders/cancel is a path beginning with /orders.
//	strings.ToLower   /Orders is a second route, because Go's mux matches the
//	                  path case-sensitively.
//	TrimSuffix("/")   /orders/ is worse than a duplicate: in Go 1.22+ a trailing
//	                  slash makes it a SUBTREE pattern, so /orders/cancel is
//	                  routed to that handler.
//
// orders_static_test.go registers all three shapes and requires the guard to
// fail on each, which is the only form of that assertion that measures anything:
// "the allowlist does not contain /orders/cancel" stays true under all three.
var consoleAccountReads = map[string]bool{"/orders": true}

// routeReadsTheAccountRecord reports the account-verb exception applying to one
// route.
//
// Three facts, all required. The path is compared byte for byte against
// consoleAccountReads. ReadOnly is the `readOnly` wrapper actually being on the
// chain, which is what makes "this is a read" a fact the table carries rather
// than an inference from the absence of a gate. And a route inside the CSRF gate
// is a state change by this file's own definition, whatever it is called.
func routeReadsTheAccountRecord(r route) bool {
	return consoleAccountReads[r.Path] && r.ReadOnly && !r.CSRFGated
}

// routeFindings is the account-mutation judgement for ONE route, as a pure
// function of the record.
//
// It is extracted from the test rather than written inside it because the
// exception above has to be measurable, and registeredRoutes parses this
// package's source off disk: a test cannot register a fake /orders/cancel to
// watch the guard refuse it. Without the extraction the only assertion available
// is "that path is not in the allowlist", which measures nothing — it holds
// under a prefix match, under ToLower and under a trailing-slash trim, which are
// exactly the three ways the exception grows.
//
// Two claims, in the two loops the table has always had:
//
//	no account verbs   nothing anywhere in the table is spelled like placing,
//	                   cancelling or amending an order, opening a gate, or
//	                   reaching a credential — including the routes in
//	                   consoleStateChanging, which ARE allowed to change
//	                   something, because none of them touches the account. (The
//	                   count is deliberately not written here: it was "five" while
//	                   the list held nine.)
//	nothing else acts  every other route is a reading. A path outside the list
//	                   above that reads like an act is either a mistake or a
//	                   requirement change, and both should stop here.
//
// The exception is consulted in BOTH. The second loop's verbs include the
// account verbs, so /orders trips it too — and the wrong repair is to add
// /orders to consoleStateChanging to silence the second one, because that list
// is what TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate demands a CSRF
// gate for, and a read route behind the CSRF gate is a page nobody can open.
func routeFindings(r route) []string {
	allowed := map[string]bool{}
	for _, path := range consoleStateChanging {
		allowed[path] = true
	}
	// Verbs that name an act rather than a reading. The allowed routes are
	// exempt: the verification control surface, the two restarts, the engine's
	// two process-control routes, and the two adoption-settings edits — which is
	// consoleStateChanging in full. The config-write vocabulary is here so a future
	// unlisted /settings/anything cannot sail past this guard the way an
	// unrecognized act otherwise would (console-adoption-controls, review P2-7).
	// "exclude" earns its place the hard way: it does not contain "include", so
	// before it was listed an unargued /settings/exclude would have sailed past
	// this guard entirely (console-excludes-in-one-click).
	// "limit" and "preset" earn their places the same hard way "exclude" did: an
	// unargued /settings/limits contains none of the verbs above, so before they
	// were listed it would have sailed past this guard entirely. ("preset"
	// happens to contain "reset" and would have tripped, which is luck, not a
	// guard — /settings/limits alone would not have.)
	actVerbs := append([]string{"start", "stop", "approve", "abort", "restart", "reset", "delete",
		"save", "include", "exclude", "enable", "config", "limit", "preset"},
		accountVerbs...)

	var findings []string
	reads := routeReadsTheAccountRecord(r)
	lowered := strings.ToLower(r.Path)
	for _, verb := range accountVerbs {
		if !strings.Contains(lowered, verb) || reads {
			continue
		}
		// The one route the spec requires to carry "gate" in its name, and only
		// that word: /settings/gate must still not be allowed to grow "credential"
		// or "sell".
		if verb == "gate" && consoleGateWriters[r.Path] {
			continue
		}
		findings = append(findings, fmt.Sprintf(
			"route %s names %q; this console has no route that touches the account, its gate "+
				"or its credentials", r.Path, verb))
	}
	if allowed[r.Path] || reads {
		return findings
	}
	for _, verb := range actVerbs {
		if strings.Contains(lowered, verb) {
			findings = append(findings, fmt.Sprintf(
				"route %s names %q but is not in consoleStateChanging; a route that acts has to be "+
					"argued for and CSRF-gated", r.Path, verb))
		}
	}
	return findings
}

// routeTableFindings is routeFindings over a whole table, plus the one claim that
// is about the table rather than about a route: every path the state-changing
// list names is actually registered. A list naming a route nobody registers has
// stopped describing this console.
func routeTableFindings(routes []route) []string {
	var findings []string
	seen := map[string]bool{}
	for _, r := range routes {
		seen[r.Path] = true
		findings = append(findings, routeFindings(r)...)
	}
	for _, path := range consoleStateChanging {
		if !seen[path] {
			findings = append(findings, fmt.Sprintf(
				"%s is in the state-changing list but is not registered", path))
		}
	}
	return findings
}

// TestNoRouteNamesAnAccountMutation applies the judgement to the real table.
func TestNoRouteNamesAnAccountMutation(t *testing.T) {
	for _, finding := range routeTableFindings(registeredRoutes(t)) {
		t.Error(finding)
	}
}

// TestTheDashboardScreensAreReads.
//
// The positions, history, overview, orders and signals routes exist, go through
// the session gate like everything else, and are NOT behind the CSRF gate — a
// read route that demanded a POST would be a page nobody can open, which is the
// failure this catches from the other side.
func TestTheDashboardScreensAreReads(t *testing.T) {
	want := map[string]bool{
		"/positions": false, "/history": false, "/dashboard": false, "/orders": false,
		"/signals": false,
	}
	found := map[string]bool{}
	for _, r := range registeredRoutes(t) {
		gated, ok := want[r.Path]
		if !ok {
			continue
		}
		found[r.Path] = true
		if r.CSRFGated != gated {
			t.Errorf("%s: CSRF-gated = %v, want %v", r.Path, r.CSRFGated, gated)
		}
		if !r.Session {
			t.Errorf("%s is registered without session0", r.Path)
		}
	}
	for path := range want {
		if !found[path] {
			t.Errorf("%s is not registered; the dashboard screen has no route", path)
		}
	}
}

// --- the injected capabilities (change console-operator-overview, task 1.4) ---------

// mutationVerbs name a request against an account, in any spelling a method or a
// type is likely to use. See sharedAccountVerbs for why the two lists differ.
var mutationVerbs = append(append([]string{}, sharedAccountVerbs...), methodOnlyMutationVerbs...)

// consoleCapabilities enumerates every field console.Options declares and what
// that field is permitted to be. A nil entry means "carries no method set": plain
// data, or a func-type seam whose name and signature are checked instead.
//
// The unit of the check is the FIELD rather than the interface, because checking
// interfaces left three holes and all three opened the moment this change added a
// file and a seam:
//
//	one file        the previous guard read packageFiles(t)["holdings.go"]. A wide
//	                interface declared in any other file failed nothing at all.
//	no func types   five of the seven injected seams are func types — StartVerify,
//	                StartEngine, StopEngine, Relaunch, RestartSoak — and a scan for
//	                *ast.InterfaceType sees none of them. `type PlaceOrderFunc
//	                func(context.Context, domain.Position) error` would have been
//	                injected straight past it: no interface to find, and no banned
//	                import either, because cmd/tossctl is what fills it in.
//	one allowlist   a single allowlist cannot cover every interface. Handoff
//	                declares Mint and Consume, AdoptionSettings declares Load and
//	                Save, and the operator-console spec requires the second of
//	                those in writing. Widening the one allowlist to fit them would
//	                have let `interface{ Holdings(...); PlaceOrder(...) }` through.
//
// Patching those one at a time moves the hardcoding from a file name to a type
// name and no further. So each field names its own permitted method set, and a
// field that is absent from this map fails on that alone — the invariant is not
// "the broker seam is narrow", it is "every capability this console is handed is
// enumerated here".
var consoleCapabilities = map[string]capability{
	// Plain data: paths, numbers and lists. Nothing to reach an account with.
	"Port":              {},
	"Remote":            {},
	"SoakRecord":        {},
	"VerifyRecord":      {},
	"VerifyRecordUS":    {},
	"Attestation":       {},
	"MinSoakDays":       {},
	"RequiredEndpoints": {},
	"JournalPath":       {},
	"RunLockPath":       {},
	"EngineMarker":      {},
	"EngineBootNote":    {},

	// Func-type seams. A func has no method set, so what is checked is its field
	// name and every type its signature mentions.
	"StartVerify":  {},
	"Relaunch":     {},
	"RestartSoak":  {},
	"CheckOpenAPI": {},
	"SaveOpenAPI":  {},
	"StartEngine":  {},
	"StopEngine":   {},
	"Now":          {},
	"Binary":       {},
	"AcquireUpdateEngineLock": {
		VerbExemptions: map[string]string{
			"AcquireUpdateEngineLock": "holds the real engine flock while replacing the local executable; it carries no account capability",
		},
	},
	"CheckUpdateVerifyActivity": {
		VerbExemptions: map[string]string{
			"CheckUpdateVerifyActivity": "reads whether external verification is active and refuses local executable replacement; it mutates no account state",
		},
	},
	// Out is io.Writer: the console's own operator lines, not an account.
	"Out": {},

	// Interface seams: exactly these methods each, and no embedding.
	"Handoff":         {Methods: []string{"Mint", "Consume"}},
	"Holdings":        {Methods: []string{"Holdings"}},
	"InstrumentNames": {Methods: []string{"Names"}},
	"Settings":        {Methods: []string{"Load", "Save"}},
	"EngineBoot":      {Methods: []string{"Load", "Save"}},
	// The scheduler screen receives display data only. Read has no parameter
	// carrying a symbol, time, holiday override, or operating command.
	"MarketSchedule": {Methods: []string{"Read"}},
	// Derived lane-performance data only. Dashboard takes a fixed typed query and
	// returns plain values; no Store, collector, journal, broker, config, or
	// operating-control handle crosses this seam.
	"Performance": {Methods: []string{"Dashboard"}},
	// Dormant strategy runtime display data only. Read accepts no operating
	// command and cannot mint lane activation or account authority.
	"StrategyRuntime": {Methods: []string{"Read"}},
	// The common exit-policy editor carries only the typed policy ID block. It
	// cannot reach broker, journal, automation gate, or trading toggles.
	"ExitPolicies": {Methods: []string{"Load", "Save"}},
	// The a050 lifecycle is a separate control-plane capability. It can create
	// immutable settings candidates and CAS versions, but its messages and
	// dependency closure expose no journal, broker, order, lane, gate, kill
	// switch or LIVE action.
	"Optimization": {
		Methods: []string{"Read", "Preview", "PreviewRollback", "Apply", "RecoverConflict"},
		VerbExemptions: map[string]string{
			"Optimization":          "the settings lifecycle capability; it cannot reach account or operating authority",
			"OptimizationCommander": "the exact settings lifecycle and read-only conflict recovery seam",
			"Apply":                 "commits one capability-bound settings CAS and append-only audit; no account mutation is in its type closure",
		},
	},
	// Generation-scoped policy writes are a deliberately narrow engine-owned
	// command capability. Its messages contain no broker, SQL, config, toggle or
	// arbitrary user text; Runtime is an additive read-only startup/tracker
	// projection and the console can only submit an opaque server action.
	"PositionPolicies": {
		Methods: []string{"List", "Runtime", "Preview", "Apply"},
		VerbExemptions: map[string]string{
			"PositionPolicies":        "the engine-owned position policy command capability; it cannot place orders or flip operating toggles",
			"PositionPolicyCommander": "the narrow lifecycle command plus read-only management runtime seam",
			"Apply":                   "commits one already-previewed position policy lifecycle CAS; it has no broker or config capability",
		},
	},
	// Broker protection is exposed only as an engine-owned current-row command
	// capability. The browser submits opaque selection/preview capabilities and
	// one checkbox; it never sends a symbol, trigger, quantity, reason, toggle,
	// journal handle, or broker client.
	"Protections": {
		Methods: []string{"List", "Preview", "Apply"},
		VerbExemptions: map[string]string{
			"Protections":         "the engine-owned protection status/current-row command capability; no broker or operating toggle is exposed",
			"ProtectionCommander": "the exact three-method opaque capability seam",
			"Apply":               "consumes one server-previewed current-row capability after the engine-enforced delay",
		},
	},
	// The Guardian limits are a read, and they are a seam of their own rather
	// than a third method on Settings: that one writes the adoption block, and a
	// screen that only wants to display a ceiling must not gain the ability to
	// edit configuration on the way (console-operator-overview D8).
	"GateLimits": {Methods: []string{"GateLimits"}},
	// The Guardian-limit editor (change console-sets-guardian-limits). A third
	// seam for the same reason GateLimits is a second one: the overview displays
	// a ceiling and must not thereby be able to change it.
	//
	// Its Save takes config.GuardianLimits — five ceilings and a currency, with
	// no field for `enabled` and none for the attestation path. That is what
	// keeps "the console cannot open the automation gate" true after this change
	// opened the numbers: the switch has nowhere to travel. settings_limits_test.go
	// reads that off the interface with reflection so the property is measured
	// rather than described.
	"Limits": {Methods: []string{"Load", "Save"}},
	// The trading-policy editor (change console-owns-the-operating-toggles). Its
	// Save takes config.TradingPolicy — place, sell, cancel and the live master —
	// so amend, conditional and fractional have nowhere to travel, exactly as
	// `enabled` has nowhere to travel through Limits.
	"TradingPolicy": {Methods: []string{"Load", "Save"}},
	// The automation gate's switch, write-only (same change).
	//
	// A fifth seam rather than a second method on Limits, and the separation is
	// the safety property rather than tidiness: each save emits bytes for its own
	// keys only, so a ceiling edit cannot carry a stale `enabled` back into the
	// file and a switch flip cannot carry stale ceilings. There is no Load
	// because Limits.Load already returns the whole block including the switch,
	// and two readers of one key is how a screen disagrees with itself.
	"Gate": {
		Methods: []string{"Save"},
		VerbExemptions: map[string]string{
			"Gate": "the Options field and the seam type. The operator-console spec " +
				"requires this name rather than tolerating it — a switch that turns the " +
				"engine loose under a vaguer name would defeat the audit it is supposed " +
				"to leave",
			"GateSwitch": "the seam type. Switch is the whole claim: one boolean key, " +
				"engine.automation_gate.enabled, and the closed member list in " +
				"internal/config's operating_io.go is what enforces it",
		},
	},
	// The discovery store's assessment, read (change add-candidate-discovery,
	// task 5.5). One method, and it fetches nothing from an account: behind it is
	// internal/candidate, whose own dependency closure is {internal/clock}, so
	// there is no order verb anywhere in what this seam can reach.
	"Signals": {Methods: []string{"Signals"}},
	"SystemUpdater": {
		Methods: []string{"Inspect", "Install"},
		VerbExemptions: map[string]string{
			"SystemUpdater": "the fixed-sibling local executable updater; its exact two-method set is enumerated here",
			"Install":       "installs reviewed local executable bytes and has no broker or account capability",
		},
	},
	// Signed release retrieval and fixed candidate publication are split. The
	// downloader can select no path and cannot install; the stager receives
	// already verified bytes and can publish only the updater's fixed sibling.
	"ReleaseDownloader":      {Methods: []string{"Fetch"}},
	"ReleaseCandidateStager": {Methods: []string{"StageCandidate"}},
	// The order record, read (change console-orders-screen, task 3.7).
	//
	// One method, and it fetches. The verb list is not touched to make room for
	// it: deleting "order" from the list would let a future /order/place and a
	// future PlaceOrder through at the same time, which is the failure the list
	// is for. Instead the exemptions below are per field and per spelling, and
	// they are the argument for each name in writing.
	"Orders": {
		Methods: []string{"Orders"},
		VerbExemptions: map[string]string{
			"Orders": "the Options field, and the seam's one method. It lists the account's " +
				"orders; naming it anything else would describe something it does not do",
			"OrdersReader": "the seam type. Reader is the whole claim and the method set above " +
				"is what enforces it",
			"OrdersReading": "the value one call returns: two lists and each one's own outcome",
			"OrderRecord": "one plain order as the broker spelled it — every field a string, " +
				"because an absent value is not a zero",
			"ConditionalRecord": "one conditional order, same shape. \"conditional\" is a banned " +
				"method verb because CreateConditionalOrder is a mutation; this is the record " +
				"of one, and the screen exists precisely because a leftover conditional is " +
				"invisible everywhere else",
		},
	},
}

// capability is what one console.Options field is permitted to be.
type capability struct {
	// Methods is the exact method set an interface seam may declare. An empty
	// list is the claim that the field carries no method set at all.
	Methods []string
	// VerbExemptions are the identifiers reachable through this field that may
	// name an account verb, each mapped to the argument for it.
	//
	// The verb scan is a name filter over the capability closure, and the closure
	// walk's own documentation says what it is: a supplementary device above the
	// real check, which is the method set. A name filter with no escape hatch
	// forces the opposite of what it is for — a seam that lists the account's
	// ORDERS either gets a name that lies about what it reads, or the verb is
	// deleted from the shared list and every future PlaceOrder walks through the
	// hole.
	//
	// So the hatch is here, at the field, one spelling at a time, with the reason
	// written down. PlaceOrder on this same seam still fails, because it is not
	// in this map — and the method set above would fail it a second time.
	VerbExemptions map[string]string
}

// TestEveryCapabilityTheConsoleReceivesIsEnumeratedAndDeclaresNothingButReads.
//
// The spec's "광폭 브로커 인터페이스 주입 차단", "새 파일에 선언된 광폭 seam",
// "func 타입으로 주입된 mutation 능력" and "열거되지 않은 새 능력" scenarios, in one
// walk. verifylive.Broker has PlaceOrder, CancelOrder, ModifyOrder and three
// conditional-order mutations on it; handing that to a read-only screen would
// make "the console places no order" a fact about what the handlers happen to
// call rather than about what they are able to call.
//
// Declarations are read out of the source rather than through reflection because
// the claim is about the declaration: a method that exists and is never called at
// runtime is exactly what this has to catch.
func TestEveryCapabilityTheConsoleReceivesIsEnumeratedAndDeclaresNothingButReads(t *testing.T) {
	files := parsedPackage(t)
	declaredTypes := packageTypes(files)

	fields := optionsFields(t, files)
	if len(fields) == 0 {
		t.Fatal("no Options fields were read; the guard is not looking at the injection surface")
	}

	declared := map[string]bool{}
	for _, field := range fields {
		for _, name := range field.Names {
			declared[name.Name] = true
			allowed, enumerated := consoleCapabilities[name.Name]
			if !enumerated {
				// Reported before the verb check so that an unenumerated field
				// gets the message that matters: an unlisted capability fails on
				// being unlisted, whatever it is called.
				t.Errorf("console.Options declares %q, which is not in consoleCapabilities; a capability "+
					"the console receives without being enumerated is one nobody argued for", name.Name)
				checkVerbs(t, "the Options field "+name.Name, name.Name)
				continue
			}
			checkVerbsExcept(t, "the Options field "+name.Name, name.Name, allowed.VerbExemptions)
			checkCapability(t, name.Name, field.Type, allowed, declaredTypes)
		}
		if len(field.Names) == 0 {
			t.Error("console.Options embeds a struct; whatever that struct gains, Options gains silently")
		}
	}

	// The allowlist rots in the other direction too. A field removed from Options
	// while its entry stayed behind fails nothing, and the next reader takes the
	// entry for a description of the injection surface — which is the whole job
	// this map has. An entry with no field is a claim about a capability nobody
	// receives any more.
	stale := make([]string, 0, len(consoleCapabilities))
	for name := range consoleCapabilities {
		if !declared[name] {
			stale = append(stale, name)
		}
	}
	sort.Strings(stale)
	for _, name := range stale {
		t.Errorf("consoleCapabilities enumerates %q and console.Options no longer declares it; an "+
			"allowlist that outlives its field stops describing the injection surface", name)
	}

	// And nothing in the package names verifylive's wide broker as its own seam.
	// The enumeration above would catch it arriving through Options; this catches
	// it arriving as a local variable, a parameter or a struct field elsewhere.
	for name, fileSrc := range packageFiles(t) {
		code := strings.Join(nonCommentLines(fileSrc), "\n")
		if strings.Contains(code, "verifylive.Broker") {
			t.Errorf("%s takes a verifylive.Broker; that interface can place, amend and cancel orders, "+
				"and the dashboard is handed HoldingsReader instead", name)
		}
	}
}

// externalOptionTypes are the types from another package an Options field may be
// declared as.
//
// The guard reads this package's source, so it cannot read another package's
// method set. A qualified type in Options is therefore an unreadable method set,
// and each one has to be argued for here rather than waved through: io.Writer
// takes the console's own operator lines and can reach nothing.
var externalOptionTypes = map[string]bool{"io.Writer": true}

// goBuiltinTypes are the spellings that positively carry no method set.
var goBuiltinTypes = map[string]bool{
	"bool": true, "string": true, "error": true, "byte": true, "rune": true,
	"int": true, "int8": true, "int16": true, "int32": true, "int64": true,
	"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,
	"uintptr": true, "float32": true, "float64": true,
	"complex64": true, "complex128": true,
}

// checkCapability resolves one Options field's declared type and holds it to the
// allowlist entry the field was enumerated with.
//
// # Why it resolves to a fixed point rather than one hop
//
// The previous version followed a field's type exactly one step and then checked
// the SPELLING of every name it had passed. That is a name filter wearing a
// capability filter's documentation, and four separate shapes went straight
// through it:
//
//	a second interface   HoldingsReader.Holdings returning an extra AccountHandle
//	                     that declares PlaceOrder, CancelOrder and Flatten. The
//	                     guard verb-checked the name "AccountHandle", found nothing,
//	                     and never opened it. Renaming the identical type to
//	                     OrderHandle failed — which is the proof: what was being
//	                     checked was the word, not the method set.
//	a generic seam       `type Desk Seam[OrderPlacer]`. The walker had no
//	                     *ast.IndexExpr case, so it returned no names at all and the
//	                     verb check ran over an empty list.
//	an alias chain       `type Ticker = Wide`. One hop lands on an *ast.Ident, not
//	                     an *ast.InterfaceType, and the error for that was skipped
//	                     for any field enumerated nil.
//	an empty static type `Feed any`, type-asserted at the use site to
//	                     `interface{ PlaceOrder(...) }`. A field that can hold
//	                     anything is a field the enumeration cannot describe.
//
// So: every type reachable from the field is verb-checked, every interface
// reachable from it is opened, and the field's own resolved shape must be one the
// guard can positively read a method set off.
func checkCapability(t *testing.T, field string, expr ast.Expr, cap capability,
	declaredTypes map[string]ast.Expr) {
	t.Helper()

	subject := "Options." + field
	allowed := cap.Methods
	names, ifaces := capabilityClosure(expr, declaredTypes)
	for _, name := range names {
		checkVerbsExcept(t, "the type "+name+" reached through "+subject, name, cap.VerbExemptions)
	}
	for _, iface := range ifaces {
		checkNoEmbedding(t, "an interface reached through "+subject, iface)
	}

	resolved := resolveDeclared(expr, declaredTypes)
	seam, isInterface := resolved.(*ast.InterfaceType)
	if !isInterface {
		if len(allowed) > 0 {
			t.Errorf("Options.%s is enumerated with the methods %v, but its declared type resolves to "+
				"%T rather than to an interface this package declares; the guard cannot read what it "+
				"was handed", field, allowed, resolved)
		}
		if !methodless(resolved, declaredTypes, map[string]bool{}) {
			t.Errorf("Options.%s resolves to %T, which is neither an interface this package declares, "+
				"nor a func type, nor plain data. A nil entry in consoleCapabilities is the claim that "+
				"the field carries no method set, and the guard has to be able to see that rather than "+
				"assume it", field, resolved)
		}
		return
	}
	if len(seam.Methods.List) == 0 {
		t.Errorf("Options.%s is an empty interface. It declares no capability and accepts every one: "+
			"the use site type-asserts it back to whatever it likes, and the enumeration describes "+
			"nothing", field)
		return
	}

	declared := map[string]bool{}
	for _, method := range seam.Methods.List {
		if len(method.Names) == 0 {
			continue // reported by checkNoEmbedding
		}
		for _, name := range method.Names {
			declared[name.Name] = true
		}
	}

	want := map[string]bool{}
	for _, name := range allowed {
		want[name] = true
	}
	for name := range declared {
		if !want[name] {
			t.Errorf("the seam injected as Options.%s declares %q, which consoleCapabilities does not "+
				"allow it; the console is handed reads and nothing else", field, name)
		}
	}
	for name := range want {
		if !declared[name] {
			t.Errorf("the seam injected as Options.%s no longer declares %q; the seam's shape is part "+
				"of the spec", field, name)
		}
	}
}

// checkNoEmbedding fails on an interface that embeds another one: whatever the
// embedded interface gains, this one gains silently.
func checkNoEmbedding(t *testing.T, subject string, iface *ast.InterfaceType) {
	t.Helper()
	for _, method := range iface.Methods.List {
		if len(method.Names) == 0 {
			t.Errorf("%s embeds another interface; whatever that interface gains, this one gains "+
				"silently", subject)
		}
	}
}

// resolveDeclared follows a chain of package-declared type names to the
// declaration that is finally not another name.
//
// One hop was not enough: `type Ticker = Wide` and `type A B; type B C` both put
// the method set two names away, and the guard's whole claim is about the method
// set.
func resolveDeclared(expr ast.Expr, declaredTypes map[string]ast.Expr) ast.Expr {
	seen := map[string]bool{}
	for {
		ident, ok := expr.(*ast.Ident)
		if !ok {
			return expr
		}
		next, ok := declaredTypes[ident.Name]
		if !ok || seen[ident.Name] {
			return expr
		}
		seen[ident.Name] = true
		expr = next
	}
}

// capabilityClosure walks everything one declaration can reach — the type itself,
// the types in its signature, the types those declarations name, to a fixed point
// — and returns every identifier worth verb-checking and every interface found on
// the way.
//
// A struct's field NAMES are checked only when the field's type could carry a
// capability. GateLimits.MaxOrderNotional is a ceiling on orders and not an
// order, and failing on it would push the next person to rename the ceiling; a
// field spelled `PlaceOrder func(...)` still fails, because a func type can.
func capabilityClosure(expr ast.Expr, declaredTypes map[string]ast.Expr) ([]string, []*ast.InterfaceType) {
	var (
		names    []string
		ifaces   []*ast.InterfaceType
		seenName = map[string]bool{}
		seenNode = map[ast.Expr]bool{}
		queue    []ast.Expr
	)
	push := func(e ast.Expr) {
		if e == nil || seenNode[e] {
			return
		}
		seenNode[e] = true
		queue = append(queue, e)
	}
	var addName func(string)
	addName = func(n string) {
		if seenName[n] {
			return
		}
		seenName[n] = true
		names = append(names, n)
		if decl, ok := declaredTypes[n]; ok {
			push(decl)
		}
	}
	pushFieldTypes := func(fl *ast.FieldList) {
		if fl == nil {
			return
		}
		for _, f := range fl.List {
			push(f.Type)
		}
	}

	push(expr)
	for len(queue) > 0 {
		e := queue[0]
		queue = queue[1:]
		switch v := e.(type) {
		case *ast.Ident:
			addName(v.Name)
		case *ast.SelectorExpr:
			addName(v.Sel.Name)
		case *ast.StarExpr:
			push(v.X)
		case *ast.ParenExpr:
			push(v.X)
		case *ast.ArrayType:
			push(v.Elt)
		case *ast.Ellipsis:
			push(v.Elt)
		case *ast.MapType:
			push(v.Key)
			push(v.Value)
		case *ast.ChanType:
			push(v.Value)
		case *ast.IndexExpr: // Seam[OrderPlacer]
			push(v.X)
			push(v.Index)
		case *ast.IndexListExpr: // Seam[A, B]
			push(v.X)
			for _, idx := range v.Indices {
				push(idx)
			}
		case *ast.FuncType:
			pushFieldTypes(v.TypeParams)
			pushFieldTypes(v.Params)
			pushFieldTypes(v.Results)
		case *ast.InterfaceType:
			ifaces = append(ifaces, v)
			for _, m := range v.Methods.List {
				for _, n := range m.Names {
					addName(n.Name)
				}
				push(m.Type)
			}
		case *ast.StructType:
			for _, f := range v.Fields.List {
				if carriesCapability(f.Type, declaredTypes, map[string]bool{}) {
					for _, n := range f.Names {
						addName(n.Name)
					}
				}
				push(f.Type)
			}
		}
	}
	return names, ifaces
}

// carriesCapability reports a type that could hold a method set or a callable.
func carriesCapability(expr ast.Expr, declaredTypes map[string]ast.Expr, seen map[string]bool) bool {
	switch v := expr.(type) {
	case *ast.FuncType, *ast.InterfaceType, *ast.IndexExpr, *ast.IndexListExpr:
		return true
	case *ast.SelectorExpr:
		return true // another package's type: unreadable here, so assumed capable
	case *ast.Ident:
		if seen[v.Name] {
			return false
		}
		seen[v.Name] = true
		if decl, ok := declaredTypes[v.Name]; ok {
			return carriesCapability(decl, declaredTypes, seen)
		}
		return false
	case *ast.StarExpr:
		return carriesCapability(v.X, declaredTypes, seen)
	case *ast.ParenExpr:
		return carriesCapability(v.X, declaredTypes, seen)
	case *ast.ArrayType:
		return carriesCapability(v.Elt, declaredTypes, seen)
	case *ast.Ellipsis:
		return carriesCapability(v.Elt, declaredTypes, seen)
	case *ast.MapType:
		return carriesCapability(v.Key, declaredTypes, seen) ||
			carriesCapability(v.Value, declaredTypes, seen)
	case *ast.ChanType:
		return carriesCapability(v.Value, declaredTypes, seen)
	case *ast.StructType:
		for _, f := range v.Fields.List {
			if carriesCapability(f.Type, declaredTypes, seen) {
				return true
			}
		}
		return false
	}
	return true
}

// methodless reports a shape the guard can positively read an empty method set
// off: a func type, plain data, or an argued-for type from another package.
//
// The point of the positive form is that "the guard did not recognise it" must
// not read as "it is fine". A nil entry in consoleCapabilities is a claim that
// the field carries no method set, and an unresolvable name is not evidence for
// that claim.
func methodless(expr ast.Expr, declaredTypes map[string]ast.Expr, seen map[string]bool) bool {
	switch v := expr.(type) {
	case *ast.FuncType:
		return true
	case *ast.StructType:
		return true // data; its field types are verb-checked by the closure
	case *ast.Ident:
		if goBuiltinTypes[v.Name] {
			return true
		}
		if seen[v.Name] {
			return false
		}
		seen[v.Name] = true
		if decl, ok := declaredTypes[v.Name]; ok {
			return methodless(decl, declaredTypes, seen)
		}
		return false // any, a generic parameter, an unresolvable name
	case *ast.SelectorExpr:
		pkg, ok := v.X.(*ast.Ident)
		return ok && externalOptionTypes[pkg.Name+"."+v.Sel.Name]
	case *ast.StarExpr:
		return methodless(v.X, declaredTypes, seen)
	case *ast.ParenExpr:
		return methodless(v.X, declaredTypes, seen)
	case *ast.ArrayType:
		return methodless(v.Elt, declaredTypes, seen)
	case *ast.MapType:
		return methodless(v.Key, declaredTypes, seen) && methodless(v.Value, declaredTypes, seen)
	case *ast.ChanType:
		return methodless(v.Value, declaredTypes, seen)
	}
	return false
}

// checkVerbs fails when an identifier names a request against the account.
func checkVerbs(t *testing.T, subject, name string) {
	t.Helper()
	checkVerbsExcept(t, subject, name, nil)
}

// checkVerbsExcept is checkVerbs with one seam's argued-for spellings allowed
// through.
//
// An exemption is keyed on the WHOLE identifier, not on the verb. Exempting the
// name "Orders" leaves "PlaceOrder", "CancelOrders" and "OrdersPlacer" failing on
// the same seam, because none of them is that string — which is what keeps the
// hatch the size of the argument written beside it.
func checkVerbsExcept(t *testing.T, subject, name string, exempt map[string]string) {
	t.Helper()
	if _, ok := exempt[name]; ok {
		return
	}
	lowered := strings.ToLower(name)
	for _, verb := range mutationVerbs {
		if strings.Contains(lowered, verb) {
			t.Errorf("%s names the mutation verb %q; this console is handed readings", subject, verb)
		}
	}
}

// optionsFields finds the Options struct wherever in the package it is declared.
func optionsFields(t *testing.T, files map[string]*ast.File) []*ast.Field {
	t.Helper()
	var out []*ast.Field
	found := 0
	for name, file := range files {
		ast.Inspect(file, func(n ast.Node) bool {
			spec, ok := n.(*ast.TypeSpec)
			if !ok || spec.Name.Name != "Options" {
				return true
			}
			st, ok := spec.Type.(*ast.StructType)
			if !ok {
				t.Fatalf("%s declares Options as something other than a struct", name)
				return false
			}
			found++
			out = append(out, st.Fields.List...)
			return false
		})
	}
	if found != 1 {
		t.Fatalf("%d Options declarations were found; the injection surface is one struct", found)
	}
	return out
}

// packageTypes indexes every type this package declares, so a field naming one
// can be resolved to what it actually is.
func packageTypes(files map[string]*ast.File) map[string]ast.Expr {
	out := map[string]ast.Expr{}
	for _, file := range files {
		ast.Inspect(file, func(n ast.Node) bool {
			if spec, ok := n.(*ast.TypeSpec); ok {
				out[spec.Name.Name] = spec.Type
			}
			return true
		})
	}
	return out
}

// parsedPackage parses every non-test source file once.
func parsedPackage(t *testing.T) map[string]*ast.File {
	t.Helper()
	out := map[string]*ast.File{}
	for name, src := range packageFiles(t) {
		out[name] = parseFile(t, name, src)
	}
	return out
}

// TestNoCapabilityReachesTheConsoleAroundOptions.
//
// The Options walk answers "is every capability handed in through that struct
// enumerated". It does not answer the sentence consoleCapabilities actually
// makes, which is "every capability this console is handed is enumerated here" —
// and the difference is one package-level variable wide:
//
//	var packageDesk Desk
//	func (c *Console) SetDesk(d Desk) { packageDesk = d }
//
// with PlaceOrder and CancelOrder on Desk passed the ENTIRE package suite. Options
// never mentions it, no banned import is needed because cmd/tossctl is what fills
// it in, and nothing else in this file was looking anywhere but at that one
// struct.
//
// So this walks the package itself: every exported method on *Console, every
// package-level var, and every interface declared anywhere in the package that is
// not one of the enumerated seams — including the inline `interface{ PlaceOrder(…) }`
// of a type assertion, which is how a capability smuggled in as `any` gets used.
func TestNoCapabilityReachesTheConsoleAroundOptions(t *testing.T) {
	files := parsedPackage(t)
	declaredTypes := packageTypes(files)
	seams := optionsSeamInterfaces(t, files, declaredTypes)

	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)

	// The three positive controls at the bottom read these. They are counted here
	// and not inferred afterwards because the thing being controlled for is a walk
	// that visited nothing, and a walk that visited nothing leaves no other trace.
	visitedInterfaces := map[*ast.InterfaceType]bool{}
	checkedConsoleMethods := 0
	for _, name := range names {
		file := files[name]

		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if d.Recv == nil || !d.Name.IsExported() || !receiverIsConsole(d.Recv) {
					continue
				}
				checkedConsoleMethods++
				subject := "the exported method (*Console)." + d.Name.Name + " in " + name
				checkVerbs(t, subject, d.Name.Name)
				closureNames, ifaces := capabilityClosure(d.Type, declaredTypes)
				for _, n := range closureNames {
					checkVerbs(t, "the type "+n+" reached through "+subject, n)
				}
				for _, iface := range ifaces {
					checkNoEmbedding(t, "an interface reached through "+subject, iface)
				}
			case *ast.GenDecl:
				if d.Tok != token.VAR {
					continue
				}
				for _, spec := range d.Specs {
					vs, ok := spec.(*ast.ValueSpec)
					if !ok {
						continue
					}
					for _, ident := range vs.Names {
						subject := "the package-level var " + ident.Name + " in " + name
						checkVerbs(t, subject, ident.Name)
						if vs.Type == nil {
							continue
						}
						closureNames, ifaces := capabilityClosure(vs.Type, declaredTypes)
						for _, n := range closureNames {
							checkVerbs(t, "the type "+n+" reached through "+subject, n)
						}
						for _, iface := range ifaces {
							checkNoEmbedding(t, "an interface reached through "+subject, iface)
						}
					}
				}
			}
		}

		ast.Inspect(file, func(n ast.Node) bool {
			iface, ok := n.(*ast.InterfaceType)
			if !ok {
				return true
			}
			visitedInterfaces[iface] = true
			if seams[iface] {
				return true
			}
			checkNoEmbedding(t, "an interface declared in "+name, iface)
			for _, m := range iface.Methods.List {
				for _, mn := range m.Names {
					checkVerbs(t, "the method "+mn.Name+" on an interface declared in "+name, mn.Name)
				}
			}
			return true
		})
	}

	// Positive controls. Every assertion above is of the form "nothing found", so
	// each of the three walks passes this test loudest when it looked at nothing at
	// all, and the controls are the only thing standing between that and a green
	// run. They are three and not one because the three walks fail independently.
	//
	// Note what is NOT asserted: that some interface outside the seam set was
	// checked. Today every interface this package declares IS an Options seam, and
	// that is the healthy state rather than a broken walk — an incidental
	// `interface{ PlaceOrder(…) }` appearing is the thing being guarded against,
	// not a precondition of guarding. So the control on that walk is that it
	// reached interface declarations at all, and that it reached the same ones the
	// Options walk resolved by its own separate route.
	if len(seams) == 0 {
		t.Error("no Options field resolved to an interface; the seam set is empty and every interface " +
			"in the package is being checked as if it were an incidental one")
	}
	if len(visitedInterfaces) == 0 {
		t.Error("the package-wide walk reached no interface declaration at all; it is vacuous, and " +
			"an interface{ PlaceOrder(...) } declared anywhere in this package — including the " +
			"inline one of a type assertion — would reach the console unreported")
	}
	missedSeams := 0
	for iface := range seams {
		if !visitedInterfaces[iface] {
			missedSeams++
		}
	}
	if missedSeams > 0 {
		t.Errorf("%d of the %d Options seams were never visited by the package-wide walk; the two "+
			"walks are not looking at the same package, so whatever the package-wide one is "+
			"missing it is missing silently", missedSeams, len(seams))
	}
	if checkedConsoleMethods == 0 {
		t.Error("no exported method on *Console was checked; the method walk is vacuous, and a " +
			"capability attached to the console rather than passed through Options — the P0-3 " +
			"shape this test exists for — would pass the entire package suite")
	}
}

// receiverIsConsole reports a method on *Console (or Console).
func receiverIsConsole(recv *ast.FieldList) bool {
	if recv == nil || len(recv.List) == 0 {
		return false
	}
	expr := recv.List[0].Type
	if star, ok := expr.(*ast.StarExpr); ok {
		expr = star.X
	}
	ident, ok := expr.(*ast.Ident)
	return ok && ident.Name == "Console"
}

// optionsSeamInterfaces is the set of interface declarations that ARE the
// enumerated seams, so the package-wide walk can hold everything else to the
// stricter "no mutation verb at all" rule without re-reporting them.
func optionsSeamInterfaces(t *testing.T, files map[string]*ast.File,
	declaredTypes map[string]ast.Expr) map[*ast.InterfaceType]bool {
	t.Helper()
	out := map[*ast.InterfaceType]bool{}
	for _, field := range optionsFields(t, files) {
		if iface, ok := resolveDeclared(field.Type, declaredTypes).(*ast.InterfaceType); ok {
			out[iface] = true
		}
	}
	return out
}

// TestTheConsoleOpensTheJournalReadOnly.
//
// The spec's "journal 쓰기 시도 차단" scenario. journal.Open creates the data
// directory and migrates the schema; journal.OpenReadOnly does neither and hands
// back a type with no write method on it. This is the guard that fails when
// somebody reaches for the writable constructor because it was the one they
// remembered.
func TestTheConsoleOpensTheJournalReadOnly(t *testing.T) {
	opener := false
	for name, src := range packageFiles(t) {
		code := strings.Join(nonCommentLines(src), "\n")
		for _, banned := range []string{
			"journal.Open(",        // the writable constructor
			"journal.Journal",      // the writable handle
			"journal.Options{",     // its options, which carry the migration override
			"OpenExitState(",       // a write
			"RecordExitJudgement(", // a write
			"BackfillTradeOutcome(",
			"PruneTradeOutcomes(",
		} {
			if strings.Contains(code, banned) {
				t.Errorf("%s contains %q; the console opens the journal with OpenReadOnly and holds "+
					"a *journal.ReadOnly, which has no method that writes", name, banned)
			}
		}
		if strings.Contains(code, "journal.OpenReadOnly(") {
			opener = true
		}
	}
	if !opener {
		t.Error("nothing in the package calls journal.OpenReadOnly; the guard is checking a claim " +
			"nobody makes any more")
	}
}

// TestTheConsoleReadsTheRunLockAndNeverWritesIt.
//
// The console yields the broker refresh to a live verification, and it learns
// about another process's run from internal/runlock's marker. Reading it is the
// whole interaction: a console that acquired the lock would pause the soak for as
// long as somebody left a browser tab open.
func TestTheConsoleReadsTheRunLockAndNeverWritesIt(t *testing.T) {
	reader := false
	for name, src := range packageFiles(t) {
		code := strings.Join(nonCommentLines(src), "\n")
		for _, banned := range []string{"runlock.Acquire(", "runlock.Hold(", ".Refresh(", ".Release()"} {
			if strings.Contains(code, banned) {
				t.Errorf("%s contains %q; the console reads the run marker and never writes or "+
					"removes it", name, banned)
			}
		}
		if strings.Contains(code, "runlock.Fresh(") {
			reader = true
		}
	}
	if !reader {
		t.Error("nothing reads runlock.Fresh; the broker refresh no longer yields to a live verification")
	}
}

// TestTheUnmanagedLabelIsSpelledOnce.
//
// Round-1 P3 fixed one label for a holding the exit policy does not manage. Two
// spellings on two screens would be two things as far as an operator is
// concerned, so the string lives in exactly one place in the Go source and the
// template refers to that.
func TestTheUnmanagedLabelIsSpelledOnce(t *testing.T) {
	const label = "관리 외(미편입)"
	// portfolio.go's positionRow.Label produces it; templates_portfolio.go
	// explains to the reader what it means. Anything else would be a second
	// definition, and a second definition drifts.
	allowed := map[string]bool{"portfolio.go": true, "templates_portfolio.go": true}

	definitions := 0
	for name, src := range packageFiles(t) {
		n := strings.Count(strings.Join(nonCommentLines(src), "\n"), label)
		if n == 0 {
			continue
		}
		if !allowed[name] {
			t.Errorf("%s spells the unmanaged label; positionRow.Label is the one definition", name)
			continue
		}
		if name == "portfolio.go" {
			definitions = n
		}
	}
	switch definitions {
	case 1:
	case 0:
		t.Error("portfolio.go no longer produces the unmanaged label; the positions screen has lost " +
			"its 관리 외 distinction")
	default:
		t.Errorf("portfolio.go produces the label %d times; it is one label with one definition",
			definitions)
	}
}

// TestTheDashboardDescribesAdoptionButCannotPerformIt.
//
// This replaces add-operator-dashboard's round-2 P1 guard, which banned the word
// "adoption" from this package outright. That was right at the dashboard's
// landing — `adoption_id` did not exist in the schema, so a screen naming it
// would have been describing a capability the engine did not have, and an
// operator would reasonably have concluded their manual holdings were protected.
//
// adopt-external-positions landed the capability, and its task 2.7 extends the
// eligibility display to cover it. So the ban is replaced by the rule that
// actually matters and always did: the console may *report* that a position was
// adopted, and it may not adopt one. The distinction is a screen that describes
// the engine versus a screen that drives it.
func TestTheDashboardDescribesAdoptionButCannotPerformIt(t *testing.T) {
	for name, src := range packageFiles(t) {
		code := strings.Join(nonCommentLines(src), "\n")
		// Writing the ledger, in any spelling. The console holds a *ReadOnly and
		// has no write method at all (readonly_test.go enumerates the method set),
		// so these would not compile — which is exactly why naming them is worth
		// failing on: an attempt to add one starts by adding the word.
		for _, banned := range []string{
			"AdoptPosition", "OpenAdoptedExitState", "AdoptionRequest",
			"UPDATE positions", "INSERT INTO position_adoptions",
		} {
			if strings.Contains(code, banned) {
				t.Errorf("%s names %q; adopting a holding is the engine's reconciliation loop's job "+
					"and this package is a read-only view of what it did", name, banned)
			}
		}
	}

	// Positive control: the display half exists. A guard that only forbids is a
	// guard that would still pass if task 2.7 had never been implemented.
	describes := false
	for name, src := range packageFiles(t) {
		if name == "portfolio.go" && strings.Contains(src, "편입 기록") {
			describes = true
		}
	}
	if !describes {
		t.Error("portfolio.go no longer distinguishes an adopted position from an entered one; the " +
			"operator cannot tell why the engine is managing a holding they bought by hand")
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
