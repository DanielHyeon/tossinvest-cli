# Function Logic Map: `Console.buildMultiMarketStrategyRuntimePage`

- Source: `internal/console/strategy_runtime_multimarket.go`
- AST evidence: `ast.json` — **구현 후 재생성**(revision current, :44–64, 분기 3). 편집 전 base `8688f74f` 는 :44–60, 분기 3.
- Risk scan: `risk-pattern-report.md`

## 편집 전후 대조 (옛/새 ast.json difflib 정렬)

분기 수·종류 그대로, B1 만 **replace**: 편집 전 `if c.opts.StrategyRuntime != nil {`(:49) → 편집 후 `if !absent {`(:53).
그 앞에 `absent := strategyprojection.StrategyRuntimeAbsent(c.opts.StrategyRuntime)`(:50, 분기 아님)를 **한 번** 묻고
`Unwired: absent`(:52)와 B1 두 자리에 쓴다 — 편집 전 두 자리의 `== nil` 이 한 벌 판정 하나로 바뀌었다.
B2(Read·Validate 실패)·B3(else Clone)은 텍스트 그대로(줄만 +4). 첫 설계의 새 helper `strategyRuntimeWired` 는
freeze 리뷰 P1-1 로 기각돼 쓰지 않았다 — 판정은 `strategyprojection` 한 벌이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c.opts.StrategyRuntime` | nil(진짜 미배선) · presence 를 말하는 wrapper(부재·sentinel·live) · 신호 없는 reader(스텁·raw client) | runConsole | 부재 → dormant(Read 0회) · 있음+실패 → 도달 불가 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if !absent {` (:53) — 부재 신호가 「있음」 | Read 1회 | — | `TestAnUnconfiguredWrapperStillRendersDormant`(부재 → Read 0·질문 1) · `TestASignallessReaderIsStillWired` · `TestStrategyRuntimeDormantPairIsHonest` |
| B2 | `if err != nil \|\| strategyprojection.Validate(value) != nil {` (:55) — 읽기·검증 실패 | LoadErr=true · UnavailableSnapshot | — | `TestAConfiguredButUnreachableWrapperRendersUnavailable` · `TestStrategyRuntimeReaderFailureAndInvalidProjectionFailClosedWithoutLeakingError` |
| B3 | `} else {` (:58) — 유효한 스냅샷 | Clone | — | `TestStrategyRuntimeMarketsRenderIndependently` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strategyprojection.StrategyRuntimeAbsent` | 부재 신호(한 벌 판정) | 부작용: wrapper 의 재부착 시도를 깨운다(비차단, rate limit) — 그래서 한 번만 묻는다(변이 K16) | AST call :50 |
| `c.opts.StrategyRuntime.Read` | 스냅샷 | 오류 → B2 | AST call :54 |

## State mutations and fallbacks

- page 값만 만든다. 부재 신호 질문은 wrapper 쪽에서 재부착 시도를 깨운다(비차단). 요청 goroutine 은 dial 하지 않는다.

## Safety conclusion

- Safe edit boundary: nil 판정 두 자리 → 한 번 물은 bool. 렌더 값·문구 무변경.
- High-risk impact: no.
