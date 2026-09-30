package strategyhandoff

// 경계 값(Handoff · Delivered)을 **만드는** 자리를 이 패키지 안에서 센다(2026-09-30 리뷰 — codex P1-2, 보이스 C #1).
//
// 동결 표면 표(escape_test.go)는 공개 이름과 서명을 고정하지만 두 모양을 못 봤다: 비공개 타입의 메서드(표는 공개 수신자만
// 본다)와 공개 var 의 값 타입(표가 "var" 한 단어만 적었다). 둘을 엮으면 `ErrNoDelivery.Mint(r)` 같은 둘째 주조 문이 표 ·
// 엔진 census 양쪽을 통과했다. 여기서는 이름이 아니라 **주조 행위**를 센다:
//   ① 결과 타입에 Handoff · Delivered 가 나오는 함수 · 메서드(수신자 공개 여부 무관)는 정확히 Admit · AdmitEachOwnerScope.
//   ② Handoff 합성 리터럴은 Admit · AdmitEachOwnerScope 안에만, 선택을 싣는(`selected:`) 리터럴은 Admit 안에만.
//   ③ Delivered 합성 리터럴은 Deliver 안에만.
//   ④ 두 타입의 비공개 필드에 대입하는 문은 어디에도 없다(리터럴을 거치지 않는 주조).

import (
	"go/ast"
	"go/types"
	"sort"
	"strings"
	"testing"
)

// 결과로 경계 값을 내놓을 수 있는 선언의 전부.
var mintingDeclarations = []string{"Admit", "AdmitEachOwnerScope"}

func declLabel(function *ast.FuncDecl) string {
	if function.Recv != nil && len(function.Recv.List) == 1 {
		return receiverTypeName(function.Recv.List[0].Type) + "." + function.Name.Name
	}
	return function.Name.Name
}

func mentionsSeamType(expr ast.Expr) bool {
	found := false
	ast.Inspect(expr, func(node ast.Node) bool {
		if ident, ok := node.(*ast.Ident); ok && (ident.Name == "Handoff" || ident.Name == "Delivered") {
			found = true
		}
		return !found
	})
	return found
}

func TestOnlyTheTwoDoorsReturnASeamValue(t *testing.T) {
	var got []string
	functions := 0
	for _, file := range productionFiles(t) {
		for _, decl := range file.Decls {
			function, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			functions++
			if function.Type.Results == nil {
				continue
			}
			for _, result := range function.Type.Results.List {
				if mentionsSeamType(result.Type) {
					got = append(got, declLabel(function))
					break
				}
			}
			// 포인터 인자로 채우는 모양(`func F(dst *Handoff)`)도 주조다.
			for _, param := range function.Type.Params.List {
				if _, pointer := param.Type.(*ast.StarExpr); pointer && mentionsSeamType(param.Type) {
					got = append(got, declLabel(function)+"(out-param)")
				}
			}
		}
	}
	if functions == 0 {
		t.Fatal("no function was scanned, so this census proves nothing")
	}
	sort.Strings(got)
	if strings.Join(got, ",") != strings.Join(mintingDeclarations, ",") {
		t.Fatalf("declarations that can hand out a seam value=%v, want exactly %v — a new door must be declared here and in the engine allow-list",
			got, mintingDeclarations)
	}
}

func TestSeamValuesAreBuiltOnlyInsideTheirDoors(t *testing.T) {
	var problems []string
	literals := map[string]int{}
	for _, file := range productionFiles(t) {
		for _, decl := range file.Decls {
			owner := "(package level)"
			if function, ok := decl.(*ast.FuncDecl); ok {
				owner = declLabel(function)
			}
			ast.Inspect(decl, func(node ast.Node) bool {
				switch value := node.(type) {
				case *ast.CompositeLit:
					typeName := types.ExprString(value.Type)
					// `[]Handoff{{…}}` 의 원소는 타입이 생략된 리터럴이다 — 바깥 슬라이스 타입으로 센다.
					if array, ok := value.Type.(*ast.ArrayType); ok {
						element := types.ExprString(array.Elt)
						for _, item := range value.Elts {
							if inner, ok := item.(*ast.CompositeLit); ok && inner.Type == nil {
								literals[owner+":"+element]++
								if element == "Handoff" && owner != "Admit" && owner != "AdmitEachOwnerScope" {
									problems = append(problems, owner+" builds an element Handoff")
								}
								if element == "Handoff" && carriesSelection(inner) && owner != "Admit" {
									problems = append(problems, owner+" builds a Handoff carrying a selection")
								}
							}
						}
						return true
					}
					switch typeName {
					case "Handoff":
						literals[owner+":Handoff"]++
						if owner != "Admit" && owner != "AdmitEachOwnerScope" {
							problems = append(problems, owner+" builds a Handoff")
						}
						if carriesSelection(value) && owner != "Admit" {
							problems = append(problems, owner+" builds a Handoff carrying a selection")
						}
					case "Delivered":
						literals[owner+":Delivered"]++
						if owner != "Handoff.Deliver" {
							problems = append(problems, owner+" builds a Delivered")
						}
					}
				case *ast.AssignStmt:
					for _, left := range value.Lhs {
						if selector, ok := left.(*ast.SelectorExpr); ok {
							switch selector.Sel.Name {
							case "selected", "refusal", "pending", "result":
								problems = append(problems, owner+" writes the field "+selector.Sel.Name+" outside a literal")
							}
						}
					}
				}
				return true
			})
		}
	}
	// 양성 대조: 오늘 알려진 주조 자리가 실제로 세어졌는가(세기가 눈멀면 위 검사는 공허하다).
	for _, want := range []string{"Admit:Handoff", "AdmitEachOwnerScope:Handoff", "Handoff.Deliver:Delivered"} {
		if literals[want] == 0 {
			t.Errorf("the census did not see the known construction %s — it is blind: %v", want, literals)
		}
	}
	if len(problems) != 0 {
		sort.Strings(problems)
		t.Fatalf("a seam value is built outside its door: %v", problems)
	}
}

func carriesSelection(literal *ast.CompositeLit) bool {
	for _, element := range literal.Elts {
		if pair, ok := element.(*ast.KeyValueExpr); ok {
			if key, ok := pair.Key.(*ast.Ident); ok && key.Name == "selected" {
				return true
			}
		}
	}
	return false
}
