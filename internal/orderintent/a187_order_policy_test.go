package orderintent

import "testing"

// a187: 엄격 호가 리더가 영문 섞인 KR 코드를 받게 되어도 주문 쪽 문법과 시장 추론은 숫자 전용 그대로임.
func TestA187OrderPolicyStaysNumericOnly(t *testing.T) {
	if looksLikeKRSymbol("0101N0") || looksLikeKRSymbol("0158S0") || !looksLikeKRSymbol("005930") {
		t.Fatal("주문 쪽 KR 종목 문법이 바뀜")
	}
	for code, want := range map[string]string{"A0101N0": "us", "0101N0": "us", "A005930": "kr"} {
		if got := InferMarketFromStockCode(code); got != want {
			t.Fatalf("InferMarketFromStockCode(%q) = %q, want %q", code, got, want)
		}
	}
}
