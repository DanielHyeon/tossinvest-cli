package trading

// a087_conditional_path_unchanged_pin_test.go — a087 tasks 4.3 「조건주문 경로 무변화 (2c 범위)」 고정 시험.
//
// a087 §1.5 는 일반 주문 관문 ③(placeIntentSupported)에 엔진 인스턴스 한정 「sell+market·가격 없음」 예외를 연다.
// 조건주문(SINGLE+MARKET 매도 STOP — measurements.md M12·M38)은 그 관문을 지나지 않는 별도 경로(ConditionalGate →
// ConditionalBroker)이고 a087 은 그 경로를 바꾸지 않는다(design 「건드리지 않는 것」·D9). 이 시험은 지금의 성질을 못 박음:
//
//   - 기본 Service 에서 일반 주문 ③ 은 비분수 KR 시장가 매도를 지원하지 않음(오늘 동작).
//   - 같은 Service 의 조건주문 경로는 SINGLE+MARKET 매도를 그 판정과 무관하게 받아, 의도를 한 글자도 바꾸지 않고 브로커에 넘김.
//
// 둘째가 첫째에 걸리기 시작하면(조건주문이 ③ 을 거치게 되면) 또는 의도가 변형되면 이 시험이 실패함.

import (
	"context"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
)

func TestA087ConditionalMarketStopBypassesThePlacePredicateUnchanged(t *testing.T) {
	t.Parallel()

	broker := &fakeConditionalBroker{}
	svc := NewService(openConditionalPolicy(), nil).WithConditional(broker)

	// 전제: 같은 기본 Service 의 일반 주문 ③ 은 오늘 이 모양을 비지원으로 판정함(PreviewPlace.LiveReady 가 ③ 의 답).
	// 이것이 참이 아니면 아래 대조가 무의미함. 술어를 직접 부르지 않는 것은 §1.5 가 술어 시그니처를 바꿔도 이 시험이 살아남게 하려는 것임.
	plainMarketSell := orderintent.PlaceIntent{
		Symbol: "005930", Market: "kr", Side: "sell", OrderType: "market", Quantity: 1, CurrencyMode: "KRW",
	}
	if svc.PreviewPlace(plainMarketSell).LiveReady {
		t.Fatal("precondition: the default place predicate already supports a plain KR market sell")
	}

	stop := orderintent.ConditionalPlaceIntent{
		Symbol: "005930", Type: "SINGLE", OrderType: "MARKET", ExpireDate: "2026-12-31", Quantity: 1,
		ClientOrderID: "a087-pin",
		// MARKET 이므로 주문가 없음(0) — 발동가만 있음.
		First: orderintent.ConditionLeg{OrderSide: "SELL", TriggerPrice: 65000},
	}
	preview := svc.PreviewConditionalPlace(stop)
	if !preview.LiveReady || !preview.MutationReady {
		t.Fatalf("preview of a SINGLE+MARKET stop: LiveReady=%v MutationReady=%v, want both true",
			preview.LiveReady, preview.MutationReady)
	}
	if _, err := svc.ConditionalPlace(context.Background(), stop, ExecuteOptions{
		Execute: true, Confirm: preview.ConfirmToken,
	}); err != nil {
		t.Fatalf("ConditionalPlace(SINGLE+MARKET SELL): %v", err)
	}
	if len(broker.created) != 1 {
		t.Fatalf("conditional broker received %d creations, want 1", len(broker.created))
	}
	if got := broker.created[0]; got != stop {
		t.Fatalf("the conditional path altered the intent:\n got %+v\nwant %+v", got, stop)
	}
}
