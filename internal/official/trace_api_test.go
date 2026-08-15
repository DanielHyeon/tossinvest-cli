package official

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"testing"
)

// TestAttemptTracePublicSurfaceIsReadOnly pins the sole exported M0 transport
// seam. It is context observation only: no request constructor, broker mutation,
// or approval capability may be added here.
func TestAttemptTracePublicSurfaceIsReadOnly(t *testing.T) {
	src, err := os.ReadFile("trace.go")
	if err != nil {
		t.Fatal(err)
	}
	f, err := parser.ParseFile(token.NewFileSet(), "trace.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Recv == nil && d.Name.IsExported() {
				got[d.Name.Name] = true
			}
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				if ts, ok := spec.(*ast.TypeSpec); ok && ts.Name.IsExported() {
					got[ts.Name.Name] = true
				}
			}
		}
	}
	want := map[string]bool{"AttemptTrace": true, "AttemptObserver": true, "WithAttemptObserver": true}
	if len(got) != len(want) {
		t.Fatalf("trace public surface = %#v, want %#v", got, want)
	}
	for name := range want {
		if !got[name] {
			t.Fatalf("trace public surface missing %s: %#v", name, got)
		}
	}
}
