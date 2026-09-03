package engine

import (
	"context"
	"go/ast"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyworker"
)

// 이 파일은 태스크 5.1.2 가 옮긴 경계를 값과 구조로 확인한다.
//
// 옮긴 것 하나: 전략군 하나를 평가하는 일이 더 이상 `*Context` 클로저가 아니다.
// 그 문장은 실행으로 증명할 수 없다 — 오늘 여덟 레인은 전부 DORMANT 라
// 어떤 실행도 "무엇을 만질 수 있었는가"를 보여 주지 못한다. 그래서 절반은
// 값으로(레인이 실제로 서고 돌고 잠긴다) 절반은 소스 구조로(레인이 도는 일에
// 원장·게이트웨이가 적혀 있지 않다) 확인한다.

func laneRuntimeFixture(t *testing.T) (*strategyLaneRuntime, *clock.Fake) {
	t.Helper()
	fake := clock.NewFake(time.Date(2026, 9, 3, 1, 0, 0, 0, time.UTC))
	runtime := newStrategyLaneRuntime(fake, nil, "")
	if runtime == nil {
		t.Fatal("생산 레인 런타임이 서지 않았다")
	}
	return runtime, fake
}

// TestTheProductionRuntimeStandsExactlyEightLanesFourPerMarket 은 스펙의
// "정확한 worker cardinality" 시나리오다: 8개의 unique lane instance,
// duplicate 또는 unknown 0건.
func TestTheProductionRuntimeStandsExactlyEightLanesFourPerMarket(t *testing.T) {
	runtime, _ := laneRuntimeFixture(t)
	if got, want := len(runtime.lanes), len(strategyworker.ProductionWorkers()); got != want {
		t.Fatalf("레인 수=%d, want %d — 목록은 strategyworker 의 생산 진입점 하나가 정한다", got, want)
	}
	seen := make(map[strategyworker.Key]int, len(runtime.lanes))
	for _, lane := range runtime.lanes {
		seen[lane.Key()]++
	}
	if len(seen) != len(runtime.lanes) {
		t.Fatalf("중복된 레인 열쇠가 있다: %v", seen)
	}
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		lanes := runtime.lanesFor(market)
		if len(lanes) != 4 {
			t.Fatalf("%s 레인 수=%d, want 4 — 네 전략군", market, len(lanes))
		}
		families := make(map[strategyrouter.Family]bool, 4)
		for _, lane := range lanes {
			if lane.Key().Market != strategyRouterMarket(market) {
				t.Fatalf("%s 목록에 %s 레인이 섞였다", market, lane.Key().Market)
			}
			families[lane.Key().Family] = true
		}
		if len(families) != 4 {
			t.Fatalf("%s 가족 수=%d, want 4 — 한 가족이 두 번 서거나 하나가 빠졌다", market, len(families))
		}
	}
}

// TestTheLaneSetOutlivesTheRefreshThatAskedForIt 는 잠금이 기억이라는 사실을 지킨다.
//
// 새로 고침마다 레인을 다시 만들면 잠긴 레인이 1 초 뒤 열린 채로 돌아온다.
// 잠근 이유는 그대로인데. 그래서 같은 Context 는 언제나 같은 레인을 준다.
func TestTheLaneSetOutlivesTheRefreshThatAskedForIt(t *testing.T) {
	c := &Context{}
	fake := clock.NewFake(time.Date(2026, 9, 3, 1, 0, 0, 0, time.UTC))
	first := mustProductionStrategyLanes(t, c, fake)
	if first == nil {
		t.Fatal("첫 호출이 레인을 세우지 않았다")
	}
	// 두 번째 호출이 다른 시계를 들고 와도 레인은 그대로여야 한다. 시계가
	// 바뀐다고 잠금이 풀리면 시계를 갈아 끼우는 것이 복구 수단이 된다.
	second := mustProductionStrategyLanes(t, c, clock.NewFake(time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)))
	if first != second {
		t.Fatal("새로 고침마다 레인을 다시 만든다 — 잠긴 레인이 열린 채로 돌아온다")
	}
	for index := range first.lanes {
		if first.lanes[index] != second.lanes[index] {
			t.Fatalf("레인 %d 이 교체됐다", index)
		}
	}
}

// TestEveryFamilyLaneIsDormantUntilASignedManifestPromotesIt 는 동결 골든이
// 여덟을 전부 `effective: OFF` 로 얼린 것과 스펙의 "Legacy 3-family approval 은
// 4-family activation 으로 자동 승격되어서는 안 된다 (MUST NOT)" 를 값으로 잡는다.
//
// DORMANT 는 거절이 아니다. 그 둘을 하나로 뭉치면 아직 안 켠 레인과 고장으로
// 잠긴 레인이 같은 값이 되고, 운영자가 할 조치가 갈린다.
func TestEveryFamilyLaneIsDormantUntilASignedManifestPromotesIt(t *testing.T) {
	runtime, _ := laneRuntimeFixture(t)
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		runtime.evaluate(context.Background(), market, 0, strategyrouter.FamilyActivation{}, nil)
	}
	observations := runtime.observations()
	if len(observations) != len(runtime.lanes) {
		t.Fatalf("관측 수=%d, want %d — 돌지 않은 레인도 목록에 남아야 한다",
			len(observations), len(runtime.lanes))
	}
	for _, observation := range observations {
		if observation.Trigger != strategyworker.TriggerEnqueued {
			t.Fatalf("%v: 투입 결과=%s, want ENQUEUED — 사이클이 아예 열리지 않았다",
				observation.Key, observation.Trigger)
		}
		if observation.Start != strategyworker.StartAdmitted {
			t.Fatalf("%v: 관문=%s, want ADMITTED", observation.Key, observation.Start)
		}
		if observation.Outcome != strategyworker.OutcomeDormant {
			t.Fatalf("%v: 결과=%s, want DORMANT.\n\n"+
				"활성화 매니페스트 없이 레인이 켜졌다. 동결 골든 four-family-runtime-v1.json 은"+
				" 여덟 서술자를 전부 effective:OFF 로 얼렸고, 스펙은 legacy 3-family 승인을"+
				" 4-family activation 으로 승격하는 것을 MUST NOT 으로 금지한다.",
				observation.Key, observation.Outcome)
		}
		if observation.Emitted {
			t.Fatalf("%v: 잠든 레인이 봉투를 냈다 — 누군가 켜진 worker 를 주조했다", observation.Key)
		}
		if observation.Latched || observation.Health != strategyworker.LaneHealthy {
			t.Fatalf("%v: 아무 고장도 없는데 상태가 %s 다", observation.Key, observation.Health)
		}
	}
}

// TestADroppedTriggerNeverDrivesACycle 는 유계 큐가 실제로 무엇을 막는지를 본다.
//
// **첫 판본은 버린 수만 셌고, 그 수는 `Offer` 가 올린다.** 그래서 "칸이 찼으면
// 사이클을 열지 않는다"는 갈래를 지워도 시험이 초록이었다(뮤테이션 M10 이
// 살아남았다). 그때 실제로 문을 닫고 있던 것은 그 갈래가 아니라 **카덴스**였다 —
// 우연이 지키던 안전이다.
//
// 여기서는 카덴스를 지나가게 한 뒤에 칸을 채운다. 그러면 버려진 투입 하나가
// 그대로 사이클을 열 수 있는 상태가 되고, 갈래가 없으면 운영자는 "버렸다"고
// 보고받은 물결이 실제로는 평가된 것을 보게 된다.
func TestADroppedTriggerNeverDrivesACycle(t *testing.T) {
	runtime, fake := laneRuntimeFixture(t)
	lane := runtime.lanesFor(StrategyMarketKR)[0]
	cadence := lane.Policy().Cadence()

	// 1) 첫 주기는 들어가고 돈다. 다음 주기 시각이 정해진다.
	runtime.evaluate(context.Background(), StrategyMarketKR, 0, strategyrouter.FamilyActivation{}, nil)
	if got := lane.Dropped(); got != 0 {
		t.Fatalf("첫 주기에 버린 수=%d, want 0", got)
	}
	// 2) 카덴스가 아직이라 투입은 칸에 남는다.
	runtime.evaluate(context.Background(), StrategyMarketKR, 0, strategyrouter.FamilyActivation{}, nil)
	if got, want := lane.Pending(), lane.Policy().QueueDepth(); got != want {
		t.Fatalf("칸에 남은 투입=%d, want %d", got, want)
	}
	// 3) 이제 카덴스를 지나 보낸다. 칸은 여전히 차 있다.
	fake.Advance(cadence)
	runtime.evaluate(context.Background(), StrategyMarketKR, 0, strategyrouter.FamilyActivation{}, nil)

	if got := lane.Dropped(); got == 0 {
		t.Fatal("칸이 찼는데 버린 수가 0 이다 — 유실이 조용히 사라진다")
	}
	for _, observation := range runtime.observations() {
		if observation.Key != lane.Key() {
			continue
		}
		if observation.Trigger != strategyworker.TriggerFull {
			t.Fatalf("칸이 찬 주기의 투입 결과=%s, want FULL", observation.Trigger)
		}
		if observation.Start != "" {
			t.Fatalf("버린 투입이 사이클을 열었다: 관문=%s, 결과=%s.\n\n"+
				"버린 수는 `Offer` 가 올리므로 그 수만 보면 이 갈래가 지워져도 보이지 않는다."+
				" 버렸다고 보고한 물결이 그대로 평가되면, 유실 계수기가 세는 것과 실제로"+
				" 일어난 일이 갈린다.", observation.Start, observation.Outcome)
		}
	}
}

// TestOnlyThePackageLevelStepEverRunsInsideALane 은 이 태스크의 문장을
// 소스 전체에서 센다.
//
// **왜 행동으로는 안 되는가.** 오늘 여덟은 전부 DORMANT 라, 어떤 실행도 레인이
// 무엇을 만질 수 **있었는지**를 보여 주지 못한다. 그리고 5.5·5.1.1·5.6.1 이
// 세 번 같은 결론에 닿았다 — 축마다 시험 하나는 끝나지 않는다. 그래서 세는
// 범위를 함수 본문이 아니라 **패키지 전체**로 올린다: 이 패키지에서
// `RunBounded` 에 넘기는 값이 무엇인지를 전부 세어 얼린다.
//
// 목록에 한 줄이 늘면 이 시험이 실패한다. 그것이 요점이다 — 레인이 도는 일을
// 새로 여는 것은 "전략군 평가에 새 능력을 주는 것"이고, `*Context` 메서드나
// `*Context` 를 담은 클로저를 그 자리에 두면 원장과 주문 경로가 함께 들어온다.
func TestOnlyThePackageLevelStepEverRunsInsideALane(t *testing.T) {
	sites := make([]string, 0, 1)
	for _, path := range engineProductionFiles(t) {
		file := parseEngineFile(t, path)
		for _, decl := range file.Decls {
			ast.Inspect(decl, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				selector, isSelector := call.Fun.(*ast.SelectorExpr)
				if !isSelector || selector.Sel.Name != "RunBounded" {
					return true
				}
				if len(call.Args) != 3 {
					t.Fatalf("%s: RunBounded 인자=%d 개 — 시험이 읽는 자리가 사라졌다",
						filepath.Base(path), len(call.Args))
				}
				sites = append(sites, filepath.Base(path)+":"+declName(decl)+":"+
					exprSpelling(call.Args[2]))
				return true
			})
		}
	}
	sort.Strings(sites)
	// 두 번째 인자가 는 것은 태스크 8.7.1 이다. 그 값은
	// `strategyrouter.FamilyActivation` — 이 패키지 밖에서는 영값만 만들 수 있고
	// 영값은 아무것도 승격하지 않는 **값**이다. 능력이 아니라는 것은 아래
	// TestTheFamilyLaneStepCarriesNothingButItsLaneAndTheSignedPromotion 이
	// 두 인자의 **타입**을 못 박아 확인한다(개수가 아니라 타입이다).
	want := []string{"strategy_lane_runtime.go:runLane:strategyFamilyLaneStep(lane, promotion)"}
	if strings.Join(sites, "\n") != strings.Join(want, "\n") {
		t.Fatalf("레인이 도는 일의 목록이 바뀌었다.\n got: %v\nwant: %v\n\n"+
			"레인 안에서 도는 일은 `*Context` 를 들 수 없어야 한다. 이 목록에 줄이 늘면"+
			" 그 값이 무엇을 담고 있는지 먼저 적을 것 — `*Context` 는 Journal 과 Gateway 를"+
			" 들고 있고, 그것이 레인 안에 들어오면 이 태스크가 옮긴 경계가 사라진다.",
			sites, want)
	}
}

// exprSpelling 은 AST 식을 소스에 적힌 대로 되돌린다. 형만 보면 서로 다른
// 값이 같은 철자로 보이므로 인자까지 적는다.
func exprSpelling(expr ast.Expr) string {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return exprName(expr)
	}
	args := make([]string, 0, len(call.Args))
	for _, arg := range call.Args {
		args = append(args, exprName(arg))
	}
	return exprName(call.Fun) + "(" + strings.Join(args, ", ") + ")"
}

func exprName(expr ast.Expr) string {
	switch value := expr.(type) {
	case *ast.Ident:
		return value.Name
	case *ast.SelectorExpr:
		return exprName(value.X) + "." + value.Sel.Name
	case *ast.FuncLit:
		return "(익명 함수)"
	}
	return "(식)"
}

// TestTheFamilyLaneStepCarriesNothingButItsLaneAndTheSignedPromotion 은 그
// 값들이 실제로 무엇을 들 수 있는지를 본다.
//
// 수신자가 없어야 하는 이유: 메서드로 두면 수신자가 무엇이든 될 수 있고,
// `*Context` 를 수신자로 두는 순간 원장과 게이트웨이가 함께 들어온다. 오늘
// 생산의 시장 주기가 정확히 그 모양이다.
//
// **개수가 아니라 타입을 못 박는다** (태스크 8.7.1 로 인자가 둘이 되었다).
// 앞 판본은 "인자는 하나" 라고 셌다. 개수 검사는 정당한 인자가 늘 때 반드시
// 걸리고, 걸렸을 때 통과시키는 유일한 방법이 **숫자를 고치는 것**이라 감시
// 대상과 같은 손짓이 된다. 타입 목록을 얼리면 새 인자는 자기 타입을 여기에
// 적어야 하고, 그 타입이 능력이면 아래 계산이 잡는다.
func TestTheFamilyLaneStepCarriesNothingButItsLaneAndTheSignedPromotion(t *testing.T) {
	file := parseEngineFile(t, filepath.Join(".", "strategy_lane_runtime.go"))
	step := engineFuncDecl(t, file, "strategyFamilyLaneStep")
	if step.Recv != nil {
		t.Fatal("strategyFamilyLaneStep 이 메서드가 됐다 — 수신자가 능력을 실어 나를 수 있다")
	}
	// 얼린 목록. 첫째는 레인 — 그 타입이 사는 패키지는 자기 import 폐포에
	// 원장·게이트웨이·브로커 변경자가 없다는 것을 `-deps`/`-deps-test` 로 훑어
	// 지킨다(strategyworker/dependency_closure_test.go). 둘째는 서명된 승격 —
	// 필드가 전부 비공개라 이 패키지에서는 영값만 만들 수 있고, 영값은 아무것도
	// 승격하지 않는다. 두 값 다 능력이 아니라 값이다.
	wantTypes := []string{"*strategyworker.Lane", "strategyrouter.FamilyActivation"}
	gotTypes := make([]string, 0, len(wantTypes))
	for _, field := range step.Type.Params.List {
		if len(field.Names) == 0 {
			t.Fatal("이름 없는 매개변수가 있다 — 시험이 읽는 자리가 흐려진다")
		}
		for range field.Names {
			gotTypes = append(gotTypes, exprTypeString(field.Type))
		}
	}
	if strings.Join(gotTypes, ", ") != strings.Join(wantTypes, ", ") {
		t.Fatalf("인자 타입 목록=%v, want %v — 새 인자를 더하려면 그 타입을 여기 적을 것."+
			" 능력을 실어 나르는 타입이면 아래 계산이 잡는다", gotTypes, wantTypes)
	}
	capabilities := engineCapabilityTypes(t)
	for _, name := range mutationCapabilityIdents {
		capabilities[name] = true
	}
	// 한정된 이름(`context.Context`)은 통째로 본다. `Sel` 만 따로 세면 표준
	// 라이브러리의 Context 가 엔진의 `*Context` 와 같은 철자라 걸린다 — 그리고
	// 그 오탐을 없애려고 세는 범위를 좁히면 진짜 능력도 함께 안 보인다.
	ast.Inspect(step, func(node ast.Node) bool {
		switch value := node.(type) {
		case *ast.SelectorExpr:
			if capabilities[exprName(value)] {
				t.Errorf("strategyFamilyLaneStep 이 %s 를 이름으로 부른다 — 그것은 변경 능력이다",
					exprName(value))
			}
			// 한정된 이름의 뒷자리는 여기서 이미 봤다. 다시 내려가면 표준
			// 라이브러리 이름이 엔진 타입 이름으로 오인된다.
			ast.Inspect(value.X, inspectBareCapability(t, capabilities))
			return false
		case *ast.Ident:
			if capabilities[value.Name] {
				t.Errorf("strategyFamilyLaneStep 이 %s 를 이름으로 부른다 — 그것은 변경 능력이다",
					value.Name)
			}
		}
		return true
	})
}

// inspectBareCapability 는 한정된 이름의 앞자리에서 능력 이름을 찾는다.
func inspectBareCapability(t *testing.T, capabilities map[string]bool) func(ast.Node) bool {
	t.Helper()
	return func(node ast.Node) bool {
		if ident, ok := node.(*ast.Ident); ok && capabilities[ident.Name] {
			t.Errorf("strategyFamilyLaneStep 이 %s 를 이름으로 부른다 — 그것은 변경 능력이다", ident.Name)
		}
		return true
	}
}

func exprTypeString(expr ast.Expr) string {
	if star, ok := expr.(*ast.StarExpr); ok {
		return "*" + exprName(star.X)
	}
	return exprName(expr)
}

// TestTheMarketCycleRunsItsLanesAndTheRefreshDoesNot 는 레인 사이클이 공유
// 새로 고침 잠금 **밖**에서 돈다는 것을 자리로 확인한다.
//
// 왜 중요한가. 레인 사이클은 마감 시한 감시견을 달고 돈다. 새로 고침 잠금
// (`c.strategyRefreshMu`) 안에서 돌리면 레인 하나의 느린 주기가 두 시장의
// 모든 권한 수집을 함께 세운다 — 그리고 그 잠금을 기다리는 쪽에는 손절과
// 무관하지 않은 경로가 붙어 있다.
func TestTheMarketCycleRunsItsLanesAndTheRefreshDoesNot(t *testing.T) {
	callers := make([]string, 0, 1)
	for _, path := range engineProductionFiles(t) {
		file := parseEngineFile(t, path)
		for _, decl := range file.Decls {
			function, ok := decl.(*ast.FuncDecl)
			if !ok || function.Body == nil {
				continue
			}
			ast.Inspect(function.Body, func(node ast.Node) bool {
				call, isCall := node.(*ast.CallExpr)
				if !isCall {
					return true
				}
				selector, isSelector := call.Fun.(*ast.SelectorExpr)
				if !isSelector || selector.Sel.Name != "evaluate" {
					return true
				}
				callers = append(callers, filepath.Base(path)+":"+function.Name.Name)
				return true
			})
		}
	}
	sort.Strings(callers)
	want := []string{"strategy_entry_supervisor.go:runProductionStrategyMarketCycle"}
	if strings.Join(callers, "\n") != strings.Join(want, "\n") {
		t.Fatalf("레인 사이클을 여는 자리=%v, want %v — 새로 고침 잠금 안으로 들어가면"+
			" 레인 하나의 느린 주기가 두 시장의 권한 수집을 함께 세운다", callers, want)
	}
}

// TestTheLaneScopeReadsTheApprovedSymbolAndNotTheProposalsOwnLineage 는 범위의
// 종목이 **승인된 후보**에서 온다는 것을 소스로 못 박는다.
//
// 왜 검사가 아니라 구조인가. 이 함수는 아무것도 검사하지 않는다 — 범위를
// 만들 뿐이다. 그런데 그 범위는 아래(조정자 `Submit`)에서 계보와 대조된다.
// 범위의 종목을 계보에서 읽으면 그 대조가 `lineage.Symbol == lineage.Symbol`
// 이 되어 언제나 참이 되고, 어긋난 제안이 남의 소유자 범위로 들어간다.
// `coordinateMarketProposals` 가 같은 이유로 같은 선택을 하고 그 이유를
// 자기 주석에 적어 두었다.
func TestTheLaneScopeReadsTheApprovedSymbolAndNotTheProposalsOwnLineage(t *testing.T) {
	file := parseEngineFile(t, filepath.Join(".", "strategy_lane_runtime.go"))
	builder := engineFuncDecl(t, file, "strategyLaneInputs")
	spellings := make(map[string]string, 4)
	ast.Inspect(builder.Body, func(node ast.Node) bool {
		pair, ok := node.(*ast.KeyValueExpr)
		if !ok {
			return true
		}
		key, isIdent := pair.Key.(*ast.Ident)
		if !isIdent {
			return true
		}
		spellings[key.Name] = exprSpelling(pair.Value)
		return true
	})
	want := map[string]string{
		"Symbol":             "entry.route.approved.Symbol()",
		"PositionGeneration": "(식).Key.PositionGeneration",
		"SnapshotDigest":     "entry.authority.SnapshotDigest()",
	}
	for field, expected := range want {
		got, present := spellings[field]
		if !present {
			t.Fatalf("strategyLaneInputs 가 %s 를 더 이상 채우지 않는다 — 이 시험이 읽는 자리가 사라졌다", field)
		}
		if got != expected {
			t.Fatalf("%s = %s, want %s.\n\n"+
				"소유자 범위의 종목이 제안의 계보에서 오면, 아래 조정자가 그 둘을 대조하는"+
				" 자리(`lineage.Symbol != scope.Symbol`)가 언제나 참이 되어 어긋난 제안이"+
				" 남의 범위로 들어간다.", field, got, expected)
		}
	}
	if spellings["Market"] != "strategyRouterMarket(authority.market)" {
		t.Fatalf("Market = %s, want strategyRouterMarket(authority.market) — 시장을 계보에서 읽으면"+
			" 남의 시장 제안이 이 시장 범위로 들어간다", spellings["Market"])
	}
}
