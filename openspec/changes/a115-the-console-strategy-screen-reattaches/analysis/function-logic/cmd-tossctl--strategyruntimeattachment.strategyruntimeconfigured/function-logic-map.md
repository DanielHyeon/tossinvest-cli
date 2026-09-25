# Function Logic Map: `strategyRuntimeAttachment.StrategyRuntimeConfigured`

- Source: `cmd/tossctl/httpapi_strategy_attach.go`
- AST evidence: `ast.json` (편집 전 base `8688f74f`, :113–117, 분기 0)
- Risk scan: `risk-pattern-report.md`

**편집하지 않는다(근거 인용 전용).** 분기 0 — 자리 상태를 읽고 **무조건 wake** 한 뒤 `reader != nil` 을 답한다(a109 G2). 콘솔 화면이 이것을 물으면 그 질문이 재부착 시도를 깨운다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 자리 | nil(부재)·sentinel·client | wrapper | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (분기 없음) | — | wake(비차단, rate limit) | `reader != nil` | `TestAnUnconfiguredWrapperStillRendersDormant`(간접) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `a.state` · `a.wake` | 상태·시도 | 비차단 | AST |

## State mutations and fallbacks

- 편집 없음.

## Safety conclusion

- Safe edit boundary: 편집 없음.
- High-risk impact: no.
