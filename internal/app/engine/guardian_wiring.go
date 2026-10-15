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

// a139: v1 → v2 — 건당 위험 예산 규칙이 일일 손실 전부에서 1/3 로 바뀜. 결정 preimage 의 PolicyVersion 이 그 결정이 어느 사이징 규칙으로
// 나왔는지 스스로 말하게 함(생산 첫 다리는 기대 버전을 Guardian 에서 읽으므로 영향 없음).
const productionGuardianPolicyVersion = "engine.automation_gate/risk-policy-v2"

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
// RiskBudget is not a sixth operator permission: it is derived from the audited
// daily-loss ceiling — a139: a third of it, rounded down to the currency's minor
// unit (risk.PerTradeRiskBudget, user decision 2026-10-15), so one stopped-out
// trade can no longer spend the whole day's loss allowance.
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
	// a139: 건당 위험 예산 = 일일 손실의 1/3(최소 단위 내림). 유도할 수 없으면(표에 없는 통화 · 0 으로 내려감) 금액을 비워
	// Policy.Validate 가 거부하게 함 — 일일 손실 전부로 대체하지 않음(완화 금지).
	if budget, err := risk.PerTradeRiskBudget(dailyLoss); err == nil {
		policy.RiskBudget = budget
	} else {
		policy.RiskBudget = riskcalc.Money{Currency: currency}
	}
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
	// a139(freeze 리뷰 P3): 위험 예산은 정책으로 옮기며 빈 값이 되므로, 그 전에 유도 실패를 이름으로 밝혀 거부함.
	if audit := perTradeRiskBudgetAudit(gate); audit.value == "" {
		return nil, fmt.Errorf("%w: production Guardian policy: per-trade risk budget %s", ErrLimitsRequired, audit.detail)
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

// perTradeRiskBudgetRule 은 기동 audit 가 적는 유도 규칙(a139) — 「일일 손실이 바뀜」과 「규칙이 바뀜」을 기록으로 가르게 함.
const perTradeRiskBudgetRule = "floor(max_daily_loss_amount / 3, minor unit) — a139"

// riskBudgetAudit 은 기동 audit · 인터록 문구가 적는 건당 위험 예산 — 값은 생산 정책 유도(riskPolicyFromAutomationGate)의 답 그대로.
type riskBudgetAudit struct{ value, detail string }

// perTradeRiskBudgetAudit 은 게이트에서 유도한 예산을 audit 모양으로 돌려줌. 유도할 수 없으면 빈 값과 사유 — 오류를 돌려주지 않음:
// 이 줄은 게이트 ON/OFF 와 무관하게 Guardian 생성 전에 기록되므로, 실패가 게이트 OFF 기동을 막으면 OFF = upstream 이 깨짐(freeze 리뷰 P1-2).
func perTradeRiskBudgetAudit(gate config.AutomationGate) riskBudgetAudit {
	budget := riskPolicyFromAutomationGate(gate).RiskBudget
	if strings.TrimSpace(budget.Amount) == "" {
		_, err := risk.PerTradeRiskBudget(riskcalc.Money{
			Amount:   strconv.FormatFloat(gate.MaxDailyLossAmount, 'f', -1, 64),
			Currency: strings.ToUpper(strings.TrimSpace(gate.LimitCurrency)),
		})
		reason := "is not set"
		if err != nil {
			reason = err.Error()
		}
		return riskBudgetAudit{detail: "cannot be derived: " + reason + " (rule " + perTradeRiskBudgetRule + ")"}
	}
	return riskBudgetAudit{value: budget.Amount, detail: "derived by " + perTradeRiskBudgetRule}
}
