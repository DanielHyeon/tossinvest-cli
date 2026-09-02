package strategyworker

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

// enginePath 는 오늘 생산에서 이 정책의 값을 들고 있는 곳이다.
//
// 왜 import 하지 않고 소스를 읽는가. `internal/app/engine` 은 journal·execgw·
// official 을 끌고 오므로, 시험이 그것을 import 하면 dependency_closure_test 의
// `-deps-test` 걸음이 깨진다 — 그 걸음이 지키는 것은 "이 패키지의 어떤 자동
// 시험도 주문 경로에 닿지 않는다"이다. 그래서 값은 파서로 읽는다.
//
// 디렉터리 전체를 읽는 이유: 한 파일만 읽으면 "완전하다"는 주장이 그 파일을
// 고른 사람의 선택이 된다. 다른 엔진 파일에 두 번째 대입이 생기면 이 시험이
// 그것을 본다.
const enginePath = "../app/engine"

func engineFiles(t *testing.T) map[string]*ast.File {
	t.Helper()
	fset := token.NewFileSet()
	packages, err := parser.ParseDir(fset, enginePath, productionOnly, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", enginePath, err)
	}
	files := make(map[string]*ast.File)
	for name, pkg := range packages {
		if name != "engine" {
			continue
		}
		for path, file := range pkg.Files {
			files[path] = file
		}
	}
	if len(files) == 0 {
		t.Fatalf("no production file was parsed from %s, so every reading below would be empty", enginePath)
	}
	return files
}

// engineConstants 는 엔진의 이름 있는 상수를 그 선언식 그대로 모은다.
func engineConstants(t *testing.T) map[string]ast.Expr {
	t.Helper()
	found := make(map[string]ast.Expr)
	for _, file := range engineFiles(t) {
		for _, decl := range file.Decls {
			general, ok := decl.(*ast.GenDecl)
			if !ok || general.Tok != token.CONST {
				continue
			}
			for _, spec := range general.Specs {
				value, ok := spec.(*ast.ValueSpec)
				if !ok || len(value.Names) != len(value.Values) {
					continue
				}
				for index, name := range value.Names {
					found[name.Name] = value.Values[index]
				}
			}
		}
	}
	return found
}

// durationOf 는 `5 * time.Second` 같은 선언식을 값으로 평가한다.
//
// 문자열 비교로 두지 않는 이유: `5 * time.Second` 와 `5000 * time.Millisecond`
// 는 같은 값인데 문자열은 다르고, 그러면 이 시험은 값이 갈라지지 않았는데도
// 빨개진다. 반대로 모르는 모양을 조용히 0 으로 돌려주면 값이 갈라져도 초록이
// 되므로, 모르는 모양은 실패다.
func durationOf(t *testing.T, expr ast.Expr) time.Duration {
	t.Helper()
	units := map[string]time.Duration{
		"Nanosecond": time.Nanosecond, "Microsecond": time.Microsecond, "Millisecond": time.Millisecond,
		"Second": time.Second, "Minute": time.Minute, "Hour": time.Hour,
	}
	unitOf := func(node ast.Expr) (time.Duration, bool) {
		selector, ok := node.(*ast.SelectorExpr)
		if !ok {
			return 0, false
		}
		pkg, ok := selector.X.(*ast.Ident)
		if !ok || pkg.Name != "time" {
			return 0, false
		}
		unit, known := units[selector.Sel.Name]
		return unit, known
	}
	if unit, ok := unitOf(expr); ok {
		return unit
	}
	binary, ok := expr.(*ast.BinaryExpr)
	if !ok || binary.Op != token.MUL {
		t.Fatalf("cannot evaluate %s as a duration", types.ExprString(expr))
	}
	literal, ok := binary.X.(*ast.BasicLit)
	if !ok || literal.Kind != token.INT {
		t.Fatalf("cannot evaluate %s as a duration", types.ExprString(expr))
	}
	count, err := strconv.ParseInt(literal.Value, 0, 64)
	if err != nil {
		t.Fatalf("cannot evaluate %s as a duration: %v", types.ExprString(expr), err)
	}
	unit, ok := unitOf(binary.Y)
	if !ok {
		t.Fatalf("cannot evaluate %s as a duration", types.ExprString(expr))
	}
	return time.Duration(count) * unit
}

func intOf(t *testing.T, expr ast.Expr) int {
	t.Helper()
	literal, ok := expr.(*ast.BasicLit)
	if !ok || literal.Kind != token.INT {
		t.Fatalf("cannot evaluate %s as an integer", types.ExprString(expr))
	}
	value, err := strconv.Atoi(literal.Value)
	if err != nil {
		t.Fatalf("cannot evaluate %s as an integer: %v", types.ExprString(expr), err)
	}
	return value
}

func engineConstant(t *testing.T, constants map[string]ast.Expr, name string) ast.Expr {
	t.Helper()
	expr, ok := constants[name]
	if !ok {
		names := make([]string, 0, len(constants))
		for known := range constants {
			names = append(names, known)
		}
		sort.Strings(names)
		t.Fatalf("the engine no longer declares %s — the receipt this policy reads is gone; known: %v", name, names)
	}
	return expr
}

// 정책의 네 값은 고른 수가 아니라 오늘 생산이 쓰는 수다.
//
// 이 시험이 없으면 두 곳이 말없이 갈라진다. 갈라지는 것을 아무도 못 보는 이유는
// 이 패키지가 아직 dormant 라서다 — 배선하는 날(5.1.2) 에야 나타나고, 그때는
// 옛 값과 새 값 중 어느 쪽이 옳았는지 아무도 모른다.
//
// "크거나 같음" 같은 느슨한 비교를 쓰지 않는다. 느슨하게 두면 그 여분은 어떤
// 영수증도 뒷받침하지 않는 수가 되고, 이 시험은 다시 "고른 수"를 통과시킨다.
func TestThePolicyValuesAreReadFromTheEngineRuntimeItReplaces(t *testing.T) {
	constants := engineConstants(t)
	policy := ProductionRuntimePolicy()

	if want := durationOf(t, engineConstant(t, constants, "DefaultStrategyCycleLimit")); policy.Cadence() != want {
		t.Errorf("cadence=%v, engine DefaultStrategyCycleLimit=%v", policy.Cadence(), want)
	}
	if want := durationOf(t, engineConstant(t, constants, "MaximumStrategyCycleLimit")); policy.CycleDeadline() != want {
		t.Errorf("cycle deadline=%v, engine MaximumStrategyCycleLimit=%v", policy.CycleDeadline(), want)
	}
	if want := durationOf(t, engineConstant(t, constants, "DefaultStrategyRestartStep")); policy.BackoffStep() != want {
		t.Errorf("backoff step=%v, engine DefaultStrategyRestartStep=%v", policy.BackoffStep(), want)
	}
	if want := durationOf(t, engineConstant(t, constants, "MaximumStrategyRestartBackoff")); policy.BackoffCeiling() != want {
		t.Errorf("backoff ceiling=%v, engine MaximumStrategyRestartBackoff=%v", policy.BackoffCeiling(), want)
	}
	if want := intOf(t, engineConstant(t, constants, "DefaultStrategyQueueDepth")); policy.QueueDepth() != want {
		t.Errorf("queue depth=%d, engine DefaultStrategyQueueDepth=%d", policy.QueueDepth(), want)
	}
}

// cadence 의 영수증은 상수 이름이 아니라 **대입**이다.
//
// `DefaultStrategyCycleLimit` 라는 이름은 "주기의 기본값"이 아니라 "사이클 한
// 번의 기본 상한"이다. 그것이 이 정책의 cadence 인 이유는 오직 생산 worker
// 서술자가 `PollInterval` 에 그 상수를 넣기 때문이고, 그 대입이 바뀌면 위
// 시험은 여전히 초록이면서 cadence 만 틀린 값이 된다.
func TestTheCadenceReceiptIsEveryProductionPollIntervalAssignment(t *testing.T) {
	sites := make([]string, 0)
	for path, file := range engineFiles(t) {
		ast.Inspect(file, func(node ast.Node) bool {
			switch typed := node.(type) {
			case *ast.KeyValueExpr:
				if key, ok := typed.Key.(*ast.Ident); ok && key.Name == "PollInterval" {
					sites = append(sites, path+" -> "+types.ExprString(typed.Value))
				}
			case *ast.AssignStmt:
				for index, target := range typed.Lhs {
					selector, ok := target.(*ast.SelectorExpr)
					if !ok || selector.Sel.Name != "PollInterval" || index >= len(typed.Rhs) {
						continue
					}
					sites = append(sites, path+" -> "+types.ExprString(typed.Rhs[index]))
				}
			}
			return true
		})
	}
	if len(sites) == 0 {
		t.Fatal("no PollInterval assignment was found, so this test read nothing")
	}
	sort.Strings(sites)
	for _, site := range sites {
		if !strings.HasSuffix(site, " -> DefaultStrategyCycleLimit") {
			t.Errorf("a production worker takes its poll interval from something else: %s", site)
		}
	}
}

// queue depth 의 영수증은 **대입이 없다는 것**이다.
//
// 생산은 `QueueDepth` 를 한 번도 지정하지 않고 0 으로 남긴다. 0 은 생성자에서
// `DefaultStrategyQueueDepth` 로 채워지므로 오늘 생산 깊이는 그 기본값이다.
// 누군가 생산 자리에 값을 지정하면 그 순간 이 정책이 읽은 영수증이 무효가
// 되므로, 지정 자체를 실패로 본다.
func TestNoProductionSiteChoosesItsOwnQueueDepth(t *testing.T) {
	scanned := 0
	for path, file := range engineFiles(t) {
		ast.Inspect(file, func(node ast.Node) bool {
			literal, ok := node.(*ast.CompositeLit)
			if !ok || types.ExprString(literal.Type) != "StrategyEntrySupervisorOptions" {
				return true
			}
			scanned++
			for _, element := range literal.Elts {
				pair, ok := element.(*ast.KeyValueExpr)
				if !ok {
					continue
				}
				key, ok := pair.Key.(*ast.Ident)
				if !ok {
					continue
				}
				if key.Name == "QueueDepth" {
					t.Errorf("%s: a production site chooses its own queue depth (%s)", path, types.ExprString(pair.Value))
				}
				if key.Name == "CycleLimit" && types.ExprString(pair.Value) != "MaximumStrategyCycleLimit" {
					t.Errorf("%s: a production site chooses its own cycle limit (%s)", path, types.ExprString(pair.Value))
				}
			}
			return true
		})
	}
	if scanned == 0 {
		t.Fatal("no StrategyEntrySupervisorOptions literal was found, so this test read nothing")
	}
}

// backoff 사다리의 모양도 엔진에서 읽는다.
//
// 값(계단 크기와 천장)이 같아도 계산이 다르면 다른 사다리가 된다. 행동 시험은
// 이 패키지가 만든 사다리만 보므로 두 구현이 갈라지는 것을 못 본다. 그래서
// 엔진 함수의 본문을 구조로 못 박는다 — 이것은 철자 census 가 아니라 그 함수의
// 문장 목록 자체다.
func TestTheBackoffLadderIsTheOneTheEngineComputes(t *testing.T) {
	var body *ast.BlockStmt
	for _, file := range engineFiles(t) {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if ok && fn.Name.Name == "strategyRestartBackoff" && fn.Recv == nil && fn.Body != nil {
				body = fn.Body
			}
		}
	}
	if body == nil {
		t.Fatal("the engine no longer declares strategyRestartBackoff, so this policy's ladder has no receipt")
	}
	got := make([]string, 0, len(body.List))
	for _, statement := range body.List {
		got = append(got, renderStatement(statement))
	}
	want := []string{
		"maximumSteps := uint64(MaximumStrategyRestartBackoff / DefaultStrategyRestartStep)",
		"if attempt >= maximumSteps { return MaximumStrategyRestartBackoff }",
		"return time.Duration(attempt) * DefaultStrategyRestartStep",
	}
	if strings.Join(got, " ; ") != strings.Join(want, " ; ") {
		t.Fatalf("the engine's backoff ladder changed.\n got: %v\nwant: %v\n\n"+
			"이 패키지의 Backoff 는 같은 사다리를 다시 계산한다. 엔진 쪽이 바뀌면 두 사다리가"+
			" 갈라지므로, 이 패키지도 함께 고치고 그 판단을 여기 want 에 적을 것.", got, want)
	}
}

// renderStatement 는 문장 하나를 한 줄로 편다. 구조 비교를 위한 것이지
// 소스 복원을 위한 것이 아니다.
func renderStatement(statement ast.Stmt) string {
	switch typed := statement.(type) {
	case *ast.AssignStmt:
		parts := make([]string, 0, len(typed.Lhs))
		for _, target := range typed.Lhs {
			parts = append(parts, types.ExprString(target))
		}
		values := make([]string, 0, len(typed.Rhs))
		for _, value := range typed.Rhs {
			values = append(values, types.ExprString(value))
		}
		return strings.Join(parts, ", ") + " " + typed.Tok.String() + " " + strings.Join(values, ", ")
	case *ast.ReturnStmt:
		values := make([]string, 0, len(typed.Results))
		for _, value := range typed.Results {
			values = append(values, types.ExprString(value))
		}
		return "return " + strings.Join(values, ", ")
	case *ast.IfStmt:
		inner := make([]string, 0, len(typed.Body.List))
		for _, nested := range typed.Body.List {
			inner = append(inner, renderStatement(nested))
		}
		rendered := "if " + types.ExprString(typed.Cond) + " { " + strings.Join(inner, "; ") + " }"
		if typed.Else != nil {
			rendered += " else {...}"
		}
		return rendered
	default:
		return "<unrenderable statement>"
	}
}

// 임계값 1 의 영수증은 **엔진에 실패 카운터가 없다는 것**이다.
//
// 오늘의 시장 worker 는 사이클이 오류를 내면 그 자리에서 `latchMarket` 을 부른다.
// 세다가 잠그는 것이 아니라 첫 실패에 잠근다. 그래서 이 정책의 임계값이 1 이면
// 오늘 동작과 같고, 1 보다 크면 **오늘이라면 잠겼을 레인이 한 번 더 진입을
// 시도한다** — 그것은 완화이고, 완화는 서명된 매니페스트가 정할 일이지 구현이
// 고를 일이 아니다.
//
// 그 사실을 필드 목록으로 못 박는다. 이름 하나가 있는지 묻는 검사는 뚫리므로
// (다른 철자의 카운터를 넣으면 그만이다) 구조체의 필드를 **전부** 세어 동결
// 목록과 견준다. 엔진에 상태가 하나라도 늘면 여기서 빨개지고, 그때 사람이
// "그 새 필드가 실패를 세는가"를 보게 된다.
func TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter(t *testing.T) {
	var fields []string
	for _, file := range engineFiles(t) {
		ast.Inspect(file, func(node ast.Node) bool {
			spec, ok := node.(*ast.TypeSpec)
			if !ok || spec.Name.Name != "strategyMarketRuntime" {
				return true
			}
			structure, ok := spec.Type.(*ast.StructType)
			if !ok {
				return true
			}
			fields = make([]string, 0)
			for _, field := range structure.Fields.List {
				for _, name := range field.Names {
					fields = append(fields, name.Name)
				}
			}
			return false
		})
	}
	if fields == nil {
		t.Fatal("the engine no longer declares strategyMarketRuntime, so the threshold below has no receipt")
	}
	want := []string{
		"descriptor", "queue", "effective", "latched", "firstFailure", "firstRefusal",
		"firstAbnormal", "latchID", "latchRevision", "abandoned", "restartAttempt", "restartNotBefore",
	}
	if strings.Join(fields, ",") != strings.Join(want, ",") {
		t.Fatalf("the engine's per-market runtime state changed.\n got: %v\nwant: %v\n\n"+
			"이 정책의 임계값 1 은 '엔진이 실패를 세지 않는다'에서 나온 값이다. 새 필드가"+
			" 실패를 센다면 임계값도 함께 정해야 하고, 안 센다면 위 목록만 갱신하면 된다.",
			fields, want)
	}
	if threshold := ProductionRuntimePolicy().FailureThreshold(); threshold != 1 {
		t.Fatalf("the production failure threshold is %d — the engine latches on the first failure, so"+
			" anything above one lets a lane try again where today it would already be latched", threshold)
	}
}
