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
	// a136: 하루 매수 총액 상한 — 해석(없음·0 → KRW 기본값, 다른 통화는 미설정)은 config 한 곳. 미설정·음수·NaN 은
	// Policy.Validate 가 거부해 기동이 ErrLimitsRequired 로 멈춤(D5).
	policy.MaxDailyBuyNotional = riskcalc.Money{
		Amount:   number(gate.EffectiveMaxDailyBuyNotional()),
		Currency: currency,
	}
	// a138: 최대 보유 포지션 수 — 해석 · 정수 판정은 config 한 곳. 쓸 수 없는 값이면 0 이 되어 Policy.Validate 가 거부함.
	policy.MaxOpenPositions, _ = config.WholeOpenPositions(gate.EffectiveMaxOpenPositions())
	return policy
}

func newProductionRiskGuardian(
	gate config.AutomationGate,
	jrn *journal.Journal,
	clk clock.Clock,
	accountRef string,
	announcer journal.ModeAnnouncer,
) (*execgw.RiskGuardian, error) {
	// a138(freeze 리뷰 P2-3): 개수 상한은 정책으로 옮기며 0 으로 바뀌므로, 그 전에 운영자가 적은 값을 담아 거부함.
	if v := gate.EffectiveMaxOpenPositions(); !wholeOpenPositions(v) {
		return nil, fmt.Errorf("%w: production Guardian policy: max_open_positions is %v; it must be a whole number of at least 1",
			ErrLimitsRequired, v)
	}
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
		// a136 D9=b: 엔진 밖 수동 매수 보드 — 보조 실행자 manual-buy-refresh 가 게시함(Context.ManualBuyRefresher).
		// 첫 게시 전에는 진입이 전부 거부됨(fail-closed).
		ManualBuys: execgw.NewManualBuyBoard(),
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

// wholeOpenPositions 는 config.WholeOpenPositions 의 참거짓 반쪽 — 판정은 그 한 곳.
func wholeOpenPositions(v float64) bool {
	_, ok := config.WholeOpenPositions(v)
	return ok
}
