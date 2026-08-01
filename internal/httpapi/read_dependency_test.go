package httpapi

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestReadPackageOwnsNoBrokerJournalWriterOrTradingCapability(t *testing.T) {
	forbidden := []string{
		"/internal/official", "/internal/journal", "/internal/trading", "/internal/execgw", "/internal/console",
	}
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	fileset := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fileset, filepath.Clean(entry.Name()), nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imported := range file.Imports {
			path, err := strconv.Unquote(imported.Path.Value)
			if err != nil {
				t.Fatal(err)
			}
			for _, banned := range forbidden {
				if strings.Contains(path, banned) {
					t.Errorf("%s imports forbidden capability package %s", entry.Name(), path)
				}
			}
		}
	}
}

func TestReaderSurfaceIsExactlySevenReads(t *testing.T) {
	typeOf := reflect.TypeOf((*Reader)(nil)).Elem()
	want := []string{"Candidates", "Engine", "Optimization", "Orders", "Performance", "Positions", "Settings"}
	if typeOf.NumMethod() != len(want) {
		t.Fatalf("Reader has %d methods, want %d", typeOf.NumMethod(), len(want))
	}
	for _, name := range want {
		method, ok := typeOf.MethodByName(name)
		if !ok {
			t.Errorf("Reader lacks %s", name)
			continue
		}
		if strings.Contains(strings.ToLower(method.Name), "write") || strings.Contains(strings.ToLower(method.Name), "place") {
			t.Errorf("Reader exposes mutation-like method %s", method.Name)
		}
	}
}

func TestPublicRowsCannotAcceptARawAccountReference(t *testing.T) {
	for _, value := range []any{Position{}, Order{}} {
		typeOf := reflect.TypeOf(value)
		if _, exists := typeOf.FieldByName("AccountRef"); exists {
			t.Errorf("%s exposes a raw AccountRef field", typeOf.Name())
		}
		field, exists := typeOf.FieldByName("AccountLabel")
		if !exists || field.Tag.Get("json") != "accountLabel" {
			t.Errorf("%s lacks the masked accountLabel contract", typeOf.Name())
		}
	}
}

func TestReadRouterContainsNoRequestBodyDecoder(t *testing.T) {
	fileset := token.NewFileSet()
	file, err := parser.ParseFile(fileset, "router.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	ast.Inspect(file, func(node ast.Node) bool {
		selector, ok := node.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if selector.Sel.Name == "Decode" || selector.Sel.Name == "ParseForm" || selector.Sel.Name == "FormValue" {
			t.Errorf("read router contains request input decoder %s", selector.Sel.Name)
		}
		return true
	})
}
