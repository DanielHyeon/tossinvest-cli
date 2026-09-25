# Function Logic Map: `StrategyRuntimeAbsent`

- Source: `internal/httpapi/strategy_runtime.go`
- AST evidence: `ast.json` (편집 전 base `8688f74f`, :56–66, 분기 2 · 반환 3)
- Risk scan: `risk-pattern-report.md`

**Pre-Edit(구현 착수 전 작성).** design D2 가 이 판정을 `internal/strategyprojection` 으로 옮기고 이 함수를
위임 한 줄로 바꾼다 — 기존 함수의 본문 편집이므로 증거가 필요하다(freeze 때 열 벌에 없던 열한 번째 번들,
issues S2). 옮겨 가는 판정은 아래 B1·B2·종단 세 갈래 그대로다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `reader` | nil · presence 를 구현한 reader(wrapper) · 신호 없는 reader(스텁·raw client) | 호출자(router REST·SSE helper·집계 스냅샷·publisher) | 없음 — bool 판정 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if reader == nil {` (:57) | 없음 | `true`(부재) | `TestAbsenceIsAskedAsAStateNotANil` · `TestTheAbsenceJudgementIsOneForBothPackages` |
| B2 | `if presence, ok := reader.(StrategyRuntimePresence); ok {` (:60) | presence 질문 — wrapper 는 재부착 시도를 깨운다(비차단) | `!StrategyRuntimeConfigured()` | `TestAbsenceIsAskedAsAStateNotANil` · `TestTheAbsenceJudgementIsOneForBothPackages` |
| (종단) | 신호 없는 non-nil reader (:65) | 없음 | `false`(있음) | `TestAbsenceIsAskedAsAStateNotANil` · `TestTheAbsenceJudgementIsOneForBothPackages` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `presence.StrategyRuntimeConfigured` | 부재 신호 | 부작용: wake(비차단, rate limit·single-flight) | AST call :61 |

## State mutations and fallbacks

- 자체 상태 없음. 편집 후: 본문이 `strategyprojection.StrategyRuntimeAbsent(reader)` 위임 한 줄 — 세 갈래는 그
  함수로 옮겨 가고 `internal/console` 두 소비자도 같은 판정을 쓴다(판정 한 벌).

## Safety conclusion

- Safe edit boundary: 본문을 위임으로. 시그니처·이름 유지(기존 네 소비처·시험 무편집).
- High-risk impact: no — 조회 전용 projection 의 부재 판정.
