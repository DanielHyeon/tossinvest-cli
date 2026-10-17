package riskbucket

// a066 6.1 (B)(2) — ApplyFill 의 target HELD 해석 가드 둘(B6 조기 검증, B32 이전 시점 재해석)은 서로를 가림:
// 하나만 끄면 다른 하나가 같은 거절 사유로 막아 변이가 살아남음(analysis/mutation-6.1/fill-gaps-ledger.tsv).
// 우연한 중복을 **선언된 층위**로 고정함(Manager 판정 (a) 2026-09-28):
//
//	① B6 가 1차 가드 — 이전 루프보다 앞에서, 함수 본문 최상위 `if len(event.TargetHeldMinor) != 0` 안에서
//	   event.ReservedMinor 의 **모든** key 를 걷고 해석 실패면 거절(return)함.
//	② B32 는 선언된 백스톱 — 이전 루프(event.ReservedMinor 의 key·값을 걷는 루프) 안에서 같은 값을 다시 해석하고 거절함.
//
// 행동 시험은 둘 중 먼저 서는 것만 보므로(TestA066ApplyFillTargetHeldShapeIsRefused) 위치·범위는 AST 로 단언함.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestA066ApplyFillTargetHeldGuardsAreLayered(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "fill.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var body *ast.BlockStmt
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "ApplyFill" && fn.Recv == nil {
			body = fn.Body
		}
	}
	if body == nil {
		t.Fatal("ApplyFill not found in fill.go")
	}
	early, transfer := -1, -1
	for i, stmt := range body.List {
		// ① 최상위 `if len(event.TargetHeldMinor) != 0 { … for key := range event.ReservedMinor { parse → return } }`.
		if ifs, ok := stmt.(*ast.IfStmt); ok && exprString(ifs.Cond) == "len(event.TargetHeldMinor) != 0" && ifs.Init == nil {
			for _, inner := range ifs.Body.List {
				// "모든 key 를 걷는다": 루프 본문의 첫 문장이 해석-거절이고 본문 어디에도 continue/break 가 없어야 함.
				if rng, ok := inner.(*ast.RangeStmt); ok && exprString(rng.X) == "event.ReservedMinor" && rng.Value == nil &&
					refusesUnparsedTargetHeld(rng.Body) && firstStmtRefuses(rng.Body) && !hasBranchStmt(rng.Body) {
					early = i
				}
			}
		}
		// ② 최상위 이전 루프 `for key, reservedRaw := range event.ReservedMinor { … parse target HELD → return … }`.
		if rng, ok := stmt.(*ast.RangeStmt); ok && exprString(rng.X) == "event.ReservedMinor" && rng.Value != nil &&
			refusesUnparsedTargetHeld(rng.Body) {
			transfer = i
		}
	}
	if early < 0 {
		t.Fatal("B6: no top-level early loop over every event.ReservedMinor key that parses TargetHeldMinor and refuses")
	}
	if transfer < 0 {
		t.Fatal("B32: the transfer loop no longer re-parses TargetHeldMinor as the declared backstop")
	}
	if early >= transfer {
		t.Fatalf("B6 (statement %d) must precede the transfer loop (statement %d)", early, transfer)
	}
}

func exprString(e ast.Expr) string {
	switch x := e.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.SelectorExpr:
		return exprString(x.X) + "." + x.Sel.Name
	case *ast.CallExpr:
		s := exprString(x.Fun) + "("
		for i, a := range x.Args {
			if i > 0 {
				s += ", "
			}
			s += exprString(a)
		}
		return s + ")"
	case *ast.BinaryExpr:
		return exprString(x.X) + " " + x.Op.String() + " " + exprString(x.Y)
	case *ast.BasicLit:
		return x.Value
	case *ast.IndexExpr:
		return exprString(x.X) + "[" + exprString(x.Index) + "]"
	}
	return "?"
}

// refusesUnparsedTargetHeld 는 블록 안에서 parseMinor(event.TargetHeldMinor[key], …) 의 오류 변수를 바로 `e != nil` 로
// 검사하고 그 갈래가 return 하는지 봄(`if _, err := parseMinor(…); err != nil { return … }` 또는 대입 다음 문장의 if).
func refusesUnparsedTargetHeld(block *ast.BlockStmt) bool {
	found := false
	ast.Inspect(block, func(n ast.Node) bool {
		var list []ast.Stmt
		switch node := n.(type) {
		case *ast.BlockStmt:
			list = node.List
		default:
			return !found
		}
		for i, stmt := range list {
			if ifs, ok := stmt.(*ast.IfStmt); ok && ifs.Init != nil {
				if name := targetHeldParseErr(ifs.Init); name != "" && exprString(ifs.Cond) == name+" != nil" && returnsInside(ifs.Body) {
					found = true
				}
			}
			if name := targetHeldParseErr(stmt); name != "" && i+1 < len(list) {
				if ifs, ok := list[i+1].(*ast.IfStmt); ok && ifs.Init == nil && exprString(ifs.Cond) == name+" != nil" && returnsInside(ifs.Body) {
					found = true
				}
			}
		}
		return !found
	})
	return found
}

// targetHeldParseErr 는 문장이 `_, e := parseMinor(event.TargetHeldMinor[key], …)` 모양이면 오류 변수 이름을 돌려줌.
func targetHeldParseErr(stmt ast.Stmt) string {
	assign, ok := stmt.(*ast.AssignStmt)
	if !ok || len(assign.Lhs) != 2 || len(assign.Rhs) != 1 {
		return ""
	}
	call, ok := assign.Rhs[0].(*ast.CallExpr)
	if !ok || exprString(call.Fun) != "parseMinor" || len(call.Args) == 0 || exprString(call.Args[0]) != "event.TargetHeldMinor[key]" {
		return ""
	}
	if id, ok := assign.Lhs[1].(*ast.Ident); ok && id.Name != "_" {
		return id.Name
	}
	return ""
}

func returnsInside(block *ast.BlockStmt) bool {
	found := false
	ast.Inspect(block, func(n ast.Node) bool {
		if _, ok := n.(*ast.ReturnStmt); ok {
			found = true
		}
		return !found
	})
	return found
}

// firstStmtRefuses 는 블록의 첫 문장이 target HELD 해석-거절 if 인지 봄.
func firstStmtRefuses(block *ast.BlockStmt) bool {
	if len(block.List) == 0 {
		return false
	}
	return refusesUnparsedTargetHeld(&ast.BlockStmt{List: block.List[:1]})
}

// hasBranchStmt 는 블록 안에 continue·break·goto 가 있는지 봄(중첩 함수 리터럴 제외).
func hasBranchStmt(block *ast.BlockStmt) bool {
	found := false
	ast.Inspect(block, func(n ast.Node) bool {
		if _, ok := n.(*ast.FuncLit); ok {
			return false
		}
		if _, ok := n.(*ast.BranchStmt); ok {
			found = true
		}
		return !found
	})
	return found
}
