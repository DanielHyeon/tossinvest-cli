package journal

// a066 6.1 — BTM 행 변이 생존: loadRiskBucketFillTransition B18(결정마다 다섯 bucket 인지 보는 루프)을 꺼도 초록이었음.
// 결정의 예약 행이 빠지면 체결 계상 앞의 verifyRiskBucketStateDigest(상태 재구성의 "bucket count")가 먼저 거절하므로
// B18·B19(전체 bucket 수) 는 그 뒤의 백스톱임 — 변이가 엄폐를 실증했으므로(Manager 규칙) 층위를 AST 로 못 박고,
// 결과(체결은 남고 owner scope 가 REPLAY_MISMATCH 로 잠김)를 행동으로 봄. 행동 시험은 넷째 층(riskbucket.ApplyFill 의
// validateFillBuckets "bucket_count")까지 있어 앞 셋을 모두 꺼도 초록임(fill-layers-ledger F4) — 그래서 셋을 자리로 고정함.

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestA066FillWithAMissingDecisionReservationLatchesAndKeepsTheFill(t *testing.T) {
	ctx := context.Background()
	j, key, decisionID, reserved := riskBucketFillFixture(t, "missing-reservation", "risk-missing-reservation")
	if err := j.RegisterRiskBucketOrder(ctx, RiskBucketOrderPlan{OrderID: "risk-missing-reservation", DecisionID: decisionID, OrderQuantity: 10, ReservedMinor: reserved, CreatedAt: riskFillNow}); err != nil {
		t.Fatal(err)
	}
	// 손상 재현: 주문 등록이 이미 이 예약을 참조하므로 외래 키를 잠시 끄고 지움(단일 연결이라 PRAGMA 가 그 연결에 걸림).
	for _, stmt := range []string{`PRAGMA foreign_keys=OFF`, `DELETE FROM risk_bucket_reservations WHERE decision_id='` + decisionID + `' AND bucket_dimension='sector'`, `PRAGMA foreign_keys=ON`} {
		if _, err := j.db.Exec(stmt); err != nil {
			t.Fatalf("%s: %v", stmt, err)
		}
	}
	result, err := j.RecordFill(ctx, observation("risk-missing-reservation", "2"))
	if err != nil || !result.Changed {
		t.Fatalf("the broker fill must be kept: result=%+v err=%v", result, err)
	}
	var latches int
	if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_scope_latches WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=? AND latch='REPLAY_MISMATCH'`,
		key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration).Scan(&latches); err != nil {
		t.Fatal(err)
	}
	if latches != 1 {
		t.Fatalf("REPLAY_MISMATCH scope latches=%d, want 1", latches)
	}
}

func TestA066FillTransitionBucketCountGuardsAreLayered(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "risk_bucket_fill.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	callers := 0
	var loader *ast.FuncDecl
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		if fn.Name.Name == "loadRiskBucketFillTransition" {
			loader = fn
			continue
		}
		if !callsAnywhere(fn.Body, "loadRiskBucketFillTransition") {
			continue
		}
		// ① 호출자마다 상태 digest 대조가 직선 경로에서 체결 계상 적재보다 앞.
		callers++
		digestAt, loadAt := -1, -1
		for i, stmt := range fn.Body.List {
			if digestAt < 0 && callsOnStraightPath(stmt, "verifyRiskBucketStateDigest") {
				digestAt = i
			}
			if loadAt < 0 && callsOnStraightPath(stmt, "loadRiskBucketFillTransition") {
				loadAt = i
			}
		}
		if digestAt < 0 || loadAt < 0 || digestAt >= loadAt {
			t.Errorf("%s: %s — verifyRiskBucketStateDigest (stmt %d) must precede loadRiskBucketFillTransition (stmt %d) on the straight path",
				fset.Position(fn.Pos()), fn.Name.Name, digestAt, loadAt)
		}
	}
	if callers != 2 {
		t.Errorf("loadRiskBucketFillTransition callers=%d, want 2 (applyRiskBucketFillInTx, completeRiskBucketFillActual)", callers)
	}
	if loader == nil {
		t.Fatal("loadRiskBucketFillTransition not found")
	}
	// ② 백스톱 둘: 결정마다 bucket 수(range decisionBuckets → return) 와 전체 bucket 수(len(state.Buckets) 대조 → return).
	perDecision, total := false, false
	for _, stmt := range loader.Body.List {
		if rng, ok := stmt.(*ast.RangeStmt); ok {
			if id, ok := rng.X.(*ast.Ident); ok && id.Name == "decisionBuckets" && !hasBranchStmtJ(rng.Body) && len(rng.Body.List) > 0 {
				// 모든 결정을 걸어야 함: 첫 문장이 비교-거절이고 분기문이 없어야 함.
				for _, inner := range rng.Body.List[:1] {
					// 조건은 정확히 `len(seen) != …` 이어야 함(`false && …` 로 꺼진 조건은 가드가 아님).
					if ifs, ok := inner.(*ast.IfStmt); ok && isNeqOfLen(ifs.Cond, "seen") && bodyReturns(ifs.Body) {
						perDecision = true
					}
				}
			}
		}
		if ifs, ok := stmt.(*ast.IfStmt); ok && bodyReturns(ifs.Body) {
			if bin, ok := ifs.Cond.(*ast.BinaryExpr); ok && bin.Op == token.LOR && isNeqOfLenSelector(bin.Y, "state.Buckets") {
				total = true
			}
		}
	}
	if !perDecision {
		t.Error("② the per-decision bucket-count backstop loop is gone")
	}
	if !total {
		t.Error("② the total bucket-count backstop is gone")
	}
}

func bodyReturns(block *ast.BlockStmt) bool {
	if len(block.List) == 0 {
		return false
	}
	_, ok := block.List[len(block.List)-1].(*ast.ReturnStmt)
	return ok
}

// isNeqOfLen 은 식이 정확히 `len(<name>) != …` 인지 봄.
func isNeqOfLen(e ast.Expr, name string) bool {
	bin, ok := e.(*ast.BinaryExpr)
	if !ok || bin.Op != token.NEQ {
		return false
	}
	call, ok := bin.X.(*ast.CallExpr)
	if !ok || len(call.Args) != 1 {
		return false
	}
	fn, ok := call.Fun.(*ast.Ident)
	arg, ok2 := call.Args[0].(*ast.Ident)
	return ok && ok2 && fn.Name == "len" && arg.Name == name
}

// isNeqOfLenSelector 는 식이 정확히 `len(a.b) != …` 인지 봄.
func isNeqOfLenSelector(e ast.Expr, name string) bool {
	bin, ok := e.(*ast.BinaryExpr)
	return ok && bin.Op == token.NEQ && mentionsLenOf(bin.X, name)
}

func returnsFromIf(block *ast.BlockStmt) bool {
	for _, stmt := range block.List {
		if ifs, ok := stmt.(*ast.IfStmt); ok && len(ifs.Body.List) > 0 {
			if _, ok := ifs.Body.List[0].(*ast.ReturnStmt); ok {
				return true
			}
		}
	}
	return false
}

// mentionsLenOf 는 조건 안에 len(<name>) 이 있는지 봄(name 은 `a.b` 모양).
func mentionsLenOf(cond ast.Expr, name string) bool {
	found := false
	ast.Inspect(cond, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || len(call.Args) != 1 {
			return true
		}
		if fn, ok := call.Fun.(*ast.Ident); ok && fn.Name == "len" {
			if sel, ok := call.Args[0].(*ast.SelectorExpr); ok {
				if x, ok := sel.X.(*ast.Ident); ok && x.Name+"."+sel.Sel.Name == name {
					found = true
				}
			}
		}
		return !found
	})
	return found
}
