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
		{"하루 매수 총액 상한", dailyBuyText(p.Gate), dailyBuyUnit(p.Gate)},
		{"최대 보유 포지션 수", openPositionsText(p.Gate.MaxOpenPositions), openPositionsUnit(p.Gate.MaxOpenPositions)},
	}
}

// openPositionsText 는 최대 보유 포지션 수의 표시(a138). 미설정이면 엔진이 실제로 쓸 기본값을 평이하게 말함 — 통화와 무관하게 6.
// 쓸 수 없는 값(음수 · 소수)이면 엔진이 기동을 거부한다고 말함(판정은 config.WholeOpenPositions 한 곳).
func openPositionsText(v float64) string {
	if v == 0 {
		return "미설정 — 기본값 " + strconv.Itoa(config.DefaultMaxOpenPositions) + " 종목이 적용된다"
	}
	if n, ok := config.WholeOpenPositions(v); ok {
		return strconv.Itoa(n)
	}
	return decimalText(v) + " — 1 이상의 정수가 아니어서 엔진이 기동을 거부한다"
}

// openPositionsUnit 은 값이 쓸 수 있을 때만 단위를 붙임(미설정 · 무효 문장은 스스로 말함).
func openPositionsUnit(v float64) string {
	if _, ok := config.WholeOpenPositions(v); ok {
		return "종목"
	}
	return ""
}

// FieldOpenPositions 는 고급 폼의 최대 보유 포지션 수 칸 값 — 미설정은 빈 칸(다른 칸과 같은 규칙).
func (p settingsPage) FieldOpenPositions() string { return limitFieldText(p.Gate.MaxOpenPositions) }

// dailyBuyText 는 하루 매수 총액 상한의 표시(a136). 미설정이면 엔진이 실제로 할 일을 평이하게 말함 — KRW 는 기본값 적용,
// 그 밖의 통화는 기동 거부(사용자 결정: USD 프리셋 값을 만들지 않고, 값이 없으면 엔진이 기동을 거부한다고 말함).
func dailyBuyText(g config.AutomationGate) string {
	if g.MaxDailyBuyNotional > 0 {
		return decimalText(g.MaxDailyBuyNotional)
	}
	switch strings.ToUpper(strings.TrimSpace(g.LimitCurrency)) {
	case "KRW":
		return "미설정 — 기본값 " + decimalText(config.DefaultMaxDailyBuyNotionalKRW) + " KRW 가 적용된다"
	case "":
		return "미설정"
	default:
		return "미설정 — " + strings.ToUpper(strings.TrimSpace(g.LimitCurrency)) +
			" 계좌에는 기본값이 없어 엔진이 기동을 거부한다. 값을 넣어야 한다"
	}
}

// dailyBuyUnit 은 값이 있을 때만 통화를 붙임(미설정 문장은 통화를 스스로 말함).
func dailyBuyUnit(g config.AutomationGate) string {
	if g.MaxDailyBuyNotional > 0 {
		return strings.TrimSpace(g.LimitCurrency)
	}
	return ""
}

// FieldDailyBuy 는 고급 폼의 하루 매수 상한 칸 값 — 미설정은 빈 칸(다른 칸과 같은 규칙).
func (p settingsPage) FieldDailyBuy() string { return limitFieldText(p.Gate.MaxDailyBuyNotional) }

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
	// Preview is what clicking this card would change, measured against the file
	// as it stands (change a055 §4). This is the one settings card where the
	// server knows BOTH sides before the click — the tier's values are fixed and
	// the current values are in the file — so it is the one card that can show
	// 현재 → 변경 rather than 현재 alone.
	Preview tierPreview
}

// LimitTierCards is what the template ranges over.
func (p settingsPage) LimitTierCards() []tierCard {
	recommended := config.DefaultGuardianTierID()
	out := make([]tierCard, 0, len(config.GuardianTiers()))
	for _, tier := range config.GuardianTiers() {
		card := tierCard{GuardianTier: tier, Recommended: tier.ID == recommended}
		card.Preview = card.previewAgainst(p.Gate.Limits(), p.Gate.LimitsSet(), p.EngineRunning)
		out = append(out, card)
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

// currencyConsequence explains the account-base unit and the paired FX
// authorities it requires.
//
// It is said on EVERY apply, not only when the currency changes. The gate still
// has one unit, but a072 defines it as account base: KR and US share one
// account-wide cap while each market supplies its own frozen quote-to-base FX.
// Neither market's operational stability is a prerequisite for implementing
// or validating the peer path.
func currencyConsequence(currency string) string {
	switch strings.ToUpper(strings.TrimSpace(currency)) {
	case "KRW":
		return "계좌 기준 통화는 KRW다 — KR은 identity FX, US는 공식 USD→KRW FX가 각각 current일 때 " +
			"같은 account-wide Guardian 한도를 사용한다. KR과 US는 같은 구현 웨이브이며 어느 시장도 peer 운영 안정화를 기다리지 않는다."
	case "USD":
		return "계좌 기준 통화는 USD다 — US는 identity FX, KR은 공식 KRW→USD FX가 각각 current일 때 " +
			"같은 account-wide Guardian 한도를 사용한다. KR과 US는 같은 구현 웨이브이며 어느 시장도 peer 운영 안정화를 기다리지 않는다."
	}
	return "계좌 기준 통화 " + currency + "는 KR/US official FX authority가 지원되지 않아 신규 자동 진입 권위를 만들지 않는다."
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
	// a136: 프리셋은 하루 매수 상한을 쓰지 않음. 통화가 바뀌었으면 writer 가 그 값을 지웠으므로 그 결과를 말함.
	daily := ""
	if gate, err := c.opts.Limits.Load(); err == nil {
		daily = " " + dailyBuyNotice(gate.LimitCurrency, gate.MaxDailyBuyNotional)
	}
	c.redirectSettings(w, r, tier.Label+" 한도 기록됨 — "+limitSummary(tier.Limits)+"."+daily+" "+
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
	// a136: 하루 매수 총액 상한 — 이 폼은 늘 보내므로 빈 칸은 「미설정(0)」으로 씀(KRW 는 기본값, 그 밖은 기동 거부).
	dailyBuy := 0.0
	if raw := strings.TrimSpace(r.PostFormValue("max_daily_buy_notional")); raw != "" {
		v, err := strconv.ParseFloat(strings.ReplaceAll(raw, ",", ""), 64)
		if err != nil {
			c.redirectSettings(w, r, "저장 안 됨 — max_daily_buy_notional이(가) 숫자가 아니다: "+raw)
			return
		}
		dailyBuy = v
	}
	next.DailyBuy = &dailyBuy
	// 구현 리뷰 P1-1: 칸은 현재 값으로 미리 채워지므로, 통화를 바꾸면서 칸을 그대로 두면 옛 통화의 숫자가 새 통화 상한으로 넘어감
	// (예: 5,000,000 KRW → 5,000,000 USD). 저장 전 파일과 비교해 「통화가 바뀌었는데 숫자는 파일 그대로」면 값을 보내지 않아
	// writer 가 지우게 함(통화 전환 = 지움, spec). 운영자가 새 통화로 다른 숫자를 쳤으면 그 값이 이김.
	if current, err := c.opts.Limits.Load(); err == nil &&
		!strings.EqualFold(strings.TrimSpace(current.LimitCurrency), strings.TrimSpace(next.Currency)) &&
		current.MaxDailyBuyNotional > 0 && current.MaxDailyBuyNotional == dailyBuy {
		next.DailyBuy = nil
		dailyBuy = 0
	}

	// a138: 최대 보유 포지션 수 — 칸이 폼에 **있을 때만** 씀(없는 옛 폼이 운영자 값을 기본값으로 되돌리지 않게). 빈 칸은 미설정(0 → 기본 6).
	// 통화를 바꾸는 저장도 이 값을 지우지 않음(개수는 통화로 재해석되지 않음).
	openPositions := -1.0
	if _, present := r.PostForm["max_open_positions"]; present {
		v := 0.0
		if raw := strings.TrimSpace(r.PostFormValue("max_open_positions")); raw != "" {
			parsed, err := strconv.ParseFloat(strings.ReplaceAll(raw, ",", ""), 64)
			if err != nil {
				c.redirectSettings(w, r, "저장 안 됨 — max_open_positions이(가) 숫자가 아니다: "+raw)
				return
			}
			v = parsed
		}
		next.OpenPositions = &v
		openPositions = v
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
	positions := ""
	if openPositions >= 0 {
		positions = " 최대 보유 포지션 수: " + openPositionsText(openPositions) + "."
	}
	c.redirectSettings(w, r, "한도 기록됨("+applied+") — "+limitSummary(next)+". "+
		dailyBuyNotice(next.Currency, dailyBuy)+positions+" "+currencyConsequence(next.Currency)+" "+effectNotice(c.engineRunning()))
}

// dailyBuyNotice 는 저장 뒤 하루 매수 상한이 실제로 무엇이 되는지 한 문장으로 말함(a136).
func dailyBuyNotice(currency string, value float64) string {
	if value > 0 {
		return "하루 매수 총액 상한 " + decimalText(value) + " " + strings.ToUpper(strings.TrimSpace(currency)) + "."
	}
	return "하루 매수 총액 상한: " + dailyBuyText(config.AutomationGate{LimitCurrency: currency}) + "."
}
