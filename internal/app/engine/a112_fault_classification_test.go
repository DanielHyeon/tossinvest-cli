//go:build tossos_testseams

package engine

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// 태스크 5.6 의 문장 앞머리는 "journal/Gateway/fence/multiple-owner 고장은 모든
// 신규 진입을 막는다"이다. 그 문장에는 **두 가지 읽기**가 있고, 둘의 차이가
// 손절을 놓는 loop 의 생사다.
//
//  1. 막는다 = 그 사이클이 주문을 내지 않는다 (진입만 닫힌다).
//  2. 막는다 = 중앙 무결성 고장으로 올려 프로세스를 세운다.
//
// 오늘 코드는 1 이다. 이 파일이 그것을 값으로 확인한다: 세 종류의 고장을 실제
// 저널·실제 Guardian·spy Gateway 로 주입하고, 매번 (a) Gateway place 호출이
// 0 이고 (b) 반환된 오류가 `isCentralStrategyIntegrity` 로 **분류되지 않는다**를
// 함께 본다.
//
// 왜 (b) 가 중요한가. 2 를 고르면 `runMarket` 이 `signalCentral` 을 부르고
// `Run` 이 반환하며, Runtime 은 첫 정지에 **모든 loop 를 취소한다** —
// fill detection·reconcile·exit observation 을 포함해서. 저널 읽기 한 번이
// 실패했다고 손절을 놓는 주체를 끄는 것은 보수적인 선택이 아니다.
//
// 이 시험은 dispatch 사이클을 **직접** 부른다. 생산 경로
// (`Context.runProductionStrategyMarketCycle`)는 이 반환값을 그대로 위로
// 올리므로(AST 산출물
// `internal-app-engine--context.runproductionstrategymarketcycle/ast.json` 의
// `453:11` dispatch 호출과 그 뒤 `return err`), 분류는 여기서 정해진다.
func TestNoJournalOrGatewayFaultInTheDispatchCycleIsClassifiedCentral(t *testing.T) {
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		lower := strings.ToLower(string(market))
		cases := []struct {
			name string
			// arm 은 fixture 를 받아 고장을 심고, 기대하는 오류 판정을 돌려준다.
			arm func(spy *strategyDispatchGatewaySpy, closeJournal func()) func(error) bool
		}{
			{
				name: "Gateway 보호 관측 실패",
				arm: func(spy *strategyDispatchGatewaySpy, _ func()) func(error) bool {
					refusal := errors.New("protection unavailable")
					spy.failProtection = map[string]error{lower: refusal}
					return func(err error) bool { return errors.Is(err, refusal) }
				},
			},
			{
				name: "Gateway 진입 게이트 거절",
				arm: func(spy *strategyDispatchGatewaySpy, _ func()) func(error) bool {
					refusal := errors.New("entry gate blocked")
					spy.failEntryGate = map[string]error{lower: refusal}
					return func(err error) bool { return errors.Is(err, refusal) }
				},
			},
			{
				// 측정: 닫힌 저널에서 dispatch 가 처음 걸리는 자리는 소유권
				// 획득(`dispatchOwner` → `AcquireStrategyDispatchOwner`)이고
				// 오류는 `sql: database is closed` 다. 그래서 이 칸의 이름은
				// "저널"이 아니라 **fence** 다 — 뒤쪽 저널 쓰기(리스 발급·CAS)는
				// 이 입력으로는 도달하지 않는다. 그 사실이 이 파일만으로는
				// 부족하다는 뜻이고, 나머지는 a112_central_integrity_census_test.go
				// 가 패키지 전체 열거로 덮는다.
				name: "소유권 fence 를 얻을 수 없다",
				arm: func(_ *strategyDispatchGatewaySpy, closeJournal func()) func(error) bool {
					closeJournal()
					return func(err error) bool { return err != nil }
				},
			},
		}
		for _, testCase := range cases {
			t.Run(string(market)+"/"+testCase.name, func(t *testing.T) {
				cycle, proposals, j, spy := pairedStrategyDispatchCycleFixture(t)
				result := proposals.forMarket(market).entries[0].authority.Proposal()
				matches := testCase.arm(spy, func() { _ = j.Close() })

				_, err := cycle.dispatch(context.Background(), deliverForTest(t, result))
				if err == nil {
					t.Fatal("고장을 심었는데 dispatch 가 성공했다 — 진입이 막히지 않았다")
				}
				if !matches(err) {
					t.Fatalf("dispatch error=%v — 심은 고장이 아니다", err)
				}
				if isCentralStrategyIntegrity(err) {
					t.Fatalf("dispatch error=%v 가 중앙 무결성으로 분류됐다.\n\n"+
						"그 분류는 runMarket 이 signalCentral 을 부르게 하고, Run 의 반환이"+
						" Runtime 의 첫 정지가 되어 **모든 loop 를 취소한다** — fill/exit/"+
						"reconcile 포함. 즉 저널·Gateway 고장 하나가 손절을 놓는 주체를 끈다."+
						" 이 change 는 그 방향으로 가지 않는다. design.md:198 의 고장표를"+
						" 그렇게 읽고 싶다면 fail-closed 의 수단이 프로세스 정지가 아니라"+
						" execgw.EntryGate.Block 이어야 하고, 그것은 사람의 결정이다.", err)
				}
				spy.mu.Lock()
				places := len(spy.calls)
				spy.mu.Unlock()
				if places != 0 {
					t.Fatalf("Gateway place calls=%d — 고장 뒤에 주문이 나갔다", places)
				}
			})
		}
	}
}
