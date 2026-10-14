package config

import (
	"math"
	"testing"
)

// a138: 최대 보유 포지션 수의 해석 — 없음 · 0 은 모든 통화에서 기본값 6, 그 밖은 그대로(검증이 거부하게).
func TestA138OpenPositionMaximumResolution(t *testing.T) {
	cases := []struct {
		name string
		gate AutomationGate
		want float64
	}{
		{"absent KRW takes the default", AutomationGate{LimitCurrency: "KRW"}, 6},
		{"absent USD takes the default too", AutomationGate{LimitCurrency: "USD"}, 6},
		{"absent currency takes the default", AutomationGate{}, 6},
		{"explicit value wins", AutomationGate{LimitCurrency: "KRW", MaxOpenPositions: 4}, 4},
		{"negative passes through for validation", AutomationGate{MaxOpenPositions: -1}, -1},
		{"fraction passes through for validation", AutomationGate{MaxOpenPositions: 2.5}, 2.5},
	}
	for _, c := range cases {
		if got := c.gate.EffectiveMaxOpenPositions(); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
	if DefaultMaxOpenPositions != 6 {
		t.Fatalf("the default maximum is the user's decision of 2026-10-15: 6, got %v", DefaultMaxOpenPositions)
	}
	if got := (AutomationGate{MaxOpenPositions: math.NaN()}).EffectiveMaxOpenPositions(); !math.IsNaN(got) {
		t.Errorf("NaN must pass through to be refused by validation, got %v", got)
	}
}

func TestA138OpenPositionMaximumIsParsedFromTheGateBlock(t *testing.T) {
	svc := writeConfig(t, `{"schema_version":4,"engine":{"automation_gate":{"limit_currency":"KRW","max_open_positions":4}}}`)
	cfg, err := svc.Load(t.Context())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := cfg.Engine.AutomationGate.MaxOpenPositions; got != 4 {
		t.Fatalf("max_open_positions = %v, want 4", got)
	}
	// 정수가 아닌 값도 읽기는 성공함 — 게이트 OFF 설정을 깨지 않고(OFF = upstream), 게이트 ON 정책 검증이 거부함.
	frac := writeConfig(t, `{"schema_version":4,"engine":{"automation_gate":{"enabled":false,"max_open_positions":2.5}}}`)
	cfg, err = frac.Load(t.Context())
	if err != nil {
		t.Fatalf("a fractional maximum must not break reading a gate-off config: %v", err)
	}
	if got := cfg.Engine.AutomationGate.MaxOpenPositions; got != 2.5 {
		t.Fatalf("max_open_positions = %v, want 2.5 as written", got)
	}
	raw, err := svc.LoadRawEngineGate()
	if err != nil {
		t.Fatal(err)
	}
	if raw.MaxOpenPositions != 4 {
		t.Fatalf("the editor read must carry the written maximum, got %v", raw.MaxOpenPositions)
	}
}

// a138: 한도 저장은 OpenPositions 가 있을 때만 그 키를 쓰고, 통화를 바꿔도 지우지 않음(개수는 통화로 재해석되지 않음).
func TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked(t *testing.T) {
	body := `{"schema_version":4,"engine":{"automation_gate":{"enabled":false,"limit_currency":"KRW",` +
		`"max_order_quantity":100,"max_order_notional":1000000,"max_total_exposure":10000000,` +
		`"max_daily_loss_amount":100000,"max_daily_loss_ratio":0.01,"max_open_positions":5}}}`
	untouched := writeConfig(t, body)
	if err := untouched.SaveEngineGateLimits(limitsFor(t, "kr-smoke")); err != nil {
		t.Fatal(err)
	}
	if got := gateOf(t, readBack(t, untouched))["max_open_positions"]; got != float64(5) {
		t.Fatalf("a save without OpenPositions must leave the key alone, got %v", got)
	}
	switched := writeConfig(t, body)
	if err := switched.SaveEngineGateLimits(limitsFor(t, "us-smoke")); err != nil {
		t.Fatal(err)
	}
	if got := gateOf(t, readBack(t, switched))["max_open_positions"]; got != float64(5) {
		t.Fatalf("a currency-changing save must not clear a count, got %v", got)
	}
	asked := writeConfig(t, body)
	l := limitsFor(t, "kr-smoke")
	three := 3.0
	l.OpenPositions = &three
	if err := asked.SaveEngineGateLimits(l); err != nil {
		t.Fatal(err)
	}
	if got := gateOf(t, readBack(t, asked))["max_open_positions"]; got != float64(3) {
		t.Fatalf("a save with OpenPositions must write it, got %v", got)
	}
	for _, bad := range []float64{-1, 2.5, math.NaN(), math.Inf(1)} {
		v := bad
		l.OpenPositions = &v
		if err := asked.SaveEngineGateLimits(l); err == nil {
			t.Fatalf("an open position maximum of %v must be refused before writing", bad)
		}
	}
	zero := 0.0
	l.OpenPositions = &zero
	if err := asked.SaveEngineGateLimits(l); err != nil {
		t.Fatalf("0 (unset → default 6) must be writable, got %v", err)
	}
}

// freeze 리뷰 P2-3: 정수 판정은 한 헬퍼 — 정책 유도 · 인터록 · 콘솔이 같은 답을 씀. math.Trunc(+Inf) == +Inf 라 Inf 를 먼저 거름.
func TestA138WholeOpenPositionsIsTheOneIntegerJudgement(t *testing.T) {
	for _, c := range []struct {
		v    float64
		want int
		ok   bool
	}{
		{6, 6, true}, {1, 1, true}, {1_000, 1_000, true},
		{0, 0, false}, {-1, 0, false}, {2.5, 0, false}, {math.NaN(), 0, false},
		{math.Inf(1), 0, false}, {math.Inf(-1), 0, false}, {1e300, 0, false}, {math.Pow(2, 63), 0, false},
	} {
		got, ok := WholeOpenPositions(c.v)
		if got != c.want || ok != c.ok {
			t.Errorf("WholeOpenPositions(%v) = %d, %v; want %d, %v", c.v, got, ok, c.want, c.ok)
		}
	}
}
