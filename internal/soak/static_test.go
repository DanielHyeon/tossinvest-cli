package soak_test

// static_test.go is the structural half of "the soak tool contains no mutation
// transport" (execution-verification: "soak 도구는 조회 전용이며 mutation transport를
// 포함해서는 안 된다").
//
// It follows internal/testenv/static_test.go's principle exactly: a property that
// no runtime test can observe — because the offending code would simply never be
// executed by a test — is asserted against the source instead.
//
// Two assertions, at two levels:
//
//	the package     internal/soak's transitive import graph contains no package
//	                that can place, amend or cancel an order. It reaches the
//	                broker only through soak.Reads, an interface whose method set
//	                has no mutation on it, so this is a compile-time guarantee
//	                rather than a convention.
//	the wiring      cmd/tossctl/soak.go — the one file that does hold a real
//	                broker client — names no mutating method and imports no
//	                mutating package.
//
// The runtime counterpart (an httptest server that fails the test on any request
// that is not a GET) is TestSoakIssuesNoMutatingRequest in cmd/tossctl.

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const modulePath = "github.com/JungHoonGhae/tossinvest-cli/"

// mutationPackages are the packages that can change a live account, plus the
// ones that exist only to drive them.
var mutationPackages = []string{
	"internal/execgw",      // the order gateway: submit, cancel, amend
	"internal/flatten",     // the liquidation saga
	"internal/official",    // PlaceOrder / CancelOrder / ModifyOrder live here
	"internal/orderintent", // the intent types a mutation is built from
	"internal/trading",     // the write-side broker service
	"internal/app/engine",  // wires all of the above together
	"internal/hybrid",      // routes writes to a broker
	"internal/client",      // the web-session client
}

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

// internalImports returns the module-internal packages pkgDir imports, from its
// non-test files only. Test files are excluded on purpose: what ships in the
// binary is the thing being constrained.
func internalImports(t *testing.T, root, pkgDir string) []string {
	t.Helper()
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, pkgDir, func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parsing %s: %v", pkgDir, err)
	}

	var out []string
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, spec := range file.Imports {
				path, err := strconv.Unquote(spec.Path.Value)
				if err != nil {
					t.Fatalf("unquoting an import in %s: %v", pkgDir, err)
				}
				if strings.HasPrefix(path, modulePath) {
					out = append(out, strings.TrimPrefix(path, modulePath))
				}
			}
		}
	}
	return out
}

// transitiveInternalImports walks the module-internal import graph from start.
func transitiveInternalImports(t *testing.T, root, start string) map[string][]string {
	t.Helper()
	// value is the path that pulled the key in, so a failure can name the chain.
	seen := map[string][]string{}
	var walk func(pkg string, chain []string)
	walk = func(pkg string, chain []string) {
		if _, done := seen[pkg]; done {
			return
		}
		seen[pkg] = chain
		for _, dep := range internalImports(t, root, filepath.Join(root, filepath.FromSlash(pkg))) {
			walk(dep, append(append([]string(nil), chain...), pkg))
		}
	}
	walk(start, nil)
	delete(seen, start)
	return seen
}

// TestSoakHasNoMutationPackageInItsImportGraph.
func TestSoakHasNoMutationPackageInItsImportGraph(t *testing.T) {
	root := repoRoot(t)
	graph := transitiveInternalImports(t, root, "internal/soak")

	if len(graph) == 0 {
		t.Fatal("the import walk found nothing; it is not looking at the package")
	}
	for _, banned := range mutationPackages {
		if chain, found := graph[banned]; found {
			t.Errorf("internal/soak depends on %s (via %s). The soak is read-only: it reaches the "+
				"broker through soak.Reads, whose method set has no mutation on it",
				banned, strings.Join(append(chain, banned), " -> "))
		}
	}
}

// TestSoakWiringNamesNoMutation checks the file that does hold a real broker
// client. The client type carries the order-mutating methods; the adapter must
// be built over a read-only interface and must never call one.
//
// A mention in a comment is allowed and a call is not, exactly as
// internal/testenv/static_test.go treats FixedFSProber: the file has to be able
// to explain what it is avoiding and why.
func TestSoakWiringNamesNoMutation(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "cmd", "tossctl", "soak.go")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	body := string(data)

	mutations := []string{
		"PlaceOrder", "CancelOrder", "ModifyOrder",
		"CreateConditionalOrder", "CancelConditionalOrder", "ModifyConditionalOrder",
	}
	for _, name := range mutations {
		if strings.Contains(body, name) && !onlyInComments(body, name) {
			t.Errorf("cmd/tossctl/soak.go calls %s outside a comment. The soak wiring must not be able "+
				"to reach a mutation", name)
		}
	}

	for _, banned := range []string{
		"internal/execgw", "internal/flatten", "internal/app/engine",
		"internal/orderintent", "internal/trading", "internal/hybrid",
	} {
		if strings.Contains(body, `"`+modulePath+banned+`"`) {
			t.Errorf("cmd/tossctl/soak.go imports %s; the soak wiring reaches reads only", banned)
		}
	}
}

// onlyInComments reports whether every occurrence of needle sits on a line that
// starts, after indentation, with a comment marker.
//
// It does not parse Go, deliberately: a false "not a comment" makes the guard
// stricter, which is the safe direction for a check like this.
func onlyInComments(src, needle string) bool {
	for _, line := range strings.Split(src, "\n") {
		if !strings.Contains(line, needle) {
			continue
		}
		if !strings.HasPrefix(strings.TrimSpace(line), "//") {
			return false
		}
	}
	return true
}
