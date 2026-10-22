package journal

// a087 tasks 3.7 — intents.price 가 NULL(시장가 보호 청산)일 때 저널 접근자 둘의 처분을 못 박음.
//
// 열거는 openspec/changes/a087-a-protective-exit-is-a-market-order/analysis/intents-price-null/ 에 있음:
// 비시험 Go 에서 intents 의 price 를 읽는 SQL 은 `intentSelect`(scanIntent — LookupIntent·Prepare)와
// `LiveOrdersForSymbol` 의 coalesce 둘뿐임. LookupIntent 의 NULL → "" 는 TestPrepareStoresEmptyPriceAsNull 이
// 이미 고정하므로, 여기서는 남은 둘 — Prepare 의 동일성 비교와 LiveOrdersForSymbol — 만 고정함.

import (
	"context"
	"errors"
	"testing"
)

// marketSellIntent 는 a087 이후 엔진이 내는 보호 청산 intent 의 모양 — KR 시장가 매도, 가격 없음.
func marketSellIntent(id string) Intent {
	return Intent{
		ID: id, Market: "kr", TradingDay: "2026-03-30", AccountRef: "acct-1",
		Symbol: "005930", Side: "SELL", OrderType: "MARKET", Quantity: "10",
		Price: "", Currency: "KRW", Source: "engine", Fingerprint: "fp-" + id,
	}
}

// TestA087ReprepareOfAMarketIntentIsNotAMutation 은 NULL 왕복이 대칭임을 고정함: 같은 시장가 intent 로 둘째
// attempt(취소)를 열 때 Prepare 는 저장된 행을 scanIntent 로 읽어 요청과 통째 비교(durability.go `existing != req.Intent`)
// 하므로, NULL 이 ""로 돌아오지 않으면 ErrIntentMutated 로 거절돼 시장가 보호 청산의 취소가 막힘.
func TestA087ReprepareOfAMarketIntentIsNotAMutation(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	if _, err := j.Prepare(ctx, PrepareRequest{
		Intent: marketSellIntent("intent-mkt"), Kind: KindPlace, AttemptID: "attempt-place",
	}); err != nil {
		t.Fatalf("first Prepare: %v", err)
	}
	_, err := j.Prepare(ctx, PrepareRequest{
		Intent: marketSellIntent("intent-mkt"), Kind: KindCancel, AttemptID: "attempt-cancel",
		TargetOrderID: "order-mkt",
	})
	if errors.Is(err, ErrIntentMutated) {
		t.Fatalf("second Prepare of the same market intent = %v — NULL price did not round-trip to \"\"", err)
	}
	if err != nil {
		t.Fatalf("second Prepare: %v", err)
	}
}

// TestA087LiveOrdersForSymbolReadsANullPriceAsEmpty 는 미체결 목록 질의가 i.price 를 빈 문자열로 coalesce 하는 것을 고정함 —
// 이 목록은 청산 전 청소(clearTheSymbol)·발의 보류 알림·reconcile 하한이 읽으므로, NULL 을 string 으로 Scan 하다
// 실패하면 그 셋이 함께 멈춤(손절 경로의 청소가 막힘).
func TestA087LiveOrdersForSymbolReadsANullPriceAsEmpty(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	attempt, err := j.Prepare(ctx, PrepareRequest{
		Intent: marketSellIntent("intent-mkt"), Kind: KindPlace, AttemptID: "attempt-place",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := attempt.MarkDispatchStarted(ctx); err != nil {
		t.Fatal(err)
	}
	if err := attempt.MarkAcked(ctx, "order-mkt"); err != nil {
		t.Fatal(err)
	}
	if err := attempt.Settle(ctx, StateConfirmed, ReasonBrokerAcknowledged, "acked"); err != nil {
		t.Fatal(err)
	}

	live, err := j.LiveOrdersForSymbol(ctx, "acct-1", "kr", "005930")
	if err != nil {
		t.Fatalf("LiveOrdersForSymbol over a NULL-priced intent: %v", err)
	}
	if len(live) != 1 || live[0].OrderID != "order-mkt" || live[0].Side != "SELL" {
		t.Fatalf("live orders = %+v, want the working market sell", live)
	}
	if live[0].Price != "" {
		t.Fatalf("live order price = %q, want \"\" for a market order (coalesce)", live[0].Price)
	}
}
