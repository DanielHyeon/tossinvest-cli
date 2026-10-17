package engine

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// a112 0.5 리뷰 유지#2: 생산 loadShadow 정의 핀(laneStepFor 핀과 같은 방식 — 몸통 · 빌드 태그 · 정의 수).
// shadow 시험은 전부 tossos_testseams 빌드라 생산 정의(`!tossos_testseams`)는 어느 shadow 시험에서도 컴파일되지 않음 —
// 몸통을 비우는 변이(M4 `return strategyshadow.FamilyShadow{}, nil`)가 무태그 엔진 스위트를 통과했음(생산에서 SHADOW 가 조용히 안 섬).
// 이 시험은 태그 없이 돌고 디스크의 소스를 읽으므로 두 빌드의 정의를 다 봄.
func TestTheProductionLoadShadowIsTheProductionLoaderOutsideTestSeams(t *testing.T) {
	type definition struct{ path, tag, last string }
	var found []definition
	for _, path := range engineProductionFiles(t) {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		tag := ""
		if first := strings.SplitN(string(raw), "\n", 2)[0]; strings.HasPrefix(first, "//go:build ") {
			tag = strings.TrimPrefix(first, "//go:build ")
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		for _, decl := range file.Decls {
			function, ok := decl.(*ast.FuncDecl)
			if !ok || function.Name.Name != "loadShadow" || function.Recv == nil {
				continue
			}
			// 생산 정의는 문장 하나여야 하므로 마지막 문장 = 유일한 문장. seam 정의는 훅이 없을 때의 마지막 문장을 봄.
			last := ""
			if n := len(function.Body.List); n > 0 {
				var buffer bytes.Buffer
				if err := printer.Fprint(&buffer, fset, function.Body.List[n-1]); err != nil {
					t.Fatal(err)
				}
				last = buffer.String()
				if tag == "!tossos_testseams" && n != 1 {
					last = "<more than one statement>"
				}
			}
			found = append(found, definition{path: filepath.Base(path), tag: tag, last: last})
		}
	}
	if len(found) != 2 {
		t.Fatalf("loadShadow definitions=%+v, want exactly two (production and test seam)", found)
	}
	const loader = "return strategyshadow.LoadProductionFamilyShadow(ctx, config)"
	production, seam := 0, 0
	for _, value := range found {
		switch value.tag {
		case "!tossos_testseams":
			production++
			if value.last != loader {
				t.Fatalf("the production loadShadow in %s is %q — a production build must run exactly the manifest loader", value.path, value.last)
			}
		case "tossos_testseams":
			seam++
			if value.last != loader {
				t.Fatalf("the seam loadShadow in %s falls back to %q — without a hook it must run the production loader", value.path, value.last)
			}
		default:
			t.Fatalf("loadShadow in %s carries build constraint %q — a seam outside the test-seam build reaches the production binary",
				value.path, value.tag)
		}
	}
	if production != 1 || seam != 1 {
		t.Fatalf("loadShadow production=%d seam=%d, want one of each", production, seam)
	}
}
