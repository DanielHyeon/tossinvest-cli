package verifylive

// pricing.go decides the one number that makes this tool safe to run: the price
// of an order that must not fill.
//
// Every live order this tool places is a limit order for one share at a price the
// market is not at. A buy goes far below the last trade, a sell far above it, and
// a stop trigger far below it. If that arithmetic is wrong the tool stops being a
// measurement and becomes a trade — so it is a small file with a table, tested
// against the boundaries, and it refuses rather than guesses.
//
// # The two constraints
//
//	the tick grid   KR prices must land on the exchange's tick, which is stepped
//	                by price band. "KR 은 호가 단위에 맞지 않으면 400 invalid-request"
//	                (OrderCreateRequest.price, docs/migration/openapi.latest.json).
//	the daily band  a KR order outside 상한가/하한가 is rejected. GET
//	                /api/v1/price-limits reports both; they are null for US, which
//	                has no daily band.
//
// Those pull in opposite directions. "Far from the market" wants a price as
// distant as possible; the daily band caps how far it can be. The functions here
// take the offset the operator asked for, snap it to the grid, clamp it into the
// band, and then check the result is still on the correct side of the last trade.
// If it is not — a stock already limit-down has no room below it — they return an
// error and the step is skipped rather than an order being placed at a price that
// could trade.

import (
	"fmt"
	"math"
	"strings"
)

// MinQuantity is the size of every order this tool places.
//
// One share. Not a flag: the whole justification for placing live orders at all
// is that the exposure is the smallest the exchange allows, and a knob that could
// raise it would be a knob that turns a verification into a position. Fractional
// quantities are not used either — they are US MARKET SELL only, and this tool
// places no market orders.
const MinQuantity = 1.0

// DefaultOffset is how far from the last trade an order is priced.
//
// 20%: comfortably outside any plausible intraday move, comfortably inside the
// KR ±30% daily band so the order is still accepted. Conservative here means
// *distant*, which is the opposite of what the word usually means about a price.
const DefaultOffset = 0.20

// MinOffset is the closest to the market this tool will price anything. Below
// this an ordinary day's volatility could reach the order.
const MinOffset = 0.05

// MaxOffset is the furthest. Past it the KR daily band clamps everything to the
// same price anyway, and the number stops describing what will happen.
const MaxOffset = 0.60

// The markets this tool's mutating steps can run against.
//
// A run carries one of them (Options.Market) and sends orders only for symbols in
// it: a symbol from the other market is skipped with a reason rather than probed
// with rules written for somewhere else. What differs between the two is the
// broker's own contract, not this tool's preference —
//
//	KR  a modify requires a quantity ("KR 주식: 필수"), and the daily price band
//	    (상한가/하한가) bounds how far from the market an order may be priced.
//	US  a modify must not carry a quantity ("US 주식: 전달 불가" →
//	    400 us-modify-quantity-not-supported), and there is no daily band —
//	    GET /api/v1/price-limits returns null, so the offset alone decides.
//
// Both citations are OrderModifyRequest / PriceLimitResponse in
// docs/migration/openapi.latest.json.
const (
	MarketKR = "KR"
	MarketUS = "US"
)

// NormalizeMarket resolves a market name. The zero value is KR, so a caller that
// says nothing gets the behaviour this tool had before it knew about US (§0.2).
func NormalizeMarket(market string) string {
	switch {
	case strings.EqualFold(strings.TrimSpace(market), MarketUS):
		return MarketUS
	default:
		return MarketKR
	}
}

// SameMarket compares two market names.
func SameMarket(a, b string) bool { return strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b)) }

// SafePrice is a price and the reasoning that produced it, so the record can say
// why an order was priced where it was.
type SafePrice struct {
	// Price is the limit or trigger price.
	Price float64
	// Basis explains it: the last trade, the offset applied, and whether the
	// daily band clamped the result.
	Basis string
	// Clamped reports that the daily band, not the offset, decided the price.
	Clamped bool
}

// ValidateOffset bounds the operator's offset.
func ValidateOffset(offset float64) (float64, error) {
	if offset == 0 {
		return DefaultOffset, nil
	}
	if offset < MinOffset || offset > MaxOffset {
		return 0, fmt.Errorf(
			"verify: an offset of %.0f%% is outside the accepted %.0f%%..%.0f%% — below the floor an ordinary "+
				"day's move could fill the order, above the ceiling the daily price band decides the price instead",
			offset*100, MinOffset*100, MaxOffset*100)
	}
	return offset, nil
}

// FarBuyLimit prices a buy that must not fill: below the market, on the tick
// grid, at or above the day's lower limit.
func FarBuyLimit(last, lowerLimit, offset float64, market string) (SafePrice, error) {
	if last <= 0 {
		return SafePrice{}, fmt.Errorf("verify: no last price is available, so no order can be priced safely")
	}
	offset, err := ValidateOffset(offset)
	if err != nil {
		return SafePrice{}, err
	}
	want := RoundDownToTick(last*(1-offset), market)
	price, clamped := want, false
	if lowerLimit > 0 && price < lowerLimit {
		price, clamped = RoundUpToTick(lowerLimit, market), true
	}
	if price <= 0 || price >= last {
		return SafePrice{}, fmt.Errorf(
			"verify: %s cannot be priced below the market and inside the day's band (last %.4f, lower limit %.4f) "+
				"— the symbol is probably already at its floor; pick another symbol",
			market, last, lowerLimit)
	}
	return SafePrice{Price: price, Basis: basis("buy", last, offset, price, clamped), Clamped: clamped}, nil
}

// FarSellLimit prices a sell that must not fill: above the market, on the tick
// grid, at or below the day's upper limit.
func FarSellLimit(last, upperLimit, offset float64, market string) (SafePrice, error) {
	if last <= 0 {
		return SafePrice{}, fmt.Errorf("verify: no last price is available, so no order can be priced safely")
	}
	offset, err := ValidateOffset(offset)
	if err != nil {
		return SafePrice{}, err
	}
	want := RoundUpToTick(last*(1+offset), market)
	price, clamped := want, false
	if upperLimit > 0 && price > upperLimit {
		price, clamped = RoundDownToTick(upperLimit, market), true
	}
	if price <= last {
		return SafePrice{}, fmt.Errorf(
			"verify: %s cannot be priced above the market and inside the day's band (last %.4f, upper limit %.4f) "+
				"— the symbol is probably already at its ceiling; pick another symbol",
			market, last, upperLimit)
	}
	return SafePrice{Price: price, Basis: basis("sell", last, offset, price, clamped), Clamped: clamped}, nil
}

// FarStopTrigger prices a stop that must not trigger: far below the market.
//
// A protective stop fires when the price falls to it, so "will not fire" and
// "will not fill" are the same arithmetic as a buy limit — which is why this
// delegates rather than repeating the clamp.
func FarStopTrigger(last, lowerLimit, offset float64, market string) (SafePrice, error) {
	p, err := FarBuyLimit(last, lowerLimit, offset, market)
	if err != nil {
		return SafePrice{}, err
	}
	p.Basis = strings.Replace(p.Basis, "buy at", "stop trigger at", 1)
	return p, nil
}

// OneTickFurther moves a price one tick away from the market, in the direction
// that keeps it un-fillable. It is what the amend and conditional-modify steps
// change, because the smallest change that the broker will accept as a change is
// the one that measures the amend and nothing else.
func OneTickFurther(price float64, side string, market string) float64 {
	tick := TickSize(price, market)
	if strings.EqualFold(side, "sell") {
		return RoundUpToTick(price+tick, market)
	}
	return RoundDownToTick(price-tick, market)
}

func basis(side string, last, offset, price float64, clamped bool) string {
	if clamped {
		return fmt.Sprintf("%s at %s — the day's price band, not the %.0f%% offset from the last trade %s",
			side, trim(price), offset*100, trim(last))
	}
	return fmt.Sprintf("%s at %s — %.0f%% from the last trade %s, snapped to the tick grid",
		side, trim(price), offset*100, trim(last))
}

// RoundDownToTick snaps a price down onto a valid tick.
func RoundDownToTick(price float64, market string) float64 {
	if price <= 0 {
		return 0
	}
	tick := TickSize(price, market)
	return round(math.Floor(price/tick) * tick)
}

// RoundUpToTick snaps a price up onto a valid tick.
func RoundUpToTick(price float64, market string) float64 {
	if price <= 0 {
		return 0
	}
	tick := TickSize(price, market)
	return round(math.Ceil(price/tick) * tick)
}

// TickSize is the exchange's minimum price increment at a price.
//
// The KR schedule is KRX's; it is the same table internal/flatten uses to price
// an emergency liquidation, and the openapi document's own example ("50,000~
// 200,000원 구간은 100원 단위") is the row it agrees with. US equities tick at a
// cent above a dollar and at a hundredth of a cent below it
// (OrderCreateRequest.price, docs/migration/openapi.latest.json).
func TickSize(price float64, market string) float64 {
	if !strings.EqualFold(market, MarketKR) {
		if price < 1 {
			return 0.0001
		}
		return 0.01
	}
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

// MarketOf classifies a symbol. KRX symbols are six digits; anything else is
// treated as US, which the mutating steps then decline to touch.
func MarketOf(symbol string) string {
	s := strings.TrimSpace(symbol)
	if len(s) != 6 {
		return "US"
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return "US"
		}
	}
	return MarketKR
}

// round removes the float noise a divide-and-multiply leaves behind, so a KR
// price is the integer the broker's pattern expects rather than 69999.999999999.
func round(v float64) float64 {
	return math.Round(v*1e6) / 1e6
}

func trim(v float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.4f", v), "0"), ".")
}
