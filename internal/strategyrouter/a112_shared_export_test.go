package strategyrouter

// a112 7.3.1 SHADOW(브리프 v3.3 §3) — shadow 적재기가 쓰는 공유 도우미는 **새 파일의 중립 wrapper** 로만 나간다.
//
//   - 기존 함수는 편집하지 않는다(편집 전후 본문 digest 대조는 착지 전 measurements/lot-7.3.1-shadow/pre-edit/body-digests.tsv).
//   - wrapper 는 읽기 전용이다: export 목록을 이름으로 고정하고, 각 본문이 기존 함수 하나만 부르는 모양을 AST 로 못 박는다
//     (서술자 표는 하나 — 복사 금지, 변환 순회만 허용).

import (
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"sort"
	"strings"
	"testing"
)

const sharedExportFile = "production_shared_export.go"

// 각 wrapper 가 부를 수 있는 기존 함수(정확히 하나).
var sharedExportCalls = map[string]string{
	"SharedProductionRouteOwnerUID":    "productionRouteOwnerUID",
	"SharedReadProductionRouteFile":    "readProductionRouteFile",
	"SharedProductionRouteDigest":      "productionRouteDigest",
	"SharedProductionRouteTime":        "productionRouteTime",
	"SharedProductionRouteIdentity":    "productionRouteIdentity",
	"SharedProductionRouteDigestValid": "productionRouteDigestValid",
	"SharedProductionRouteDescriptors": "productionRouteDescriptors",
}

func TestTheSharedExportFileHoldsOnlyReadOnlyWrappers(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), sharedExportFile, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var funcs, types []string
	for _, decl := range file.Decls {
		switch value := decl.(type) {
		case *ast.FuncDecl:
			funcs = append(funcs, value.Name.Name)
			want, ok := sharedExportCalls[value.Name.Name]
			if !ok || value.Recv != nil {
				t.Errorf("unexpected declaration %s in %s", value.Name.Name, sharedExportFile)
				continue
			}
			var calls []string
			ast.Inspect(value.Body, func(node ast.Node) bool {
				if call, ok := node.(*ast.CallExpr); ok {
					switch fun := call.Fun.(type) {
					case *ast.Ident:
						calls = append(calls, fun.Name)
					case *ast.SelectorExpr:
						calls = append(calls, exprName(fun))
					default:
						calls = append(calls, "<dynamic>")
					}
				}
				return true
			})
			allowed := map[string]bool{want: true, "append": true, "make": true, "len": true, "sort.Slice": true}
			found := false
			for _, call := range calls {
				found = found || call == want
				if !allowed[call] {
					t.Errorf("%s calls %s — a wrapper may only call %s (plus the conversion loop in the descriptor wrapper)", value.Name.Name, call, want)
				}
			}
			if !found {
				t.Errorf("%s does not call %s", value.Name.Name, want)
			}
		case *ast.GenDecl:
			for _, spec := range value.Specs {
				switch spec := spec.(type) {
				case *ast.TypeSpec:
					types = append(types, spec.Name.Name)
				case *ast.ValueSpec:
					t.Errorf("%s declares variables or constants (%v) — wrappers only", sharedExportFile, spec.Names)
				}
			}
		}
	}
	sort.Strings(funcs)
	want := make([]string, 0, len(sharedExportCalls))
	for name := range sharedExportCalls {
		want = append(want, name)
	}
	sort.Strings(want)
	if !reflect.DeepEqual(funcs, want) || strings.Join(types, ",") != "SharedLaneDescriptor" {
		t.Fatalf("exports funcs=%v types=%v, want funcs=%v types=[SharedLaneDescriptor]", funcs, types, want)
	}
	for _, spec := range file.Imports {
		if path := strings.Trim(spec.Path.Value, `"`); path != "os" && path != "sort" && path != "time" {
			t.Errorf("%s imports %s — read-only wrappers need only os (FileMode) · sort · time", sharedExportFile, path)
		}
	}
}

func exprName(selector *ast.SelectorExpr) string {
	if ident, ok := selector.X.(*ast.Ident); ok {
		return ident.Name + "." + selector.Sel.Name
	}
	return "<selector>"
}

// 서술자 wrapper 는 **같은** 표를 돌려준다(시장마다 넷, 가족 순).
func TestTheSharedDescriptorsAreTheOneTable(t *testing.T) {
	for _, market := range []Market{MarketKR, MarketUS} {
		table := productionRouteDescriptors(market)
		shared := SharedProductionRouteDescriptors(market)
		if len(shared) != len(table) || len(shared) != 4 {
			t.Fatalf("%s: shared=%d table=%d, want 4", market, len(shared), len(table))
		}
		for index, lane := range shared {
			want, ok := table[lane.LaneID]
			if !ok || want.Family != lane.Family || want.Horizon != lane.Horizon || want.LaneVersion != lane.LaneVersion {
				t.Fatalf("%s: shared lane %+v is not the table's", market, lane)
			}
			if index > 0 && shared[index-1].Family >= lane.Family {
				t.Fatalf("%s: shared descriptors are not in family order", market)
			}
		}
	}
	if SharedProductionRouteDescriptors("JP") != nil {
		t.Fatal("an unknown market must have no descriptors")
	}
}
