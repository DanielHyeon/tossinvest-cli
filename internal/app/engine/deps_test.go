package engine_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const modulePath = "github.com/JungHoonGhae/tossinvest-cli/"

// TestEngineDependencyGraphExcludesWTSMutators is a structural safety test, not
// a behavioural one: it proves the engine cannot reach a web-session order
// mutator even by mistake, because no such type is in its import graph.
//
// A runtime test can only show that a particular code path did not call WTS on
// the day it was written. This one shows the call is unspellable: internal/client
// (whose *Client implements PlacePendingOrder/CancelPendingOrder/
// AmendPendingOrder against the scraped web session) and internal/hybrid (whose
// broker silently routes writes to WTS whenever the official client is absent)
// are absent from everything the engine transitively imports.
//
// Test files are deliberately not walked. What matters is what the shipped
// engine can call; the isolation test next door legitimately constructs the WTS
// broker to prove its spy is live.
func TestEngineDependencyGraphExcludesWTSMutators(t *testing.T) {
	t.Parallel()

	deps := transitiveInternalDeps(t, modulePath+"internal/app/engine")

	forbidden := map[string]string{
		modulePath + "internal/client": "WTS order mutators (PlacePendingOrder/CancelPendingOrder/AmendPendingOrder over the scraped web session)",
		modulePath + "internal/hybrid": "hybrid router — falls back to WTS for order writes whenever the official client is nil",
		modulePath + "internal/app":    "CLI wiring profile — transitively pulls in both of the above",
	}
	for pkg, why := range forbidden {
		if deps[pkg] {
			t.Errorf("engine dependency graph must not contain %s: %s\nfull graph: %v", pkg, why, sortedKeys(deps))
		}
	}

	// Positive control: an empty or truncated walk would make the assertions
	// above pass for the wrong reason.
	for _, required := range []string{
		modulePath + "internal/official",
		modulePath + "internal/trading",
		modulePath + "internal/orderintent",
		modulePath + "internal/config",
	} {
		if !deps[required] {
			t.Errorf("expected %s in the engine dependency graph; walker found only %v", required, sortedKeys(deps))
		}
	}
}

// transitiveInternalDeps returns every in-module package reachable from pkg by
// production (non-test) imports. It parses imports directly rather than shelling
// out to `go list` so the test needs no toolchain subprocess and no module
// downloads.
func transitiveInternalDeps(t *testing.T, pkg string) map[string]bool {
	t.Helper()
	root := repoRoot(t)

	seen := map[string]bool{}
	var walk func(string)
	walk = func(p string) {
		if seen[p] {
			return
		}
		seen[p] = true

		dir := filepath.Join(root, filepath.FromSlash(strings.TrimPrefix(p, modulePath)))
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("reading package dir %s: %v", dir, err)
		}
		fset := token.NewFileSet()
		for _, e := range entries {
			name := e.Name()
			if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
				continue
			}
			f, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, parser.ImportsOnly)
			if err != nil {
				t.Fatalf("parsing %s: %v", name, err)
			}
			for _, imp := range f.Imports {
				path := strings.Trim(imp.Path.Value, `"`)
				if strings.HasPrefix(path, modulePath) {
					walk(path)
				}
			}
		}
	}
	walk(pkg)
	delete(seen, pkg) // the package itself is not one of its dependencies
	return seen
}

// repoRoot walks up from the test's working directory (the package dir) to the
// module root.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not locate go.mod above the test directory")
		}
		dir = parent
	}
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, strings.TrimPrefix(k, modulePath))
	}
	// insertion sort keeps the helper dependency-free and the list is tiny
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j] < out[j-1]; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}
