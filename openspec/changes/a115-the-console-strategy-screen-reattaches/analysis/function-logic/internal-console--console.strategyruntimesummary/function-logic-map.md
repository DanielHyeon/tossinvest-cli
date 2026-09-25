# Function Logic Map: `Console.strategyRuntimeSummary`

- Source: `internal/console/settings_tabs.go`
- AST evidence: `ast.json` (편집 전 base `8688f74f`, :290–301, 분기 2)
- Risk scan: `risk-pattern-report.md`

a115 편집: B1(:291) `c.opts.StrategyRuntime == nil` 을 `!strategyRuntimeWired(c.opts.StrategyRuntime)` 로 — 설정 화면 요약도 같은 판정(정정의 단위는 값이다). B2 무변경.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c.opts.StrategyRuntime` | 위와 같음 | runConsole | 편집 전: nil 만 dormant 문구 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if c.opts.StrategyRuntime == nil {` (:291) → `!strategyRuntimeWired(...)` | 없음 | "… — dormant 미배선" | `TestTheSummaryAsksThePresenceSignalToo` · `TestStrategyRuntimeSummaryUsesPairedDormantTruth` |
| B2 | Read 오류·검증 실패 | 없음 | "읽지 못함 — …" | `TestStrategyRuntimeSummaryReportsReadFailure` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strategyRuntimeWired` | 부재 신호 | 위와 같음 | 새 파일 |

## State mutations and fallbacks

- 없음(문자열 반환).

## Safety conclusion

- Safe edit boundary: 조건 한 줄.
- High-risk impact: no.
