package engine_test

// a090 R15(tasks 2.15) — workingSet 포지션 순회의 이탈 전수 핀.
//
// 미관측 계수는 「표시(markHeld) 뒤에 판정 진입이 없음」 으로 셈. 그래서 표시 **앞**에 새 조기 탈락이 생기면 그 포지션은
// 조용히 안 세어짐(루프 머리 continue 교훈). 저장된 ast.json 에는 문장 순서 · continue 목록이 없으므로 현재 소스를
// go/parser 로 직접 읽어, 순회 본문의 모든 continue · return 을 **얼린 목록**과 다중집합으로 대조함. 좌표는 편집에 안정된
// 조상 경로(그 이탈까지의 if 조건 원문 — else 갈래는 "else " 접두) + 같은 경로의 출현 순번 + 이탈 종류.

import (
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// a090FrozenWorkingSetExits 는 설계 단계 FLM(workingSet 분기 22)의 이탈 열 개임.
var a090FrozenWorkingSetExits = []string{
	"continue @ p.State == journal.PositionClosed || isZeroQuantity(p.Quantity) #0", // B5 보유 아님
	"continue @ !p.ExitEligible() #0",                                               // B6 미관리(표시 앞)
	"continue @ !ok > err != nil #0",                                                // B8 열기 실패
	`continue @ !ok > opened.PositionID == "" #0`,                                   // B10 완료(해제 뒤)
	"continue @ result.Corruption != nil > qerr != nil #0",                          // B12 격리 쓰기 실패
	"continue @ result.Corruption != nil #0",                                        // B11 손상 → refused
	"continue @ qerr != nil #0",                                                     // B14 격리 읽기 실패
	"continue @ else qerr != nil > active && !q.NeedsReJudgement() #0",              // B17 격리 → refused
	"continue @ identityErr != nil > qerr != nil #0",                                // B21 격리 쓰기 실패
	"continue @ identityErr != nil #0",                                              // B20 신원 오류 → refused
}

type a090Exit struct {
	kind string
	path []string
}

func (e a090Exit) key(seq int) string {
	return e.kind + " @ " + strings.Join(e.path, " > ") + " #" + strconv.Itoa(seq)
}

// a090WorkingSetLoop 는 exitloop.go 의 ExitObserver.workingSet 에서 `for _, p := range positions` 본문을 돌려줌.
func a090WorkingSetLoop(t *testing.T, src []byte) (*token.FileSet, *ast.BlockStmt) {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "exitloop.go", src, 0)
	if err != nil {
		t.Fatalf("parsing exitloop.go: %v", err)
	}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "workingSet" || fn.Recv == nil {
			continue
		}
		for _, stmt := range fn.Body.List {
			if rng, ok := stmt.(*ast.RangeStmt); ok {
				if id, ok := rng.X.(*ast.Ident); ok && id.Name == "positions" {
					return fset, rng.Body
				}
			}
		}
	}
	t.Fatal("workingSet's position loop was not found")
	return nil, nil
}

func a090Src(fset *token.FileSet, n ast.Node) string {
	var b strings.Builder
	_ = printer.Fprint(&b, fset, n)
	return b.String()
}

// a090Exits 는 순회 본문의 모든 continue · return(함수 리터럴 안 제외)을 조상 if 경로와 함께 모음.
func a090Exits(fset *token.FileSet, body *ast.BlockStmt) []string {
	var exits []a090Exit
	var walkStmts func(stmts []ast.Stmt, path []string)
	var walk func(s ast.Stmt, path []string)
	walk = func(s ast.Stmt, path []string) {
		switch n := s.(type) {
		case *ast.BranchStmt:
			// continue 만이 아니라 break · goto 도 이탈임(리뷰 P2-5) — break 는 남은 포지션 전부를 판정과 계수에서 함께 뺌.
			exits = append(exits, a090Exit{kind: strings.ToLower(n.Tok.String()), path: append([]string(nil), path...)})
		case *ast.ExprStmt:
			if call, ok := n.X.(*ast.CallExpr); ok {
				if id, ok := call.Fun.(*ast.Ident); ok && id.Name == "panic" {
					exits = append(exits, a090Exit{kind: "panic", path: append([]string(nil), path...)})
				}
			}
		case *ast.ReturnStmt:
			exits = append(exits, a090Exit{kind: "return", path: append([]string(nil), path...)})
		case *ast.IfStmt:
			cond := a090Src(fset, n.Cond)
			walkStmts(n.Body.List, append(append([]string(nil), path...), cond))
			if n.Else != nil {
				elsePath := append(append([]string(nil), path...), "else "+cond)
				switch e := n.Else.(type) {
				case *ast.IfStmt:
					walk(e, elsePath)
				case *ast.BlockStmt:
					walkStmts(e.List, elsePath)
				}
			}
		case *ast.BlockStmt:
			walkStmts(n.List, path)
		case *ast.ForStmt, *ast.RangeStmt:
			// 안쪽 순회의 continue 는 이 순회의 이탈이 아님 — 지금 소스에는 없음. 생기면 여기서 드러나도록 셈에 넣음.
			ast.Inspect(n, func(x ast.Node) bool {
				if _, ok := x.(*ast.ReturnStmt); ok {
					exits = append(exits, a090Exit{kind: "return", path: append(append([]string(nil), path...), "nested-loop")})
				}
				return true
			})
		case *ast.SwitchStmt, *ast.TypeSwitchStmt, *ast.SelectStmt:
			ast.Inspect(n, func(x ast.Node) bool {
				switch b := x.(type) {
				case *ast.ReturnStmt:
					exits = append(exits, a090Exit{kind: "return", path: append(append([]string(nil), path...), "switch")})
				case *ast.BranchStmt:
					if b.Tok != token.BREAK { // switch 안의 break 는 switch 를 나갈 뿐 — continue · goto 는 순회 이탈
						exits = append(exits, a090Exit{kind: strings.ToLower(b.Tok.String()), path: append(append([]string(nil), path...), "switch")})
					}
				}
				return true
			})
		}
	}
	walkStmts = func(stmts []ast.Stmt, path []string) {
		for _, s := range stmts {
			walk(s, path)
		}
	}
	walkStmts(body.List, nil)

	seen := map[string]int{}
	out := make([]string, 0, len(exits))
	for _, e := range exits {
		base := e.kind + " @ " + strings.Join(e.path, " > ")
		out = append(out, e.key(seen[base]))
		seen[base]++
	}
	return out
}

func a090ReadExitloop(t *testing.T) []byte {
	t.Helper()
	src, err := os.ReadFile("exitloop.go")
	if err != nil {
		t.Fatalf("reading exitloop.go: %v", err)
	}
	return src
}

// a090CheckWorkingSet 는 src 의 workingSet 이 얼린 이탈 목록 · 표시 자리 · 해제 자리를 지키는지 보고, 어긋남을 문장으로 돌려줌.
func a090CheckWorkingSet(t *testing.T, src []byte) []string {
	t.Helper()
	fset, body := a090WorkingSetLoop(t, src)
	var problems []string

	got := a090Exits(fset, body)
	want := append([]string(nil), a090FrozenWorkingSetExits...)
	sort.Strings(got)
	sort.Strings(want)
	gotSet := map[string]int{}
	for _, g := range got {
		gotSet[g]++
	}
	for _, w := range want {
		if gotSet[w] == 0 {
			problems = append(problems, "frozen exit missing: "+w)
			continue
		}
		gotSet[w]--
	}
	for g, n := range gotSet {
		for ; n > 0; n-- {
			problems = append(problems, "unpinned exit: "+g)
		}
	}

	// 표시는 B6 블록 바로 뒤 형제 문장.
	marked := false
	for i, s := range body.List {
		ifs, ok := s.(*ast.IfStmt)
		if !ok || a090Src(fset, ifs.Cond) != "!p.ExitEligible()" {
			continue
		}
		if i+1 < len(body.List) && strings.TrimSpace(a090Src(fset, body.List[i+1])) == "o.markHeld(cycle, p)" {
			marked = true
		}
	}
	if !marked {
		problems = append(problems, "o.markHeld(cycle, p) is not the statement right after the B6 block")
	}

	// 해제는 B10 블록의 첫 문장.
	unmarked := false
	ast.Inspect(body, func(n ast.Node) bool {
		ifs, ok := n.(*ast.IfStmt)
		if !ok || a090Src(fset, ifs.Cond) != `opened.PositionID == ""` {
			return true
		}
		if len(ifs.Body.List) > 0 && strings.TrimSpace(a090Src(fset, ifs.Body.List[0])) == "o.unmarkHeld(cycle, p.ID)" {
			unmarked = true
		}
		return true
	})
	if !unmarked {
		problems = append(problems, "o.unmarkHeld(cycle, p.ID) is not the first statement of the B10 block")
	}
	return problems
}

func TestA090R15WorkingSetExitsAreTheFrozenTenAndTheMarkSitsBeforeThem(t *testing.T) {
	if problems := a090CheckWorkingSet(t, a090ReadExitloop(t)); len(problems) > 0 {
		t.Fatalf("workingSet exit census:\n  %s", strings.Join(problems, "\n  "))
	}
}

// 핀이 무는지: tasks 2.15 의 변이 다섯 + 리뷰 P2-5 의 break · panic 이 각각 빨강이어야 함.
func TestA090R15TheCensusCatchesTheFiveMutations(t *testing.T) {
	src := string(a090ReadExitloop(t))
	const mark = "\t\to.markHeld(cycle, p)\n"
	const unmark = "\t\t\t\to.unmarkHeld(cycle, p.ID)\n"
	const b6Tail = "\t\t\to.alertUnmanaged(ctx, p)\n\t\t\tcontinue\n\t\t}\n"
	const b8Continue = "\t\t\t\tif cycle.Err == nil {\n\t\t\t\t\tcycle.Err = err\n\t\t\t\t}\n\t\t\t\tcontinue\n"
	const b5Tail = "\t\tif p.State == journal.PositionClosed || isZeroQuantity(p.Quantity) {\n\t\t\tcontinue\n\t\t}\n"
	for _, needle := range []string{mark, unmark, b6Tail, b8Continue, b5Tail} {
		if strings.Count(src, needle) != 1 {
			t.Fatalf("mutation anchor not unique (count %d): %q", strings.Count(src, needle), needle)
		}
	}
	mutations := map[string]string{
		"1_mark_moved_after_B10": strings.Replace(strings.Replace(src, mark, "", 1),
			"\t\t\tcycle.Opened++\n", "\t\t\tcycle.Opened++\n"+mark, 1),
		"2_unmark_deleted": strings.Replace(src, unmark, "", 1),
		"3_continue_between_mark_and_B10": strings.Replace(src, mark,
			mark+"\t\tif p.Symbol == \"\" {\n\t\t\tcontinue\n\t\t}\n", 1),
		"4_early_exit_before_mark": strings.Replace(src, b5Tail,
			b5Tail+"\t\tif p.Market == \"\" {\n\t\t\tcontinue\n\t\t}\n", 1),
		"6_break_before_mark": strings.Replace(src, b5Tail,
			b5Tail+"\t\tif p.Market == \"\" {\n\t\t\tbreak\n\t\t}\n", 1),
		"7_panic_before_mark": strings.Replace(src, b5Tail,
			b5Tail+"\t\tif p.Market == \"\" {\n\t\t\tpanic(\"x\")\n\t\t}\n", 1),
		"5_return_in_B8": strings.Replace(src, b8Continue,
			"\t\t\t\tif cycle.Err == nil {\n\t\t\t\t\tcycle.Err = err\n\t\t\t\t\treturn nil, err\n\t\t\t\t}\n\t\t\t\tcontinue\n", 1),
	}
	for name, mutated := range mutations {
		if mutated == src {
			t.Fatalf("%s did not change the source", name)
		}
		if problems := a090CheckWorkingSet(t, []byte(mutated)); len(problems) == 0 {
			t.Errorf("mutation %s passed the census", name)
		}
	}
}
