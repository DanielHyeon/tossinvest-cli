package strategyrouter

// a127 — route 적재기는 주입된 현재 원장 스키마와 정확히 같은 원장만 읽는다(design D1 · D2 · D3 · D7, 반증 S4 · S5 · S6 · S8 · S9 · S12).
// 방향 문구는 공개 경계(LoadProductionRouteAuthorityBatch)에서 단언(S12 — `:352` 감싸기가 원인을 지우면 실패). 실제 원장(`journal.Open`) 양성은
// 외부 시험 패키지(a127_real_journal_test.go, tossos_testseams)가 잰다.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
)

func a127RouteLoad(t *testing.T, fixture *productionRouteFixture, config ProductionRouteConfig) error {
	t.Helper()
	_, err := LoadProductionRouteAuthorityBatch(context.Background(), config, []ProductionRouteTarget{{Symbol: config.Symbol}})
	return err
}

func a127RouteExec(t *testing.T, fixture *productionRouteFixture, statements ...string) {
	t.Helper()
	db, err := sql.Open("sqlite", fixture.journal)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			t.Fatalf("%s: %v", statement, err)
		}
	}
}

// 대조 — 주입 값과 같은 원장은 읽힘.
func TestA127RouteLoaderReadsTheLedgerWhoseSchemaEqualsTheInjectedVersion(t *testing.T) {
	fixture := newProductionRouteFixture(t)
	if err := a127RouteLoad(t, fixture, fixture.config[MarketKR]); err != nil {
		t.Fatalf("matching ledger refused: %v", err)
	}
}

// S4 · S9 · S12 — 이 빌드보다 새 원장(읽기 집합 온전, 버전만 증가)은 거절, 방향 문구가 Batch 경계까지 옴.
func TestA127RouteLoaderRefusesANewerLedgerAndSaysSoAtTheBatchBoundary(t *testing.T) {
	fixture := newProductionRouteFixture(t)
	a127RouteExec(t, fixture, fmt.Sprintf(`PRAGMA user_version=%d`, productionRouteTestJournalSchema+1))
	err := a127RouteLoad(t, fixture, fixture.config[MarketKR])
	if !errors.Is(err, ErrProductionRouteUnavailable) || !strings.Contains(err.Error(), "newer than this build") {
		t.Fatalf("err=%v, want an Unavailable refusal naming the newer ledger", err)
	}
}

// S5 · S9 · S12 — 마이그레이션되지 않은(더 옛) 원장은 거절.
func TestA127RouteLoaderRefusesAnOlderLedgerAndSaysSoAtTheBatchBoundary(t *testing.T) {
	fixture := newProductionRouteFixture(t)
	config := fixture.config[MarketKR]
	config.JournalSchemaVersion = productionRouteTestJournalSchema + 1
	err := a127RouteLoad(t, fixture, config)
	if !errors.Is(err, ErrProductionRouteUnavailable) || !strings.Contains(err.Error(), "older than this build") {
		t.Fatalf("err=%v, want an Unavailable refusal naming the older ledger", err)
	}
}

// S6 — 주입 누락(0 · 음수)은 원장을 열기 전에 거절: 존재하지 않는 원장 경로로 불러도 주입 누락 문구.
func TestA127RouteLoaderRefusesAMissingInjectionBeforeOpeningTheLedger(t *testing.T) {
	for _, injected := range []int{0, -1} {
		fixture := newProductionRouteFixture(t)
		config := fixture.config[MarketKR]
		config.JournalSchemaVersion = injected
		config.JournalPath = filepath.Join(t.TempDir(), "absent.db")
		err := a127RouteLoad(t, fixture, config)
		if !errors.Is(err, ErrProductionRouteUnavailable) || !strings.Contains(err.Error(), "journal schema version not injected") {
			t.Fatalf("injected=%d err=%v, want the injection refusal before any ledger access", injected, err)
		}
	}
}

// S8 — 조건부 질의: active owner 가 없는 범위는 campaign 질의를 실행하지 않으므로, campaign 전용 열이 없는 원장의 결함은 판독 전 prepare 가
// 드러내야 함. 픽스처 원장에는 owner 행이 없음(= active owner 없음).
func TestA127RouteLoaderRefusesALedgerMissingACampaignColumnEvenWithoutAnActiveOwner(t *testing.T) {
	fixture := newProductionRouteFixture(t)
	a127RouteExec(t, fixture,
		`CREATE TABLE a127_pc AS SELECT id,account_ref,market,symbol,lane_id,lane_version,prospective_token,actual_position_generation,state FROM position_campaigns`,
		`DROP TABLE position_campaigns`,
		`ALTER TABLE a127_pc RENAME TO position_campaigns`)
	if err := a127RouteLoad(t, fixture, fixture.config[MarketKR]); !errors.Is(err, ErrProductionRouteUnavailable) || !strings.Contains(err.Error(), "journal read set unavailable") {
		t.Fatalf("err=%v, want the missing campaign column refused by the pre-read prepare", err)
	}
}

// codex 구현 리뷰 P2 #1 · #2 — route 의 판독 순서와 트랜잭션 수명을 구조로 단언: opener 는 읽기 전용 BeginTx 하나 · 원장 데이터 질의 없음 · 모든 prepare
// 뒤에만 성공 반환, Batch 는 opener 를 owner 재구성(데이터 질의)보다 먼저 부름. 열 삭제 시험은 prepare 를 뒤로 옮긴 변이를 못 가르므로 순서는 여기서 잼.
func TestA127RouteLoaderPreparesBeforeAnyLedgerDataRead(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "production.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	funcs := map[string]*ast.FuncDecl{}
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			funcs[fn.Name.Name] = fn
		}
	}
	opener, batch := funcs["openProductionRouteSnapshot"], funcs["LoadProductionRouteAuthorityBatch"]
	if opener == nil || batch == nil {
		t.Fatal("opener or batch not found")
	}
	begins, readOnly := 0, false
	var lastPrepare, success token.Pos
	ast.Inspect(opener.Body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.CallExpr:
			sel, ok := node.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			switch sel.Sel.Name {
			case "BeginTx":
				begins++
				if unary, ok := node.Args[1].(*ast.UnaryExpr); ok {
					if lit, ok := unary.X.(*ast.CompositeLit); ok {
						for _, element := range lit.Elts {
							if kv, ok := element.(*ast.KeyValueExpr); ok {
								if key, ok := kv.Key.(*ast.Ident); ok && key.Name == "ReadOnly" {
									if value, ok := kv.Value.(*ast.Ident); ok && value.Name == "true" {
										readOnly = true
									}
								}
							}
						}
					}
				}
			case "PrepareContext":
				if node.Pos() > lastPrepare {
					lastPrepare = node.Pos()
				}
			case "QueryRowContext", "QueryContext":
				if lit, ok := node.Args[1].(*ast.BasicLit); !ok || !strings.Contains(lit.Value, "PRAGMA user_version") {
					t.Errorf("%s: the opener reads ledger data before handing the transaction over", fset.Position(node.Pos()))
				}
			}
		case *ast.ReturnStmt:
			if len(node.Results) == 3 {
				if id, ok := node.Results[2].(*ast.Ident); ok && id.Name == "nil" {
					success = node.Pos()
				}
			}
		}
		return true
	})
	if begins != 1 || !readOnly || lastPrepare == token.NoPos || success == token.NoPos || !(lastPrepare < success) {
		t.Fatalf("opener BeginTx=%d readOnly=%v last prepare %s success return %s — want one read-only tx returned only after every prepare",
			begins, readOnly, fset.Position(lastPrepare), fset.Position(success))
	}
	var open, owners token.Pos
	ast.Inspect(batch.Body, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if id, ok := call.Fun.(*ast.Ident); ok {
				switch id.Name {
				case "openProductionRouteSnapshot":
					open = call.Pos()
				case "loadProductionRouteOwnersFrom":
					if owners == token.NoPos {
						owners = call.Pos()
					}
				}
			}
		}
		return true
	})
	if open == token.NoPos || owners == token.NoPos || !(open < owners) {
		t.Fatalf("batch opens the snapshot at %s and first reads owners at %s — the prepares must come first", fset.Position(open), fset.Position(owners))
	}
}
