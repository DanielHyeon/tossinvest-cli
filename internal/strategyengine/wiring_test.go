package strategyengine

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
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
		"internal/strategyobserve/observe.go":   true,
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
		file, parseErr := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if parseErr != nil {
			return parseErr
		}
		for _, spec := range file.Imports {
			matched, err := importMatches(spec, modulePath)
			if err != nil {
				return err
			}
			if matched && !strings.HasPrefix(filepath.ToSlash(relative), "internal/strategyengine/") {
				rel := filepath.ToSlash(relative)
				if !allowedDormantContracts[rel] {
					t.Errorf("a047 dormant engine is runtime-wired by %s", rel)
				}
				if rel == "internal/strategyobserve/observe.go" {
					if err := validateObservationEngineReferences(file, spec); err != nil {
						t.Errorf("%s: %v", rel, err)
					}
				}
				if rel == "internal/console/strategy_runtime.go" || rel == "internal/httpapi/read.go" {
					assertConsoleOnlyCallsDormantDescriptor(t, file)
				}
			}
		}
		if err := validateObservationImports(file, filepath.ToSlash(relative)); err != nil {
			t.Error(err)
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

func validateObservationImports(file *ast.File, path string) error {
	for _, spec := range file.Imports {
		matched, err := importMatches(spec, "github.com/JungHoonGhae/tossinvest-cli/internal/strategyobserve")
		if err != nil {
			return err
		}
		if matched && path != "tools/a155-observe-versioned-strategy-decisions/main.go" && filepath.ToSlash(filepath.Dir(path)) != "internal/strategyobserve" {
			return fmt.Errorf("observation package imported by runtime: %s", path)
		}
	}
	return nil
}

func validateObservationEngineReferences(file *ast.File, spec *ast.ImportSpec) error {
	alias := "strategyengine"
	if spec.Name != nil {
		alias = spec.Name.Name
	}
	if alias == "." || alias == "_" {
		return fmt.Errorf("observation engine import must be named")
	}
	allowed := map[string]bool{"ParkerRRFromDecision": true, "ParkerRRDiagnostics": true, "DecisionRecord": true, "LaneID": true, "LaneVersion": true}
	var violation error
	ast.Inspect(file, func(node ast.Node) bool {
		sel, ok := node.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		pkg, ok := sel.X.(*ast.Ident)
		if ok && pkg.Name == alias && !allowed[sel.Sel.Name] {
			violation = fmt.Errorf("forbidden engine reference %s.%s", alias, sel.Sel.Name)
		}
		return true
	})
	return violation
}

func importMatches(spec *ast.ImportSpec, path string) (bool, error) {
	imported, err := strconv.Unquote(spec.Path.Value)
	return imported == path, err
}
