package config

// limits.go is the Guardian tier registry, its ceiling backstop, and the
// validation the console's limit editor is held to (change
// console-sets-guardian-limits, tasks 2.x).
//
// # Where the numbers come from
//
// This is a port of StockOS's packages/trading/stockos_trading/risk_profiles.py
// — the (market, profile) → RiskProfile table. The numbers are transcribed
// rather than re-derived so an audit can put the two files side by side, which
// is what risk-management's 정책 수치의 provenance asks for: "모든 한도·정책
// 수치는 코드에 출처(StockOS 파일·검증 상태)와 함께 기록되어야 한다".
//
// TossOS spells currency where StockOS spells market. StockOS pins KRX→KRW and
// NASD/NYSE/AMEX→USD in its own MarketMeta registry and its US tiers are
// identical across the three US venues, so the currency axis carries every
// distinction the market axis carried. The automation gate has one currency
// (config.AutomationGate.LimitCurrency), so this is also the axis the gate can
// actually express.
//
// # What is NOT ported
//
// StockOS registers a third KRX tier, paper_demo, sized for a KIS demo
// account's 5억 seed — its own docstring says so, and says why it is not
// mirrored onto the US markets. TossOS has no such seed on the official Open
// API. Porting it would do one thing beyond adding an unusable row: by the
// max-across-tiers rule below it would become the KRW ceiling, raising the
// most permissive KRW order from 1,000,000 to 5,000,000. That is a loosening
// with no evidence behind it (§0.6).

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// GuardianLimits is the automation gate's five ceilings and the currency the
// money ones are expressed in.
//
// It is a separate type from AutomationGate on purpose, and the separation is
// load-bearing: it is the shape the console's editor seam accepts, and it has
// no `enabled` field. The console cannot send the switch because there is no
// place in the message to put it (design D6/D7).
type GuardianLimits struct {
	MaxOrderQuantity   float64
	MaxOrderNotional   float64
	MaxTotalExposure   float64
	MaxDailyLossAmount float64
	MaxDailyLossRatio  float64
	Currency           string
}

// GuardianTier is one registered set of ceilings.
type GuardianTier struct {
	// ID is the stable identifier a form posts back.
	ID string
	// Label is what the screen shows.
	Label string
	// Limits are the ceilings this tier installs.
	Limits GuardianLimits
}

// guardianTiers is the registry, in display order.
//
// The per-tier provenance comment names the StockOS symbol the row transcribes.
// The quantity column is the one value with no StockOS counterpart — see
// guardianOrderQuantityCap.
var guardianTiers = []GuardianTier{
	{
		ID:    "kr-smoke",
		Label: "국내 스모크 — 주문 하나가 실패해도 노출에 해가 없는 크기",
		// StockOS risk_profiles.py _KR_SMOKE.
		Limits: GuardianLimits{
			MaxOrderQuantity:   guardianOrderQuantityCap,
			MaxOrderNotional:   500_000,
			MaxTotalExposure:   5_000_000,
			MaxDailyLossAmount: 100_000,
			MaxDailyLossRatio:  0.01,
			Currency:           "KRW",
		},
	},
	{
		ID:    "kr-small-live",
		Label: "국내 소액 실거래 (권장 기본값)",
		// StockOS risk_profiles.py _KR_SMALL_LIVE. These five are also the set
		// risk-management's 정책 수치의 provenance approved verbatim: 주문당
		// notional 1,000,000 KRW·주문당 수량 100주·총 노출 10,000,000 KRW·일일
		// 손실 100,000 KRW·일일 손실 자본비 1%·통화 KRW.
		Limits: GuardianLimits{
			MaxOrderQuantity:   guardianOrderQuantityCap,
			MaxOrderNotional:   1_000_000,
			MaxTotalExposure:   10_000_000,
			MaxDailyLossAmount: 100_000,
			MaxDailyLossRatio:  0.01,
			Currency:           "KRW",
		},
	},
	{
		ID:    "us-smoke",
		Label: "미국 스모크 — 국내 자동 진입은 통화 불일치로 닫힌다",
		// StockOS risk_profiles.py _US_SMOKE ($100 / $300 / $10, 1%).
		Limits: GuardianLimits{
			MaxOrderQuantity:   guardianOrderQuantityCap,
			MaxOrderNotional:   100,
			MaxTotalExposure:   300,
			MaxDailyLossAmount: 10,
			MaxDailyLossRatio:  0.01,
			Currency:           "USD",
		},
	},
	{
		ID:    "us-small-live",
		Label: "미국 소액 실거래 — 국내 자동 진입은 통화 불일치로 닫힌다",
		// StockOS risk_profiles.py _US_SMALL_LIVE ($300 / $1,000 / $50, 1%).
		Limits: GuardianLimits{
			MaxOrderQuantity:   guardianOrderQuantityCap,
			MaxOrderNotional:   300,
			MaxTotalExposure:   1_000,
			MaxDailyLossAmount: 50,
			MaxDailyLossRatio:  0.01,
			Currency:           "USD",
		},
	},
}

// guardianOrderQuantityCap is the share-count ceiling, common to every tier.
//
// StockOS bounds order size by money alone (max_order_krw); there is no
// counterpart to copy. TossOS's startup interlock requires a quantity bound as
// one of its five fields, so this number is TossOS's own — and its source is
// risk-management's approved set, which names 주문당 수량 100주.
//
// It does not vary by tier, and that is the decision rather than an omission.
// A quantity ceiling is a fat-finger backstop, not a sizing rule: the risk
// differences between tiers are already carried in full by the three money
// axes. Scaling it per tier would mean assuming a nominal share price and
// dividing — StockOS wrote such a price once in prose ("1 lot of a $100
// stock"), which is a rationale, not a number this code may multiply by. A
// ceiling computed from a price nobody observed is exactly the kind of figure
// this repository refuses elsewhere.
//
// The consequence is that the money axis binds first in almost every case (100
// shares stay under the USD smoke tier's $100 only below $1 a share). That is
// what a backstop is: silent on ordinary days, and there on the day a decimal
// point moves.
const guardianOrderQuantityCap = 100

// GuardianTiers returns the registry in display order.
func GuardianTiers() []GuardianTier {
	out := make([]GuardianTier, len(guardianTiers))
	copy(out, guardianTiers)
	return out
}

// GuardianTierByID looks one tier up.
func GuardianTierByID(id string) (GuardianTier, bool) {
	for _, tier := range guardianTiers {
		if tier.ID == id {
			return tier, true
		}
	}
	return GuardianTier{}, false
}

// DefaultGuardianTierID names the tier the screen recommends.
//
// It is not applied implicitly anywhere. The console offers it and a click
// writes it into the file explicitly, so the number the screen shows and the
// number the engine's interlock reads are the same bytes (design D1).
func DefaultGuardianTierID() string { return "kr-small-live" }

// normaliseCurrency is the one spelling rule: trimmed and upper-cased.
func normaliseCurrency(s string) string {
	return strings.ToUpper(strings.TrimSpace(s))
}

// GuardianCeiling is the per-field maximum across every tier registered for a
// currency — StockOS ADR §2.6's _market_ceiling, which its own docstring calls
// the absolute backstop.
//
// A currency with no registered tier has no ceiling, and no ceiling means no
// backstop, so it is an error rather than an unbounded pass. StockOS fails
// closed here for the same reason and says so.
func GuardianCeiling(currency string) (GuardianLimits, error) {
	name := normaliseCurrency(currency)
	if name == "" {
		return GuardianLimits{}, errors.New(
			"config: a limit block names no currency, so it has no registered ceiling")
	}
	ceiling := GuardianLimits{Currency: name}
	found := false
	for _, tier := range guardianTiers {
		if normaliseCurrency(tier.Limits.Currency) != name {
			continue
		}
		found = true
		ceiling.MaxOrderQuantity = math.Max(ceiling.MaxOrderQuantity, tier.Limits.MaxOrderQuantity)
		ceiling.MaxOrderNotional = math.Max(ceiling.MaxOrderNotional, tier.Limits.MaxOrderNotional)
		ceiling.MaxTotalExposure = math.Max(ceiling.MaxTotalExposure, tier.Limits.MaxTotalExposure)
		ceiling.MaxDailyLossAmount = math.Max(ceiling.MaxDailyLossAmount, tier.Limits.MaxDailyLossAmount)
		ceiling.MaxDailyLossRatio = math.Max(ceiling.MaxDailyLossRatio, tier.Limits.MaxDailyLossRatio)
	}
	if !found {
		return GuardianLimits{}, fmt.Errorf(
			"config: %s has no registered Guardian tier, so there is no ceiling to check a limit "+
				"against; the limit editor is fail-closed", name)
	}
	return ceiling, nil
}

// GuardianCurrencies lists the currencies that have a registered tier, in
// display order, so a refusal can name what the operator may actually pick
// instead of hardcoding the answer at the call site.
func GuardianCurrencies() []string {
	var out []string
	seen := map[string]bool{}
	for _, tier := range guardianTiers {
		name := normaliseCurrency(tier.Limits.Currency)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		out = append(out, name)
	}
	return out
}

// guardianFields is the field set, spelled once. Every per-field walk in this
// file goes through it, so a sixth ceiling cannot be added to GuardianLimits and
// silently skip the cap check.
func guardianFields() []struct {
	Key string
	Get func(GuardianLimits) float64
} {
	return []struct {
		Key string
		Get func(GuardianLimits) float64
	}{
		{"max_order_quantity", func(l GuardianLimits) float64 { return l.MaxOrderQuantity }},
		{"max_order_notional", func(l GuardianLimits) float64 { return l.MaxOrderNotional }},
		{"max_total_exposure", func(l GuardianLimits) float64 { return l.MaxTotalExposure }},
		{"max_daily_loss_amount", func(l GuardianLimits) float64 { return l.MaxDailyLossAmount }},
		{"max_daily_loss_ratio", func(l GuardianLimits) float64 { return l.MaxDailyLossRatio }},
	}
}

// Validate reports why the startup interlock would refuse this block.
//
// It mirrors execgw.Limits.Validate field for field, because the console's
// contract is that it never records a block the engine would refuse to start
// on. The two live in different packages — internal/execgw reaches this one
// transitively, so it cannot be imported here — and the agreement between them
// is asserted over a generated corpus in
// internal/app/engine/limits_equivalence_test.go.
//
// Deliberately NOT checked here: whether the currency has a registered tier.
// The interlock does not care, so neither may this, or the two verdicts diverge.
// That check belongs to the ceiling, which is this package's own rule and is
// applied at the write path.
func (l GuardianLimits) Validate() error {
	for _, f := range guardianFields() {
		v := f.Get(l)
		switch {
		case math.IsNaN(v) || math.IsInf(v, 0):
			return fmt.Errorf("%s is not a finite number", f.Key)
		case v <= 0:
			return fmt.Errorf(
				"%s is %v; a limit that authorises nothing is not a limit", f.Key, v)
		}
	}
	if l.MaxDailyLossRatio > 1 {
		return fmt.Errorf(
			"max_daily_loss_ratio is %v; a ratio above 1 bounds nothing", l.MaxDailyLossRatio)
	}
	if strings.TrimSpace(l.Currency) == "" {
		return errors.New("the limit block names no currency, so its money bounds mean nothing")
	}
	return nil
}

// CeilingViolations lists the fields raised past the registered ceiling —
// StockOS ADR §2.6's real_safety_gate_ceiling_violations. An empty list means
// the block is acceptable.
//
// The boundary is inclusive: every tier sits at the ceiling on at least one
// field, so an exclusive comparison would make the presets unsaveable.
func (l GuardianLimits) CeilingViolations() []string {
	ceiling, err := GuardianCeiling(l.Currency)
	if err != nil {
		return []string{err.Error()}
	}
	var violations []string
	for _, f := range guardianFields() {
		value, limit := f.Get(l), f.Get(ceiling)
		if math.IsNaN(value) || value <= limit {
			continue
		}
		violations = append(violations, fmt.Sprintf(
			"%s=%s는 %s 상한 %s를 넘는다", f.Key, formatLimit(value),
			ceiling.Currency, formatLimit(limit)))
	}
	return violations
}

// MatchingGuardianTier names the tier this block spells exactly, or "".
//
// It is a comparison against the registry, not a claim about where the numbers
// came from. Somebody may have typed the same five values by hand, and nothing
// in the file distinguishes that from a preset click — so the screen says "이
// 티어와 일치", never "이 티어를 골랐다" (design D12).
func MatchingGuardianTier(l GuardianLimits) string {
	for _, tier := range guardianTiers {
		want := tier.Limits
		if normaliseCurrency(l.Currency) != normaliseCurrency(want.Currency) {
			continue
		}
		same := true
		for _, f := range guardianFields() {
			if f.Get(l) != f.Get(want) {
				same = false
				break
			}
		}
		if same {
			return tier.ID
		}
	}
	return ""
}

// formatLimit renders a ceiling without exponent notation or a trailing ".0",
// so a message names the number the operator typed.
func formatLimit(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}
