package strategyworker

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"strconv"
	"strings"
	"testing"
)

// 이 파일은 유계 사이클의 **계약**을 엔진 소스에서 읽어 대조한다.
//
// 왜 필요한가. `invokeBounded` 는 엔진 `invokeBoundedStrategyCycle` 을 옮겨
// 적은 것이다. 부르지 않고 옮겨 적은 이유는 폐포 때문이고(엔진을 들여오면 원장과
// 게이트웨이가 따라온다), 옮겨 적은 것은 갈라진다. 갈라지는 것을 사람이 눈으로
// 막을 수는 없으므로 두 벌을 여기서 견준다.

// 투입 결과의 세 철자는 엔진의 것과 같아야 한다.
//
// 부분집합만 보면 안 된다. 엔진에 다섯 번째 결과가 생겼는데 여기서 아무 말도
// 안 하면, 레인은 엔진이 이미 구분하는 상태 하나를 조용히 뭉개게 된다. 그래서
// **차집합까지** 못 박는다 — 오늘 쓰지 않는 것은 정확히 `INVALID_MARKET` 하나다.
func TestTheTriggerNamesAreTheEnginesOwnSpelling(t *testing.T) {
	engine := make(map[string]string)
	for _, file := range engineFiles(t) {
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.CONST {
				continue
			}
			for _, spec := range gen.Specs {
				value, ok := spec.(*ast.ValueSpec)
				if !ok || value.Type == nil || types.ExprString(value.Type) != "StrategyTriggerResult" {
					continue
				}
				for index, name := range value.Names {
					if index >= len(value.Values) {
						continue
					}
					literal, ok := value.Values[index].(*ast.BasicLit)
					if !ok || literal.Kind != token.STRING {
						t.Fatalf("%s is not a string literal, so its spelling cannot be read", name.Name)
					}
					unquoted, err := strconv.Unquote(literal.Value)
					if err != nil {
						t.Fatalf("unquote %s: %v", name.Name, err)
					}
					engine[name.Name] = unquoted
				}
			}
		}
	}
	if len(engine) == 0 {
		t.Fatal("the engine declares no StrategyTriggerResult constant, so this comparison read nothing")
	}

	mine := map[string]Trigger{
		"StrategyTriggerEnqueued": TriggerEnqueued,
		"StrategyTriggerDisabled": TriggerDisabled,
		"StrategyTriggerFull":     TriggerFull,
	}
	for name, want := range mine {
		spelling, declared := engine[name]
		if !declared {
			t.Errorf("the engine no longer declares %s, so %q has no receipt", name, want)
			continue
		}
		if spelling != string(want) {
			t.Errorf("%s is %q in the engine and %q here", name, spelling, want)
		}
	}
	unused := make([]string, 0)
	for name := range engine {
		if _, mirrored := mine[name]; !mirrored {
			unused = append(unused, name)
		}
	}
	if strings.Join(unused, ",") != "StrategyTriggerInvalid" {
		t.Fatalf("the engine's trigger results this lane does not mirror are %v, want exactly"+
			" [StrategyTriggerInvalid].\n\n레인은 시장을 고르지 않으므로 INVALID_MARKET 이 생길"+
			" 자리가 없다. 그 밖의 결과가 엔진에 생겼다면 레인이 그것을 뭉개고 있다는 뜻이므로,"+
			" 짝을 만들거나 왜 안 만드는지 여기 적을 것.", unused)
	}
}

// 감시견의 세 갈래는 엔진의 것과 같은 순서·같은 결론이어야 한다.
//
// 특히 세 번째 갈래의 `true, false, true`(abnormal, cancelled, abandoned)가
// 중요하다. 설계 고장표는 deadline 을 "보통 오류" 줄에 두었으므로 두 정본이
// 갈리는 자리이고, 이 패키지는 **엔진 쪽**을 따랐다. 엔진이 그 결론을 바꾸면
// 여기서 빨개지고, 그때 사람이 어느 정본을 따를지 다시 고르게 된다.
func TestTheEngineWatchdogStillDecidesWhatThisLaneCopied(t *testing.T) {
	got := selectArms(t, engineWatchdogBody(t))
	want := []string{
		"case <-ctx.Done(): return ctx.Err(), false, true, true",
		"case outcome := <-result: return outcome.err, outcome.abnormal, false, false",
		"case <-deadline: if ctx.Err() != nil { return ctx.Err(), false, true, true }; " +
			"return ErrStrategyCycleDeadline, true, false, true",
	}
	if strings.Join(got, " ;; ") != strings.Join(want, " ;; ") {
		t.Fatalf("the engine's bounded cycle changed.\n got: %v\nwant: %v\n\n"+
			"이 패키지의 invokeBounded 는 이것을 옮겨 적은 것이다. 엔진이 바뀌면 두 벌이"+
			" 갈라지므로 이 패키지도 함께 고치고 그 판단을 여기 want 에 적을 것.", got, want)
	}
}

// 옮겨 적은 쪽도 못 박는다.
//
// 엔진만 고정하면 이 패키지의 select 가 조용히 바뀌어도 초록이다. 행동 시험이
// 세 갈래를 각각 재기는 하지만, 갈래의 **순서**는 행동으로 관측되지 않는다 —
// 세 채널이 동시에 준비되면 Go 는 무작위로 고르므로, 순서가 뒤바뀌어도 어떤
// 단일 시험도 반드시 빨개지지는 않는다.
func TestThisLaneWatchdogHasTheSameThreeArmsInTheSameOrder(t *testing.T) {
	got := selectArms(t, laneWatchdogBody(t))
	want := []string{
		"case <-ctx.Done(): return BoundedCycle{Err: ctx.Err(), Cancelled: true, Abandoned: true}",
		"case outcome := <-result: return outcome",
		"case <-deadline: if ctx.Err() != nil { return BoundedCycle{Err: ctx.Err(), Cancelled: true, Abandoned: true} }; " +
			"return BoundedCycle{Err: ErrCycleDeadline, Abnormal: true, Abandoned: true}",
	}
	if strings.Join(got, " ;; ") != strings.Join(want, " ;; ") {
		t.Fatalf("this lane's bounded cycle changed.\n got: %v\nwant: %v", got, want)
	}
}

func engineWatchdogBody(t *testing.T) *ast.BlockStmt {
	t.Helper()
	for _, file := range engineFiles(t) {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if ok && fn.Name.Name == "invokeBoundedStrategyCycle" && fn.Recv == nil && fn.Body != nil {
				return fn.Body
			}
		}
	}
	t.Fatal("the engine no longer declares invokeBoundedStrategyCycle, so this lane's watchdog has no receipt")
	return nil
}

func laneWatchdogBody(t *testing.T) *ast.BlockStmt {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "bounded.go", nil, 0)
	if err != nil {
		t.Fatalf("parse bounded.go: %v", err)
	}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if ok && fn.Name.Name == "invokeBounded" && fn.Recv != nil && fn.Body != nil {
			return fn.Body
		}
	}
	t.Fatal("bounded.go no longer declares invokeBounded")
	return nil
}

// selectArms 는 본문 안의 select 하나를 찾아 갈래마다 한 줄로 편다.
func selectArms(t *testing.T, body *ast.BlockStmt) []string {
	t.Helper()
	var found *ast.SelectStmt
	for _, statement := range body.List {
		if typed, ok := statement.(*ast.SelectStmt); ok {
			if found != nil {
				t.Fatal("this body has more than one select, so the comparison below is ambiguous")
			}
			found = typed
		}
	}
	if found == nil {
		t.Fatal("this body has no select, so it has no watchdog")
	}
	arms := make([]string, 0, len(found.Body.List))
	for _, clause := range found.Body.List {
		comm, ok := clause.(*ast.CommClause)
		if !ok {
			t.Fatal("a select body held something that is not a case")
		}
		inner := make([]string, 0, len(comm.Body))
		for _, statement := range comm.Body {
			inner = append(inner, renderStatement(statement))
		}
		head := "default:"
		if comm.Comm != nil {
			head = "case " + renderStatement(comm.Comm) + ":"
		}
		arms = append(arms, head+" "+strings.Join(inner, "; "))
	}
	return arms
}
