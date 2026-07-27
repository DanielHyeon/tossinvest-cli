package riskcalc_test

// riskcalc tests (extend-execution-contract task 6.1).
//
// The requirement is order-execution "총계 한도의 계산 계약": the aggregate
// quantities must be *defined* before anything reserves against them, and an
// input that is stale or unknown refuses the entry rather than being guessed at.
//
// Every test supplies its own instant. There is no clock here to fake, which is
// the property being pinned as much as any assertion below.

import (
	"errors"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

var now = time.Date(2026, 3, 30, 1, 30, 0, 0, time.UTC) // 10:30 KST

func fresh(d time.Duration) time.Time { return now.Add(-d) }

// --- LIMIT-only automated entry ---------------------------------------------

// TestMarketEntryIsRefused is the spec's "시장가 진입 시도" scenario. A market
// order has no price until it fills, so its exposure valuation does not exist —
// and a limit measured against an undefined number is not a limit.
func TestMarketEntryIsRefused(t *testing.T) {
	err := riskcalc.CheckAutomatedEntry(riskcalc.EntryIntent{
		Symbol: "AAPL", OrderType: riskcalc.OrderTypeMarket,
		Quantity: "10", Currency: "USD",
	})
	if !errors.Is(err, riskcalc.ErrMarketEntry) {
		t.Fatalf("err = %v, want ErrMarketEntry", err)
	}
}

// TestUnknownOrderTypeIsRefused: fail-closed applies to the order type too. Only
// LIMIT is allowed, not "anything that is not MARKET".
func TestUnknownOrderTypeIsRefused(t *testing.T) {
	for _, ot := range []riskcalc.OrderType{"", "STOP", "limit-ish"} {
		err := riskcalc.CheckAutomatedEntry(riskcalc.EntryIntent{
			Symbol: "AAPL", OrderType: ot, LimitPrice: "200", Quantity: "10", Currency: "USD",
		})
		if !errors.Is(err, riskcalc.ErrMarketEntry) {
			t.Errorf("order type %q: err = %v, want ErrMarketEntry", ot, err)
		}
	}
}

// TestLimitEntryExposureIsPriceTimesQuantityPlusCosts transcribes the spec's
// "진입 노출 평가 = 지정가 × 수량 + 과대 추정 비용".
func TestLimitEntryExposureIsPriceTimesQuantityPlusCosts(t *testing.T) {
	got, err := riskcalc.EntryExposure(riskcalc.EntryIntent{
		Symbol: "AAPL", OrderType: riskcalc.OrderTypeLimit,
		LimitPrice: "200.5", Quantity: "10", CostOverestimate: "3.25", Currency: "USD",
	})
	if err != nil {
		t.Fatalf("EntryExposure: %v", err)
	}
	if got.Amount != "2008.25" || got.Currency != "USD" {
		t.Fatalf("exposure = %+v, want 2008.25 USD", got)
	}
}

// TestEntryExposureNeedsItsCostEstimate: the cost term is not optional. An empty
// one is an unknown input, not a zero — "no fee" is a claim about the broker
// nobody measured.
func TestEntryExposureNeedsItsCostEstimate(t *testing.T) {
	_, err := riskcalc.EntryExposure(riskcalc.EntryIntent{
		Symbol: "AAPL", OrderType: riskcalc.OrderTypeLimit,
		LimitPrice: "200.5", Quantity: "10", Currency: "USD",
	})
	if !errors.Is(err, riskcalc.ErrUnknownInput) {
		t.Fatalf("err = %v, want ErrUnknownInput", err)
	}
}

// TestMarketEntryExposureIsRefusedNotComputed: the rejection lives in the
// valuation too, so no caller can reach a number by skipping the check.
func TestMarketEntryExposureIsRefusedNotComputed(t *testing.T) {
	_, err := riskcalc.EntryExposure(riskcalc.EntryIntent{
		Symbol: "AAPL", OrderType: riskcalc.OrderTypeMarket,
		LimitPrice: "200.5", Quantity: "10", CostOverestimate: "3", Currency: "USD",
	})
	if !errors.Is(err, riskcalc.ErrMarketEntry) {
		t.Fatalf("err = %v, want ErrMarketEntry", err)
	}
}

// --- gross long exposure ----------------------------------------------------

func krwSnapshot(age time.Duration) riskcalc.AccountSnapshot {
	return riskcalc.AccountSnapshot{
		AsOf: fresh(age),
		Holdings: []riskcalc.Holding{
			{Symbol: "005930", Quantity: "10", Price: "70000", Currency: "KRW"},
		},
		UnfilledBuys: []riskcalc.UnfilledBuy{
			{Symbol: "000660", LimitPrice: "150000", RemainingQuantity: "2", Currency: "KRW"},
		},
	}
}

// TestGrossLongComposition is the spec's "보유 평가액 + 미체결 매수 평가액 +
// 유효 진입 예약의 합". Long-only gross: nothing nets anything else off.
func TestGrossLongComposition(t *testing.T) {
	got, err := riskcalc.GrossLongExposure(riskcalc.GrossLongInputs{
		Now:            now,
		TargetCurrency: "KRW",
		Snapshot:       krwSnapshot(2 * time.Second),
		Reservations: []riskcalc.EntryReservation{
			{DecisionID: "d-1", Amount: "500000", Currency: "KRW"},
		},
	})
	if err != nil {
		t.Fatalf("GrossLongExposure: %v", err)
	}
	// 10×70000 + 2×150000 + 500000
	if got.Amount != "1500000" || got.Currency != "KRW" {
		t.Fatalf("gross long = %+v, want 1500000 KRW", got)
	}
}

// TestGrossLongCountsEveryComponentSeparately proves each term is really added,
// rather than one of them accidentally standing in for the total.
func TestGrossLongCountsEveryComponentSeparately(t *testing.T) {
	base := riskcalc.GrossLongInputs{Now: now, TargetCurrency: "KRW",
		Snapshot: riskcalc.AccountSnapshot{AsOf: fresh(time.Second)}}

	cases := []struct {
		name string
		in   riskcalc.GrossLongInputs
		want string
	}{
		{name: "nothing at all", in: base, want: "0"},
		{name: "holdings only", want: "700000", in: func() riskcalc.GrossLongInputs {
			in := base
			in.Snapshot.Holdings = []riskcalc.Holding{
				{Symbol: "005930", Quantity: "10", Price: "70000", Currency: "KRW"}}
			return in
		}()},
		{name: "unfilled buys only", want: "300000", in: func() riskcalc.GrossLongInputs {
			in := base
			in.Snapshot.UnfilledBuys = []riskcalc.UnfilledBuy{
				{Symbol: "000660", LimitPrice: "150000", RemainingQuantity: "2", Currency: "KRW"}}
			return in
		}()},
		{name: "reservations only", want: "500000", in: func() riskcalc.GrossLongInputs {
			in := base
			in.Reservations = []riskcalc.EntryReservation{
				{DecisionID: "d-1", Amount: "500000", Currency: "KRW"}}
			return in
		}()},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := riskcalc.GrossLongExposure(tc.in)
			if err != nil {
				t.Fatalf("GrossLongExposure: %v", err)
			}
			if got.Amount != tc.want {
				t.Errorf("gross long = %q, want %q", got.Amount, tc.want)
			}
		})
	}
}

// TestStaleAccountSnapshotIsRefused: the exposure is a claim about *now*, and a
// snapshot older than the bound cannot support it.
func TestStaleAccountSnapshotIsRefused(t *testing.T) {
	_, err := riskcalc.GrossLongExposure(riskcalc.GrossLongInputs{
		Now:            now,
		TargetCurrency: "KRW",
		Snapshot:       krwSnapshot(riskcalc.AccountSnapshotStaleness + time.Second),
	})
	if !errors.Is(err, riskcalc.ErrStaleInput) {
		t.Fatalf("err = %v, want ErrStaleInput", err)
	}
}

// TestSnapshotWithNoAsOfIsUnknownNotFresh: a zero as-of is the absence of the
// fact, and absence must not read as "just now".
func TestSnapshotWithNoAsOfIsUnknownNotFresh(t *testing.T) {
	in := riskcalc.GrossLongInputs{Now: now, TargetCurrency: "KRW", Snapshot: krwSnapshot(0)}
	in.Snapshot.AsOf = time.Time{}
	if _, err := riskcalc.GrossLongExposure(in); !errors.Is(err, riskcalc.ErrUnknownInput) {
		t.Fatalf("err = %v, want ErrUnknownInput", err)
	}
}

// TestFutureDatedSnapshotIsRefused: a negative age means the clocks disagree, and
// a clock nobody can trust is not a freshness proof.
func TestFutureDatedSnapshotIsRefused(t *testing.T) {
	in := riskcalc.GrossLongInputs{Now: now, TargetCurrency: "KRW", Snapshot: krwSnapshot(0)}
	in.Snapshot.AsOf = now.Add(2 * time.Second)
	if _, err := riskcalc.GrossLongExposure(in); !errors.Is(err, riskcalc.ErrStaleInput) {
		t.Fatalf("err = %v, want ErrStaleInput", err)
	}
}

// TestMissingNowIsUnknown: the caller supplies the instant, so forgetting it is
// an unknown input rather than the zero time.
func TestMissingNowIsUnknown(t *testing.T) {
	in := riskcalc.GrossLongInputs{TargetCurrency: "KRW", Snapshot: krwSnapshot(time.Second)}
	if _, err := riskcalc.GrossLongExposure(in); !errors.Is(err, riskcalc.ErrUnknownInput) {
		t.Fatalf("err = %v, want ErrUnknownInput", err)
	}
}

// TestUnreadableHoldingIsRefused: a price that is not a decimal is not a zero.
func TestUnreadableHoldingIsRefused(t *testing.T) {
	in := riskcalc.GrossLongInputs{Now: now, TargetCurrency: "KRW",
		Snapshot: riskcalc.AccountSnapshot{AsOf: fresh(time.Second),
			Holdings: []riskcalc.Holding{
				{Symbol: "005930", Quantity: "10", Price: "seventy thousand", Currency: "KRW"}}}}
	if _, err := riskcalc.GrossLongExposure(in); !errors.Is(err, riskcalc.ErrUnknownInput) {
		t.Fatalf("err = %v, want ErrUnknownInput", err)
	}
}

// TestHoldingWithoutACurrencyIsRefused: an amount with no currency cannot be
// normalised, and treating it as the target currency would be a guess.
func TestHoldingWithoutACurrencyIsRefused(t *testing.T) {
	in := riskcalc.GrossLongInputs{Now: now, TargetCurrency: "KRW",
		Snapshot: riskcalc.AccountSnapshot{AsOf: fresh(time.Second),
			Holdings: []riskcalc.Holding{{Symbol: "005930", Quantity: "10", Price: "70000"}}}}
	if _, err := riskcalc.GrossLongExposure(in); !errors.Is(err, riskcalc.ErrUnknownInput) {
		t.Fatalf("err = %v, want ErrUnknownInput", err)
	}
}

// TestMissingTargetCurrencyIsRefused: there is no default currency.
func TestMissingTargetCurrencyIsRefused(t *testing.T) {
	in := riskcalc.GrossLongInputs{Now: now, Snapshot: krwSnapshot(time.Second)}
	if _, err := riskcalc.GrossLongExposure(in); !errors.Is(err, riskcalc.ErrUnknownInput) {
		t.Fatalf("err = %v, want ErrUnknownInput", err)
	}
}

// --- currency normalisation -------------------------------------------------

// TestStaleFXRateRefusesTheEntry is the spec's "환율 stale" scenario.
func TestStaleFXRateRefusesTheEntry(t *testing.T) {
	in := riskcalc.GrossLongInputs{
		Now: now, TargetCurrency: "KRW",
		Snapshot: riskcalc.AccountSnapshot{AsOf: fresh(time.Second),
			Holdings: []riskcalc.Holding{
				{Symbol: "AAPL", Quantity: "10", Price: "200", Currency: "USD"}}},
		Rates: []riskcalc.FXRate{
			{From: "USD", To: "KRW", Rate: "1350", AsOf: fresh(riskcalc.FXRateStaleness + time.Second)}},
	}
	if _, err := riskcalc.GrossLongExposure(in); !errors.Is(err, riskcalc.ErrStaleInput) {
		t.Fatalf("err = %v, want ErrStaleInput", err)
	}
}

// TestFreshFXRateNormalises: the same holding with a rate inside the bound.
func TestFreshFXRateNormalises(t *testing.T) {
	in := riskcalc.GrossLongInputs{
		Now: now, TargetCurrency: "KRW",
		Snapshot: riskcalc.AccountSnapshot{AsOf: fresh(time.Second),
			Holdings: []riskcalc.Holding{
				{Symbol: "AAPL", Quantity: "10", Price: "200", Currency: "USD"}}},
		Rates: []riskcalc.FXRate{
			{From: "USD", To: "KRW", Rate: "1350", AsOf: fresh(30 * time.Second)}},
	}
	got, err := riskcalc.GrossLongExposure(in)
	if err != nil {
		t.Fatalf("GrossLongExposure: %v", err)
	}
	if got.Amount != "2700000" { // 10×200×1350
		t.Fatalf("gross long = %q, want 2700000", got.Amount)
	}
}

// TestMissingFXRateIsUnknownNotOne: a foreign amount with no rate at all is the
// unknown input the spec fails closed on, not a 1:1 conversion.
func TestMissingFXRateIsUnknownNotOne(t *testing.T) {
	in := riskcalc.GrossLongInputs{
		Now: now, TargetCurrency: "KRW",
		Snapshot: riskcalc.AccountSnapshot{AsOf: fresh(time.Second),
			Holdings: []riskcalc.Holding{
				{Symbol: "AAPL", Quantity: "10", Price: "200", Currency: "USD"}}},
	}
	if _, err := riskcalc.GrossLongExposure(in); !errors.Is(err, riskcalc.ErrUnknownInput) {
		t.Fatalf("err = %v, want ErrUnknownInput", err)
	}
}

// TestSameCurrencyNeedsNoRate: normalisation is not a round trip through a rate
// that does not exist.
func TestSameCurrencyNeedsNoRate(t *testing.T) {
	got, err := riskcalc.GrossLongExposure(riskcalc.GrossLongInputs{
		Now: now, TargetCurrency: "KRW", Snapshot: krwSnapshot(time.Second),
	})
	if err != nil {
		t.Fatalf("GrossLongExposure: %v", err)
	}
	if got.Amount != "1000000" {
		t.Fatalf("gross long = %q, want 1000000", got.Amount)
	}
}

// TestNonPositiveFXRateIsRefused: a zero or negative rate would shrink exposure,
// which is the one direction a bad input must never move it.
func TestNonPositiveFXRateIsRefused(t *testing.T) {
	for _, rate := range []string{"0", "-1350", ""} {
		in := riskcalc.GrossLongInputs{
			Now: now, TargetCurrency: "KRW",
			Snapshot: riskcalc.AccountSnapshot{AsOf: fresh(time.Second),
				Holdings: []riskcalc.Holding{
					{Symbol: "AAPL", Quantity: "10", Price: "200", Currency: "USD"}}},
			Rates: []riskcalc.FXRate{
				{From: "USD", To: "KRW", Rate: rate, AsOf: fresh(time.Second)}},
		}
		if _, err := riskcalc.GrossLongExposure(in); !errors.Is(err, riskcalc.ErrUnknownInput) {
			t.Errorf("rate %q: err = %v, want ErrUnknownInput", rate, err)
		}
	}
}

// --- daily loss -------------------------------------------------------------

// TestDailyLossIsRealisedAfterCosts transcribes "일일 손실은 실현 손익(비용 차감
// 후) 기준". A loss is reported as a positive magnitude; a profit is zero loss.
func TestDailyLossIsRealisedAfterCosts(t *testing.T) {
	in := riskcalc.DailyLossInputs{
		Now: now, TradingDay: "2026-03-30", TargetCurrency: "KRW",
		Realized: []riskcalc.RealizedPnL{
			{TradingDay: "2026-03-30", Amount: "-120000", Currency: "KRW",
				CostsDeducted: true, AsOf: fresh(time.Second)},
			{TradingDay: "2026-03-30", Amount: "40000", Currency: "KRW",
				CostsDeducted: true, AsOf: fresh(time.Second)},
		},
	}
	got, err := riskcalc.DailyLoss(in)
	if err != nil {
		t.Fatalf("DailyLoss: %v", err)
	}
	if got.Amount != "80000" || got.Currency != "KRW" {
		t.Fatalf("daily loss = %+v, want 80000 KRW", got)
	}
}

// TestProfitableDayHasNoLoss: the quantity is a loss, so a net gain is 0 and
// never a negative "loss" a comparison would read as headroom.
func TestProfitableDayHasNoLoss(t *testing.T) {
	got, err := riskcalc.DailyLoss(riskcalc.DailyLossInputs{
		Now: now, TradingDay: "2026-03-30", TargetCurrency: "KRW",
		Realized: []riskcalc.RealizedPnL{
			{TradingDay: "2026-03-30", Amount: "40000", Currency: "KRW",
				CostsDeducted: true, AsOf: fresh(time.Second)}},
	})
	if err != nil {
		t.Fatalf("DailyLoss: %v", err)
	}
	if got.Amount != "0" {
		t.Fatalf("daily loss = %q, want 0", got.Amount)
	}
}

// TestGrossPnLWithoutCostsIsRefused: "비용 차감 후" is the definition, not a
// preference. A caller that cannot say costs were deducted has an unknown input.
func TestGrossPnLWithoutCostsIsRefused(t *testing.T) {
	_, err := riskcalc.DailyLoss(riskcalc.DailyLossInputs{
		Now: now, TradingDay: "2026-03-30", TargetCurrency: "KRW",
		Realized: []riskcalc.RealizedPnL{
			{TradingDay: "2026-03-30", Amount: "-120000", Currency: "KRW", AsOf: fresh(time.Second)}},
	})
	if !errors.Is(err, riskcalc.ErrUnknownInput) {
		t.Fatalf("err = %v, want ErrUnknownInput", err)
	}
}

// TestRealizedFromAnotherTradingDayIsRefused: a daily limit measured over the
// wrong day is not the limit anyone configured.
func TestRealizedFromAnotherTradingDayIsRefused(t *testing.T) {
	_, err := riskcalc.DailyLoss(riskcalc.DailyLossInputs{
		Now: now, TradingDay: "2026-03-30", TargetCurrency: "KRW",
		Realized: []riskcalc.RealizedPnL{
			{TradingDay: "2026-03-27", Amount: "-120000", Currency: "KRW",
				CostsDeducted: true, AsOf: fresh(time.Second)}},
	})
	if !errors.Is(err, riskcalc.ErrUnknownInput) {
		t.Fatalf("err = %v, want ErrUnknownInput", err)
	}
}

// TestStaleRealizedPnLIsRefused: the realised total comes from the broker
// snapshot, so it carries the account snapshot's bound.
func TestStaleRealizedPnLIsRefused(t *testing.T) {
	_, err := riskcalc.DailyLoss(riskcalc.DailyLossInputs{
		Now: now, TradingDay: "2026-03-30", TargetCurrency: "KRW",
		Realized: []riskcalc.RealizedPnL{
			{TradingDay: "2026-03-30", Amount: "-120000", Currency: "KRW", CostsDeducted: true,
				AsOf: fresh(riskcalc.AccountSnapshotStaleness + time.Second)}},
	})
	if !errors.Is(err, riskcalc.ErrStaleInput) {
		t.Fatalf("err = %v, want ErrStaleInput", err)
	}
}

// TestDailyLossNormalisesCurrency: a foreign realised loss needs a fresh rate
// like everything else.
func TestDailyLossNormalisesCurrency(t *testing.T) {
	in := riskcalc.DailyLossInputs{
		Now: now, TradingDay: "2026-03-30", TargetCurrency: "KRW",
		Realized: []riskcalc.RealizedPnL{
			{TradingDay: "2026-03-30", Amount: "-100", Currency: "USD",
				CostsDeducted: true, AsOf: fresh(time.Second)}},
		Rates: []riskcalc.FXRate{
			{From: "USD", To: "KRW", Rate: "1350", AsOf: fresh(time.Second)}},
	}
	got, err := riskcalc.DailyLoss(in)
	if err != nil {
		t.Fatalf("DailyLoss: %v", err)
	}
	if got.Amount != "135000" {
		t.Fatalf("daily loss = %q, want 135000", got.Amount)
	}

	in.Rates[0].AsOf = fresh(riskcalc.FXRateStaleness + time.Second)
	if _, err := riskcalc.DailyLoss(in); !errors.Is(err, riskcalc.ErrStaleInput) {
		t.Fatalf("with a stale rate: err = %v, want ErrStaleInput", err)
	}
}

// --- limit comparison -------------------------------------------------------

// TestWithinLimitIsFailClosedOnTies: reaching the limit is not staying under it.
func TestWithinLimitIsFailClosedOnTies(t *testing.T) {
	cases := []struct {
		value, limit string
		want         bool
	}{
		{value: "999999", limit: "1000000", want: true},
		{value: "1000000", limit: "1000000", want: false},
		{value: "1000001", limit: "1000000", want: false},
	}
	for _, tc := range cases {
		got, err := riskcalc.WithinLimit(
			riskcalc.Money{Amount: tc.value, Currency: "KRW"},
			riskcalc.Money{Amount: tc.limit, Currency: "KRW"})
		if err != nil {
			t.Fatalf("WithinLimit(%s, %s): %v", tc.value, tc.limit, err)
		}
		if got != tc.want {
			t.Errorf("WithinLimit(%s, %s) = %v, want %v", tc.value, tc.limit, got, tc.want)
		}
	}
}

// TestWithinLimitRefusesAnUnconfiguredLimit: "no limit" must not read as
// "unlimited". An unset or non-positive limit is an unknown input.
func TestWithinLimitRefusesAnUnconfiguredLimit(t *testing.T) {
	for _, limit := range []string{"", "0", "-1"} {
		_, err := riskcalc.WithinLimit(
			riskcalc.Money{Amount: "1", Currency: "KRW"},
			riskcalc.Money{Amount: limit, Currency: "KRW"})
		if !errors.Is(err, riskcalc.ErrUnknownInput) {
			t.Errorf("limit %q: err = %v, want ErrUnknownInput", limit, err)
		}
	}
}

// TestWithinLimitRefusesMixedCurrencies: normalisation happens before the
// comparison, never inside it — comparing across currencies without a rate is
// exactly the guess this package exists to prevent.
func TestWithinLimitRefusesMixedCurrencies(t *testing.T) {
	_, err := riskcalc.WithinLimit(
		riskcalc.Money{Amount: "1000", Currency: "USD"},
		riskcalc.Money{Amount: "1000000", Currency: "KRW"})
	if !errors.Is(err, riskcalc.ErrUnknownInput) {
		t.Fatalf("err = %v, want ErrUnknownInput", err)
	}
}

// --- the conservative defaults ----------------------------------------------

// TestStalenessDefaultsAreTheTranscribedNumbers pins D6a. They are not tuning
// knobs: WORKFLOW §0.9 allows only a conservative (shorter) replacement, and
// anything else needs a human's approval and an audit record (§0.5, §0.7).
func TestStalenessDefaultsAreTheTranscribedNumbers(t *testing.T) {
	if riskcalc.AccountSnapshotStaleness != 10*time.Second {
		t.Errorf("AccountSnapshotStaleness = %s, want 10s (design D6a)",
			riskcalc.AccountSnapshotStaleness)
	}
	if riskcalc.FXRateStaleness != 60*time.Second {
		t.Errorf("FXRateStaleness = %s, want 60s (design D6a)", riskcalc.FXRateStaleness)
	}
	d := riskcalc.DefaultStaleness()
	if d.AccountSnapshot != riskcalc.AccountSnapshotStaleness || d.FXRate != riskcalc.FXRateStaleness {
		t.Errorf("DefaultStaleness() = %+v, want the constants", d)
	}
}

// TestTighterBoundsAreAccepted: a caller may only narrow the window. Widening it
// is the one direction §0.9 forbids without a human.
func TestTighterBoundsAreAccepted(t *testing.T) {
	in := riskcalc.GrossLongInputs{
		Now: now, TargetCurrency: "KRW",
		Bounds:   riskcalc.Staleness{AccountSnapshot: 2 * time.Second, FXRate: 2 * time.Second},
		Snapshot: krwSnapshot(5 * time.Second),
	}
	if _, err := riskcalc.GrossLongExposure(in); !errors.Is(err, riskcalc.ErrStaleInput) {
		t.Fatalf("err = %v, want ErrStaleInput under a 2s bound", err)
	}
}

// TestWiderBoundsAreRefused: a caller cannot loosen the transcribed default by
// passing a longer window.
func TestWiderBoundsAreRefused(t *testing.T) {
	in := riskcalc.GrossLongInputs{
		Now: now, TargetCurrency: "KRW",
		Bounds:   riskcalc.Staleness{AccountSnapshot: time.Hour, FXRate: time.Hour},
		Snapshot: krwSnapshot(30 * time.Second),
	}
	if _, err := riskcalc.GrossLongExposure(in); !errors.Is(err, riskcalc.ErrStaleInput) {
		t.Fatalf("err = %v, want the entry still refused: a wider bound must not take effect", err)
	}
}

// TestInputErrorNamesWhatWasMissing: fail-closed is only operable if the
// operator can see which input failed.
func TestInputErrorNamesWhatWasMissing(t *testing.T) {
	_, err := riskcalc.GrossLongExposure(riskcalc.GrossLongInputs{
		Now: now, TargetCurrency: "KRW",
		Snapshot: krwSnapshot(riskcalc.AccountSnapshotStaleness + time.Second),
	})
	var inputErr *riskcalc.InputError
	if !errors.As(err, &inputErr) {
		t.Fatalf("err = %v (%T), want a *riskcalc.InputError", err, err)
	}
	if inputErr.Input != "account snapshot" {
		t.Errorf("Input = %q, want the named input", inputErr.Input)
	}
	if inputErr.Bound != riskcalc.AccountSnapshotStaleness {
		t.Errorf("Bound = %s, want the bound that was exceeded", inputErr.Bound)
	}
}
