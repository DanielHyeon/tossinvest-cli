package main

// sizing.go 는 그림자 측정의 가상 사이징 — 어떤 생산 값도 아님(design §4 · §2.1-8).

import (
	"errors"
	"math"
	"math/big"

	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
)

const ppmScale = 1_000_000

// krTickMinor 는 KRX 호가 단위(원). verifylive.TickSize · flatten.krTick 과 같은 표 — 시험이 경계마다 대조함.
// 정수로 다시 적은 이유: breakout 설정의 TickMinor 는 정수 minor 이고 이 도구는 float 가격을 쓰지 않음.
func krTickMinor(price uint64) uint64 {
	switch {
	case price < 2000:
		return 1
	case price < 5000:
		return 5
	case price < 20000:
		return 10
	case price < 50000:
		return 50
	case price < 200000:
		return 100
	case price < 500000:
		return 500
	default:
		return 1000
	}
}

// costRatePPM 은 KR 왕복 비용 요율(ppm) = 매수 수수료 + 매도측 요율(수수료 · 거래세 · 환전)의 요율별 올림 합.
// 요율마다 올리는 이유: 0.003×1e6 같은 값이 부동소수 잡음으로 3000.0000000000005 가 되어 1 ppm 이 더 붙지 않게 —
// 1e-6 ppm 아래의 잡음은 버리고 그 위의 소수는 올림(과대 추정 쪽).
func costRatePPM(model costs.Model) (uint64, error) {
	buy, err := model.BuyCommissionRate(costs.MarketKR)
	if err != nil {
		return 0, err
	}
	sell, err := model.SellCommissionRate(costs.MarketKR)
	if err != nil {
		return 0, err
	}
	tax, err := model.SellTaxRate(costs.MarketKR)
	if err != nil {
		return 0, err
	}
	fx, err := model.FXConversionFeeRate(costs.MarketKR)
	if err != nil {
		return 0, err
	}
	total := uint64(0)
	for _, rate := range []float64{buy, sell, tax, fx} {
		if math.IsNaN(rate) || math.IsInf(rate, 0) || rate < 0 || rate > 1 {
			return 0, errors.New("a cost rate is not a fraction between 0 and 1")
		}
		total += uint64(math.Ceil(rate*ppmScale - 1e-6))
	}
	return total, nil
}

// sizingPlan 은 가상 진입 하나의 숫자 — 원장 행 · 라벨러가 같은 값을 씀.
type sizingPlan struct {
	Entry, Ask, Worst, Stop, RoundTrip, Risk, Target uint64
}

// planSizing 은 평가기 size() 와 같은 식으로 위험을 세고, 비용 포함 순 RR = rrPPM 이 되는 목표를 거꾸로 계산함.
//
//	worst = max(진입 제안가, ask) · 왕복 = ceil(worst × 요율 / 1e6) · 위험 = worst − 손절 + 왕복
//	목표 = worst + 왕복 + ceil(위험 × rr / 1e6)
//
// 평가기의 RR 검사(netReward ≥ ceil(위험 × rr))와 같은 올림이라 그 검사는 구성상 통과함 — 보고서가 그렇게 밝힘.
func planSizing(entry, ask, stop, costPPM, rrPPM uint64) (sizingPlan, error) {
	worst := max(entry, ask)
	if entry == 0 || stop == 0 || stop >= worst || stop >= entry {
		return sizingPlan{}, errors.New("the stop is not below the entry")
	}
	roundTrip, ok := mulDivCeil(worst, costPPM, ppmScale)
	if !ok {
		return sizingPlan{}, errors.New("the round-trip cost overflows")
	}
	risk := worst - stop + roundTrip
	reward, ok := mulDivCeil(risk, rrPPM, ppmScale)
	if !ok || worst > math.MaxUint64-roundTrip-reward {
		return sizingPlan{}, errors.New("the target overflows")
	}
	return sizingPlan{Entry: entry, Ask: ask, Worst: worst, Stop: stop, RoundTrip: roundTrip, Risk: risk,
		Target: worst + roundTrip + reward}, nil
}

// rMultiplePPM 은 floor((exit − worst − 왕복) × 1e6 / 위험), 음수는 −∞ 쪽 내림. 위험 0 이면 0.
func rMultiplePPM(exit uint64, plan sizingPlan) int64 {
	if plan.Risk == 0 {
		return 0
	}
	net := new(big.Int).SetUint64(exit)
	net.Sub(net, new(big.Int).SetUint64(plan.Worst))
	net.Sub(net, new(big.Int).SetUint64(plan.RoundTrip))
	net.Mul(net, big.NewInt(ppmScale))
	q := new(big.Int)
	m := new(big.Int)
	q.DivMod(net, new(big.Int).SetUint64(plan.Risk), m) // DivMod 은 유클리드 나눗셈 — 양의 제수에서 −∞ 쪽 내림과 같음
	if !q.IsInt64() {
		if q.Sign() > 0 {
			return math.MaxInt64
		}
		return math.MinInt64
	}
	return q.Int64()
}

// mulDivCeil 은 ceil(a × b / d). 결과가 uint64 를 넘으면 ok = false.
func mulDivCeil(a, b, d uint64) (uint64, bool) {
	if d == 0 {
		return 0, false
	}
	n := new(big.Int).Mul(new(big.Int).SetUint64(a), new(big.Int).SetUint64(b))
	q, r := new(big.Int).QuoRem(n, new(big.Int).SetUint64(d), new(big.Int))
	if r.Sign() != 0 {
		q.Add(q, big.NewInt(1))
	}
	if !q.IsUint64() {
		return 0, false
	}
	return q.Uint64(), true
}
