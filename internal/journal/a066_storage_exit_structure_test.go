package journal

// a066 6.1 (B)(3) — 저장 오류 출구의 fail-closed 를 행별 RED 대신 **구조**로 고정함.
//
// not-applicable(행별 RED): 저장소·드라이버 오류 출구(`if err != nil { return … }`)를 시험에서 빨갛게 만들려면
// 생산 코드에 시험 전용 fault-injection seam 을 심어야 함 — 이 change 범위 밖(Manager 판정 2026-09-28, 비례 원칙).
// 대신 AST 로 두 성질을 단언함:
//
//	P1 오류 출구의 모든 return 은 오류 자리(마지막 결과)에 nil 이 아닌 값을 돌려줌 — 오류를 삼켜 성공처럼 보이지 않음.
//	P2 오류 출구 안에서 쓰기(ExecContext·Exec·Commit)가 없음 — 실패한 뒤 절반만 쓴 상태를 커밋하지 않음. 트랜잭션을
//	   직접 여는 함수는 BeginTx 바로 뒤 `defer tx.Rollback()` 을 가져야 함(출구가 롤백에 맡기는 근거).
//
// 대상: internal/journal/risk_bucket*.go(비시험)의 모든 함수 + (*Journal).RecordFill + riskbucket 생산 snapshot reader 파일
// 전체 + execgw Gateway.checkReservation·submit. 대상 집합과 출구 수를 census 로
// 고정함 — 새 파일·새 출구가 생기면 숫자가 달라져 빨개지고, 새 출구는 같은 두 성질로 자동 검사됨.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// census 는 이 시험이 걸은 범위임. 값이 바뀌면 새 출구·새 함수가 검사 범위에 들어온 것이므로 숫자를 갱신하기 전에
// 새 출구가 두 성질을 지키는지(이 시험이 이미 검사함)와 대상 파일 목록이 의도대로인지 확인할 것.
// others 는 `err != nil` 이지만 err 가 저장소 호출에서 오지 않는 출구(해석·의미 판정) — 이 시험의 두 성질 밖이고
// 세기만 함(범위가 조용히 줄지 않게).
var a066StorageExitCensus = struct {
	files, funcs, exits, others, txOpeners int
}{files: 11, funcs: 128, exits: 340, others: 91, txOpeners: 13}

// a066ProductionSnapshotStorageExits 는 걷는 범위 중 riskbucket 생산 snapshot reader 의 저장소 오류 출구를 **함수별 이름으로** 얼린다.
// 336 → 337(2026-10-01, a112 5.2.2.2 리뷰 수리 `face8d0d` — J4 = (A)): `loadProductionRiskEntries` 의 scope latch 조회가 `err != nil ||
// latches != 0` 한 갈래에서 둘로 갈렸다 — 조회 결함(`scope latch unreadable: %w`, 이 새 저장소 출구 — 결함이라 nil 아닌 오류 · 쓰기 0)과 latch
// 존재(범위 국소 sentinel, 저장소 출구 아님). 편집 전에는 합쳐진 조건이라 `err != nil` 출구로 세어지지 않았다. 정당한 이유: 원장 조회 결함을
// 범위 국소 거절과 가르려면 그 결함이 자기 출구를 가져야 한다(a112 J4).
// 337 → 340 · txOpeners 12 → 13(2026-10-01, a127 D1 · D7): `loadProductionRiskEntries` 가 판독을 읽기 전용 tx 하나로 묶고(BeginTx 출구 +1, 트랜잭션
// 여는 함수 +1 — `defer tx.Rollback()` 이 BeginTx 검사 바로 뒤) user_version 판독 실패를 버전 불일치와 가르며(+1) 원장 데이터 질의를 판독 전에
// prepare 함(+1). 세 출구 모두 nil 아닌 오류 · 쓰기 0(이 시험이 검사), 대상 파일 목록 불변.
var a066ProductionSnapshotStorageExits = map[string]int{
	"LoadProductionRiskSnapshotAuthority": 2,
	"loadProductionRiskEntries":           6, // a112 5.2.2.2: +1 「scope latch unreadable」 · a127: +3(읽기 tx BeginTx · user_version 판독 · 판독 전 prepare)
	"ReadJournalBucketUsage":              1,
	"readProductionRiskUsage":             2,
}

func TestA066StorageErrorExitsFailClosed(t *testing.T) {
	names, err := filepath.Glob("risk_bucket*.go")
	if err != nil {
		t.Fatal(err)
	}
	var files []string
	for _, name := range names {
		if !strings.HasSuffix(name, "_test.go") {
			files = append(files, name)
		}
	}
	files = append(files, "fills.go")
	sort.Strings(files)
	// BTM 의 저장 출구 행이 있는 다른 패키지 두 파일도 같은 두 성질로 걸음: 생산 snapshot reader 전체와 Gateway 의
	// a066 번들 함수 둘(checkReservation · submit).
	files = append(files, "../riskbucket/production_snapshot_authority.go", "../execgw/gateway.go")

	fset := token.NewFileSet()
	var funcs, exits, others, txOpeners int
	productionExits := map[string]int{}
	for _, name := range files {
		src, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		file, err := parser.ParseFile(fset, name, src, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			// fills.go 에서는 a066 이 계상 경로를 붙인 RecordFill 하나만 봄(나머지는 이 change 의 함수가 아님).
			if name == "fills.go" && fn.Name.Name != "RecordFill" {
				continue
			}
			if name == "../execgw/gateway.go" && fn.Name.Name != "checkReservation" && fn.Name.Name != "submit" {
				continue
			}
			funcs++
			if opensTx(fn) {
				txOpeners++
				if !rollbackDeferredAfterBegin(fn) {
					t.Errorf("%s: %s opens a transaction without `defer tx.Rollback()` right after BeginTx", fset.Position(fn.Pos()), fn.Name.Name)
				}
			}
			// 블록 안에서 `if … err != nil` 을 찾고, 그 err 를 만든 호출(if 의 init 문이나 바로 앞 문장)이 저장소 호출인지 가름.
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				// 문장 목록은 블록 말고 switch/select 의 case 절에도 있음 — 둘 다 걸어야 case 안의 출구가 빠지지 않음.
				var list []ast.Stmt
				switch node := n.(type) {
				case *ast.BlockStmt:
					list = node.List
				case *ast.CaseClause:
					list = node.Body
				case *ast.CommClause:
					list = node.Body
				default:
					return true
				}
				for i, s := range list {
					// `switch { case err == nil: … default: return …, err }` 의 default 도 저장소 오류 출구임 — err 출처는
					// switch 앞에서 err 를 마지막으로 대입한 문장.
					if sw, ok := s.(*ast.SwitchStmt); ok && sw.Tag == nil {
						for _, clause := range sw.Body.List {
							cc, ok := clause.(*ast.CaseClause)
							if !ok || cc.List != nil {
								continue
							}
							var source ast.Stmt = sw.Init
							for k := i - 1; source == nil && k >= 0; k-- {
								if assignsErr(list[k]) {
									source = list[k]
								}
							}
							if assignsErrFromStorageCall(source) {
								exits++
								if name == "../riskbucket/production_snapshot_authority.go" {
									productionExits[fn.Name.Name]++
								}
								if os.Getenv("A066_LIST_STORAGE_EXITS") != "" {
									t.Logf("storage exit %s %s (switch default)", fset.Position(cc.Pos()), fn.Name.Name)
								}
								checkStorageExit(t, fset, fn.Name.Name, cc.Pos(), &ast.BlockStmt{List: cc.Body})
							}
						}
					}
					stmt, ok := s.(*ast.IfStmt)
					// `} else if err := f(); err != nil {` 사슬도 출구임 — else 쪽 IfStmt 는 블록 목록에 없으므로 사슬을 따라감.
					// 사슬의 else-if 가 자기 init 없이 앞 if 의 init 이 만든 err 를 다시 보면(`if err := q.Scan(); err == nil {…}
					// else if !errors.Is(err, sql.ErrNoRows) {…}`) 그 init 이 출처임.
					var chainInit ast.Stmt
					for first := true; ok; stmt, ok = stmt.Else.(*ast.IfStmt) {
						if stmt.Init != nil {
							chainInit = stmt.Init
						}
						if isErrNotNil(stmt.Cond) {
							var source ast.Stmt = stmt.Init
							if source == nil && !first {
								source = chainInit
							}
							// init 이 없으면 err 를 마지막으로 대입한 앞 문장을 거슬러 찾음 — 사이에 `if errors.Is(err, sql.ErrNoRows)`
							// 같은 갈래가 끼어도 err 의 출처는 그 앞의 대입임.
							for k := i - 1; source == nil && first && k >= 0; k-- {
								if assignsErr(list[k]) {
									source = list[k]
								}
							}
							if assignsErrFromStorageCall(source) {
								exits++
								if name == "../riskbucket/production_snapshot_authority.go" {
									productionExits[fn.Name.Name]++
								}
								if os.Getenv("A066_LIST_STORAGE_EXITS") != "" {
									t.Logf("storage exit %s %s", fset.Position(stmt.Pos()), fn.Name.Name)
								}
								checkStorageExit(t, fset, fn.Name.Name, stmt.Pos(), stmt.Body)
							} else {
								others++
								if os.Getenv("A066_LIST_OTHER_EXITS") != "" {
									t.Logf("other exit %s %s", fset.Position(stmt.Pos()), fn.Name.Name)
								}
							}
						}
						first = false
					}
				}
				return true
			})
		}
	}
	if !reflect.DeepEqual(productionExits, a066ProductionSnapshotStorageExits) {
		t.Errorf("riskbucket production snapshot storage exits by function = %v, census %v — name the new exit", productionExits,
			a066ProductionSnapshotStorageExits)
	}
	got := struct{ files, funcs, exits, others, txOpeners int }{len(files), funcs, exits, others, txOpeners}
	if got != a066StorageExitCensus {
		t.Errorf("walked scope changed: got files=%d funcs=%d exits=%d others=%d txOpeners=%d, census %+v (files %v)",
			got.files, got.funcs, got.exits, got.others, got.txOpeners, a066StorageExitCensus, files)
	}
}

// checkStorageExit 는 저장소 오류 출구 하나에 P1·P2 를 단언함.
func checkStorageExit(t *testing.T, fset *token.FileSet, fn string, pos token.Pos, body ast.Node) {
	t.Helper()
	where := fset.Position(pos)
	// P2: 출구 본문 안의 쓰기 호출.
	ast.Inspect(body, func(m ast.Node) bool {
		if call, ok := m.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				switch sel.Sel.Name {
				case "ExecContext", "Exec", "Commit":
					t.Errorf("%s: %s storage error exit writes (%s) before returning", where, fn, sel.Sel.Name)
				}
			}
		}
		return true
	})
	// P1: 출구 본문의 모든 return 이 nil 아닌 오류를 돌려줌. 중첩 함수 리터럴의 return 은 이 함수의 출구가 아님.
	returns := 0
	ast.Inspect(body, func(m ast.Node) bool {
		if _, ok := m.(*ast.FuncLit); ok {
			return false
		}
		ret, ok := m.(*ast.ReturnStmt)
		if !ok {
			return true
		}
		returns++
		if len(ret.Results) == 0 {
			// 이름 붙은 결과의 맨 return — 오류 결과에 무엇이 담겼는지 AST 만으로는 보장 못 하므로 거절함.
			t.Errorf("%s: %s storage error exit uses a bare return", where, fn)
			return true
		}
		if id, ok := ret.Results[len(ret.Results)-1].(*ast.Ident); ok && id.Name == "nil" {
			t.Errorf("%s: %s storage error exit returns a nil error", where, fn)
		}
		return true
	})
	if returns == 0 {
		t.Errorf("%s: %s storage error exit does not return", where, fn)
	}
}

// storageMethods 는 database/sql 의 오류를 내는 메서드 이름임.
var storageMethods = map[string]bool{
	"BeginTx": true, "Commit": true, "ExecContext": true, "Exec": true, "QueryContext": true, "Query": true,
	"QueryRowContext": true, "QueryRow": true, "Scan": true, "Err": true, "RowsAffected": true, "LastInsertId": true,
	"Close": true,
}

// assignsErr 는 문장이 err 에 대입하는지 봄.
func assignsErr(stmt ast.Stmt) bool {
	assign, ok := stmt.(*ast.AssignStmt)
	if !ok {
		return false
	}
	for _, lhs := range assign.Lhs {
		if id, ok := lhs.(*ast.Ident); ok && id.Name == "err" {
			return true
		}
	}
	return false
}

// assignsErrFromStorageCall 은 문장이 err 에 호출 결과를 대입하고, 그 호출이 저장소 호출인지 봄:
// database/sql 메서드이거나(체인 포함), 인자에 트랜잭션·질의자(tx·q·db·j.db)나 ctx 를 넘기는 헬퍼.
func assignsErrFromStorageCall(stmt ast.Stmt) bool {
	assign, ok := stmt.(*ast.AssignStmt)
	if !ok || len(assign.Rhs) != 1 {
		return false
	}
	namesErr := false
	for _, lhs := range assign.Lhs {
		if id, ok := lhs.(*ast.Ident); ok && id.Name == "err" {
			namesErr = true
		}
	}
	call, ok := assign.Rhs[0].(*ast.CallExpr)
	if !namesErr || !ok {
		return false
	}
	storage := false
	ast.Inspect(call, func(n ast.Node) bool {
		c, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if sel, ok := c.Fun.(*ast.SelectorExpr); ok && storageMethods[sel.Sel.Name] {
			storage = true
		}
		for _, arg := range c.Args {
			switch a := arg.(type) {
			case *ast.Ident:
				// ctx 를 받는 헬퍼는 저장소(또는 수집) 입출력을 함 — 그 오류도 같은 두 성질을 지켜야 함.
				if a.Name == "tx" || a.Name == "q" || a.Name == "db" || a.Name == "ctx" {
					storage = true
				}
			case *ast.SelectorExpr:
				if a.Sel.Name == "db" {
					storage = true
				}
			}
		}
		return !storage
	})
	return storage
}

// isErrNotNil 은 조건이 `err != nil` 이거나 `!errors.Is(err, sql.ErrNoRows)`(행 없음 밖의 저장 오류)인지 봄
// (`if x, err := f(); err != nil` 포함 — Cond 만 봄).
func isErrNotNil(cond ast.Expr) bool {
	if un, ok := cond.(*ast.UnaryExpr); ok && un.Op == token.NOT {
		if call, ok := un.X.(*ast.CallExpr); ok && len(call.Args) == 2 {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "Is" {
				if id, ok := call.Args[0].(*ast.Ident); ok && id.Name == "err" {
					if target, ok := call.Args[1].(*ast.SelectorExpr); ok && target.Sel.Name == "ErrNoRows" {
						return true
					}
				}
			}
		}
	}
	bin, ok := cond.(*ast.BinaryExpr)
	if !ok || bin.Op != token.NEQ {
		return false
	}
	x, xok := bin.X.(*ast.Ident)
	y, yok := bin.Y.(*ast.Ident)
	return xok && yok && x.Name == "err" && y.Name == "nil"
}

// opensTx 는 함수 본문이 BeginTx 를 직접 부르는지 봄.
func opensTx(fn *ast.FuncDecl) bool {
	found := false
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "BeginTx" {
				found = true
			}
		}
		return !found
	})
	return found
}

// rollbackDeferredAfterBegin 은 BeginTx 를 담은 문장 뒤, 그 오류 검사 다음 문장이 `defer tx.Rollback()` 인지 봄.
func rollbackDeferredAfterBegin(fn *ast.FuncDecl) bool {
	ok := false
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		block, isBlock := n.(*ast.BlockStmt)
		if !isBlock {
			return true
		}
		for i, stmt := range block.List {
			if !containsBeginTx(stmt) || i+2 >= len(block.List) {
				continue
			}
			if def, isDefer := block.List[i+2].(*ast.DeferStmt); isDefer {
				if sel, isSel := def.Call.Fun.(*ast.SelectorExpr); isSel && sel.Sel.Name == "Rollback" {
					ok = true
				}
			}
		}
		return true
	})
	return ok
}

func containsBeginTx(stmt ast.Stmt) bool {
	found := false
	ast.Inspect(stmt, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "BeginTx" {
				found = true
			}
		}
		return !found
	})
	return found
}
