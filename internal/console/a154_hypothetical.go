package console

import (
	"math/big"
	"regexp"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
)

type hypotheticalInput struct{ Market, Entry, Low, ATR, Tick, Min, Max, Budget, MaxShares, MaxNotional string }
type hypotheticalArm struct{ Name, Stop, Distance, Percent, Quantity, Risk, Notional, State string }
type hypotheticalResult struct {
	Input    hypotheticalInput
	Currency string
	Errors   map[string]string
	Arms     []hypotheticalArm
}

var hypotheticalDecimal = regexp.MustCompile(`^[0-9]+(?:\.[0-9]{1,8})?$`)

func hypotheticalNumber(s string) (*big.Rat, bool) {
	if len(s) == 0 || len(s) > 32 || !hypotheticalDecimal.MatchString(s) {
		return nil, false
	}
	n, ok := new(big.Rat).SetString(s)
	return n, ok && n.Sign() > 0
}
func hypotheticalText(x *big.Rat) string {
	s := x.FloatString(8)
	return strings.TrimRight(strings.TrimRight(s, "0"), ".")
}
func hypotheticalFloor(x, tick *big.Rat) *big.Rat {
	q := new(big.Rat).Quo(x, tick)
	i := new(big.Int).Quo(q.Num(), q.Denom())
	if q.Sign() < 0 && new(big.Int).Rem(q.Num(), q.Denom()).Sign() != 0 {
		i.Sub(i, big.NewInt(1))
	}
	return new(big.Rat).Mul(new(big.Rat).SetInt(i), tick)
}
func calculateHypothetical(in hypotheticalInput) hypotheticalResult {
	out := hypotheticalResult{Input: in, Errors: map[string]string{}, Arms: []hypotheticalArm{{Name: "고정 0.7%"}, {Name: "신호봉 저가 아래 1틱"}, {Name: "ATR14 × 1.5"}}}
	if in.Market == "KR" {
		out.Currency = "KRW"
	} else if in.Market == "US" {
		out.Currency = "USD"
	} else {
		out.Errors["market"] = "시장을 선택하세요."
	}
	parse := func(key, s string) *big.Rat {
		if s == "" {
			return nil
		}
		n, ok := hypotheticalNumber(s)
		if !ok {
			out.Errors[key] = "양의 숫자를 입력하세요 (최대 32자, 소수 8자리)."
			return nil
		}
		return n
	}
	entry, tick, min, max := parse("entry", in.Entry), parse("tick", in.Tick), parse("min", in.Min), parse("max", in.Max)
	low, atr := parse("low", in.Low), parse("atr", in.ATR)
	budget, shares, notional := parse("budget", in.Budget), parse("max_shares", in.MaxShares), parse("max_notional", in.MaxNotional)
	if shares != nil && (strings.Contains(in.MaxShares, ".") || !shares.IsInt() || shares.Cmp(big.NewRat(1000000, 1)) > 0) {
		out.Errors["max_shares"] = "100만 이하의 정수를 입력하세요."
		shares = nil
	}
	common := entry != nil && tick != nil && min != nil && max != nil && out.Currency != ""
	if common && (min.Cmp(max) > 0 || entry.Cmp(min) < 0 || entry.Cmp(max) > 0 || hypotheticalFloor(entry, tick).Cmp(entry) != 0) {
		out.Errors["entry"] = "진입가와 가격 범위 또는 호가 간격을 확인하세요."
		common = false
	}
	if low != nil && entry != nil && low.Cmp(entry) > 0 {
		out.Errors["low"] = "신호봉 저가는 진입가 이하이어야 합니다."
		low = nil
	}
	for i := range out.Arms {
		a := &out.Arms[i]
		a.State = "입력 부족"
		if !common {
			if out.Errors["market"] != "" || out.Errors["entry"] != "" || out.Errors["tick"] != "" || out.Errors["min"] != "" || out.Errors["max"] != "" {
				a.State = "입력 오류"
			}
			continue
		}
		var raw *big.Rat
		switch i {
		case 0:
			raw = new(big.Rat).Mul(entry, big.NewRat(993, 1000))
		case 1:
			if low == nil {
				if out.Errors["low"] != "" {
					a.State = "입력 오류"
				}
				continue
			}
			raw = new(big.Rat).Sub(hypotheticalFloor(low, tick), tick)
		case 2:
			if atr == nil {
				if out.Errors["atr"] != "" {
					a.State = "입력 오류"
				}
				continue
			}
			raw = new(big.Rat).Sub(entry, new(big.Rat).Mul(atr, big.NewRat(3, 2)))
		}
		stop := hypotheticalFloor(raw, tick)
		if stop.Sign() <= 0 || stop.Cmp(entry) >= 0 || stop.Cmp(min) < 0 || stop.Cmp(max) > 0 {
			a.State = "가격 범위 또는 손절가 오류"
			continue
		}
		dist := new(big.Rat).Sub(entry, stop)
		a.Stop = hypotheticalText(stop)
		a.Distance = hypotheticalText(dist)
		pct := new(big.Rat).Quo(new(big.Rat).Mul(dist, big.NewRat(100, 1)), entry)
		if pct.Cmp(big.NewRat(1, 1000000)) < 0 {
			a.Percent = "<0.000001%"
		} else {
			a.Percent = pct.FloatString(6) + "%"
		}
		a.State = "계산됨"
		if budget == nil || shares == nil || notional == nil {
			continue
		}
		qstr, err := risk.RiskBasedQuantity(in.Budget, in.Entry, a.Stop)
		if err != nil {
			a.State = "수량 계산 오류"
			continue
		}
		q, _ := new(big.Int).SetString(qstr, 10)
		cap := new(big.Rat).Quo(notional, entry)
		nc := new(big.Int).Quo(cap.Num(), cap.Denom())
		if q.Cmp(shares.Num()) > 0 {
			q.Set(shares.Num())
		}
		if q.Cmp(nc) > 0 {
			q.Set(nc)
		}
		qty := new(big.Rat).SetInt(q)
		if q.Sign() == 0 {
			a.State = "매수 가능 수량 없음"
		}
		a.Quantity = q.String()
		a.Risk = hypotheticalText(new(big.Rat).Mul(qty, dist))
		a.Notional = hypotheticalText(new(big.Rat).Mul(qty, entry))
	}
	return out
}
