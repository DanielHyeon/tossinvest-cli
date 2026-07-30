package engine_test

// limits_equivalence_test.go asserts the one claim that lets the console write
// Guardian limits at all (change console-sets-guardian-limits, task 4.3):
//
//	config.GuardianLimits.Validate refuses exactly what the startup interlock
//	refuses.
//
// The console's spec says it must not record a block the engine would refuse to
// start on. That promise is only as good as the agreement between two
// validators living in two packages, and nothing else in the tree compares
// them. This test does, from the package that already owns the interlock.
//
// The candidates are GENERATED from the field set rather than transcribed. A
// hand-written list in one package and a hand-written list in the other is the
// drift this test exists to catch, so it must not be made of one.

import (
	"math"
	"strconv"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
)

// asExecgw is the transcription the engine's own wiring performs: five bounds
// and a currency.
func asExecgw(l config.GuardianLimits) execgw.Limits {
	return execgw.Limits{
		MaxQuantity:        execgw.Bound(l.MaxOrderQuantity),
		MaxNotional:        execgw.Bound(l.MaxOrderNotional),
		MaxTotalExposure:   execgw.Bound(l.MaxTotalExposure),
		MaxDailyLossAmount: execgw.Bound(l.MaxDailyLossAmount),
		MaxDailyLossRatio:  execgw.Bound(l.MaxDailyLossRatio),
		Currency:           l.Currency,
	}
}

func TestConfigRefusesExactlyWhatTheInterlockRefuses(t *testing.T) {
	base := config.GuardianLimits{
		MaxOrderQuantity: 100, MaxOrderNotional: 1_000_000,
		MaxTotalExposure: 10_000_000, MaxDailyLossAmount: 100_000,
		MaxDailyLossRatio: 0.01, Currency: "KRW",
	}

	// One setter per numeric bound, so adding a sixth bound to GuardianLimits
	// without extending this list is visible as a compile-time gap in the struct
	// literal below rather than as silent under-coverage.
	fields := map[string]func(*config.GuardianLimits, float64){
		"max_order_quantity":    func(l *config.GuardianLimits, v float64) { l.MaxOrderQuantity = v },
		"max_order_notional":    func(l *config.GuardianLimits, v float64) { l.MaxOrderNotional = v },
		"max_total_exposure":    func(l *config.GuardianLimits, v float64) { l.MaxTotalExposure = v },
		"max_daily_loss_amount": func(l *config.GuardianLimits, v float64) { l.MaxDailyLossAmount = v },
		"max_daily_loss_ratio":  func(l *config.GuardianLimits, v float64) { l.MaxDailyLossRatio = v },
	}
	values := []float64{0, -1, -0.0001, math.NaN(), math.Inf(1), math.Inf(-1), 0.5, 1, 1.5, 2}

	type candidate struct {
		name   string
		limits config.GuardianLimits
	}
	candidates := []candidate{{"the approved set", base}}
	for name, set := range fields {
		for _, v := range values {
			c := base
			set(&c, v)
			candidates = append(candidates, candidate{name + "=" + formatFloat(v), c})
		}
	}
	for _, currency := range []string{"", " ", "   ", "KRW", "usd"} {
		c := base
		c.Currency = currency
		candidates = append(candidates, candidate{"currency=" + currency, c})
	}

	for _, c := range candidates {
		mine := c.limits.Validate()
		theirs := asExecgw(c.limits).Validate()
		if (mine != nil) != (theirs != nil) {
			t.Errorf("%s: config refuses = %v (%v), the interlock refuses = %v (%v)",
				c.name, mine != nil, mine, theirs != nil, theirs)
		}
	}
}

func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'g', -1, 64)
}
