package engine

// a087 tasks 3.7 — 청산 전 청소(clearTheSymbol)가 미체결 주문의 가격을 읽는 자리(workingOrderPrice)의 NULL 처분을
// 못 박음. 가격이 빈 주문(시장가 — intents.price NULL 을 LiveOrdersForSymbol 이 ""로 줌)을 실패로 읽으면
// clearTheSymbol 의 `qerr != nil || perr != nil` 갈래가 청소를 실패시켜 보호 청산을 보류시킴(a094 3.B3).
// 열거: openspec/changes/a087-a-protective-exit-is-a-market-order/analysis/intents-price-null/.

import "testing"

func TestA087AnEmptyWorkingOrderPriceIsZeroNotAFailure(t *testing.T) {
	for _, price := range []string{"", "  "} {
		got, err := workingOrderPrice(price)
		if err != nil || got != 0 {
			t.Fatalf("workingOrderPrice(%q) = (%v, %v), want (0, nil)", price, got, err)
		}
	}
	// 대조군: 가격이 있으면 그 값, 읽을 수 없으면 여전히 실패(빈 값만 너그러움).
	if got, err := workingOrderPrice("70000"); err != nil || got != 70000 {
		t.Fatalf("workingOrderPrice(70000) = (%v, %v)", got, err)
	}
	if _, err := workingOrderPrice("not-a-price"); err == nil {
		t.Fatal("an unreadable price was accepted — only an empty one may be")
	}
}
