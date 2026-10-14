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
		`{"orderId":"m1","side":"BUY","status":"CLOSED","quantity":"20","price":"70000","currency":"KRW","orderedAt":"2026-10-14T01:00:00Z","execution":{"filledQuantity":"20","averageFilledPrice":"70000","filledAmount":"1400000"}}`,
		// KR 부분 체결 + 미체결(OPEN): 체결 3 × 10,000(체결액 없음) + 잔량 7 × 10,100 = 100,700
		`{"orderId":"m2","side":"BUY","status":"OPEN","quantity":"10","price":"10100","currency":"KRW","orderedAt":"2026-10-14T02:00:00Z","execution":{"filledQuantity":"3","averageFilledPrice":"10000"}}`,
		// US 미체결: 2 × 150.25 = 300.5 (뉴욕 날짜 10-14 정규장 = KST 10-14 22:30 이후)
		`{"orderId":"m3","side":"BUY","status":"OPEN","quantity":"2","price":"150.25","currency":"USD","orderedAt":"2026-10-14T14:00:00Z","execution":null}`,
		// 취소된 무체결 매수: 0
		`{"orderId":"m4","side":"BUY","status":"CLOSED","quantity":"5","price":"1000","currency":"KRW","orderedAt":"2026-10-14T03:00:00Z","execution":{"filledQuantity":"0"}}`,
		// 매도 · 엔진 주문 · 다른 하루(KST 10-14 07:59)는 셈에서 빠짐
		`{"orderId":"s1","side":"SELL","status":"CLOSED","quantity":"1","price":"1","currency":"KRW","orderedAt":"2026-10-14T01:00:00Z"}`,
		`{"orderId":"e1","side":"BUY","status":"OPEN","quantity":"100","price":"100000","currency":"KRW","orderedAt":"2026-10-14T01:00:00Z"}`,
		`{"orderId":"old","side":"BUY","status":"CLOSED","quantity":"1","price":"1","currency":"KRW","orderedAt":"2026-10-13T22:59:00Z","execution":{"filledQuantity":"1","filledAmount":"999999"}}`,
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
		"pending market order": `{"orderId":"x","side":"BUY","status":"OPEN","quantity":"1","price":"","currency":"KRW","orderedAt":"2026-10-14T01:00:00Z"}`,
		"no order id":          `{"orderId":"","side":"BUY","status":"OPEN","quantity":"1","price":"1","currency":"KRW","orderedAt":"2026-10-14T01:00:00Z"}`,
		"no ordered at":        `{"orderId":"x","side":"BUY","status":"OPEN","quantity":"1","price":"1","currency":"KRW","orderedAt":""}`,
		"unknown currency":     `{"orderId":"x","side":"BUY","status":"OPEN","quantity":"1","price":"1","currency":"JPY","orderedAt":"2026-10-14T01:00:00Z"}`,
		"exponent quantity":    `{"orderId":"x","side":"BUY","status":"OPEN","quantity":"1e3","price":"1","currency":"KRW","orderedAt":"2026-10-14T01:00:00Z"}`,
		"filled without price": `{"orderId":"x","side":"BUY","status":"CLOSED","quantity":"1","price":"1","currency":"KRW","orderedAt":"2026-10-14T01:00:00Z","execution":{"filledQuantity":"1"}}`,
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
		`{"orderId":"d","side":"BUY","status":"OPEN","quantity":"10","price":"100","currency":"KRW","orderedAt":"2026-10-14T01:00:00Z","execution":{"filledQuantity":"0"}}`,
		`{"orderId":"d","side":"BUY","status":"CLOSED","quantity":"10","price":"100","currency":"KRW","orderedAt":"2026-10-14T01:00:00Z","execution":{"filledQuantity":"4","filledAmount":"400"}}`,
	), nil, day)
	if err != nil {
		t.Fatal(err)
	}
	if got["KRW"] != "1000" {
		t.Fatalf("the same order must count once at its larger amount, got %v", got)
	}
}
