package strategymarket

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const strategyMarketImportPath = "github.com/JungHoonGhae/tossinvest-cli/internal/strategymarket"

func scalarSealerReferences(path string, source any) ([]string, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, source, 0)
	if err != nil {
		return nil, err
	}
	parents := map[ast.Node]ast.Node{}
	var stack []ast.Node
	ast.Inspect(file, func(node ast.Node) bool {
		if node == nil {
			stack = stack[:len(stack)-1]
			return false
		}
		if len(stack) != 0 {
			parents[node] = stack[len(stack)-1]
		}
		stack = append(stack, node)
		return true
	})
	var findings []string
	for _, spec := range file.Imports {
		importPath, _ := strconv.Unquote(spec.Path.Value)
		if importPath == strategyMarketImportPath && spec.Name != nil && spec.Name.Name == "." {
			findings = append(findings, "dot-import:"+path)
		}
	}
	ast.Inspect(file, func(node ast.Node) bool {
		identifier, ok := node.(*ast.Ident)
		if !ok || identifier.Name != "SealAdaptedOfficialMinutePage" {
			return true
		}
		if declaration, ok := parents[identifier].(*ast.FuncDecl); ok && declaration.Name == identifier {
			return true
		}
		findings = append(findings, "reference:"+path)
		return true
	})
	return findings, nil
}

func TestOfficialPageScalarSealerIsUsedOnlyByStrategyCandleAdapter(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", ".."))
	var references []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil || entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return walkErr
		}
		found, inspectErr := scalarSealerReferences(path, nil)
		if inspectErr != nil {
			return inspectErr
		}
		references = append(references, found...)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(references) != 1 || !strings.HasSuffix(references[0], "internal/strategycandle/adapter.go") {
		t.Fatalf("official page scalar sealer production references = %v, want only strategycandle adapter", references)
	}
}

func TestOfficialPageScalarSealerGuardDetectsFunctionAliasesAndDotImports(t *testing.T) {
	alias := `package bad
import sm "github.com/JungHoonGhae/tossinvest-cli/internal/strategymarket"
var forge = sm.SealAdaptedOfficialMinutePage
`
	dot := `package bad
import . "github.com/JungHoonGhae/tossinvest-cli/internal/strategymarket"
var forge = SealAdaptedOfficialMinutePage
`
	samePackage := `package strategymarket
var forge = SealAdaptedOfficialMinutePage
`
	for name, source := range map[string]string{"function alias": alias, "dot import": dot, "same-package alias": samePackage} {
		t.Run(name, func(t *testing.T) {
			findings, err := scalarSealerReferences(name+".go", source)
			if err != nil || len(findings) == 0 {
				t.Fatalf("findings=%v err=%v", findings, err)
			}
		})
	}
}
