//go:build ignore

// census 는 a087 tasks 3.7 ① 의 열거기임 — 비시험 Go 파일 전체를 go/parser 로 읽어, 최상위 선언(함수·const·var)
// 하나에 든 문자열 리터럴을 이어 붙인 SQL 텍스트가 intents 표를 읽고(FROM/JOIN/쉼표 조인) 동시에 `price` 낱말을
// 담는 선언을 셈. 손으로 고른 파일이 아니라 저장소 전체를 걸으므로 볼 곳을 고르지 않음.
//
// 한계(문장으로 남김): 문자열을 실행 시점에 조립(fmt.Sprintf·+ 연결 변수)하면 선언 단위로만 묶이므로,
// 다른 선언의 const 를 이어 붙이는 질의는 const 쪽 선언으로 잡힘 — 그 const 의 사용처는 ② 에서 따라감.
// `price` 가 intents 의 열인지 다른 표의 열인지는 이 도구가 가르지 않음 — 출력의 줄을 사람이 판정함(README 표).
//
// 사용(저장소 루트에서):
//
//	go run ./openspec/changes/a087-a-protective-exit-is-a-market-order/analysis/intents-price-null/census.go
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	// intents 표를 읽는 자리 — FROM/JOIN 뒤 또는 쉼표 조인. INSERT INTO intents 는 쓰기라 제외됨.
	readsIntents = regexp.MustCompile(`(?i)\b(from|join)\s+intents\b|,\s*intents\s+(as\s+)?[a-z_]+\b`)
	// price 낱말 — average_price·prev_avg_price 같은 다른 열은 `_` 가 낱말 문자라 걸리지 않음.
	priceWord = regexp.MustCompile(`(?i)\bprice\b`)
	// 열 이름 없이 price 를 읽는 자리 — SELECT * 또는 별칭.* (낱말 price 가 없어도 열을 가져감).
	starSelect = regexp.MustCompile(`(?i)\bselect\s+(distinct\s+)?\*|\b[a-z_]+\.\*`)
)

type hit struct {
	file, decl string
	line       int
	priceLines []string
	star       bool
}

func main() {
	var hits []hit
	scanned := 0
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "openspec", "vendor", "node_modules", ".sdd", ".codegraph":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		fset := token.NewFileSet()
		f, perr := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if perr != nil {
			return fmt.Errorf("%s: %w", path, perr)
		}
		scanned++
		for _, decl := range f.Decls {
			var parts []string
			var lines []string
			ast.Inspect(decl, func(n ast.Node) bool {
				lit, ok := n.(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					return true
				}
				s, uerr := strconv.Unquote(lit.Value)
				if uerr != nil {
					s = lit.Value
				}
				parts = append(parts, s)
				// price 가 든 줄을 원본 좌표와 함께 남김 — 사람 판정의 입력.
				base := fset.Position(lit.Pos()).Line
				for i, l := range strings.Split(s, "\n") {
					if priceWord.MatchString(l) {
						lines = append(lines, fmt.Sprintf("%d: %s", base+i, strings.TrimSpace(l)))
					}
				}
				return true
			})
			text := strings.Join(parts, "\n")
			star := starSelect.MatchString(text)
			if !readsIntents.MatchString(text) || !(priceWord.MatchString(text) || star) {
				continue
			}
			hits = append(hits, hit{
				file: filepath.ToSlash(path), decl: declName(decl),
				line: fset.Position(decl.Pos()).Line, priceLines: lines, star: star,
			})
		}
		return nil
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	sort.Slice(hits, func(i, j int) bool {
		if hits[i].file != hits[j].file {
			return hits[i].file < hits[j].file
		}
		return hits[i].line < hits[j].line
	})
	fmt.Printf("scanned non-test Go files: %d\nhits (decl reads intents AND (mentions price OR selects *)): %d\n\n", scanned, len(hits))
	for _, h := range hits {
		mark := ""
		if h.star {
			mark = "  [select-star]"
		}
		fmt.Printf("%s:%d %s%s\n", h.file, h.line, h.decl, mark)
		for _, l := range h.priceLines {
			fmt.Printf("    %s\n", l)
		}
	}
}

// declName 은 선언의 이름 — 메서드는 「수신자.이름」, const/var 는 첫 이름.
func declName(d ast.Decl) string {
	switch v := d.(type) {
	case *ast.FuncDecl:
		if v.Recv != nil && len(v.Recv.List) > 0 {
			t := v.Recv.List[0].Type
			if s, ok := t.(*ast.StarExpr); ok {
				t = s.X
			}
			if id, ok := t.(*ast.Ident); ok {
				return id.Name + "." + v.Name.Name
			}
		}
		return v.Name.Name
	case *ast.GenDecl:
		for _, spec := range v.Specs {
			if vs, ok := spec.(*ast.ValueSpec); ok && len(vs.Names) > 0 {
				return v.Tok.String() + " " + vs.Names[0].Name
			}
		}
		return v.Tok.String()
	}
	return "?"
}
