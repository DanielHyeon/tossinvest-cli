package strategyhandoff

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"sort"
	"strings"
	"testing"
)

// 이 파일은 "거절 여부를 안 보고 값을 읽을 수 없다"는 성질을 **타입 서명**에
// 못 박는다.
//
// 첫 판본에서 그 성질은 AST 가드 하나가 지키고 있었다. 그 가드는 handoff 값을
// 쓰는 함수가 Admitted() 를 **부르는지**만 봤고, 그것이 실제로 관문 역할을
// 하는지는 보지 않았다. 그래서 `_ = handoff.Admitted()` 로 호출만 남기고 관문을
// 지우거나, `var` 선언으로 바인딩하거나, 아예 바인딩 없이 필드를 바로 읽으면
// 전부 통과했다 — 두 스위트가 모두 초록인 채로.
//
// 지금은 값이 Single() 로만 나가고 그 서명이 bool 을 함께 준다. 무시하려면
// 명시적으로 버려야 한다. 남은 위험은 하나뿐이다: 누가 나중에 Result 를 그냥
// 돌려주는 접근자를 새로 다는 것. 이 검사는 그 문 하나를 잠근다.
func TestNoAccessorHandsOutAResultWithoutSayingWhetherItWasAdmitted(t *testing.T) {
	fset := token.NewFileSet()
	packages, err := parser.ParseDir(fset, ".", productionOnly, 0)
	if err != nil {
		t.Fatalf("parse package directory: %v", err)
	}
	checked := 0
	offenders := make([]string, 0)
	for _, pkg := range packages {
		for path, file := range pkg.Files {
			for _, decl := range file.Decls {
				function, ok := decl.(*ast.FuncDecl)
				if !ok || function.Recv == nil || len(function.Recv.List) != 1 {
					continue
				}
				if receiverTypeName(function.Recv.List[0].Type) != "Handoff" {
					continue
				}
				checked++
				results, carries := 0, false
				if function.Type.Results != nil {
					for _, field := range function.Type.Results.List {
						names := len(field.Names)
						if names == 0 {
							names = 1
						}
						results += names
						if isResultType(field.Type) {
							carries = true
						}
					}
				}
				// Result 를 내주는 접근자는 반드시 두 값을 돌려준다. 하나만
				// 돌려주면 부르는 쪽이 거절을 볼 방법이 없다.
				if carries && results < 2 {
					offenders = append(offenders, path+":"+function.Name.Name)
				}
			}
		}
	}
	if checked == 0 {
		t.Fatal("no method on Handoff was scanned, so this guard proves nothing")
	}
	if len(offenders) != 0 {
		sort.Strings(offenders)
		t.Fatalf("these accessors hand out a strategyflow.Result without a companion admission answer: %v", offenders)
	}
}

// Handoff 의 필드는 모두 비공개여야 한다. 하나라도 열리면 위의 서명 검사는
// 우회된다 — 부르는 쪽이 접근자를 거치지 않고 값을 읽으면 되기 때문이다.
func TestTheHandoffCarriesNoExportedField(t *testing.T) {
	fset := token.NewFileSet()
	packages, err := parser.ParseDir(fset, ".", productionOnly, 0)
	if err != nil {
		t.Fatalf("parse package directory: %v", err)
	}
	checked := 0
	for _, pkg := range packages {
		for path, file := range pkg.Files {
			ast.Inspect(file, func(node ast.Node) bool {
				spec, ok := node.(*ast.TypeSpec)
				if !ok || spec.Name.Name != "Handoff" {
					return true
				}
				structure, ok := spec.Type.(*ast.StructType)
				if !ok || structure.Fields == nil {
					t.Fatalf("%s: Handoff is no longer a struct, so this guard reads nothing", path)
				}
				for _, field := range structure.Fields.List {
					if len(field.Names) == 0 {
						t.Errorf("%s: Handoff embeds a type, which can re-export a carried value", path)
					}
					for _, name := range field.Names {
						checked++
						if name.IsExported() {
							t.Errorf("%s: Handoff.%s is exported; the admission answer can then be bypassed", path, name.Name)
						}
					}
				}
				return true
			})
		}
	}
	if checked == 0 {
		t.Fatal("no Handoff field was scanned, so this guard proves nothing")
	}
}

// productionOnly 는 테스트 파일을 뺀다. 약속의 대상은 출하되는 코드다.
func productionOnly(info fs.FileInfo) bool {
	return !strings.HasSuffix(info.Name(), "_test.go")
}

func receiverTypeName(expr ast.Expr) string {
	if star, ok := expr.(*ast.StarExpr); ok {
		expr = star.X
	}
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return ""
	}
	return ident.Name
}

func isResultType(expr ast.Expr) bool {
	switch value := expr.(type) {
	case *ast.SelectorExpr:
		pkg, ok := value.X.(*ast.Ident)
		return ok && pkg.Name == "strategyflow" && value.Sel.Name == "Result"
	case *ast.ArrayType:
		return isResultType(value.Elt)
	case *ast.StarExpr:
		return isResultType(value.X)
	}
	return false
}
