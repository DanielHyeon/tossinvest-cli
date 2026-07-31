package markout

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var productionImportAllowlist = map[string]bool{
	"math/big": true,
	"regexp":   true,
	"sort":     true,
	"strings":  true,
	"time":     true,
}

func TestMarkoutProductionUsesOnlyPureImportsAndNoClockOrPolling(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		source, err := os.ReadFile(entry.Name())
		if err != nil {
			t.Fatal(err)
		}
		if findings := productionIsolationFindings(entry.Name(), source); len(findings) != 0 {
			t.Errorf("%s violates markout supplied-observation isolation:\n  %s",
				entry.Name(), strings.Join(findings, "\n  "))
		}
	}
}

func TestProductionIsolationDetectorPositiveControl(t *testing.T) {
	source := []byte(`package markout
import (
  clock "time"
  "net/http"
)
func pollQuotes() {
  _ = clock.Now()
  _ = clock.NewTicker(clock.Second)
  _, _ = http.Get("https://example.invalid")
}`)
	findings := strings.Join(productionIsolationFindings("positive_control.go", source), "\n")
	for _, want := range []string{`import "net/http" is not allowlisted`, "clock source time.Now", "polling primitive time.NewTicker", "poll-like function pollQuotes"} {
		if !strings.Contains(findings, want) {
			t.Errorf("positive control did not detect %q; findings:\n%s", want, findings)
		}
	}
}

func productionIsolationFindings(filename string, source []byte) []string {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, source, 0)
	if err != nil {
		return []string{fmt.Sprintf("parse: %v", err)}
	}
	imports := map[string]string{}
	var findings []string
	for _, spec := range file.Imports {
		path := strings.Trim(spec.Path.Value, `"`)
		if !productionImportAllowlist[path] {
			findings = append(findings, fmt.Sprintf("import %q is not allowlisted", path))
		}
		name := filepath.Base(path)
		if spec.Name != nil {
			name = spec.Name.Name
		}
		imports[name] = path
	}
	ast.Inspect(file, func(node ast.Node) bool {
		switch value := node.(type) {
		case *ast.FuncDecl:
			if strings.Contains(strings.ToLower(value.Name.Name), "poll") {
				findings = append(findings, "poll-like function "+value.Name.Name)
			}
		case *ast.CallExpr:
			selector, ok := value.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			identifier, ok := selector.X.(*ast.Ident)
			if !ok || imports[identifier.Name] != "time" {
				return true
			}
			switch selector.Sel.Name {
			case "Now", "Since", "Until":
				findings = append(findings, "clock source time."+selector.Sel.Name)
			case "NewTicker", "Tick", "After", "AfterFunc", "NewTimer", "Sleep":
				findings = append(findings, "polling primitive time."+selector.Sel.Name)
			}
		}
		return true
	})
	return findings
}
