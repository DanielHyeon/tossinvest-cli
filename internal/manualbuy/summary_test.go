package manualbuy

import (
	"encoding/json"
	"errors"
	"testing"
)

// a136 D9=b: 수동 매수 합계 — 매수만, 엔진 주문 제외, 계좌 하루 귀속, 체결액 + 미체결 잔량 × 지정가, 모르는 값은 산정 불가.

func rows(t *testing.T, objs ...string) []json.RawMessage {
	t.Helper()
	out := make([]json.RawMessage, 0, len(objs))
	for _, o := range objs {
		out = append(out, json.RawMessage(o))
	}
	return out
}

const day = "2026-10-14" // KST 2026-10-14 08:00 ~ 2026-10-15 08:00

func TestA136ManualBuysAreSummedPerCurrency(t *testing.T) {
	got, err := Summarize(rows(t,
		// KR 체결 완료(체결액 있음) 1,400,000
		`{"orderId":"m1","side":"BUY","status":"FILLED","quantity":"20","price":"70000","currency":"KRW","orderedAt":"2026-10-14T01:00:00Z","execution":{"filledQuantity":"20","averageFilledPrice":"70000","filledAmount":"1400000"}}`,
		// KR 부분 체결 + 미체결(OPEN): 체결 3 × 10,000(체결액 없음) + 잔량 7 × 10,100 = 100,700
		`{"orderId":"m2","side":"BUY","status":"PARTIAL_FILLED","quantity":"10","price":"10100","currency":"KRW","orderedAt":"2026-10-14T02:00:00Z","execution":{"filledQuantity":"3","averageFilledPrice":"10000"}}`,
		// US 미체결: 2 × 150.25 = 300.5 (뉴욕 날짜 10-14 정규장 = KST 10-14 22:30 이후)
		`{"orderId":"m3","side":"BUY","status":"PENDING","quantity":"2","price":"150.25","currency":"USD","orderedAt":"2026-10-14T14:00:00Z","execution":null}`,
		// 취소된 무체결 매수: 0
		`{"orderId":"m4","side":"BUY","status":"CANCELED","quantity":"5","price":"1000","currency":"KRW","orderedAt":"2026-10-14T03:00:00Z","execution":{"filledQuantity":"0"}}`,
		// 매도 · 엔진 주문 · 다른 하루(KST 10-14 07:59)는 셈에서 빠짐
		`{"orderId":"s1","side":"SELL","status":"FILLED","quantity":"1","price":"1","currency":"KRW","orderedAt":"2026-10-14T01:00:00Z"}`,
		`{"orderId":"e1","side":"BUY","status":"PENDING","quantity":"100","price":"100000","currency":"KRW","orderedAt":"2026-10-14T01:00:00Z"}`,
		`{"orderId":"old","side":"BUY","status":"FILLED","quantity":"1","price":"1","currency":"KRW","orderedAt":"2026-10-13T22:59:00Z","execution":{"filledQuantity":"1","filledAmount":"999999"}}`,
	), map[string]bool{"e1": true}, day)
	if err != nil {
		t.Fatal(err)
	}
	if got["KRW"] != "1500700" || got["USD"] != "300.5" || len(got) != 2 {
		t.Fatalf("totals = %v, want KRW 1500700 and USD 300.5", got)
	}
}

func TestA136AnUnvaluableManualBuyFailsTheWholeSnapshot(t *testing.T) {
	for name, row := range map[string]string{
		"pending market order": `{"orderId":"x","side":"BUY","status":"PENDING","quantity":"1","price":"","currency":"KRW","orderedAt":"2026-10-14T01:00:00Z"}`,
		"no order id":          `{"orderId":"","side":"BUY","status":"PENDING","quantity":"1","price":"1","currency":"KRW","orderedAt":"2026-10-14T01:00:00Z"}`,
		"no ordered at":        `{"orderId":"x","side":"BUY","status":"PENDING","quantity":"1","price":"1","currency":"KRW","orderedAt":""}`,
		"unknown currency":     `{"orderId":"x","side":"BUY","status":"PENDING","quantity":"1","price":"1","currency":"JPY","orderedAt":"2026-10-14T01:00:00Z"}`,
		"exponent quantity":    `{"orderId":"x","side":"BUY","status":"PENDING","quantity":"1e3","price":"1","currency":"KRW","orderedAt":"2026-10-14T01:00:00Z"}`,
		"filled without price": `{"orderId":"x","side":"BUY","status":"FILLED","quantity":"1","price":"1","currency":"KRW","orderedAt":"2026-10-14T01:00:00Z","execution":{"filledQuantity":"1"}}`,
		"not json":             `{"orderId":`,
	} {
		raw := []json.RawMessage{json.RawMessage(row)}
		if name == "not json" {
			raw = []json.RawMessage{json.RawMessage(`"string"`)}
		}
		if _, err := Summarize(raw, nil, day); !errors.Is(err, ErrUnvaluable) {
			t.Errorf("%s: want ErrUnvaluable, got %v", name, err)
		}
	}
}

func TestA136ADuplicatedOrderCountsOnceAtTheLargerAmount(t *testing.T) {
	got, err := Summarize(rows(t,
		`{"orderId":"d","side":"BUY","status":"PENDING","quantity":"10","price":"100","currency":"KRW","orderedAt":"2026-10-14T01:00:00Z","execution":{"filledQuantity":"0"}}`,
		`{"orderId":"d","side":"BUY","status":"CANCELED","quantity":"10","price":"100","currency":"KRW","orderedAt":"2026-10-14T01:00:00Z","execution":{"filledQuantity":"4","filledAmount":"400"}}`,
	), nil, day)
	if err != nil {
		t.Fatal(err)
	}
	if got["KRW"] != "1000" {
		t.Fatalf("the same order must count once at its larger amount, got %v", got)
	}
}

// 구현 리뷰 P1-2: 세부 상태는 OpenAPI OrderStatus 열거 — 종결(취소 · 거절 · 대체 · 거부 레코드)은 잔량을 세지 않고, 지정가 없는 거절 시장가도
// 산정 불가로 만들지 않음. 모르는 상태는 미체결처럼 잔량을 셈(과대 집계).
func TestA136TerminalStatusesCountOnlyWhatFilled(t *testing.T) {
	got, err := Summarize(rows(t,
		`{"orderId":"c","side":"BUY","status":"CANCELED","quantity":"5","price":"1000","currency":"KRW","orderedAt":"2026-10-14T01:00:00Z","execution":{"filledQuantity":"0"}}`,
		`{"orderId":"r","side":"BUY","status":"REJECTED","quantity":"10","price":"","currency":"KRW","orderedAt":"2026-10-14T01:00:00Z","execution":{"filledQuantity":"0"}}`,
		`{"orderId":"p","side":"BUY","status":"REPLACED","quantity":"10","price":"100","currency":"KRW","orderedAt":"2026-10-14T01:00:00Z","execution":{"filledQuantity":"2","filledAmount":"200"}}`,
		`{"orderId":"cr","side":"BUY","status":"CANCEL_REJECTED","quantity":"10","price":"100","currency":"KRW","orderedAt":"2026-10-14T01:00:00Z"}`,
		`{"orderId":"u","side":"BUY","status":"SOMETHING_NEW","quantity":"3","price":"7","currency":"KRW","orderedAt":"2026-10-14T01:00:00Z"}`,
	), nil, day)
	if err != nil {
		t.Fatalf("terminal statuses must not make the snapshot unvaluable: %v", err)
	}
	// 취소 0 + 거절 시장가 0 + 대체 원주문의 체결분 200 + 거부 레코드 0 + 모르는 상태 3 × 7 = 221
	if got["KRW"] != "221" {
		t.Fatalf("totals = %v, want KRW 221", got)
	}
}

// 구현 리뷰 P2-1: 금액 기반 US 시장가 매수는 orderAmount 로 셈 — 수량이 0 이어도 과소 집계하지 않음.
func TestA136AnAmountBasedMarketBuyCountsItsOrderAmount(t *testing.T) {
	got, err := Summarize(rows(t,
		`{"orderId":"a","side":"BUY","status":"PENDING","orderType":"MARKET","quantity":"0","price":null,"orderAmount":"250.5","currency":"USD","orderedAt":"2026-10-14T14:00:00Z"}`,
		`{"orderId":"b","side":"BUY","status":"PARTIAL_FILLED","orderType":"MARKET","quantity":"0","orderAmount":"100","currency":"USD","orderedAt":"2026-10-14T14:00:00Z","execution":{"filledQuantity":"1","filledAmount":"120"}}`,
	), nil, day)
	if err != nil {
		t.Fatal(err)
	}
	if got["USD"] != "370.5" {
		t.Fatalf("totals = %v, want USD 370.5 (250.5 + max(100, 120))", got)
	}
}
