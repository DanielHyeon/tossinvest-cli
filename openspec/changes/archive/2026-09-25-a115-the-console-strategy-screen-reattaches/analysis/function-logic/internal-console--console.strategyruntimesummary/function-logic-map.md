# Function Logic Map: `Console.strategyRuntimeSummary`

- Source: `internal/console/settings_tabs.go`
- AST evidence: `ast.json` — **구현 후 재생성**(revision current, :290–303, 분기 2). 편집 전 base `8688f74f` 는 :290–301, 분기 2.
- Risk scan: `risk-pattern-report.md`

## 편집 전후 대조 (옛/새 ast.json difflib 정렬)

분기 수·종류 그대로, B1 만 **replace**: 편집 전 `if c.opts.StrategyRuntime == nil {`(:291) → 편집 후
`if strategyprojection.StrategyRuntimeAbsent(c.opts.StrategyRuntime) {`(:293) — 설정 화면 요약도 전략 페이지와 **같은
판정**(정정의 단위는 값이다). B2 텍스트 그대로(줄 +2, 주석 두 줄).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c.opts.StrategyRuntime` | nil · presence 를 말하는 wrapper · 신호 없는 reader | runConsole | 부재 → dormant 문구(Read 0) · 있음+실패 → 「읽지 못함」 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if strategyprojection.StrategyRuntimeAbsent(c.opts.StrategyRuntime) {` (:293) | 없음(판정의 wake 부작용 1회) | "… — dormant 미배선" | `TestTheSummaryAsksThePresenceSignalToo` · `TestStrategyRuntimeSummaryUsesPairedDormantTruth` |
| B2 | `if err != nil \|\| strategyprojection.Validate(snapshot) != nil {` (:297) | 없음 | "읽지 못함 — …" | `TestStrategyRuntimeSummaryReportsReadFailure` · `TestTheSummaryAsksThePresenceSignalToo` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strategyprojection.StrategyRuntimeAbsent` | 부재 신호(한 벌 판정) | 부작용: wrapper 재부착 시도 깨우기(비차단) | AST call :293 |
| `c.opts.StrategyRuntime.Read` | 스냅샷 | 오류 → B2 | AST call :296 |

## State mutations and fallbacks

- 없음(문자열 반환).

## Safety conclusion

- Safe edit boundary: 조건 한 줄.
- High-risk impact: no.
