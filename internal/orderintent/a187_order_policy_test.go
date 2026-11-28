package orderintent

import "testing"

// a187: 엄격 호가 리더가 영문 섞인 KR 코드를 받게 되어도 주문 쪽 「시장 추론」은 숫자 전용 그대로임.
// 고정하는 것은 시장 추론(looksLikeKRSymbol·InferMarketFromStockCode)이며, 주문 검증 전체가 숫자 전용이라는 뜻이 아님
// (리뷰 r1 발견 1).
func TestA187MarketInferenceStaysNumericOnly(t *testing.T) {
	if looksLikeKRSymbol("0101N0") || looksLikeKRSymbol("0158S0") || !looksLikeKRSymbol("005930") {
		t.Fatal("주문 쪽 KR 시장 추론 문법이 바뀜")
	}
	for code, want := range map[string]string{"A0101N0": "us", "0101N0": "us", "A005930": "kr"} {
		if got := InferMarketFromStockCode(code); got != want {
			t.Fatalf("InferMarketFromStockCode(%q) = %q, want %q", code, got, want)
		}
	}
}

// 현 동작 고정: 시장을 "kr" 로 명시하면 NormalizePlace 는 영문 섞인 KR 코드를 거절하지 않고 그대로 돌려줌.
// 이 동작은 a187 이전부터 있던 것이고(a187 은 internal/orderintent 를 고치지 않음) 여기서는 좋고 나쁨을 판단하지 않고
// 바뀌지 않았음만 고정함. 실행 승인·설정·confirm 가드는 이 함수 밖에 있음.
func TestA187NormalizePlaceExplicitKRAlphanumericUnchanged(t *testing.T) {
	got, err := NormalizePlace(PlaceInput{Market: "kr", Symbol: "0101N0", Side: "buy", OrderType: "limit", Quantity: 1, Price: 1000})
	if err != nil {
		t.Fatalf("NormalizePlace: %v", err)
	}
	if got.Market != "kr" || got.Symbol != "0101N0" {
		t.Fatalf("market=%q symbol=%q", got.Market, got.Symbol)
	}
}

// 현 동작 고정(gstack 발견 7): 시장을 "us" 로 주거나 생략하면 영문 섞인 6자리는 KR 로 추론되지 않고 "us" 로 남고,
// 숫자 6자리는 "kr" 로 바뀜. 시장 추론 호출 자리(intent.go 의 looksLikeKRSymbol)를 넓히는 변경을 잡기 위함.
func TestA187NormalizePlaceMarketInferenceUnchanged(t *testing.T) {
	for _, c := range []struct {
		market, symbol, want string
	}{
		{"us", "0101N0", "us"},
		{"", "0101N0", "us"},
		{"us", "005930", "kr"},
		{"", "005930", "kr"},
	} {
		got, err := NormalizePlace(PlaceInput{Market: c.market, Symbol: c.symbol, Side: "buy", OrderType: "limit", Quantity: 1, Price: 1000})
		if err != nil {
			t.Fatalf("NormalizePlace(%q, %q): %v", c.market, c.symbol, err)
		}
		if got.Market != c.want || got.Symbol != c.symbol {
			t.Fatalf("NormalizePlace(%q, %q) market=%q symbol=%q, want market %q", c.market, c.symbol, got.Market, got.Symbol, c.want)
		}
	}
}
