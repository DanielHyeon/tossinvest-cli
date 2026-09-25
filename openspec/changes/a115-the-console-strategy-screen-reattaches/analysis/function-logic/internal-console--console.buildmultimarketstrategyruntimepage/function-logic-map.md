# Function Logic Map: `Console.buildMultiMarketStrategyRuntimePage`

- Source: `internal/console/strategy_runtime_multimarket.go`
- AST evidence: `ast.json` (편집 전 base `8688f74f`, :44–60, 분기 3)
- Risk scan: `risk-pattern-report.md`

a115 편집: `Unwired: c.opts.StrategyRuntime == nil`(:48)과 B1 `c.opts.StrategyRuntime != nil`(:49)을 **한 번 물은** `strategyRuntimeWired(c.opts.StrategyRuntime)` 로 바꾼다. B2(Read 실패·검증 실패 → LoadErr + UnavailableSnapshot)·B3(Clone) 무변경.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c.opts.StrategyRuntime` | nil · 부재 신호를 가진 wrapper · 그 밖의 reader | runConsole | 편집 전: nil 만 dormant |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if c.opts.StrategyRuntime != nil {` (:49) → a115 `if wired {` | Read 1회 | — | `TestAnUnconfiguredWrapperStillRendersDormant`(Read 0회) · `TestStrategyRuntimeMarketsRenderIndependently`(a110 회귀) |
| B2 | Read 오류 또는 검증 실패 | LoadErr=true · UnavailableSnapshot | — | `TestAConfiguredButUnreachableWrapperRendersUnavailable` |
| B3 | else — 유효한 스냅샷 | Clone | — | `TestStrategyRuntimeMarketsRenderIndependently` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strategyRuntimeWired` (a115 새 helper) | 부재 신호 | 부작용: wrapper 의 재부착 시도를 깨운다(rate limit 안) | 새 파일 |
| `StrategyRuntime.Read` | 스냅샷 | 오류 → B2 | AST call :50 |

## State mutations and fallbacks

- page 값만 만든다. 부재 신호 질문은 wrapper 쪽에서 재부착 시도를 깨운다(비차단).

## Safety conclusion

- Safe edit boundary: nil 판정 두 자리 → 한 번 물은 bool. 렌더 값·문구 무변경.
- High-risk impact: no.
