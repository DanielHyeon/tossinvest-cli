package engine

import (
	"go/ast"
	"go/constant"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

func TestA087ProtectiveActionValuesFollowTypedActionSet(t *testing.T) {
	_, here, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	dir := filepath.Clean(filepath.Join(filepath.Dir(here), "..", "..", "exitpolicy"))
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	var files []*ast.File
	declared := map[string]bool{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, filepath.Join(dir, entry.Name()), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		if a087ActionConstNames(file, declared) || entry.Name() == "ratchet.go" {
			files = append(files, file)
		}
	}
	config := types.Config{Importer: importer.Default(), IgnoreFuncBodies: true}
	pkg, err := config.Check("github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy", fset, files, nil)
	if err != nil {
		t.Fatalf("go/types exitpolicy 검사: %v", err)
	}
	actionName, ok := pkg.Scope().Lookup("Action").(*types.TypeName)
	if !ok {
		t.Fatal("go/types Action 타입을 찾지 못함")
	}
	all := map[string][]string{}
	typed := map[string]bool{}
	for _, name := range pkg.Scope().Names() {
		constantObject, ok := pkg.Scope().Lookup(name).(*types.Const)
		if !ok || !types.Identical(constantObject.Type(), actionName.Type()) {
			continue
		}
		value := constant.StringVal(constantObject.Val())
		all[value] = append(all[value], name)
		typed[name] = true
	}
	if !a087SameStringSet(declared, typed) {
		t.Fatalf("AST Action 상수=%v, go/types Action 상수=%v", a087SortedKeys(declared), a087SortedKeys(typed))
	}
	if len(all) == 0 {
		t.Fatal("Action 상수 집합이 비어 있음")
	}
	got := map[string]bool{}
	for value, names := range all {
		sort.Strings(names)
		t.Logf("Action 값 %q: %s", value, strings.Join(names, ","))
		if isProtective(exitpolicy.Proposal{Action: exitpolicy.Action(value)}) {
			got[value] = true
		}
	}
	if len(got) == 0 {
		t.Fatal("보호 Action 값 집합이 비어 있음")
	}
	want := map[string]bool{}
	for _, value := range protectiveActionValues() {
		want[value] = true
	}
	if len(want) == 0 {
		t.Fatal("재구성 action 값 집합이 비어 있음")
	}
	if !a087SameStringSet(got, want) {
		t.Fatalf("보호 Action 값=%v, 재구성 action 값=%v", a087SortedKeys(got), a087SortedKeys(want))
	}
}

// a087ActionConstNames 는 패키지 전체의 Action 상수 선언을 잡아, 타입 검사 대상 파일이 누락되지 않게 함.
func a087ActionConstNames(file *ast.File, names map[string]bool) bool {
	found := false
	for _, declaration := range file.Decls {
		group, ok := declaration.(*ast.GenDecl)
		if !ok || group.Tok != token.CONST {
			continue
		}
		for _, specification := range group.Specs {
			value, ok := specification.(*ast.ValueSpec)
			if !ok || !a087ActionTypedValue(value) {
				continue
			}
			for _, name := range value.Names {
				names[name.Name] = true
			}
			found = true
		}
	}
	return found
}

func a087ActionTypedValue(value *ast.ValueSpec) bool {
	if named, ok := value.Type.(*ast.Ident); ok && named.Name == "Action" {
		return true
	}
	for _, expression := range value.Values {
		call, ok := expression.(*ast.CallExpr)
		if !ok {
			continue
		}
		if named, ok := call.Fun.(*ast.Ident); ok && named.Name == "Action" {
			return true
		}
	}
	return false
}

func a087SameStringSet(left, right map[string]bool) bool {
	if len(left) != len(right) {
		return false
	}
	for value := range left {
		if !right[value] {
			return false
		}
	}
	return true
}

func a087SortedKeys(values map[string]bool) []string {
	out := make([]string, 0, len(values))
	for value := range values {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}
