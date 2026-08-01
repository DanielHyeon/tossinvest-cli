package strategyengine

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestA047ShipsNoRuntimeOrderOrExitWiring(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", ".."))
	modulePath := "github.com/JungHoonGhae/tossinvest-cli/internal/strategyengine"
	allowedDormantContracts := map[string]bool{
		"internal/execgw/riskguardian.go":       true,
		"internal/strategydispatch/adapters.go": true,
		"internal/strategydispatch/dispatch.go": true,
		"internal/console/strategy_runtime.go":  true,
		"internal/httpapi/read.go":              true,
	}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == ".git" || entry.Name() == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		relative, _ := filepath.Rel(root, path)
		if strings.HasPrefix(filepath.ToSlash(relative), "internal/strategyengine/") {
			return nil
		}
		file, parseErr := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if parseErr != nil {
			return parseErr
		}
		for _, spec := range file.Imports {
			if strings.Trim(spec.Path.Value, "\"") == modulePath {
				rel := filepath.ToSlash(relative)
				if !allowedDormantContracts[rel] {
					t.Errorf("a047 dormant engine is runtime-wired by %s", rel)
				}
				if rel == "internal/console/strategy_runtime.go" || rel == "internal/httpapi/read.go" {
					assertConsoleOnlyCallsDormantDescriptor(t, file)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func assertConsoleOnlyCallsDormantDescriptor(t *testing.T, file *ast.File) {
	t.Helper()
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		pkg, ok := selector.X.(*ast.Ident)
		if !ok || pkg.Name != "strategyengine" {
			return true
		}
		if selector.Sel.Name != "DormantRuntimeDescriptor" {
			t.Errorf("console calls strategyengine.%s; only the dormant read-only descriptor is allowed", selector.Sel.Name)
		}
		return true
	})
}
