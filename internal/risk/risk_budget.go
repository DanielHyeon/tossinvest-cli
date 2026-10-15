package risk

// risk_budget.go 는 건당 위험 예산의 유도 한 곳임(a139, risk-management 「Automation gate limits are the production Guardian policy source」).
//
// # 왜 1/3 인가
//
// 출처: 사용자 결정 2026-10-15(change a139) — 건당 위험 예산 = 일일 손실 한도 / 3. 그 전에는 예산이 일일 손실 한도 전부라서 손절에 걸린
// 거래 **하나**가 그날의 손실 한도를 다 쓸 수 있었음. 예산이 작아지면 위험 기반 수량 floor(예산 / 손절 폭)도 작아지므로 사이징을 줄이는
// 방향(§0.9)만임. 새 결정 없이 완화하지 않음.
//
// # 내림의 단위
//
// 1/3 은 유한 소수가 아니므로 반드시 자름. 자르는 단위는 그 통화의 최소 단위(KRW 1 · USD 0.01 — 저장소의 minor-v1 관례, continuationlane
// currencyMinorScale 과 같은 표)이고 방향은 언제나 내림(보수). 최소 단위를 모르는 통화 · 0 으로 내려가는 값은 오류 — 호출자(정책 유도)는
// 예산을 비워 Policy.Validate 가 기동을 거부하게 함.

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// PerTradeRiskBudgetDivisor 는 일일 손실 한도를 나누는 수(사용자 결정 2026-10-15, a139).
const PerTradeRiskBudgetDivisor = 3

// riskBudgetMinorScale 은 통화의 최소 단위 소수 자릿수 — KRW 0 · USD 2(minor-v1). 표에 없는 통화는 예산을 정할 수 없음.
var riskBudgetMinorScale = map[string]int{"KRW": 0, "USD": 2}

// PerTradeRiskBudget 은 일일 손실 한도에서 건당 위험 예산을 유도함: floor(일일 손실 / 3) 을 통화 최소 단위로.
//
// 일일 손실이 양수 decimal 이 아니거나, 통화의 최소 단위를 모르거나, 결과가 0 이면 오류.
func PerTradeRiskBudget(dailyLoss riskcalc.Money) (riskcalc.Money, error) {
	currency := strings.ToUpper(strings.TrimSpace(dailyLoss.Currency))
	scale, ok := riskBudgetMinorScale[currency]
	if !ok {
		return riskcalc.Money{}, fmt.Errorf("risk: per-trade risk budget: currency %q has no known minor unit", dailyLoss.Currency)
	}
	amount, err := parseDecimal("maximum daily loss", dailyLoss.Amount)
	if err != nil {
		return riskcalc.Money{}, fmt.Errorf("risk: per-trade risk budget: %w", err)
	}
	if amount.Sign() <= 0 {
		return riskcalc.Money{}, fmt.Errorf("risk: per-trade risk budget: maximum daily loss %q is not positive", dailyLoss.Amount)
	}
	// floor(amount / 3 × 10^scale) 를 최소 단위 정수로 — 음이 아닌 유리수의 floor 는 분자 ÷ 분모의 정수 나눗셈.
	unit := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale)), nil)
	minor := new(big.Rat).Mul(amount, new(big.Rat).SetFrac(unit, big.NewInt(PerTradeRiskBudgetDivisor)))
	floored := new(big.Int).Quo(minor.Num(), minor.Denom())
	if floored.Sign() <= 0 {
		return riskcalc.Money{}, fmt.Errorf("risk: per-trade risk budget: a third of %s %s rounds down to zero", dailyLoss.Amount, currency)
	}
	return riskcalc.Money{Amount: minorText(floored, scale), Currency: currency}, nil
}

// minorText 는 최소 단위 정수를 decimal 문자열로 씀(소수부 끝의 0 은 뺌 — 2000 cents = "20").
func minorText(minor *big.Int, scale int) string {
	digits := minor.String()
	if scale == 0 {
		return digits
	}
	for len(digits) <= scale {
		digits = "0" + digits
	}
	whole, frac := digits[:len(digits)-scale], strings.TrimRight(digits[len(digits)-scale:], "0")
	if frac == "" {
		return whole
	}
	return whole + "." + frac
}
