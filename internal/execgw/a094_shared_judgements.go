package execgw

// a094_shared_judgements.go 는 a094 가 두 호출자에게 **한 함수**로 내보낸 판정 둘임(판정을 둘로 두지 않음 —
// 저장소 교훈 「판정이 둘이면 반증이 죽는다」).
//
//  1. UnsettledOnSymbol — 같은 종목의 미종결 attempt. 주문 경로(checkSymbolFree)와 exit 청소(clearTheSymbol)가 같이 부름
//     (a094 D−4.7 · Q6-2). 청소가 다른 목록으로 치움을 판정하던 틈이 「치움 → 무장 → SymbolInFlight → 해제」 반복을 만들었음.
//  2. ConfirmPlacedOrder — 접수된 발주를 기록 번호로 읽어 확인. 발주 직후 확인(confirmCreatedOrder)과 기동의 ACKED 확정
//     (reconcile.Recovery)이 같이 부름(D−4.2 · D−5.1). 응답 종목이 비면 확인 실패 — 계약상 종목은 필수임.

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// UnsettledOnSymbol 은 같은 시장 · 종목을 겨누는 미종결(PendingAttempts) attempt 를 전부 돌려줌.
func (g *Gateway) UnsettledOnSymbol(ctx context.Context, market, symbol string) ([]journal.AttemptRecord, error) {
	return g.unsettledFor(ctx, mutationPlan{
		market: strings.ToLower(strings.TrimSpace(market)),
		symbol: strings.ToUpper(strings.TrimSpace(symbol)),
	})
}

// unsettledFor 는 판정 본체임 — 원장의 미종결 목록을 그 attempt 의 intent 시장 · 종목으로 거름(attemptTargets).
// intent 를 못 읽으면 오류 — 모르는 attempt 를 「다른 종목」 으로 치우면 차단이 조용히 빠짐.
func (g *Gateway) unsettledFor(ctx context.Context, plan mutationPlan) ([]journal.AttemptRecord, error) {
	pending, err := g.journal.PendingAttempts(ctx)
	if err != nil {
		return nil, fmt.Errorf("execgw: checking for in-flight mutations on %s: %w", plan.symbol, err)
	}
	var out []journal.AttemptRecord
	for _, rec := range pending {
		same, err := g.attemptTargets(ctx, rec, plan)
		if err != nil {
			return nil, err
		}
		if same {
			out = append(out, rec)
		}
	}
	return out, nil
}

// ConfirmPlacedOrder 는 접수된 발주를 기록 번호로 한 번 읽어 확인함. nil 이면 확인됨.
//
// 판정: 읽기 성공(시한 roundTripTimeout) · 응답 번호가 기록 번호와 **바이트 일치**(orderId 는 불투명 식별자 — 정규화 없음) ·
// 응답 종목이 **비어 있지 않고** 발주 종목과 같음(정규화 ToUpper(TrimSpace) 뒤). 그 밖은 전부 확인 실패임.
func ConfirmPlacedOrder(ctx context.Context, orders OrderReader, brokerOrderID, symbol string) error {
	if orders == nil {
		return errNoOrderReader
	}
	rctx, cancel := context.WithTimeout(ctx, roundTripTimeout)
	defer cancel()
	raw, err := orders.OrderRaw(rctx, brokerOrderID)
	if err != nil {
		return fmt.Errorf("reading order %q back: %w", brokerOrderID, err)
	}
	return judgePlacedOrder(raw, brokerOrderID, symbol)
}

// judgePlacedOrder 는 읽어 온 응답의 판정만 함(순수 함수).
func judgePlacedOrder(raw json.RawMessage, brokerOrderID, symbol string) error {
	facts, err := parseOrderFacts(unwrapOrderEnvelope(raw))
	if err != nil {
		return fmt.Errorf("the read-back of order %q could not be read: %w", brokerOrderID, err)
	}
	// Byte-exact. An id that differs only in whitespace or case is a different id.
	if facts.OrderID != brokerOrderID {
		return fmt.Errorf("the broker acked order %q but the read-back names %q", brokerOrderID, facts.OrderID)
	}
	want := strings.ToUpper(strings.TrimSpace(symbol))
	// 응답 종목이 비면 확인 실패(a094 D−5.1) — 종전에는 비면 통과시켰음. 계약(openapi Order)은 symbol 을 필수로 둔다.
	if facts.Symbol == "" {
		return fmt.Errorf("the read-back of order %q names no symbol, so it cannot be tied to %s", brokerOrderID, want)
	}
	if want == "" || facts.Symbol != want {
		return fmt.Errorf(
			"order %q was acked for %s but the broker reports it on %s — the same identifier appears in a conflicting context",
			brokerOrderID, want, facts.Symbol)
	}
	return nil
}
