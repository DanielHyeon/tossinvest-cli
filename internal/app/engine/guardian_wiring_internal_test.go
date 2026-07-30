package engine

import (
	"math"
	"reflect"
	"strconv"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

func TestAutomationGateRiskPolicyUsesEveryConfiguredUSDLimit(t *testing.T) {
	gate := config.AutomationGate{
		Enabled:            true,
		MaxOrderQuantity:   100,
		MaxOrderNotional:   300.25,
		MaxTotalExposure:   1000.5,
		MaxDailyLossAmount: 50.75,
		MaxDailyLossRatio:  0.01,
		LimitCurrency:      " usd ",
	}

	got := riskPolicyFromAutomationGate(gate)
	want := risk.DefaultPolicy()
	want.MaxOrderQuantity = "100"
	want.MaxOrderNotional = riskcalc.Money{Amount: "300.25", Currency: "USD"}
	want.MaxOpenExposure = riskcalc.Money{Amount: "1000.5", Currency: "USD"}
	want.MaxDailyLoss = riskcalc.Money{Amount: "50.75", Currency: "USD"}
	want.MaxDailyLossRatio = "0.01"
	want.RiskBudget = riskcalc.Money{Amount: "50.75", Currency: "USD"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("policy = %#v\nwant   = %#v", got, want)
	}

	limits, err := execgw.ExposureLimitsFor(got)
	if err != nil {
		t.Fatalf("ExposureLimitsFor: %v", err)
	}
	if want := gateLimits(gate); !reflect.DeepEqual(limits, want) {
		t.Fatalf("Guardian limits = %#v\ninterlock limits = %#v", limits, want)
	}
}

func TestAutomationGateRiskPolicyPreservesInvalidNumbersForFailClosedValidation(t *testing.T) {
	for _, value := range []float64{0, -1, math.NaN(), math.Inf(1), math.Inf(-1)} {
		t.Run(strconv.FormatFloat(value, 'g', -1, 64), func(t *testing.T) {
			gate := config.AutomationGate{
				Enabled:            true,
				MaxOrderQuantity:   100,
				MaxOrderNotional:   300,
				MaxTotalExposure:   1000,
				MaxDailyLossAmount: 50,
				MaxDailyLossRatio:  value,
				LimitCurrency:      "USD",
			}
			if err := riskPolicyFromAutomationGate(gate).Validate(); err == nil {
				t.Fatalf("Validate accepted daily loss ratio %v", value)
			}
		})
	}
}
