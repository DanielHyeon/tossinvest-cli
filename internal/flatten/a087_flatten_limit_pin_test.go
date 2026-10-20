package flatten_test

// a087_flatten_limit_pin_test.go — a087 tasks 4.1 「flatten 무변화」 고정 시험.
//
// a087 은 엔진 보호 청산을 MARKET 으로 바꾸려고 실행 관문 ②(execgw.checkOrderShape)에 축소 시장가 예외를 연다.
// flatten 의 청산 매도(Saga.sell → Gateway → CheckPlace)는 그 관문의 상류이므로(a087 analysis/gate-0.4-codegraph.md)
// ② 가 열려도 flatten 이 내는 주문은 지금 그대로 LIMIT 이어야 함 — 가격은 거래소 하한가가 1순위이고, 가격 제한폭을 못 읽으면
// 최근 체결가에서 공격적 할인을 뺀 틱 내림값임(liquidate.go aggressivePrice). 이 시험은 실제 Gateway·trading.Service 를 지나
// 브로커 대역에 닿은 주문을 두 가격 출처 모두에서 고정함.

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
)

func TestA087FlattenStillSellsAtAnAggressiveLimit(t *testing.T) {
	cases := []struct {
		name      string
		lower     map[string]float64
		last      map[string]float64
		wantPrice float64
	}{
		{
			// 1순위: 거래소 하한가 그대로.
			name:      "exchange lower limit first",
			lower:     map[string]float64{"005930": 63000},
			last:      map[string]float64{"005930": 71234},
			wantPrice: 63000,
		},
		{
			// 가격 제한폭이 없으면 최근 체결가 × (1 − 기본 할인 5%) 를 100원 틱으로 내림 — 그래도 LIMIT.
			name:      "last trade less the discount",
			last:      map[string]float64{"005930": 71234},
			wantPrice: 67600,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			account := &accountFake{
				positions: []domain.Position{position("005930", "kr", 2)},
				sellable:  map[string]float64{"005930": 2},
				lower:     tc.lower,
				last:      tc.last,
			}
			h := newLiqHarness(t, [][]json.RawMessage{{}}, account)
			ctx := context.Background()
			saga := h.saga(false)
			h.sell.after = account.setFlat

			if _, err := saga.CancelAll(ctx); err != nil {
				t.Fatalf("CancelAll: %v", err)
			}
			if _, err := saga.Liquidate(ctx); err != nil {
				t.Fatalf("Liquidate: %v", err)
			}

			// 주문이 0건이면 관문이 거부했거나 가격을 못 정한 것 — 어느 쪽이든 flatten 동작이 바뀐 것이라 실패임.
			orders := h.sell.orders()
			if len(orders) != 1 {
				t.Fatalf("orders that reached the broker = %+v, want exactly one", orders)
			}
			got := orders[0]
			if got.OrderType != "limit" {
				t.Errorf("order type = %q, want limit — a087 must not turn flatten into a market sell", got.OrderType)
			}
			if got.Side != "sell" || got.Quantity != 2 {
				t.Errorf("side/quantity = %s/%v, want sell/2", got.Side, got.Quantity)
			}
			if got.Price != tc.wantPrice {
				t.Errorf("price = %v, want %v", got.Price, tc.wantPrice)
			}
		})
	}
}
