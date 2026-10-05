package verifylive

// reconcile_structure_test.go — a121 tasks 2.2.2(R1)·2.4.1(봉인 census·생산 nil 핀).
//
// 구조 시험은 행동 시험이 못 보는 것을 못 박는다: 대사 파일에서 쓰기 메서드 호출·type assertion 이 없고, 의존
// 인터페이스가 공식 GET 읽기만 노출하고, Q1 보존 한도가 생산 빌드에서 nil 이며, 대사 줄 StepID 가 카탈로그 밖이다.
// 「주문·조건주문 변이 도달 0」 이 주장의 전부다 — 토큰 갱신 POST(token.go:138, auth 기반)는 이 경계 밖이고
// 행동 쪽 짝(TestReconcileSucceedsWithOnlyGetRequestsAndTheTokenPost)이 그 하나만 허용한다.

import (
	"go/ast"
	"go/build/constraint"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

// reconcileWriteMethods 는 design G2 P1-1·tasks 2.4.1 의 쓰기 7이름이다.
var reconcileWriteMethods = []string{
	"PlaceOrder", "CancelOrder", "ModifyOrder",
	"CreateConditionalOrder", "ModifyConditionalOrder", "ModifyConditionalOrderRef", "CancelConditionalOrder",
}

// reconcileReadMethods 는 대사 의존이 노출해도 되는 이름이다(공식 GET 읽기).
var reconcileReadMethods = map[string]bool{
	"ReconcileConditionalOrdersPage": true,
	"ReconcileOpenOrdersPage":        true,
	"ReconcileInstrument":            true,
}

// reconcileSourceFiles 는 census 대상 — 대사 경로의 비시험 소스 전부(cmd·verifylive·official).
func reconcileSourceFiles(t *testing.T) []string {
	t.Helper()
	var files []string
	for _, pattern := range []string{
		"reconcile*.go",
		filepath.Join("..", "official", "reconcile*.go"),
		filepath.Join("..", "..", "cmd", "tossctl", "verify_reconcile*.go"),
	} {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatal(err)
		}
		n := 0
		for _, m := range matches {
			if strings.HasSuffix(m, "_test.go") {
				continue
			}
			files = append(files, m)
			n++
		}
		if n == 0 {
			t.Fatalf("census found no reconcile source for %s — the census would pass on an empty sample", pattern)
		}
	}
	sort.Strings(files)
	return files
}

// TestReconcileFilesCallNoWriteMethodAndAssertNoType 는 봉인 census 다: 대사 파일에서 쓰기 7이름 호출 0,
// type assertion·type switch 0, verify 의 브로커 생성자 재사용 0, Broker·*official.Client 를 돌려주는 함수 0.
func TestReconcileFilesCallNoWriteMethodAndAssertNoType(t *testing.T) {
	forbiddenCalls := map[string]bool{}
	for _, m := range reconcileWriteMethods {
		forbiddenCalls[m] = true
	}
	forbiddenIdents := map[string]bool{"verifyBrokerFactory": true, "buildVerifyBroker": true}
	for _, path := range reconcileSourceFiles(t) {
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(file, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.CallExpr:
				name := ""
				switch fn := n.Fun.(type) {
				case *ast.SelectorExpr:
					name = fn.Sel.Name
				case *ast.Ident:
					name = fn.Name
				}
				if forbiddenCalls[name] {
					t.Errorf("%s: calls write method %s", fset.Position(n.Pos()), name)
				}
			case *ast.SelectorExpr:
				// 메서드 값으로 꺼내 쓰는 것도 호출 경로다.
				if forbiddenCalls[n.Sel.Name] {
					t.Errorf("%s: references write method %s", fset.Position(n.Pos()), n.Sel.Name)
				}
			case *ast.TypeAssertExpr:
				t.Errorf("%s: type assertion in a reconcile file (a concrete client in scope breaks the seal)", fset.Position(n.Pos()))
			case *ast.TypeSwitchStmt:
				t.Errorf("%s: type switch in a reconcile file", fset.Position(n.Pos()))
			case *ast.Ident:
				if forbiddenIdents[n.Name] {
					t.Errorf("%s: reuses %s — the reconcile path must not obtain a Broker", fset.Position(n.Pos()), n.Name)
				}
			case *ast.FuncDecl:
				if n.Type.Results == nil {
					return true
				}
				for _, field := range n.Type.Results.List {
					if returnsBrokerOrClient(field.Type) {
						t.Errorf("%s: %s returns a Broker or *official.Client", fset.Position(n.Pos()), n.Name.Name)
					}
				}
			}
			return true
		})
	}
}

func returnsBrokerOrClient(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.StarExpr:
		return returnsBrokerOrClient(e.X)
	case *ast.Ident:
		return e.Name == "Broker" || e.Name == "Client"
	case *ast.SelectorExpr:
		return e.Sel.Name == "Broker" || e.Sel.Name == "Client"
	}
	return false
}

// TestReconcileReaderExposesOnlyOfficialGetReads — 대사 의존 인터페이스의 메서드 집합이 읽기 목록 안에 있다.
func TestReconcileReaderExposesOnlyOfficialGetReads(t *testing.T) {
	typ := reflect.TypeOf((*ReconcileReader)(nil)).Elem()
	if typ.NumMethod() == 0 {
		t.Fatal("ReconcileReader has no methods")
	}
	for i := 0; i < typ.NumMethod(); i++ {
		name := typ.Method(i).Name
		if !reconcileReadMethods[name] {
			t.Fatalf("ReconcileReader exposes %s, which is not an official GET read", name)
		}
	}
}

// TestReconcileRetentionIsNilInProductionBuilds 는 design G1-4 P0-1 의 AST 핀이다 — 생산 빌드(시험 파일·
// tossos_testseams 태그 파일 제외)에서 reconcileRetention 은 초기값 없이 선언되고 어디서도 대입·주소 획득되지 않는다.
func TestReconcileRetentionIsNilInProductionBuilds(t *testing.T) {
	requireProductionVar(t, "reconcileRetention", "")
}

// TestReconcileFreshnessBoundIsTheApprovedConstantInProduction 은 Q3(리뷰 승인 15초)의 핀이다 — 생산 빌드에서
// reconcileFreshnessBound 는 상수 reconcileFreshnessBoundValue 로만 초기화되고 다른 대입(init 포함)·주소 획득이 없다.
func TestReconcileFreshnessBoundIsTheApprovedConstantInProduction(t *testing.T) {
	if reconcileFreshnessBoundValue != 15*time.Second {
		t.Fatalf("Q3 constant is %v, the reviewed value is 15s", reconcileFreshnessBoundValue)
	}
	requireProductionVar(t, "reconcileFreshnessBound", "reconcileFreshnessBoundValue")
}

// requireProductionVar 는 생산 파일에서 name 이 정확히 한 번 선언되고(초기값은 initIdent 식별자 하나, 빈 값이면
// 초기값 없음) 대입·증감·주소 획득이 없음을 단언한다.
func requireProductionVar(t *testing.T, name, initIdent string) {
	t.Helper()
	matches, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	declared := 0
	for _, path := range matches {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if seamOnly(t, src) {
			continue
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, src, parser.ParseComments)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(file, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.ValueSpec:
				for i, id := range n.Names {
					if id.Name != name {
						continue
					}
					declared++
					switch {
					case initIdent == "" && len(n.Values) > 0:
						t.Errorf("%s: %s has an initial value in a production build", fset.Position(id.Pos()), name)
					case initIdent != "":
						if len(n.Values) <= i {
							t.Errorf("%s: %s must be initialised from %s", fset.Position(id.Pos()), name, initIdent)
						} else if v, ok := n.Values[i].(*ast.Ident); !ok || v.Name != initIdent {
							t.Errorf("%s: %s is initialised from something other than %s", fset.Position(id.Pos()), name, initIdent)
						}
					}
				}
			case *ast.AssignStmt:
				for _, lhs := range n.Lhs {
					if id, ok := lhs.(*ast.Ident); ok && id.Name == name {
						t.Errorf("%s: %s is assigned in a production build", fset.Position(n.Pos()), name)
					}
				}
			case *ast.IncDecStmt:
				if id, ok := n.X.(*ast.Ident); ok && id.Name == name {
					t.Errorf("%s: %s is modified in a production build", fset.Position(n.Pos()), name)
				}
			case *ast.UnaryExpr:
				if id, ok := n.X.(*ast.Ident); ok && id.Name == name && n.Op == token.AND {
					t.Errorf("%s: the address of %s escapes in a production build", fset.Position(n.Pos()), name)
				}
			}
			return true
		})
	}
	if declared != 1 {
		t.Fatalf("%s declared %d time(s) in production files, want exactly 1", name, declared)
	}
}

// seamOnly 는 파일이 tossos_testseams 태그에서만 빌드되는지 본다.
func seamOnly(t *testing.T, src []byte) bool {
	t.Helper()
	for _, line := range strings.Split(string(src), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "package ") {
			return false
		}
		if !constraint.IsGoBuild(line) {
			continue
		}
		expr, err := constraint.Parse(line)
		if err != nil {
			t.Fatal(err)
		}
		return !expr.Eval(func(tag string) bool { return tag != "tossos_testseams" })
	}
	return false
}

// TestReconcileStepIDIsOutsideTheCatalogueCleanupAndAbort 는 로트 1 R1 이다 — StepID 만 비교하는 소비자 넷
// (LastEntry·heldAfter·m0ManualReconcileIDs·baselineSellable)이 대사 줄을 단계 판정으로 읽지 않게 한다.
func TestReconcileStepIDIsOutsideTheCatalogueCleanupAndAbort(t *testing.T) {
	if StepReconcile == "" {
		t.Fatal("the reconcile line needs its own StepID; an empty one is shared with every legacy line")
	}
	taken := map[StepID]bool{StepCleanup: true, StepAbort: true}
	for _, s := range Steps() {
		taken[s.ID] = true
	}
	if taken[StepReconcile] {
		t.Fatalf("StepReconcile %q collides with the catalogue, cleanup or abort", StepReconcile)
	}
	if StepReconcile == StepConditionalCancel {
		t.Fatal("StepReconcile reuses conditional-cancel — it would release or re-hold conditionals")
	}
}
