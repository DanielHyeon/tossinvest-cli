package protectionlifecycle_test

import (
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

func TestProductionAPIExportsNoAuthorityMintingFunction(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), entry.Name(), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for name, object := range file.Scope.Objects {
			if object.Kind == 5 && token.IsExported(name) {
				t.Fatalf("production package exports authority-capable function %s", name)
			}
		}
	}
}
