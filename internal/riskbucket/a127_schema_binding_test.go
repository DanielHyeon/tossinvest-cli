//go:build tossos_testseams

package riskbucket

// a127 — 위험 snapshot 적재기는 주입된 현재 원장 스키마와 정확히 같은 원장만 읽는다(design D1 · D2 · D3 · D7, 반증 S3 · S5 · S6 · S9 · S11 · S13 · S14).
// 실제 원장(`journal.Open`) 양성은 journal 을 import 할 수 있는 engine 패키지가 잰다(a112 트립와이어 반전).

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func a127RiskFixture(t *testing.T) productionRiskFixture {
	t.Helper()
	return newProductionRiskFixture(t, MarketKR, time.Date(2026, 8, 4, 2, 0, 0, 0, time.UTC))
}

func a127Exec(t *testing.T, fixture productionRiskFixture, statements ...string) {
	t.Helper()
	db := openProductionRiskDB(t, fixture.config.JournalPath)
	defer db.Close()
	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			t.Fatalf("%s: %v", statement, err)
		}
	}
}

// 대조 — 주입 값과 같은 원장은 읽힘(축소 원장, 주입 1).
func TestA127RiskLoaderReadsTheLedgerWhoseSchemaEqualsTheInjectedVersion(t *testing.T) {
	fixture := a127RiskFixture(t)
	if _, err := LoadProductionRiskSnapshotAuthority(context.Background(), fixture.config, fixture.input); err != nil {
		t.Fatalf("matching ledger refused: %v", err)
	}
}

// S3 · S9 · S11 — 이 빌드보다 새 원장(읽기 집합은 온전, 버전만 증가)은 거절, 문구가 방향을 말하고 범위 국소 신원이 아님.
func TestA127RiskLoaderRefusesANewerLedger(t *testing.T) {
	fixture := a127RiskFixture(t)
	a127Exec(t, fixture, fmt.Sprintf(`PRAGMA user_version=%d`, productionRiskTestJournalSchema+1))
	_, err := LoadProductionRiskSnapshotAuthority(context.Background(), fixture.config, fixture.input)
	if !errors.Is(err, ErrProductionRiskSnapshotUnavailable) || errors.Is(err, ErrProductionRiskScopeRefused) || !strings.Contains(err.Error(), "newer than this build") {
		t.Fatalf("err=%v, want an Unavailable (not scope-refused) refusal naming the newer ledger", err)
	}
}

// S5 · S9 — 마이그레이션되지 않은(더 옛) 원장은 표 · 열이 다 있어도 거절.
func TestA127RiskLoaderRefusesAnOlderLedger(t *testing.T) {
	fixture := a127RiskFixture(t)
	fixture.config.JournalSchemaVersion = productionRiskTestJournalSchema + 1
	_, err := LoadProductionRiskSnapshotAuthority(context.Background(), fixture.config, fixture.input)
	if !errors.Is(err, ErrProductionRiskSnapshotUnavailable) || errors.Is(err, ErrProductionRiskScopeRefused) || !strings.Contains(err.Error(), "older than this build") {
		t.Fatalf("err=%v, want an Unavailable refusal naming the older ledger", err)
	}
}

// S6 — 주입 누락(0 · 음수)은 원장을 열기 전에 거절: 존재하지 않는 원장 경로로 불러도 오류는 경로 실패가 아니라 주입 누락 문구.
func TestA127RiskLoaderRefusesAMissingInjectionBeforeOpeningTheLedger(t *testing.T) {
	for _, injected := range []int{0, -1} {
		fixture := a127RiskFixture(t)
		fixture.config.JournalSchemaVersion = injected
		fixture.config.JournalPath = filepath.Join(t.TempDir(), "absent.db")
		_, err := LoadProductionRiskSnapshotAuthority(context.Background(), fixture.config, fixture.input)
		if !errors.Is(err, ErrProductionRiskSnapshotUnavailable) || !strings.Contains(err.Error(), "journal schema version not injected") {
			t.Fatalf("injected=%d err=%v, want the injection refusal before any ledger access", injected, err)
		}
	}
}

// S14 — 조건부 질의: latch 가 선 범위는 사용량 판독 전에 범위 국소 거절로 돌아가므로, 사용량 전용 열이 없는 원장의 결함은 판독 전 prepare 가
// 먼저 드러내야 함(결함 신원 — ScopeRefused 아님). 대조: 같은 latch 에 온전한 원장이면 ScopeRefused.
func TestA127RiskLoaderReportsAMissingUsageColumnAsADefectEvenOnALatchedScope(t *testing.T) {
	latch := func(fixture productionRiskFixture) string {
		return fmt.Sprintf(`INSERT INTO risk_bucket_scope_latches(account_ref,market,symbol,prospective_generation) VALUES('%s','%s','005930','g-latched')`,
			fixture.config.AccountID, fixture.config.Market)
	}
	control := a127RiskFixture(t)
	a127Exec(t, control, latch(control))
	if _, err := LoadProductionRiskSnapshotAuthority(context.Background(), control.config, control.input); !errors.Is(err, ErrProductionRiskScopeRefused) {
		t.Fatalf("control: latched scope on an intact ledger err=%v, want the scope refusal", err)
	}
	fixture := a127RiskFixture(t)
	a127Exec(t, fixture, latch(fixture),
		`CREATE TABLE a127_d AS SELECT decision_id,account_ref,market,symbol FROM risk_bucket_final_decisions`,
		`DROP TABLE risk_bucket_final_decisions`,
		`ALTER TABLE a127_d RENAME TO risk_bucket_final_decisions`)
	_, err := LoadProductionRiskSnapshotAuthority(context.Background(), fixture.config, fixture.input)
	if err == nil || errors.Is(err, ErrProductionRiskScopeRefused) || !errors.Is(err, ErrProductionRiskSnapshotUnavailable) {
		t.Fatalf("err=%v, want a defect (Unavailable, not scope-refused) from the missing usage column", err)
	}
}

// S13 — 버전 확인 · scope latch · 다섯 사용량 판독이 전부 같은 읽기 트랜잭션에서(구조 단언 — 판독 하나라도 `db` 로 돌아가면 실패).
func TestA127RiskLoaderReadsEverythingInOneReadOnlyTransaction(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "production_snapshot_authority.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var fn *ast.FuncDecl
	for _, decl := range file.Decls {
		if d, ok := decl.(*ast.FuncDecl); ok && d.Name.Name == "loadProductionRiskEntries" {
			fn = d
		}
	}
	if fn == nil {
		t.Fatal("loadProductionRiskEntries not found")
	}
	txName := ""
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if as, ok := n.(*ast.AssignStmt); ok && len(as.Rhs) == 1 {
			if call, ok := as.Rhs[0].(*ast.CallExpr); ok {
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "BeginTx" {
					if id, ok := as.Lhs[0].(*ast.Ident); ok {
						txName = id.Name
					}
				}
			}
		}
		return true
	})
	if txName == "" {
		t.Fatal("no BeginTx in loadProductionRiskEntries — the reads are not in one transaction")
	}
	reads, usage := 0, 0
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		switch fun := call.Fun.(type) {
		case *ast.SelectorExpr:
			switch fun.Sel.Name {
			case "QueryRowContext", "QueryContext", "ExecContext", "PrepareContext":
				if id, ok := fun.X.(*ast.Ident); !ok || id.Name != txName {
					t.Errorf("%s: %s on %s, want the read-only transaction %q", fset.Position(call.Pos()), fun.Sel.Name, exprName(fun.X), txName)
				}
				reads++
			}
		case *ast.Ident:
			if fun.Name == "ReadJournalBucketUsage" {
				usage++
				if id, ok := call.Args[1].(*ast.Ident); !ok || id.Name != txName {
					t.Errorf("%s: ReadJournalBucketUsage reads through %s, want %q", fset.Position(call.Pos()), exprName(call.Args[1]), txName)
				}
			}
		}
		return true
	})
	if reads < 2 || usage != 1 {
		t.Fatalf("reads=%d usage calls=%d — the census found less than the version, latch and usage reads", reads, usage)
	}
	a127AssertOneReadOnlyTransactionLifetime(t, fset, fn, txName)
	a127AssertPrepareBeforeFirstDataQuery(t, fset, fn)
}

// a127AssertOneReadOnlyTransactionLifetime — codex 구현 리뷰 P2 #1: 수신자 철자만으로는 같은 트랜잭션임을 못 보임(같은 이름으로 tx 를 닫고 다시 열면 통과).
// 그래서 BeginTx 는 정확히 하나 · 옵션은 ReadOnly: true · tx 대입은 그 하나뿐 · tx 의 Rollback/Commit 은 defer 된 Rollback 하나뿐임을 단언함.
func a127AssertOneReadOnlyTransactionLifetime(t *testing.T, fset *token.FileSet, fn *ast.FuncDecl, txName string) {
	t.Helper()
	begins, assigns, closes := 0, 0, 0
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.DeferStmt:
			if sel, ok := node.Call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "Rollback" {
				if id, ok := sel.X.(*ast.Ident); ok && id.Name == txName {
					return false // 허용된 유일한 닫기 — 함수 끝에서만 실행됨
				}
			}
		case *ast.AssignStmt:
			for _, lhs := range node.Lhs {
				if id, ok := lhs.(*ast.Ident); ok && id.Name == txName {
					assigns++
				}
			}
		case *ast.CallExpr:
			sel, ok := node.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			switch sel.Sel.Name {
			case "BeginTx":
				begins++
				if len(node.Args) != 2 || !a127ReadOnlyOption(node.Args[1]) {
					t.Errorf("%s: BeginTx without &sql.TxOptions{ReadOnly: true}", fset.Position(node.Pos()))
				}
			case "Rollback", "Commit":
				if id, ok := sel.X.(*ast.Ident); ok && id.Name == txName {
					closes++
					t.Errorf("%s: %s.%s ends the read transaction before the function does", fset.Position(node.Pos()), txName, sel.Sel.Name)
				}
			}
		}
		return true
	})
	if begins != 1 || assigns != 1 || closes != 0 {
		t.Fatalf("BeginTx=%d tx assignments=%d early closes=%d — want one read-only transaction living until the function returns", begins, assigns, closes)
	}
}

func a127ReadOnlyOption(arg ast.Expr) bool {
	unary, ok := arg.(*ast.UnaryExpr)
	if !ok || unary.Op != token.AND {
		return false
	}
	lit, ok := unary.X.(*ast.CompositeLit)
	if !ok {
		return false
	}
	for _, element := range lit.Elts {
		if kv, ok := element.(*ast.KeyValueExpr); ok {
			if key, ok := kv.Key.(*ast.Ident); ok && key.Name == "ReadOnly" {
				if value, ok := kv.Value.(*ast.Ident); ok && value.Name == "true" {
					return true
				}
			}
		}
	}
	return false
}

// a127AssertPrepareBeforeFirstDataQuery — codex 구현 리뷰 P2 #2: 열 삭제 시험은 prepare 를 latch 질의 뒤로 옮긴 변이를 못 가름. 소스 순서로 단언함:
// 원장 데이터 질의(PRAGMA user_version 을 뺀 Query*Context · ReadJournalBucketUsage)는 전부 마지막 PrepareContext 보다 뒤.
func a127AssertPrepareBeforeFirstDataQuery(t *testing.T, fset *token.FileSet, fn *ast.FuncDecl) {
	t.Helper()
	var lastPrepare, firstData token.Pos
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		data := false
		switch fun := call.Fun.(type) {
		case *ast.SelectorExpr:
			switch fun.Sel.Name {
			case "PrepareContext":
				if call.Pos() > lastPrepare {
					lastPrepare = call.Pos()
				}
			case "QueryRowContext", "QueryContext":
				data = !(len(call.Args) >= 2 && a127IsPragmaUserVersion(call.Args[1]))
			}
		case *ast.Ident:
			data = fun.Name == "ReadJournalBucketUsage"
		}
		if data && (firstData == token.NoPos || call.Pos() < firstData) {
			firstData = call.Pos()
		}
		return true
	})
	if lastPrepare == token.NoPos || firstData == token.NoPos || !(lastPrepare < firstData) {
		t.Fatalf("last prepare at %s, first ledger-data query at %s — every data query must come after the prepares",
			fset.Position(lastPrepare), fset.Position(firstData))
	}
}

func a127IsPragmaUserVersion(expr ast.Expr) bool {
	lit, ok := expr.(*ast.BasicLit)
	return ok && strings.Contains(lit.Value, "PRAGMA user_version")
}

func exprName(expr ast.Expr) string {
	if id, ok := expr.(*ast.Ident); ok {
		return id.Name
	}
	return fmt.Sprintf("%T", expr)
}
