package verifylive

// pricing_test.go is where "the order cannot fill" is actually proven.
//
// Everything else in this package is procedure; this file is the arithmetic that
// makes the procedure safe. The cases are the ones that would produce a fillable
// order: a price on the wrong side of the market, a price the daily band pushes
// back across it, and a tick that rounds the wrong way.

import (
	"math"
	"testing"
)

func TestFarBuyLimitIsBelowTheMarketAndOnTheTick(t *testing.T) {
	got, err := FarBuyLimit(70000, 49000, DefaultOffset, MarketKR)
	if err != nil {
		t.Fatalf("FarBuyLimit: %v", err)
	}
	if got.Price >= 70000 {
		t.Fatalf("price %v is not below the last trade", got.Price)
	}
	if math.Mod(got.Price, TickSize(got.Price, MarketKR)) != 0 {
		t.Errorf("price %v is not on the tick grid (tick %v)", got.Price, TickSize(got.Price, MarketKR))
	}
	if got.Price != 56000 {
		t.Errorf("price = %v, want 56000 (70000 less 20%%, snapped down to the 100원 tick)", got.Price)
	}
	if got.Clamped {
		t.Error("the price was reported clamped although the band was nowhere near")
	}
}

func TestFarBuyLimitIsClampedIntoTheDailyBand(t *testing.T) {
	// A 20% discount lands below the day's floor; the floor wins, and the record
	// has to say so or the "20% away" claim would be a lie.
	got, err := FarBuyLimit(70000, 65000, DefaultOffset, MarketKR)
	if err != nil {
		t.Fatalf("FarBuyLimit: %v", err)
	}
	if got.Price < 65000 {
		t.Errorf("price %v is below the day's lower limit 65000 and would be rejected", got.Price)
	}
	if !got.Clamped {
		t.Error("the price was decided by the band and the basis does not say so")
	}
}

// TestFarBuyLimitRefusesWhenThereIsNoRoom. A stock already at its floor cannot be
// given an un-fillable buy price, and the honest answer is to refuse rather than
// place something at the last trade.
func TestFarBuyLimitRefusesWhenThereIsNoRoom(t *testing.T) {
	if _, err := FarBuyLimit(49000, 49000, DefaultOffset, MarketKR); err == nil {
		t.Fatal("a symbol pinned at its lower limit produced a 'safe' buy price")
	}
	if _, err := FarBuyLimit(0, 0, DefaultOffset, MarketKR); err == nil {
		t.Fatal("a missing last price produced a price")
	}
}

func TestFarSellLimitIsAboveTheMarketAndInsideTheBand(t *testing.T) {
	got, err := FarSellLimit(70000, 91000, DefaultOffset, MarketKR)
	if err != nil {
		t.Fatalf("FarSellLimit: %v", err)
	}
	if got.Price <= 70000 {
		t.Fatalf("price %v is not above the last trade", got.Price)
	}
	if got.Price > 91000 {
		t.Errorf("price %v is above the day's upper limit", got.Price)
	}

	clamped, err := FarSellLimit(80000, 82000, DefaultOffset, MarketKR)
	if err != nil {
		t.Fatalf("FarSellLimit: %v", err)
	}
	if clamped.Price > 82000 || !clamped.Clamped {
		t.Errorf("price = %+v, want it clamped to the 82000 ceiling", clamped)
	}
}

func TestFarStopTriggerIsFarBelow(t *testing.T) {
	got, err := FarStopTrigger(70000, 49000, DefaultOffset, MarketKR)
	if err != nil {
		t.Fatalf("FarStopTrigger: %v", err)
	}
	if got.Price >= 70000 {
		t.Fatalf("a stop at %v would be triggered by the current price", got.Price)
	}
}

func TestOffsetBounds(t *testing.T) {
	if got, err := ValidateOffset(0); err != nil || got != DefaultOffset {
		t.Errorf("ValidateOffset(0) = %v, %v; want the conservative default", got, err)
	}
	for _, bad := range []float64{0.001, 0.049, 0.61, 5} {
		if _, err := ValidateOffset(bad); err == nil {
			t.Errorf("ValidateOffset(%v) was accepted; %v%% is outside the safe range", bad, bad*100)
		}
	}
	if _, err := ValidateOffset(0.3); err != nil {
		t.Errorf("ValidateOffset(0.3): %v", err)
	}
}

// TestKRTickSchedule is the KRX table, including the boundary the openapi
// document names by example ("50,000~200,000원 구간은 100원 단위").
func TestKRTickSchedule(t *testing.T) {
	cases := []struct {
		price float64
		tick  float64
	}{
		{1999, 1}, {2000, 5}, {4999, 5}, {5000, 10}, {19999, 10},
		{20000, 50}, {49999, 50}, {50000, 100}, {199999, 100},
		{200000, 500}, {499999, 500}, {500000, 1000},
	}
	for _, c := range cases {
		if got := TickSize(c.price, MarketKR); got != c.tick {
			t.Errorf("TickSize(%v) = %v, want %v", c.price, got, c.tick)
		}
	}
	if got := TickSize(5, "US"); got != 0.01 {
		t.Errorf("TickSize(5, US) = %v, want a cent", got)
	}
	if got := TickSize(0.5, "US"); got != 0.0001 {
		t.Errorf("TickSize(0.5, US) = %v, want a hundredth of a cent below a dollar", got)
	}
}

func TestRoundingDirection(t *testing.T) {
	if got := RoundDownToTick(56050, MarketKR); got != 56000 {
		t.Errorf("RoundDownToTick(56050) = %v, want 56000", got)
	}
	if got := RoundUpToTick(56050, MarketKR); got != 56100 {
		t.Errorf("RoundUpToTick(56050) = %v, want 56100", got)
	}
	// The float noise a divide-and-multiply leaves behind would fail the
	// broker's integer pattern for KR prices.
	if got := RoundDownToTick(70000*0.8, MarketKR); got != 56000 {
		t.Errorf("RoundDownToTick(70000*0.8) = %v, want exactly 56000", got)
	}
}

func TestOneTickFurtherMovesAwayFromTheMarket(t *testing.T) {
	if got := OneTickFurther(56000, "buy", MarketKR); got != 55900 {
		t.Errorf("a buy moved to %v, want 55900 (one 100원 tick lower)", got)
	}
	if got := OneTickFurther(84000, "sell", MarketKR); got != 84100 {
		t.Errorf("a sell moved to %v, want 84100 (one tick higher)", got)
	}
}

func TestMarketOf(t *testing.T) {
	for symbol, want := range map[string]string{
		"005930": MarketKR, "000660": MarketKR, "AAPL": "US", "12345": "US", "0059300": "US", "00593A": "US",
	} {
		if got := MarketOf(symbol); got != want {
			t.Errorf("MarketOf(%q) = %q, want %q", symbol, got, want)
		}
	}
}

// TestMinQuantityIsOne. The exposure argument for this whole tool rests on it.
func TestMinQuantityIsOne(t *testing.T) {
	if MinQuantity != 1 {
		t.Fatalf("MinQuantity = %v; every live order this tool places must be one share", MinQuantity)
	}
}
