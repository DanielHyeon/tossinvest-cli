package console

// settings_limits.go is the Guardian-limit editor (change
// console-sets-guardian-limits, task 5.x): a section on the settings screen, a
// one-click preset apply, and an advanced form behind a fold.
//
// # What this surface may write, and how that is enforced
//
// The five ceilings and the currency. Not `enabled`, not the kill switch — the
// operator-console spec keeps the §0.7 switch outside the console, and this
// change separates the switch from the numbers rather than loosening the rule.
//
// The enforcement is the shape of the seam, not the care of this file: Save
// takes config.GuardianLimits, and that type has no field for the gate. A
// handler here cannot flip the gate by mistake, by race, or by a future edit,
// because it has nowhere to put the value. config's writer closes the same door
// from the other side by splicing six named keys and never emitting `enabled`
// (design D6).
//
// # Why there is no implicit default
//
// The registry's recommended tier is offered as a preset, and a click writes it
// into the file. Nothing here fills the five values in silently. A screen that
// displayed 1,000,000 for a file that contains nothing would disagree with the
// engine's startup interlock, which reads the file — and the operator would be
// looking at a configured gate while the engine refused to start on an
// unconfigured one (design D1).

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
)

// LimitSettings is the console's write surface for the automation gate's
// ceilings. Exactly two methods — settings_static_test.go fails if it grows,
// and settings_limits_test.go fails if Save ever accepts a type that can carry
// the switch.
type LimitSettings interface {
	// Load returns the gate block as written in the file. It includes `enabled`
	// because the screen must say whether the gate is on; the write side cannot
	// accept it back.
	Load() (config.AutomationGate, error)
	// Save validates against the startup interlock's rules and the registered
	// ceiling, then writes the six values and nothing else.
	Save(config.GuardianLimits) error
}

// limitRow is one ceiling as the screen renders it.
type limitRow struct {
	Label string
	// Value is the number as written, or 미설정. It is never a default.
	Value string
	// Unit is the currency for the money rows, empty otherwise.
	Unit string
}

// LimitRows renders the five ceilings the file currently spells.
func (p settingsPage) LimitRows() []limitRow {
	currency := strings.TrimSpace(p.Gate.LimitCurrency)
	return []limitRow{
		{"1회 주문 수량 상한", limitValueText(p.Gate.MaxOrderQuantity), "주"},
		{"1회 주문 금액 상한", limitValueText(p.Gate.MaxOrderNotional), currency},
		{"계좌 개방 노출 상한", limitValueText(p.Gate.MaxTotalExposure), currency},
		{"일일 손실 한도(금액)", limitValueText(p.Gate.MaxDailyLossAmount), currency},
		{"일일 손실 한도(비율)", limitRatioText(p.Gate.MaxDailyLossRatio), ""},
	}
}

// limitValueText is the two-state rendering the whole design rests on: a
// number, or the honest absence of one.
func limitValueText(v float64) string {
	if v <= 0 {
		return "미설정"
	}
	return decimalText(v)
}

// limitRatioText renders the capital fraction as a percentage, which is how the
// tier table and StockOS both talk about it.
func limitRatioText(v float64) string {
	if v <= 0 {
		return "미설정"
	}
	return strconv.FormatFloat(v*100, 'f', -1, 64) + "%"
}

// The advanced form's field values.
//
// Two rules, both learned the hard way in review. An unset ceiling renders as
// the EMPTY string, never as 0 — the table above already says 미설정 about the
// same field, and a box pre-filled with 0 says the ceiling is zero, which is the
// one reading this codebase spends its comments refusing (A1). And the number is
// formatted rather than handed to the template, because Go renders float64
// 10000000 as 1e+07; that round-trips through ParseFloat perfectly well and is
// still an invitation to retype it wrong (A2).
func (p settingsPage) FieldQuantity() string  { return limitFieldText(p.Gate.MaxOrderQuantity) }
func (p settingsPage) FieldNotional() string  { return limitFieldText(p.Gate.MaxOrderNotional) }
func (p settingsPage) FieldExposure() string  { return limitFieldText(p.Gate.MaxTotalExposure) }
func (p settingsPage) FieldDailyLoss() string { return limitFieldText(p.Gate.MaxDailyLossAmount) }
func (p settingsPage) FieldRatio() string     { return limitFieldText(p.Gate.MaxDailyLossRatio) }

// FieldCurrency is the currency as written, trimmed.
func (p settingsPage) FieldCurrency() string { return strings.TrimSpace(p.Gate.LimitCurrency) }

// limitFieldText is a plain decimal with no grouping — the form posts it back,
// and a value that survives the round trip unchanged is one fewer thing to get
// wrong.
func limitFieldText(v float64) string {
	if v <= 0 {
		return ""
	}
	return strconv.FormatFloat(v, 'f', -1, 64)
}

// LimitCurrencyText is the currency as the file spells it, or 미지정.
func (p settingsPage) LimitCurrencyText() string {
	if c := strings.TrimSpace(p.Gate.LimitCurrency); c != "" {
		return c
	}
	return "미지정"
}

// LimitsConfigured reports a block the startup interlock would accept.
func (p settingsPage) LimitsConfigured() bool {
	return p.Gate.Limits().Validate() == nil
}

// LimitsPartlyConfigured is the state design D2 refuses to repair silently:
// something is written, but not enough for the engine to start.
func (p settingsPage) LimitsPartlyConfigured() bool {
	return p.Gate.LimitsSet() && !p.LimitsConfigured()
}

// LimitsUnset is the empty block.
func (p settingsPage) LimitsUnset() bool { return !p.Gate.LimitsSet() }

// LimitVerdict is why the interlock would refuse the current block.
func (p settingsPage) LimitVerdict() string {
	if err := p.Gate.Limits().Validate(); err != nil {
		return err.Error()
	}
	return ""
}

// LimitTiers is the registry, for the preset cards.
func (p settingsPage) LimitTiers() []config.GuardianTier { return config.GuardianTiers() }

// RecommendedTier names the tier the screen marks 권장.
func (p settingsPage) RecommendedTier() string { return config.DefaultGuardianTierID() }

// MatchingTier names the registry row the file currently spells, or "".
func (p settingsPage) MatchingTier() string {
	if !p.Gate.LimitsSet() {
		return ""
	}
	return config.MatchingGuardianTier(p.Gate.Limits())
}

// TierSummary renders one tier's five values for its card.
func (t tierCard) Summary() string { return limitSummary(t.Limits) }

// tierCard is a registry row plus the screen's view of it.
type tierCard struct {
	config.GuardianTier
	// Recommended marks the default tier.
	Recommended bool
}

// LimitTierCards is what the template ranges over.
func (p settingsPage) LimitTierCards() []tierCard {
	recommended := config.DefaultGuardianTierID()
	out := make([]tierCard, 0, len(config.GuardianTiers()))
	for _, tier := range config.GuardianTiers() {
		out = append(out, tierCard{GuardianTier: tier, Recommended: tier.ID == recommended})
	}
	return out
}

// limitSummary is the one-line rendering used both on a preset card and in the
// answer after applying it, so the operator reads the same sentence before and
// after the click.
func limitSummary(l config.GuardianLimits) string {
	return "주문 " + decimalText(l.MaxOrderQuantity) + "주 · " +
		decimalText(l.MaxOrderNotional) + " " + l.Currency + " / 총 노출 " +
		decimalText(l.MaxTotalExposure) + " " + l.Currency + " / 일일 손실 " +
		decimalText(l.MaxDailyLossAmount) + " " + l.Currency +
		" (자본비 " + limitRatioText(l.MaxDailyLossRatio) + ")"
}

// currencyConsequence states what the limit currency costs.
//
// It is said on EVERY apply, not only when the currency changes. The gate has
// one currency and risk's chain refuses an intent priced in another
// ("a090's mixed-currency guard"), so choosing a currency closes a market —
// and the moment that most needs saying is the first time the operator sets
// one, which a change-detecting branch would pass over in silence.
func currencyConsequence(currency string) string {
	switch strings.ToUpper(strings.TrimSpace(currency)) {
	case "KRW":
		return "한도 통화는 KRW다 — Guardian 체인은 한도 통화와 다른 통화의 진입을 거부하므로 " +
			"미국 시장의 자동 진입은 닫힌다."
	case "USD":
		return "한도 통화는 USD다 — Guardian 체인은 한도 통화와 다른 통화의 진입을 거부하므로 " +
			"국내 시장의 자동 진입은 닫힌다."
	}
	return "한도 통화는 " + currency + "다 — Guardian 체인은 한도 통화와 다른 통화의 진입을 거부한다."
}

// handleSettingsLimitPreset applies one registered tier in a single click.
func (c *Console) handleSettingsLimitPreset(w http.ResponseWriter, r *http.Request) {
	if c.opts.Limits == nil {
		c.refuse(w, http.StatusNotImplemented, "한도 저장이 배선되지 않았다",
			"이 빌드의 콘솔에는 Guardian 한도 저장 seam이 주입되지 않았다.")
		return
	}
	id := strings.TrimSpace(r.PostFormValue("tier"))
	tier, ok := config.GuardianTierByID(id)
	if !ok {
		// Not a redirect: an unregistered id did not come from this screen, and a
		// notice would report on a choice nobody was offered.
		c.refuse(w, http.StatusBadRequest, "등록되지 않은 한도 티어",
			"요청한 티어가 레지스트리에 없다: "+id)
		return
	}
	if err := c.opts.Limits.Save(tier.Limits); err != nil {
		c.redirectSettings(w, r, "저장 안 됨 — "+err.Error())
		return
	}
	// Recorded, not applied: a running engine keeps the limits it started with,
	// and the Guardian's snapshot is taken from them. effectNotice is the one
	// place that says when this bites (design D9).
	c.redirectSettings(w, r, tier.Label+" 한도 기록됨 — "+limitSummary(tier.Limits)+". "+
		currencyConsequence(tier.Limits.Currency)+" "+effectNotice(c.engineRunning()))
}

// handleSettingsLimits saves individually entered values.
func (c *Console) handleSettingsLimits(w http.ResponseWriter, r *http.Request) {
	if c.opts.Limits == nil {
		c.refuse(w, http.StatusNotImplemented, "한도 저장이 배선되지 않았다",
			"이 빌드의 콘솔에는 Guardian 한도 저장 seam이 주입되지 않았다.")
		return
	}

	next := config.GuardianLimits{Currency: strings.TrimSpace(r.PostFormValue("limit_currency"))}
	targets := []struct {
		key string
		set func(*config.GuardianLimits, float64)
	}{
		{"max_order_quantity", func(l *config.GuardianLimits, v float64) { l.MaxOrderQuantity = v }},
		{"max_order_notional", func(l *config.GuardianLimits, v float64) { l.MaxOrderNotional = v }},
		{"max_total_exposure", func(l *config.GuardianLimits, v float64) { l.MaxTotalExposure = v }},
		{"max_daily_loss_amount", func(l *config.GuardianLimits, v float64) { l.MaxDailyLossAmount = v }},
		{"max_daily_loss_ratio", func(l *config.GuardianLimits, v float64) { l.MaxDailyLossRatio = v }},
	}
	for _, target := range targets {
		raw := strings.TrimSpace(r.PostFormValue(target.key))
		if raw == "" {
			// Left at zero deliberately: an absent field is an unset limit, and the
			// validation below refuses the block for the same reason the interlock
			// would. Filling it from the registry here would be the implicit default
			// design D1 rejected — with the extra sin of doing it inside a form the
			// operator thought they had control of.
			continue
		}
		v, err := strconv.ParseFloat(strings.ReplaceAll(raw, ",", ""), 64)
		if err != nil {
			c.redirectSettings(w, r, "저장 안 됨 — "+target.key+"이(가) 숫자가 아니다: "+raw)
			return
		}
		target.set(&next, v)
	}

	if err := next.Validate(); err != nil {
		c.redirectSettings(w, r, "저장 안 됨 — 엔진이 이 한도로는 기동을 거부한다: "+err.Error())
		return
	}
	// An unregistered currency and an over-ceiling value are different mistakes
	// and get different sentences. CeilingViolations folds the first into the
	// second (it has to — it returns one list), and reporting "you exceeded the
	// ceiling" for a currency that has no ceiling sends the operator to lower a
	// number that was never the problem (adversarial review A10).
	if _, err := config.GuardianCeiling(next.Currency); err != nil {
		c.redirectSettings(w, r, "저장 안 됨 — "+err.Error()+" 등록된 통화는 "+
			strings.Join(config.GuardianCurrencies(), "·")+"다.")
		return
	}
	// Checked here as well as in the writer so the message names the field. The
	// refusal that counts is still the writer's — a seam that skipped this one
	// would be stopped there (design D5).
	if violations := next.CeilingViolations(); len(violations) > 0 {
		c.redirectSettings(w, r, "저장 안 됨 — 등록된 티어 상한을 넘는다: "+
			strings.Join(violations, "; ")+" 상한 위로 올리는 것은 콘솔 밖 행위다.")
		return
	}
	if err := c.opts.Limits.Save(next); err != nil {
		c.redirectSettings(w, r, "저장 안 됨 — "+err.Error())
		return
	}

	applied := "사용자 지정값"
	if id := config.MatchingGuardianTier(next); id != "" {
		applied = id + "와 일치"
	}
	c.redirectSettings(w, r, "한도 기록됨("+applied+") — "+limitSummary(next)+". "+
		currencyConsequence(next.Currency)+" "+effectNotice(c.engineRunning()))
}
