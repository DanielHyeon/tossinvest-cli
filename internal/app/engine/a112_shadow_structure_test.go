//go:build tossos_testseams

package engine

// a112 7.3.1 SHADOW 구조 핀(브리프 v3.3 §2 ② ③ · §4 ⓐⓑⓒ · §5 ①~⑤).
//
// 행동 시험이 「무엇이 일어나는가」 를 재면, 여기서는 「어디에서만 일어날 수 있는가」 를 생산 빌드(무태그) 타입 검사로 못 박는다:
//   ① runProductionStrategyMarketCycle 본문에서 shadow 타입 식은 evaluate 인자 index 5 의 부분 트리 하나, 그 안의 유일한 호출은
//      strategyShadowPair.forMarket. forMarket 본문은 형제 strategyProposalAuthorityPair.forMarket 과 같은 모양(시장 비교 · 필드 반환 · 호출 0).
//   ②(census) shadow 타입(묶음 · 짝 · ShadowInput · FamilyShadow · shadow 설정)을 **사용**하는 함수는 허용 목록뿐 — 선언이 아니라 사용
//      (types.Info.Uses 와 본문 식 타입)으로 센다. 양성 대조: 세탁 접근자 모양의 합성 패키지를 같은 걸음이 잡는다.
//   ③ shadow 단계 호출 폐포에 refresh · 조립 · 파도 합류 · 활성화 적재 · 조정 · dispatch · 원장 쓰기 0, 공유 캐시 필드 사용 0.
//   ④ evaluate 안 묶음은 record 로 넘기는 인자 하나뿐, ⑤ record 의 대입은 파도 증가와 같은 임계 구역.
//   ⓐⓑⓒ 조정자의 수집 helper 문장 하나 · 충돌 return 은 부재 값 리터럴 · 정상 return 만 묶음. helper · 생성자 shape.
//   cycle 클로저: 두 생산 자리가 productionStrategyCycle 하나를 부르고, 그 몸통은 recover 없는 defer(성공 플래그) · 주기 run 뒤의 시작.

import (
	"bytes"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/printer"
	"go/token"
	"go/types"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/testenv"
)

const a112EnginePath = "github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"

// a112ShadowTypeNames 는 census ② 가 세는 shadow 타입(패키지 경로 · 이름)이다.
var a112ShadowTypeNames = map[string]bool{
	a112EnginePath + ".strategyShadowBatch":                                       true,
	a112EnginePath + ".strategyShadowPair":                                        true,
	a112EnginePath + ".strategyShadowCell":                                        true,
	a112EnginePath + ".strategyShadowObservation":                                 true,
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyworker.ShadowInput":  true,
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyshadow.FamilyShadow": true,
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyshadow.Config":       true,
}

// a112ShadowAllowed 는 shadow 타입을 쓸 수 있는 함수(운반 · 보관 · 단계 · 투영)다. 이름은 「수신자.이름」.
var a112ShadowAllowed = map[string]bool{
	// 운반
	"strategyShadowBatch.collect": true, "strategyShadowBatch.boundTo": true, "strategyShadowPair.forMarket": true,
	"strategyProposalAuthorityLoader.shadowConfig": true, "coordinateMarketProposals": true,
	"strategyProposalAuthorityLoader.collectMarket": true, "strategyProposalAuthorityLoader.collect": true,
	"Context.NewPairedStrategyEntryProductionAssembly": true,
	// 주기 함수 — evaluate 인자 한 자리(핀 ①이 모양을 못 박는다)
	"Context.runProductionStrategyMarketCycle": true,
	// 보관 · 단계 · 투영
	"strategyLaneRuntime.evaluate": true, "strategyLaneRuntime.record": true, "newStrategyLaneRuntime": true,
	"strategyLaneRuntime.startShadowStep": true, "strategyLaneRuntime.superviseShadowStep": true,
	"strategyLaneRuntime.runShadowStep": true, "strategyLaneRuntime.publishShadow": true, "strategyLaneRuntime.loadShadow": true,
	"strategyLaneRuntime.invalidateShadow": true, "strategyLaneRuntime.projection": true, "strategyLaneProjection": true,
	"shadowObservationUsable": true,
}

// a112ShadowForbidden 는 shadow 타입이 절대 닿으면 안 되는 주문 경로 함수다(허용 목록 밖이면 어차피 실패하지만 이름으로 한 번 더 적는다).
var a112ShadowForbidden = []string{"strategyFamilyGate.admit", "strategyProposalMarketAuthority.dispatchHandoffs",
	"strategyProposalMarketAuthority.dispatchHandoff", "strategyProposalMarketAuthority.authorityForOwnerScope",
	"strategyMarketArbitration.entries", "dispatchStrategyMarketHandoffs", "strategyDispatchCycle.dispatch"}

func a112CheckedEngine(t *testing.T) testenv.CheckedPackage {
	t.Helper()
	return testenv.TypeCheckProduction(t, ".", a112EnginePath)
}

func a112FuncName(decl *ast.FuncDecl) string {
	if decl.Recv == nil || len(decl.Recv.List) == 0 {
		return decl.Name.Name
	}
	typ := decl.Recv.List[0].Type
	if star, ok := typ.(*ast.StarExpr); ok {
		typ = star.X
	}
	if ident, ok := typ.(*ast.Ident); ok {
		return ident.Name + "." + decl.Name.Name
	}
	return "?." + decl.Name.Name
}

// a112ShadowHolders 는 shadow 상태를 필드로 **보관하도록 허용된** 이름 있는 타입이다(레인 런타임 · 조립). 이 타입을 다루기만 하는 함수는
// 세지 않고, 그 shadow 필드를 실제로 꺼내는 식(`fresh.shadow` · `runtime.shadowCells[market]`)만 그 식의 타입으로 센다. 다른 이름 있는
// 타입은 밑바탕 구조체까지 걷는다 — 그래서 authority 같은 주문 경로 타입에 shadow 필드를 숨기면 그 타입을 쓰는 모든 함수가 걸린다.
var a112ShadowHolders = map[string]bool{
	a112EnginePath + ".strategyLaneRuntime":             true,
	a112EnginePath + ".StrategyEntryProductionAssembly": true,
}

// a112ContainsShadowType 는 타입이 shadow 타입을 품는지 걷는다(포인터 · 슬라이스 · 배열 · 맵 · 채널 · 구조체 필드 · 함수 서명 · 별칭 · 인스턴스).
func a112ContainsShadowType(typ types.Type, names map[string]bool, seen map[types.Type]bool) bool {
	if typ == nil || seen[typ] {
		return false
	}
	seen[typ] = true
	switch value := types.Unalias(typ).(type) {
	case *types.Named:
		object := value.Obj()
		if object != nil && object.Pkg() != nil && names[object.Pkg().Path()+"."+object.Name()] {
			return true
		}
		if object != nil && object.Pkg() != nil && a112ShadowHolders[object.Pkg().Path()+"."+object.Name()] {
			return false
		}
		for index := range value.TypeArgs().Len() {
			if a112ContainsShadowType(value.TypeArgs().At(index), names, seen) {
				return true
			}
		}
		return a112ContainsShadowType(value.Underlying(), names, seen)
	case *types.Pointer:
		return a112ContainsShadowType(value.Elem(), names, seen)
	case *types.Slice:
		return a112ContainsShadowType(value.Elem(), names, seen)
	case *types.Array:
		return a112ContainsShadowType(value.Elem(), names, seen)
	case *types.Map:
		return a112ContainsShadowType(value.Key(), names, seen) || a112ContainsShadowType(value.Elem(), names, seen)
	case *types.Chan:
		return a112ContainsShadowType(value.Elem(), names, seen)
	case *types.Struct:
		for index := range value.NumFields() {
			if a112ContainsShadowType(value.Field(index).Type(), names, seen) {
				return true
			}
		}
	case *types.Signature:
		for _, tuple := range []*types.Tuple{value.Params(), value.Results()} {
			for index := range tuple.Len() {
				if a112ContainsShadowType(tuple.At(index).Type(), names, seen) {
					return true
				}
			}
		}
	}
	return false
}

// a112CalleeExprs 는 호출식의 피호출 자리(`f` · `x.m`)다. 그 식의 타입은 함수 서명이라 매개변수에 shadow 타입이 있으면 걸리는데,
// 그것은 「값을 쓴다」 가 아니라 「그 함수를 부른다」 다 — 값을 넘기면 인자 식이 따로 걸린다. 그래서 피호출 자리만 뺀다(함수 값을 인자로
// 넘기는 세탁은 여전히 걸린다 — 그 식은 피호출 자리가 아니다).
func a112CalleeExprs(node ast.Node) map[ast.Expr]bool {
	callees := map[ast.Expr]bool{}
	ast.Inspect(node, func(child ast.Node) bool {
		if call, ok := child.(*ast.CallExpr); ok {
			callees[call.Fun] = true
			if selector, ok := call.Fun.(*ast.SelectorExpr); ok {
				callees[selector.Sel] = true
			}
		}
		return true
	})
	return callees
}

// a112ShadowUsers 는 shadow 타입을 쓰는 함수 이름을 센다: 본문의 모든 식 타입과 Uses 가 가리키는 객체 타입.
func a112ShadowUsers(files []*ast.File, info *types.Info, names map[string]bool) map[string][]string {
	users := map[string][]string{}
	for _, file := range files {
		for _, decl := range file.Decls {
			function, ok := decl.(*ast.FuncDecl)
			if !ok || function.Body == nil {
				continue
			}
			name := a112FuncName(function)
			callees := a112CalleeExprs(function)
			ast.Inspect(function, func(node ast.Node) bool {
				expression, ok := node.(ast.Expr)
				if !ok || callees[expression] {
					return true
				}
				if tv, ok := info.Types[expression]; ok && a112ContainsShadowType(tv.Type, names, map[types.Type]bool{}) {
					users[name] = append(users[name], types.ExprString(expression))
					return true
				}
				if ident, ok := expression.(*ast.Ident); ok {
					if object := info.Uses[ident]; object != nil && a112ContainsShadowType(object.Type(), names, map[types.Type]bool{}) {
						users[name] = append(users[name], ident.Name)
					}
				}
				return true
			})
		}
	}
	return users
}

func TestOnlyTheAllowedFunctionsEverTouchAShadowType(t *testing.T) {
	checked := a112CheckedEngine(t)
	if len(checked.Info.Types) == 0 {
		t.Fatal("the type-check recorded no expression types — the census would read nothing")
	}
	users := a112ShadowUsers(checked.Files, checked.Info, a112ShadowTypeNames)
	var names []string
	for name, uses := range users {
		names = append(names, name)
		if !a112ShadowAllowed[name] {
			t.Errorf("%s touches a shadow type (%s) — only the carry · keep · step · projection functions may", name, strings.Join(uses, ", "))
		}
	}
	for _, forbidden := range a112ShadowForbidden {
		if len(users[forbidden]) != 0 {
			t.Errorf("order-path function %s touches a shadow type: %v", forbidden, users[forbidden])
		}
	}
	// 하한: 운반 · 단계 · 투영이 실제로 세어졌다.
	for _, must := range []string{"coordinateMarketProposals", "strategyLaneRuntime.record", "strategyLaneRuntime.runShadowStep", "strategyLaneProjection"} {
		if len(users[must]) == 0 {
			sort.Strings(names)
			t.Errorf("the census did not see %s use a shadow type (seen: %v)", must, names)
		}
	}
}

// 양성 대조: 세탁 접근자(새 메서드가 묶음을 any 로 돌려주고 주문 경로가 그것을 받는 모양 — 보이스 3 P1-A)를 같은 걸음이 잡는다.
func TestTheShadowCensusCatchesALaunderingAccessor(t *testing.T) {
	const source = `package launder
type strategyShadowBatch struct{ inputs []int }
type strategyProposalMarketAuthority struct{ hidden any }
func (authority strategyProposalMarketAuthority) carried() any { return strategyShadowBatch{} }
type wrapper[T any] struct{ value T }
func (authority strategyProposalMarketAuthority) dispatchHandoffs() []any {
	held := wrapper[strategyShadowBatch]{}
	_ = held
	return []any{authority.carried()}
}
func (authority strategyProposalMarketAuthority) authorityForOwnerScope() func() strategyShadowBatch { return nil }
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "launder.go", source, 0)
	if err != nil {
		t.Fatal(err)
	}
	info := &types.Info{Types: map[ast.Expr]types.TypeAndValue{}, Uses: map[*ast.Ident]types.Object{}, Defs: map[*ast.Ident]types.Object{}}
	if _, err := (&types.Config{Importer: importer.Default()}).Check("launder", fset, []*ast.File{file}, info); err != nil {
		t.Fatal(err)
	}
	users := a112ShadowUsers([]*ast.File{file}, info, map[string]bool{"launder.strategyShadowBatch": true})
	for _, want := range []string{"strategyProposalMarketAuthority.carried", "strategyProposalMarketAuthority.dispatchHandoffs",
		"strategyProposalMarketAuthority.authorityForOwnerScope"} {
		if len(users[want]) == 0 {
			t.Errorf("the census missed %s (generic instance · func value · accessor) — users=%v", want, users)
		}
	}
}

func a112Decl(t *testing.T, files []*ast.File, name string) *ast.FuncDecl {
	t.Helper()
	var found *ast.FuncDecl
	for _, file := range files {
		for _, decl := range file.Decls {
			if function, ok := decl.(*ast.FuncDecl); ok && a112FuncName(function) == name {
				if found != nil {
					t.Fatalf("%s is declared twice in the production build", name)
				}
				found = function
			}
		}
	}
	if found == nil {
		t.Fatalf("%s is not declared in the production build", name)
	}
	return found
}

func a112Calls(node ast.Node) []string {
	var calls []string
	ast.Inspect(node, func(child ast.Node) bool {
		if call, ok := child.(*ast.CallExpr); ok {
			calls = append(calls, calleeText(call.Fun))
		}
		return true
	})
	return calls
}

// ① 주기 함수의 shadow 식은 evaluate 의 마지막 인자(index 5) 하나, 그 유일한 호출은 forMarket. forMarket 은 형제와 같은 모양.
func TestTheMarketCycleCarriesShadowOnlyAsTheLastEvaluateArgument(t *testing.T) {
	checked := a112CheckedEngine(t)
	cycle := a112Decl(t, checked.Files, "Context.runProductionStrategyMarketCycle")
	var evaluate *ast.CallExpr
	ast.Inspect(cycle.Body, func(node ast.Node) bool {
		if call, ok := node.(*ast.CallExpr); ok && strings.HasSuffix(calleeText(call.Fun), ".evaluate") {
			evaluate = call
		}
		return true
	})
	if evaluate == nil || len(evaluate.Args) != 6 {
		t.Fatalf("evaluate call missing or has %d args, want 6 (shadow appended at index 5)", len(evaluate.Args))
	}
	carried := evaluate.Args[5]
	if got := types.ExprString(carried); got != "fresh.shadow.forMarket(market)" {
		t.Fatalf("evaluate Args[5]=%s, want fresh.shadow.forMarket(market)", got)
	}
	callees := a112CalleeExprs(cycle.Body)
	ast.Inspect(cycle.Body, func(node ast.Node) bool {
		expression, ok := node.(ast.Expr)
		if !ok || callees[expression] {
			return true
		}
		if expression.Pos() >= carried.Pos() && expression.End() <= carried.End() {
			return false
		}
		if tv, ok := checked.Info.Types[expression]; ok && a112ContainsShadowType(tv.Type, a112ShadowTypeNames, map[types.Type]bool{}) {
			t.Errorf("a shadow-typed expression %s sits outside evaluate Args[5]", types.ExprString(expression))
		}
		return true
	})
	if calls := a112Calls(carried); strings.Join(calls, ",") != "fresh.shadow.forMarket" {
		t.Fatalf("Args[5] calls %v, want only the accessor fresh.shadow.forMarket", calls)
	}
	for _, name := range []string{"strategyShadowPair.forMarket", "strategyProposalAuthorityPair.forMarket"} {
		decl := a112Decl(t, checked.Files, name)
		if calls := a112Calls(decl.Body); len(calls) != 0 {
			t.Errorf("%s calls %v — a pair accessor compares the market and returns a field", name, calls)
		}
		shape := []string{}
		for _, statement := range decl.Body.List {
			switch statement.(type) {
			case *ast.IfStmt:
				shape = append(shape, "if")
			case *ast.ReturnStmt:
				shape = append(shape, "return")
			default:
				shape = append(shape, "other")
			}
		}
		if strings.Join(shape, ",") != "if,if,return" {
			t.Errorf("%s shape=%v, want the sibling's if,if,return", name, shape)
		}
	}
}

// ④ evaluate 안의 묶음: record 로 넘기는 인자 하나(메서드 호출 · 순회 0). ⑤ record 의 칸 대입은 Lock 뒤 · 파도 증가 뒤, defer Unlock.
func TestTheLaneRuntimeOnlyStoresTheShadowBatchBesideTheWave(t *testing.T) {
	checked := a112CheckedEngine(t)
	evaluate := a112Decl(t, checked.Files, "strategyLaneRuntime.evaluate")
	uses := 0
	ast.Inspect(evaluate.Body, func(node ast.Node) bool {
		switch value := node.(type) {
		case *ast.SelectorExpr:
			if ident, ok := value.X.(*ast.Ident); ok && ident.Name == "shadow" {
				t.Errorf("evaluate selects %s on the shadow batch — it may only pass it to record", types.ExprString(value))
			}
		case *ast.RangeStmt:
			if ident, ok := value.X.(*ast.Ident); ok && ident.Name == "shadow" {
				t.Error("evaluate ranges over the shadow batch")
			}
		case *ast.CallExpr:
			for _, arg := range value.Args {
				if ident, ok := arg.(*ast.Ident); ok && ident.Name == "shadow" {
					uses++
					if calleeText(value.Fun) != "runtime.record" {
						t.Errorf("evaluate passes the shadow batch to %s, want only runtime.record", calleeText(value.Fun))
					}
				}
			}
		}
		return true
	})
	if uses != 1 {
		t.Fatalf("evaluate passes the shadow batch %d times, want exactly once (to record)", uses)
	}
	record := a112Decl(t, checked.Files, "strategyLaneRuntime.record")
	lock, wave, cell, deferred := -1, -1, -1, false
	for index, statement := range record.Body.List {
		text := strings.Join(strings.Fields(types.ExprString(a112StatementExpr(statement))), "")
		switch {
		case strings.Contains(text, "runtime.mu.Lock"):
			lock = index
		case strings.HasPrefix(text, "deferruntime.mu.Unlock"):
			deferred = true
		}
		if ifStatement, ok := statement.(*ast.IfStmt); ok && strings.Contains(types.ExprString(ifStatement.Cond), "runtime.waves[market]") {
			wave = index
		}
		if assign, ok := statement.(*ast.AssignStmt); ok && len(assign.Lhs) == 1 && types.ExprString(assign.Lhs[0]) == "runtime.shadowCells[market]" {
			cell = index
		}
	}
	if !(lock >= 0 && deferred && lock < wave && wave < cell) {
		t.Fatalf("record: lock=%d wave=%d cell=%d deferUnlock=%v — the cell must be written after the wave increment under the same lock",
			lock, wave, cell, deferred)
	}
}

// a112StatementExpr 는 문장 하나를 비교용 식 문자열로 근사한다(defer · 식 문장 · 그 밖).
func a112StatementExpr(statement ast.Stmt) ast.Expr {
	switch value := statement.(type) {
	case *ast.ExprStmt:
		return value.X
	case *ast.DeferStmt:
		return &ast.Ident{Name: "defer" + types.ExprString(value.Call)}
	}
	return &ast.Ident{Name: "-"}
}

// ⓐⓑⓒ + helper · 생성자 shape.
func TestTheCoordinatorCollectsWithOneStatementAndReturnsTheAbsentValueOnCollision(t *testing.T) {
	checked := a112CheckedEngine(t)
	coordinate := a112Decl(t, checked.Files, "coordinateMarketProposals")
	var collects, admits []token.Pos
	declared := false
	ast.Inspect(coordinate.Body, func(node ast.Node) bool {
		switch value := node.(type) {
		case *ast.AssignStmt:
			if len(value.Lhs) == 1 && types.ExprString(value.Lhs[0]) == "shadow" && value.Tok == token.DEFINE &&
				a112Source(checked.Fset, value.Rhs[0]) == "strategyShadowBatch{observed: true}" {
				declared = true
			}
		case *ast.ExprStmt:
			if call, ok := value.X.(*ast.CallExpr); ok && calleeText(call.Fun) == "shadow.collect" {
				collects = append(collects, value.Pos())
			}
		case *ast.CallExpr:
			if calleeText(value.Fun) == "gate.admit" {
				admits = append(admits, value.Pos())
			}
		}
		return true
	})
	if !declared || len(collects) != 1 || len(admits) != 1 || collects[0] > admits[0] {
		t.Fatalf("declared=%v collects=%d admits=%d — one statement-level shadow.collect before gate.admit", declared, len(collects), len(admits))
	}
	var returns []*ast.ReturnStmt
	ast.Inspect(coordinate.Body, func(node ast.Node) bool {
		if value, ok := node.(*ast.ReturnStmt); ok {
			returns = append(returns, value)
		}
		return true
	})
	if len(returns) != 2 {
		t.Fatalf("coordinateMarketProposals has %d returns, want 2 (collision · normal)", len(returns))
	}
	if got := types.ExprString(returns[0].Results[2]); got != "strategyShadowBatch{}" {
		t.Errorf("ⓑ the collision return carries %s, want the absent value strategyShadowBatch{}", got)
	}
	if got := types.ExprString(returns[1].Results[2]); got != "shadow" {
		t.Errorf("ⓒ the normal return carries %s, want the collected shadow", got)
	}
	collect := a112Decl(t, checked.Files, "strategyShadowBatch.collect")
	if len(collect.Body.List) != 1 || types.ExprString(collect.Body.List[0].(*ast.AssignStmt).Rhs[0]) !=
		"append(batch.inputs, strategyworker.NewShadowInput(proposal))" {
		t.Errorf("the collection helper must be one append of the ShadowInput constructor")
	}
	if star, ok := collect.Recv.List[0].Type.(*ast.StarExpr); !ok || types.ExprString(star.X) != "strategyShadowBatch" {
		t.Error("the collection helper must have a pointer receiver on the local batch")
	}
	worker, err := parser.ParseFile(token.NewFileSet(), filepath.Join("..", "..", "strategyworker", "shadow.go"), nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	constructor := a112Decl(t, []*ast.File{worker}, "NewShadowInput")
	if len(constructor.Body.List) != 1 || len(a112Calls(constructor.Body)) != 0 {
		t.Error("NewShadowInput must be one return of a composite literal with no call")
	}
	// collectMarket: shadow 는 첫 문장의 부재 값 대입과 조정 바로 뒤의 대입, 그 둘뿐.
	collectMarket := a112Decl(t, checked.Files, "strategyProposalAuthorityLoader.collectMarket")
	var stores []string
	ast.Inspect(collectMarket.Body, func(node ast.Node) bool {
		if assign, ok := node.(*ast.AssignStmt); ok && len(assign.Lhs) == 1 && types.ExprString(assign.Lhs[0]) == "*shadow" {
			stores = append(stores, types.ExprString(assign.Rhs[0]))
		}
		return true
	})
	first, ok := collectMarket.Body.List[0].(*ast.AssignStmt)
	if !ok || types.ExprString(first.Lhs[0]) != "*shadow" || types.ExprString(first.Rhs[0]) != "strategyShadowBatch{}" ||
		strings.Join(stores, " | ") != "strategyShadowBatch{} | collected.boundTo(loader.shadowConfig(market, schedule, routes))" {
		t.Fatalf("collectMarket shadow stores=%v — want the absent value first, then the coordinated batch with its binding", stores)
	}
}

// cycle 클로저: 두 생산 자리 → productionStrategyCycle 하나 → strategyCycleWithShadow(run = 주기 함수). 몸통: recover 0, defer 는
// `if !returnedNil { 접근자().invalidateShadow(market) }` 하나, run 호출 < returnedNil=true < startShadowStep. 접근자 · 폐기 shape.
func TestTheCycleClosureStartsTheShadowOnlyAfterANilCycleAndDiscardsOtherwise(t *testing.T) {
	checked := a112CheckedEngine(t)
	for _, name := range []string{"Context.NewRefreshingPairedStrategyEntrySupervisor", "Context.productionStrategyWorker"} {
		decl := a112Decl(t, checked.Files, name)
		found := false
		for _, call := range a112Calls(decl.Body) {
			found = found || call == "c.productionStrategyCycle"
			if call == "c.runProductionStrategyMarketCycle" {
				t.Errorf("%s calls the market cycle directly — it must go through productionStrategyCycle", name)
			}
		}
		if !found {
			t.Errorf("%s does not use productionStrategyCycle", name)
		}
	}
	production := a112Decl(t, checked.Files, "Context.productionStrategyCycle")
	if calls := a112Calls(production.Body); strings.Join(calls, ",") != "strategyCycleWithShadow,c.runProductionStrategyMarketCycle" {
		t.Fatalf("productionStrategyCycle calls %v, want strategyCycleWithShadow wrapping c.runProductionStrategyMarketCycle", calls)
	}
	wrapper := a112Decl(t, checked.Files, "strategyCycleWithShadow")
	var defers []*ast.DeferStmt
	var run, start, flag token.Pos
	ast.Inspect(wrapper.Body, func(node ast.Node) bool {
		switch value := node.(type) {
		case *ast.DeferStmt:
			defers = append(defers, value)
		case *ast.CallExpr:
			switch calleeText(value.Fun) {
			case "recover":
				t.Error("strategyCycleWithShadow must not recover — the panic belongs to invokeStrategyCycle")
			case "run":
				run = value.Pos()
			}
			if strings.HasSuffix(calleeText(value.Fun), ".startShadowStep") {
				start = value.Pos()
			}
		case *ast.AssignStmt:
			if types.ExprString(value.Lhs[0]) == "returnedNil" && types.ExprString(value.Rhs[0]) == "true" {
				flag = value.Pos()
			}
		}
		return true
	})
	if len(defers) != 1 {
		t.Fatalf("strategyCycleWithShadow has %d defers, want 1", len(defers))
	}
	literal, ok := defers[0].Call.Fun.(*ast.FuncLit)
	if !ok || len(literal.Body.List) != 1 {
		t.Fatal("the defer must be one func literal with one statement")
	}
	guard, ok := literal.Body.List[0].(*ast.IfStmt)
	if !ok || types.ExprString(guard.Cond) != "!returnedNil" || guard.Else != nil || len(guard.Body.List) != 1 {
		t.Fatalf("the defer must be `if !returnedNil { c.strategyLaneRuntimeIfAny().invalidateShadow(market) }`")
	}
	if call, ok := guard.Body.List[0].(*ast.ExprStmt); !ok || types.ExprString(call.X) != "c.strategyLaneRuntimeIfAny().invalidateShadow(market)" {
		t.Fatalf("the defer body must be the one call c.strategyLaneRuntimeIfAny().invalidateShadow(market)")
	}
	if !(run.IsValid() && flag.IsValid() && start.IsValid() && run < flag && flag < start) {
		t.Fatalf("order run=%v flag=%v start=%v — the shadow starts only after a nil cycle", run, flag, start)
	}
	accessor := a112Decl(t, checked.Files, "Context.strategyLaneRuntimeIfAny")
	if calls := a112Calls(accessor.Body); strings.Join(calls, ",") != "c.strategyLanesMu.Lock,c.strategyLanesMu.Unlock" {
		t.Errorf("the runtime accessor calls %v — lock and read only", calls)
	}
	invalidate := a112Decl(t, checked.Files, "strategyLaneRuntime.invalidateShadow")
	if calls := a112Calls(invalidate.Body); strings.Join(calls, ",") != "runtime.mu.Lock,runtime.mu.Unlock,delete" {
		t.Errorf("invalidateShadow calls %v — nil guard, lock, epoch++, delete only", calls)
	}
}

// ③ shadow 단계 호출 폐포(엔진 안 정적 호출을 전이로 따라감): 금지 함수 · 공유 캐시 필드 0, 하한 둘(적재기 · 판정).
func TestTheShadowStepClosureNeverRefreshesCoordinatesDispatchesOrWrites(t *testing.T) {
	checked := a112CheckedEngine(t)
	decls := map[string]*ast.FuncDecl{}
	for _, file := range checked.Files {
		for _, decl := range file.Decls {
			if function, ok := decl.(*ast.FuncDecl); ok {
				decls[a112FuncName(function)] = function
			}
		}
	}
	closure := func(roots ...string) (map[string]bool, map[string]bool) {
		callees, fields := map[string]bool{}, map[string]bool{}
		stack, seen := append([]string(nil), roots...), map[string]bool{}
		for len(stack) > 0 {
			name := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if seen[name] {
				continue
			}
			seen[name] = true
			decl := decls[name]
			if decl == nil {
				continue
			}
			ast.Inspect(decl.Body, func(node ast.Node) bool {
				ident, ok := node.(*ast.Ident)
				if !ok {
					return true
				}
				switch object := checked.Info.Uses[ident].(type) {
				case *types.Func:
					full := object.FullName()
					callees[full] = true
					if object.Pkg() != nil && object.Pkg().Path() == a112EnginePath {
						key := object.Name()
						if signature, ok := object.Type().(*types.Signature); ok && signature.Recv() != nil {
							receiver := signature.Recv().Type()
							if pointer, ok := receiver.(*types.Pointer); ok {
								receiver = pointer.Elem()
							}
							if named, ok := receiver.(*types.Named); ok {
								key = named.Obj().Name() + "." + key
							}
						}
						stack = append(stack, key)
					}
				case *types.Var:
					if object.IsField() {
						fields[object.Name()] = true
					}
				}
				return true
			})
		}
		return callees, fields
	}
	callees, fields := closure("strategyLaneRuntime.startShadowStep")
	forbidden := []string{"refreshPairedStrategyEntryProductionAssembly", "NewPairedStrategyEntryProductionAssembly", "joinStrategyRefreshWave",
		"collectStrategyRefreshWave", "awaitStrategyRefreshWave", "loadFamilyActivation", "familyGateFor", "LoadProductionFamilyActivation",
		"coordinateMarketProposals", "Submit", "dispatch", "Record", "Recover", "Place", "EncodeProductionFamilyActivation"}
	for callee := range callees {
		for _, word := range forbidden {
			if strings.Contains(callee, word) && !strings.Contains(callee, "strategyshadow") {
				t.Errorf("the shadow step closure reaches %s (forbidden: %s)", callee, word)
			}
		}
		if strings.Contains(callee, "/internal/journal.") {
			t.Errorf("the shadow step closure reaches the journal: %s", callee)
		}
	}
	for _, field := range []string{"strategyRefresh", "strategyRefreshAt", "strategyRefreshWave"} {
		if fields[field] {
			t.Errorf("the shadow step closure uses the shared cache field %s", field)
		}
	}
	for _, must := range []string{"strategyshadow.LoadProductionFamilyShadow", "ShadowOutcomeOver"} {
		found := false
		for callee := range callees {
			found = found || strings.Contains(callee, must)
		}
		if !found {
			t.Errorf("lower bound: the closure walk never reached %s — it read nothing", must)
		}
	}
	// 양성 대조: 같은 걸음이 주기 함수에서는 refresh 를 본다.
	control, _ := closure("Context.runProductionStrategyMarketCycle")
	seen := false
	for callee := range control {
		seen = seen || strings.Contains(callee, "refreshPairedStrategyEntryProductionAssembly")
	}
	if !seen {
		t.Fatal("control: the walk from the market cycle did not reach the refresh — the walker is blind")
	}
}

// a112Source 는 식의 원문(printer)이다 — types.ExprString 은 composite literal 원소를 「…」 로 줄인다.
func a112Source(fset *token.FileSet, node ast.Node) string {
	var buffer bytes.Buffer
	if err := printer.Fprint(&buffer, fset, node); err != nil {
		return "<unprintable>"
	}
	return buffer.String()
}

// 운반의 나머지 두 자리 · 게시 조건: collect 의 회복 갈래는 묶음을 부재 값으로 되돌리고, 조립은 collect 의 둘째 값을 shadow 필드에 그대로
// 담고, 게시는 세대 ∧ 파도 둘 다를 본다(파도 쪽은 투영의 신선도 규칙과 겹치므로 행동 시험이 아니라 여기서 못 박는다).
func TestTheRemainingCarryAndPublishSitesKeepTheirShape(t *testing.T) {
	checked := a112CheckedEngine(t)
	collect := a112Decl(t, checked.Files, "strategyProposalAuthorityLoader.collect")
	reset := false
	ast.Inspect(collect.Body, func(node ast.Node) bool {
		if call, ok := node.(*ast.CallExpr); ok && calleeText(call.Fun) == "recover" {
			reset = true
		}
		return true
	})
	resets := 0
	ast.Inspect(collect.Body, func(node ast.Node) bool {
		if assign, ok := node.(*ast.AssignStmt); ok && len(assign.Lhs) == 1 && types.ExprString(assign.Lhs[0]) == "shadow" &&
			a112Source(checked.Fset, assign.Rhs[0]) == "strategyShadowBatch{}" {
			resets++
		}
		return true
	})
	if !reset || resets != 1 {
		t.Fatalf("collect's recover must reset the shadow batch to the absent value (recover=%v resets=%d)", reset, resets)
	}
	assembly := a112Decl(t, checked.Files, "Context.NewPairedStrategyEntryProductionAssembly")
	bound, stored := false, false
	ast.Inspect(assembly.Body, func(node ast.Node) bool {
		switch value := node.(type) {
		case *ast.AssignStmt:
			if len(value.Lhs) == 2 && types.ExprString(value.Lhs[1]) == "shadowAuthority" && strings.HasSuffix(a112Source(checked.Fset, value.Rhs[0]), "collect(ctx, scheduleAuthority, routeAuthority, fxAuthority)") {
				bound = true
			}
		case *ast.KeyValueExpr:
			if types.ExprString(value.Key) == "shadow" && types.ExprString(value.Value) == "shadowAuthority" {
				stored = true
			}
		}
		return true
	})
	if !bound || !stored {
		t.Fatalf("the assembly must take collect's shadow pair (bound=%v) into its shadow field (stored=%v)", bound, stored)
	}
	publish := a112Decl(t, checked.Files, "strategyLaneRuntime.publishShadow")
	guard, ok := publish.Body.List[2].(*ast.IfStmt)
	if !ok || a112Source(checked.Fset, guard.Cond) != "runtime.shadowEpochs[market] != epoch || runtime.shadowCells[market].wave != wave" {
		t.Fatal("publishShadow must compare both the epoch and the wave it copied at the start (CAS)")
	}
}
