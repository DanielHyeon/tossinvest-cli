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

// 이 파일은 셈으로 답한다: **무엇이 엔진을 세울 수 있는가.**
//
// `ErrStrategyCentralIntegrity` 가 `runMarket` 을 지나면 `Run` 이 반환하고,
// Runtime 은 첫 정지에 모든 loop 를 취소한다 — fill detection·reconcile·
// exit observation 을 포함해서. 엔진이 서면 손절을 놓는 주체가 사라진다.
// 그러므로 "이 값을 만들 수 있는 자리"의 목록은 안전 계약이다.
//
// 왜 행동 시험으로는 부족한가. 행동은 **주입한 고장**만 본다. 저널 리스 발급을
// 중앙으로 승격시키는 변이(M13)는 이 change 의 반증 배터리에서 **살아남았다** —
// 닫힌 저널이 그 줄에 닿기 전에 소유권 획득에서 먼저 실패하기 때문이다.
// 즉 "각 고장마다 시험 하나"는 여기서도 끝나지 않는다(5.5·5.1.1 과 같은 결론).
// 그래서 셈의 **범위를 함수 본문에서 패키지 전체로 올린다**: 비시험 파일 전부를
// 훑어 이 값을 만드는 자리를 열거하고, 얼어붙은 목록과 대조한다.
//
// 목록에 한 줄이 늘면 이 시험이 실패한다. 그것이 요점이다 — 새 자리를 여는 것은
// "엔진을 세울 새 이유를 만드는 것"이고, 그 판단은 조용히 지나가면 안 된다.
func TestTheCompleteCensusOfWhatCanStopTheEngineOnAStrategyFault(t *testing.T) {
	mints, callers := centralIntegrityCensus(t)

	// (1) `StrategyCentralIntegrityFailure` 는 **생산 호출자가 없다.** 그 함수는
	// 평가 콜백이 자기 불변식 위반을 프로세스 전체로 올리라고 만들어 둔 문이고,
	// 오늘 그 문을 여는 생산 코드는 없다. 저널·Gateway·fence 고장은 그냥 오류로
	// 올라가 시장 국소로 남는다 — a112_fault_classification_test.go 가 값으로 확인한다.
	if len(callers) != 0 {
		t.Fatalf("생산 코드가 StrategyCentralIntegrityFailure 를 부른다: %v\n\n"+
			"그 호출은 해당 고장을 프로세스 정지로 승격시킨다. 진입만 닫고 싶다면"+
			" 수단은 execgw.EntryGate.Block 이지 이것이 아니다. design.md:198 의"+
			" 고장표를 '프로세스 정지'로 읽는 결정은 사람이 내려야 한다.", callers)
	}

	// (2) sentinel 이 나타나는 자리 **전부**를 세어 얼린다. 어느 것이 "만들기"이고
	// 어느 것이 "읽기"인지를 시험이 판정하지 않는 이유는, 그 판정 자체가 우회
	// 지점이 되기 때문이다 — 열거는 판정하지 않으므로 우회할 것이 없다.
	// 역할은 아래 주석이 지고, 목록은 코드가 진다.
	//
	// **좌표는 파일 절대 줄이 아니라 감싼 선언의 시작으로부터의 거리다.**
	// 첫 판본은 절대 줄을 얼렸고, 그 함수와 아무 상관 없는 위쪽 편집(태스크
	// 5.1.2 가 `runProductionStrategyMarketCycle` 에 열여섯 줄을 넣은 것)에
	// 전부 어긋났다. 그때 통과시키는 방법은 기대 목록을 고치는 것뿐인데,
	// 그것은 이 시험이 막으려던 바로 그 편집과 구분되지 않는다. 거리로 재면
	// 남의 편집에는 흔들리지 않고, 이 선언 **안에서** 자리가 옮겨지거나
	// 개수가 바뀌는 것은 그대로 잡는다.
	//
	//	(package)+0                       sentinel 선언 (var 블록의 첫 줄)
	//	Error+2, Error+4                  오류 문구 조립 (읽기)
	//	StrategyCentralIntegrityFailure+2 기본 cause (문 자체)
	//	isCentralStrategyIntegrity+2      분류 (읽기)
	//	Run+2                             nil 감독자 — 구성 오류
	//	Run+7                             Run 중복 호출 — 구성 오류
	//	Run+46                            runMarket 이 올린 중앙 신호의 중계
	//
	// **만들기는 이 중 셋뿐이고 셋 다 `Run` 안에 있다.** 평가 실패는 하나도 없다.
	want := []string{
		"(package)+0",
		"Error+2",
		"Error+4",
		"StrategyCentralIntegrityFailure+2",
		"isCentralStrategyIntegrity+2",
		"Run+2",
		"Run+7",
		"Run+46",
	}
	if strings.Join(mints, "\n") != strings.Join(want, "\n") {
		t.Fatalf("중앙 무결성 값을 만드는 자리가 바뀌었다.\n got:\n  %s\nwant:\n  %s\n\n"+
			"줄이 늘었다면 엔진을 세울 새 이유가 생긴 것이다. 줄이 줄었다면 fail-closed"+
			" 하나가 사라진 것이다. 어느 쪽이든 이 목록을 고치기 전에 무엇이 바뀌는지"+
			" 적을 것 — 이 값의 대가는 fill/exit/reconcile loop 의 취소다.",
			strings.Join(mints, "\n  "), strings.Join(want, "\n  "))
	}
}

// centralIntegrityCensus 는 `internal/app/engine` 의 **모든 비시험 파일**을 훑는다.
// 파일 하나를 고르면 완전성이 고른 사람의 것이 되기 때문이다(이 change 가
// 정책 영수증에서 이미 배운 것).
func centralIntegrityCensus(t *testing.T) (mints, callers []string) {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read engine package: %v", err)
	}
	fset := token.NewFileSet()
	scanned := 0
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
			owner := "(package)"
			if isFunc {
				owner = fn.Name.Name
			}
			ast.Inspect(decl, func(node ast.Node) bool {
				switch typed := node.(type) {
				case *ast.Ident:
					if typed.Name == "ErrStrategyCentralIntegrity" {
						mints = append(mints, fmt.Sprintf("%s+%d", owner,
							fset.Position(typed.Pos()).Line-fset.Position(decl.Pos()).Line))
					}
				case *ast.CallExpr:
					if types.ExprString(typed.Fun) == "StrategyCentralIntegrityFailure" {
						callers = append(callers, fmt.Sprintf("%s:%s:%d", name, owner,
							fset.Position(typed.Pos()).Line))
					}
				}
				return true
			})
		}
	}
	if scanned < 10 {
		t.Fatalf("비시험 파일 %d 개만 훑었다 — 셈의 범위가 무너졌다", scanned)
	}
	sort.Strings(callers)
	return mints, callers
}
