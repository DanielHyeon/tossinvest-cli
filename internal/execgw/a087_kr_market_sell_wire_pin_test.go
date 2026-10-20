package execgw_test

// a087_kr_market_sell_wire_pin_test.go — a087 tasks 4.5 「전송 표류 가드 KR market sell 골든 행」.
//
// decision_test.go TestPlaceWireBodyShapes 의 골든에는 US market sell 만 있다. a087 Phase 2 가 엔진 보호 청산을
// KR MARKET 매도로 보내면, 원장에 저장되는 사본 렌더러(PlaceWireBody, 재생이 다시 보내는 바이트)와 실제 클라이언트 빌더
// (official buildOrderCreate, PlaceOrder 가 보내는 바이트)가 KR 행에서도 같은 바이트를 내야 함 — 다르면 재생이 같은
// 멱등 키로 다른 본문을 보내 「422 idempotency-key-conflict」를 받는다(TestStoredWireBodyIsWhatTheBrokerReceived 머리말).
//
// 고정하는 것: 두 렌더러 모두 KR 시장가 매도에 가격·timeInForce 를 싣지 않음. intent 에 가격이 남아 있어도(관측가가
// 새는 경우) 싣지 않음. 골든 모양은 2026-10-19 a087 5.1 실측 영수증의 wire_body(브로커 HTTP 200 접수)와 같은 필드 구성이다
// (a087 analysis/market-sell-receipts/a087ms-20261019T145815-8bdcb862.json).

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
)

const a087KRMarketSellGolden = `{"symbol":"005930","side":"SELL","orderType":"MARKET","quantity":"1",` +
	`"clientOrderId":"tos-key","confirmHighValueOrder":false}`

// sentByOfficial 은 official.Client.PlaceOrder 가 실제로 보낸 POST /api/v1/orders 본문을 httptest 서버로 잡아 돌려줌.
func sentByOfficial(t *testing.T, intent orderintent.PlaceIntent) string {
	t.Helper()
	var captured string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/oauth2/token":
			_, _ = w.Write([]byte(`{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/orders":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Errorf("reading the captured body: %v", err)
			}
			captured = string(body)
			_, _ = w.Write([]byte(`{"result":{"orderId":"O-1","clientOrderId":"tos-key"}}`))
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	off := official.New(official.Credentials{APIKey: "k", SecretKey: "s"},
		filepath.Join(t.TempDir(), "token.json"),
		official.WithBaseURL(srv.URL), official.WithHTTPClient(srv.Client()),
		official.WithAccountSeq(7))
	if _, err := off.PlaceOrder(context.Background(), intent); err != nil {
		t.Fatalf("official PlaceOrder: %v", err)
	}
	if captured == "" {
		t.Fatal("no request body was captured")
	}
	return captured
}

func TestA087KRMarketSellWireBodyCarriesNoPriceOrTIF(t *testing.T) {
	base := orderintent.PlaceIntent{
		Symbol: "005930", Market: "kr", Side: "sell", OrderType: "market",
		Quantity: 1, CurrencyMode: "KRW", ClientOrderID: "tos-key",
	}
	stray := base
	stray.Price = 245750 // 관측가가 intent 에 남아도 시장가 본문에는 실리지 않아야 함

	for _, tc := range []struct {
		name   string
		intent orderintent.PlaceIntent
	}{
		{"kr market sell", base},
		{"kr market sell with a stray price", stray},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stored, err := execgw.PlaceWireBody(tc.intent)
			if err != nil {
				t.Fatalf("PlaceWireBody: %v", err)
			}
			if stored != a087KRMarketSellGolden {
				t.Errorf("stored copy (PlaceWireBody):\n got %s\nwant %s", stored, a087KRMarketSellGolden)
			}
			if sent := sentByOfficial(t, tc.intent); sent != a087KRMarketSellGolden {
				t.Errorf("sent bytes (official buildOrderCreate):\n got %s\nwant %s", sent, a087KRMarketSellGolden)
			}
		})
	}
}
