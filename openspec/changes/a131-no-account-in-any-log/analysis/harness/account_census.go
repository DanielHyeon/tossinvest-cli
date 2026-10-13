//go:build ignore

// account_census 는 a131 의 유출 자리 전수 열거기임 — 비시험 Go 파일 전체를 go/parser 로 읽어,
// 계좌를 싣는 식(식별자·필드 이름이 계좌를 뜻하거나 obs.FieldAccount)이 **싱크 호출의 인자**로
// 들어가는 자리를 셈. 존재가 아니라 역할(싱크 인자인가)로 셈.
//
// 싱크 분류:
//
//	obs     — obs.Logger 의 Event/Warn/Error/OrderTransition/Reconcile/With 및 o.log/logErr 류 래퍼 호출,
//	          그리고 map 리터럴 키가 obs.FieldAccount 인 자리(알림 Fields → 로그 줄)
//	errtext — fmt.Errorf / fmt.Sprintf / errors.New 로 문자열에 계좌를 넣는 자리(어느 싱크로든 흘러갈 수 있음)
//	print   — fmt.Fprintf/Fprintln/Printf 로 쓰기 대상에 계좌를 직접 쓰는 자리
//	record  — json 태그가 계좌를 뜻하는 구조체 필드(기록 파일 직렬화)
//
// 판별은 이름 정규식 + 싱크 호출 인자 위치다(역할 판정이 아님 — a131 리뷰 R-6). 2판: DisplayName 추가 · 클로저 포함.
//
// 사용: go run account_census.go <repo-root>
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var accountName = regexp.MustCompile(`(?i)^(account|accountref|accountno|accountnumber|accountseq|accountmask|maskedaccount|currentref|recordmask|currentmask|displayname)$`)

type hit struct {
	file, kind, fn string
	line           int
	text           string
}

func main() {
	root := os.Args[1]
	fset := token.NewFileSet()
	var hits []hit
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		if info.IsDir() {
			base := info.Name()
			if base == ".git" || base == "openspec" || base == "node_modules" || base == ".sdd" || base == "tools" || base == "vendor" || strings.HasPrefix(base, ".") && rel != "." {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, perr := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if perr != nil {
			return nil
		}
		fnName := ""
		ast.Inspect(f, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.FuncDecl:
				fnName = x.Name.Name
				if x.Recv != nil && len(x.Recv.List) > 0 {
					fnName = recvName(x.Recv.List[0].Type) + "." + fnName
				}
			case *ast.Field:
				if x.Tag != nil && strings.Contains(strings.ToLower(x.Tag.Value), `json:"account`) {
					hits = append(hits, hit{rel, "record", fnName, fset.Position(x.Pos()).Line, x.Tag.Value})
				}
			case *ast.KeyValueExpr:
				if isFieldAccount(x.Key) {
					hits = append(hits, hit{rel, "obs", fnName, fset.Position(x.Pos()).Line, "map key obs.FieldAccount"})
				}
			case *ast.CallExpr:
				kind := sinkKind(x)
				if kind == "" {
					return true
				}
				for _, a := range x.Args {
					if carriesAccount(a) {
						hits = append(hits, hit{rel, kind, fnName, fset.Position(x.Pos()).Line, callName(x)})
						break
					}
				}
			}
			return true
		})
		return nil
	})
	sort.Slice(hits, func(i, j int) bool {
		if hits[i].kind != hits[j].kind {
			return hits[i].kind < hits[j].kind
		}
		if hits[i].file != hits[j].file {
			return hits[i].file < hits[j].file
		}
		return hits[i].line < hits[j].line
	})
	counts := map[string]int{}
	for _, h := range hits {
		counts[h.kind]++
		fmt.Printf("%s\t%s:%d\t%s\t%s\n", h.kind, h.file, h.line, h.fn, h.text)
	}
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("# total %s %d\n", k, counts[k])
	}
}

func recvName(e ast.Expr) string {
	switch t := e.(type) {
	case *ast.StarExpr:
		return recvName(t.X)
	case *ast.Ident:
		return t.Name
	case *ast.IndexExpr:
		return recvName(t.X)
	}
	return "?"
}

func isFieldAccount(e ast.Expr) bool {
	s, ok := e.(*ast.SelectorExpr)
	return ok && s.Sel.Name == "FieldAccount"
}

// carriesAccount 는 인자 식 안에 계좌를 뜻하는 이름이나 obs.FieldAccount 가 있는지 봄.
func carriesAccount(e ast.Expr) bool {
	found := false
	ast.Inspect(e, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.SelectorExpr:
			if x.Sel.Name == "FieldAccount" || accountName.MatchString(x.Sel.Name) {
				found = true
			}
		case *ast.Ident:
			if accountName.MatchString(x.Name) {
				found = true
			}
		}
		return !found
	})
	return found
}

func callName(c *ast.CallExpr) string {
	switch f := c.Fun.(type) {
	case *ast.SelectorExpr:
		if id, ok := f.X.(*ast.Ident); ok {
			return id.Name + "." + f.Sel.Name
		}
		return "…." + f.Sel.Name
	case *ast.Ident:
		return f.Name
	}
	return "?"
}

func sinkKind(c *ast.CallExpr) string {
	name := callName(c)
	switch name {
	case "fmt.Errorf", "fmt.Sprintf", "errors.New":
		return "errtext"
	case "fmt.Fprintf", "fmt.Fprintln", "fmt.Printf", "fmt.Println", "fmt.Fprint":
		return "print"
	}
	sel, ok := c.Fun.(*ast.SelectorExpr)
	if !ok {
		return ""
	}
	switch sel.Sel.Name {
	case "Event", "Warn", "Error", "OrderTransition", "Reconcile", "log", "logErr", "warn", "event", "logUnobserved":
		return "obs"
	case "With":
		return "obs"
	}
	return ""
}
