package console

import (
	"net/url"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
)

// a138: 콘솔 한도 편집기의 최대 보유 포지션 수 — 정수 입력, 미설정이면 기본 6 을 평이하게 말함, 칸이 폼에 있을 때만 씀,
// 프리셋 · 통화 변경이 건드리지 않음, 타이핑 확인 없음.

func a138PostLimits(t *testing.T, h *dashboardHarness, extra url.Values) {
	t.Helper()
	form := url.Values{
		"csrf": {h.csrf}, "max_order_quantity": {"5"}, "max_order_notional": {"200000"},
		"max_total_exposure": {"400000"}, "max_daily_loss_amount": {"20000"}, "max_daily_loss_ratio": {"0.005"},
		"limit_currency": {"KRW"}, "max_daily_buy_notional": {""},
	}
	for k, v := range extra {
		form[k] = v
	}
	h.post(t, "/settings/limits", form)
}

func TestA138TheAdvancedFormSavesTheOpenPositionMaximum(t *testing.T) {
	seam := &fakeLimits{}
	h := limitsHarness(t, seam)
	h.authenticate(t)
	a138PostLimits(t, h, url.Values{"max_open_positions": {"4"}})
	writes := seam.writes()
	if len(writes) != 1 || writes[0].OpenPositions == nil || *writes[0].OpenPositions != 4 {
		t.Fatalf("writes = %+v, want one save carrying 4", writes)
	}

	blank := &fakeLimits{}
	h = limitsHarness(t, blank)
	h.authenticate(t)
	a138PostLimits(t, h, url.Values{"max_open_positions": {""}})
	if w := blank.writes(); len(w) != 1 || w[0].OpenPositions == nil || *w[0].OpenPositions != 0 {
		t.Fatalf("a blank field is unset (0 → default 6), got %+v", w)
	}

	// 칸이 없는 폼(옛 화면)은 운영자 값을 건드리지 않음.
	absent := &fakeLimits{}
	h = limitsHarness(t, absent)
	h.authenticate(t)
	a138PostLimits(t, h, nil)
	if w := absent.writes(); len(w) != 1 || w[0].OpenPositions != nil {
		t.Fatalf("a form without the field must leave the key alone, got %+v", w)
	}

	for _, bad := range []string{"lots", "2.5", "-1"} {
		refused := &fakeLimits{}
		h = limitsHarness(t, refused)
		h.authenticate(t)
		a138PostLimits(t, h, url.Values{"max_open_positions": {bad}})
		if len(refused.writes()) != 0 {
			t.Fatalf("an open position maximum of %q must not be written", bad)
		}
	}
}

func TestA138APresetLeavesTheOpenPositionMaximumAlone(t *testing.T) {
	seam := &fakeLimits{gate: config.AutomationGate{LimitCurrency: "KRW", MaxOpenPositions: 4}}
	h := limitsHarness(t, seam)
	h.authenticate(t)
	h.post(t, "/settings/limits/preset", url.Values{"csrf": {h.csrf}, "tier": {"kr-small-live"}})
	writes := seam.writes()
	if len(writes) != 1 || writes[0].OpenPositions != nil {
		t.Fatalf("a preset must not carry the open position maximum, got %+v", writes)
	}
}

func TestA138TheScreenSaysWhatTheOpenPositionMaximumIs(t *testing.T) {
	h := limitsHarness(t, &fakeLimits{gate: config.AutomationGate{LimitCurrency: "USD"}})
	h.authenticate(t)
	if current := limitCurrent(t, h.page(t, pathSettingsDaily)); !strings.Contains(current, "기본값 6 종목이 적용된다") {
		t.Error("an unset maximum must say the default 6 applies, in every currency")
	}
	h = limitsHarness(t, &fakeLimits{gate: config.AutomationGate{LimitCurrency: "KRW", MaxOpenPositions: 2.5}})
	h.authenticate(t)
	if current := limitCurrent(t, h.page(t, pathSettingsDaily)); !strings.Contains(current, "엔진이 기동을 거부한다") {
		t.Error("an unusable maximum must say the engine refuses to start")
	}
	h = limitsHarness(t, &fakeLimits{gate: config.AutomationGate{LimitCurrency: "KRW", MaxOpenPositions: 4}})
	h.authenticate(t)
	page := h.page(t, pathSettingsDaily)
	if !strings.Contains(page, `name="max_open_positions" value="4"`) || !strings.Contains(page, `inputmode="numeric"`) ||
		!strings.Contains(page, `aria-describedby="open-positions-help"`) || !strings.Contains(page, `id="open-positions-help"`) {
		t.Error("the advanced form must carry the open position field, prefilled, with its help text linked")
	}
	if strings.Contains(strings.ToLower(page), "type the") || strings.Contains(page, "confirm-text") {
		t.Error("no typed confirmation may guard the open position maximum")
	}
}
