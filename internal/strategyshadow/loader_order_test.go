package strategyshadow

// a112 7.3.1(브리프 v3.3 §3): shadow 적재기는 활성화 적재기의 형태를 옮긴 사본이다 — 그래서 **양쪽**을 AST 로 못 박는다. 두 함수의 가드
// (if 문) 순서가 같아야 한다: 미선언 · ctx nil · ctx 취소 · 설정 결속 · 파일 읽기 · 핀 · 해석(정규) · 폐기 · 결속/수명/서술자 · ctx 취소.
// 한쪽만 바뀌면 이 시험이 깨진다(옮겨 적은 코드는 양쪽을 다 못 박아야 한다).

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
)

func loaderGuards(t *testing.T, path, name string) []string {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var body *ast.BlockStmt
	for _, decl := range file.Decls {
		if function, ok := decl.(*ast.FuncDecl); ok && function.Name.Name == name {
			body = function.Body
		}
	}
	if body == nil {
		t.Fatalf("%s not found in %s", name, path)
	}
	// 짝 이름을 한 어휘로 접는다(공유 wrapper ↔ 원 함수, 두 매니페스트의 같은 자리 함수).
	normalize := strings.NewReplacer(
		"strategyrouter.SharedProductionRouteOwnerUID", "ownerUID", "productionRouteOwnerUID", "ownerUID",
		"strategyrouter.SharedReadProductionRouteFile", "readFile", "readProductionRouteFile", "readFile",
		"strategyrouter.SharedProductionRouteDigestValid", "digestValid", "productionRouteDigestValid", "digestValid",
		"strategyrouter.SharedProductionRouteDigest", "digest", "productionRouteDigest", "digest",
		"strategyrouter.SharedProductionRouteIdentity", "identity", "productionRouteIdentity", "identity",
		"decodeProductionFamilyShadow", "decode", "decodeProductionFamilyActivation", "decode",
		"validateProductionFamilyShadow", "validate", "validateProductionFamilyActivation", "validate",
	)
	var guards []string
	var previous ast.Stmt
	for _, statement := range body.List {
		guard, ok := statement.(*ast.IfStmt)
		if !ok {
			previous = statement
			continue
		}
		var calls []string
		ast.Inspect(&ast.BlockStmt{List: []ast.Stmt{&ast.ExprStmt{X: guard.Cond}}}, func(node ast.Node) bool {
			if call, ok := node.(*ast.CallExpr); ok {
				calls = append(calls, normalize.Replace(exprText(call.Fun)))
			}
			return true
		})
		if guard.Init != nil {
			ast.Inspect(guard.Init, func(node ast.Node) bool {
				if call, ok := node.(*ast.CallExpr); ok {
					calls = append(calls, normalize.Replace(exprText(call.Fun)))
				}
				return true
			})
		}
		if len(calls) == 0 {
			calls = []string{normalize.Replace(exprText(guard.Cond))}
		}
		// `if err != nil` 은 바로 앞 대입의 호출로 이름 붙인다(읽기 · 해석 · 검증 중 무엇의 오류인지가 순서의 일부다).
		if assign, ok := previous.(*ast.AssignStmt); ok && len(assign.Rhs) == 1 {
			if call, ok := assign.Rhs[0].(*ast.CallExpr); ok && strings.Contains(exprText(guard.Cond), "err") {
				calls = append([]string{"after " + normalize.Replace(exprText(call.Fun))}, calls...)
			}
		}
		guards = append(guards, strings.Join(calls, "+"))
		previous = statement
	}
	return guards
}

func exprText(expression ast.Expr) string {
	switch value := expression.(type) {
	case *ast.Ident:
		return value.Name
	case *ast.SelectorExpr:
		return exprText(value.X) + "." + value.Sel.Name
	case *ast.BinaryExpr:
		return exprText(value.X) + value.Op.String() + exprText(value.Y)
	case *ast.CallExpr:
		return exprText(value.Fun) + "()"
	default:
		return "_"
	}
}

func TestTheShadowLoaderKeepsTheActivationLoadersGuardOrder(t *testing.T) {
	shadow := loaderGuards(t, "shadow.go", "LoadProductionFamilyShadow")
	activation := loaderGuards(t, filepath.Join("..", "strategyrouter", "production_family_activation.go"), "LoadProductionFamilyActivation")
	if strings.Join(shadow, " | ") != strings.Join(activation, " | ") {
		t.Fatalf("guard order diverged\n shadow:     %v\n activation: %v", shadow, activation)
	}
	t.Logf("guards: %v", shadow)
	if len(shadow) < 9 {
		t.Fatalf("only %d guards were read — the pin reads nothing: %v", len(shadow), shadow)
	}
}
