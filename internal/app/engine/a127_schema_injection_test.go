package engine

// a127 D2 · S7 · S10 — 엔진의 두 적재기 호출 자리는 원장 스키마 주입 값으로 `journal.SchemaVersion` **상수 선택자**를 넘긴다. 런타임에 파일에서 읽은
// 값(같은 파일을 자기와 비교하는 공허한 검사)이나 리터럴(다음 마이그레이션에서 같은 결함 재현)은 값 비교로는 오늘 같아서 못 잡으므로 구조로 잰다.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"testing"
)

const a127JournalImport = "github.com/JungHoonGhae/tossinvest-cli/internal/journal"

func TestA127EngineInjectsTheJournalSchemaVersionConstant(t *testing.T) {
	for _, site := range []struct{ file, pkg, typ string }{
		{"strategy_risk_authority.go", "github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket", "ProductionRiskSnapshotConfig"},
		{"strategy_route_authority.go", "github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter", "ProductionRouteConfig"},
	} {
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, site.file, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		names := map[string]string{} // import path → local name
		for _, spec := range file.Imports {
			path, _ := strconv.Unquote(spec.Path.Value)
			name := path[len(path)-len(lastSegment(path)):]
			if spec.Name != nil {
				name = spec.Name.Name
			}
			names[path] = name
		}
		journalName, typeName := names[a127JournalImport], names[site.pkg]
		if journalName == "" || typeName == "" {
			t.Fatalf("%s: imports journal=%q %s=%q — the site must import both", site.file, journalName, site.typ, typeName)
		}
		literals := 0
		ast.Inspect(file, func(n ast.Node) bool {
			lit, ok := n.(*ast.CompositeLit)
			if !ok {
				return true
			}
			typ, ok := lit.Type.(*ast.SelectorExpr)
			if !ok || typ.Sel.Name != site.typ {
				return true
			}
			if x, ok := typ.X.(*ast.Ident); !ok || x.Name != typeName {
				return true
			}
			literals++
			found := false
			for _, element := range lit.Elts {
				kv, ok := element.(*ast.KeyValueExpr)
				if !ok {
					continue
				}
				if key, ok := kv.Key.(*ast.Ident); !ok || key.Name != "JournalSchemaVersion" {
					continue
				}
				found = true
				value, ok := kv.Value.(*ast.SelectorExpr)
				if !ok || value.Sel.Name != "SchemaVersion" {
					t.Errorf("%s: JournalSchemaVersion is %T, want the selector %s.SchemaVersion", fset.Position(kv.Pos()), kv.Value, journalName)
					continue
				}
				if x, ok := value.X.(*ast.Ident); !ok || x.Name != journalName {
					t.Errorf("%s: JournalSchemaVersion selects from %v, want the journal import %q", fset.Position(kv.Pos()), value.X, journalName)
				}
			}
			if !found {
				t.Errorf("%s: %s literal without JournalSchemaVersion — the loader would refuse an uninjected config", fset.Position(lit.Pos()), site.typ)
			}
			return true
		})
		if literals != 1 {
			t.Fatalf("%s: %d %s literals, want exactly the one production call site", site.file, literals, site.typ)
		}
	}
}

func lastSegment(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
}
