package engine

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

const productionGuardianPolicyVersion = "engine.automation_gate/risk-policy-v1"

type productionGuardianFactory func(
	config.AutomationGate,
	*journal.Journal,
	clock.Clock,
	string,
	journal.ModeAnnouncer,
) (execgw.Guardian, error)

// riskPolicyFromAutomationGate transcribes the one audited source of production
// exposure limits into the policy the Guardian evaluates.
//
// The non-limit policy fields deliberately remain risk.DefaultPolicy values.
// RiskBudget is not a sixth operator permission: using the configured daily-loss
// ceiling is the largest per-trade budget that adds no permission beyond that
// already audited ceiling.
func riskPolicyFromAutomationGate(gate config.AutomationGate) risk.Policy {
	number := func(value float64) string {
		// The risk decimal parser deliberately rejects exponent notation. 'f'
		// with precision -1 is still the shortest round-tripping decimal, but
		// spells 1,000,000 as "1000000" rather than "1e+06".
		return strconv.FormatFloat(value, 'f', -1, 64)
	}
	currency := strings.ToUpper(strings.TrimSpace(gate.LimitCurrency))
	dailyLoss := riskcalc.Money{
		Amount:   number(gate.MaxDailyLossAmount),
		Currency: currency,
	}

	policy := risk.DefaultPolicy()
	policy.MaxOrderQuantity = number(gate.MaxOrderQuantity)
	policy.MaxOrderNotional = riskcalc.Money{
		Amount:   number(gate.MaxOrderNotional),
		Currency: currency,
	}
	policy.MaxOpenExposure = riskcalc.Money{
		Amount:   number(gate.MaxTotalExposure),
		Currency: currency,
	}
	policy.MaxDailyLoss = dailyLoss
	policy.MaxDailyLossRatio = number(gate.MaxDailyLossRatio)
	policy.RiskBudget = dailyLoss
	return policy
}

func newProductionRiskGuardian(
	gate config.AutomationGate,
	jrn *journal.Journal,
	clk clock.Clock,
	accountRef string,
	announcer journal.ModeAnnouncer,
) (*execgw.RiskGuardian, error) {
	policy := riskPolicyFromAutomationGate(gate)
	if err := policy.Validate(); err != nil {
		return nil, fmt.Errorf("%w: production Guardian policy: %v", ErrLimitsRequired, err)
	}
	guardian, err := execgw.NewRiskGuardian(execgw.RiskGuardianOptions{
		Journal:       jrn,
		Clock:         clk,
		AccountRef:    accountRef,
		Policy:        policy,
		Costs:         costs.DefaultModel(),
		PolicyVersion: productionGuardianPolicyVersion,
		Announcer:     announcer,
	})
	if err != nil {
		return nil, fmt.Errorf("engine: constructing the production Guardian: %w", err)
	}
	return guardian, nil
}

func defaultProductionGuardianFactory(
	gate config.AutomationGate,
	jrn *journal.Journal,
	clk clock.Clock,
	accountRef string,
	announcer journal.ModeAnnouncer,
) (execgw.Guardian, error) {
	return newProductionRiskGuardian(gate, jrn, clk, accountRef, announcer)
}
