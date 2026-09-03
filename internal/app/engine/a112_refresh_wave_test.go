package engine

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// 이 파일은 태스크 5.2 의 두 절 중 하나를 지킨다: **remote evidence refresh 를
// 공유 assembly mutex 밖에 둔다.**
//
// 왜 셈으로 지키는가. "이 잠금을 들고 원격을 부르지 않는다"는 문장은 한 함수의
// 성질이 아니라 **패키지의 성질**이다. 한 자리를 고쳐 놓아도 다음 사람이 다른
// 자리에서 같은 잠금을 들면 같은 결함이 돌아오고, 행동 시험은 그것을 못 본다 —
// 오늘 이 패키지에서 파도를 느리게 만드는 방법이 없기 때문이다(원격은 실제
// official client 뒤에 있고 시계로는 늦출 수 없다). 그래서 이 change 가 5.5·
// 5.1.1·5.6.1·5.1.2.1 에서 네 번 도달한 답을 다섯 번째로 쓴다: **세는 범위를
// 함수 본문에서 패키지 전체로 올린다.**
//
// 잠금을 만지는 함수 목록과 그 함수들이 부르는 것의 목록을 **둘 다** 얼린다.
// 하나만 얼리면 한 다리 건너 우회된다 — 잠금 안에서 헬퍼를 부르고 그 헬퍼가
// 원격을 부르면 첫 목록은 그대로다. 이것이 5.5 의 적대 리뷰가 실제로 뚫은
// 방법이고, 기억에 "존재 검사는 역할 검사가 아니다"로 남아 있다.

// strategyRefreshMuxName 은 이 셈의 대상이다. 문자열 하나로 두는 이유는 아래
// 두 시험이 **같은 이름**을 세게 하기 위해서다. 서로 다른 철자를 세면 한쪽만
// 통과하는 편집이 생긴다.
const strategyRefreshMutexName = "strategyRefreshMu"

// strategyRemoteWaveCall 은 원격 권한 수집 한 번의 **메서드 이름**이다.
//
// 이 함수가 원격을 부른다는 것은 추측이 아니라 읽은 것이다: 본문이
// `newStrategyScheduleAuthorityLoader(..., c.official, ...)` 와
// `newStrategyFXAuthorityLoader(..., c.official, ...)` 를 만들고 그 둘의
// collect 가 각각 official 의 `TypedMarketCalendar` 와
// `officialfx.ProductionAuthorityService` 를 탄다.
//
// 수신자 철자(`c.`)를 넣어 통째로 비교하지 않는다. 수신자 이름을 바꾸는 편집
// 하나로 이 셈이 조용히 아무것도 못 보게 되기 때문이다 — 이 change 가 세 번
// 겪은 "이름이 있는지 보는 검사는 뚫린다"와 같은 모양이다.
const strategyRemoteWaveCall = "NewPairedStrategyEntryProductionAssembly"

// TestTheRemoteAuthorityWaveNeverRunsUnderTheSharedAssemblyMutex 은 이 로트의
// RED 다. 고치기 전에는 `refreshPairedStrategyEntryProductionAssembly` 하나가
// 잠금을 들고 원격 파도를 돌리므로 빨갛다.
func TestTheRemoteAuthorityWaveNeverRunsUnderTheSharedAssemblyMutex(t *testing.T) {
	holders, waveCallers := strategyRefreshMutexCensus(t)

	// (1) 잠금을 만지는 함수와 원격 파도를 부르는 함수는 **서로소**여야 한다.
	for name := range waveCallers {
		if callees, held := holders[name]; held {
			t.Fatalf("%s 가 공유 assembly mutex 를 들고 원격 권한 파도를 돌린다.\n"+
				"그 함수의 호출 목록: %v\n\n"+
				"이 잠금은 두 시장의 새로 고침이 공유하는 하나다. 그것을 든 채 원격을"+
				" 부르면 한 시장의 느린 official 응답이 다른 시장의 주기 전체를 세운다 —"+
				" 그리고 그 주기는 레인 평가와 dispatch 를 함께 들고 있다.",
				name, callees)
		}
	}

	// (2) 잠금을 만지는 함수 목록을 얼린다. 목록이 비면 (1) 이 공허해지고,
	// 목록이 늘면 새 자리가 생긴 것이다.
	wantHolders := []string{
		"Context.joinStrategyRefreshWave",
		"Context.publishStrategyRefreshWave",
	}
	if got := sortedKeys(holders); strings.Join(got, "\n") != strings.Join(wantHolders, "\n") {
		t.Fatalf("공유 assembly mutex 를 만지는 함수가 바뀌었다.\n got:\n  %s\nwant:\n  %s\n\n"+
			"이 목록을 늘리기 전에 그 함수가 잠금을 든 채 무엇을 부르는지 적을 것.",
			strings.Join(got, "\n  "), strings.Join(wantHolders, "\n  "))
	}

	// (3) 그 두 함수가 **부르는 것**까지 얼린다. (1) 은 한 다리만 본다. 잠금
	// 안에서 헬퍼를 부르고 그 헬퍼가 원격을 부르면 (1) 은 초록이다.
	wantCallees := map[string][]string{
		"Context.joinStrategyRefreshWave": {
			"c.strategyRefreshMu.Lock",
			"c.strategyRefreshMu.Unlock",
			"make",
			"now.Before",
			"now.Sub",
		},
		"Context.publishStrategyRefreshWave": {
			"c.strategyRefreshMu.Lock",
			"c.strategyRefreshMu.Unlock",
			"close",
		},
	}
	for _, name := range wantHolders {
		got := holders[name]
		want := wantCallees[name]
		if strings.Join(got, "\n") != strings.Join(want, "\n") {
			t.Fatalf("%s 가 잠금을 든 채 부르는 것이 바뀌었다.\n got:\n  %s\nwant:\n  %s\n\n"+
				"여기에 새 호출이 들어오면 그 호출이 원격·파일·저널을 타는지 먼저 잴 것.",
				name, strings.Join(got, "\n  "), strings.Join(want, "\n  "))
		}
	}
}

// TestExactlyOneFunctionRunsTheRemoteAuthorityWave 는 파도를 도는 자리가 하나뿐
// 임을 얼린다. 자리가 둘이면 single-flight 가 하나에만 걸려 있어도 시험은
// 초록이고, 두 시장은 다시 각자 원격을 탄다.
func TestExactlyOneFunctionRunsTheRemoteAuthorityWave(t *testing.T) {
	_, waveCallers := strategyRefreshMutexCensus(t)
	// 이 이름은 잠금을 만지는 두 함수 어느 쪽도 아니다. 그 서로소 관계를
	// 확인하는 것은 위 시험이고, 여기서는 자리가 **하나**임만 얼린다.
	want := []string{"Context.refreshPairedStrategyEntryProductionAssembly"}
	if got := sortedKeys(waveCallers); strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("원격 권한 파도를 도는 자리가 바뀌었다.\n got:\n  %s\nwant:\n  %s\n\n"+
			"자리가 늘면 single-flight 를 지나치는 경로가 생긴 것이다.",
			strings.Join(got, "\n  "), strings.Join(want, "\n  "))
	}
}

// strategyRefreshMutexCensus 는 `internal/app/engine` 의 **모든 비시험 파일**을
// 훑는다. 파일 하나를 고르면 완전성이 고른 사람의 것이 된다.
//
// holders 는 잠금 식별자를 본문에 가진 함수 이름 → 그 함수의 호출 목록.
// waveCallers 는 원격 파도를 부르는 함수 이름의 집합이다.
func strategyRefreshMutexCensus(t *testing.T) (holders map[string][]string, waveCallers map[string]struct{}) {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read engine package: %v", err)
	}
	holders = map[string][]string{}
	waveCallers = map[string]struct{}{}
	fset := token.NewFileSet()
	scanned := 0
	declarations := 0
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, filepath.Join(".", name), nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		scanned++
		for _, decl := range file.Decls {
			fn, isFunc := decl.(*ast.FuncDecl)
			if !isFunc || fn.Body == nil {
				// 선언 자리는 함수가 아니다. 개수만 세어 아래에서 확인한다.
				ast.Inspect(decl, func(node ast.Node) bool {
					field, isField := node.(*ast.Field)
					if !isField {
						return true
					}
					for _, fieldName := range field.Names {
						if fieldName.Name == strategyRefreshMutexName {
							declarations++
						}
					}
					return true
				})
				continue
			}
			owner := strategyFunctionOwner(fn)
			touches := false
			calls := []string{}
			ast.Inspect(fn.Body, func(node ast.Node) bool {
				switch typed := node.(type) {
				case *ast.Ident:
					if typed.Name == strategyRefreshMutexName {
						touches = true
					}
				case *ast.CallExpr:
					spelling := types.ExprString(typed.Fun)
					calls = append(calls, spelling)
					if calleeMethodName(typed.Fun) == strategyRemoteWaveCall {
						waveCallers[owner] = struct{}{}
					}
				}
				return true
			})
			if touches {
				sort.Strings(calls)
				holders[owner] = dedupeStrings(calls)
			}
		}
	}
	if scanned < 10 {
		t.Fatalf("비시험 파일 %d 개만 훑었다 — 셈의 범위가 무너졌다", scanned)
	}
	if declarations != 1 {
		t.Fatalf("%s 선언이 %d 개다 — 공유 잠금이 하나가 아니면 이 셈은 무의미하다",
			strategyRefreshMutexName, declarations)
	}
	return holders, waveCallers
}

// calleeMethodName 은 **메서드 호출**의 이름이다. `c.Foo` 는 `Foo` 가 되고 맨
// 이름 `Foo` 는 빈 문자열이 된다.
//
// 맨 이름을 빼는 것은 취향이 아니라 필요다. 이 패키지에는 같은 이름이 둘 있다:
// `*Context` 의 메서드(원격을 타는 파도)와, 관측값 하나로 두 dormant worker 만
// 만드는 패키지 수준 함수(`strategy_entry_supervisor.go` 의 값 전용 생성자,
// `NewDormantStrategyEntrySupervisor` 가 부른다). 후자는 원격을 타지 않는다.
func calleeMethodName(fun ast.Expr) string {
	selector, isSelector := fun.(*ast.SelectorExpr)
	if !isSelector {
		return ""
	}
	return selector.Sel.Name
}

// strategyFunctionOwner 는 수신자를 포함한 함수 이름이다. 수신자를 빼면 같은
// 이름의 메서드와 함수가 한 칸으로 접힌다.
func strategyFunctionOwner(fn *ast.FuncDecl) string {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return fn.Name.Name
	}
	return fmt.Sprintf("%s.%s", strings.TrimPrefix(types.ExprString(fn.Recv.List[0].Type), "*"), fn.Name.Name)
}

func dedupeStrings(values []string) []string {
	unique := make([]string, 0, len(values))
	for index, value := range values {
		if index > 0 && values[index-1] == value {
			continue
		}
		unique = append(unique, value)
	}
	return unique
}

func sortedKeys[V any](source map[string]V) []string {
	keys := make([]string, 0, len(source))
	for key := range source {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
