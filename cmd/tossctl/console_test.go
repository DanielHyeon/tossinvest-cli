package main

// console_test.go covers the CLI surface of `tossctl console` (task 1.6).
//
// The console's behaviour — the session gate, the CSRF gate, the typed nonce, the
// process boundary — is tested in internal/console against a fake broker. What is
// tested here is the thing only the command can get wrong, and it is the condition
// under which task 1.6 permits a web approval to exist at all:
//
//	"CLI 단독으로 비대화 승인이 가능해지는 플래그·경로를 만들지 않는다(콘솔 내부 배선만)."
//
// So: `verify run` still confirms at the terminal, in the source; the console
// offers no flag that could answer anything; and no file but console.go can even
// reach the console package.

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/console"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
)

// --- registration and flags -------------------------------------------------------

func TestConsoleIsRegisteredAndAnnotated(t *testing.T) {
	cmd := findCommand(newRootCmd(), "console")
	if cmd == nil {
		t.Fatal("console is not registered on the root command")
	}
	if got := cmd.Annotations["source"]; got != "official" {
		t.Errorf("source = %q, want official — the verification it drives reaches the Open API", got)
	}
	if cmd.Annotations["mutating"] != "true" {
		t.Error("console is not annotated mutating; it can place live orders through the verify runner")
	}
	if strings.TrimSpace(cmd.Short) == "" {
		t.Error("console has no Short description")
	}
	for _, want := range []string{"127.0.0.1", "session token", "CSRF", "read-only"} {
		if !strings.Contains(cmd.Long, want) {
			t.Errorf("the help does not mention %q; an operator has to know what they are opening", want)
		}
	}
}

// TestConsoleOffersOnlyThePortFlag.
//
// Every banned name here is a different way of making the console answer for a
// person: a preset token, a preset approval, or an interface somebody else can
// reach. The task's wording is "127.0.0.1 전용 바인딩(비루프백 거부)" and there is no
// flag that negotiates with it.
func TestConsoleOffersOnlyThePortFlag(t *testing.T) {
	cmd := findCommand(newRootCmd(), "console")
	if cmd == nil {
		t.Fatal("console is not registered")
	}
	if cmd.Flags().Lookup("port") == nil {
		t.Error("console has no --port")
	}

	banned := []string{
		// the interface
		"host", "bind", "address", "addr", "interface", "listen", "public", "remote", "expose", "lan",
		// the authentication
		"token", "session", "session-token", "csrf", "no-auth", "insecure", "open",
		// the approval
		"yes", "y", "force", "f", "approve", "auto-approve", "nonce", "confirm", "no-confirm",
		"non-interactive", "unattended", "skip-confirmation",
		// the evidence's provenance
		"base-url", "api-base-url", "endpoint",
		// the finer gate the console deliberately does not offer
		"confirm-each",
	}
	for _, name := range banned {
		if cmd.Flags().Lookup(name) != nil {
			t.Errorf("console offers --%s", name)
		}
		if len(name) == 1 && cmd.Flags().ShorthandLookup(name) != nil {
			t.Errorf("console offers -%s", name)
		}
	}

	// And nothing beyond --port. Asserted against the source rather than the
	// flag set, because the flag set also carries the root command's persistent
	// flags and a source count is what makes "a flag added here has to be
	// justified in this test" true.
	if got := declaredFlags(t, "console.go"); len(got) != 1 || got[0] != "port" {
		t.Errorf("console.go registers %v, want exactly [port] — every other knob on this command would be "+
			"a knob on an approval surface", got)
	}
}

// declaredFlags reads the flag names a file registers.
func declaredFlags(t *testing.T, name string) []string {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), name, readSource(t, name), 0)
	if err != nil {
		t.Fatalf("parsing %s: %v", name, err)
	}
	var names []string
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || !strings.HasSuffix(sel.Sel.Name, "Var") || len(call.Args) < 2 {
			return true
		}
		lit, ok := call.Args[1].(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}
		names = append(names, strings.Trim(lit.Value, `"`))
		return true
	})
	return names
}

// TestConsoleProbeSymbolMatchesVerifyRunsDefault. The two are spelled separately
// (console.go says why); this is what stops them drifting apart.
func TestConsoleProbeSymbolMatchesVerifyRunsDefault(t *testing.T) {
	run := findCommand(findCommand(newRootCmd(), "verify"), "run")
	if run == nil {
		t.Fatal("verify run is not registered")
	}
	f := run.Flags().Lookup("symbol")
	if f == nil {
		t.Fatal("verify run has no --symbol")
	}
	if f.DefValue != consoleProbeSymbolKR {
		t.Errorf("console probes %s but `verify run --symbol` defaults to %s; the two must agree",
			consoleProbeSymbolKR, f.DefValue)
	}
}

// --- the CLI must not gain a non-interactive approval path ---------------------------

// TestVerifyRunStillConfirmsAtTheTerminalOnly.
//
// This is the source-level half of task 1.6's condition. The console supplies its
// own confirmer to its own runner; `verify run` must keep supplying the terminal's,
// and no runtime test would notice if it stopped — the offending line would simply
// never be executed by a test binary that has no terminal anyway.
func TestVerifyRunStillConfirmsAtTheTerminalOnly(t *testing.T) {
	want := map[string]string{
		"Confirm":      "terminalConfirmer",
		"ConfirmBatch": "terminalBatchConfirmer",
	}
	got := runnerConfirmers(t, "verify.go")
	for field, fn := range want {
		if got[field] != fn {
			t.Errorf("verify.go builds its runner with %s: %s, want %s(cmd) — `verify run` must never be "+
				"able to reach the console's web approval", field, orNothing(got[field]), fn)
		}
	}

	src := readSource(t, "verify.go")
	if strings.Contains(src, modulePathPrefix+"internal/console") {
		t.Error("verify.go imports internal/console; the two approval surfaces must not be wired together")
	}
}

// TestConsoleWiresTheWebConfirmerAndRefusesPerMutationPrompts.
//
// The positive counterpart: the console's runner takes the web batch confirmer,
// and its per-mutation confirmer is the refusing one — so the gate the console
// does not offer fails closed rather than being silently satisfied.
func TestConsoleWiresTheWebConfirmerAndRefusesPerMutationPrompts(t *testing.T) {
	got := runnerConfirmers(t, "console.go")
	if got["ConfirmBatch"] != "confirm" {
		t.Errorf("console.go builds its runner with ConfirmBatch: %s, want the injected web confirmer",
			orNothing(got["ConfirmBatch"]))
	}
	if got["Confirm"] != "consoleMutationConfirmer" {
		t.Errorf("console.go builds its runner with Confirm: %s, want consoleMutationConfirmer()",
			orNothing(got["Confirm"]))
	}
	if strings.Contains(readSource(t, "console.go"), "ConfirmEach:") {
		t.Error("console.go sets ConfirmEach; the console is batch-only by design")
	}
}

func TestConsoleMutationConfirmerRefuses(t *testing.T) {
	err := consoleMutationConfirmer()(verifylive.Mutation{Step: verifylive.StepOrderCancel})
	if err == nil {
		t.Fatal("the console's per-mutation confirmer approved a mutation")
	}
	if err != verifylive.ErrNotATerminal { //nolint:errorlint // the sentinel identity is the assertion
		t.Errorf("err = %v, want ErrNotATerminal", err)
	}
}

// TestOnlyConsoleGoReachesTheConsolePackage.
//
// "콘솔 내부 배선만": the console's route to a live account exists in exactly one
// file. A second importer would be a second place the web confirmer could be
// wired, and reviewing one file is the only way that claim stays checkable.
func TestOnlyConsoleGoReachesTheConsolePackage(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("reading the package directory: %v", err)
	}
	importers := 0
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		if !strings.Contains(readSource(t, name), `"`+modulePathPrefix+`internal/console"`) {
			continue
		}
		importers++
		if name != "console.go" {
			t.Errorf("%s imports internal/console; only console.go may", name)
		}
	}
	if importers != 1 {
		t.Errorf("%d file(s) import internal/console, want exactly console.go", importers)
	}
}

// --- the dashboard's broker (change add-operator-dashboard) ------------------------

// TestTheConsoleIsHandedOneCapabilityAndNotABroker.
//
// internal/console declares a one-method interface and a static test there keeps
// it that way. This is the other end of the same claim: the value this file hands
// across the boundary has exactly one method too, so the Open API client's
// PlaceOrder / CancelOrder / ModifyOrder are not reachable from the console even
// by type assertion.
func TestTheConsoleIsHandedOneCapabilityAndNotABroker(t *testing.T) {
	// One shared resolver behind both seams, as runConsole wires them
	// (TestOpeningEveryConsoleReadScreenResolvesTheAccountOnce owns that half).
	// Sharing the resolution changes nothing about the shape of what crosses the
	// boundary, which is what this test is here to keep saying.
	shared := newConsoleBroker(&rootOptions{})

	var reader console.HoldingsReader = newConsoleHoldings(shared)

	typ := reflect.TypeOf(reader)
	if typ.NumMethod() != 1 || typ.Method(0).Name != "Holdings" {
		names := make([]string, 0, typ.NumMethod())
		for i := 0; i < typ.NumMethod(); i++ {
			names = append(names, typ.Method(i).Name)
		}
		t.Fatalf("the console's holdings reader declares %v, want exactly [Holdings]", names)
	}

	// The orders screen's read is the same claim a second time (change
	// console-orders-screen, task 4.6): one method, and the wide broker it came
	// from is not reachable from the value that crosses the boundary.
	var orders console.OrdersReader = consoleOrdersSeam(shared)
	ordersType := reflect.TypeOf(orders)
	if ordersType.NumMethod() != 1 || ordersType.Method(0).Name != "Orders" {
		names := make([]string, 0, ordersType.NumMethod())
		for i := 0; i < ordersType.NumMethod(); i++ {
			names = append(names, ordersType.Method(i).Name)
		}
		t.Fatalf("the console's orders reader declares %v, want exactly [Orders]", names)
	}

	// And the console.Options literal never receives a broker of any kind. The
	// field list is read out of the source because the failure this guards against
	// is somebody adding one, which no runtime test would ever execute.
	fields := consoleOptionFields(t)
	if len(fields) == 0 {
		t.Fatal("no console.Options literal was found; the guard is not reading the wiring")
	}
	if !fields["Holdings"] {
		t.Error("console.Options is built without Holdings; the positions screen has no broker")
	}
	if !fields["Orders"] {
		t.Error("console.Options is built without Orders; the orders screen has no read")
	}
	for name := range fields {
		if why, exempt := consoleFieldExemptions[name]; exempt {
			if strings.TrimSpace(why) == "" {
				t.Errorf("the exemption for the %s field has no reason; an unargued exemption is "+
					"this ban quietly getting shorter", name)
			}
			continue
		}
		lowered := strings.ToLower(name)
		for _, banned := range []string{"broker", "client", "order", "place", "cancel"} {
			if strings.Contains(lowered, banned) {
				t.Errorf("console.Options is given a %s field; the dashboard's whole broker is "+
					"console.HoldingsReader", name)
			}
		}
	}
}

// consoleFieldExemptions are the console.Options fields whose names may contain a
// word the ban above looks for, each with the argument.
//
// It is a hole and it is the size of this map. The alternative — deleting "order"
// from the list — would stop catching a future PlaceOrder field at the same
// moment, and that is what the list is for. A field named here is still held to
// the shape check above: the value handed across the boundary has one method, and
// the wide broker is not reachable from it.
var consoleFieldExemptions = map[string]string{
	"Orders": "the orders screen's read (change console-orders-screen). It lists the account's " +
		"orders and cannot be named without the word; the seam declares exactly Orders, and " +
		"lazyOrders takes its two read method values off the shared client per call and keeps " +
		"neither them nor the verifylive.Broker they came from",
}

// consoleOptionFields reads the keys of the console.Options literal in console.go.
func consoleOptionFields(t *testing.T) map[string]bool {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), "console.go", readSource(t, "console.go"), 0)
	if err != nil {
		t.Fatalf("parsing console.go: %v", err)
	}
	out := map[string]bool{}
	ast.Inspect(file, func(n ast.Node) bool {
		lit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}
		sel, ok := lit.Type.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "Options" {
			return true
		}
		pkg, ok := sel.X.(*ast.Ident)
		if !ok || pkg.Name != "console" {
			return true
		}
		for _, elt := range lit.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			if key, ok := kv.Key.(*ast.Ident); ok {
				out[key.Name] = true
			}
		}
		return false
	})
	return out
}

// TestTheConsoleReadsTheJournalPathAndTheRunLockFromTheSamePlacesEverythingElseDoes.
//
// The dashboard has no paths of its own: the journal is journal.DefaultPath()
// (the engine's), and the run marker is the one `verify run` and the soak already
// agree on. A second answer to "where is the journal" would be a console
// reporting on a database nobody else is using.
func TestTheConsoleReadsTheJournalPathAndTheRunLockFromTheSamePlacesEverythingElseDoes(t *testing.T) {
	src := readSource(t, "console.go")
	for _, want := range []string{"journal.DefaultPath()", "verifyRunLockPath(verifyRecord)"} {
		if !strings.Contains(src, want) {
			t.Errorf("console.go no longer resolves the dashboard's paths with %s", want)
		}
	}
	if strings.Contains(src, "journal.Open(") {
		t.Error("console.go opens the journal writable for the console; the dashboard opens it read-only " +
			"itself, and this path would create and migrate it")
	}
}

// --- helpers ----------------------------------------------------------------------

const modulePathPrefix = "github.com/JungHoonGhae/tossinvest-cli/"

func readSource(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Clean(name))
	if err != nil {
		t.Fatalf("reading %s: %v", name, err)
	}
	return string(data)
}

// runnerConfirmers reports which function each confirmer field of a file's
// verifylive.Options literal is built from.
func runnerConfirmers(t *testing.T, name string) map[string]string {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), name, readSource(t, name), 0)
	if err != nil {
		t.Fatalf("parsing %s: %v", name, err)
	}

	out := map[string]string{}
	found := false
	ast.Inspect(file, func(n ast.Node) bool {
		lit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}
		sel, ok := lit.Type.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "Options" {
			return true
		}
		pkg, ok := sel.X.(*ast.Ident)
		if !ok || pkg.Name != "verifylive" {
			return true
		}
		found = true
		for _, elt := range lit.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			key, ok := kv.Key.(*ast.Ident)
			if !ok || (key.Name != "Confirm" && key.Name != "ConfirmBatch") {
				continue
			}
			out[key.Name] = describeConfirmer(kv.Value)
		}
		return true
	})
	if !found {
		t.Fatalf("%s builds no verifylive.Options; the guard is not reading the wiring", name)
	}
	return out
}

// describeConfirmer names the expression a confirmer field is set from: the
// function that is called, or the identifier that is passed through.
func describeConfirmer(expr ast.Expr) string {
	switch v := expr.(type) {
	case *ast.CallExpr:
		if id, ok := v.Fun.(*ast.Ident); ok {
			return id.Name
		}
		if sel, ok := v.Fun.(*ast.SelectorExpr); ok {
			return sel.Sel.Name
		}
	case *ast.Ident:
		return v.Name
	}
	return "an expression this guard does not recognise"
}

func orNothing(s string) string {
	if s == "" {
		return "(nothing)"
	}
	return s
}

// --- the overview's Guardian limits (change console-operator-overview, task 5.1) ---

// TestTheConsoleIsHandedTheLimitsAsNumbersAndNoWayToWriteThem.
//
// The overview displays the automation gate's ceilings. What crosses the boundary
// to do that is five float64s and a currency string — not the config service,
// which owns Init and a surgical writer, and not the gate's own type, which
// internal/console is forbidden to name at all. A seam with a second method would
// be the console gaining the ability to move a risk limit from a browser, which
// its spec says in writing it does not have.
func TestTheConsoleIsHandedTheLimitsAsNumbersAndNoWayToWriteThem(t *testing.T) {
	seam := consoleGateLimitsSeam(&rootOptions{configDir: t.TempDir()})
	if seam == nil {
		t.Fatal("no limits seam was built for a resolvable config directory")
	}

	typ := reflect.TypeOf(seam)
	if typ.NumMethod() != 1 || typ.Method(0).Name != "GateLimits" {
		names := make([]string, 0, typ.NumMethod())
		for i := 0; i < typ.NumMethod(); i++ {
			names = append(names, typ.Method(i).Name)
		}
		t.Fatalf("the limits seam declares %v, want exactly [GateLimits]", names)
	}

	if !consoleOptionFields(t)["GateLimits"] {
		t.Error("console.Options is built without GateLimits; the overview's safety panel has no " +
			"ceilings to show and renders them as seam 미배선 forever")
	}
}

// TestTheLimitsReadIsBounded.
//
// The seam takes no context, so the deadline has to be set inside it. What is on
// the other end is a screen that reloads every 30 seconds and stays open all day:
// a config read with a bare context.Background() behind it would hold an HTTP
// handler open with no bound on a wedged filesystem, and a render that never
// returns is the one failure the overview has no words for — every other one is a
// sentence on the page.
func TestTheLimitsReadIsBounded(t *testing.T) {
	src, err := os.ReadFile("console.go")
	if err != nil {
		t.Fatalf("reading console.go: %v", err)
	}
	code := string(src)
	if strings.Contains(code, "s.svc.Load(context.Background())") {
		t.Error("consoleGateLimits.GateLimits loads the config with an unbounded context")
	}
	if !strings.Contains(code, "context.WithTimeout(context.Background(), consoleGateLimitsTimeout)") {
		t.Error("consoleGateLimits.GateLimits no longer bounds its read with consoleGateLimitsTimeout")
	}
	if consoleGateLimitsTimeout <= 0 {
		t.Errorf("consoleGateLimitsTimeout is %s; a non-positive deadline expires immediately",
			consoleGateLimitsTimeout)
	}
}

// TestTheConsoleComesUpWithoutTheLimitsSeam.
//
// A console that refused to start because it could not resolve a config file
// would be a console unavailable at exactly the moment somebody is trying to work
// out what is wrong. The seam is nil and the overview says so, panel by panel.
func TestTheConsoleComesUpWithoutTheLimitsSeam(t *testing.T) {
	c, err := console.New(console.Options{
		StartVerify: func(
			context.Context, verifylive.BatchConfirmer, io.Writer, string, []verifylive.StepID,
		) (verifylive.Summary, []verifylive.Entry, error) {
			return verifylive.Summary{}, nil, nil
		},
	})
	if err != nil {
		t.Fatalf("a console with no limits seam refused to start: %v", err)
	}
	if c.Handler() == nil {
		t.Error("the console came up with no handler")
	}
}

// --- the orders screen's read (change console-orders-screen, task 4.6) -------------

// TestOneRefreshAsksTheOpenGroupAndTheClosedGroupSeparatelyAndTheLiveOneWhole.
//
// This is the wire, and it is the finding an adversarial review of the first
// implementation opened with (design D2, amended). `status` is `required: true`
// on GET /api/v1/orders, and the two groups behave differently in a way that
// decides whether this screen can do its job at all:
//
//	status=OPEN    "모든 대기 중 주문을 전량 반환합니다. limit, cursor 는 무시되며"
//	status=CLOSED  "limit (기본 20, 최대 100), cursor, from/to 파라미터 모두 적용"
//
// The first implementation sent neither — `?limit=100` and nothing else. That is
// one of two failures depending on how the broker treats a missing required
// parameter: a rejection, which makes the live count permanently unmeasured, or
// one hundred rows spanning the account's whole history, in which case a live
// order past row 100 is absent from the table AND from the count, rendered as
// "0건 이상". The leftover then hides behind an 이상 nobody can resolve, because
// this adapter deliberately does not walk pages (I-3).
//
// So the live half is asked for by the name that cannot truncate, and the closed
// half is a separate request because a cancelled or rejected order never becomes
// a trade and therefore never appears on /history either. Three calls, and the
// honesty is what the number is for.
func TestOneRefreshAsksTheOpenGroupAndTheClosedGroupSeparatelyAndTheLiveOneWhole(t *testing.T) {
	type wire struct {
		path  string
		query url.Values
	}
	var (
		mu   sync.Mutex
		seen []wire
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth2/token" {
			fmt.Fprint(w, `{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`)
			return
		}
		mu.Lock()
		seen = append(seen, wire{path: r.URL.Path, query: r.URL.Query()})
		mu.Unlock()

		switch {
		case r.URL.Path == "/api/v1/orders" && r.URL.Query().Get("status") == "OPEN":
			fmt.Fprint(w, `{"result":{"orders":[{"orderId":"live-1","symbol":"005930","side":"BUY",`+
				`"orderType":"MARKET","status":"PENDING","quantity":"10","price":null,`+
				`"currency":"KRW","orderedAt":"2026-07-27T09:30:00+09:00","canceledAt":null,`+
				`"execution":null}],"nextCursor":null,"hasNext":false}}`)
		case r.URL.Path == "/api/v1/orders" && r.URL.Query().Get("status") == "CLOSED":
			fmt.Fprint(w, `{"result":{"orders":[{"orderId":"done-1","symbol":"AAPL","side":"SELL",`+
				`"orderType":"LIMIT","status":"CANCELED","quantity":"5","price":"200",`+
				`"currency":"USD","orderedAt":"2026-07-27T00:10:00Z",`+
				`"canceledAt":"2026-07-27T00:11:00Z","execution":null}],`+
				`"nextCursor":"c2","hasNext":true}}`)
		case r.URL.Path == "/api/v1/orders":
			// The shape this test exists to make impossible. The broker documents
			// status as required; answering 400 here is the friendlier of the two
			// things a real one does.
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error":{"code":"missing-required-parameter","message":"status"}}`)
		case r.URL.Path == "/api/v1/conditional-orders":
			fmt.Fprint(w, `{"result":{"conditionalOrders":[],"nextCursor":null,"hasNext":false}}`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	previous := verifyBrokerFactory
	verifyBrokerFactory = func(*rootOptions) (verifylive.Broker, string, error) {
		return official.New(
			official.Credentials{APIKey: "k", SecretKey: "s"},
			filepath.Join(t.TempDir(), "token.json"),
			official.WithBaseURL(srv.URL),
			official.WithHTTPClient(srv.Client()),
			official.WithAccountSeq(7),
		), "123-45-678901", nil
	}
	t.Cleanup(func() { verifyBrokerFactory = previous })

	reading, err := consoleOrdersSeam(newConsoleBroker(&rootOptions{})).Orders(context.Background())
	if err != nil {
		t.Fatalf("Orders: %v", err)
	}

	mu.Lock()
	got := append([]wire(nil), seen...)
	mu.Unlock()

	if len(got) != 3 {
		t.Fatalf("one refresh made %d broker call(s): %+v; the budget for this screen is three — "+
			"미체결 1콜 + 종결 1콜 + 조건주문 1콜", len(got), got)
	}

	live, closed, conditional := got[0], got[1], got[2]
	if live.path != "/api/v1/orders" || live.query.Get("status") != "OPEN" {
		t.Errorf("the live read went out as %s?%s; without status=OPEN the broker either refuses "+
			"the request or answers one page over the account's whole history, and the live order "+
			"past row 100 is then missing from the table and from the count",
			live.path, live.query.Encode())
	}
	if live.query.Get("limit") != "" || live.query.Get("cursor") != "" {
		t.Errorf("the live read carried limit=%q cursor=%q; the API ignores both for OPEN, and "+
			"sending them says on the wire that this call can be truncated when it cannot",
			live.query.Get("limit"), live.query.Get("cursor"))
	}
	if closed.path != "/api/v1/orders" || closed.query.Get("status") != "CLOSED" {
		t.Errorf("the second read went out as %s?%s, want the CLOSED group — a cancelled or "+
			"rejected order never becomes a trade, so /history never shows it and this call is "+
			"the only place that information exists", closed.path, closed.query.Encode())
	}
	if closed.query.Get("limit") != strconv.Itoa(consoleOrdersPageLimit) {
		t.Errorf("the closed read asked for limit=%q, want %d", closed.query.Get("limit"),
			consoleOrdersPageLimit)
	}
	if conditional.path != "/api/v1/conditional-orders" {
		t.Errorf("the third call was %s, want the conditional endpoint", conditional.path)
	}

	// And the reading keeps the two plain lists apart, because only one of them
	// can be truncated.
	if len(reading.Open) != 1 || reading.Open[0].ID != "live-1" {
		t.Errorf("the open list is %+v, want the one pending order", reading.Open)
	}
	if reading.OpenTruncated {
		t.Error("the open list reports a truncation; OPEN returns every pending order, so the " +
			"live count is a number rather than a floor")
	}
	if len(reading.Closed) != 1 || reading.Closed[0].ID != "done-1" {
		t.Errorf("the closed list is %+v, want the one cancelled order", reading.Closed)
	}
	if !reading.ClosedTruncated {
		t.Error("hasNext was dropped from the closed list; that count would be a settled number")
	}
}

// TestTheOrdersSeamResolvesTheAccountOnceAndBuildsNoSecondClient.
//
// verifylive.Broker does not declare the raw order reads, so the path has to be
// chosen explicitly, and the tempting choice is a second *official.Client built
// here. That client resolves the account sequence again — the /api/v1/accounts
// read that came back 429 three times on 2026-07-26 and cost a run three steps
// (measurements.md M4) — for an answer the first resolution already had.
//
// So the seam reads through the console's shared client, and this is what fails if
// somebody adds a second construction path.
//
// This is the *per-seam* half only: it counts builds behind one seam across
// refreshes and says nothing about what a whole console session costs. The
// console-level claim — every read screen together, one resolution — is
// TestOpeningEveryConsoleReadScreenResolvesTheAccountOnce, and it was written
// because this test passed for the whole time a session that opened the positions
// screen and /orders resolved the account twice.
func TestTheOrdersSeamResolvesTheAccountOnceAndBuildsNoSecondClient(t *testing.T) {
	srv := newVerifyServer(t)
	built := 0
	previous := verifyBrokerFactory
	verifyBrokerFactory = func(*rootOptions) (verifylive.Broker, string, error) {
		built++
		return official.New(
			official.Credentials{APIKey: "k", SecretKey: "s"},
			filepath.Join(t.TempDir(), "token.json"),
			official.WithBaseURL(srv.URL),
			official.WithHTTPClient(srv.Client()),
			official.WithAccountSeq(7),
		), "123-45-678901", nil
	}
	t.Cleanup(func() { verifyBrokerFactory = previous })

	seam := consoleOrdersSeam(newConsoleBroker(&rootOptions{}))
	for i := 0; i < 3; i++ {
		if _, err := seam.Orders(context.Background()); err != nil {
			t.Fatalf("Orders: %v", err)
		}
	}
	if built != 1 {
		t.Errorf("the broker was built %d times for three refreshes; each build resolves the "+
			"account sequence and that is the read that gets rate limited", built)
	}

	// The source half: nothing in this file constructs a client of its own.
	src := readSource(t, "console.go")
	if strings.Contains(src, "official.New(") {
		t.Error("console.go constructs an *official.Client. The orders seam reuses the one " +
			"verifyBrokerFactory already resolved the account for; a second one duplicates that " +
			"resolution and doubles the 429 exposure")
	}
	if !strings.Contains(src, "l.shared.resolve()") {
		t.Error("the orders seam no longer reads through the console's shared client; a resolver " +
			"of its own is a second /api/v1/accounts read per session")
	}
}

// TestTheOrdersSeamCarriesEachListsOutcomeSeparately.
//
// One refresh is three calls, and losing one of them must not take the others
// down. The console renders the missing part as unmeasured and refuses to add the
// counts; returning an error for the set would throw away the parts that answered
// and leave the screen with nothing to be honest about.
func TestTheOrdersSeamCarriesEachListsOutcomeSeparately(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/oauth2/token":
			fmt.Fprint(w, `{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`)
		case r.URL.Path == "/api/v1/orders" && r.URL.Query().Get("status") == "OPEN":
			fmt.Fprint(w, `{"result":{"orders":[{"orderId":"o-1","symbol":"005930","side":"BUY",`+
				`"orderType":"MARKET","status":"PENDING","quantity":"10","price":null,`+
				`"currency":"KRW","orderedAt":"2026-07-27T09:30:00+09:00","canceledAt":null,`+
				`"execution":null}],"nextCursor":null,"hasNext":false}}`)
		case r.URL.Path == "/api/v1/orders":
			// The closed page is the one that can be truncated.
			fmt.Fprint(w, `{"result":{"orders":[{"orderId":"o-0","symbol":"005930","side":"BUY",`+
				`"orderType":"LIMIT","status":"FILLED","quantity":"10","price":"70000",`+
				`"currency":"KRW","orderedAt":"2026-07-26T09:30:00+09:00","canceledAt":null,`+
				`"execution":{"filledQuantity":"10","averageFilledPrice":"70000"}}],`+
				`"nextCursor":"c2","hasNext":true}}`)
		case r.URL.Path == "/api/v1/conditional-orders":
			// The part that fails.
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprint(w, `{"error":{"code":"rate-limited"}}`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	previous := verifyBrokerFactory
	verifyBrokerFactory = func(*rootOptions) (verifylive.Broker, string, error) {
		return official.New(
			official.Credentials{APIKey: "k", SecretKey: "s"},
			filepath.Join(t.TempDir(), "token.json"),
			official.WithBaseURL(srv.URL),
			official.WithHTTPClient(srv.Client()),
			official.WithAccountSeq(7),
		), "123-45-678901", nil
	}
	t.Cleanup(func() { verifyBrokerFactory = previous })

	reading, err := consoleOrdersSeam(newConsoleBroker(&rootOptions{})).Orders(context.Background())
	if err != nil {
		t.Fatalf("Orders returned an error for a half-answered reading: %v", err)
	}
	if len(reading.Open) != 1 || reading.Open[0].ID != "o-1" {
		t.Fatalf("the open list did not survive the conditional failure: %+v", reading.Open)
	}
	if reading.OpenError != "" {
		t.Errorf("the open list carries an error it did not have: %q", reading.OpenError)
	}
	if len(reading.Closed) != 1 || reading.Closed[0].ID != "o-0" {
		t.Fatalf("the closed list did not survive the conditional failure: %+v", reading.Closed)
	}
	if !reading.ClosedTruncated {
		t.Error("hasNext was dropped; the closed count would be rendered as a settled number")
	}
	if reading.ConditionalError == "" {
		t.Error("the conditional failure was swallowed; the screen would render 0 conditional " +
			"orders as a measured value while a leftover held the exposure cap")
	}
	// And the absent values arrived absent rather than as zero.
	if reading.Open[0].Price != "" || reading.Open[0].AverageFilledPrice != "" {
		t.Errorf("a null price/execution arrived as %q/%q; the raw read exists so that an absent "+
			"value is not a number", reading.Open[0].Price, reading.Open[0].AverageFilledPrice)
	}
}

// consoleBrokerBuildSites are the functions in console.go that may build a live
// client of their own, each with the argument for why its resolution is separate.
//
// Every entry costs one more /api/v1/accounts read per console process — the read
// that came back 429 three times on 2026-07-26 and cost a verification run three
// steps (measurements.md M4). The map is the whole permitted set: a seam that
// resolves on its own instead of sharing has to be argued for here first, which is
// what makes "the console resolves the account once" checkable rather than a claim
// somebody has to re-derive from the wiring.
var consoleBrokerBuildSites = map[string]string{
	"consoleBroker.resolve": "the console's one read client. Every read seam is handed this same " +
		"resolver, so opening every screen the console serves costs one resolution per process " +
		"rather than one per screen",
	"consoleVerifyStarter": "a verification run resolves its own account at run start, on purpose. " +
		"The record it writes names the account it confirmed at the moment real orders were about " +
		"to be sent, not one resolved whenever a read screen happened to be opened — possibly " +
		"before `tossctl openapi login` replaced the credentials. One resolution per run started",
}

// TestOpeningEveryConsoleReadScreenResolvesTheAccountOnce.
//
// The claim is about the console, not about one seam. Each read seam already
// resolved the account once *inside itself*, and that was true while a session
// that opened the positions screen and then /orders resolved it twice — which is
// the shape a per-seam test cannot see and this one is for.
//
// The runtime half opens every read screen the console serves, twice each and then
// both at once, against a factory that counts. The static half binds it to the
// wiring: runConsole builds one shared resolver and hands it to every seam, and
// nothing else in console.go builds a client except the sites argued for in
// consoleBrokerBuildSites.
func TestOpeningEveryConsoleReadScreenResolvesTheAccountOnce(t *testing.T) {
	srv := newVerifyServer(t)
	var (
		mu    sync.Mutex
		built int
	)
	previous := verifyBrokerFactory
	// The factory stands in for buildVerifyBroker, whose body is the
	// /api/v1/accounts read this test counts: one call here is one resolution.
	verifyBrokerFactory = func(*rootOptions) (verifylive.Broker, string, error) {
		mu.Lock()
		built++
		mu.Unlock()
		return official.New(
			official.Credentials{APIKey: "k", SecretKey: "s"},
			filepath.Join(t.TempDir(), "token.json"),
			official.WithBaseURL(srv.URL),
			official.WithHTTPClient(srv.Client()),
			official.WithAccountSeq(7),
		), "123-45-678901", nil
	}
	t.Cleanup(func() { verifyBrokerFactory = previous })

	resolutions := func() int {
		mu.Lock()
		defer mu.Unlock()
		return built
	}

	// The seams as runConsole wires them: one resolver, built once at startup,
	// every screen opened off it afterwards.
	shared := newConsoleBroker(&rootOptions{})
	positions, orders := newConsoleHoldings(shared), consoleOrdersSeam(shared)
	screens := []struct {
		name string
		open func() error
	}{
		{"/dashboard 포지션", func() error {
			_, err := positions.Holdings(context.Background(), "")
			return err
		}},
		{"/orders", func() error {
			_, err := orders.Orders(context.Background())
			return err
		}},
	}
	for round := 0; round < 2; round++ {
		for _, screen := range screens {
			if err := screen.open(); err != nil {
				t.Fatalf("opening %s (round %d): %v", screen.name, round+1, err)
			}
		}
	}
	if got := resolutions(); got != 1 {
		t.Errorf("opening every console read screen resolved the account %d times, want 1. "+
			"Each seam resolving for itself is one /api/v1/accounts read per screen, and that "+
			"read is the one that came back 429 three times on 2026-07-26 (measurements.md M4)", got)
	}

	// And two screens opened in the same second are still one resolution: the
	// build is serialised rather than raced, which is the case a warm cache hides.
	before := resolutions()
	cold := newConsoleBroker(&rootOptions{})
	coldPositions, coldOrders := newConsoleHoldings(cold), consoleOrdersSeam(cold)
	var wg sync.WaitGroup
	errs := make([]error, len(screens))
	for i := range screens {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			switch i {
			case 0:
				_, errs[i] = coldPositions.Holdings(context.Background(), "")
			default:
				_, errs[i] = coldOrders.Orders(context.Background())
			}
		}(i)
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Fatalf("opening %s concurrently: %v", screens[i].name, err)
		}
	}
	if got := resolutions() - before; got != 1 {
		t.Errorf("two screens opened at once resolved the account %d times, want 1; the build "+
			"has to be serialised, not raced", got)
	}

	// --- the wiring half -----------------------------------------------------
	//
	// The runtime half above built the seams the way runConsole does. This is what
	// keeps that true, and what a future seam has to pass.
	file, err := parser.ParseFile(token.NewFileSet(), "console.go", readSource(t, "console.go"), 0)
	if err != nil {
		t.Fatalf("parsing console.go: %v", err)
	}

	// Nothing but the argued-for sites builds a client.
	sites := map[string]int{}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		name := fn.Name.Name
		if fn.Recv != nil && len(fn.Recv.List) > 0 {
			receiver := strings.TrimPrefix(types.ExprString(fn.Recv.List[0].Type), "*")
			name = receiver + "." + name
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "verifyBrokerFactory" {
				sites[name]++
			}
			return true
		})
	}
	for name := range sites {
		if why, argued := consoleBrokerBuildSites[name]; !argued || strings.TrimSpace(why) == "" {
			t.Errorf("%s builds a live client of its own and is not argued for in "+
				"consoleBrokerBuildSites; every separate build is one more /api/v1/accounts "+
				"read per console process", name)
		}
	}
	for name := range consoleBrokerBuildSites {
		if sites[name] == 0 {
			t.Errorf("consoleBrokerBuildSites still argues for %s, which no longer builds a "+
				"client; an exemption nobody uses is one the next reader will reuse", name)
		}
	}

	// And runConsole builds ONE resolver and hands it to every read seam.
	var (
		holder    string
		builds    int
		seamArgs  = map[string]string{}
		seamNames = []string{"newConsoleHoldings", "consoleOrdersSeam"}
	)
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "runConsole" {
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.AssignStmt:
				for i, rhs := range node.Rhs {
					call, ok := rhs.(*ast.CallExpr)
					if !ok {
						continue
					}
					ident, ok := call.Fun.(*ast.Ident)
					if !ok || ident.Name != "newConsoleBroker" || i >= len(node.Lhs) {
						continue
					}
					builds++
					if target, ok := node.Lhs[i].(*ast.Ident); ok {
						holder = target.Name
					}
				}
			case *ast.CallExpr:
				ident, ok := node.Fun.(*ast.Ident)
				if !ok || len(node.Args) != 1 {
					return true
				}
				for _, seam := range seamNames {
					if ident.Name == seam {
						seamArgs[seam] = types.ExprString(node.Args[0])
					}
				}
			}
			return true
		})
	}
	if builds != 1 {
		t.Fatalf("runConsole builds %d shared resolvers, want exactly 1; a second one is a "+
			"second account resolution wearing the shared resolver's name", builds)
	}
	for _, seam := range seamNames {
		got, wired := seamArgs[seam]
		if !wired {
			t.Errorf("runConsole no longer wires %s; the guard is not reading the console's "+
				"read seams", seam)
			continue
		}
		if got != holder {
			t.Errorf("runConsole hands %s %q while the shared resolver is %q; a seam with its "+
				"own resolver resolves the account a second time", seam, got, holder)
		}
	}
}
