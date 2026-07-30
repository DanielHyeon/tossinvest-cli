package config

// limits_test.go covers the Guardian tier registry, its ceiling backstop and the
// interlock-equivalent validation (change console-sets-guardian-limits, tasks
// 2.x and 3.x).
//
// The registry is a port of StockOS's risk_profiles.py, so the first test is a
// transcription check: an audit that puts the two files side by side must find
// the same numbers. It is deliberately written as literals rather than derived
// from the registry — a test that reads the table it is checking checks nothing.

// The equivalence with execgw.Limits.Validate — the claim that this package
// refuses exactly what the startup interlock refuses — cannot be asserted here:
// internal/execgw reaches internal/config transitively, so importing it from
// this test is an import cycle. It lives in internal/app/engine, which is the
// interlock's own package and already imports both
// (limits_equivalence_test.go).

import (
	"math"
	"testing"
)

// TestTheTiersTranscribeStockOS pins every registered tier against the numbers
// in packages/trading/stockos_trading/risk_profiles.py.
func TestTheTiersTranscribeStockOS(t *testing.T) {
	want := map[string]GuardianLimits{
		// _KR_SMOKE
		"kr-smoke": {
			MaxOrderQuantity: 100, MaxOrderNotional: 500_000,
			MaxTotalExposure: 5_000_000, MaxDailyLossAmount: 100_000,
			MaxDailyLossRatio: 0.01, Currency: "KRW",
		},
		// _KR_SMALL_LIVE — also the five numbers risk-management approved.
		"kr-small-live": {
			MaxOrderQuantity: 100, MaxOrderNotional: 1_000_000,
			MaxTotalExposure: 10_000_000, MaxDailyLossAmount: 100_000,
			MaxDailyLossRatio: 0.01, Currency: "KRW",
		},
		// _US_SMOKE
		"us-smoke": {
			MaxOrderQuantity: 100, MaxOrderNotional: 100,
			MaxTotalExposure: 300, MaxDailyLossAmount: 10,
			MaxDailyLossRatio: 0.01, Currency: "USD",
		},
		// _US_SMALL_LIVE
		"us-small-live": {
			MaxOrderQuantity: 100, MaxOrderNotional: 300,
			MaxTotalExposure: 1_000, MaxDailyLossAmount: 50,
			MaxDailyLossRatio: 0.01, Currency: "USD",
		},
	}

	tiers := GuardianTiers()
	if len(tiers) != len(want) {
		t.Fatalf("GuardianTiers() returned %d tiers, want %d", len(tiers), len(want))
	}
	for _, tier := range tiers {
		expected, ok := want[tier.ID]
		if !ok {
			t.Errorf("unregistered tier %q; adding a tier raises the ceiling and needs its own argument", tier.ID)
			continue
		}
		if tier.Limits != expected {
			t.Errorf("tier %s = %+v, want %+v", tier.ID, tier.Limits, expected)
		}
		if tier.Label == "" {
			t.Errorf("tier %s has no label; the screen shows it before applying", tier.ID)
		}
	}
}

// TestPaperDemoIsNotPorted: StockOS registers a KRX paper_demo tier sized for a
// KIS demo-account seed. Porting it would make 5,000,000 the KRW ceiling by the
// max-across-tiers rule, so its absence is a load-bearing fact (design D3).
func TestPaperDemoIsNotPorted(t *testing.T) {
	for _, tier := range GuardianTiers() {
		if tier.ID == "kr-paper-demo" || tier.Label == "paper_demo" {
			t.Fatalf("paper_demo was ported as %q; it would become the KRW ceiling", tier.ID)
		}
	}
}

// TestTheDefaultTierIsTheApprovedSet: risk-management's 정책 수치의 provenance
// names five numbers. The default tier is exactly those.
func TestTheDefaultTierIsTheApprovedSet(t *testing.T) {
	tier, ok := GuardianTierByID(DefaultGuardianTierID())
	if !ok {
		t.Fatalf("the default tier %q is not registered", DefaultGuardianTierID())
	}
	want := GuardianLimits{
		MaxOrderQuantity: 100, MaxOrderNotional: 1_000_000,
		MaxTotalExposure: 10_000_000, MaxDailyLossAmount: 100_000,
		MaxDailyLossRatio: 0.01, Currency: "KRW",
	}
	if tier.Limits != want {
		t.Errorf("default tier = %+v, want the approved set %+v", tier.Limits, want)
	}
}

// TestTheCeilingIsTheMaxAcrossRegisteredTiers is StockOS ADR §2.6's
// _market_ceiling, per currency.
func TestTheCeilingIsTheMaxAcrossRegisteredTiers(t *testing.T) {
	for _, tc := range []struct {
		currency string
		want     GuardianLimits
	}{
		{"KRW", GuardianLimits{
			MaxOrderQuantity: 100, MaxOrderNotional: 1_000_000,
			MaxTotalExposure: 10_000_000, MaxDailyLossAmount: 100_000,
			MaxDailyLossRatio: 0.01, Currency: "KRW",
		}},
		{"USD", GuardianLimits{
			MaxOrderQuantity: 100, MaxOrderNotional: 300,
			MaxTotalExposure: 1_000, MaxDailyLossAmount: 50,
			MaxDailyLossRatio: 0.01, Currency: "USD",
		}},
	} {
		got, err := GuardianCeiling(tc.currency)
		if err != nil {
			t.Fatalf("GuardianCeiling(%q): %v", tc.currency, err)
		}
		if got != tc.want {
			t.Errorf("ceiling(%s) = %+v, want %+v", tc.currency, got, tc.want)
		}
	}
}

// TestTheCeilingIsCaseInsensitiveOnCurrency: a block spelling "krw" names the
// same ceiling. Fail-closed must not fire on spelling.
func TestTheCeilingIsCaseInsensitiveOnCurrency(t *testing.T) {
	lower, err := GuardianCeiling(" krw ")
	if err != nil {
		t.Fatalf("GuardianCeiling(\" krw \"): %v", err)
	}
	upper, _ := GuardianCeiling("KRW")
	if lower != upper {
		t.Errorf("ceiling(\" krw \") = %+v, want the KRW ceiling %+v", lower, upper)
	}
}

// TestAnUnregisteredCurrencyFailsClosed mirrors StockOS's _market_ceiling: a
// currency with no registered tier has no ceiling, and no ceiling means no save.
func TestAnUnregisteredCurrencyFailsClosed(t *testing.T) {
	for _, currency := range []string{"", "JPY", "EUR"} {
		if _, err := GuardianCeiling(currency); err == nil {
			t.Errorf("GuardianCeiling(%q) returned no error; the override path is fail-closed", currency)
		}
	}
}

// TestCeilingViolationsNameTheFieldAndTheCap: the operator has to be able to fix
// it from the message alone.
func TestCeilingViolationsNameTheFieldAndTheCap(t *testing.T) {
	over := GuardianLimits{
		MaxOrderQuantity: 100, MaxOrderNotional: 2_000_000,
		MaxTotalExposure: 10_000_000, MaxDailyLossAmount: 100_000,
		MaxDailyLossRatio: 0.01, Currency: "KRW",
	}
	violations := over.CeilingViolations()
	if len(violations) != 1 {
		t.Fatalf("violations = %v, want exactly the notional one", violations)
	}
	if !contains(violations[0], "max_order_notional") || !contains(violations[0], "1000000") {
		t.Errorf("violation %q names neither the field nor the cap", violations[0])
	}
}

// TestEveryFieldIsCapped: a ceiling that only checks some fields is a ceiling
// with a hole, and the hole is where a fat finger goes.
func TestEveryFieldIsCapped(t *testing.T) {
	base, _ := GuardianCeiling("KRW")
	for _, tc := range []struct {
		name string
		bend func(*GuardianLimits)
	}{
		{"quantity", func(l *GuardianLimits) { l.MaxOrderQuantity *= 2 }},
		{"notional", func(l *GuardianLimits) { l.MaxOrderNotional *= 2 }},
		{"exposure", func(l *GuardianLimits) { l.MaxTotalExposure *= 2 }},
		{"daily loss", func(l *GuardianLimits) { l.MaxDailyLossAmount *= 2 }},
		{"ratio", func(l *GuardianLimits) { l.MaxDailyLossRatio *= 2 }},
	} {
		over := base
		tc.bend(&over)
		if got := over.CeilingViolations(); len(got) == 0 {
			t.Errorf("raising the %s past the ceiling produced no violation", tc.name)
		}
	}
}

// TestTheCeilingItselfIsAcceptable: the boundary is inclusive, or no preset
// could be saved — every tier sits at the ceiling on at least one field.
func TestTheCeilingItselfIsAcceptable(t *testing.T) {
	for _, tier := range GuardianTiers() {
		if got := tier.Limits.CeilingViolations(); len(got) != 0 {
			t.Errorf("tier %s cannot be saved through its own ceiling: %v", tier.ID, got)
		}
	}
}

// TestLoweringIsNeverACeilingViolation is the conservative direction (§0.6).
func TestLoweringIsNeverACeilingViolation(t *testing.T) {
	tier, _ := GuardianTierByID("kr-small-live")
	low := tier.Limits
	low.MaxOrderQuantity = 1
	low.MaxOrderNotional = 1
	low.MaxTotalExposure = 1
	low.MaxDailyLossAmount = 1
	low.MaxDailyLossRatio = 0.0001
	if got := low.CeilingViolations(); len(got) != 0 {
		t.Errorf("lowering every limit produced violations: %v", got)
	}
}

// guardianCandidates is this package's corpus for the validation rules.
//
// The interlock equivalence test in internal/app/engine does NOT share this
// list — it cannot, since a helper in a _test.go file is invisible across
// packages, and exporting a corpus from production code to serve a test is
// worse than the drift it would prevent. That test generates its candidates
// mechanically from the field set instead, which is drift-proof by
// construction.
func guardianCandidates() []GuardianLimits {
	ok := GuardianLimits{
		MaxOrderQuantity: 100, MaxOrderNotional: 1_000_000,
		MaxTotalExposure: 10_000_000, MaxDailyLossAmount: 100_000,
		MaxDailyLossRatio: 0.01, Currency: "KRW",
	}
	mutate := func(f func(*GuardianLimits)) GuardianLimits {
		out := ok
		f(&out)
		return out
	}
	return []GuardianLimits{
		ok,
		mutate(func(l *GuardianLimits) { l.MaxOrderQuantity = 0 }),
		mutate(func(l *GuardianLimits) { l.MaxOrderNotional = 0 }),
		mutate(func(l *GuardianLimits) { l.MaxTotalExposure = 0 }),
		mutate(func(l *GuardianLimits) { l.MaxDailyLossAmount = 0 }),
		mutate(func(l *GuardianLimits) { l.MaxDailyLossRatio = 0 }),
		mutate(func(l *GuardianLimits) { l.MaxOrderQuantity = -1 }),
		mutate(func(l *GuardianLimits) { l.MaxDailyLossRatio = 1 }),
		mutate(func(l *GuardianLimits) { l.MaxDailyLossRatio = 1.5 }),
		mutate(func(l *GuardianLimits) { l.MaxOrderNotional = math.NaN() }),
		mutate(func(l *GuardianLimits) { l.MaxTotalExposure = math.Inf(1) }),
		mutate(func(l *GuardianLimits) { l.Currency = "" }),
		mutate(func(l *GuardianLimits) { l.Currency = "   " }),
	}
}

// TestValidationRefusesWhatTheInterlockRefuses states the rules directly. The
// equivalence with execgw.Limits.Validate over the same corpus is asserted in
// internal/app/engine/limits_equivalence_test.go.
func TestValidationRefusesWhatTheInterlockRefuses(t *testing.T) {
	candidates := guardianCandidates()
	if err := candidates[0].Validate(); err != nil {
		t.Fatalf("the approved set was refused: %v", err)
	}
	// A ratio of exactly 1 is a ceiling of the whole capital: permitted by the
	// interlock (it refuses only > 1), and so permitted here.
	for i, c := range candidates[1:] {
		if c.Validate() == nil && c.MaxDailyLossRatio != 1 {
			t.Errorf("candidate %d %+v was accepted", i+1, c)
		}
	}
}

// TestMatchingTierNamesTheRegistryRow — the screen says which tier the file
// currently spells, and says nothing when it spells none of them (design D12).
func TestMatchingTierNamesTheRegistryRow(t *testing.T) {
	tier, _ := GuardianTierByID("us-smoke")
	if got := MatchingGuardianTier(tier.Limits); got != "us-smoke" {
		t.Errorf("MatchingGuardianTier(us-smoke limits) = %q, want us-smoke", got)
	}
	custom := tier.Limits
	custom.MaxDailyLossAmount = 7
	if got := MatchingGuardianTier(custom); got != "" {
		t.Errorf("MatchingGuardianTier(one field changed) = %q, want the empty string", got)
	}
	if got := MatchingGuardianTier(GuardianLimits{}); got != "" {
		t.Errorf("MatchingGuardianTier(zero) = %q, want the empty string", got)
	}
}

// TestMatchingTierIgnoresCurrencySpelling: "krw" and "KRW" are the same block.
func TestMatchingTierIgnoresCurrencySpelling(t *testing.T) {
	tier, _ := GuardianTierByID("kr-smoke")
	lower := tier.Limits
	lower.Currency = "krw"
	if got := MatchingGuardianTier(lower); got != "kr-smoke" {
		t.Errorf("MatchingGuardianTier with lower-case currency = %q, want kr-smoke", got)
	}
}

// TestParsingStillInventsNothing is design D1 stated as a test in this change's
// own name: the config layer does not inject a default tier, so the screen and
// the engine's interlock read the same file.
func TestParsingStillInventsNothing(t *testing.T) {
	svc := writeConfig(t, `{"schema_version":5,"trading":{},"engine":{"automation_gate":{"enabled":false}}}`)
	cfg, err := svc.Load(t.Context())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Engine.AutomationGate.LimitsSet() {
		t.Errorf("parsing injected limits %+v; an implicit default splits the screen from the engine",
			cfg.Engine.AutomationGate)
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && indexOf(haystack, needle) >= 0
}

func indexOf(haystack, needle string) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}
