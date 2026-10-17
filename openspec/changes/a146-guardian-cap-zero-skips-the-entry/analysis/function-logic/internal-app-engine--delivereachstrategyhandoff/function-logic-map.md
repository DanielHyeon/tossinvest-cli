# Function Logic Map: `deliverEachStrategyHandoff`

- Source: `internal/app/engine/strategy_market_handoff_delivery.go` (66-98)
- Qualified: `deliverEachStrategyHandoff`
- AST evidence: `ast.json`(현재) — 현재 AST branches 6 · returns 4
- 편집 전 기준: `ast-base.json`(base, `source_sha256` 44ed84069d5204e7…) — 편집 전 갈래 수 5
- Risk scan: `risk-pattern-report.md`

**역할.** 한 시장 주기의 handoff 를 조정자 순서대로 몸통에 건넴 — 범위 거절은 건너뛰어 모으고, 일일 손실 표지는 주기를 조용히 끝내며, 그 밖은 멈춤.

**a146 편집.** 범위 거절 판정 다음에 `*strategyEntrySkip` 이면 `continue`(모으지 않음 — 주기 오류가 되지 않게). 새 분기 하나.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 상태 | 소스 참조 | 현재 HEAD | 분기 표 참조 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for _, handoff := range handoffs {` (range 68) | — | 조정자 순서대로 하나씩 | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|Scope|Handoff|Delivery'` · `TestA146ASkipContinuesToTheNextScopeWithoutAnError` · `TestA146ASkipIsNeverJoinedIntoTheCycleError` · `TestA146OnlyAPureSkipIsSkippedEveryLookalikeStops` · `TestOnlyATypedScopeRefusalIsSkippedEveryOtherFaultStops` · `TestA137AMixedCycleStillReportsItsSkippedScopes` |
| B2 | `if err == nil {` (if 70) | — | 다음 범위 | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|Scope|Handoff|Delivery'` · `TestA146ASkipContinuesToTheNextScopeWithoutAnError` · `TestA146ASkipIsNeverJoinedIntoTheCycleError` · `TestA146OnlyAPureSkipIsSkippedEveryLookalikeStops` · `TestOnlyATypedScopeRefusalIsSkippedEveryOtherFaultStops` · `TestA137AMixedCycleStillReportsItsSkippedScopes` |
| B3 | `if errors.As(err, &scope) {` (if 74) | skipped 에 모음 | 다음 범위(끝에 Join 으로 보고 — 무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|Scope|Handoff|Delivery'` · `TestA146ASkipContinuesToTheNextScopeWithoutAnError` · `TestA146ASkipIsNeverJoinedIntoTheCycleError` · `TestA146OnlyAPureSkipIsSkippedEveryLookalikeStops` · `TestOnlyATypedScopeRefusalIsSkippedEveryOtherFaultStops` · `TestA137AMixedCycleStillReportsItsSkippedScopes` |
| B4 | `if _, isSkip := strategyEntrySkipOf(err); isSkip {` (if 80) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|Scope|Handoff|Delivery'` · `TestA146ASkipContinuesToTheNextScopeWithoutAnError` · `TestA146ASkipIsNeverJoinedIntoTheCycleError` · `TestA146OnlyAPureSkipIsSkippedEveryLookalikeStops` · `TestOnlyATypedScopeRefusalIsSkippedEveryOtherFaultStops` · `TestA137AMixedCycleStillReportsItsSkippedScopes` |
| B5 | `if errors.As(err, &blocked) \|\| errors.As(err, &unmeasurable) {` (if 87) | — | Join(skipped) — 일일 손실 계좌 차단은 주기를 조용히 끝냄(무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|Scope|Handoff|Delivery'` · `TestA146ASkipContinuesToTheNextScopeWithoutAnError` · `TestA146ASkipIsNeverJoinedIntoTheCycleError` · `TestA146OnlyAPureSkipIsSkippedEveryLookalikeStops` · `TestOnlyATypedScopeRefusalIsSkippedEveryOtherFaultStops` · `TestA137AMixedCycleStillReportsItsSkippedScopes` |
| B6 | `if len(skipped) == 0 {` (if 92) | — | 멈추는 결함 그대로 / skipped 와 Join(무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|Scope|Handoff|Delivery'` · `TestA146ASkipContinuesToTheNextScopeWithoutAnError` · `TestA146ASkipIsNeverJoinedIntoTheCycleError` · `TestA146OnlyAPureSkipIsSkippedEveryLookalikeStops` · `TestOnlyATypedScopeRefusalIsSkippedEveryOtherFaultStops` · `TestA137AMixedCycleStillReportsItsSkippedScopes` |

Exact AST return positions: 89 · 93 · 95 · 97

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (호출 6종 — 대표: `append`, `errors.As`, `errors.Join`, `handoff.Deliver`, `len`, `strategyEntrySkipOf`) | 본문 참조 | 소스 참조 | AST |

## State mutations and fallbacks

- 편집은 위 「a146 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a146 편집」에 적은 한 자리.
- High-risk impact: yes(전략 첫 다리 진입 분류 — 주문을 늘리지 않음: 건너뛴 진입은 브로커 0, 매도 · 손절 · 청산 경로 무관)
