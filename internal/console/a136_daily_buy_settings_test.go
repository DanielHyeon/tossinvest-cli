package console

import (
	"net/url"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
)

// a136: 콘솔 한도 편집기의 하루 매수 총액 상한 — 현재 통화의 자유 숫자 입력, USD 프리셋 없음, 미설정의 실제 결과를 평이하게 말함,
// 프리셋은 그 값을 건드리지 않음, 타이핑 확인 없음.

func a136PostLimits(t *testing.T, h *dashboardHarness, daily string) {
	t.Helper()
	h.post(t, "/settings/limits", url.Values{
		"csrf": {h.csrf}, "max_order_quantity": {"5"}, "max_order_notional": {"200000"},
		"max_total_exposure": {"400000"}, "max_daily_loss_amount": {"20000"}, "max_daily_loss_ratio": {"0.005"},
		"limit_currency": {"KRW"}, "max_daily_buy_notional": {daily},
	})
}

func TestA136TheAdvancedFormSavesTheDailyBuyCeiling(t *testing.T) {
	seam := &fakeLimits{}
	h := limitsHarness(t, seam)
	h.authenticate(t)
	a136PostLimits(t, h, "5,000,000")
	writes := seam.writes()
	if len(writes) != 1 || writes[0].DailyBuy == nil || *writes[0].DailyBuy != 5_000_000 {
		t.Fatalf("writes = %+v, want one save carrying 5,000,000", writes)
	}

	refused := &fakeLimits{}
	h = limitsHarness(t, refused)
	h.authenticate(t)
	a136PostLimits(t, h, "lots")
	if len(refused.writes()) != 0 {
		t.Fatal("a non-numeric daily buy ceiling must not be written")
	}
}

func TestA136APresetLeavesTheDailyBuyCeilingAlone(t *testing.T) {
	seam := &fakeLimits{gate: config.AutomationGate{LimitCurrency: "KRW", MaxDailyBuyNotional: 3_000_000}}
	h := limitsHarness(t, seam)
	h.authenticate(t)
	h.post(t, "/settings/limits/preset", url.Values{"csrf": {h.csrf}, "tier": {"kr-small-live"}})
	writes := seam.writes()
	if len(writes) != 1 || writes[0].DailyBuy != nil {
		t.Fatalf("a preset must not carry the daily buy ceiling, got %+v", writes)
	}
}

func TestA136TheScreenSaysWhatAnUnsetDailyBuyCeilingDoes(t *testing.T) {
	for currency, want := range map[string]string{
		"KRW": "기본값 10,000,000 KRW 가 적용된다",
		"USD": "엔진이 기동을 거부한다",
	} {
		h := limitsHarness(t, &fakeLimits{gate: config.AutomationGate{LimitCurrency: currency}})
		h.authenticate(t)
		if current := limitCurrent(t, h.page(t, pathSettingsDaily)); !strings.Contains(current, want) {
			t.Errorf("%s: the current-value table must say %q", currency, want)
		}
	}
	h := limitsHarness(t, &fakeLimits{gate: config.AutomationGate{LimitCurrency: "KRW", MaxDailyBuyNotional: 7_000_000}})
	h.authenticate(t)
	page := h.page(t, pathSettingsDaily)
	if current := limitCurrent(t, page); !strings.Contains(current, "7,000,000") {
		t.Error("a set daily buy ceiling must render as its number")
	}
	if !strings.Contains(page, `name="max_daily_buy_notional" value="7000000"`) ||
		!strings.Contains(page, `aria-describedby="daily-buy-help"`) || !strings.Contains(page, `id="daily-buy-help"`) {
		t.Error("the advanced form must carry the daily buy field, prefilled, with its help text linked")
	}
	if strings.Contains(strings.ToLower(page), "type the") || strings.Contains(page, "confirm-text") {
		t.Error("no typed confirmation may guard the daily buy ceiling")
	}
}
