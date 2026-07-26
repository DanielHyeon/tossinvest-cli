package costs

// override_test.go is the port of StockOS `tests/test_costs_env_override.py`
// (16 cases) onto the TossOS **config-injection** seam.
//
// # What was rewritten, and why
//
// StockOS resolves overrides from process env under `KIS_*` keys
// (costs.py:11-21, :118-167). trade-analytics says both of those do not come
// across: KIS 수치·`KIS_*` 명명은 이식하지 않는다(SHALL NOT) … override는 설정
// 주입 방식으로 재구현하고 test_costs(4)·test_costs_env_override(16)는 **검증
// 게이트·주입 구조**에 대해 이식한다(수치 단언은 Toss 보수값으로 재작성).
//
// So: every numeric assertion is written against either a value the case
// injects itself or the TossOS `[미검증]` placeholders in costs.go, and every
// key is a TossOS config key. What is preserved exactly is the *gate*: the five
// rejection rules (empty / non-numeric / non-finite / negative / above
// MaxRate), the inclusive cap, whitespace tolerance, per-key isolation, and the
// registry sweep that catches a rate added to the model without being wired
// through the gate.
//
// Case mapping (StockOS → TossOS):
//
//	 1 test_from_env_with_empty_env_returns_dataclass_defaults      → TestNoOverridesReturnsDefaults
//	 2 test_from_env_with_unrelated_keys_returns_defaults           → TestUnknownOverrideKeyIsRefused  [의도적 반전]
//	 3 test_from_env_applies_each_valid_override_individually       → TestEachOverrideAppliesToItsOwnFieldOnly
//	 4 test_from_env_accepts_zero_rate                              → TestZeroRateIsAccepted
//	 5 test_from_env_accepts_value_at_cap                           → TestValueAtCapIsAccepted
//	 6 test_from_env_strips_surrounding_whitespace                  → TestSurroundingWhitespaceIsStripped
//	 7 test_from_env_rejects_non_numeric_value                      → TestNonNumericValueIsRefused
//	 8 test_from_env_rejects_empty_string                           → TestEmptyValueIsRefused
//	 9 test_from_env_rejects_negative_rate                          → TestNegativeRateIsRefused
//	10 test_from_env_rejects_value_above_cap                        → TestValueAboveCapIsRefused
//	11 test_from_env_rejects_nan                                    → TestNaNIsRefused
//	12 test_from_env_rejects_positive_infinity                      → TestInfinityIsRefused
//	13 test_every_overrideable_key_validates_negative_and_above_cap → TestEveryOverrideKeyGoesThroughTheGate
//	14 test_max_rate_constant_is_a_finite_positive_float            → TestMaxRateConstantShape
//	15 test_us_trade_cost_includes_commission_regulatory_fee_and_fx_fee → TestUSTradeCostIncludesCommissionRegulatoryFeeAndFXFee
//	16 test_us_break_even_floor_converts_krw_buffer_and_fx_fee_to_usd   → TestUSBreakEvenAppliesNativeProfitFloorAndFXFee  [FX 환산 제외]
//
// Two deliberate divergences, both recorded in the change's issues.md:
//
//   - **Case 2 is inverted.** An env map legitimately carries hundreds of
//     unrelated keys, so StockOS must ignore them. A TossOS override map is a
//     dedicated config block, so an unrecognised key is a typo — and a typo that
//     silently keeps the default is a rate nobody set behaving like a rate
//     somebody set. It is refused. The property case 2 actually protects (an
//     override the model does not know about must not perturb the model) is
//     preserved by TestNoOverridesReturnsDefaults plus the refusal.
//   - **Case 16 drops the KRW→USD conversion.** StockOS converts a KRW profit
//     floor into USD inside the cost model (costs.py:269-284). In TossOS,
//     currency normalisation with a staleness bound belongs to internal/riskcalc
//     (riskcalc.convert / FXRateStaleness) and must not be re-implemented here
//     with an unbounded rate. The floor is therefore native-currency, and the
//     case pins the same arithmetic (floor + FX fee inside the sell-side rate).

import (
	"math"
	"strconv"
	"strings"
	"testing"
)

func TestNoOverridesReturnsDefaults(t *testing.T) {
	for _, overrides := range []map[string]string{nil, {}} {
		got, err := NewModel(overrides)
		if err != nil {
			t.Fatalf("NewModel(%v): %v", overrides, err)
		}
		if got != DefaultModel() {
			t.Fatalf("NewModel(%v) = %+v, want the defaults %+v — an empty override set must not shift any rate",
				overrides, got, DefaultModel())
		}
	}
}

func TestUnknownOverrideKeyIsRefused(t *testing.T) {
	// See the header: inverted from StockOS on purpose.
	_, err := NewModel(map[string]string{"kr.buy_comission_rate": "0.001"})
	if err == nil {
		t.Fatal("an unrecognised override key was accepted; a mistyped key must not silently keep the default")
	}
	if !strings.Contains(err.Error(), "kr.buy_comission_rate") {
		t.Fatalf("error %q does not echo the offending key", err)
	}
}

func TestEachOverrideAppliesToItsOwnFieldOnly(t *testing.T) {
	// A copy-paste edit that rebinds one key to another key's field is
	// invisible to a single-key smoke test; this sweeps every key.
	values := map[string]string{
		KeyKRBuyCommissionRate:     "0.00042",
		KeyKRSellCommissionRate:    "0.00043",
		KeyKRSellTaxRate:           "0.00044",
		KeyUSBuyCommissionRate:     "0.00045",
		KeyUSSellCommissionRate:    "0.00046",
		KeyUSSellRegulatoryFeeRate: "0.00047",
		KeyUSFXConversionFeeRate:   "0.00048",
	}
	if len(values) != len(OverrideKeys()) {
		t.Fatalf("this case covers %d keys but the model has %d — a new rate was added without a case here",
			len(values), len(OverrideKeys()))
	}

	defaults := DefaultModel()
	for _, key := range OverrideKeys() {
		value, ok := values[key]
		if !ok {
			t.Fatalf("no test value for override key %q", key)
		}
		m, err := NewModel(map[string]string{key: value})
		if err != nil {
			t.Fatalf("NewModel(%s=%s): %v", key, value, err)
		}
		want, err := strconv.ParseFloat(value, 64)
		if err != nil {
			t.Fatalf("bad test value %q: %v", value, err)
		}
		if got := m.Rate(key); got != want {
			t.Fatalf("%s = %v, want %v", key, got, want)
		}
		for _, other := range OverrideKeys() {
			if other == key {
				continue
			}
			if got, def := m.Rate(other), defaults.Rate(other); got != def {
				t.Fatalf("overriding %s also moved %s (%v, default %v)", key, other, got, def)
			}
		}
	}
}

func TestZeroRateIsAccepted(t *testing.T) {
	// Zero is a legitimate rate (a promotional commission-free window), and
	// refusing it would make the operator express it as a lie instead.
	m := modelWith(t, map[string]string{KeyKRBuyCommissionRate: "0"})
	if got := m.Rate(KeyKRBuyCommissionRate); got != 0 {
		t.Fatalf("rate = %v, want 0", got)
	}
}

func TestValueAtCapIsAccepted(t *testing.T) {
	// The cap is inclusive: exactly MaxRate passes.
	m := modelWith(t, map[string]string{KeyKRSellTaxRate: strconv.FormatFloat(MaxRate, 'f', -1, 64)})
	if got := m.Rate(KeyKRSellTaxRate); got != MaxRate {
		t.Fatalf("rate = %v, want the cap %v", got, MaxRate)
	}
}

func TestSurroundingWhitespaceIsStripped(t *testing.T) {
	m := modelWith(t, map[string]string{KeyKRBuyCommissionRate: "  0.0001  "})
	if got := m.Rate(KeyKRBuyCommissionRate); got != 0.0001 {
		t.Fatalf("rate = %v, want 0.0001", got)
	}
}

func TestNonNumericValueIsRefused(t *testing.T) {
	_, err := NewModel(map[string]string{KeyKRBuyCommissionRate: "not-a-number"})
	requireRateError(t, err, KeyKRBuyCommissionRate, "numeric")
}

func TestEmptyValueIsRefused(t *testing.T) {
	// An override present but empty is "the operator forgot the value". Falling
	// through to the default would mask a broken deployment artifact.
	_, err := NewModel(map[string]string{KeyKRSellTaxRate: "   "})
	requireRateError(t, err, KeyKRSellTaxRate, "numeric")
}

func TestNegativeRateIsRefused(t *testing.T) {
	// A negative rate flips the sign of both P&L and the break-even price: a
	// misconfiguration would read as a trade that pays you to make it.
	_, err := NewModel(map[string]string{KeyKRSellCommissionRate: "-0.0001"})
	requireRateError(t, err, KeyKRSellCommissionRate, "non-negative")
}

func TestValueAboveCapIsRefused(t *testing.T) {
	_, err := NewModel(map[string]string{KeyKRSellTaxRate: "0.20"})
	requireRateError(t, err, KeyKRSellTaxRate, "")
	// The message must point at the off-by-100 cause, so an operator fixes the
	// value instead of raising the cap.
	if msg := strings.ToLower(err.Error()); !strings.Contains(msg, "typo") && !strings.Contains(msg, "fraction") {
		t.Fatalf("error %q does not explain the cap", err)
	}
}

func TestNaNIsRefused(t *testing.T) {
	_, err := NewModel(map[string]string{KeyKRBuyCommissionRate: "nan"})
	requireRateError(t, err, KeyKRBuyCommissionRate, "finite")
}

func TestInfinityIsRefused(t *testing.T) {
	_, err := NewModel(map[string]string{KeyKRBuyCommissionRate: "inf"})
	requireRateError(t, err, KeyKRBuyCommissionRate, "finite")
}

func TestEveryOverrideKeyGoesThroughTheGate(t *testing.T) {
	// The regression this catches: a rate added to the model and to the
	// override registry, but resolved without the gate. It would show up here
	// as one key accepting -0.001.
	if len(OverrideKeys()) == 0 {
		t.Fatal("the override registry is empty; the sweep would assert nothing")
	}
	for _, key := range OverrideKeys() {
		for _, bad := range []string{"-0.001", strconv.FormatFloat(MaxRate+0.001, 'f', -1, 64), "nan", "inf", "", "0.2%"} {
			if _, err := NewModel(map[string]string{key: bad}); err == nil {
				t.Fatalf("%s accepted %q; this key does not go through the validation gate", key, bad)
			}
		}
	}
}

func TestMaxRateConstantShape(t *testing.T) {
	// Lock the cap's shape so a later edit cannot make it negative, non-finite
	// or ≥ 1 — a sell-side rate of 1 divides the break-even price by zero.
	v := float64(MaxRate)
	if math.IsNaN(v) || math.IsInf(v, 0) {
		t.Fatalf("MaxRate = %v, want finite", v)
	}
	if v <= 0 || v >= 1 {
		t.Fatalf("MaxRate = %v, want 0 < MaxRate < 1", v)
	}
}

func TestUSTradeCostIncludesCommissionRegulatoryFeeAndFXFee(t *testing.T) {
	m := modelWith(t, map[string]string{
		KeyUSBuyCommissionRate:     "0.001",
		KeyUSSellCommissionRate:    "0.002",
		KeyUSSellRegulatoryFeeRate: "0.0001",
		KeyUSFXConversionFeeRate:   "0.0005",
	})

	buy, err := m.EstimateTradeCost("1000", SideBuy, MarketUS)
	if err != nil {
		t.Fatalf("EstimateTradeCost(BUY): %v", err)
	}
	closeTo(t, amount(t, buy.Commission), 1.0, 1e-9, "US buy commission")
	closeTo(t, amount(t, buy.FXFee), 0.5, 1e-9, "US buy FX fee")
	if got := amount(t, buy.Tax); got != 0 {
		t.Fatalf("US buy regulatory fee = %v, want 0 — it is sell-side only", got)
	}
	closeTo(t, amount(t, buy.Total), 1.5, 1e-9, "US buy total")

	sell, err := m.EstimateTradeCost("1000", SideSell, MarketUS)
	if err != nil {
		t.Fatalf("EstimateTradeCost(SELL): %v", err)
	}
	closeTo(t, amount(t, sell.Commission), 2.0, 1e-9, "US sell commission")
	closeTo(t, amount(t, sell.Tax), 0.1, 1e-9, "US sell regulatory fee")
	closeTo(t, amount(t, sell.FXFee), 0.5, 1e-9, "US sell FX fee")
	closeTo(t, amount(t, sell.Total), 2.6, 1e-9, "US sell total")
}

func TestUSBreakEvenAppliesNativeProfitFloorAndFXFee(t *testing.T) {
	// StockOS's case converts a 1,350 KRW floor at 1,350 KRW/USD into 1 USD
	// inside the cost model. TossOS takes the floor already in the trade's own
	// currency (see the header): the arithmetic under test is the same one.
	m := modelWith(t, map[string]string{
		KeyUSBuyCommissionRate:     "0.001",
		KeyUSSellCommissionRate:    "0.001",
		KeyUSSellRegulatoryFeeRate: "0",
		KeyUSFXConversionFeeRate:   "0.0005",
	})

	got, err := m.BreakEvenSellPriceWithFloor("100", "1", MarketUS, "1")
	if err != nil {
		t.Fatalf("BreakEvenSellPriceWithFloor: %v", err)
	}
	// Buy leg: commission 0.10 + FX 0.05 = 0.15. Required profit: 1.
	// Sell-side rate: commission 0.001 + FX 0.0005 = 0.0015.
	want := (100.0 + 0.15 + 1.0) / (1 - 0.0015)
	closeTo(t, amount(t, got), want, 1e-9, "US break-even with a native profit floor")
}

func requireRateError(t *testing.T, err error, key, detail string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: the gate accepted an invalid rate", key)
	}
	if !strings.Contains(err.Error(), key) {
		t.Fatalf("error %q does not echo the key %q; an operator cannot find the bad value", err, key)
	}
	if detail != "" && !strings.Contains(err.Error(), detail) {
		t.Fatalf("error %q does not say why (%q)", err, detail)
	}
}
