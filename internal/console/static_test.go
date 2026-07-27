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
	// Seven verification-console routes, the two dashboard screens
	// (add-operator-dashboard) and the engine's two process-control routes
	// (add-engine-runtime). The floor is asserted so that a guard which stops
	// parsing the table cannot pass by reading nothing.
	if len(routes) < 13 {
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
		// The engine's process control (add-engine-runtime task 2.1). Same
		// reasoning one step up: neither touches an account — the engine's own
		// startup interlock decides whether the process it starts may trade — but
		// both are acts, and starting a trading engine from an embedded image is
		// not a thing a page should be able to do.
		"/engine/start": true,
		"/engine/stop":  true,
		// The adoption-settings edits (console-adoption-controls task 3.2). The
		// only thing either writes is the engine.adoption config block through
		// the injected seam — no journal, no broker, no account — but a config
		// that outlives the console is exactly what CSRF must gate.
		"/settings/save":    true,
		"/settings/include": true,
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
// 엔진 시작/정지), 편입 설정 편집(편입 설정 저장·종목 편입 지정)뿐이다(SHALL —
// 계좌 무접촉; 편입 설정 편집의 대상은 engine.adoption config 블록만이다).
//
// It is the same set TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate uses,
// named separately here because the two tests ask different questions of it: one
// asks whether these routes are gated, the other whether anything outside them
// even looks like an act.
var consoleStateChanging = []string{
	"/verify/start", "/verify/approve", "/verify/abort", "/restart", "/soak/restart",
	"/engine/start", "/engine/stop", "/settings/save", "/settings/include",
}

// TestNoRouteNamesAnAccountMutation.
//
// Two claims in one walk of the route table:
//
//	no account verbs   nothing anywhere in the table is spelled like placing,
//	                   cancelling or amending an order, opening a gate, or
//	                   reaching a credential — including the five routes that ARE
//	                   allowed to change something, because none of them touches
//	                   the account.
//	nothing else acts  every other route is a reading. A path outside the list
//	                   above that reads like an act is either a mistake or a
//	                   requirement change, and both should stop here.
func TestNoRouteNamesAnAccountMutation(t *testing.T) {
	allowed := map[string]bool{}
	for _, path := range consoleStateChanging {
		allowed[path] = true
	}
	// Verbs that would name a request against the account, or a way to open one.
	accountVerbs := []string{
		"order", "sell", "buy", "cancel", "modify", "amend", "flatten",
		"gate", "credential", "secret", "token", "adopt", "enroll",
	}
	// Verbs that name an act rather than a reading. The allowed routes are
	// exempt: the verification control surface, the restarts, and the two
	// adoption-settings edits. The config-write vocabulary is here so a future
	// unlisted /settings/anything cannot sail past this guard the way an
	// unrecognized act otherwise would (console-adoption-controls, review P2-7).
	actVerbs := append([]string{"start", "stop", "approve", "abort", "restart", "reset", "delete",
		"save", "include", "enable", "config"},
		accountVerbs...)

	seen := map[string]bool{}
	for _, r := range registeredRoutes(t) {
		seen[r.Path] = true
		lowered := strings.ToLower(r.Path)
		for _, verb := range accountVerbs {
			if strings.Contains(lowered, verb) {
				t.Errorf("route %s names %q; this console has no route that touches the account, "+
					"its gate or its credentials", r.Path, verb)
			}
		}
		if allowed[r.Path] {
			continue
		}
		for _, verb := range actVerbs {
			if strings.Contains(lowered, verb) {
				t.Errorf("route %s names %q but is not in consoleStateChanging; a route that acts has to "+
					"be argued for and CSRF-gated", r.Path, verb)
			}
		}
	}
	for _, path := range consoleStateChanging {
		if !seen[path] {
			t.Errorf("%s is in the state-changing list but is not registered", path)
		}
	}
}

// TestTheDashboardScreensAreReads.
//
// The positions and history routes exist, go through the session gate like
// everything else, and are NOT behind the CSRF gate — a read route that demanded
// a POST would be a page nobody can open, which is the failure this catches from
// the other side.
func TestTheDashboardScreensAreReads(t *testing.T) {
	want := map[string]bool{"/positions": false, "/history": false}
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

// TestTheConsoleBrokerInterfaceDeclaresNothingButReads.
//
// The spec's "광폭 브로커 인터페이스 주입 차단" scenario. verifylive.Broker has
// PlaceOrder, CancelOrder, ModifyOrder and three conditional-order mutations on
// it; handing that to a read-only screen would make "the console places no order"
// a fact about what the handlers happen to call rather than about what they can.
//
// The interface is read out of the source rather than through reflection because
// the claim is about the declaration: a method that exists but is never called at
// runtime is exactly what this has to catch.
func TestTheConsoleBrokerInterfaceDeclaresNothingButReads(t *testing.T) {
	src := packageFiles(t)["holdings.go"]
	if src == "" {
		t.Fatal("holdings.go is missing; the broker interface cannot be checked")
	}
	file := parseFile(t, "holdings.go", src)

	allowed := map[string]bool{"Holdings": true}
	banned := []string{
		"place", "order", "cancel", "modify", "amend", "sell", "buy", "create",
		"delete", "update", "submit", "transfer", "withdraw", "conditional",
	}

	var declared []string
	ast.Inspect(file, func(n ast.Node) bool {
		spec, ok := n.(*ast.TypeSpec)
		if !ok || spec.Name.Name != "HoldingsReader" {
			return true
		}
		iface, ok := spec.Type.(*ast.InterfaceType)
		if !ok {
			t.Fatal("HoldingsReader is not an interface; the console's broker seam must stay a narrow one")
			return false
		}
		for _, method := range iface.Methods.List {
			for _, name := range method.Names {
				declared = append(declared, name.Name)
			}
			if len(method.Names) == 0 {
				t.Error("HoldingsReader embeds another interface; whatever that interface gains, " +
					"this one gains silently")
			}
		}
		return false
	})

	if len(declared) == 0 {
		t.Fatal("no HoldingsReader methods were read; the guard is not looking at the interface")
	}
	for _, name := range declared {
		if !allowed[name] {
			t.Errorf("HoldingsReader declares %q; the console is handed reads and nothing else", name)
		}
		lowered := strings.ToLower(name)
		for _, verb := range banned {
			if strings.Contains(lowered, verb) {
				t.Errorf("HoldingsReader declares %q, which names the mutation verb %q", name, verb)
			}
		}
	}

	// And nothing in the package names verifylive's wide broker as its own seam.
	for name, fileSrc := range packageFiles(t) {
		code := strings.Join(nonCommentLines(fileSrc), "\n")
		if strings.Contains(code, "verifylive.Broker") {
			t.Errorf("%s takes a verifylive.Broker; that interface can place, amend and cancel orders, "+
				"and the dashboard is handed HoldingsReader instead", name)
		}
	}
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
