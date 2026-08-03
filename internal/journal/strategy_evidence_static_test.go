package journal

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
	"testing"
)

func TestStrategyEvidenceReadBoundaryIsStructurallySelectOnly(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "strategy_evidence.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, imported := range file.Imports {
		path, err := strconv.Unquote(imported.Path.Value)
		if err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range []string{"net/http", "/broker", "/dispatch", "/execgw", "/guardian", "/operating", "/runtime", "/toggle"} {
			if strings.Contains(path, forbidden) {
				t.Fatalf("dormant journal boundary imports mutation/runtime path %q", path)
			}
		}
	}
	method := findMethod(t, file, "ConsumedSnapshot")
	ast.Inspect(method.Body, func(node ast.Node) bool {
		switch value := node.(type) {
		case *ast.SelectorExpr:
			for _, forbidden := range []string{"Exec", "ExecContext", "Begin", "BeginTx"} {
				if value.Sel.Name == forbidden {
					t.Errorf("ConsumedSnapshot contains mutating database call %s", forbidden)
				}
			}
		case *ast.BasicLit:
			if value.Kind != token.STRING {
				break
			}
			literal, unquoteErr := strconv.Unquote(value.Value)
			if unquoteErr != nil {
				t.Errorf("unquote SQL literal: %v", unquoteErr)
				break
			}
			upper := strings.ToUpper(literal)
			for _, forbidden := range []string{"INSERT ", "UPDATE ", "DELETE ", "ALTER ", "DROP ", "PRAGMA "} {
				if strings.Contains(upper, forbidden) {
					t.Errorf("ConsumedSnapshot contains mutating SQL %q", forbidden)
				}
			}
		}
		return true
	})
}

func findMethod(t *testing.T, file *ast.File, name string) *ast.FuncDecl {
	t.Helper()
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if ok && function.Recv != nil && function.Name.Name == name {
			return function
		}
	}
	t.Fatalf("method %s not found", name)
	return nil
}
