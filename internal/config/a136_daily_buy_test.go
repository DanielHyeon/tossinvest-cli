package config

import (
	"math"
	"testing"
)

// a136 D5: 하루 매수 총액 상한의 해석 — 없음·0 은 KRW 에서만 기본값, 다른 통화는 미설정(→ 기동 거부).
func TestA136DailyBuyCeilingResolution(t *testing.T) {
	cases := []struct {
		name string
		gate AutomationGate
		want float64
	}{
		{"absent KRW takes the default", AutomationGate{LimitCurrency: "KRW"}, 10_000_000},
		{"absent lower-case krw takes the default", AutomationGate{LimitCurrency: " krw "}, 10_000_000},
		{"absent USD stays unset", AutomationGate{LimitCurrency: "USD"}, 0},
		{"absent currency stays unset", AutomationGate{}, 0},
		{"explicit KRW value wins", AutomationGate{LimitCurrency: "KRW", MaxDailyBuyNotional: 5_000_000}, 5_000_000},
		{"explicit USD value wins", AutomationGate{LimitCurrency: "USD", MaxDailyBuyNotional: 2_000}, 2_000},
		{"negative passes through for validation", AutomationGate{LimitCurrency: "KRW", MaxDailyBuyNotional: -1}, -1},
	}
	for _, c := range cases {
		if got := c.gate.EffectiveMaxDailyBuyNotional(); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
	if got := (AutomationGate{LimitCurrency: "KRW", MaxDailyBuyNotional: math.NaN()}).EffectiveMaxDailyBuyNotional(); !math.IsNaN(got) {
		t.Errorf("NaN must pass through to be refused by validation, got %v", got)
	}
}

func TestA136DailyBuyCeilingIsParsedFromTheGateBlock(t *testing.T) {
	svc := writeConfig(t, `{"schema_version":4,"engine":{"automation_gate":{"limit_currency":"KRW","max_daily_buy_notional":3000000}}}`)
	cfg, err := svc.Load(t.Context())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := cfg.Engine.AutomationGate.MaxDailyBuyNotional; got != 3_000_000 {
		t.Fatalf("max_daily_buy_notional = %v, want 3000000", got)
	}
	absent := writeConfig(t, `{"schema_version":4,"engine":{"automation_gate":{"limit_currency":"KRW"}}}`)
	cfg, err = absent.Load(t.Context())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := cfg.Engine.AutomationGate.MaxDailyBuyNotional; got != 0 {
		t.Fatalf("an absent key must read as 0 (unset), got %v", got)
	}
}

// a136 freeze 리뷰 P1-2: 콘솔 한도 저장이 통화를 바꾸면 옛 통화로 적힌 하루 매수 상한을 지워(0) 재해석을 막음.
// 같은 통화 저장은 그 키를 건드리지 않음(편집기의 칸이 아니므로).
func TestA136ACurrencyChangingSaveClearsTheDailyBuyCeiling(t *testing.T) {
	body := `{"schema_version":4,"engine":{"automation_gate":{"enabled":false,"limit_currency":"KRW",` +
		`"max_order_quantity":100,"max_order_notional":1000000,"max_total_exposure":10000000,` +
		`"max_daily_loss_amount":100000,"max_daily_loss_ratio":0.01,"max_daily_buy_notional":5000000}}}`

	same := writeConfig(t, body)
	if err := same.SaveEngineGateLimits(limitsFor(t, "kr-smoke")); err != nil {
		t.Fatalf("same-currency save: %v", err)
	}
	if got := gateOf(t, readBack(t, same))["max_daily_buy_notional"]; got != float64(5000000) {
		t.Fatalf("a same-currency save must leave the daily buy ceiling alone, got %v", got)
	}

	changed := writeConfig(t, body)
	if err := changed.SaveEngineGateLimits(limitsFor(t, "us-smoke")); err != nil {
		t.Fatalf("currency-changing save: %v", err)
	}
	if got := gateOf(t, readBack(t, changed))["max_daily_buy_notional"]; got != float64(0) {
		t.Fatalf("a KRW figure must not survive as a USD ceiling; want 0 (unset), got %v", got)
	}
	cfg, err := changed.Load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.Engine.AutomationGate.EffectiveMaxDailyBuyNotional(); got != 0 {
		t.Fatalf("after switching to USD the ceiling must be unset (startup refusal), got %v", got)
	}
}
