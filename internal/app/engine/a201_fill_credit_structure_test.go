package engine_test

// a201_fill_credit_structure_test.go 는 design §3-11 의 구조를 AST 로 고정함 — 행동 시험이 못 보는 「자리」를 못 박음.
//
//   - RunOnce: 표식 읽기는 로컬 읽기 앞 1자리, 부여는 Observe 오류 반환 뒤 · 편입 판정 앞의 최상위 문장 1자리.
//   - Observe: credit 반박 호출은 둘(차단 갈래 기존 1 + 깨끗한 갈래 권위 오류 반환 직전 V4b 1), V4b 는 그 오류 갈래 몸
//     안에서 return 바로 앞에만 있음.
//   - 체결 credit 은 Tracker.adjusted 를 쓰는 기존 함수로만 쓰이고(새로 adjusted 를 만지는 함수는 부여 메서드 하나),
//     곁 맵 filled 는 부여 메서드와 Observe 의 원인 표지 한 문장에서만 쓰임 — 새 사용 · 반박 · 삭제 분기 0.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// a201Funcs 는 dir 의 시험 아닌 Go 파일에서 함수 선언을 「수신자.이름」 으로 모음.
func a201Funcs(t *testing.T, dir string) map[string]*ast.FuncDecl {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir(%s): %v", dir, err)
	}
	out := map[string]*ast.FuncDecl{}
	fset := token.NewFileSet()
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			key := fn.Name.Name
			if fn.Recv != nil && len(fn.Recv.List) == 1 {
				recv := fn.Recv.List[0].Type
				if star, ok := recv.(*ast.StarExpr); ok {
					recv = star.X
				}
				if id, ok := recv.(*ast.Ident); ok {
					key = id.Name + "." + key
				}
			}
			out[key] = fn
		}
	}
	return out
}

// a201Calls 는 node 안에서 이름(선택자의 끝 또는 식별자)이 name 인 호출 수를 셈.
func a201Calls(node ast.Node, name string) int {
	n := 0
	ast.Inspect(node, func(x ast.Node) bool {
		call, ok := x.(*ast.CallExpr)
		if !ok {
			return true
		}
		switch fun := call.Fun.(type) {
		case *ast.SelectorExpr:
			if fun.Sel.Name == name {
				n++
			}
		case *ast.Ident:
			if fun.Name == name {
				n++
			}
		}
		return true
	})
	return n
}

// a201Selects 는 node 안에서 선택자 이름이 field 인 식(t.adjusted 등) 수를 셈.
func a201Selects(node ast.Node, field string) int {
	n := 0
	ast.Inspect(node, func(x ast.Node) bool {
		if sel, ok := x.(*ast.SelectorExpr); ok && sel.Sel.Name == field {
			n++
		}
		return true
	})
	return n
}

// a201TopIndex 는 함수 몸의 최상위 문장 가운데 name 호출을 담은 첫 문장의 번호임(없으면 -1).
func a201TopIndex(fn *ast.FuncDecl, name string) int {
	for i, stmt := range fn.Body.List {
		if a201Calls(stmt, name) > 0 {
			return i
		}
	}
	return -1
}

func TestA201TheFillCreditIsGrantedAtOneSiteAfterASuccessfulObserve(t *testing.T) {
	fn := a201Funcs(t, ".")["ReconcileDriver.RunOnce"]
	if fn == nil {
		t.Fatal("ReconcileDriver.RunOnce not found")
	}
	for _, name := range []string{"MarkBeforeLocalRead", "FillsOutranAdjustment", "FillOutranSymbols"} {
		if got := a201Calls(fn, name); got != 1 {
			t.Fatalf("RunOnce calls %s %d times, want exactly one", name, got)
		}
	}
	mark, local := a201TopIndex(fn, "MarkBeforeLocalRead"), a201TopIndex(fn, "LocalStateFromJournal")
	// 바로 다음 문장이어야 함 — 표식과 로컬 읽기 사이가 벌어지면(예: stabilise 앞으로 옮김) 그 사이 체결이 전부 「표식 뒤」가 됨.
	if mark < 0 || local < 0 || local != mark+1 {
		t.Fatalf("the fill mark (statement %d) must be the statement right before the local read (statement %d)", mark, local)
	}
	observe, grant, judge := a201TopIndex(fn, "Observe"), a201TopIndex(fn, "FillsOutranAdjustment"), a201TopIndex(fn, "judgeHoldings")
	if observe < 0 || grant < 0 || judge < 0 || !(observe < grant && grant < judge) {
		t.Fatalf("statements observe=%d grant=%d judge=%d, want observe < grant < judge", observe, grant, judge)
	}
	if _, ok := fn.Body.List[grant].(*ast.ExprStmt); !ok {
		t.Fatalf("the grant is %T, want a top-level call statement (no new branch)", fn.Body.List[grant])
	}
	// Observe 와 부여 사이에 Observe 오류로 돌아가는 갈래가 있어야 함 — 오류 주기에는 부여하지 않음(V3).
	returns := false
	for _, stmt := range fn.Body.List[observe+1 : grant] {
		if ifs, ok := stmt.(*ast.IfStmt); ok {
			ast.Inspect(ifs.Body, func(x ast.Node) bool {
				if _, ok := x.(*ast.ReturnStmt); ok {
					returns = true
				}
				return true
			})
		}
	}
	if !returns {
		t.Fatal("no Observe-error return between the Observe call and the grant")
	}
}

func TestA201TheFillCreditRidesTheAdjustmentCreditsOwnBranches(t *testing.T) {
	funcs := a201Funcs(t, filepath.Join("..", "..", "reconcile"))
	observe := funcs["Tracker.Observe"]
	if observe == nil {
		t.Fatal("Tracker.Observe not found")
	}

	// 반박 호출은 둘 — V4b 는 깨끗한 갈래의 권위 오류 갈래 몸 안, return 바로 앞.
	if got := a201Calls(observe, "refuteUsableAdjustmentCredits"); got != 2 {
		t.Fatalf("Observe calls refuteUsableAdjustmentCredits %d times, want 2", got)
	}
	var authority *ast.IfStmt
	ast.Inspect(observe.Body, func(x ast.Node) bool {
		if ifs, ok := x.(*ast.IfStmt); ok && authority == nil && ifs.Init != nil &&
			a201Calls(ifs.Init, "resolvePendingPermanentAuthority") > 0 {
			authority = ifs
			return false
		}
		return true
	})
	if authority == nil {
		t.Fatal("the clean branch's authority check not found")
	}
	body := authority.Body.List
	if len(body) != 2 || a201Calls(body[0], "refuteUsableAdjustmentCredits") != 1 {
		t.Fatalf("the clean branch's authority-error body has %d statements, want the refutation then the return", len(body))
	}
	if _, ok := body[1].(*ast.ReturnStmt); !ok {
		t.Fatalf("the authority-error body ends with %T, want the return", body[1])
	}

	// adjusted 를 만지는 함수 집합: 기존 다섯 + 부여 메서드. filled 는 부여 메서드와 Observe 뿐.
	var adjusted, filled []string
	for key, fn := range funcs {
		if a201Selects(fn, "adjusted") > 0 {
			adjusted = append(adjusted, key)
		}
		if a201Selects(fn, "filled") > 0 {
			filled = append(filled, key)
		}
	}
	sort.Strings(adjusted)
	sort.Strings(filled)
	wantAdjusted := "Tracker.AdjustmentApplied,Tracker.FillsOutranAdjustment,Tracker.Observe,Tracker.Resolve,Tracker.Restore,Tracker.refuteUsableAdjustmentCredits"
	if got := strings.Join(adjusted, ","); got != wantAdjusted {
		t.Fatalf("functions touching adjusted = %s, want %s", got, wantAdjusted)
	}
	if got := strings.Join(filled, ","); got != "Tracker.FillsOutranAdjustment,Tracker.Observe" {
		t.Fatalf("functions touching filled = %s, want the grant and Observe only", got)
	}

	// Observe 안에서 filled 는 원인 표지 한 문장에만, 그 문장 밖의 adjusted 사용 수는 편집 전과 같은 3.
	var tags []ast.Stmt
	ast.Inspect(observe.Body, func(x ast.Node) bool {
		if assign, ok := x.(*ast.AssignStmt); ok && a201Selects(assign, "filled") > 0 {
			tags = append(tags, assign)
		}
		return true
	})
	if len(tags) != 1 || a201Selects(observe, "filled") != a201Selects(tags[0], "filled") {
		t.Fatalf("Observe uses filled outside one cause-tag assignment (%d assignments)", len(tags))
	}
	tag := tags[0].(*ast.AssignStmt)
	if sel, ok := tag.Lhs[0].(*ast.SelectorExpr); !ok || sel.Sel.Name != "releasedByFill" {
		t.Fatal("the filled assignment in Observe must set the block's cause tag only")
	}
	if got := a201Selects(observe, "adjusted") - a201Selects(tag, "adjusted"); got != 3 {
		t.Fatalf("Observe uses adjusted %d times outside the cause tag, want the 3 it had before", got)
	}
}

// TestA201ARefreshFailureRefutesOnlyOnItsOwnReturnPath: V4c(gstack r1 G1) 의 자리 — 반박 호출은 RunOnce 의 Refresh 오류
// 갈래 몸 안에서 return 바로 앞 1자리뿐(오류 없는 경로 무변), 새 메서드는 추적기 잠금 아래에서 기존 반박 함수만 부름
// (잠금 없이 부르는 변이는 -race 가 못 잡을 수 있어 구조로 고정).
func TestA201ARefreshFailureRefutesOnlyOnItsOwnReturnPath(t *testing.T) {
	run := a201Funcs(t, ".")["ReconcileDriver.RunOnce"]
	if run == nil {
		t.Fatal("ReconcileDriver.RunOnce not found")
	}
	if got := a201Calls(run, "RefuteCreditsDisputedBy"); got != 1 {
		t.Fatalf("RunOnce calls RefuteCreditsDisputedBy %d times, want exactly one", got)
	}
	var refresh *ast.IfStmt
	for _, stmt := range run.Body.List {
		if ifs, ok := stmt.(*ast.IfStmt); ok && ifs.Init != nil && a201Calls(ifs.Init, "Refresh") > 0 {
			refresh = ifs
		}
	}
	if refresh == nil {
		t.Fatal("the Refresh-error branch of RunOnce not found")
	}
	body := refresh.Body.List
	if len(body) < 2 || a201Calls(body[len(body)-2], "RefuteCreditsDisputedBy") != 1 {
		t.Fatal("the refutation must be the statement right before the Refresh-error return")
	}
	if _, ok := body[len(body)-1].(*ast.ReturnStmt); !ok {
		t.Fatalf("the Refresh-error branch ends with %T, want the return", body[len(body)-1])
	}

	method := a201Funcs(t, filepath.Join("..", "..", "reconcile"))["Tracker.RefuteCreditsDisputedBy"]
	if method == nil {
		t.Fatal("Tracker.RefuteCreditsDisputedBy not found")
	}
	if a201Calls(method, "Lock") != 1 || a201Calls(method, "Unlock") != 1 ||
		a201Calls(method, "refuteUsableAdjustmentCredits") != 1 || a201Calls(method, "symbolsInDispute") != 1 {
		t.Fatal("RefuteCreditsDisputedBy must take the tracker lock and call only the existing refutation over the comparison's disputes")
	}
}
