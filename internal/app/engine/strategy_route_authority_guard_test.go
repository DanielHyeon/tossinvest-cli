package engine

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

// 태스크 4.3.2: 아래 정확한 생산 호출자 파일에서 strategyrouter.Route 선택 호출을 금지하고,
// 최소 한 번의 RouteSet 호출을 요구한다. import 별칭을 실제로 해석해서 확인하므로
// 별칭을 바꿔 우회할 수 없다. 이 닫힘 밖의 레거시 Route 와 그 호출자는 그대로 둔다.
const strategyRouterImportPath = "github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"

func strategyRouteSetGuardFiles(t *testing.T) []string {
	t.Helper()
	exact := []string{
		"strategy_route_authority.go",
		"strategy_proposal_authority.go",
		"strategy_entry_supervisor.go",
	}
	files := make([]string, 0, len(exact)+2)
	for _, name := range exact {
		path := filepath.Join(".", name)
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("guarded production caller %s is missing: %v", name, err)
		}
		files = append(files, path)
	}
	// 앞으로 생길 조정자 파일도 같은 규칙을 받는다.
	coordinators, err := filepath.Glob(filepath.Join(".", "strategy_*coordinator*.go"))
	if err != nil {
		t.Fatalf("coordinator glob: %v", err)
	}
	for _, path := range coordinators {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		files = append(files, path)
	}
	return files
}

// resolvedStrategyRouterName 은 그 파일이 strategyrouter 를 어떤 이름으로 부르는지 찾는다.
// 별칭이 없으면 패키지 이름이 그대로 쓰인다.
func resolvedStrategyRouterName(file *ast.File) (string, bool) {
	for _, spec := range file.Imports {
		path, err := strconv.Unquote(spec.Path.Value)
		if err != nil || path != strategyRouterImportPath {
			continue
		}
		if spec.Name != nil {
			if spec.Name.Name == "_" || spec.Name.Name == "." {
				return "", false
			}
			return spec.Name.Name, true
		}
		return "strategyrouter", true
	}
	return "", false
}

func TestGuardedProductionCallersUseRouteSetAndNeverRoute(t *testing.T) {
	fset := token.NewFileSet()
	guarded := strategyRouteSetGuardFiles(t)
	sawRouteSetSomewhere := false
	for _, path := range guarded {
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		name, ok := resolvedStrategyRouterName(file)
		if !ok {
			// 이 파일은 라우터를 전혀 쓰지 않는다. 그럼 금지할 것도 없다.
			continue
		}
		routeCalls, routeSetCalls := 0, 0
		ast.Inspect(file, func(node ast.Node) bool {
			call, isCall := node.(*ast.CallExpr)
			if !isCall {
				return true
			}
			selector, isSelector := call.Fun.(*ast.SelectorExpr)
			if !isSelector {
				return true
			}
			ident, isIdent := selector.X.(*ast.Ident)
			if !isIdent || ident.Name != name {
				return true
			}
			switch selector.Sel.Name {
			case "Route":
				routeCalls++
			case "RouteSet":
				routeSetCalls++
			}
			return true
		})
		if routeCalls != 0 {
			t.Fatalf("%s calls %s.Route %d time(s); pre-evaluation selection is forbidden in this closure", path, name, routeCalls)
		}
		if routeSetCalls > 0 {
			sawRouteSetSomewhere = true
		}
	}
	if !sawRouteSetSomewhere {
		t.Fatal("no guarded production caller resolves a RouteSet call, so the sealed route-set authority is unused")
	}
}

// 이 닫힘 밖의 레거시 Route 는 그대로 살아 있어야 한다. 지워 버리면
// "행동이 변하지 않았다"는 주장을 확인할 대상 자체가 사라진다.
func TestLegacyRouteStillExistsOutsideTheGuardedClosure(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("..", "..", "strategyrouter", "router.go"))
	if err != nil {
		t.Fatalf("read legacy router: %v", err)
	}
	if !strings.Contains(string(source), "func Route(request RouteRequest) RouteResult {") {
		t.Fatal("legacy Route was removed or its signature changed")
	}
}
