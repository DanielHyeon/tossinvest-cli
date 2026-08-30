package strategyhandoff

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"io/fs"
	"sort"
	"strings"
	"testing"
)

// 이 파일은 "거절 여부를 안 보고 값을 읽을 수 없다"는 성질을 지킨다.
//
// 세 판본이 있었고 앞의 둘은 뚫렸다.
//
//  1. AST 가드: handoff 값을 쓰는 함수가 `Admitted()` 를 **부르는지**만 봤다.
//     호출만 남기고 관문을 지우면 통과했다.
//  2. 서명 검사: `Handoff` 의 **메서드**가 `strategyflow.Result` 를 돌려주면
//     두 값을 돌려주도록 요구했다. 적대 리뷰어 셋이 각자 같은 구멍을 찾았다 —
//     `func Escape(h Handoff) strategyflow.Result` 는 메서드가 아니라서 스캔
//     자체에 안 걸렸고, `Into(dst *Result)`·`Any() any`·`map[int]Result`·
//     `(Result, Result)`·`<-chan Result` 도 전부 통과했다. 모양을 열거하는
//     검사는 언제나 열거하지 못한 모양을 남긴다.
//  3. **지금 판본: 모양을 세지 않고 공개 표면 전체를 고정한다.** 이 패키지
//     밖에서 쓸 수 있는 것은 아래 표가 전부이고, 이름과 서명까지 같아야 한다.
//     새 탈출구는 예외 없이 새 공개 이름을 요구하므로 이 표와 어긋난다.
//
// 이 표를 늘리는 것 자체는 금지가 아니다. 금지는 **조용히** 늘리는 것이다.
var exportedSurface = map[string]string{
	"Capacity":      "const",
	"Refusal":       "type",
	"Admitted":      "const",
	"MarketClosed":  "const",
	"NoSelection":   "const",
	"OverCapacity":  "const",
	"OverCarried":   "const",
	"ErrNoDelivery": "var",
	"Handoff":       "type",

	"Admit": "func(ready bool, selected []strategyflow.Result) Handoff",

	"Handoff.Refusal": "func() Refusal",
	"Handoff.Pending": "func() int",
	// 값이 나가는 두 문. 서명까지 고정한다 — Single 이 bool 을 떼거나
	// Deliver 가 몸통 대신 값을 돌려주면 이 표가 깨진다.
	"Handoff.Single":  "func() (strategyflow.Result, bool)",
	"Handoff.Deliver": "func(to func(strategyflow.Result) error) error",
}

func TestThePackageExposesExactlyTheSurfaceTheSeamNeeds(t *testing.T) {
	found := make(map[string]string)
	for _, file := range productionFiles(t) {
		for _, decl := range file.Decls {
			switch value := decl.(type) {
			case *ast.FuncDecl:
				name := value.Name.Name
				if value.Recv != nil {
					if len(value.Recv.List) != 1 {
						continue
					}
					receiver := receiverTypeName(value.Recv.List[0].Type)
					if !ast.IsExported(receiver) {
						continue
					}
					name = receiver + "." + name
				}
				if !value.Name.IsExported() {
					continue
				}
				found[name] = types.ExprString(value.Type)
			case *ast.GenDecl:
				kind := map[token.Token]string{token.CONST: "const", token.VAR: "var", token.TYPE: "type"}[value.Tok]
				if kind == "" {
					continue
				}
				for _, spec := range value.Specs {
					switch inner := spec.(type) {
					case *ast.TypeSpec:
						if inner.Name.IsExported() {
							found[inner.Name.Name] = kind
						}
					case *ast.ValueSpec:
						for _, name := range inner.Names {
							if name.IsExported() {
								found[name.Name] = kind
							}
						}
					}
				}
			}
		}
	}
	if len(found) == 0 {
		t.Fatal("no exported declaration was scanned, so this surface is not pinned at all")
	}
	problems := make([]string, 0)
	for name, signature := range found {
		want, listed := exportedSurface[name]
		if !listed {
			problems = append(problems, "undeclared export "+name+" ("+signature+")")
			continue
		}
		if want != signature {
			problems = append(problems, name+" is "+signature+", the surface says "+want)
		}
	}
	for name, want := range exportedSurface {
		if _, ok := found[name]; !ok {
			problems = append(problems, name+" is on the surface ("+want+") but the package does not declare it")
		}
	}
	if len(problems) != 0 {
		sort.Strings(problems)
		t.Fatalf("the seam's public surface changed; every new door out of the boundary must be declared here: %v", problems)
	}
}

// Handoff 의 필드는 모두 비공개여야 한다. 하나라도 열리면 위의 표는 우회된다 —
// 부르는 쪽이 어떤 접근자도 거치지 않고 값을 읽으면 되기 때문이다.
func TestTheHandoffCarriesNoExportedField(t *testing.T) {
	checked := 0
	for _, file := range productionFiles(t) {
		ast.Inspect(file, func(node ast.Node) bool {
			spec, ok := node.(*ast.TypeSpec)
			if !ok || spec.Name.Name != "Handoff" {
				return true
			}
			structure, ok := spec.Type.(*ast.StructType)
			if !ok || structure.Fields == nil {
				t.Fatal("Handoff is no longer a struct, so this guard reads nothing")
			}
			for _, field := range structure.Fields.List {
				if len(field.Names) == 0 {
					t.Error("Handoff embeds a type, which can re-export a carried value")
				}
				for _, name := range field.Names {
					checked++
					if name.IsExported() {
						t.Errorf("Handoff.%s is exported; the admission answer can then be bypassed", name.Name)
					}
				}
			}
			return true
		})
	}
	if checked == 0 {
		t.Fatal("no Handoff field was scanned, so this guard proves nothing")
	}
}

// productionOnly 는 테스트 파일을 뺀다. 약속의 대상은 출하되는 코드다.
func productionOnly(info fs.FileInfo) bool {
	return !strings.HasSuffix(info.Name(), "_test.go")
}

func productionFiles(t *testing.T) []*ast.File {
	t.Helper()
	packages, err := parser.ParseDir(token.NewFileSet(), ".", productionOnly, 0)
	if err != nil {
		t.Fatalf("parse package directory: %v", err)
	}
	out := make([]*ast.File, 0, 2)
	for name, pkg := range packages {
		if name != "strategyhandoff" {
			t.Fatalf("this directory declares package %q", name)
		}
		for _, file := range pkg.Files {
			out = append(out, file)
		}
	}
	if len(out) == 0 {
		t.Fatal("no production file was parsed")
	}
	return out
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
