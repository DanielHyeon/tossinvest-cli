package console

// settings_limits_test.go covers the Guardian-limit editor (change
// console-sets-guardian-limits, tasks 5.x and 6.x).
//
// The claim these tests exist to defend is narrow and load-bearing: the console
// may write the five ceilings and the currency, and may not write the switch.
// The seam's Save takes config.GuardianLimits, which has no `enabled` field, so
// the shape of the message is the enforcement — TestTheLimitSeamCannotCarryTheSwitch
// reads that off the interface with reflection rather than trusting the handler.

import (
	"errors"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
)

type fakeLimits struct {
	mu      sync.Mutex
	gate    config.AutomationGate
	loadErr error
	saveErr error
	saved   []config.GuardianLimits
}

func (f *fakeLimits) Load() (config.AutomationGate, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.gate, f.loadErr
}

// Save mirrors config.Service.SaveEngineGateLimits: the same two refusals, and
// `enabled` is untouched because the message cannot carry it.
func (f *fakeLimits) Save(l config.GuardianLimits) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.saveErr != nil {
		return f.saveErr
	}
	if err := l.Validate(); err != nil {
		return err
	}
	if v := l.CeilingViolations(); len(v) > 0 {
		return errors.New(strings.Join(v, "; "))
	}
	f.saved = append(f.saved, l)
	f.gate.MaxOrderQuantity = l.MaxOrderQuantity
	f.gate.MaxOrderNotional = l.MaxOrderNotional
	f.gate.MaxTotalExposure = l.MaxTotalExposure
	f.gate.MaxDailyLossAmount = l.MaxDailyLossAmount
	f.gate.MaxDailyLossRatio = l.MaxDailyLossRatio
	f.gate.LimitCurrency = l.Currency
	return nil
}

func (f *fakeLimits) writes() []config.GuardianLimits {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]config.GuardianLimits(nil), f.saved...)
}

func limitsHarness(t *testing.T, seam *fakeLimits) *dashboardHarness {
	t.Helper()
	return newDashboardHarness(t, func(o *Options) {
		o.Settings = &fakeSettings{}
		if seam != nil {
			o.Limits = seam
		}
	})
}

func tierLimits(t *testing.T, id string) config.GuardianLimits {
	t.Helper()
	tier, ok := config.GuardianTierByID(id)
	if !ok {
		t.Fatalf("tier %q is not registered", id)
	}
	return tier.Limits
}

// TestTheLimitSeamCannotCarryTheSwitch is design D6/D7 read off the type: the
// write side takes GuardianLimits, and no field of it names the gate.
func TestTheLimitSeamCannotCarryTheSwitch(t *testing.T) {
	iface := reflect.TypeOf((*LimitSettings)(nil)).Elem()
	if iface.NumMethod() != 2 {
		t.Fatalf("LimitSettings declares %d methods, want exactly Load and Save", iface.NumMethod())
	}
	save, ok := iface.MethodByName("Save")
	if !ok {
		t.Fatal("LimitSettings has no Save")
	}
	arg := save.Type.In(0)
	if arg != reflect.TypeOf(config.GuardianLimits{}) {
		t.Fatalf("Save takes %s; it must take config.GuardianLimits, which has no enabled field", arg)
	}
	for i := 0; i < arg.NumField(); i++ {
		switch strings.ToLower(arg.Field(i).Name) {
		case "enabled", "attestationfile":
			t.Errorf("GuardianLimits carries %s; the console must not be able to send it",
				arg.Field(i).Name)
		}
	}
}

// TestApplyingAPresetWritesAllFiveAtOnce is the one-click path (사용자 결정
// 2026-07-30).
func TestApplyingAPresetWritesAllFiveAtOnce(t *testing.T) {
	seam := &fakeLimits{gate: config.AutomationGate{Enabled: false}}
	h := limitsHarness(t, seam)
	h.authenticate(t)

	h.post(t, "/settings/limits/preset", url.Values{"csrf": {h.csrf}, "tier": {"kr-small-live"}})

	writes := seam.writes()
	if len(writes) != 1 {
		t.Fatalf("writes = %d, want 1", len(writes))
	}
	if writes[0] != tierLimits(t, "kr-small-live") {
		t.Errorf("wrote %+v, want the tier's limits", writes[0])
	}
}

// TestThePresetControlsAskForNoTyping — the preset forms carry a hidden tier id
// and nothing to fill in (타이핑 확인·추가 승인 마찰 금지, 사용자 결정 2026-07-27).
func TestThePresetControlsAskForNoTyping(t *testing.T) {
	h := limitsHarness(t, &fakeLimits{})
	h.authenticate(t)
	page := h.page(t, "/settings")

	start := strings.Index(page, `action="/settings/limits/preset"`)
	if start < 0 {
		t.Fatal("the settings screen offers no preset control")
	}
	end := strings.Index(page[start:], "</form>")
	if end < 0 {
		t.Fatal("the preset form is not closed")
	}
	form := page[start : start+end]
	for _, forbidden := range []string{`type="text"`, `type="number"`, "<textarea"} {
		if strings.Contains(form, forbidden) {
			t.Errorf("the preset form contains %s; applying a preset must be one click", forbidden)
		}
	}
	if !strings.Contains(form, "onsubmit=\"return confirm(") {
		t.Error("the preset form has no confirmation dialog")
	}
}

// TestEveryRegisteredTierIsOfferedWithItsNumbers — the operator sees what a
// click will install before clicking it.
func TestEveryRegisteredTierIsOfferedWithItsNumbers(t *testing.T) {
	h := limitsHarness(t, &fakeLimits{})
	h.authenticate(t)
	page := h.page(t, "/settings")

	for _, tier := range config.GuardianTiers() {
		if !strings.Contains(page, `value="`+tier.ID+`"`) {
			t.Errorf("tier %s is not offered", tier.ID)
		}
		if !strings.Contains(page, tier.Label) {
			t.Errorf("tier %s renders without its label", tier.ID)
		}
		if !strings.Contains(page, tier.Limits.Currency) {
			t.Errorf("tier %s renders without its currency", tier.ID)
		}
	}
	// The default tier is marked as the recommendation rather than applied.
	if !strings.Contains(page, "권장") {
		t.Error("no tier is marked as the recommended default")
	}
}

// TestUnsetLimitsRenderAsUnsetAndNotAsTheDefault is design D1: the screen must
// not draw a number the file does not contain, because the engine's interlock
// reads the file.
func TestUnsetLimitsRenderAsUnsetAndNotAsTheDefault(t *testing.T) {
	h := limitsHarness(t, &fakeLimits{})
	h.authenticate(t)
	page := h.page(t, "/settings")

	// Scoped to the current-values table on purpose: the preset cards DO show the
	// default tier's numbers, because the operator has to see what a click will
	// install. What must not happen is those numbers appearing as the gate's
	// current state.
	current := limitCurrent(t, page)
	if !strings.Contains(current, "미설정") {
		t.Error("an empty gate block does not render as 미설정")
	}
	for _, spelling := range []string{"1000000", "1,000,000", "10000000", "10,000,000"} {
		if strings.Contains(current, spelling) {
			t.Errorf("the current-value table drew %s as if it were configured; the engine "+
				"would refuse to start on this file", spelling)
		}
	}
}

// TestAPartialBlockSaysTheInterlockRefusesIt (design D2).
func TestAPartialBlockSaysTheInterlockRefusesIt(t *testing.T) {
	seam := &fakeLimits{gate: config.AutomationGate{
		MaxOrderQuantity: 10, LimitCurrency: "KRW",
	}}
	h := limitsHarness(t, seam)
	h.authenticate(t)

	section := limitSection(t, h.page(t, "/settings"))
	for _, want := range []string{"기동", "거부", "프리셋"} {
		if !strings.Contains(section, want) {
			t.Errorf("a partially configured gate does not say %q", want)
		}
	}
}

// TestTheScreenNamesTheMatchingTier (design D12).
func TestTheScreenNamesTheMatchingTier(t *testing.T) {
	match := tierLimits(t, "us-smoke")
	seam := &fakeLimits{gate: config.AutomationGate{
		MaxOrderQuantity: match.MaxOrderQuantity, MaxOrderNotional: match.MaxOrderNotional,
		MaxTotalExposure: match.MaxTotalExposure, MaxDailyLossAmount: match.MaxDailyLossAmount,
		MaxDailyLossRatio: match.MaxDailyLossRatio, LimitCurrency: match.Currency,
	}}
	h := limitsHarness(t, seam)
	h.authenticate(t)
	if section := limitSection(t, h.page(t, "/settings")); !strings.Contains(section, "us-smoke") {
		t.Error("the screen does not name the tier the file currently spells")
	}

	seam.gate.MaxDailyLossAmount = 7
	if section := limitSection(t, h.page(t, "/settings")); !strings.Contains(section, "사용자 지정값") {
		t.Error("a block matching no tier is not reported as a custom one")
	}
}

// TestTheApplyAnswerNamesTheMarketTheCurrencyCloses (design D8).
func TestTheApplyAnswerNamesTheMarketTheCurrencyCloses(t *testing.T) {
	for _, tc := range []struct{ tier, closed string }{
		{"us-small-live", "국내"},
		{"kr-small-live", "미국"},
	} {
		seam := &fakeLimits{}
		h := limitsHarness(t, seam)
		h.authenticate(t)
		page := body(t, h.post(t, "/settings/limits/preset",
			url.Values{"csrf": {h.csrf}, "tier": {tc.tier}}))
		if !strings.Contains(page, tc.closed) {
			t.Errorf("%s: the answer does not say %s 자동 진입이 닫힌다", tc.tier, tc.closed)
		}
		if !strings.Contains(page, "통화") {
			t.Errorf("%s: the answer does not mention the limit currency", tc.tier)
		}
	}
}

// TestTheApplyAnswerDefersToTheEngineRestart — no present-tense guarantee
// (design D9; the A1 lesson from console-excludes-in-one-click).
func TestTheApplyAnswerDefersToTheEngineRestart(t *testing.T) {
	h := limitsHarness(t, &fakeLimits{})
	h.authenticate(t)
	page := body(t, h.post(t, "/settings/limits/preset",
		url.Values{"csrf": {h.csrf}, "tier": {"kr-smoke"}}))
	if !strings.Contains(page, "다음 엔진 기동부터 반영") {
		t.Error("the answer does not defer to the next engine start")
	}
	for _, lie := range []string{"지금부터 적용된다", "즉시 적용된다"} {
		if strings.Contains(page, lie) {
			t.Errorf("the answer claims %q", lie)
		}
	}
}

// TestApplyingAPresetDoesNotTouchTheAdoptionBlock — two editors, one file.
func TestApplyingAPresetDoesNotTouchTheAdoptionBlock(t *testing.T) {
	adoption := &fakeSettings{block: config.Adoption{
		Enabled: true, DefaultStopPct: 0.07, ExcludeSymbols: []string{"TSLA"},
	}}
	seam := &fakeLimits{}
	h := newDashboardHarness(t, func(o *Options) {
		o.Settings = adoption
		o.Limits = seam
	})
	h.authenticate(t)

	h.post(t, "/settings/limits/preset", url.Values{"csrf": {h.csrf}, "tier": {"kr-smoke"}})

	if _, saves := adoption.saved(); saves != 0 {
		t.Errorf("the adoption seam was written %d times by a limit save", saves)
	}
}

// TestTheAdvancedFormSavesIndividualValues — the fold, mirroring the adoption
// screen's list-editing fold.
func TestTheAdvancedFormSavesIndividualValues(t *testing.T) {
	seam := &fakeLimits{}
	h := limitsHarness(t, seam)
	h.authenticate(t)

	h.post(t, "/settings/limits", url.Values{
		"csrf":                  {h.csrf},
		"max_order_quantity":    {"5"},
		"max_order_notional":    {"200000"},
		"max_total_exposure":    {"400000"},
		"max_daily_loss_amount": {"20000"},
		"max_daily_loss_ratio":  {"0.005"},
		"limit_currency":        {"KRW"},
	})

	writes := seam.writes()
	if len(writes) != 1 {
		t.Fatalf("writes = %d, want 1", len(writes))
	}
	want := config.GuardianLimits{
		MaxOrderQuantity: 5, MaxOrderNotional: 200_000, MaxTotalExposure: 400_000,
		MaxDailyLossAmount: 20_000, MaxDailyLossRatio: 0.005, Currency: "KRW",
	}
	if writes[0] != want {
		t.Errorf("wrote %+v, want %+v", writes[0], want)
	}
}

// TestTheAdvancedFormIsInsideAFold (사용자 결정: 1차 경로는 프리셋).
func TestTheAdvancedFormIsInsideAFold(t *testing.T) {
	h := limitsHarness(t, &fakeLimits{})
	h.authenticate(t)
	section := limitSection(t, h.page(t, "/settings"))

	form := strings.Index(section, `action="/settings/limits"`)
	if form < 0 {
		t.Fatal("the advanced limit form is not rendered")
	}
	fold := strings.LastIndex(section[:form], "<details")
	if fold < 0 {
		t.Fatal("the advanced limit form is not inside a fold")
	}
	if closed := strings.Index(section[fold:form], "</details>"); closed >= 0 {
		t.Error("the advanced limit form sits outside the fold that precedes it")
	}
}

// TestTheAdvancedFormShowsNoZeroForAnUnsetLimit — adversarial review A1.
//
// The current-value table says 미설정 and the form must not say 0 about the same
// field. Zero is the value this codebase spends paragraphs refusing to conflate
// with "nobody chose one", and a form pre-filled with it invites the operator to
// believe the ceiling is zero — or to leave it and submit a block the interlock
// refuses for a reason the screen already denied.
func TestTheAdvancedFormShowsNoZeroForAnUnsetLimit(t *testing.T) {
	h := limitsHarness(t, &fakeLimits{})
	h.authenticate(t)
	section := limitSection(t, h.page(t, "/settings"))

	for _, field := range []string{
		"max_order_quantity", "max_order_notional", "max_total_exposure",
		"max_daily_loss_amount", "max_daily_loss_ratio", "limit_currency",
	} {
		want := `name="` + field + `" value=""`
		if !strings.Contains(section, want) {
			t.Errorf("the advanced form does not render %s as empty; it must not pre-fill 0", field)
		}
	}
}

// TestTheAdvancedFormRendersNumbersTheOperatorCanRead — adversarial review A2.
//
// Go's default float rendering turns 10000000 into 1e+07. It round-trips through
// ParseFloat, so nothing breaks — but an operator who reads 1e+07 in a text box
// and "corrects" it is one keystroke from a limit off by a factor of ten.
func TestTheAdvancedFormRendersNumbersTheOperatorCanRead(t *testing.T) {
	seam := &fakeLimits{gate: config.AutomationGate{
		MaxOrderQuantity: 100, MaxOrderNotional: 1_000_000,
		MaxTotalExposure: 10_000_000, MaxDailyLossAmount: 100_000,
		MaxDailyLossRatio: 0.01, LimitCurrency: "KRW",
	}}
	h := limitsHarness(t, seam)
	h.authenticate(t)
	section := limitSection(t, h.page(t, "/settings"))

	if strings.Contains(section, "e+0") {
		t.Error("the screen renders a limit in exponent notation")
	}
	if !strings.Contains(section, `name="max_total_exposure" value="10000000"`) {
		t.Error("the advanced form does not round-trip the exposure ceiling as a plain decimal")
	}
}

// TestAnUnregisteredCurrencyIsRefusedForTheRightReason — adversarial review A10.
//
// "JPY has no registered tier" is not "you exceeded the ceiling". Reporting the
// first as the second sends the operator to lower a number that was never the
// problem.
func TestAnUnregisteredCurrencyIsRefusedForTheRightReason(t *testing.T) {
	seam := &fakeLimits{}
	h := limitsHarness(t, seam)
	h.authenticate(t)

	page := body(t, h.post(t, "/settings/limits", url.Values{
		"csrf":                  {h.csrf},
		"max_order_quantity":    {"100"},
		"max_order_notional":    {"1000"},
		"max_total_exposure":    {"10000"},
		"max_daily_loss_amount": {"100"},
		"max_daily_loss_ratio":  {"0.01"},
		"limit_currency":        {"JPY"},
	}))

	if len(seam.writes()) != 0 {
		t.Fatal("an unregistered currency was written")
	}
	if strings.Contains(page, "상한을 넘는다") {
		t.Error("an unregistered currency was reported as a ceiling violation")
	}
	for _, want := range []string{"JPY", "KRW", "USD"} {
		if !strings.Contains(page, want) {
			t.Errorf("the refusal does not mention %s; it must name the registered currencies", want)
		}
	}
}

// TestAValueAboveTheCeilingIsRefusedWithItsReason (design D5).
func TestAValueAboveTheCeilingIsRefusedWithItsReason(t *testing.T) {
	seam := &fakeLimits{}
	h := limitsHarness(t, seam)
	h.authenticate(t)

	page := body(t, h.post(t, "/settings/limits", url.Values{
		"csrf":                  {h.csrf},
		"max_order_quantity":    {"100"},
		"max_order_notional":    {"9000000"},
		"max_total_exposure":    {"10000000"},
		"max_daily_loss_amount": {"100000"},
		"max_daily_loss_ratio":  {"0.01"},
		"limit_currency":        {"KRW"},
	}))

	if len(seam.writes()) != 0 {
		t.Fatal("a value above the ceiling was written")
	}
	if !strings.Contains(page, "max_order_notional") {
		t.Error("the refusal does not name the field that broke the ceiling")
	}
}

// TestABlockTheInterlockWouldRefuseIsNotWritten — the console does not record a
// gate the engine would refuse to start on.
func TestABlockTheInterlockWouldRefuseIsNotWritten(t *testing.T) {
	for _, tc := range []struct {
		name string
		form url.Values
	}{
		{"a missing field", url.Values{
			"max_order_quantity": {"100"}, "limit_currency": {"KRW"},
		}},
		{"a non-numeric field", url.Values{
			"max_order_quantity": {"열"}, "max_order_notional": {"1000000"},
			"max_total_exposure": {"10000000"}, "max_daily_loss_amount": {"100000"},
			"max_daily_loss_ratio": {"0.01"}, "limit_currency": {"KRW"},
		}},
		{"a ratio above one", url.Values{
			"max_order_quantity": {"100"}, "max_order_notional": {"1000000"},
			"max_total_exposure": {"10000000"}, "max_daily_loss_amount": {"100000"},
			"max_daily_loss_ratio": {"1.5"}, "limit_currency": {"KRW"},
		}},
	} {
		seam := &fakeLimits{}
		h := limitsHarness(t, seam)
		h.authenticate(t)
		form := tc.form
		form.Set("csrf", h.csrf)

		page := body(t, h.post(t, "/settings/limits", form))
		if len(seam.writes()) != 0 {
			t.Errorf("%s: was written", tc.name)
		}
		if !strings.Contains(page, "저장 안 됨") {
			t.Errorf("%s: the answer does not report a refusal", tc.name)
		}
	}
}

// TestAnUnknownTierIsRefused — a form value nobody offered.
func TestAnUnknownTierIsRefused(t *testing.T) {
	seam := &fakeLimits{}
	h := limitsHarness(t, seam)
	h.authenticate(t)

	resp := h.post(t, "/settings/limits/preset", url.Values{"csrf": {h.csrf}, "tier": {"kr-paper-demo"}})
	if resp.StatusCode == http.StatusOK && len(seam.writes()) != 0 {
		t.Fatal("an unregistered tier was applied")
	}
	if len(seam.writes()) != 0 {
		t.Fatal("an unregistered tier was applied")
	}
}

// TestWithoutASeamTheLimitEditorRefusesRatherThanPretends.
func TestWithoutASeamTheLimitEditorRefusesRatherThanPretends(t *testing.T) {
	h := limitsHarness(t, nil)
	h.authenticate(t)

	page := h.page(t, "/settings")
	if strings.Contains(page, `action="/settings/limits/preset"`) {
		t.Error("the preset controls render without a seam to save through")
	}
	for _, path := range []string{"/settings/limits", "/settings/limits/preset"} {
		resp := h.post(t, path, url.Values{"csrf": {h.csrf}, "tier": {"kr-smoke"}})
		if resp.StatusCode != http.StatusNotImplemented {
			t.Errorf("POST %s without a seam = %d, want 501", path, resp.StatusCode)
		}
	}
}

// TestAnUnreadableConfigDoesNotHideTheRestOfTheScreen.
func TestAnUnreadableConfigDoesNotHideTheRestOfTheScreen(t *testing.T) {
	h := limitsHarness(t, &fakeLimits{loadErr: errors.New("permission denied")})
	h.authenticate(t)

	page := h.page(t, "/settings")
	if !strings.Contains(page, "permission denied") {
		t.Error("the read failure is not reported")
	}
	if !strings.Contains(page, "편입 설정") {
		t.Error("a limit read failure took the adoption section down with it")
	}
}

// TestTheGateStateIsShownButNotOffered: the screen says whether the gate is on,
// and offers nothing that changes it.
func TestTheGateStateIsShownButNotOffered(t *testing.T) {
	h := limitsHarness(t, &fakeLimits{gate: config.AutomationGate{Enabled: true}})
	h.authenticate(t)
	page := h.page(t, "/settings")

	if !strings.Contains(page, "automation gate") || !strings.Contains(page, "콘솔에서 편집할 수 없다") {
		t.Error("the screen no longer says the gate itself is not editable here")
	}
	for _, forbidden := range []string{`name="enabled" `, `name="gate_enabled"`} {
		if strings.Contains(limitSection(t, page), forbidden) {
			t.Errorf("the limit section carries %s", forbidden)
		}
	}
}

// limitSection returns the Guardian-limit part of the settings page, so an
// assertion about it cannot be satisfied by a word that appears in the adoption
// part.
func limitSection(t *testing.T, page string) string {
	t.Helper()
	start := strings.Index(page, `id="guardian-limits"`)
	if start < 0 {
		t.Fatal("the settings screen has no Guardian-limit section")
	}
	end := strings.Index(page[start:], "</section>")
	if end < 0 {
		t.Fatal("the Guardian-limit section is not closed")
	}
	return page[start : start+end]
}

// limitCurrent returns just the table of what the file currently spells.
func limitCurrent(t *testing.T, page string) string {
	t.Helper()
	start := strings.Index(page, `id="guardian-current"`)
	if start < 0 {
		t.Fatal("the settings screen has no current-limit table")
	}
	end := strings.Index(page[start:], "</table>")
	if end < 0 {
		t.Fatal("the current-limit table is not closed")
	}
	return page[start : start+end]
}
