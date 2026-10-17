package engine

// a112 3.8(-e) 감사 보강(Manager 판정 2026-10-04): 전략 경로에는 Toss 조건부 주문(수동 조건 주문) 변경자에 닿는 자리가 없다.
//
// 레인 · 증거 · 후보 쪽은 폐포 가드가 이미 막는다(internal/official 이 그 패키지들의 폐포에 없음). 엔진은 다르다 — 엔진은 보호 주문을
// 내야 하므로 internal/official 을 들여오고 조건부 주문 변경자를 실제로 부른다(broker.go 의 officialBroker 어댑터 — 손절 · 익절 보호
// 경로). 그래서 여기서는 **자리** 를 센다:
//   ① 엔진 생산 파일에서 조건부 주문 변경자 이름(Create/Cancel/Modify ConditionalOrder · ModifyConditionalOrderRef)과 조건부 주문 의도
//      타입(orderintent.Conditional*) 이 나오는 파일은 broker.go 하나뿐이다.
//   ② 전략 파일(strategy_*.go) 은 그 이름 · 타입을 하나도 쓰지 않는다 — 어댑터를 거치는 새 호출도 이름이 같으므로 여기서 걸린다.
// 빈 표본 실패: 생산 파일 · 전략 파일을 하나 이상 읽어야 하고, broker.go 에서 변경자 이름을 실제로 봐야 한다(census 가 눈먼 경우 배제).

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

var a112ConditionalOrderMutators = map[string]bool{
	"CreateConditionalOrder": true, "CancelConditionalOrder": true, "ModifyConditionalOrder": true, "ModifyConditionalOrderRef": true,
}

func TestNoStrategyFileReachesAConditionalOrderMutator(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	sitesByFile := map[string][]string{}
	production, strategy := 0, 0
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		production++
		if strings.HasPrefix(path, "strategy_") {
			strategy++
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		ast.Inspect(file, func(node ast.Node) bool {
			switch value := node.(type) {
			case *ast.SelectorExpr:
				name := value.Sel.Name
				if a112ConditionalOrderMutators[name] || strings.HasPrefix(name, "Conditional") && isOrderIntentSelector(value) {
					sitesByFile[path] = append(sitesByFile[path], fset.Position(value.Pos()).String()+" "+name)
				}
			case *ast.FuncDecl:
				if a112ConditionalOrderMutators[value.Name.Name] {
					sitesByFile[path] = append(sitesByFile[path], fset.Position(value.Pos()).String()+" func "+value.Name.Name)
				}
			}
			return true
		})
	}
	if production == 0 || strategy == 0 {
		t.Fatalf("read %d production and %d strategy files — the census read nothing", production, strategy)
	}
	if len(sitesByFile["broker.go"]) == 0 {
		t.Fatal("broker.go names no conditional-order mutator — the census is blind (the protection adapter does call them)")
	}
	var outside []string
	for path, sites := range sitesByFile {
		if path == "broker.go" {
			continue
		}
		outside = append(outside, sites...)
	}
	sort.Strings(outside)
	if len(outside) != 0 {
		t.Fatalf("a conditional-order mutator or intent is named outside the protection adapter (broker.go):\n%s", strings.Join(outside, "\n"))
	}
}

// isOrderIntentSelector 는 `orderintent.Conditional…` 꼴(패키지 선택자)인지 본다.
func isOrderIntentSelector(selector *ast.SelectorExpr) bool {
	pkg, ok := selector.X.(*ast.Ident)
	return ok && pkg.Name == "orderintent"
}
