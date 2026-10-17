package execgw_test

// a094 4.N4e · D−5.1 — 응답 종목이 없거나 비면 확인 실패. 발주 직후 확인(Gateway.Place 의 read-back)과 기동 확정
// (ConfirmPlacedOrder — reconcile 이 부름)이 같은 함수라 한 변이가 둘을 깬다.

import (
	"context"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

func a094DetailWithSymbolField(id, symbolField string) string {
	body := `{"result":{"orderId":"` + id + `",` + symbolField + `"side":"BUY","status":"PENDING","quantity":"2","price":"70000",` +
		`"currency":"KRW","orderedAt":"2026-03-30T10:30:00+09:00","canceledAt":null,"execution":{"filledQuantity":"0"}}}`
	return body
}

func TestA094TheReadBackNeedsANamedSymbol(t *testing.T) {
	const acked = "O-1"
	for _, tc := range []struct {
		name, field string
	}{
		{"no symbol field", ``},
		{"null symbol", `"symbol":null,`},
		{"empty symbol", `"symbol":"",`},
		{"blank symbol", `"symbol":"   ",`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			detail := a094DetailWithSymbolField(acked, tc.field)

			// 발주 직후 확인 — IN_DOUBT(ack_round_trip_unconfirmed).
			broker := &fakeBroker{result: domain.MutationResult{Kind: "place", Status: "accepted", OrderID: acked}}
			reader := newOrderReader()
			reader.set(acked, detail)
			gw, j, clk := newGatewayWithOrders(t, broker, reader)
			out, err := gw.Place(context.Background(), placeRequest(t, j, clk))
			if err == nil || out.State != journal.StateInDoubt {
				t.Fatalf("post-place: state %s err %v, want IN_DOUBT — a read-back naming no symbol confirms nothing", out.State, err)
			}

			// 기동 확정이 부르는 같은 함수.
			if err := execgw.ConfirmPlacedOrder(context.Background(), reader, acked, "005930"); err == nil ||
				!strings.Contains(err.Error(), "symbol") {
				t.Fatalf("boot confirm: err = %v, want a symbol refusal", err)
			}
		})
	}
	// 대조 — 종목이 있으면 두 경로 다 확인.
	reader := newOrderReader()
	reader.set(acked, orderDetailJSON(acked, "005930"))
	if err := execgw.ConfirmPlacedOrder(context.Background(), reader, acked, " 005930 "); err != nil {
		t.Fatalf("control: a named, matching symbol must confirm: %v", err)
	}
}
