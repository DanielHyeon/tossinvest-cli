# Function Logic Map: `deliverEachStrategyHandoff`

- Source: `internal/app/engine/strategy_market_handoff_delivery.go` (62-89, 편집 뒤)
- Qualified: `deliverEachStrategyHandoff`
- AST evidence: `ast.json`(편집 뒤, `source_sha256` e972e7224635b825…) · `ast-base.json`(편집 전, base `a000cf29`, `source_sha256` fb503c7549f1fd1f…, 편집 전 분기수 4) — 현재 branches 5 · returns 4
- Risk scan: `risk-pattern-report.md`

**역할.** handoff 를 조정자 순서대로 몸통에 건네고 범위 거절만 건너뜀(J4).

**a137 편집.** 몸통 오류가 `*execgw.DailyLossEntryBlock`(errors.As)이면 남은 전달을 멈추고 `errors.Join(skipped...)` 를 돌려줌(B4 추가) — 앞서 건너뛴 범위 거절이 없으면 nil(주기 무오류 → runMarket B11, 삼킴 · 잠금 없음), 있으면 지금처럼 그 보고. 분기 4 → 5. 구현 리뷰 P1-3: `*execgw.DailyLossUnmeasurable` 도 같은 B4 에서 조용히 끝냄.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| handoffs | 조정자 순서 | 조정자 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for _, handoff := range handoffs {` (range 64) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine -run ScopeRefusal` |
| B2 | `if err == nil {` (if 66) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine -run ScopeRefusal` |
| B3 | `if errors.As(err, &scope) {` (if 70) | — | 분기 본문 참조 | 기존 `TestOnlyATypedScopeRefusalIsSkippedEveryOtherFaultStops` |
| B4 | `if errors.As(err, &blocked) \|\| errors.As(err, &unmeasurable) {` (if 78) | — | 분기 본문 참조 | `TestA137ADailyLossModeBlockEndsTheCycleQuietly` · `TestA137AMixedCycleStillReportsItsSkippedScopes` · `TestA137OtherModeBlocksBehaveAsBefore` · `TestA137AnUnmeasurableDayIsAPolicyOutcomeNotALaneFault` · `TestA137TheDailyLossStopsNeverLatchOrSwallowOnEitherLaneKind`(effective · refreshOnly 둘) |
| B5 | `if len(skipped) == 0 {` (if 83) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine -run ScopeRefusal` |

Exact AST return positions: 80 · 84 · 86 · 88

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `handoff.Deliver` | 전달 | 오류 분류 | AST calls |

## State mutations and fallbacks

- 없음.

## Safety conclusion

- Safe edit boundary: 표지 오류의 분류 하나.
- High-risk impact: yes(전략 주기) — 주문을 더 내지 않는 방향(뒤 범위 전달 멈춤).
