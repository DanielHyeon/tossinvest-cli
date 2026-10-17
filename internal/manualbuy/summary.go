// Package manualbuy 는 공식 주문 조회 원문에서 엔진 밖 수동 매수(토스 앱 등)의 오늘 합계를 셈(a136 D9=b).
//
// 순수 함수만 둠 — 브로커 읽기는 엔진 보조 실행자(app/engine)의 일이고, 판정은 journal 예약 tx 의 일임.
// 모르는 값은 0 으로 두지 않음: 금액을 정할 수 없는 행이 하나라도 있으면 합계 전체가 산정 불가(오류)이고, 그 하루의
// 매수는 거부됨(fail-closed). 매도 · 엔진 자신의 주문 · 다른 하루의 주문은 세지 않음.
package manualbuy

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// ErrUnvaluable 는 금액을 정할 수 없는 주문이 있어 합계를 낼 수 없다는 뜻임.
var ErrUnvaluable = errors.New("manualbuy: a buy order outside the engine cannot be valued")

// order 는 GET /api/v1/orders 한 행 중 이 계산이 읽는 칸(official.apiOrder 와 같은 철자).
type order struct {
	OrderID   string `json:"orderId"`
	Side      string `json:"side"`
	Status    string `json:"status"`
	Quantity  string `json:"quantity"`
	Price     string `json:"price"`
	Currency  string `json:"currency"`
	OrderedAt string `json:"orderedAt"`
	// OrderAmount 는 금액 기반 US 시장가 매수에만 있음(OpenAPI Order.orderAmount).
	OrderAmount string `json:"orderAmount"`
	Execution   *struct {
		FilledQuantity     string `json:"filledQuantity"`
		AverageFilledPrice string `json:"averageFilledPrice"`
		FilledAmount       string `json:"filledAmount"`
	} `json:"execution"`
}

// Summarize 는 주문 원문들에서 계좌 하루 day 의 수동 매수 금액을 통화별로 더함.
//
// 규칙:
//   - side 가 BUY 가 아니면 셈에서 뺌. orderId 가 engineOrders 에 있으면 뺌(예약 원장이 이미 셈).
//   - 귀속은 주문 시각(orderedAt)의 clock.AccountTradingDay — 엔진 매수가 발급일에 귀속되는 것과 같은 규칙.
//   - 금액 = 체결액(filledAmount, 없으면 filledQuantity × averageFilledPrice) + 미체결 잔량 × 지정가. 잔량은 세부 상태
//     (`orders[].status` — OpenAPI OrderStatus 열거, 그룹 라벨 OPEN/CLOSED 가 아님)가 종결이면 0, 그 밖(모르는 값 포함)이면 셈.
//     지정가 없는 미체결은 금액 기반 시장가(orderAmount)면 max(orderAmount, 체결액), 아니면 산정 불가.
//   - 한 행이라도 orderId · 시각 · 통화 · 수량 · 가격을 해석할 수 없으면 ErrUnvaluable(지정가 없는 미체결 시장가 주문 포함).
//
// 같은 주문 id 가 두 번 오면(OPEN · CLOSED 조회가 겹침) 한 번만 셈 — 두 금액 중 큰 쪽(과대 집계 방향)을 씀.
func Summarize(raws []json.RawMessage, engineOrders map[string]bool, day string) (map[string]string, error) {
	perOrder := map[string]*big.Rat{}
	currencyOf := map[string]string{}
	for i, raw := range raws {
		var o order
		if err := json.Unmarshal(raw, &o); err != nil {
			return nil, fmt.Errorf("%w: row %d is not an order: %v", ErrUnvaluable, i, err)
		}
		if !strings.EqualFold(strings.TrimSpace(o.Side), "BUY") {
			continue
		}
		id := o.OrderID
		if strings.TrimSpace(id) == "" {
			return nil, fmt.Errorf("%w: row %d is a buy with no orderId, so it cannot be told from the engine's own", ErrUnvaluable, i)
		}
		if engineOrders[id] {
			continue
		}
		at, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(o.OrderedAt))
		if err != nil {
			return nil, fmt.Errorf("%w: order %s orderedAt %q: %v", ErrUnvaluable, id, o.OrderedAt, err)
		}
		label, err := clock.AccountTradingDay(at)
		if err != nil {
			return nil, fmt.Errorf("%w: order %s account day: %v", ErrUnvaluable, id, err)
		}
		if label != day {
			continue
		}
		currency := strings.ToUpper(strings.TrimSpace(o.Currency))
		if currency != "KRW" && currency != "USD" {
			return nil, fmt.Errorf("%w: order %s currency %q", ErrUnvaluable, id, o.Currency)
		}
		amount, err := valueOf(o)
		if err != nil {
			return nil, fmt.Errorf("%w: order %s: %v", ErrUnvaluable, id, err)
		}
		if prev, seen := perOrder[id]; seen {
			if currencyOf[id] != currency {
				return nil, fmt.Errorf("%w: order %s is reported in both %s and %s", ErrUnvaluable, id, currencyOf[id], currency)
			}
			if prev.Cmp(amount) >= 0 {
				continue
			}
		}
		perOrder[id], currencyOf[id] = amount, currency
	}
	totals := map[string]*big.Rat{}
	for id, amount := range perOrder {
		c := currencyOf[id]
		if totals[c] == nil {
			totals[c] = new(big.Rat)
		}
		totals[c].Add(totals[c], amount)
	}
	out := map[string]string{}
	for c, total := range totals {
		out[c] = decimalString(total)
	}
	return out, nil
}

// terminalStatuses 는 잔량이 더 체결되지 않는 세부 상태(OpenAPI OrderStatus). CANCEL_REJECTED · REPLACE_REJECTED 는 거부된 취소 · 정정의
// 별도 레코드이고 원주문은 이전 상태로 돌아가 따로 나오므로 이 레코드 자체는 새 매수가 아님(체결분만, 보통 0).
// PARTIAL_FILLED 는 CLOSED 그룹에도 나오지만 잔량이 아직 살아 있을 수 있으므로 종결로 보지 않음(과대 집계 방향).
var terminalStatuses = map[string]bool{
	"FILLED": true, "CANCELED": true, "REJECTED": true, "REPLACED": true,
	"CANCEL_REJECTED": true, "REPLACE_REJECTED": true,
}

// valueOf 는 한 매수 주문이 오늘 쓴(또는 쓸) 금액임.
func valueOf(o order) (*big.Rat, error) {
	status := strings.ToUpper(strings.TrimSpace(o.Status))
	filled := new(big.Rat)
	value := new(big.Rat)
	var err error
	if o.Execution != nil {
		if filled, err = decimal("execution.filledQuantity", o.Execution.FilledQuantity, true); err != nil {
			return nil, err
		}
		switch {
		case strings.TrimSpace(o.Execution.FilledAmount) != "":
			if value, err = decimal("execution.filledAmount", o.Execution.FilledAmount, false); err != nil {
				return nil, err
			}
		case filled.Sign() > 0:
			price, err := decimal("execution.averageFilledPrice", o.Execution.AverageFilledPrice, false)
			if err != nil {
				return nil, err
			}
			value = new(big.Rat).Mul(filled, price)
		}
	}
	if terminalStatuses[status] {
		return value, nil
	}
	// 미체결이 남을 수 있는 주문. 금액 기반 시장가(orderAmount)는 수량이 아니라 금액으로 셈 — 수량이 0 · 빈 값이어도 과소 집계 안 함.
	if strings.TrimSpace(o.OrderAmount) != "" {
		amount, err := decimal("orderAmount", o.OrderAmount, false)
		if err != nil {
			return nil, err
		}
		if amount.Cmp(value) > 0 {
			return amount, nil
		}
		return value, nil
	}
	quantity, err := decimal("quantity", o.Quantity, false)
	if err != nil {
		return nil, err
	}
	remaining := new(big.Rat).Sub(quantity, filled)
	if remaining.Sign() > 0 {
		price, err := decimal("price", o.Price, false)
		if err != nil {
			return nil, fmt.Errorf("a pending buy without a limit price or order amount cannot be valued: %w", err)
		}
		value.Add(value, new(big.Rat).Mul(remaining, price))
	}
	return value, nil
}

// decimal 은 브로커의 decimal 문자열을 정확히 읽음. 빈 값은 allowEmpty 일 때만 0(체결 전의 filledQuantity), 음수 · 지수 표기는 거부.
func decimal(field, raw string, allowEmpty bool) (*big.Rat, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		if allowEmpty {
			return new(big.Rat), nil
		}
		return nil, fmt.Errorf("%s is empty", field)
	}
	for _, r := range s {
		if (r < '0' || r > '9') && r != '.' {
			return nil, fmt.Errorf("%s %q is not a plain non-negative decimal", field, raw)
		}
	}
	value, ok := new(big.Rat).SetString(s)
	if !ok || strings.Count(s, ".") > 1 {
		return nil, fmt.Errorf("%s %q is not a decimal", field, raw)
	}
	return value, nil
}

// decimalString 은 유한 소수 Rat 를 정확한 decimal 문자열로 씀(유한 소수의 곱 · 합이므로 분모는 2·5 의 거듭제곱).
func decimalString(r *big.Rat) string {
	for places := 0; places <= 30; places++ {
		s := r.FloatString(places)
		back, ok := new(big.Rat).SetString(s)
		if ok && back.Cmp(r) == 0 {
			return s
		}
	}
	// 30 자리로도 정확하지 않으면 올림 방향(과대 집계)으로 자름.
	scale := new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(30), nil))
	scaled := new(big.Rat).Mul(r, scale)
	ceil := new(big.Int).Quo(scaled.Num(), scaled.Denom())
	if new(big.Rat).SetInt(ceil).Cmp(scaled) < 0 {
		ceil.Add(ceil, big.NewInt(1))
	}
	return new(big.Rat).Quo(new(big.Rat).SetInt(ceil), scale).FloatString(30)
}
