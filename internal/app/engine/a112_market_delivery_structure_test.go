package engine

// a112 5.2.2.1 리뷰 수리(codex P1-1 · 보이스 B #2) — 주문 경로의 배선을 **이름이 아니라 식별자 해소(go/types)** 로 못 박는다.
//
// 앞 판본의 구조 못은 호출식의 이름 문자열만 셌다. 보이스 B 가 세 우회를 실측했다: 이름만 같은 다른 타입의 `dispatchHandoffs`
// (X08) · 몸통 closure 안 카운터로 둘째 범위부터 버림(X09) · `deliverEachStrategyHandoff := …` 지역 섀도(X10). 셋 다 이름은
// 같으므로 이름 세기는 원리적으로 못 본다. 여기서는 각 식별자가 **무엇으로 해소되는지** 본다:
//   ① `runProductionStrategyMarketCycle` 의 마지막 문장 = `return <패키지 수준 func dispatchStrategyMarketHandoffs>(ctx,
//      <Context.Journal 필드>, <StrategyEntryProductionAssembly.dispatch 필드>, <handoff 원천>)`.
//      handoff 원천 = `strategyProposalMarketAuthority.dispatchHandoffs` 메서드 호출, 또는 그 호출을 한 번만 대입받은 지역 변수
//      (동작이 같은 리팩터를 거짓 양성으로 막지 않으려고 — 보이스 B X11).
//   ② `dispatchStrategyMarketHandoffs` 의 본문 = `return <패키지 수준 func deliverEachStrategyHandoff>(<매개변수 handoffs>, <함수 리터럴>)`
//      하나, 그리고 그 함수 리터럴은 **자기 밖에서 선언된 변수에 쓰지 않는다**(캡처 쓰기 금지 — X09 의 모양).
// 위치는 절대 좌표가 아니라 감싼 선언 기준이다(「함수의 마지막 문장」 · 「본문의 유일한 문장」).
//
// 타입 검사는 엔진 생산 파일만으로 한다(의존 패키지는 빈 스텁 — 그쪽 식별자는 해소되지 않아 오류가 나지만 무시한다). 여기서
// 해소가 필요한 것은 전부 이 패키지 안의 선언이다.
//
// **이 파일은 2차 방어이고 동결됐다**(Manager 판정 2026-10-01). 의미 — 「dispatch 가 이 시장의 handoff 를 전부 받는다」 — 는
// 주기 함수를 통째로 도는 행동 시험 a112_market_cycle_delivery_test.go 가 잰다(어떤 철자의 우회든 dispatch 수가 준다). 여기의
// 모양 검사를 더 조이지 않는다. 이름 붙여 두는 미해소 모양(codex 재확인 #2): 캡처 쓰기 금지는 **식별자 대입 · 증감만** 본다 —
// 포인터 경유(`(*p)++`) · 필드(`s.n++`) · 맵 원소(`m[k]++`) · 채널 송신(`ch <- v`)은 보지 않고, 스텁 importer 로 해소되지 않는
// 외부 패키지 식별자(dot import 변수 등)는 「밖」으로 판정하지 못해 허용된다. `handoffSource` 는 지역 변수의 **대입**만 세므로
// 원소 변경(`clear(hs[1:])`) · 다른 호출로의 전달은 보지 않는다 — 그 우회는 위 행동 시험이 잡는다(변이 원장 H 세트).

import (
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"go/types"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

type a112StubImporter struct{}

func (a112StubImporter) Import(importPath string) (*types.Package, error) {
	pkg := types.NewPackage(importPath, path.Base(importPath))
	pkg.MarkComplete()
	return pkg, nil
}

type a112EngineTypes struct {
	pkg   *types.Package
	info  *types.Info
	decls map[string]*ast.FuncDecl
}

func a112TypeCheckEngine(t *testing.T) a112EngineTypes {
	t.Helper()
	fset := token.NewFileSet()
	var files []*ast.File
	decls := map[string]*ast.FuncDecl{}
	for _, name := range engineProductionFiles(t) {
		// 기본 빌드 구성(태그 없음, 이 GOOS)의 파일만 — 생산 이진이 보는 선언 집합.
		if ok, err := build.Default.MatchFile(filepath.Dir(name), filepath.Base(name)); err != nil || !ok {
			continue
		}
		file, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		files = append(files, file)
		for _, decl := range file.Decls {
			if function, ok := decl.(*ast.FuncDecl); ok {
				key := function.Name.Name
				if function.Recv != nil {
					key = types.ExprString(function.Recv.List[0].Type) + "." + key
				}
				decls[key] = function
			}
		}
	}
	info := &types.Info{Uses: map[*ast.Ident]types.Object{}, Defs: map[*ast.Ident]types.Object{},
		Selections: map[*ast.SelectorExpr]*types.Selection{}}
	config := types.Config{Importer: a112StubImporter{}, Error: func(error) {}}
	pkg, _ := config.Check("github.com/JungHoonGhae/tossinvest-cli/internal/app/engine", fset, files, info)
	if pkg == nil {
		t.Fatal("engine package did not type-check at all")
	}
	return a112EngineTypes{pkg: pkg, info: info, decls: decls}
}

// packageFunc 는 식별자가 이 패키지 수준의 func `name` 으로 해소되는지 답함.
func (e a112EngineTypes) packageFunc(ident *ast.Ident, name string) bool {
	object, ok := e.info.Uses[ident].(*types.Func)
	return ok && object.Name() == name && object.Pkg() == e.pkg && object.Parent() == e.pkg.Scope()
}

// selectsFrom 은 선택자가 `owner` 타입의 `name` (필드 또는 메서드)로 해소되는지 답함.
func (e a112EngineTypes) selectsFrom(selector *ast.SelectorExpr, owner, name string) bool {
	selection := e.info.Selections[selector]
	if selection == nil || selection.Obj().Name() != name || selection.Obj().Pkg() != e.pkg {
		return false
	}
	receiver := selection.Recv()
	if pointer, ok := receiver.(*types.Pointer); ok {
		receiver = pointer.Elem()
	}
	named, ok := receiver.(*types.Named)
	return ok && named.Obj().Name() == owner && named.Obj().Pkg() == e.pkg
}

func (e a112EngineTypes) isDispatchHandoffsCall(expr ast.Expr) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok || len(call.Args) != 0 {
		return false
	}
	selector, ok := call.Fun.(*ast.SelectorExpr)
	return ok && e.selectsFrom(selector, "strategyProposalMarketAuthority", "dispatchHandoffs")
}

// handoffSource 는 인자가 dispatchHandoffs 호출이거나, 그 호출을 **정확히 한 번** 대입받고 달리 쓰이지 않는 지역 변수인지 답함.
func (e a112EngineTypes) handoffSource(function *ast.FuncDecl, arg ast.Expr) bool {
	if e.isDispatchHandoffsCall(arg) {
		return true
	}
	ident, ok := arg.(*ast.Ident)
	if !ok {
		return false
	}
	variable, ok := e.info.Uses[ident].(*types.Var)
	if !ok || variable.Parent() == e.pkg.Scope() {
		return false
	}
	writes, good := 0, 0
	ast.Inspect(function.Body, func(node ast.Node) bool {
		// `var hs = …dispatchHandoffs()` 도 같은 동작이다(codex 재확인 #4 — 앞 판은 ValueSpec 을 안 세서 거짓 양성).
		if spec, ok := node.(*ast.ValueSpec); ok {
			for index, name := range spec.Names {
				if e.info.Defs[name] != variable {
					continue
				}
				writes++
				if len(spec.Values) == len(spec.Names) && e.isDispatchHandoffsCall(spec.Values[index]) {
					good++
				}
			}
			return true
		}
		assign, ok := node.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for index, left := range assign.Lhs {
			name, ok := left.(*ast.Ident)
			if !ok || (e.info.Defs[name] != variable && e.info.Uses[name] != variable) {
				continue
			}
			writes++
			if len(assign.Rhs) == len(assign.Lhs) && e.isDispatchHandoffsCall(assign.Rhs[index]) {
				good++
			}
		}
		return true
	})
	return writes == 1 && good == 1
}

func TestTheProductionCycleEndsByDeliveringEveryOwnerScopeHandoff(t *testing.T) {
	engine := a112TypeCheckEngine(t)
	cycle := engine.decls["*Context.runProductionStrategyMarketCycle"]
	if cycle == nil || len(cycle.Body.List) == 0 {
		t.Fatal("runProductionStrategyMarketCycle not found")
	}
	last, ok := cycle.Body.List[len(cycle.Body.List)-1].(*ast.ReturnStmt)
	if !ok || len(last.Results) != 1 {
		t.Fatalf("the cycle's last statement is %T, want `return dispatchStrategyMarketHandoffs(…)`", cycle.Body.List[len(cycle.Body.List)-1])
	}
	call, ok := last.Results[0].(*ast.CallExpr)
	if !ok || len(call.Args) != 4 {
		t.Fatalf("the cycle's last statement returns %s, want a 4-argument delivery call", types.ExprString(last.Results[0]))
	}
	callee, ok := call.Fun.(*ast.Ident)
	if !ok || !engine.packageFunc(callee, "dispatchStrategyMarketHandoffs") {
		t.Fatalf("the cycle delivers through %s, which does not resolve to the package-level dispatchStrategyMarketHandoffs",
			types.ExprString(call.Fun))
	}
	if journal, ok := call.Args[1].(*ast.SelectorExpr); !ok || !engine.selectsFrom(journal, "Context", "Journal") {
		t.Errorf("campaign reader argument %s is not the Context's own journal", types.ExprString(call.Args[1]))
	}
	if dispatch, ok := call.Args[2].(*ast.SelectorExpr); !ok || !engine.selectsFrom(dispatch, "StrategyEntryProductionAssembly", "dispatch") {
		t.Errorf("dispatcher argument %s is not the refreshed assembly's dispatch cycle", types.ExprString(call.Args[2]))
	}
	if !engine.handoffSource(cycle, call.Args[3]) {
		t.Errorf("handoff argument %s is not strategyProposalMarketAuthority.dispatchHandoffs()", types.ExprString(call.Args[3]))
	}
}

func TestTheDeliveryBodyHandsEveryHandoffToOneClosureThatWritesNothingOutside(t *testing.T) {
	engine := a112TypeCheckEngine(t)
	delivery := engine.decls["dispatchStrategyMarketHandoffs"]
	if delivery == nil || len(delivery.Body.List) != 1 {
		t.Fatal("dispatchStrategyMarketHandoffs must be a single return statement")
	}
	ret, ok := delivery.Body.List[0].(*ast.ReturnStmt)
	if !ok || len(ret.Results) != 1 {
		t.Fatal("dispatchStrategyMarketHandoffs must be a single return statement")
	}
	call, ok := ret.Results[0].(*ast.CallExpr)
	if !ok || len(call.Args) != 2 {
		t.Fatalf("delivery returns %s, want deliverEachStrategyHandoff(handoffs, func…)", types.ExprString(ret.Results[0]))
	}
	callee, ok := call.Fun.(*ast.Ident)
	if !ok || !engine.packageFunc(callee, "deliverEachStrategyHandoff") {
		t.Fatalf("delivery iterates through %s, not the package-level deliverEachStrategyHandoff", types.ExprString(call.Fun))
	}
	handoffs, ok := call.Args[0].(*ast.Ident)
	parameter := delivery.Type.Params.List[len(delivery.Type.Params.List)-1].Names[0]
	if !ok || engine.info.Uses[handoffs] == nil || engine.info.Uses[handoffs] != engine.info.Defs[parameter] {
		t.Errorf("delivery iterates over %s, not its own handoffs parameter", types.ExprString(call.Args[0]))
	}
	body, ok := call.Args[1].(*ast.FuncLit)
	if !ok {
		t.Fatalf("delivery body is %T, want a function literal", call.Args[1])
	}
	// 캡처 쓰기 금지: 리터럴 밖에서 선언된 변수에 대입 · 증감하면 호출 사이에 상태가 생긴다(「둘째부터 버림」 카운터의 모양).
	outside := func(ident *ast.Ident) bool {
		object := engine.info.Uses[ident]
		return object != nil && (object.Pos() < body.Pos() || object.Pos() > body.End())
	}
	var captured []string
	ast.Inspect(body.Body, func(node ast.Node) bool {
		switch value := node.(type) {
		case *ast.AssignStmt:
			for _, left := range value.Lhs {
				if ident, ok := ast.Unparen(left).(*ast.Ident); ok && outside(ident) {
					captured = append(captured, ident.Name)
				}
			}
		case *ast.IncDecStmt:
			if ident, ok := ast.Unparen(value.X).(*ast.Ident); ok && outside(ident) {
				captured = append(captured, ident.Name)
			}
		}
		return true
	})
	if len(captured) != 0 {
		t.Fatalf("the delivery closure writes variables captured from outside it: %s", strings.Join(captured, ","))
	}
}
