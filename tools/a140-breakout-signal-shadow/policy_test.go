package main

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestEffectivePolicyDigestCoversRuntimePolicy(t *testing.T) {
	base := defaultProbeConfig()
	base.Params = testParams()
	baseDigest := effectivePolicyFor(base, sessionSourceCalendar).digest()
	cases := []struct {
		name   string
		mutate func(*probeConfig)
	}{
		{"top", func(c *probeConfig) { c.Top++ }},
		{"retest", func(c *probeConfig) { c.Params.RetestTolerancePPM++ }},
		{"quote age", func(c *probeConfig) { c.Params.MaxQuoteAgeMS++ }},
		{"spread", func(c *probeConfig) { c.Params.MaxSpreadPPM++ }},
		{"drift", func(c *probeConfig) { c.Params.MaxEntryDriftPPM++ }},
		{"risk", func(c *probeConfig) { c.Params.RiskBudgetKRW++ }},
		{"notional", func(c *probeConfig) { c.Params.NotionalCapKRW++ }},
		{"cost", func(c *probeConfig) { c.Params.CostPPM++ }},
		{"minimum RR", func(c *probeConfig) { c.Params.MinRRPPM++ }},
		{"read gap", func(c *probeConfig) { c.ReadGap += time.Nanosecond }},
		{"retry delay", func(c *probeConfig) { c.RetryDelays[0] += time.Nanosecond }},
		{"quote suspension", func(c *probeConfig) { c.QuoteSuspend += time.Nanosecond }},
		{"armed attempts", func(c *probeConfig) { c.ArmedAttempts++ }},
		{"armed window", func(c *probeConfig) { c.ArmedWindow += time.Nanosecond }},
		{"gap persistence", func(c *probeConfig) { c.GapPersist += time.Nanosecond }},
		{"label grace", func(c *probeConfig) { c.LabelGrace += time.Nanosecond }},
		{"budget cadence", func(c *probeConfig) { c.BudgetEvery += time.Nanosecond }},
		{"poll lookback", func(c *probeConfig) { c.PollLookback += time.Nanosecond }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := base
			cfg.RetryDelays = append([]time.Duration(nil), base.RetryDelays...)
			tc.mutate(&cfg)
			if got := effectivePolicyFor(cfg, sessionSourceCalendar).digest(); got == baseDigest {
				t.Fatal("policy digest did not change")
			}
		})
	}
	if got := effectivePolicyFor(base, sessionSourceManual).digest(); got == baseDigest {
		t.Fatal("session source did not change policy digest")
	}
}

func TestEffectivePolicyIsCanonicalAndExcludesRuntimeSource(t *testing.T) {
	cfg := defaultProbeConfig()
	cfg.Params = testParams()
	cfg.ParamsSource = "/private/config.json token=secret"
	policy := effectivePolicyFor(cfg, sessionSourceCalendar)
	raw, err := json.Marshal(policy)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), cfg.ParamsSource) {
		t.Fatalf("canonical policy retained runtime source: %s", raw)
	}
	if policy.digest() == "" || lotForPolicy(policy.digest()) == "" {
		t.Fatal("policy digest or lot is empty")
	}
}

func TestEffectivePolicyLotsKeepDifferentSessionsBelowVerdictThreshold(t *testing.T) {
	base := defaultProbeConfig()
	base.Params = testParams()
	changed := base
	changed.Params.RiskBudgetKRW++
	lotA := lotForPolicy(effectivePolicyFor(base, sessionSourceCalendar).digest())
	lotB := lotForPolicy(effectivePolicyFor(changed, sessionSourceCalendar).digest())
	if lotA == lotB {
		t.Fatal("changed effective policy reused a lot")
	}
	rows := make([]ledgerRow, 0, 30)
	for i := 0; i < 29; i++ {
		rows = append(rows, ledgerRow{Kind: kindLabel, Session: "KRX:2026-08-14", Lot: lotA, Outcome: "target", RPPM: 1_500_000})
	}
	rows = append(rows, ledgerRow{Kind: kindLabel, Session: "KRX:2026-08-15", Lot: lotB, Outcome: "target", RPPM: 1_500_000})
	report := renderReport(rows)
	for _, want := range []string{lotA + ": 라벨 29 — 결과 판정 보류(n<30)", lotB + ": 라벨 1 — 결과 판정 보류(n<30)"} {
		if !strings.Contains(report, want) {
			t.Fatalf("policy lots were not separately withheld for %q:\n%s", want, report)
		}
	}
	if strings.Contains(report, "라벨 30 · 목표 선착") {
		t.Fatalf("different policy lots were combined into a verdict:\n%s", report)
	}
}
