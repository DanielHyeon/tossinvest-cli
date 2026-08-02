package console

// settings_card_test.go pins the card standard (change
// a055-console-settings-cadence §4).
//
// The requirement that is easy to write and hard to check is ③: "저장할 수 없으면
// 이름 붙은 사유를 저장 표면이 있어야 할 자리에 표시한다". The obvious check looks
// for a disabled attribute — and this console does not disable a form whose seam
// is missing, it declines to render it. That check finds zero disabled controls,
// passes, and measures nothing. So the claim is stated from the other side: every
// card either renders a submit control or renders a named reason, and a card with
// neither is the failure.

import (
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
)

var settingsCardBlock = regexp.MustCompile(`data-settings-card="([^"]*)"`)

// cardsOn cuts each settings card out of a rendered tab.
func cardsOn(t *testing.T, page string) map[string]string {
	t.Helper()
	out := map[string]string{}
	for _, m := range settingsCardBlock.FindAllStringSubmatchIndex(page, -1) {
		name := page[m[2]:m[3]]
		rest := page[m[0]:]
		end := strings.Index(rest, "</section>")
		if end < 0 {
			t.Fatalf("the card %q is not closed", name)
		}
		out[name] = rest[:end]
	}
	return out
}

// TestEveryCardEitherSavesOrSaysWhyNot (task 4.3).
//
// Run over BOTH a console with every seam wired and one with none, because the
// two exercise opposite halves: with the seams the cards render forms, without
// them every single card is in the state the requirement is actually about.
func TestEveryCardEitherSavesOrSaysWhyNot(t *testing.T) {
	for _, build := range []struct {
		name string
		h    *harness
	}{
		{"every seam wired", fullSettingsHarness(t)},
		{"no seam wired", newHarness(t)},
	} {
		build.h.authenticate(t)
		found := 0
		for _, path := range settingsTabPaths {
			for name, card := range cardsOn(t, body(t, build.h.get(t, path))) {
				found++
				// A card may hold more than one form — the system-update card holds
				// the download and the install — so the claim is made per form and
				// then once for the card as a whole.
				forms := formBlock.FindAllStringSubmatch(card, -1)
				for _, form := range forms {
					if !strings.Contains(form[2], `<button type="submit">`) {
						t.Errorf("%s/%s (%s): the form posting to %s has no submit control and no "+
							"reason took its place", path, name, build.name, form[1])
					}
				}
				if len(forms) == 0 && !strings.Contains(card, "data-save-block=") {
					t.Errorf("%s/%s (%s): the save surface is missing and nothing says why. "+
						"An empty space gives the operator no next action:\n%s",
						path, name, build.name, card)
				}
			}
		}
		if found == 0 {
			t.Fatalf("%s: no settings card was found; the scan is not reading the tabs", build.name)
		}
	}
}

// TestEveryCardShowsWhatChangesAndWhen (tasks 4.1, 4.2).
func TestEveryCardShowsWhatChangesAndWhen(t *testing.T) {
	h := fullSettingsHarness(t)
	h.authenticate(t)
	for _, path := range settingsTabPaths {
		for name, card := range cardsOn(t, body(t, h.get(t, path))) {
			if !strings.Contains(card, `class="card-preview"`) {
				t.Errorf("%s/%s renders no 적용 후 preview; the operator presses without being told "+
					"what the press writes", path, name)
			}
		}
	}

	// 반영 시점 is one of the eight things that may never be folded, so it is a
	// notice and not muted prose, and it is on every card that writes a file.
	standing := body(t, h.get(t, pathSettingsStanding))
	for _, card := range []string{"adoption", "gate"} {
		marker := `data-effect-timing="` + card + `"`
		i := strings.Index(standing, marker)
		if i < 0 {
			t.Errorf("the %s card does not say when a save takes effect", card)
			continue
		}
		if !strings.Contains(standing[max(0, i-60):i], `class="notice"`) {
			t.Errorf("the %s card states 반영 시점 as something other than a notice; it is on the "+
				"never-folded list", card)
		}
	}
}

// TestEachBlockingReasonAppearsUnderItsOwnState (task 4.3, second half).
//
// The seven reasons of design §3, each with the state that produces it. None of
// them is a new judgement: every one already existed on the old screen as a
// paragraph, and this change gave it a name.
func TestEachBlockingReasonAppearsUnderItsOwnState(t *testing.T) {
	offGrid := config.Adoption{Enabled: true, DefaultStopPct: 0.076}
	for _, tc := range []struct {
		name   string
		tab    string
		reason string
		tweak  func(*Options)
	}{{
		name: "저장 seam 미배선", tab: pathSettingsStanding, reason: `data-save-block="저장 미배선"`,
		tweak: func(o *Options) { o.Settings = nil },
	}, {
		name: "설정 파일 읽기 실패", tab: pathSettingsStanding,
		reason: `data-save-block="설정 파일 읽기 실패"`,
		tweak:  func(o *Options) { o.Settings = &fakeSettings{loadErr: errors.New("permission denied")} },
	}, {
		name: "엔진이 거부할 블록", tab: pathSettingsStanding,
		reason: `data-save-caution="엔진이 거부할 블록"`,
		tweak:  func(o *Options) { o.Settings = &fakeSettings{verdict: "stop fraction out of band"} },
	}, {
		name: "한도 미설정", tab: pathSettingsDaily, reason: `data-save-caution="한도 미설정"`,
		tweak: func(o *Options) { o.Limits = &fakeLimits{} },
	}, {
		name: "손절폭 단위 grid 불일치", tab: pathSettingsStanding,
		reason: `data-save-caution="손절폭 단위 불일치"`,
		tweak:  func(o *Options) { o.Settings = &fakeSettings{block: offGrid} },
	}, {
		name: "지금 켜면 기동 거부", tab: pathSettingsStanding,
		reason: `data-save-caution="지금 켜면 기동이 거부된다"`,
		tweak:  func(o *Options) { o.TradingPolicy = &fakeTrading{block: config.Trading{}} },
	}} {
		t.Run(tc.name, func(t *testing.T) {
			h := fullSettingsHarness(t, tc.tweak)
			h.authenticate(t)
			page := body(t, h.get(t, tc.tab))
			if !strings.Contains(page, tc.reason) {
				t.Errorf("the %s state renders no %s:\n%s", tc.name, tc.reason, page)
			}
		})
	}
}

// TestARunningEngineIsSaidOnEveryCardThatWritesAFile (reason ⑦).
func TestARunningEngineIsSaidOnEveryCardThatWritesAFile(t *testing.T) {
	h := newEngineHarness(t, func(o *Options) {
		o.Settings = &fakeSettings{}
		o.Limits = &fakeLimits{gate: fullLimits()}
		o.TradingPolicy = &fakeTrading{block: fullPolicy()}
		o.Gate = &fakeGate{}
	})
	h.authenticate(t)
	holdEngineMarker(t, h.marker, engineNow)

	for _, tab := range []string{pathSettingsStanding, pathSettingsDaily} {
		page := body(t, h.get(t, tab))
		if !strings.Contains(page, "재시작") {
			t.Errorf("%s does not say a running engine keeps its startup snapshot:\n%s", tab, page)
		}
	}
}

// --- the Guardian limits' direction (tasks 4.5, 4.5b) ------------------------------

// TestTighteningAndLooseningAreMarkedDifferently (task 4.5).
//
// Within one currency. And the screen says the direction — it does not refuse
// the widening one: allowed-or-not is the server's existing validation, and a
// screen with a second opinion about it is a second rule to keep in step.
func TestTighteningAndLooseningAreMarkedDifferently(t *testing.T) {
	tier := tierLimits(t, "kr-small-live")
	current := tier
	// One ceiling wider than the tier and one narrower, in the tier's own currency.
	current.MaxOrderNotional = tier.MaxOrderNotional * 2
	current.MaxTotalExposure = tier.MaxTotalExposure / 2

	card := tierCard{GuardianTier: config.GuardianTier{ID: "kr-small-live", Limits: tier}}
	preview := card.previewAgainst(current, true, false)
	if preview.CurrencyChange {
		t.Fatal("a same-currency comparison reported a currency change")
	}
	axes := map[string]string{}
	for _, row := range preview.Rows {
		axes[row.Label] = row.Axis
	}
	if got := axes["1회 주문 금액 상한"]; got != "강화" {
		t.Errorf("halving a ceiling is marked %q, want 강화", got)
	}
	if got := axes["계좌 개방 노출 상한"]; got != "완화" {
		t.Errorf("doubling a ceiling is marked %q, want 완화", got)
	}

	// And nothing about the widening one is refused: the form still posts.
	h := limitsHarness(t, &fakeLimits{gate: config.AutomationGate{
		LimitCurrency: current.Currency, MaxOrderQuantity: current.MaxOrderQuantity,
		MaxOrderNotional: current.MaxOrderNotional, MaxTotalExposure: current.MaxTotalExposure,
		MaxDailyLossAmount: current.MaxDailyLossAmount, MaxDailyLossRatio: current.MaxDailyLossRatio,
	}})
	h.authenticate(t)
	page := h.page(t, pathSettingsDaily)
	if !strings.Contains(page, `action="/settings/limits/preset"`) {
		t.Error("the screen withheld the preset form because a ceiling would widen; the direction " +
			"is a display and the refusal belongs to the server")
	}
}

// TestACurrencyChangeIsNeitherTighteningNorLoosening (task 4.5b).
//
// KRW 500,000 → USD 3,000 is a smaller number and the opposite of a tightening:
// the gate has one currency, risk's chain refuses an intent priced in another,
// so what actually happened is that one market's automatic entry closed and the
// other's opened. Reading the numbers as a direction would be false in exactly
// the direction that matters.
func TestACurrencyChangeIsNeitherTighteningNorLoosening(t *testing.T) {
	krw := config.GuardianLimits{
		Currency: "KRW", MaxOrderQuantity: 10, MaxOrderNotional: 500_000,
		MaxTotalExposure: 2_000_000, MaxDailyLossAmount: 200_000, MaxDailyLossRatio: 0.02,
	}
	usd := config.GuardianLimits{
		Currency: "USD", MaxOrderQuantity: 10, MaxOrderNotional: 3_000,
		MaxTotalExposure: 12_000, MaxDailyLossAmount: 1_200, MaxDailyLossRatio: 0.02,
	}
	card := tierCard{GuardianTier: config.GuardianTier{ID: "x", Limits: usd}}
	preview := card.previewAgainst(krw, true, false)

	if !preview.CurrencyChange {
		t.Fatal("KRW → USD is not reported as a currency change")
	}
	for _, row := range preview.Rows {
		if row.Axis != "" {
			t.Errorf("%s is marked %q across a currency change; the two numbers are not on the "+
				"same scale", row.Label, row.Axis)
		}
	}
	if !strings.Contains(preview.Consequence, "국내") {
		t.Errorf("a USD limit does not say which market's automatic entry closes: %q",
			preview.Consequence)
	}

	// And the first setting of a currency is neither direction either.
	first := card.previewAgainst(config.GuardianLimits{}, false, false)
	if !first.FirstTime || first.CurrencyChange {
		t.Errorf("setting a currency for the first time reads as a change: %+v", first)
	}
	for _, row := range first.Rows {
		if row.Axis != "" {
			t.Errorf("%s is marked %q with nothing to compare against", row.Label, row.Axis)
		}
	}
}

// TestTheCurrencyConsequenceIsRenderedAndNotFolded (never-folded item ⑦).
func TestTheCurrencyConsequenceIsRenderedAndNotFolded(t *testing.T) {
	h := limitsHarness(t, &fakeLimits{gate: fullLimits()})
	h.authenticate(t)
	page := h.page(t, pathSettingsDaily)
	i := strings.Index(page, `data-currency-consequence=`)
	if i < 0 {
		t.Fatal("the screen does not say what the limit currency costs")
	}
	if !strings.Contains(page[max(0, i-60):i], `class="notice"`) {
		t.Error("the currency consequence is muted prose; it is on the never-folded list")
	}
	if !strings.Contains(page, "자동 진입") {
		t.Error("the consequence does not name what closes")
	}
}

// --- no friction (task 4.7) ---------------------------------------------------------

// TestTheSettingsScreensAskForNoTyping is the user instruction of 2026-07-27,
// asserted rather than remembered: no confirmation phrase, no free-text reason,
// no mandatory justification. Settings saves are already audited by the server,
// so an input demanded for traceability would be friction bought with nothing.
func TestTheSettingsScreensAskForNoTyping(t *testing.T) {
	h := fullSettingsHarness(t)
	h.authenticate(t)

	// The fields the settings screen legitimately has. Everything else that takes
	// free text is the thing this test exists to refuse.
	allowed := map[string]bool{
		"csrf": true, "tier": true, "reviewed_sha256": true,
		"exclude_symbols": true, "include_symbols": true, "default_stop_percent": true,
		"max_order_quantity": true, "max_order_notional": true, "max_total_exposure": true,
		"max_daily_loss_amount": true, "max_daily_loss_ratio": true, "limit_currency": true,
		"enabled": true, "place": true, "sell": true, "cancel": true,
		"allow_live_order_actions": true,
	}
	for _, path := range settingsTabPaths {
		page := body(t, h.get(t, path))
		if strings.Contains(page, "<textarea") {
			t.Errorf("%s has a textarea; a settings save does not ask for an essay", path)
		}
		if strings.Contains(page, `type="password"`) {
			t.Errorf("%s asks for a typed phrase", path)
		}
		for _, form := range formBlock.FindAllStringSubmatch(page, -1) {
			for _, field := range inputName.FindAllStringSubmatch(form[2], -1) {
				if !allowed[field[1]] {
					t.Errorf("%s: %s takes a field %q that is not one of the settings this screen "+
						"owns; a required reason or a confirmation phrase is exactly this shape",
						path, form[1], field[1])
				}
			}
		}
		for _, word := range []string{"입력하세요", "정확히 입력", "사유를 입력"} {
			if strings.Contains(page, word) {
				t.Errorf("%s asks the operator to type %q", path, word)
			}
		}
	}
}
