//go:build ignore

// flow 는 a087 tasks 3.7 ② 의 열거기임 — 저널 접근자(LookupIntent·LiveOrdersForSymbol·scanIntent)의 결과를 받은
// 변수를, 그것을 부르는 비시험 함수 안에서 go/ast 로 따라가 그 변수(와 range 로 꺼낸 원소)가 쓰이는 자리를 전부 셈.
// 각 자리는 「필드 읽기(.Price 여부)」·「통째 전달(호출 인자·반환·대입)」 중 하나로 분류됨. 통째 전달은 그
// 받는 쪽에서 다시 따라가야 하는 자리라 README 에 그 처분을 적음.
//
// 호출 자리는 이 도구가 고르지 않음 — 저장소 비시험 Go 전체에서 접근자 이름의 호출식을 찾음(CodeGraph 결과와
// 대조는 README 에). 사용(저장소 루트에서):
//
//	go run ./openspec/changes/a087-a-protective-exit-is-a-market-order/analysis/intents-price-null/flow.go
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// accessors 는 intents.price 를 읽는 질의를 감싼 접근자 — census.go 의 열거 결과에서 옴.
var accessors = map[string]bool{"LookupIntent": true, "LiveOrdersForSymbol": true, "scanIntent": true}

// follow 는 첫 단계 출력에서 값을 통째로 받은 함수 — 「함수 이름 → 받은 매개변수 이름」. 둘째 단계가 그 매개변수를
// 같은 방식으로 따라감. 둘째 단계 출력에 또 통째 전달이 나오면 여기에 더해 다시 돌림(고정점까지).
var follow = map[string]string{
	"clearTargets": "live", "resolveCancel": "intent", "resolveAmend": "intent", "newMatcher": "intent",
	"absenceCorroborated": "intent", "parkAlert": "intent", "windowContaminated": "intent", "intentNotional": "intent",
}

type use struct {
	line int
	kind string // field / pass
	text string
}

func main() {
	type site struct {
		file, fn string
		line     int
		accessor string
		uses     []use
	}
	var sites []site
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
		f, perr := parser.ParseFile(fset, path, nil, 0)
		if perr != nil {
			return perr
		}
		for _, decl := range f.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Body == nil {
				continue
			}
			if param, ok := follow[fd.Name.Name]; ok {
				for _, field := range fd.Type.Params.List {
					for _, name := range field.Names {
						if name.Name != param || name.Obj == nil {
							continue
						}
						tracked := map[*ast.Object]bool{name.Obj: true}
						trackRange(fd.Body, tracked)
						sites = append(sites, site{file: filepath.ToSlash(path), fn: funcName(fd),
							line: fset.Position(fd.Pos()).Line, accessor: "param " + param,
							uses: collectUses(fset, fd.Body, tracked, nil)})
					}
				}
			}
			ast.Inspect(fd.Body, func(n ast.Node) bool {
				as, ok := n.(*ast.AssignStmt)
				if !ok || len(as.Rhs) != 1 {
					return true
				}
				call, ok := as.Rhs[0].(*ast.CallExpr)
				if !ok {
					return true
				}
				name := calleeName(call.Fun)
				if !accessors[name] {
					return true
				}
				id, ok := as.Lhs[0].(*ast.Ident)
				if !ok || id.Obj == nil {
					return true
				}
				s := site{file: filepath.ToSlash(path), fn: funcName(fd), line: fset.Position(call.Pos()).Line, accessor: name}
				tracked := map[*ast.Object]bool{id.Obj: true}
				trackRange(fd.Body, tracked)
				s.uses = collectUses(fset, fd.Body, tracked, as)
				sites = append(sites, s)
				return true
			})
		}
		return nil
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	sort.Slice(sites, func(i, j int) bool {
		if sites[i].file != sites[j].file {
			return sites[i].file < sites[j].file
		}
		return sites[i].line < sites[j].line
	})
	fmt.Printf("accessor call sites + followed params (non-test): %d\n\n", len(sites))
	for _, s := range sites {
		fmt.Printf("%s:%d %s ← %s\n", s.file, s.line, s.fn, s.accessor)
		for _, u := range s.uses {
			fmt.Printf("    %d %-5s %s\n", u.line, u.kind, u.text)
		}
	}
}

// trackRange 는 range 로 꺼낸 원소도 같은 값으로 봄 — 두 번 돌아 원소 변수까지 추적함.
func trackRange(body *ast.BlockStmt, tracked map[*ast.Object]bool) {
	for pass := 0; pass < 2; pass++ {
		ast.Inspect(body, func(m ast.Node) bool {
			rs, ok := m.(*ast.RangeStmt)
			if !ok {
				return true
			}
			if x, ok := rs.X.(*ast.Ident); ok && x.Obj != nil && tracked[x.Obj] {
				if v, ok := rs.Value.(*ast.Ident); ok && v.Obj != nil {
					tracked[v.Obj] = true
				}
			}
			return true
		})
	}
}

// collectUses 는 추적 변수의 모든 등장을 부모 문맥으로 분류함. 대입문 자신의 좌변은 제외함.
func collectUses(fset *token.FileSet, body *ast.BlockStmt, tracked map[*ast.Object]bool, origin *ast.AssignStmt) []use {
	var out []use
	var stack []ast.Node
	ast.Inspect(body, func(n ast.Node) bool {
		if n == nil {
			stack = stack[:len(stack)-1]
			return false
		}
		defer func() { stack = append(stack, n) }()
		id, ok := n.(*ast.Ident)
		if !ok || id.Obj == nil || !tracked[id.Obj] || len(stack) == 0 {
			return true
		}
		parent := stack[len(stack)-1]
		if parent == origin {
			return true
		}
		if rs, ok := parent.(*ast.RangeStmt); ok && (rs.Value == id || rs.Key == id) {
			return true
		}
		line := fset.Position(id.Pos()).Line
		if sel, ok := parent.(*ast.SelectorExpr); ok && sel.X == id {
			// 필드 읽기 — 한 단계 위가 대입 좌변이면 쓰기지만 여기서는 읽기·쓰기 모두 적음(문맥은 text).
			ctx := sel.Sel.Name
			if len(stack) >= 2 {
				ctx = render(fset, stack[len(stack)-2])
			}
			out = append(out, use{line, "field", sel.Sel.Name + " :: " + clip(ctx)})
			return true
		}
		out = append(out, use{line, "pass", clip(render(fset, parent))})
		return true
	})
	return out
}

func render(fset *token.FileSet, n ast.Node) string {
	var b bytes.Buffer
	_ = printer.Fprint(&b, fset, n)
	return strings.Join(strings.Fields(b.String()), " ")
}

func clip(s string) string {
	if len(s) > 160 {
		return s[:160] + "…"
	}
	return s
}

func calleeName(e ast.Expr) string {
	switch v := e.(type) {
	case *ast.SelectorExpr:
		return v.Sel.Name
	case *ast.Ident:
		return v.Name
	}
	return ""
}

func funcName(fd *ast.FuncDecl) string {
	if fd.Recv != nil && len(fd.Recv.List) > 0 {
		t := fd.Recv.List[0].Type
		if s, ok := t.(*ast.StarExpr); ok {
			t = s.X
		}
		if id, ok := t.(*ast.Ident); ok {
			return id.Name + "." + fd.Name.Name
		}
	}
	return fd.Name.Name
}
