# Function Logic Map: `ExitObserver.ObserveOnce`

- Source: `internal/app/engine/exitloop.go` (467-539)
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- 기준: 편집 뒤(브랜치 `a163-impl` GREEN 로트, 2026-10-22). 편집 전 번들 문서는 커밋 `78efe1951` · 생성 입력 `../../flm-gen/spec-base.json`
- AST 분기 8 · 반환 5 · defer 2 · go 0
- 위험 등급: High-risk(청산 관측 주기 — 손절 판정 순회)

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `o.opts.SLO` | nil 또는 압력 원천 | 배선 | 뒤처지면 주기 양보(B1) |
| `states` | 작업 집합 | `workingSet` | 오류면 B2 |
| `quotes` | 관측 맵 | `observe` | 오류면 B4 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if 475 `if o.opts.SLO != nil && o.opts.SLO.FillDetectionBehind() {` | `cycle.Deferred`, `checkOutage` | `return cycle` | 기존 시험 |
| B2 | if 485 `if err != nil {` | `cycle.Err` | `return cycle` | 기존 시험 |
| B3 | if 489 `if len(states) == 0 {` | `settleUnobserved`, 두절 시계 리셋 | `return cycle` | 기존 시험 |
| B4 | if 502 `if err != nil {` | `cycle.Err`, `checkOutage` | `return cycle` | `TestAFailedObservationHoldsTheJudgement` |
| B5 | range 511 `for _, state := range states {` | 판정 순회 | — | 기존 시험 |
| B6 | if 513 `if !ok {` | `noteUnobservedCause(no_quote)` | `continue` | `TestA090R2AnAbsentSiblingIsCountedLikeAZeroQuote` |
| B7 | if 521 `if !o.quoteUsable(quote) {` | `noteUnobservedCause(quote_expired)` | `continue` | `TestA090R7AQuoteThatExpiredMidCycleIsCountedAsExpired` |
| B8 | if 529 `if err := o.judge(ctx, state, quote, observation, &cycle); err != nil && cycle.Err == nil {` | `cycle.Err` 첫 오류만 | — | 기존 시험 |

반환 좌표(AST): 481 · 487 · 499 · 505 · 538

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `o.workingSet` | 보유·대상 포지션 | 오류는 B2 | AST |
| `o.observe` | 주기의 단일 시세 읽기 | 오류는 B4 | AST |
| `o.quoteUsable` / `o.judge` | 판정 | 판정 오류는 `cycle.Err` | AST |
| `o.settleUnobserved` | a090 순회 뒤 집계 정리 | 실패는 로그 | AST |
| defer `o.endExitOrderCycle` · `o.flushStrategyProtectionAlerts` | a129 · 전략 보호 알림 | — | AST(defer 2) |

## State mutations and fallbacks

- `o.lastObserved` · `o.outageRaised` · `o.protectionAlerts` · a090 주기 집계.

## a163 편집

- a163 3.3: B5 판정 순회 **뒤**, `o.settleUnobserved(ctx, &cycle)` 바로 앞에 보고 호출 한 줄(`o.reportSourceAge(ctx, &cycle)` — 토글 판단은 그 함수 안, design D4·D8). 분기 8 · 반환 5 · defer 2 불변. 순회 안 새 코드 0. — **적용됨**, 편집 전·뒤 AST 분기·반환·defer 수 대조는 `../../flm-gen/README.md`

## Safety conclusion

- Safe edit boundary: 순회 밖 호출 한 줄. B3(무보유)·조기 반환 경로에는 보고 없음.
- High-risk impact: yes — 순서(2.7)·패닉(2.10)·토글 위치(㉒) 시험.
