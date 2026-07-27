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
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/console"
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
	var reader console.HoldingsReader = newConsoleHoldings(&rootOptions{})

	typ := reflect.TypeOf(reader)
	if typ.NumMethod() != 1 || typ.Method(0).Name != "Holdings" {
		names := make([]string, 0, typ.NumMethod())
		for i := 0; i < typ.NumMethod(); i++ {
			names = append(names, typ.Method(i).Name)
		}
		t.Fatalf("the console's holdings reader declares %v, want exactly [Holdings]", names)
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
	for name := range fields {
		lowered := strings.ToLower(name)
		for _, banned := range []string{"broker", "client", "order", "place", "cancel"} {
			if strings.Contains(lowered, banned) {
				t.Errorf("console.Options is given a %s field; the dashboard's whole broker is "+
					"console.HoldingsReader", name)
			}
		}
	}
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
