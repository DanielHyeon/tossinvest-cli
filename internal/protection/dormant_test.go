package protection_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
)

func TestProtectionRemainsUnwiredAndCorePackageHasNoBrokerTransport(t *testing.T) {
	if execgw.ProfileProtection != execgw.ProtectionUnwired {
		t.Fatalf("ProfileProtection = %q, want UNWIRED", execgw.ProfileProtection)
	}

	root := moduleRoot(t)
	packageDir := filepath.Join(root, "internal", "protection")
	entries, err := os.ReadDir(packageDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(packageDir, entry.Name())
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imp := range file.Imports {
			name := strings.Trim(imp.Path.Value, `"`)
			if name == "net/http" || strings.Contains(name, "/internal/official") || strings.Contains(name, "/internal/trading") {
				t.Errorf("%s imports mutation-capable package %s", entry.Name(), name)
			}
		}
	}

	// No shipped engine or command may activate the otherwise testable official
	// adapter. The controller's opaque activation has no exported minter.
	for _, dir := range []string{filepath.Join(root, "cmd"), filepath.Join(root, "internal", "app")} {
		err := filepath.WalkDir(dir, func(path string, d os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			file, parseErr := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
			if parseErr != nil {
				return parseErr
			}
			for _, imp := range file.Imports {
				if strings.Contains(strings.Trim(imp.Path.Value, `"`), "/internal/protection") {
					t.Errorf("%s wires the dormant protection package", path)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestNoPaperShadowOrCanaryProtectionPath(t *testing.T) {
	root := moduleRoot(t)
	packageDir := filepath.Join(root, "internal", "protection")
	err := filepath.WalkDir(packageDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		src, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		for _, forbidden := range []string{"paper", "shadow", "canary"} {
			if strings.Contains(strings.ToLower(string(src)), forbidden) {
				t.Errorf("%s contains forbidden protection path %q", path, forbidden)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestOfficialProtectionAdapterHasNoPaperShadowCanaryOrFallbackTransport(t *testing.T) {
	root := moduleRoot(t)
	packageDir := filepath.Join(root, "internal", "protectionofficial")
	err := filepath.WalkDir(packageDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		src, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		lower := strings.ToLower(string(src))
		for _, forbidden := range []string{"paper", "shadow", "canary", "/internal/wts", "/internal/hybrid", "/internal/trading"} {
			if strings.Contains(lower, forbidden) {
				t.Errorf("%s contains forbidden adapter path %q", path, forbidden)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}
