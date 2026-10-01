package scheduler

import (
	"crypto/sha256"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

// a112 7.1 — 전략 capability 의 `(market, horizon, family, poll_class)` subscope 결합(Manager 판정 Q1~Q4, 2026-10-01).
//
// 범위는 capability 를 **묶을** 뿐 용량을 나누지 않는다: 모든 범위는 같은 물리 endpoint · reset 세대의 남은 수 · 안전 예비 · commitment 집합 · 발급
// 상한 · observation cycle 을 공유한다(`tryAcquire` 가 범위와 무관하게 endpoint 하나의 집합에서 판정). 범위가 하는 일은 하나 — 발급 때의 범위와
// 다른 범위(또는 범위 없음)로는 그 capability 를 완료할 수 없게 하는 것(교차 family replay 금지).
//
// family → horizon 의 레인 표 정합은 강제하지 않는다 — 스펙은 결합만 요구한다(필요해지면 6.1 처럼 strategyrouter 정본 표에서 유도).
// 생산 호출자는 0 이다(핀 `TestTheStrategyBudgetAPIHasNoProductionCallerYet`) — 레인 evidence polling 배선은 7.5 또는 활성화 로트의 몫.

// StrategyBudgetDeferred 는 전략 subscope API 의 유일한 거절이다(동결 골든 `refusal_enums.scheduler`). 세부 사유는 Detail.
const StrategyBudgetDeferred = "BUDGET_DEFERRED"

// StrategyScope 는 전략 capability 의 admission subscope 다.
type StrategyScope struct {
	Market  strategyrouter.Market
	Horizon strategyrouter.Horizon
	Family  strategyrouter.Family
}

// Valid 는 범위의 세 값이 모두 정규 이름인지다.
func (scope StrategyScope) Valid() bool {
	return (scope.Market == strategyrouter.MarketKR || scope.Market == strategyrouter.MarketUS) &&
		(scope.Horizon == strategyrouter.HorizonShort || scope.Horizon == strategyrouter.HorizonWeekly) && scope.Family.Known()
}

func (scope StrategyScope) digest() [sha256.Size]byte {
	return sha256.Sum256([]byte("TossOS/scheduler-strategy-scope/v1\x00" + string(scope.Market) + "\x00" + string(scope.Horizon) + "\x00" + string(scope.Family)))
}

// StrategyCommitmentToken 은 범위에 묶인 capability 다. 안의 토큰은 범위 없는 완료 API 로 꺼낼 수 없다.
type StrategyCommitmentToken struct {
	token CommitmentToken
	scope [sha256.Size]byte
}

// StrategyBudgetGrant 는 전략 subscope 발급 결과다. 거절이면 Refusal = BUDGET_DEFERRED 이고 Detail 이 세부 사유다.
type StrategyBudgetGrant struct {
	Allowed    bool
	Refusal    string
	Detail     BudgetReason
	Remaining  int
	Reserve    int
	Available  int
	Reset      time.Time
	ObservedAt time.Time
	Commitment StrategyCommitmentToken
}

// TryAcquireStrategy 는 범위에 묶인 low-priority capability 하나를 발급한다. 안전 등급은 전략 API 의 대상이 아니다(안전 예비를 쓰지 않음).
func (c *BudgetCoordinator) TryAcquireStrategy(key string, scope StrategyScope, class PollClass, now time.Time) StrategyBudgetGrant {
	if !scope.Valid() {
		return StrategyBudgetGrant{Refusal: StrategyBudgetDeferred, Detail: BudgetInvalidBounds}
	}
	if !isKnownPollClass(class) || isSafetyClass(class) {
		return StrategyBudgetGrant{Refusal: StrategyBudgetDeferred, Detail: BudgetUnknownClass}
	}
	digest := scope.digest()
	grant := c.tryAcquire(key, class, now, digest)
	result := StrategyBudgetGrant{Allowed: grant.Allowed, Detail: grant.Reason, Remaining: grant.Remaining, Reserve: grant.Reserve,
		Available: grant.Available, Reset: grant.Reset, ObservedAt: grant.ObservedAt}
	if !grant.Allowed {
		result.Refusal = StrategyBudgetDeferred
		return result
	}
	result.Commitment = StrategyCommitmentToken{token: grant.Commitment, scope: digest}
	return result
}

// CompleteStrategy 는 범위에 묶인 capability 하나를 완료한다. 발급 때와 다른 범위 · 범위 없는 토큰 · 위조 토큰은 false 이고 상태는 바뀌지 않는다.
func (c *BudgetCoordinator) CompleteStrategy(key string, scope StrategyScope, token StrategyCommitmentToken) bool {
	if !scope.Valid() || token.scope != scope.digest() {
		return false
	}
	return c.complete(key, token.token, token.scope)
}
