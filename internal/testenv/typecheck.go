package testenv

// typecheck.go — 심볼 census 의 공용 부품(a112 8.2). 「이 패키지가 저 패키지의 무엇을 쓰는가」를 철자가 아니라 **해소된 객체**로 센다:
// import 별칭 · 점 import · 같은 이름의 지역 식별자는 철자 census 를 속이지만 타입 검사기의 Uses 표는 속이지 못한다.

import (
	"go/ast"
	"go/build"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// CheckedPackage 는 타입 검사가 끝난 생산 파일 묶음이다.
type CheckedPackage struct {
	Files []*ast.File
	Fset  *token.FileSet
	Info  *types.Info
}

// TypeCheckProduction 은 dir 의 생산(.go, _test.go 제외) 파일을 tags 빌드 구성으로 골라 진짜 importer(소스 모드)로 타입 검사한다.
// 태그 뒤 생산 파일(`*_testseam.go`)도 그 태그를 주면 함께 센다 — 생산 빌드에 없는 문도 census 에서 빠지지 않게.
func TypeCheckProduction(t *testing.T, dir, importPath string, tags ...string) CheckedPackage {
	t.Helper()
	// `-trimpath` 시험 이진에서는 runtime.GOROOT() 가 비어 소스 importer 가 표준 패키지를 못 찾는다 — 그때만 go 도구에게 묻는다.
	if build.Default.GOROOT == "" {
		out, err := exec.Command("go", "env", "GOROOT").Output()
		if err != nil {
			t.Fatalf("GOROOT unavailable for the source importer: %v", err)
		}
		build.Default.GOROOT = strings.TrimSpace(string(out))
	}
	context := build.Default
	context.BuildTags = append(append([]string(nil), context.BuildTags...), tags...)
	names, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	var files []*ast.File
	for _, path := range names {
		base := filepath.Base(path)
		if strings.HasSuffix(base, "_test.go") {
			continue
		}
		match, err := context.MatchFile(dir, base)
		if err != nil {
			t.Fatalf("match %s: %v", base, err)
		}
		if !match {
			continue
		}
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("parse %s: %v", base, err)
		}
		files = append(files, file)
	}
	if len(files) == 0 {
		t.Fatalf("no production file selected in %s — the census would read nothing", dir)
	}
	info := &types.Info{Uses: map[*ast.Ident]types.Object{}, Selections: map[*ast.SelectorExpr]*types.Selection{}}
	config := types.Config{Importer: importer.ForCompiler(fset, "source", nil)}
	if _, err := config.Check(importPath, fset, files, info); err != nil {
		// 미해소를 허용하면 census 가 약해진다 — 타입 검사가 끝까지 서야 한다.
		t.Fatalf("type-check %s with the real importer: %v", importPath, err)
	}
	return CheckedPackage{Files: files, Fset: fset, Info: info}
}

// SymbolUse 는 한 패키지 객체의 사용 한 건이다.
type SymbolUse struct {
	Object   types.Object
	Position string
}

// UsesFrom 는 이 묶음이 pkgPath 패키지에서 쓰는 객체(함수 · 메서드 · 타입 · 필드 · 변수 · 상수) 전부를 위치 순으로 돌려준다.
func (checked CheckedPackage) UsesFrom(pkgPath string) []SymbolUse {
	var uses []SymbolUse
	for ident, object := range checked.Info.Uses {
		if object == nil || object.Pkg() == nil || object.Pkg().Path() != pkgPath {
			continue
		}
		uses = append(uses, SymbolUse{Object: object, Position: checked.Fset.Position(ident.Pos()).String()})
	}
	sort.Slice(uses, func(i, j int) bool { return uses[i].Position < uses[j].Position })
	return uses
}

// Describe 는 객체를 「종류 이름(수신자)」 꼴로 적는다 — 실패 메시지가 무엇을 썼는지 말하게.
func Describe(object types.Object) string {
	switch value := object.(type) {
	case *types.Func:
		if receiver := value.Type().(*types.Signature).Recv(); receiver != nil {
			return "method " + types.TypeString(receiver.Type(), nil) + "." + value.Name()
		}
		return "func " + value.Name()
	case *types.TypeName:
		return "type " + value.Name()
	case *types.Var:
		if value.IsField() {
			return "field " + value.Name()
		}
		return "var " + value.Name()
	case *types.Const:
		return "const " + value.Name()
	}
	return object.String()
}

// ReceiverNamed 는 메서드 수신자의 이름 붙은 타입 이름을 돌려준다(포인터 수신자 포함). 메서드가 아니면 "".
func ReceiverNamed(object types.Object) string {
	function, ok := object.(*types.Func)
	if !ok {
		return ""
	}
	receiver := function.Type().(*types.Signature).Recv()
	if receiver == nil {
		return ""
	}
	typ := receiver.Type()
	if pointer, ok := typ.(*types.Pointer); ok {
		typ = pointer.Elem()
	}
	if named, ok := typ.(*types.Named); ok {
		return named.Obj().Name()
	}
	return ""
}
