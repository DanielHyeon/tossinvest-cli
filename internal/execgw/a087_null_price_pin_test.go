package execgw

// a087 tasks 3.7 — intents.price 가 NULL 이어서 Intent.Price == "" 로 온 시장가 보호 청산 intent 를 IN_DOUBT
// 해소기가 읽는 두 자리의 처분을 못 박음(열거: openspec/changes/a087-…/analysis/intents-price-null/).
//
//   - newMatcher → parseOptionalDecimal("") = 0 → matcher.matches 의 가격 조건(`m.price > 0 && …`)이 꺼짐 —
//     가격 없는 주문은 가격으로 거르지 않고 종목·방향·수량·제출 창으로만 대조함(와일드카드).
//   - intentNotional = 0 → absenceCorroborated 의 매수가능금액 교차 확인(`notional > 0 && …`)이 꺼짐 — 부재 확증은
//     보유수량 불변(그 앞 갈래)에 기대고, 매도 접수가 매수가능금액을 어떻게 움직이는지는 [미측정] 이라 이 갈래가
//     지정가 매도에서 실제로 무엇을 거르는지도 미검증임(issues.md I-R10).

import (
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

func a087MarketSellIntent() journal.Intent {
	return journal.Intent{
		ID: "intent-mkt", Market: "kr", TradingDay: "2026-03-30", AccountRef: "acct-1",
		Symbol: "005930", Side: "SELL", OrderType: "MARKET", Quantity: "10", Price: "",
		Currency: "KRW", Source: "engine", Fingerprint: "fp-mkt",
	}
}

func TestA087AnEmptyIntentPriceMatchesAnyBrokerPrice(t *testing.T) {
	m, err := newMatcher(a087MarketSellIntent(), journal.AttemptRecord{}, ResolveConfig{})
	if err != nil {
		t.Fatalf("newMatcher over a market intent: %v", err)
	}
	if m.price != 0 {
		t.Fatalf("matcher price = %v, want 0 (no price to compare)", m.price)
	}
	// 시장가 주문을 브로커가 체결가로 보고하든 0 으로 보고하든 같은 주문으로 대조돼야 함.
	for _, brokerPrice := range []float64{0, 70100} {
		order := orderFacts{OrderID: "order-mkt", Symbol: "005930", Side: "SELL", Quantity: 10, Price: brokerPrice}
		if !m.matches(order) {
			t.Fatalf("broker price %v: the market intent did not match its own order", brokerPrice)
		}
	}
	// 가격이 꺼져도 나머지 대조는 살아 있음 — 수량이 다르면 다른 주문.
	if m.matches(orderFacts{Symbol: "005930", Side: "SELL", Quantity: 9, Price: 70100}) {
		t.Fatal("a different quantity matched — the wildcard must cover the price only")
	}
	// 대조군: 가격이 있는 intent 는 다른 가격의 주문과 대조되지 않음(와일드카드가 빈 가격에서만 켜짐).
	limit := a087MarketSellIntent()
	limit.OrderType, limit.Price = "LIMIT", "70000"
	ml, err := newMatcher(limit, journal.AttemptRecord{}, ResolveConfig{})
	if err != nil {
		t.Fatal(err)
	}
	if ml.matches(orderFacts{Symbol: "005930", Side: "SELL", Quantity: 10, Price: 70100}) {
		t.Fatal("a priced intent matched a different broker price — control failed")
	}
}

func TestA087AnEmptyIntentPriceHasNoNotional(t *testing.T) {
	if got := intentNotional(a087MarketSellIntent()); got != 0 {
		t.Fatalf("notional of a market intent = %v, want 0", got)
	}
	// 대조군: 같은 수량의 지정가 intent 는 수량×가격.
	limit := a087MarketSellIntent()
	limit.Price = "70000"
	if got := intentNotional(limit); got != 700000 {
		t.Fatalf("notional of a priced intent = %v, want 700000", got)
	}
}
