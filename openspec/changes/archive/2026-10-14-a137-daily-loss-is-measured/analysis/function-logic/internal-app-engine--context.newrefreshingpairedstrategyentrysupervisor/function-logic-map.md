# Function Logic Map: `Context.NewRefreshingPairedStrategyEntrySupervisor`

- Source: `internal/app/engine/strategy_entry_supervisor.go` (399-428, 편집 뒤)
- Qualified: `Context.NewRefreshingPairedStrategyEntrySupervisor`
- AST evidence: `ast.json`(편집 뒤, `source_sha256` 6f7c999a13c4abdc…) · `ast-base.json`(편집 전, base `63fd0c92`, `source_sha256` 9cb510c1c9c7f44a…, 편집 전 분기수 4) — 현재 branches 4 · returns 4
- Risk scan: `risk-pattern-report.md`

**역할.** 생산 전략 감독자 조립.

**a137 편집.** 재리뷰 P1-N1: `DailyLossNotice: c.strategyDailyLossNotice()`(경고 줄 하나, 로그 없으면 nil). 분기 불변(4).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| c · clk | 엔진 Context | 조립 | B1 · B2 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if c == nil \|\| clk == nil {` (if 400) | — | 분기 본문 참조 | 기존 조립 시험 · `TestA137TheProductionNoticeIsOneWarningLine` |
| B2 | `if c.Entry == nil {` (if 404) | — | 분기 본문 참조 | 기존 조립 시험 · `TestA137TheProductionNoticeIsOneWarningLine` |
| B3 | `for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {` (range 408) | — | 분기 본문 참조 | 기존 조립 시험 · `TestA137TheProductionNoticeIsOneWarningLine` |
| B4 | `if err != nil {` (if 421) | — | 분기 본문 참조 | 기존 조립 시험 · `TestA137TheProductionNoticeIsOneWarningLine` |

Exact AST return positions: 401 · 405 · 422 · 427

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `c.strategyDailyLossNotice` | 알림 자리(a137) | — | AST calls |

## State mutations and fallbacks

- c.strategySupervisor(기존).

## Safety conclusion

- Safe edit boundary: 옵션 칸 하나.
- High-risk impact: no — 관측.
