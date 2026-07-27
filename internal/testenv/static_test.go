package testenv_test

// static_test.go is the source-level half of the completion gate (task 4.6).
//
// Some safety properties cannot be observed at runtime because the code that
// would violate them is never executed in a test — a test-only escape hatch left
// in a production path does nothing until the day it does. So they are asserted
// against the source instead.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// repoRoot is this package's directory, two levels up.
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

// productionFiles walks cmd/ and internal/ and returns every .go file that is
// compiled into the binary (that is, not a _test.go file).
func productionFiles(t *testing.T) []string {
	t.Helper()
	root := repoRoot(t)

	var out []string
	for _, dir := range []string{"cmd", "internal"} {
		err := filepath.WalkDir(filepath.Join(root, dir), func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if d.Name() == "testdata" {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			out = append(out, path)
			return nil
		})
		if err != nil {
			t.Fatalf("walking %s: %v", dir, err)
		}
	}
	if len(out) < 50 {
		t.Fatalf("only %d production files found; the walk is not covering the repository", len(out))
	}
	return out
}

// TestFixedFSProberIsTestOnly.
//
// FixedFSProber makes the filesystem guard answer whatever the caller says, which
// is exactly right for a test on a tmpfs and exactly wrong anywhere else: the
// guard exists because fsync does not mean fsync on some mounts, and a production
// path that stubs it out has turned the durability contract into a comment.
//
// The definition itself is production code and is allowed; every *use* of it must
// be in a _test.go file.
//
// There are two definitions, and that is deliberate rather than sloppy.
// internal/candidate carries its own copy of the guard because importing
// internal/journal would put the order ledger's API inside the discovery
// package's type universe, which the discovery isolation requirement forbids
// (openspec candidate-discovery, "발굴은 주문 경로에 도달할 수 없어야 한다"). The
// two allowlists are held together by internal/candidate/fsguard_drift_test.go,
// which reads the ledger's source and fails when they diverge.
func TestFixedFSProberIsTestOnly(t *testing.T) {
	root := repoRoot(t)
	// The files allowed to mention it: the ones that define it. Listed
	// explicitly, so a third copy of the durability stub has to be argued for in
	// a diff to this test rather than appearing quietly beside it.
	definitions := []string{
		filepath.Join(root, "internal", "journal", "fsguard.go"),
		filepath.Join(root, "internal", "candidate", "fsguard.go"),
	}
	isDefinition := map[string]bool{}
	for _, path := range definitions {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("reading %s: %v", path, err)
		}
		// An allowlist entry earns its place by actually declaring the function.
		// Otherwise this list becomes a way to exempt a file that merely calls it,
		// which is the exact thing being prevented.
		if !strings.Contains(string(data), "func FixedFSProber(") {
			rel, _ := filepath.Rel(root, path)
			t.Fatalf("%s is allowlisted as a definition of FixedFSProber but does not declare it; "+
				"the allowlist is exempting a use", rel)
		}
		isDefinition[path] = true
	}

	// journal.go's Options doc explains why tests inject it. A comment is not a
	// use, but allowing the file wholesale would be too broad, so the doc mention
	// is matched exactly.
	docMention := filepath.Join(root, "internal", "journal", "journal.go")

	for _, path := range productionFiles(t) {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("reading %s: %v", path, err)
		}
		if !strings.Contains(string(data), "FixedFSProber") {
			continue
		}
		rel, _ := filepath.Rel(root, path)

		// A definition file is exempt for its declaration, not wholesale. The
		// exemption exists so `func FixedFSProber(...)` can be written down; a
		// *call* to it inside the same file is the same defect as a call anywhere
		// else, and is in fact the likelier one — the tempting "fix" for the
		// nil-prober branch in CheckFilesystem is to default it to a fixed prober,
		// which turns the durability guard off for every caller that omits one.
		if isDefinition[path] {
			if line, called := callsFixedFSProber(string(data)); called {
				t.Errorf("%s:%d calls FixedFSProber. Declaring it is allowed here; "+
					"calling it disables the filesystem durability guard for every caller", rel, line)
			}
			continue
		}
		if path == docMention && onlyInComments(string(data), "FixedFSProber") {
			continue
		}
		t.Errorf("%s uses FixedFSProber. It disables the filesystem durability guard and "+
			"belongs only in _test.go files", rel)
	}
}

// callsFixedFSProber reports a call site — `FixedFSProber(` not preceded by
// `func ` — outside comments, and the 1-based line it is on.
func callsFixedFSProber(src string) (int, bool) {
	for i, line := range strings.Split(src, "\n") {
		code := line
		if idx := strings.Index(code, "//"); idx >= 0 {
			code = code[:idx]
		}
		idx := strings.Index(code, "FixedFSProber(")
		if idx < 0 {
			continue
		}
		if strings.HasSuffix(strings.TrimSpace(code[:idx]), "func") {
			continue
		}
		return i + 1, true
	}
	return 0, false
}

// TestNoProductionCodeHardcodesATestTransport catches the other shape of the same
// mistake: a production file that installs a test double for the HTTP transport.
func TestNoProductionCodeHardcodesATestTransport(t *testing.T) {
	root := repoRoot(t)
	banned := []string{
		"net/http/httptest",
		"internal/testenv",
	}

	for _, path := range productionFiles(t) {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("reading %s: %v", path, err)
		}
		body := string(data)
		for _, pkg := range banned {
			if !strings.Contains(body, `"`+pkg+`"`) &&
				!strings.Contains(body, `"github.com/JungHoonGhae/tossinvest-cli/`+pkg+`"`) {
				continue
			}
			rel, _ := filepath.Rel(root, path)
			t.Errorf("%s imports %s from production code; test scaffolding must not ship in the binary", rel, pkg)
		}
	}
}

// TestTestenvItselfIsNotImportedByProductionCode. This package exists to be
// imported by tests. A production import of it would mean the binary carries a
// transport that can be told to fail, which is the opposite of the point.
func TestTestenvItselfIsNotImportedByProductionCode(t *testing.T) {
	root := repoRoot(t)
	for _, path := range productionFiles(t) {
		// This package's own non-test file obviously mentions its own name.
		if strings.HasPrefix(path, filepath.Join(root, "internal", "testenv")) {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("reading %s: %v", path, err)
		}
		if strings.Contains(string(data), "tossinvest-cli/internal/testenv") {
			rel, _ := filepath.Rel(root, path)
			t.Errorf("%s imports internal/testenv from production code", rel)
		}
	}
}

// onlyInComments reports whether every occurrence of needle in src sits on a line
// that starts (after indentation) with a comment marker.
//
// It is deliberately simple: it does not parse Go. A false "not a comment" makes
// the test stricter, which is the safe direction for a guard like this.
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
