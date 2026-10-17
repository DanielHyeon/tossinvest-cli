package officialbars

// a112 8.2(Manager 판정 2026-10-04): officialbars 는 공식 API 의 **읽기** 셋(StrictMinuteCandles · StrictOrderbookTop ·
// StrictLastPrice)을 쓰려고 internal/official 을 들여오고, 그 패키지는 주문 변경자(PlaceOrder · CancelOrder · ModifyOrder · 조건부
// 주문 3종)와 trading · orderintent · config 를 함께 끌고 온다. 기존 가드(guard_test.go)는 직접 import 허용 목록이라 「official 이
// 있다」만 보고 그 안의 무엇을 부르는지는 보지 않는다 — 존재 검사다. 여기서 셋을 묶는다:
//   ① 폐포 걸음 넷(생산 · 시험 이진 × 무태그 · 태그)에서 능력 패키지는 **internal/official 을 거쳐서만** 닿는다(official 을 잘라 낸
//      걸음에 능력 0). journal · execgw · ops 같은 것은 어떤 길로도 없다.
//   ② 생산 파일이 official 에서 쓰는 것은 읽기 응답의 **값 타입** 넷과 그 필드 · 메서드뿐이다 — 타입 검사기가 해소한 객체로 센다.
//      패키지 수준 함수(official.New 포함)와 *official.Client 의 메서드는 하나도 없다(주문 변경자는 그 부분집합).
//   ③ 시험 파일은 주문 변경자 이름을 부르지 않는다(a112_live_host_guard_test.go 가 시험 쪽 나머지를 맡음).

import (
	"go/types"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/testenv"
)

const (
	a112BarsPath     = testenv.ModulePath + "internal/officialbars"
	a112OfficialPath = testenv.ModulePath + "internal/official"
)

// a112OfficialMutators 는 *official.Client 의 주문 변경 메서드다(internal/official orders_write.go · conditional_writes.go).
var a112OfficialMutators = map[string]bool{
	"PlaceOrder": true, "CancelOrder": true, "ModifyOrder": true,
	"CreateConditionalOrder": true, "CancelConditionalOrder": true, "ModifyConditionalOrder": true, "ModifyConditionalOrderRef": true,
}

func TestOfficialBarsReachesMutationCapabilitiesOnlyThroughTheOfficialClient(t *testing.T) {
	for _, mode := range testenv.WalkModes() {
		t.Run(mode.Name, func(t *testing.T) {
			graph := testenv.ListDeps(t, ".", mode)
			roots := graph.Roots(a112BarsPath)
			if len(roots) == 0 {
				t.Fatalf("the walk has no root for %s", a112BarsPath)
			}
			whole := graph.Reachable(roots, nil)
			// 양성 대조: 잘라 낼 대상이 실제로 폐포에 있다 — 없으면 아래 「잘라 낸 걸음」 은 아무것도 자르지 않은 것이다.
			if !whole[a112OfficialPath] {
				t.Fatalf("%s: internal/official is not in the closure — the cut below would prove nothing", mode.Name)
			}
			// official 을 거치지 않는 길에는 능력이 하나도 없다.
			cut := graph.Reachable(roots, map[string]bool{a112OfficialPath: true})
			if found := testenv.ForbiddenIn(cut); len(found) != 0 {
				t.Fatalf("%s: officialbars reaches a mutation capability by a path that does not go through internal/official:\n%v", mode.Name, found)
			}
			// official 을 거쳐 닿는 것도 그 패키지가 실제로 들여오는 넷(official 자신 · trading · orderintent · config)뿐이다 —
			// journal · execgw · ops 같은 것이 official 뒤로 새로 들어오면 여기서 멈춘다.
			if found := testenv.ForbiddenIn(whole, "internal/official", "internal/trading", "internal/orderintent", "internal/config"); len(found) != 0 {
				t.Fatalf("%s: the closure behind internal/official grew a new capability:\n%v", mode.Name, found)
			}
		})
	}
}

func TestOfficialBarsProductionUsesOnlyTheReadResponseTypesOfTheOfficialClient(t *testing.T) {
	checked := testenv.TypeCheckProduction(t, ".", a112BarsPath, "tossos_testseams")
	valueTypes := map[string]bool{
		"StrictMinutePage": true, "RawMinuteCandle": true, // 분봉 읽기 응답
		"StrictTopOfBook": true, "StrictLastPrice": true, // 호가 · 현재가 읽기 응답
		"RateBudget": true, // 분봉 응답이 싣는 호출 한도 값(StrictMinutePage.Budget) — Exhausted 는 값 판정
	}
	seenTypes := 0
	for _, use := range checked.UsesFrom(a112OfficialPath) {
		switch object := use.Object.(type) {
		case *types.Func:
			receiver := testenv.ReceiverNamed(object)
			if a112OfficialMutators[object.Name()] {
				t.Errorf("%s: calls the order mutator %s", use.Position, testenv.Describe(object))
				continue
			}
			if receiver == "" || !valueTypes[receiver] {
				t.Errorf("%s: uses %s — production officialbars may use only the read-response value types, never a client or package function", use.Position, testenv.Describe(object))
			}
		case *types.TypeName:
			if !valueTypes[object.Name()] {
				t.Errorf("%s: names %s — not one of the read-response value types", use.Position, testenv.Describe(object))
			}
			seenTypes++
		case *types.Var:
			if !object.IsField() {
				t.Errorf("%s: uses %s — package-level official state", use.Position, testenv.Describe(object))
			}
		case *types.Const:
			// 상수는 값이라 능력이 아니다 — 허용.
		default:
			t.Errorf("%s: uses %s — unclassified official symbol", use.Position, testenv.Describe(object))
		}
	}
	if seenTypes == 0 {
		t.Fatal("the census saw no official type — it read nothing")
	}
}
