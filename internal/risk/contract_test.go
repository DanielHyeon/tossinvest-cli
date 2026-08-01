package risk

// contract_test.go is task 2.2: the stop contract, the risk-based quantity and
// the minimum reward:risk — ported from StockOS
// `tests/test_target_stop_contract.py` (29 cases) and
// `tests/test_a090_tradeplan_entry_contract.py` (36 cases).
//
// # test_target_stop_contract.py — 13 이식 / 16 제외
//
// 이식 (layers 1-2, the constructor invariants and the Guardian verdicts —
// TossOS has no separate constructor, so both layers land on the chain):
//
//	test_candidate_requires_target_above_entry_for_buy  → TestTargetAtOrBelowEntryIsRefused
//	test_candidate_rejects_stop_at_or_above_entry_for_buy → TestStopAtOrAboveEntryIsRefused
//	test_candidate_rejects_zero_target                  → TestNonPositiveOrUnreadableTargetStopIsRefused
//	test_candidate_rejects_zero_stop                    → TestNonPositiveOrUnreadableTargetStopIsRefused
//	test_candidate_rejects_negative_target              → TestNonPositiveOrUnreadableTargetStopIsRefused
//	test_candidate_rejects_nan_target                   → TestNonPositiveOrUnreadableTargetStopIsRefused
//	test_candidate_rejects_inf_stop                     → TestNonPositiveOrUnreadableTargetStopIsRefused
//	test_candidate_rejects_zero_quantity                → TestSizeZeroIsRefused
//	test_guardian_blocks_target_below_break_even        → TestTargetBelowBreakEvenIsRefused
//	test_guardian_does_not_compute_target_stop          → TestTheChainNeverRepairsATargetOrStop
//	test_guardian_blocks_stop_at_or_above_entry         → TestStopAtOrAboveEntryIsRefused
//	test_guardian_blocks_target_at_or_below_entry       → TestTargetAtOrBelowEntryIsRefused
//	test_guardian_target_stop_check_runs_before_size_limit → TestStopContractRunsBeforeTheSizeLimit
//
// 제외 (layer 3, 16건 — 전부 전략·신호 계층): test_strategy_* (6),
// test_select_target_* (5), test_select_stop_* (4), test_candidate_metadata_
// carries_audit_labels (1). 사유: Bollinger 밴드·fractal 투영·pct floor는 신호
// 계층 산물이고, risk-management는 체인의 모든 입력이 의도 필드·브로커 스냅샷·
// journal 상태에서 온다고 못박는다(SHALL — 신호 계층 산물 없음: 구조적 RR 계산·
// 등급배수는 P3). TossOS의 Guardian은 target/stop을 **계산하지 않고 검사만** 한다
// — 원본의 layer 3은 계산하는 쪽의 테스트다.
//
// # test_a090_tradeplan_entry_contract.py — 12 이식 / 24 제외
//
// 이식:
//
//	test_config_rejects_invalid_thresholds       → TestPolicyRejectsInvalidThresholds
//	test_stop_missing_rejected                   → TestStopMissingIsRefusedBeforeSizing
//	test_stop_not_protective_rejected            → TestStopAtOrAboveEntryIsRefused
//	test_rr_below_min_rejected                   → TestRewardRiskBelowTheMinimumIsRefused
//	test_rr_missing_fail_closed                  → TestMissingTargetIsRefused
//	test_rr_unparsable_fail_closed               → TestNonPositiveOrUnreadableTargetStopIsRefused
//	test_grade_multiplier_sizing_spec_example    → TestRiskBasedQuantity (배수 1.0 — 아래 주석)
//	test_a_grade_full_size                       → TestRiskBasedQuantity (같은 이유)
//	test_sizing_single_floor_on_non_divisible_width → TestRiskBasedQuantityFloorsExactlyOnce
//	test_size_zero_rejected                      → TestSizeZeroIsRefused
//	test_inputs_unavailable_on_bad_entry         → TestUnusableEntryPriceIsInputUnavailable
//	test_build_trade_plan_does_not_mutate_inputs → TestEvaluateDoesNotMutateItsInput (chain_test.go)
//	test_us_without_fx_fails_closed_without_mixed_currency_qty → TestForeignCurrencyIntentIsNotSizedAgainstADomesticBudget (chain_test.go)
//	test_us_market_stop_rr_rungs_still_enforced  → 같은 케이스의 후반부
//	test_unknown_market_fail_closed              → TestUnusableInputsRefuseRatherThanDefault (chain_test.go)
//
// 제외와 사유:
//
//	test_mode_resolver_defaults_off / test_flat_values_roundtrip_trade_plan_keys /
//	test_unset_trade_plan_resolves_dormant_defaults / test_parameter_registry_has_
//	trade_plan_specs / test_effective_settings_exposes_trade_plan (5)
//	  → StockOS의 trade-plan shadow 모드·설정 레지스트리. TossOS의 대응물은
//	    자동화 게이트(config.AutomationGate)와 openspec 계약이지 이 체인이 아니다.
//	test_grade_below_floor_rejected / test_grade_missing_fail_closed /
//	test_grade_at_floor_passes / test_lowercase_fail_normalised_to_grade_below_floor /
//	test_us_market_skips_grade_rung / test_sizing_markets_subset_of_grade_capable (6)
//	  → 등급(abc_grade)은 신호 계층 산물. 등급배수는 P3(배수 1.0 고정 — 보수 하한).
//	test_us_native_sizing_can_reject_size_zero_after_fx / test_us_stale_fx_reason_
//	is_audited / test_market_value_normalised / test_krx_market_identical_to_default (4)
//	  → USD-native 예산·FX 신선도는 riskcalc 소관(FXRateStaleness), 4.x 배선 후.
//	test_audit_payload_contains_contract_fields / test_audit_payload_scope_fields_us /
//	test_audit_payload_scope_fields_default_evaluated / test_us_rejected_still_
//	reports_skipped_scopes (4)
//	  → audit payload는 journal 결정 행(2a preimage)이 담당한다(4.1).
//	test_record_trade_plan_shadow_writes_audit / test_record_trade_plan_shadow_
//	swallows_failure (2)
//	  → shadow 기록 헬퍼. TossOS는 shadow 모드를 두지 않는다.
//	test_us_without_fx… 계열 3건은 위에서 이식됨.
//
// **배수 1.0에 대하여**: 원본의 사이징 예시는 등급 B 배수 0.6을 곱해 300주를
// 낸다. risk-management는 등급배수를 P3로 미루고 **1.0 고정(보수 하한)**을
// 규정하므로, 같은 입력의 TossOS 기대값은 500주다. 원본이 지키던 성질(단일
// 내림)은 TestRiskBasedQuantityFloorsExactlyOnce가 그대로 지킨다.

import (
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

func TestStrategyEntryQuantityUsesExactMinimumOfGuardianCaps(t *testing.T) {
	tests := []struct {
		name   string
		policy Policy
		entry  string
		stop   string
		want   string
	}{
		{name: "risk budget", policy: func() Policy {
			p := DefaultPolicy()
			p.RiskBudget = krw("2500")
			p.MaxOrderQuantity = "1000"
			return p
		}(), entry: "100", stop: "90", want: "250"},
		{name: "default quantity cap", policy: DefaultPolicy(), entry: "100", stop: "99", want: "100"},
		{name: "notional floor", policy: func() Policy {
			p := DefaultPolicy()
			p.MaxOrderQuantity = "1000"
			p.MaxOrderNotional = krw("1000")
			return p
		}(), entry: "101", stop: "100", want: "9"},
		{name: "exact notional boundary", policy: func() Policy {
			p := DefaultPolicy()
			p.MaxOrderQuantity = "1000"
			p.MaxOrderNotional = krw("1000")
			return p
		}(), entry: "100", stop: "99", want: "10"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := StrategyEntryQuantity(tc.policy, tc.entry, tc.stop)
			if err != nil || got != tc.want {
				t.Fatalf("quantity=%q err=%v want=%q", got, err, tc.want)
			}
		})
	}
}

func TestStrategyEntryQuantityRefusesZeroCapacity(t *testing.T) {
	policy := DefaultPolicy()
	policy.MaxOrderNotional = krw("1")
	for _, tc := range []struct{ entry, stop string }{{"100", "99"}, {"100", "100"}} {
		if got, err := StrategyEntryQuantity(policy, tc.entry, tc.stop); got != "" || !errors.Is(err, ErrStrategyQuantityZero) {
			t.Fatalf("entry/stop=%s/%s quantity=%q err=%v", tc.entry, tc.stop, got, err)
		}
	}
}

// --- No Stop = No Trade -------------------------------------------------------

func TestStopMissingIsRefusedBeforeSizing(t *testing.T) {
	// risk-management: 손절가가 없거나 보호적이지 않은 진입 의도는 **수량 계산
	// 이전 단계에서** 거부되어야 한다(SHALL). The intent below is also
	// unsizeable (quantity 0); the stop is what is reported, which is only
	// possible if the stop check ran first.
	in := entryInput()
	in.Intent.StopPrice = ""
	in.Intent.Quantity = "0"
	requireRefused(t, in, ReasonStopMissing)

	// Whitespace is absence, not a number.
	in.Intent.StopPrice = "   "
	requireRefused(t, in, ReasonStopMissing)
}

func TestStopAtOrAboveEntryIsRefused(t *testing.T) {
	// long-only: 진입은 매수이고 보호적 손절은 stop < entry. A stop at the entry
	// protects nothing, and a stop above it is a market order to lose money.
	for _, stop := range []string{"10000", "10500"} {
		in := entryInput()
		in.Intent.StopPrice = stop
		requireRefused(t, in, ReasonStopNotBelowEntry)
	}
}

func TestTargetAtOrBelowEntryIsRefused(t *testing.T) {
	for _, target := range []string{"10000", "9900"} {
		in := entryInput()
		in.Intent.TargetPrice = target
		requireRefused(t, in, ReasonTargetNotAboveEntry)
	}
}

func TestMissingTargetIsRefused(t *testing.T) {
	// a090 test_rr_missing_fail_closed. TossOS reports the missing target at the
	// stop-contract rung rather than at the RR rung: a target that does not exist
	// cannot be compared to the entry or to break-even, so the refusal happens
	// where the value is first required. Either way it is 0 대체 금지 — the RR is
	// never treated as zero and never as passing.
	in := entryInput()
	in.Intent.TargetPrice = ""
	requireRefused(t, in, ReasonInvalidTargetStop)
}

func TestNonPositiveOrUnreadableTargetStopIsRefused(t *testing.T) {
	// Ported from the constructor invariants: zero, negative, NaN, Inf and
	// non-numeric are all INVALID_TARGET_STOP.
	//
	// One deliberate difference from a090: a **zero** stop is INVALID_TARGET_STOP
	// here, not STOP_MISSING. a090 collapses the two because its input is a
	// float where 0.0 is the "unset" sentinel; TossOS's input is a decimal
	// string, so "absent" ("") and "present but not a price" ("0") are
	// distinguishable — and distinguishing them tells an operator whether the
	// signal produced nothing or produced nonsense.
	bad := []string{"0", "-1", "nan", "inf", "-inf", "n/a", "1e", ""}
	for _, value := range bad {
		if value != "" { // "" on the stop is STOP_MISSING, tested above
			in := entryInput()
			in.Intent.StopPrice = value
			requireRefused(t, in, ReasonInvalidTargetStop)
		}
		in := entryInput()
		in.Intent.TargetPrice = value
		requireRefused(t, in, ReasonInvalidTargetStop)
	}
}

func TestTheChainNeverRepairsATargetOrStop(t *testing.T) {
	// StockOS's "guardian does not compute target/stop": a malformed value is
	// refused with the chain's own code, never silently replaced with a computed
	// one. Guardian audits the contract; strategy owns the numbers — and in
	// TossOS there is no strategy layer at all yet (P3), which makes a repaired
	// value a value nobody chose.
	in := entryInput()
	in.Intent.StopPrice = "nan"
	got := requireRefused(t, in, ReasonInvalidTargetStop)
	if in.Intent.StopPrice != "nan" {
		t.Fatalf("the chain rewrote the stop to %q", in.Intent.StopPrice)
	}
	if !strings.Contains(got.Detail, "stop") {
		t.Fatalf("detail %q does not name which side of the contract failed", got.Detail)
	}
}

func TestStopContractRunsBeforeTheSizeLimit(t *testing.T) {
	// The audit trail should show the most fundamental violation. An intent that
	// is both unaffordable and contract-breaking reports the contract.
	in := entryInput()
	in.Intent.TargetPrice = "inf"
	in.Intent.Quantity = "100000"
	in.Account.CashAvailable = krw("1")
	requireRefused(t, in, ReasonInvalidTargetStop)
}

// --- 실질 본전 ------------------------------------------------------------------

func TestTargetBelowBreakEvenIsRefused(t *testing.T) {
	// A target above the entry but below the cost-inclusive break-even describes
	// a trade that loses money when it *works*.
	in := entryInput()
	in.Intent.TargetPrice = "10001"
	got := requireRefused(t, in, ReasonTargetBelowBreakEven)
	if !strings.Contains(got.Detail, "10001") {
		t.Fatalf("detail %q does not show the target against the break-even", got.Detail)
	}
}

func TestBreakEvenUsesTheInjectedCostModel(t *testing.T) {
	// trade-analytics: Guardian의 현금·비용 검증, exit 정책의 실질 본전, 성과의
	// 비용 차감이 같은 모델을 사용한다(SHALL — 이중 정의 금지). Raising the
	// injected rates must move this verdict; if it did not, the chain would be
	// reading a second, hard-coded cost model.
	expensive, err := costs.NewModel(map[string]string{
		costs.KeyKRBuyCommissionRate:  "0.05",
		costs.KeyKRSellCommissionRate: "0.05",
		costs.KeyKRSellTaxRate:        "0.05",
	})
	if err != nil {
		t.Fatalf("NewModel: %v", err)
	}
	in := entryInput()
	requireAllowed(t, in) // 10,500 clears the placeholder break-even (≈10,050)
	in.Costs = expensive
	requireRefused(t, in, ReasonTargetBelowBreakEven) // …but not ≈11,667
}

// --- 위험 기반 수량 ---------------------------------------------------------------

func TestRiskBasedQuantity(t *testing.T) {
	// floor(위험예산 / (entry − stop)), 등급배수 1.0 고정.
	cases := []struct {
		name                string
		budget, entry, stop string
		want                string
	}{
		// a090's spec example, with the grade multiplier at 1.0: 100,000 ÷ 200.
		{"spec example at multiplier 1.0", "100000", "10000", "9800", "500"},
		// a090 test_a_grade_full_size: the top tier was already ×1.0, so it is
		// the one case whose original expectation survives unchanged.
		{"full size", "100000", "10000", "9800", "500"},
		{"budget smaller than one unit of risk", "100", "10000", "9800", "0"},
		{"budget exactly one unit of risk", "200", "10000", "9800", "1"},
		{"fractional prices", "1000", "8.00", "7.88", "8333"},
		// stop 폭 0 이하는 수량 0(fail-closed) — the stop contract refuses these
		// before sizing, but the sizing rule itself must not divide by zero or
		// return a negative quantity if some later caller reaches it directly.
		{"zero width", "100000", "10000", "10000", "0"},
		{"inverted width", "100000", "10000", "10200", "0"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := RiskBasedQuantity(tc.budget, tc.entry, tc.stop)
			if err != nil {
				t.Fatalf("RiskBasedQuantity: %v", err)
			}
			if got != tc.want {
				t.Fatalf("RiskBasedQuantity(%s, %s, %s) = %s, want %s",
					tc.budget, tc.entry, tc.stop, got, tc.want)
			}
		})
	}
}

func TestRiskBasedQuantityFloorsExactlyOnce(t *testing.T) {
	// a090 리뷰 M1: 계약 공식 = (예산 × 배수) ÷ 폭 → 내림 한 번. Two floors on the
	// same quotient lose a unit whenever the intermediate is not an integer, and
	// the difference is a whole share of exposure the sizing rule did not intend
	// to drop. With the multiplier fixed at 1.0 the second floor has nowhere to
	// hide, and this case pins the exact quotient rather than a rounded one.
	got, err := RiskBasedQuantity("100000", "10000", "9700")
	if err != nil {
		t.Fatalf("RiskBasedQuantity: %v", err)
	}
	if got != "333" { // 100000/300 = 333.33…
		t.Fatalf("quantity = %s, want 333", got)
	}

	// The arithmetic is exact rather than float: 0.1-wide risk on a budget that
	// a binary float would round below the boundary still yields the full count.
	got, err = RiskBasedQuantity("3", "1.3", "1.0")
	if err != nil {
		t.Fatalf("RiskBasedQuantity: %v", err)
	}
	if got != "10" { // 3 / 0.3 = exactly 10; float64 gives 9.999999999999998
		t.Fatalf("quantity = %s, want 10 — the quotient was computed in binary floating point", got)
	}
}

func TestRiskBasedQuantityRefusesUnreadableInputs(t *testing.T) {
	for _, tc := range [][3]string{
		{"", "10000", "9800"},
		{"100000", "", "9800"},
		{"100000", "10000", ""},
		{"abc", "10000", "9800"},
		{"-100000", "10000", "9800"},
		{"100000", "nan", "9800"},
	} {
		if got, err := RiskBasedQuantity(tc[0], tc[1], tc[2]); err == nil {
			t.Fatalf("RiskBasedQuantity(%q, %q, %q) = %s with no error", tc[0], tc[1], tc[2], got)
		}
	}
}

func TestQuantityAboveTheRiskBudgetIsRefused(t *testing.T) {
	// The chain does not trust the issuer's sizing: it recomputes the cap from
	// the stop it just validated. Budget 100,000 over a 2,000-wide stop is 50
	// shares; 51 is refused even though it is inside every configured per-order
	// cap (51 ≤ 100 shares, 510,000 ≤ 1,000,000 KRW).
	in := entryInput()
	in.Intent.StopPrice = "8000"
	in.Intent.TargetPrice = "14000"
	in.Intent.Quantity = "51"
	in.Account.CashAvailable = krw("2000000")
	got := requireRefused(t, in, ReasonInvalidOrderSize)
	if !strings.Contains(got.Detail, "50") {
		t.Fatalf("detail %q does not show the cap the quantity exceeded", got.Detail)
	}

	in.Intent.Quantity = "50"
	requireAllowed(t, in)
}

func TestSizeZeroIsRefused(t *testing.T) {
	// a090 test_size_zero_rejected (손절폭이 리스크 예산보다 큼 → 1주도 못 삼) and
	// test_candidate_rejects_zero_quantity, which land on the same rung here.
	in := entryInput()
	in.Intent.Quantity = "0"
	requireRefused(t, in, ReasonInvalidOrderSize)

	in = entryInput()
	in.Intent.Quantity = "-1"
	requireRefused(t, in, ReasonInvalidOrderSize)

	in = entryInput()
	in.Intent.Quantity = "9.5" // a fractional share is not an order this places
	requireRefused(t, in, ReasonInvalidOrderSize)

	// The budget cannot buy one unit of risk: the cap is zero, so any quantity
	// is above it.
	in = entryInput()
	in.Policy.RiskBudget = krw("100")
	requireRefused(t, in, ReasonInvalidOrderSize)
}

func TestUnusableEntryPriceIsInputUnavailable(t *testing.T) {
	// a090 test_inputs_unavailable_on_bad_entry. The entry price is the anchor
	// every other check is relative to, so an unusable one is an input failure
	// rather than a contract violation — reporting STOP_NOT_BELOW_ENTRY for an
	// entry of 0 would send an operator to the wrong end of the pipeline.
	for _, price := range []string{"0", "-1", "", "abc", "inf"} {
		in := entryInput()
		in.Intent.LimitPrice = price
		requireRefused(t, in, ReasonInputUnavailable)
	}
}

// --- 최소 RR --------------------------------------------------------------------

func TestDefaultMinimumRewardRiskIsTwo(t *testing.T) {
	// risk-management: 기본값 2.0 미달·계산 불가는 거부한다(SHALL — 0 대체 금지.
	// provenance: StockOS parker_vwap §22 lock 2.0; 1.5는 최저 티어 값이라 기각).
	if got := DefaultPolicy().MinRewardRisk; got != DefaultMinRewardRisk {
		t.Fatalf("default policy carries %q, want %q", got, DefaultMinRewardRisk)
	}
	if DefaultMinRewardRisk != "2.0" {
		t.Fatalf("DefaultMinRewardRisk = %q, want 2.0 — lowering it is a §0.9 relaxation",
			DefaultMinRewardRisk)
	}
}

func TestRewardRiskBelowTheMinimumIsRefused(t *testing.T) {
	// (target − entry) / (entry − stop) = 300/200 = 1.5 — a090's rejected value,
	// and the tier TossOS explicitly declined to adopt.
	in := entryInput()
	in.Intent.TargetPrice = "10300"
	got := requireRefused(t, in, ReasonMinRRNotMet)
	if !strings.Contains(got.Detail, "1.5") || !strings.Contains(got.Detail, "2.0") {
		t.Fatalf("detail %q does not show the ratio against the minimum", got.Detail)
	}
}

func TestRewardRiskAtTheMinimumPasses(t *testing.T) {
	// 미달 is below, not "not above": exactly 2.0 passes. The arithmetic is
	// exact, so this boundary is decided by the rule and not by rounding.
	in := entryInput()
	in.Intent.TargetPrice = "10400" // 400/200 = exactly 2.0
	requireAllowed(t, in)

	in.Intent.TargetPrice = "10399.99" // 1.99995
	requireRefused(t, in, ReasonMinRRNotMet)
}

func TestRewardRiskIsExactArithmetic(t *testing.T) {
	cases := []struct {
		entry, stop, target string
		want                string
	}{
		{"10000", "9800", "10500", "5/2"},
		{"10000", "9700", "10480", "8/5"},
		{"8.00", "7.88", "8.30", "5/2"},
		// A ratio a float64 would land just under 2: 0.1-scale prices.
		{"1.3", "1.0", "1.9", "2"},
	}
	for _, tc := range cases {
		got, err := RewardRisk(tc.entry, tc.stop, tc.target)
		if err != nil {
			t.Fatalf("RewardRisk(%s, %s, %s): %v", tc.entry, tc.stop, tc.target, err)
		}
		want, ok := new(big.Rat).SetString(tc.want)
		if !ok {
			t.Fatalf("bad expectation %q", tc.want)
		}
		if got.Cmp(want) != 0 {
			t.Fatalf("RewardRisk(%s, %s, %s) = %s, want %s",
				tc.entry, tc.stop, tc.target, got.RatString(), want.RatString())
		}
	}
}

func TestRewardRiskRefusesRatherThanReturningZero(t *testing.T) {
	// SHALL — 0 대체 금지. A ratio that cannot be computed is not a ratio of
	// zero: zero would be a *measured* failure, and the two must not be
	// indistinguishable in an audit record.
	for _, tc := range [][3]string{
		{"10000", "10000", "10500"}, // zero-wide stop
		{"10000", "10200", "10500"}, // inverted stop
		{"10000", "9800", ""},       // no target
		{"10000", "9800", "abc"},
		{"", "9800", "10500"},
	} {
		if got, err := RewardRisk(tc[0], tc[1], tc[2]); err == nil {
			t.Fatalf("RewardRisk(%q, %q, %q) = %s with no error", tc[0], tc[1], tc[2], got.RatString())
		}
	}

	// And the chain maps that refusal to MIN_RR_NOT_MET rather than to a pass.
	// (The stop contract normally catches a zero-wide stop first, so the step is
	// called directly here — the mapping must exist even if it is unreachable
	// through the current order.)
	in := entryInput()
	in.Intent.StopPrice = "10000"
	got := checkMinRewardRisk(in)
	if got.Allowed || got.Reason != ReasonMinRRNotMet {
		t.Fatalf("uncomputable RR gave %+v, want a %s refusal", got, ReasonMinRRNotMet)
	}
}

// --- long-only ------------------------------------------------------------------

func TestShortExposureIsStructurallyImpossible(t *testing.T) {
	// risk-management: short 노출은 구조적으로 금지된다(SHALL NOT). There is no
	// "sell to open": a SELL is a reduction, and a reduction is bounded by what
	// the account holds.
	in := reductionInput()
	in.Account.HeldQuantity = "0"
	in.Intent.Quantity = "1"
	requireRefused(t, in, ReasonSellExceedsHoldings)
}

// --- policy validation ------------------------------------------------------------

func TestPolicyRejectsInvalidThresholds(t *testing.T) {
	// a090 test_config_rejects_invalid_thresholds.
	cases := []struct {
		name string
		mut  func(*Policy)
	}{
		{"zero minimum RR", func(p *Policy) { p.MinRewardRisk = "0" }},
		{"negative minimum RR", func(p *Policy) { p.MinRewardRisk = "-1" }},
		{"unreadable minimum RR", func(p *Policy) { p.MinRewardRisk = "two" }},
		{"negative risk budget", func(p *Policy) { p.RiskBudget = krw("-1") }},
		{"risk budget without currency", func(p *Policy) { p.RiskBudget = riskcalc.Money{Amount: "100000"} }},
		{"zero order quantity cap", func(p *Policy) { p.MaxOrderQuantity = "0" }},
		{"unset notional cap", func(p *Policy) { p.MaxOrderNotional = riskcalc.Money{} }},
		{"mixed limit currencies", func(p *Policy) { p.MaxDailyLoss = riskcalc.Money{Amount: "100000", Currency: "USD"} }},
		{"daily loss ratio above one", func(p *Policy) { p.MaxDailyLossRatio = "1.5" }},
		{"negative cooldown", func(p *Policy) { p.Reentry.Cooldown = -1 }},
		{"zero same-day entries", func(p *Policy) { p.Reentry.MaxEntriesPerSymbol = 0 }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := DefaultPolicy()
			tc.mut(&p)
			if err := p.Validate(); err == nil {
				t.Fatal("Validate accepted the policy")
			}
		})
	}

	if err := DefaultPolicy().Validate(); err != nil {
		t.Fatalf("the conservative default policy does not validate: %v", err)
	}
}

func TestDefaultPolicyIsTheConservativeDefaultSet(t *testing.T) {
	// risk-management: 사용자 미확정 시 보수 기본값 전체 집합을 사용한다(SHALL —
	// 인터록 5필드를 전부 충족): 주문당 notional 1,000,000 KRW·주문당 수량 100주·
	// 총 노출 10,000,000 KRW·일일 손실 100,000 KRW·일일 손실 자본비 1%·통화 KRW.
	p := DefaultPolicy()
	for _, tc := range []struct{ name, got, want string }{
		{"per-order notional", p.MaxOrderNotional.Amount, "1000000"},
		{"per-order quantity", p.MaxOrderQuantity, "100"},
		{"total exposure", p.MaxOpenExposure.Amount, "10000000"},
		{"daily loss", p.MaxDailyLoss.Amount, "100000"},
		{"daily loss ratio", p.MaxDailyLossRatio, "0.01"},
		{"currency", p.MaxOrderNotional.Currency, "KRW"},
	} {
		if tc.got != tc.want {
			t.Errorf("%s = %q, want %q", tc.name, tc.got, tc.want)
		}
	}
	// 재진입: 쿨다운 기본 30분 `[미검증]`·당일 심볼당 최대 진입 수 기본 2회 `[미검증]`.
	if p.Reentry.MaxEntriesPerSymbol != 2 {
		t.Errorf("same-day entries = %d, want 2", p.Reentry.MaxEntriesPerSymbol)
	}
	if p.Reentry.Cooldown.Minutes() != 30 {
		t.Errorf("cooldown = %s, want 30m", p.Reentry.Cooldown)
	}
}

// --- the reason vocabulary ---------------------------------------------------------

func TestEveryReasonCodeIsReachableAndStable(t *testing.T) {
	// The codes are written into journal rows and operator alerts, so the set is
	// a contract: renaming one rewrites history. This pins both the spelling and
	// that nothing was added without being listed.
	want := map[ReasonCode]bool{
		ReasonAllowed:                      true,
		ReasonKillSwitchActive:             true,
		ReasonOperatingModeBlocked:         true,
		ReasonEntryGateBlocked:             true,
		ReasonSymbolNotAllowed:             true,
		ReasonStopMissing:                  true,
		ReasonStopNotBelowEntry:            true,
		ReasonTargetNotAboveEntry:          true,
		ReasonInvalidTargetStop:            true,
		ReasonTargetBelowBreakEven:         true,
		ReasonSellExceedsHoldings:          true,
		ReasonInvalidOrderSize:             true,
		ReasonMaxOrderExceeded:             true,
		ReasonMinRRNotMet:                  true,
		ReasonInsufficientCash:             true,
		ReasonPendingBuyOrderBlocked:       true,
		ReasonSameDayReentryBlocked:        true,
		ReasonSameDayReentryCooldownActive: true,
		ReasonOpenExposureExceeded:         true,
		ReasonDailyLossLimitReached:        true,
		ReasonDuplicateOrder:               true,
		ReasonInputUnavailable:             true,
	}
	got := AllReasonCodes()
	if len(got) != len(want) {
		t.Fatalf("AllReasonCodes has %d entries, this case lists %d: %v", len(got), len(want), got)
	}
	seen := map[ReasonCode]bool{}
	for _, code := range got {
		if !want[code] {
			t.Errorf("%q is not in the expected set", code)
		}
		if seen[code] {
			t.Errorf("%q appears twice", code)
		}
		seen[code] = true
		if code != ReasonCode(strings.ToUpper(string(code))) {
			t.Errorf("%q is not upper-snake; the spec's vocabulary is", code)
		}
	}
}
