package markout

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

func TestMarkoutHasNoPollingTransportOrClockSource(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), entry.Name(), nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, spec := range file.Imports {
			path := strings.Trim(spec.Path.Value, `"`)
			if path == "net" || path == "net/http" || path == "database/sql" || strings.Contains(path, "/internal/") {
				t.Errorf("%s imports %q; markout may consume supplied observations only", entry.Name(), path)
			}
		}
	}
	source, err := os.ReadFile("markout.go")
	if err != nil {
		t.Fatal(err)
	}
	file, err := parser.ParseFile(token.NewFileSet(), "markout.go", source, parser.SkipObjectResolution)
	if err != nil {
		t.Fatal(err)
	}
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if id, ok := selector.X.(*ast.Ident); ok && id.Name == "time" && selector.Sel.Name == "Now" {
			t.Error("markout reads time.Now; the decision instant and observations must be supplied")
		}
		return true
	})
}
