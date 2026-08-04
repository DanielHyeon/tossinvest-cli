package engine_test

// entryreach_test.go is design D5: the observation this change rests on, frozen
// as a test.
//
// interlock-gates-entry-not-exit argues that refusing the runtime over clause 6
// costs protection and buys nothing, and the "buys nothing" half is an
// observation about today's code — a started engine has no reachable path that
// raises exposure. Observations expire. This one expires the moment 2c-A wires
// candidate discovery to entry issuance, which is exactly when somebody needs to
// be told that the argument no longer holds.
//
// It is a structural test in the sense deps_test.go means: not "this path did not
// place a buy on the day the test ran", but "the call is not spelled anywhere the
// runtime can reach".
//
// Going red is not a bug in this test. It is the notice that the runtime gained
// an entry path, and the thing to do about it is to satisfy the gateway's
// protection refusal (execgw's protection_test.go) — not to delete this file.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// tracerSymbols are the only entry-issuing constructs in internal/app/engine.
//
// The tracer is a deliberate island (add-core-domain D8: 실전 실행은 verify
// 트랙). It contains the package's single `Side: "buy"` and its single
// IssueEntry call, and nothing in the shipped command graph names it.
var tracerSymbols = []string{"NewTracer", "TracerOptions", "TracerParams"}

// TestTheRuntimeCannotReachEntryIssuance walks the non-test sources of the engine
// package and the command package, and asserts that the tracer is unreferenced
// outside its own file.
func TestTheRuntimeCannotReachEntryIssuance(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	for _, dir := range []string{
		filepath.Join(root, "internal", "app", "engine"),
		filepath.Join(root, "cmd", "tossctl"),
	} {
		for _, file := range goSources(t, dir) {
			if filepath.Base(file) == "tracer.go" {
				continue
			}
			for _, sym := range referencedIdents(t, file, tracerSymbols) {
				t.Errorf("%s references %s: the runtime can now reach entry issuance, "+
					"so interlock-gates-entry-not-exit's premise no longer holds. The entry must be "+
					"refused at the gateway's protection check, not permitted by this test's absence",
					mustRel(root, file), sym)
			}
		}
	}
}

// TestTheEngineSpellsExactlyOneBuy is the historical test name; its current
// contract pins the reviewed BUY paths: the package
// contains the legacy unreachable tracer plus the claimed-lease strategy path.
//
// A second one appearing anywhere is the same notice as above, caught even if the
// new caller reaches execgw directly instead of going through the tracer.
func TestTheEngineSpellsExactlyOneBuy(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	var found []string
	for _, file := range goSources(t, filepath.Join(root, "internal", "app", "engine")) {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("reading %s: %v", file, err)
		}
		for i, line := range strings.Split(string(src), "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "//") {
				continue
			}
			if strings.Contains(trimmed, `Side: "buy"`) || strings.Contains(trimmed, `Side: "BUY"`) {
				found = append(found, mustRel(root, file)+":"+strconv.Itoa(i+1))
			}
		}
	}
	want := []string{"strategy_dispatch_cycle.go", "tracer.go"}
	if len(found) != len(want) {
		t.Fatalf("buy-side intents = %v, want exactly reviewed paths %v", found, want)
	}
	for i, file := range want {
		if !strings.Contains(found[i], file) {
			t.Errorf("buy-side intents = %v, want index %d in %s", found, i, file)
		}
	}
}

// --- helpers ----------------------------------------------------------------

func goSources(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading %s: %v", dir, err)
	}
	var out []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		out = append(out, filepath.Join(dir, name))
	}
	return out
}

// referencedIdents returns which of `want` the file names, ignoring comments.
func referencedIdents(t *testing.T, file string, want []string) []string {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, nil, 0) // no comments
	if err != nil {
		t.Fatalf("parsing %s: %v", file, err)
	}
	wanted := map[string]bool{}
	for _, w := range want {
		wanted[w] = true
	}
	seen := map[string]bool{}
	var out []string
	ast.Inspect(f, func(n ast.Node) bool {
		id, ok := n.(*ast.Ident)
		if !ok || !wanted[id.Name] || seen[id.Name] {
			return true
		}
		seen[id.Name] = true
		out = append(out, id.Name)
		return true
	})
	return out
}

func mustRel(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return rel
}
