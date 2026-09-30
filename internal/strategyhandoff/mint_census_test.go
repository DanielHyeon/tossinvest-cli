package strategyhandoff

// 경계 값(Handoff · Delivered)을 **만드는** 자리를 이 패키지 안에서 **타입 동일성**으로 센다.
//
// 이력: 1판(2026-09-30 리뷰 수리)은 철자 `Handoff` · `Delivered` 로 셌고, codex 재확인이 셋을 뚫었다 — 비공개 별칭
// (`type hiddenDelivered = Delivered`)을 돌려주는 비공개 수신자 메서드, `init()` 에서 공개 var 를 재대입하는 모양, 결과가 없는
// out-param 함수(검사 전에 건너뜀). 철자는 별칭으로 언제나 다시 쓸 수 있으므로 여기서는 이름이 아니라 **타입**을 본다:
// 진짜 importer(소스 모드)로 패키지를 타입 검사하고,
//   ① 결과 타입(또는 out-param — 포인터 · 슬라이스 · 맵 · 채널)이 경계 타입을 **품는** 함수 · 메서드(수신자 공개 여부 · 별칭 무관)는
//      정확히 Admit · AdmitEachOwnerScope,
//   ② 경계 타입을 품는 패키지 수준 var 는 없다(함수 값 var 로 문을 다시 내놓는 모양),
//   ③ 패키지 수준 var 는 선언 초기화 밖에서 대입되지 않는다(`init()` 재대입 포함),
//   ④ 타입이 Handoff 인 합성 리터럴은 두 문 안에만, 선택을 싣는 것은 Admit 안에만, Delivered 리터럴은 Deliver 안에만,
//   ⑤ 두 타입의 비공개 필드에 리터럴 밖에서 쓰지 않는다.
// 「품는다」는 포인터 · 슬라이스 · 배열 · 맵 · 채널 · 구조체 필드 · 함수 서명(인자 · 결과) · 인터페이스 메서드를 따라 내려간다.
//
// **완전성 주장 철회(2026-10-01, codex 3차).** 이 census 도 모양을 센다 — 제네릭 주조(`T{result: r}` 의 T 가 타입 매개변수) +
// `any` 반환 + 매개변수 포인터 경유 재대입은 못 봤고, 쌍둥이 구조체 변환(`Delivered(twin{r})`)도 못 본다. 이 파일은 「왜 두 문뿐인가」의
// 설명이다. **종결은 source_freeze_test.go 의 소스 digest 동결이 진다.**

import (
	"go/ast"
	"go/build"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os/exec"
	"sort"
	"strings"
	"testing"
)

var mintingDeclarations = []string{"Admit", "AdmitEachOwnerScope"}

type seamPackage struct {
	pkg       *types.Package
	info      *types.Info
	files     []*ast.File
	handoff   types.Type
	delivered types.Type
}

func typeCheckSeam(t *testing.T) seamPackage {
	t.Helper()
	fset := token.NewFileSet()
	packages, err := parser.ParseDir(fset, ".", productionOnly, 0)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	var files []*ast.File
	for _, pkg := range packages {
		for _, file := range pkg.Files {
			files = append(files, file)
		}
	}
	info := &types.Info{Types: map[ast.Expr]types.TypeAndValue{}, Defs: map[*ast.Ident]types.Object{},
		Uses: map[*ast.Ident]types.Object{}, Selections: map[*ast.SelectorExpr]*types.Selection{}}
	// `-trimpath` 로 빌드한 시험 이진에서는 runtime.GOROOT() 가 비어 소스 importer 가 표준 패키지를 못 찾는다
	// (변이 하네스 · 게이트가 -trimpath 를 쓴다). 그때만 go 도구에게 묻는다.
	if build.Default.GOROOT == "" {
		out, err := exec.Command("go", "env", "GOROOT").Output()
		if err != nil {
			t.Fatalf("GOROOT unavailable for the source importer: %v", err)
		}
		build.Default.GOROOT = strings.TrimSpace(string(out))
	}
	config := types.Config{Importer: importer.ForCompiler(fset, "source", nil)}
	pkg, err := config.Check("github.com/JungHoonGhae/tossinvest-cli/internal/strategyhandoff", fset, files, info)
	if err != nil {
		// 미해소를 허용하면 판정이 약해진다 — 타입 검사가 끝까지 서야 한다.
		t.Fatalf("type-check the seam with the real importer: %v", err)
	}
	return seamPackage{pkg: pkg, info: info, files: files,
		handoff: pkg.Scope().Lookup("Handoff").Type(), delivered: pkg.Scope().Lookup("Delivered").Type()}
}

// carries 는 타입이 경계 타입을 품는지 답함(별칭은 타입 검사가 이미 풀었다).
func (s seamPackage) carries(typ types.Type) bool {
	return s.carriesSeen(typ, map[types.Type]bool{})
}

func (s seamPackage) carriesSeen(typ types.Type, seen map[types.Type]bool) bool {
	if typ == nil || seen[typ] {
		return false
	}
	seen[typ] = true
	if types.Identical(typ, s.handoff) || types.Identical(typ, s.delivered) {
		return true
	}
	switch value := typ.(type) {
	case *types.Named:
		return s.carriesSeen(value.Underlying(), seen)
	case *types.Pointer:
		return s.carriesSeen(value.Elem(), seen)
	case *types.Slice:
		return s.carriesSeen(value.Elem(), seen)
	case *types.Array:
		return s.carriesSeen(value.Elem(), seen)
	case *types.Map:
		return s.carriesSeen(value.Key(), seen) || s.carriesSeen(value.Elem(), seen)
	case *types.Chan:
		return s.carriesSeen(value.Elem(), seen)
	case *types.Struct:
		for i := 0; i < value.NumFields(); i++ {
			if s.carriesSeen(value.Field(i).Type(), seen) {
				return true
			}
		}
	case *types.Signature:
		return s.carriesSeen(value.Params(), seen) || s.carriesSeen(value.Results(), seen)
	case *types.Tuple:
		for i := 0; i < value.Len(); i++ {
			if s.carriesSeen(value.At(i).Type(), seen) {
				return true
			}
		}
	case *types.Interface:
		for i := 0; i < value.NumMethods(); i++ {
			if s.carriesSeen(value.Method(i).Type(), seen) {
				return true
			}
		}
	case *types.TypeParam:
		return s.carriesSeen(value.Constraint(), seen)
	}
	return false
}

// hands 는 서명이 경계 값을 밖으로 내놓을 수 있는지 답함: 결과가 품거나, 인자가 쓰기 가능한 그릇(포인터 · 슬라이스 · 맵 ·
// 채널)으로 품는 경우. 함수 인자(Deliver 의 몸통 `func(Delivered) error`)는 값을 **받는** 쪽이라 세지 않는다.
func (s seamPackage) hands(signature *types.Signature) bool {
	if s.carries(signature.Results()) {
		return true
	}
	for i := 0; i < signature.Params().Len(); i++ {
		switch param := signature.Params().At(i).Type().Underlying().(type) {
		case *types.Pointer, *types.Slice, *types.Map, *types.Chan:
			if s.carries(param) {
				return true
			}
		}
	}
	return false
}

func funcLabel(function *ast.FuncDecl) string {
	if function.Recv != nil && len(function.Recv.List) == 1 {
		return receiverTypeName(function.Recv.List[0].Type) + "." + function.Name.Name
	}
	return function.Name.Name
}

func TestOnlyTheTwoDoorsHandOutASeamValue(t *testing.T) {
	seam := typeCheckSeam(t)
	var got []string
	functions := 0
	for _, file := range seam.files {
		for _, decl := range file.Decls {
			function, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			functions++
			object, ok := seam.info.Defs[function.Name].(*types.Func)
			if !ok {
				t.Fatalf("%s has no type object", funcLabel(function))
			}
			if seam.hands(object.Type().(*types.Signature)) {
				got = append(got, funcLabel(function))
			}
		}
	}
	if functions == 0 {
		t.Fatal("no function was scanned")
	}
	sort.Strings(got)
	if strings.Join(got, ",") != strings.Join(mintingDeclarations, ",") {
		t.Fatalf("declarations that can hand out a seam value=%v, want exactly %v", got, mintingDeclarations)
	}
}

func TestNoPackageVariableCarriesOrIsReassigned(t *testing.T) {
	seam := typeCheckSeam(t)
	var problems []string
	scope := seam.pkg.Scope()
	variables := 0
	for _, name := range scope.Names() {
		variable, ok := scope.Lookup(name).(*types.Var)
		if !ok {
			continue
		}
		variables++
		if seam.carries(variable.Type()) {
			problems = append(problems, "package variable "+name+" carries a seam type ("+variable.Type().String()+")")
		}
	}
	if variables == 0 {
		t.Fatal("no package variable was scanned (ErrNoDelivery should be one)")
	}
	for _, file := range seam.files {
		ast.Inspect(file, func(node ast.Node) bool {
			var targets []ast.Expr
			switch value := node.(type) {
			case *ast.AssignStmt:
				targets = value.Lhs
			case *ast.IncDecStmt:
				targets = []ast.Expr{value.X}
			}
			for _, target := range targets {
				ast.Inspect(target, func(inner ast.Node) bool {
					if ident, ok := inner.(*ast.Ident); ok {
						if variable, ok := seam.info.Uses[ident].(*types.Var); ok && variable.Parent() == scope {
							problems = append(problems, "package variable "+ident.Name+" is assigned outside its declaration")
						}
					}
					return true
				})
			}
			return true
		})
	}
	if len(problems) != 0 {
		sort.Strings(problems)
		t.Fatalf("%v", problems)
	}
}

func TestSeamValuesAreBuiltOnlyInsideTheirDoors(t *testing.T) {
	seam := typeCheckSeam(t)
	var problems []string
	literals := map[string]int{}
	for _, file := range seam.files {
		for _, decl := range file.Decls {
			owner := "(package level)"
			if function, ok := decl.(*ast.FuncDecl); ok {
				owner = funcLabel(function)
			}
			ast.Inspect(decl, func(node ast.Node) bool {
				switch value := node.(type) {
				case *ast.CompositeLit:
					typ := seam.info.Types[value].Type
					switch {
					case types.Identical(typ, seam.handoff):
						literals[owner+":Handoff"]++
						if owner != "Admit" && owner != "AdmitEachOwnerScope" {
							problems = append(problems, owner+" builds a Handoff")
						}
						if carriesSelection(value) && owner != "Admit" {
							problems = append(problems, owner+" builds a Handoff carrying a selection")
						}
					case types.Identical(typ, seam.delivered):
						literals[owner+":Delivered"]++
						if owner != "Handoff.Deliver" {
							problems = append(problems, owner+" builds a Delivered")
						}
					}
				case *ast.AssignStmt:
					for _, left := range value.Lhs {
						if selector, ok := left.(*ast.SelectorExpr); ok {
							if selection := seam.info.Selections[selector]; selection != nil && selection.Kind() == types.FieldVal {
								receiver := selection.Recv()
								if pointer, ok := receiver.(*types.Pointer); ok {
									receiver = pointer.Elem()
								}
								if types.Identical(receiver, seam.handoff) || types.Identical(receiver, seam.delivered) {
									problems = append(problems, owner+" writes the field "+selector.Sel.Name+" outside a literal")
								}
							}
						}
					}
				}
				return true
			})
		}
	}
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
