package engine

import (
	"go/ast"
	"go/parser"
	"go/token"
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
	// 유일하게 허용된 자리는 seam 의 색인 하나뿐이다. seam 의 개수 비교는
	// 리터럴 1 이 아니라 strategyMarketHandoffCapacity 라서 여기 잡히지 않는다 —
	// 그것이 요점이다. 상한이 이름을 가지면 한 곳에서 바꿀 수 있다.
	handoffSeamFile: 1,
	// L6 소유(태스크 6.2 q_final/owner admission). 그 로트가 열릴 때까지 남는다.
	// 5 는 센 값이다: len 비교 두 개와 색인 세 개.
	"strategy_account_first_leg_authority.go": 5,
}

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

// TestNoProductionSiteReadsAHandoffWithoutAskingWhetherItWasAdmitted 는
// seam 의 값을 거절 여부를 보지 않고 쓰는 것을 막는다.
//
// 이 검사는 뮤테이션이 요구해서 생겼다. worker 승격 조건에서 `!handoff.Admitted()`
// 를 지워도 그때 있던 시험은 전부 통과했다 — 거절된 handoff 의 result 가
// 영값이라 바로 아래 `ValidProposal()` 이 대신 걸러 주기 때문이다. 즉 그
// 자리의 안전은 handoff 가 아니라 **우연**이 지키고 있었다. 거절 때도 값을
// 채우도록 누가 바꾸면 그 우연은 사라지고, 어떤 동작 시험도 그것을 잡지 못한다.
// 그래서 여기서는 동작이 아니라 **쓰여 있는 것**을 본다.
func TestNoProductionSiteReadsAHandoffWithoutAskingWhetherItWasAdmitted(t *testing.T) {
	checked := 0
	for _, path := range engineProductionFiles(t) {
		for _, decl := range parseEngineFile(t, path).Decls {
			function, ok := decl.(*ast.FuncDecl)
			if !ok || function.Body == nil {
				continue
			}
			bound := make(map[string]bool)
			asked := make(map[string]bool)
			ast.Inspect(function.Body, func(node ast.Node) bool {
				if assign, ok := node.(*ast.AssignStmt); ok {
					for index, value := range assign.Rhs {
						call, isCall := value.(*ast.CallExpr)
						if !isCall || index >= len(assign.Lhs) {
							continue
						}
						selector, isSelector := call.Fun.(*ast.SelectorExpr)
						if !isSelector || selector.Sel.Name != "dispatchHandoff" {
							continue
						}
						if name, isIdent := assign.Lhs[index].(*ast.Ident); isIdent {
							bound[name.Name] = true
						}
					}
				}
				call, isCall := node.(*ast.CallExpr)
				if !isCall {
					return true
				}
				selector, isSelector := call.Fun.(*ast.SelectorExpr)
				if !isSelector || selector.Sel.Name != "Admitted" {
					return true
				}
				if receiver, isIdent := selector.X.(*ast.Ident); isIdent {
					asked[receiver.Name] = true
				}
				return true
			})
			for name := range bound {
				checked++
				if !asked[name] {
					t.Errorf("%s: %s reads the handoff %q without calling %s.Admitted()",
						filepath.Base(path), function.Name.Name, name, name)
				}
			}
		}
	}
	if checked == 0 {
		t.Fatal("no production function binds a dispatchHandoff, so this guard proves nothing")
	}
}

// mutationCapabilityImports 는 이 패키지 안에서 브로커나 원장을 바꿀 수 있는
// 유일한 통로다. dispatch cycle 이 실제로 무엇을 들여오는지 읽어서 적었다.
var mutationCapabilityImports = []string{
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw",
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal",
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent",
}

// mutationCapabilityIdents 는 같은 패키지 안에 있어서 import 로는 막을 수 없는
// 능력들이다. 이 이름 중 하나라도 들고 있으면 그 파일은 주문이나 lease 를
// 만들 수 있다.
var mutationCapabilityIdents = []string{
	"strategyDispatchCycle",
	"newStrategyDispatchCycle",
	"strategyDispatchGateway",
	"strategyDispatchOwnerCoordinator",
	"strategyFirstLegAdmissionBridge",
}

func TestTheCoordinatorAndHandoffSeamsHoldNoMutationCapability(t *testing.T) {
	for _, name := range []string{coordinatorSeamFile, handoffSeamFile} {
		path := filepath.Join(".", name)
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("guarded seam %s is missing: %v", name, err)
		}
		file := parseEngineFile(t, path)
		for _, spec := range file.Imports {
			value, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				t.Fatalf("%s: unquote %s: %v", name, spec.Path.Value, err)
			}
			for _, forbidden := range mutationCapabilityImports {
				if value == forbidden {
					t.Errorf("%s imports %s, so the seam can reach a mutation capability", name, value)
				}
			}
		}
		ast.Inspect(file, func(node ast.Node) bool {
			ident, ok := node.(*ast.Ident)
			if !ok {
				return true
			}
			for _, forbidden := range mutationCapabilityIdents {
				if ident.Name == forbidden {
					t.Errorf("%s names %s, so the seam holds a mutation capability", name, forbidden)
				}
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
func TestExactlyOneProductionCallSiteTurnsAHandoffIntoADispatch(t *testing.T) {
	type site struct {
		file, enclosing string
		argument        ast.Expr
	}
	sites := make([]site, 0, 1)
	for _, path := range engineProductionFiles(t) {
		file := parseEngineFile(t, path)
		for _, decl := range file.Decls {
			function, ok := decl.(*ast.FuncDecl)
			if !ok || function.Body == nil {
				continue
			}
			ast.Inspect(function.Body, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				selector, ok := call.Fun.(*ast.SelectorExpr)
				if !ok || selector.Sel.Name != "dispatch" || len(call.Args) != 2 {
					return true
				}
				sites = append(sites, site{file: filepath.Base(path), enclosing: function.Name.Name, argument: call.Args[1]})
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
	selector, ok := sites[0].argument.(*ast.SelectorExpr)
	if !ok {
		t.Fatalf("the dispatched value is %T, not a field of the bounded handoff", sites[0].argument)
	}
	receiver, isIdent := selector.X.(*ast.Ident)
	if !isIdent || receiver.Name != "handoff" || selector.Sel.Name != "result" {
		t.Fatalf("the dispatched value is not handoff.result, so something other than the bounded handoff chose it")
	}
}
