# Function Logic Map: `StrategyEntrySupervisor.Snapshot`

- Source: `internal/app/engine/strategy_entry_supervisor.go` (782-804, 편집 뒤)
- Qualified: `StrategyEntrySupervisor.Snapshot`
- AST evidence: `ast.json`(편집 뒤, `source_sha256` 6f7c999a13c4abdc…) · `ast-base.json`(편집 전, base `63fd0c92`, `source_sha256` 9cb510c1c9c7f44a…, 편집 전 분기수 2) — 현재 branches 2 · returns 3
- Risk scan: `risk-pattern-report.md`

**역할.** 시장별 읽기 전용 스냅숏.

**a137 편집.** 재리뷰 P1-N1: `DailyLossUnmeasurableCycles` · `FirstDailyLossUnmeasurable` 두 칸을 채움. 분기 불변(2).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| market | KR · US | 호출자 | 무효 → B1 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if s == nil \|\| !validStrategyMarket(market) {` (if 783) | — | 분기 본문 참조 | 기존 스냅숏 시험 |
| B2 | `if worker == nil {` (if 789) | — | 분기 본문 참조 | `TestA137AnUnmeasurableDayIsCountedAndAnnouncedOncePerAccountDay`(effective · refreshOnly) |

Exact AST return positions: 784 · 790 · 792

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| — | — | — | — |

## State mutations and fallbacks

- 없음(읽기).

## Safety conclusion

- Safe edit boundary: 칸 둘.
- High-risk impact: no — 관측.
