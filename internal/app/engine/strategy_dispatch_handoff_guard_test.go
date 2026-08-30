package engine

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// 태스크 5.5 가 요구하는 "compile/dependency 수준의 증명"이 이 파일이다.
//
// 여기 있는 검사는 실행 중에 무슨 일이 일어났는지 보지 않는다. 소스에 무엇이
// **쓰여 있는지**를 본다. 그래서 "그날 주문이 안 나갔다"가 아니라 "주문을 낼
// 방법이 그 자리에 적혀 있지 않다"를 증명한다.

// handoffSeamFile 은 조정자가 고른 것이 공유 dispatch 로 건너가는 유일한 자리다.
const handoffSeamFile = "strategy_dispatch_handoff.go"

// coordinatorSeamFile 은 시장 조정자를 엔진 안에서 감싸는 자리다.
const coordinatorSeamFile = "strategy_market_coordinator.go"

// dispatchCallSiteFunc 는 공유 dispatch 를 부르는 유일한 생산 함수다.
const dispatchCallSiteFunc = "runProductionStrategyMarketCycle"

// workerBuilderFunc 는 시장 worker 를 Effective 로 올리는 함수다.
const workerBuilderFunc = "buildProductionStrategyMarketWorker"

// enginePackagePrefix 는 이 모듈의 import 경로 앞자리다.
const enginePackagePrefix = "github.com/JungHoonGhae/tossinvest-cli/"

func engineProductionFiles(t *testing.T) []string {
	t.Helper()
	all, err := filepath.Glob(filepath.Join(".", "*.go"))
	if err != nil {
		t.Fatalf("engine source glob: %v", err)
	}
	files := make([]string, 0, len(all))
	for _, path := range all {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		files = append(files, path)
	}
	if len(files) == 0 {
		t.Fatal("no engine production source was scanned, so every guard below is vacuous")
	}
	sort.Strings(files)
	return files
}

func parseEngineFile(t *testing.T, path string) *ast.File {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return file
}

// singleProposalAssumptionSites 는 "한 시장의 제안은 정확히 하나"라는 가정이
// 코드에 몇 번 적혀 있는지 센다. 그 가정은 정확히 두 모양으로만 나타난다.
//
//	len(x.entries) <비교> 1
//	x.entries[0]
//
// 이 두 모양만 보므로 `len(routes.entries) == 0` 같은 다른 목록 검사는 걸리지
// 않는다. 걸리게 하면 세는 대상이 흐려져서 census 가 아무것도 고정하지 못한다.
func singleProposalAssumptionSites(file *ast.File) int {
	sites := 0
	entriesSelector := func(node ast.Expr) bool {
		selector, ok := node.(*ast.SelectorExpr)
		return ok && selector.Sel != nil && selector.Sel.Name == "entries"
	}
	isOne := func(node ast.Expr) bool {
		literal, ok := node.(*ast.BasicLit)
		return ok && literal.Value == "1"
	}
	ast.Inspect(file, func(node ast.Node) bool {
		switch value := node.(type) {
		case *ast.BinaryExpr:
			for _, pair := range [][2]ast.Expr{{value.X, value.Y}, {value.Y, value.X}} {
				call, ok := pair[0].(*ast.CallExpr)
				if !ok || len(call.Args) != 1 {
					continue
				}
				name, ok := call.Fun.(*ast.Ident)
				if !ok || name.Name != "len" || !entriesSelector(call.Args[0]) || !isOne(pair[1]) {
					continue
				}
				sites++
			}
		case *ast.IndexExpr:
			if entriesSelector(value.X) && isZero(value.Index) {
				sites++
			}
		}
		return true
	})
	return sites
}

func isZero(node ast.Expr) bool {
	literal, ok := node.(*ast.BasicLit)
	return ok && literal.Value == "0"
}

// singleProposalAssumptionCensus 는 그 가정이 남아 있어도 되는 자리와 그 수다.
//
// 이 표는 계약이 아니라 **빚 목록**이다. 동결 골든
// (analysis/goldens/four-family-runtime-v1.json)의
// queue.market_wide_single_proposal_assumption_forbidden 은 true 이고
// queue.selected_limit 는 "시장마다 하나"가 아니라 "소유자 범위마다 하나"다.
// 즉 이 가정은 결국 전부 사라져야 한다. 5.5 는 dispatch 경로의 사본을 한 곳
// (handoff seam)으로 모으는 데까지만 한다. 남은 줄은 그 사본을 없앨 로트를
// 이름으로 달고 여기 남는다 — 목록에서 지우면 조용한 생략이 된다.
var singleProposalAssumptionCensus = map[string]int{
	// L6 소유(태스크 6.2 q_final/owner admission). 그 로트가 열릴 때까지 남는다.
	// 5 는 센 값이다: len 비교 두 개와 색인 세 개.
	"strategy_account_first_leg_authority.go": 5,
}

// handoffSeamFile 이 이 표에 없는 것은 빠뜨린 것이 아니다. dispatch 경로에서
// 그 가정은 이 패키지에서 사라졌다 — 어댑터는 entries 를 순서대로 옮기기만
// 하고, 몇 개까지 되는지는 internal/strategyhandoff 의 Capacity 가 정한다.
// 그 상수는 리터럴 1 과의 비교가 아니라 이름과의 비교라서 여기 잡히지 않고,
// 잡혀서도 안 된다.

func TestTheSingleProposalAssumptionLivesOnlyWhereTheCensusSaysItDoes(t *testing.T) {
	found := make(map[string]int)
	for _, path := range engineProductionFiles(t) {
		if sites := singleProposalAssumptionSites(parseEngineFile(t, path)); sites != 0 {
			found[filepath.Base(path)] = sites
		}
	}
	unexpected := make([]string, 0)
	for name, sites := range found {
		if want, ok := singleProposalAssumptionCensus[name]; !ok || want != sites {
			unexpected = append(unexpected, name+" has "+strconv.Itoa(sites)+" site(s), census says "+strconv.Itoa(singleProposalAssumptionCensus[name]))
		}
	}
	for name, want := range singleProposalAssumptionCensus {
		if found[name] != want {
			unexpected = append(unexpected, name+" is in the census with "+strconv.Itoa(want)+" site(s) but the source has "+strconv.Itoa(found[name]))
		}
	}
	if len(unexpected) != 0 {
		sort.Strings(unexpected)
		t.Fatalf("the market-wide single-proposal assumption moved: %v", unexpected)
	}
}

// TestEveryMarketAuthorityCarryingEntriesAlsoReportsReady 는 handoff 가 준비
// 상태를 보게 되어도 지금 동작이 바뀌지 않는다는 근거다.
//
// handoff 이전에 `ResultAuthority` 와 읽기 전용 projection 은 개수만 보고
// 준비 상태를 보지 않았다. 이제 둘 다 handoff 를 쓰므로 준비 상태 검사가
// 새로 걸린다. 그 검사가 **거부하게 될 정상 입력**이 무엇인지 세어 보면
// 답은 없음이다: 이 패키지에서 항목을 담아 만드는 시장 권한은 모두 준비된
// 시장이기 때문이다. 그 사실을 문장이 아니라 소스에서 확인한다.
func TestEveryMarketAuthorityCarryingEntriesAlsoReportsReady(t *testing.T) {
	checked := 0
	for _, path := range engineProductionFiles(t) {
		file := parseEngineFile(t, path)
		ast.Inspect(file, func(node ast.Node) bool {
			literal, ok := node.(*ast.CompositeLit)
			if !ok {
				return true
			}
			name, ok := literal.Type.(*ast.Ident)
			if !ok || name.Name != "strategyProposalMarketAuthority" {
				return true
			}
			fields := compositeFields(literal)
			if _, carries := fields["entries"]; !carries {
				return true
			}
			checked++
			snapshot, ok := fields["snapshot"].(*ast.CompositeLit)
			if !ok {
				t.Errorf("%s builds a market authority with entries but no literal snapshot", filepath.Base(path))
				return true
			}
			ready, ok := compositeFields(snapshot)["Ready"].(*ast.Ident)
			if !ok || ready.Name != "true" {
				t.Errorf("%s builds a market authority carrying entries while Ready is not true; "+
					"the handoff's readiness check would now refuse it", filepath.Base(path))
			}
			return true
		})
	}
	if checked == 0 {
		t.Fatal("no production site builds a market authority with entries, so this invariant proves nothing")
	}
}

func compositeFields(literal *ast.CompositeLit) map[string]ast.Expr {
	fields := make(map[string]ast.Expr, len(literal.Elts))
	for _, element := range literal.Elts {
		pair, ok := element.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, ok := pair.Key.(*ast.Ident)
		if !ok {
			continue
		}
		fields[key.Name] = pair.Value
	}
	return fields
}

// 예전에 여기 있던 TestNoProductionSiteReadsAHandoffWithoutAskingWhetherItWasAdmitted
// 는 지웠다. 조용히 지운 것이 아니라 **더 강한 것으로 바꿔서** 지웠다.
//
// 그 검사는 handoff 값을 쓰는 함수가 Admitted() 를 부르는지만 봤고, 그것이
// 실제로 관문 역할을 하는지는 보지 않았다. 그래서 `_ = handoff.Admitted()` 로
// 호출만 남기고 관문을 지우거나, `var` 로 바인딩하거나, 아예 바인딩 없이
// 필드를 바로 읽으면 전부 통과했다 — 두 스위트가 모두 초록인 채로.
//
// 지금 그 성질은 검사가 아니라 **타입**이 지킨다. 경계 값에서 제안이 나오는
// 길은 strategyhandoff.Handoff.Single() 하나뿐이고, 그 서명이 bool 을 함께
// 준다. 무시하려면 명시적으로 버려야 한다. 새 접근자를 달아 우회하는 문은
// internal/strategyhandoff/escape_test.go 가 잠근다.

// seamImportAllowlist 는 두 seam 파일이 들여올 수 있는 것 **전부**다.
//
// 금지 목록이 아니라 허용 목록인 이유: 첫 판본은 금지 목록을
// strategy_dispatch_cycle.go 의 import 에서 베껴 왔는데, 그것은 그 파일의
// 영수증이지 이 모듈의 변경 표면이 아니다. 그래서 internal/official 이 통째로
// 빠져 있었고, 빠진 것은 아무 소리도 내지 않았다. 허용 목록은 빠뜨릴 수 없다.
var seamImportAllowlist = map[string]map[string]bool{
	coordinatorSeamFile: {
		"time": true,
		"github.com/JungHoonGhae/tossinvest-cli/internal/strategyarbiter":     true,
		"github.com/JungHoonGhae/tossinvest-cli/internal/strategycoordinator": true,
		"github.com/JungHoonGhae/tossinvest-cli/internal/strategyproposal":    true,
		"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter":      true,
	},
	handoffSeamFile: {
		"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow":    true,
		"github.com/JungHoonGhae/tossinvest-cli/internal/strategyhandoff": true,
	},
}

// mutationCapabilityPackages 는 이 모듈에서 브로커·원장·주문 의도를 실제로
// 바꿀 수 있는 패키지다. 아래 계산은 이 목록을 **타입을 분류하는 데**만 쓰므로,
// 넉넉히 잡는 쪽이 안전하다.
var mutationCapabilityPackages = []string{
	"internal/official", "internal/hybrid", "internal/client", "internal/trading",
	"internal/orderintent", "internal/execgw", "internal/journal", "internal/ops",
	"internal/protectionofficial", "internal/verifylive", "internal/config",
}

// mutationCapabilityIdents 는 순수한 타입만 노출해서 위 계산이 잡지 못하지만
// 실제로는 능력인 것들이다. 계산과 선언을 합집합으로 쓴다.
var mutationCapabilityIdents = []string{
	"strategyDispatchCycle",
	"newStrategyDispatchCycle",
	"strategyDispatchGateway",
	"strategyDispatchOwnerCoordinator",
	"strategyFirstLegAdmissionBridge",
}

// engineCapabilityTypes 는 package engine 안에서 변경 능력을 손에 쥐고 있는
// 타입 이름을 **소스에서 계산한다**.
//
// 손으로 적은 목록이 왜 안 되는지는 실측으로 드러났다. 첫 판본은 import 와
// 이름 몇 개만 금지했는데, seam 파일에 import 를 하나도 늘리지 않고
// `func (b *officialBroker) …{ b.off.CancelConditionalOrder(ctx, id) }` 를
// 넣으면 통과했다. 즉 보호 손절을 취소하는 코드가 "변경 능력 없음" 판정을
// 받았다. 능력은 import 가 아니라 **패키지 안의 타입**을 통해 들어온다.
//
// 계산은 필드 타입을 따라 고정점까지 넓힌다: 능력 패키지의 타입을 필드로
// 들고 있으면 능력이고, 능력 타입을 필드로 들고 있어도 능력이다.
func engineCapabilityTypes(t *testing.T) map[string]bool {
	t.Helper()
	type decl struct {
		name string
		expr ast.Expr
		// aliases 는 그 파일에서 능력 패키지를 가리키는 지역 이름이다.
		aliases map[string]bool
	}
	decls := make([]decl, 0)
	for _, path := range engineProductionFiles(t) {
		file := parseEngineFile(t, path)
		aliases := make(map[string]bool)
		for _, spec := range file.Imports {
			value, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				t.Fatalf("%s: unquote %s: %v", path, spec.Path.Value, err)
			}
			for _, capability := range mutationCapabilityPackages {
				if value != enginePackagePrefix+capability {
					continue
				}
				local := value[strings.LastIndex(value, "/")+1:]
				if spec.Name != nil {
					local = spec.Name.Name
				}
				aliases[local] = true
			}
		}
		ast.Inspect(file, func(node ast.Node) bool {
			spec, ok := node.(*ast.TypeSpec)
			if !ok {
				return true
			}
			decls = append(decls, decl{name: spec.Name.Name, expr: spec.Type, aliases: aliases})
			return true
		})
	}
	if len(decls) == 0 {
		t.Fatal("no engine type declaration was scanned, so the capability computation reads nothing")
	}
	marked := make(map[string]bool)
	for changed := true; changed; {
		changed = false
		for _, value := range decls {
			if marked[value.name] {
				continue
			}
			holds := false
			ast.Inspect(value.expr, func(node ast.Node) bool {
				switch inner := node.(type) {
				case *ast.SelectorExpr:
					if pkg, ok := inner.X.(*ast.Ident); ok && value.aliases[pkg.Name] {
						holds = true
					}
				case *ast.Ident:
					if marked[inner.Name] {
						holds = true
					}
				}
				return !holds
			})
			if holds {
				marked[value.name] = true
				changed = true
			}
		}
	}
	// 양성 대조. 계산이 조용히 빈 집합을 돌려주면 아래 단언은 **틀린 이유로**
	// 통과한다. officialBroker 는 이 패키지의 유일한 주문 경로이므로 반드시
	// 잡혀야 하고, 잡히지 않으면 계산이 고장난 것이다.
	for _, required := range []string{"officialBroker", "Context"} {
		if !marked[required] {
			names := make([]string, 0, len(marked))
			for name := range marked {
				names = append(names, name)
			}
			sort.Strings(names)
			t.Fatalf("capability computation missed %s; it found only %v", required, names)
		}
	}
	return marked
}

// TestTheCoordinatorAndHandoffSeamsHoldNoMutationCapability 는 두 seam 파일이
// 주문이나 lease 를 만들 수 있는 무엇도 손에 넣지 못한다는 것을 증명한다.
//
// 이 검사가 증명할 수 있는 범위를 정확히 적어 둔다. 파일 단위 import 허용
// 목록은 그 파일이 밖에서 무엇을 들여오는지 전부 고정한다. 같은 패키지 안의
// 능력은 import 로 막을 수 없으므로 계산된 타입 이름으로 막는다. 그래도
// package engine 안에 있는 한 이것은 **파일 규율**이지 컴파일러 보장이 아니다.
// 컴파일러 보장이 필요한 자리 — 무엇이 건너갈지 고르는 판단 — 는 그래서
// 이 패키지 밖(internal/strategyhandoff)으로 옮겼고, 거기서는 import 폐쇄가
// 결정적이다.
func TestTheCoordinatorAndHandoffSeamsHoldNoMutationCapability(t *testing.T) {
	capabilityTypes := engineCapabilityTypes(t)
	for _, name := range mutationCapabilityIdents {
		capabilityTypes[name] = true
	}
	for _, name := range []string{coordinatorSeamFile, handoffSeamFile} {
		path := filepath.Join(".", name)
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("guarded seam %s is missing: %v", name, err)
		}
		allowed, ok := seamImportAllowlist[name]
		if !ok {
			t.Fatalf("seam %s has no import allowlist, so nothing constrains what it may reach", name)
		}
		file := parseEngineFile(t, path)
		for _, spec := range file.Imports {
			value, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				t.Fatalf("%s: unquote %s: %v", name, spec.Path.Value, err)
			}
			if !allowed[value] {
				t.Errorf("%s imports %s, which is outside its allowed closure", name, value)
			}
		}
		ast.Inspect(file, func(node ast.Node) bool {
			ident, ok := node.(*ast.Ident)
			if !ok {
				return true
			}
			if capabilityTypes[ident.Name] {
				t.Errorf("%s names %s, which holds a mutation capability", name, ident.Name)
			}
			return true
		})
	}
}

func engineFuncDecl(t *testing.T, file *ast.File, name string) *ast.FuncDecl {
	t.Helper()
	for _, decl := range file.Decls {
		function, ok := decl.(*ast.FuncDecl)
		if ok && function.Name != nil && function.Name.Name == name {
			return function
		}
	}
	t.Fatalf("%s is not declared in the parsed file", name)
	return nil
}

// TestTheWorkerBuilderOnlyObservesThroughTheGateway 는 worker 승격 경로가
// 게이트웨이를 **읽기만** 한다는 것을 증명한다. worker 가 Effective 로 올라가는
// 판단 안에서 주문을 낼 수 있으면, 그 판단은 더 이상 관측이 아니다.
func TestTheWorkerBuilderOnlyObservesThroughTheGateway(t *testing.T) {
	file := parseEngineFile(t, filepath.Join(".", "strategy_entry_supervisor.go"))
	builder := engineFuncDecl(t, file, workerBuilderFunc)
	gatewayParam := ""
	for _, field := range builder.Type.Params.List {
		ident, ok := field.Type.(*ast.Ident)
		if !ok || ident.Name != "strategyDispatchGateway" || len(field.Names) != 1 {
			continue
		}
		gatewayParam = field.Names[0].Name
	}
	if gatewayParam == "" {
		t.Fatalf("%s no longer takes a single named strategyDispatchGateway, so this guard reads nothing", workerBuilderFunc)
	}
	observed := 0
	ast.Inspect(builder.Body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		receiver, ok := selector.X.(*ast.Ident)
		if !ok || receiver.Name != gatewayParam {
			return true
		}
		if !strings.HasPrefix(selector.Sel.Name, "Observe") {
			t.Errorf("%s calls %s.%s: the worker builder may only observe, never place",
				workerBuilderFunc, gatewayParam, selector.Sel.Name)
			return true
		}
		observed++
		return true
	})
	if observed == 0 {
		t.Fatalf("%s makes no gateway observation, so this guard proves nothing", workerBuilderFunc)
	}
}

// TestExactlyOneProductionCallSiteTurnsAHandoffIntoADispatch 는 "하나의
// bounded handoff" 를 글자 그대로 고정한다. 부르는 자리가 둘이 되는 순간
// 두 자리가 서로 다른 조건을 갖게 되고, 그 차이는 아무도 보고하지 않는다.
//
// 첫 판본은 두 가지를 놓쳤다. (1) `dispatch` 를 메서드 값으로 꺼내 두면
// (`send := fresh.dispatch.dispatch`) 호출식이 아니라서 세지 않았고, 그러면
// 두 번째 살아 있는 dispatch 가 생기는데도 sites=1 이라고 보고했다.
// (2) 건네는 값이 `handoff.result` 인지만 봤으므로, `handoff` 라는 이름의
// 구조체 리터럴에 `result` 필드를 달면 경계를 거치지 않은 값이 통과했다.
//
// 지금은 건네는 값이 **같은 함수 안에서 dispatchHandoff().Single() 이 묶어
// 준 이름**이어야 한다. 그 서명은 흉내 낼 수 없다.
func TestExactlyOneProductionCallSiteTurnsAHandoffIntoADispatch(t *testing.T) {
	type site struct {
		file, enclosing string
		argument        ast.Expr
		fromSeam        map[string]bool
	}
	sites := make([]site, 0, 1)
	for _, path := range engineProductionFiles(t) {
		file := parseEngineFile(t, path)
		for _, decl := range file.Decls {
			function, ok := decl.(*ast.FuncDecl)
			if !ok || function.Body == nil {
				continue
			}
			// dispatchHandoff().Single() 이 첫 자리에 묶어 준 이름들.
			fromSeam := make(map[string]bool)
			ast.Inspect(function.Body, func(node ast.Node) bool {
				assign, ok := node.(*ast.AssignStmt)
				if !ok || len(assign.Rhs) != 1 || len(assign.Lhs) != 2 {
					return true
				}
				outer, ok := assign.Rhs[0].(*ast.CallExpr)
				if !ok {
					return true
				}
				selector, ok := outer.Fun.(*ast.SelectorExpr)
				if !ok || selector.Sel.Name != "Single" {
					return true
				}
				inner, ok := selector.X.(*ast.CallExpr)
				if !ok {
					return true
				}
				source, ok := inner.Fun.(*ast.SelectorExpr)
				if !ok || source.Sel.Name != "dispatchHandoff" {
					return true
				}
				if name, ok := assign.Lhs[0].(*ast.Ident); ok {
					fromSeam[name.Name] = true
				}
				return true
			})
			ast.Inspect(function.Body, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				selector, ok := call.Fun.(*ast.SelectorExpr)
				if !ok || selector.Sel.Name != "dispatch" || len(call.Args) != 2 {
					return true
				}
				sites = append(sites, site{file: filepath.Base(path), enclosing: function.Name.Name,
					argument: call.Args[1], fromSeam: fromSeam})
				return true
			})
		}
	}
	if len(sites) != 1 {
		names := make([]string, 0, len(sites))
		for _, value := range sites {
			names = append(names, value.file+":"+value.enclosing)
		}
		sort.Strings(names)
		t.Fatalf("production dispatch call sites=%d, want exactly one bounded handoff: %v", len(sites), names)
	}
	if sites[0].enclosing != dispatchCallSiteFunc {
		t.Fatalf("the dispatch call moved into %s, not %s", sites[0].enclosing, dispatchCallSiteFunc)
	}
	dispatched, ok := sites[0].argument.(*ast.Ident)
	if !ok {
		t.Fatalf("the dispatched value is %T, not a name the seam bound", sites[0].argument)
	}
	if !sites[0].fromSeam[dispatched.Name] {
		t.Fatalf("the dispatched value %q was not bound by dispatchHandoff().Single(), "+
			"so something other than the bounded handoff chose it", dispatched.Name)
	}
	// dispatch 를 메서드 값으로 꺼내 두면 위의 호출 세기를 통째로 우회한다.
	assertEveryDispatchMentionIsInTheCensus(t)
}

// TestNoProductionSiteDiscardsTheSeamsAdmissionAnswer 는 경계가 준 답을
// 버리는 유일한 철자를 막는다.
//
// Single 은 두 값을 돌려주므로, 답을 안 받는 방법은 `_` 로 버리는 것 하나뿐이다
// (한 값짜리 문맥에서 쓰면 컴파일이 안 된다). 그 하나를 여기서 막으면 나머지는
// 컴파일러가 막는다 — 받아 놓고 아무 데서도 안 쓰면 "declared and not used" 다.
//
// 즉 이 검사와 컴파일러가 함께, 예전 AST 가드가 못 하던 일을 한다. 그 가드는
// Admitted() 가 **불렸는지**만 봤고 그것이 관문인지는 보지 않아서, 호출만
// 남기고 관문을 지우면 통과했다.
func TestNoProductionSiteDiscardsTheSeamsAdmissionAnswer(t *testing.T) {
	checked := 0
	for _, path := range engineProductionFiles(t) {
		for _, decl := range parseEngineFile(t, path).Decls {
			function, ok := decl.(*ast.FuncDecl)
			if !ok || function.Body == nil {
				continue
			}
			for _, answer := range seamAdmissionAnswers(t, filepath.Base(path), function) {
				checked++
				if !identAppearsInSomeIfCondition(function.Body, answer) {
					t.Errorf("%s: %s binds the seam's admission answer %q but never tests it in a condition",
						filepath.Base(path), function.Name.Name, answer)
				}
				if identIsBlankAssigned(function.Body, answer) {
					t.Errorf("%s: %s silences the seam's admission answer with `_ = %s`",
						filepath.Base(path), function.Name.Name, answer)
				}
			}
		}
	}
	if checked == 0 {
		t.Fatal("no production site reads the seam, so this guard proves nothing")
	}
	// 경계를 읽는 자리는 셋이다: worker 승격, dispatch 주기, 결과 권한,
	// 그리고 읽기 전용 projection — 넷이다. 센 값을 적는다.
	if checked != 4 {
		t.Fatalf("production sites reading the seam=%d, want 4 (worker promotion, dispatch cycle, result authority, projection)", checked)
	}
}

// seamAdmissionAnswers 는 이 함수 안에서 dispatchHandoff().Single() 이 묶어 준
// bool 이름들을 돌려준다. `_` 로 버렸으면 그 자리에서 바로 실패시킨다.
func seamAdmissionAnswers(t *testing.T, file string, function *ast.FuncDecl) []string {
	t.Helper()
	names := make([]string, 0, 1)
	ast.Inspect(function.Body, func(node ast.Node) bool {
		assign, ok := node.(*ast.AssignStmt)
		if !ok || len(assign.Rhs) != 1 {
			return true
		}
		outer, ok := assign.Rhs[0].(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := outer.Fun.(*ast.SelectorExpr)
		if !ok || selector.Sel.Name != "Single" {
			return true
		}
		inner, ok := selector.X.(*ast.CallExpr)
		if !ok {
			return true
		}
		if source, ok := inner.Fun.(*ast.SelectorExpr); !ok || source.Sel.Name != "dispatchHandoff" {
			return true
		}
		if len(assign.Lhs) != 2 {
			t.Errorf("%s: %s binds a dispatchHandoff().Single() to %d name(s), want the value and its admission answer",
				file, function.Name.Name, len(assign.Lhs))
			return true
		}
		name, ok := assign.Lhs[1].(*ast.Ident)
		if !ok || name.Name == "_" {
			t.Errorf("%s: %s throws away the seam's admission answer with `_`", file, function.Name.Name)
			return true
		}
		names = append(names, name.Name)
		return true
	})
	return names
}

// identAppearsInSomeIfCondition 은 그 이름이 어떤 if 의 조건 안에 나오는지 본다.
//
// 이것은 지배 관계의 **근사**다. 조건 안에 있다고 해서 그 조건이 값을 쓰기
// 전에 반드시 돌아간다는 뜻은 아니다. 진짜 지배를 증명하려면 go/types 와
// 제어 흐름 그래프가 필요하고, 그건 이 검사의 범위 밖이다. 여기서 막는 것은
// **실제로 일어난 실패 방식**이다: 관문을 지우고 답만 어딘가에 남겨 두는 편집.
func identAppearsInSomeIfCondition(body *ast.BlockStmt, name string) bool {
	found := false
	ast.Inspect(body, func(node ast.Node) bool {
		branch, ok := node.(*ast.IfStmt)
		if !ok || branch.Cond == nil {
			return true
		}
		ast.Inspect(branch.Cond, func(inner ast.Node) bool {
			if ident, ok := inner.(*ast.Ident); ok && ident.Name == name {
				found = true
			}
			return !found
		})
		return !found
	})
	return found
}

// identIsBlankAssigned 는 `_ = name` 을 찾는다. 컴파일러의 미사용 변수 검사를
// 달래려고 관문을 지운 자리에 남기는 바로 그 철자다.
func identIsBlankAssigned(body *ast.BlockStmt, name string) bool {
	blanked := false
	ast.Inspect(body, func(node ast.Node) bool {
		assign, ok := node.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for index, left := range assign.Lhs {
			target, ok := left.(*ast.Ident)
			if !ok || target.Name != "_" || index >= len(assign.Rhs) {
				continue
			}
			if value, ok := assign.Rhs[index].(*ast.Ident); ok && value.Name == name {
				blanked = true
			}
		}
		return !blanked
	})
	return blanked
}

// dispatchMentionCensus 는 생산 코드가 `dispatch` 를 손에 쥐는 자리 **전부**와
// 그 수다. 철자까지 고정한다.
//
// 왜 세는가: 호출식만 세면 `send := fresh.dispatch.dispatch` 처럼 부르지 않고
// 쥐고만 있는 형태를 놓친다. 그러면 살아 있는 dispatch 가 둘인데도 "부르는
// 자리는 하나"가 참으로 보고된다. 이름을 못 붙이게 하는 대신 **몇 번 어떻게
// 적혀 있는지**를 고정하면, 새로 쥐는 자리가 생기는 순간 이 표와 어긋난다.
//
// 값은 센 것이다. 생산 코드에서 `.dispatch` 는 두 줄뿐이고, 그중 한 줄이
// 수신자와 호출로 두 번 나타난다.
var dispatchMentionCensus = map[string]int{
	"strategy_entry_supervisor.go: fresh.dispatch":          2,
	"strategy_entry_supervisor.go: fresh.dispatch.dispatch": 1,
}

// assertEveryDispatchMentionIsInTheCensus 는 위 표를 소스와 맞춰 본다.
func assertEveryDispatchMentionIsInTheCensus(t *testing.T) {
	t.Helper()
	found := make(map[string]int)
	for _, path := range engineProductionFiles(t) {
		ast.Inspect(parseEngineFile(t, path), func(node ast.Node) bool {
			selector, ok := node.(*ast.SelectorExpr)
			if !ok || selector.Sel.Name != "dispatch" {
				return true
			}
			found[filepath.Base(path)+": "+types.ExprString(selector)]++
			return true
		})
	}
	if len(found) == 0 {
		t.Fatal("no production mention of dispatch was found, so this census fixes nothing")
	}
	mismatched := make([]string, 0)
	for spelling, count := range found {
		if dispatchMentionCensus[spelling] != count {
			mismatched = append(mismatched, spelling+" appears "+strconv.Itoa(count)+
				" time(s), census says "+strconv.Itoa(dispatchMentionCensus[spelling]))
		}
	}
	for spelling, count := range dispatchMentionCensus {
		if found[spelling] != count {
			mismatched = append(mismatched, spelling+" is in the census "+strconv.Itoa(count)+
				" time(s) but the source has "+strconv.Itoa(found[spelling]))
		}
	}
	if len(mismatched) != 0 {
		sort.Strings(mismatched)
		t.Fatalf("a production site took hold of dispatch in a way the census does not name: %v", mismatched)
	}
}
