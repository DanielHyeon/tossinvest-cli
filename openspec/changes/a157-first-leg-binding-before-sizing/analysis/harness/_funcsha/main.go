// _funcsha 는 Go 파일 하나에서 함수 선언 하나의 본문 sha256 을 냄(a157 design D3 · freeze 리뷰 Opus P2-1).
//
// logic-map 번들의 source_sha256 은 파일 sha 라서, 편집하는 파일 안의 무편집 함수(IssuePrecheckedQFinalCampaignFirstLeg)의
// 불변은 그것으로 잴 수 없음. 그래서 `func` 키워드부터 닫는 괄호까지(문서 주석 제외)의 원문 구간을 gofmt 로 정규화해 해시함.
// 디렉터리 이름이 `_` 로 시작하므로 `./...` 패턴(빌드 · vet · 시험)에는 잡히지 않음.
//
// 사용: go run ./openspec/changes/a157-first-leg-binding-before-sizing/analysis/harness/_funcsha <file.go> <Recv.Method|Func>
package main

import (
	"crypto/sha256"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: _funcsha <file.go> <Recv.Method|Func>")
		os.Exit(2)
	}
	path, want := os.Args[1], os.Args[2]
	src, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, src, parser.ParseComments)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	found := 0
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || qualified(fn) != want {
			continue
		}
		found++
		span := src[fset.Position(fn.Pos()).Offset:fset.Position(fn.End()).Offset]
		normalized, err := format.Source(span)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("%x  %s %s (%d-%d)\n", sha256.Sum256(normalized), path, want,
			fset.Position(fn.Pos()).Line, fset.Position(fn.End()).Line)
	}
	if found != 1 {
		fmt.Fprintf(os.Stderr, "want exactly one %s, found %d\n", want, found)
		os.Exit(1)
	}
}

// qualified 는 `Recv.Method`(포인터 수신자는 별표 없이) 또는 함수 이름.
func qualified(fn *ast.FuncDecl) string {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return fn.Name.Name
	}
	recv := fn.Recv.List[0].Type
	if star, ok := recv.(*ast.StarExpr); ok {
		recv = star.X
	}
	name := fmt.Sprint(recv)
	if ident, ok := recv.(*ast.Ident); ok {
		name = ident.Name
	}
	return strings.TrimSpace(name) + "." + fn.Name.Name
}
