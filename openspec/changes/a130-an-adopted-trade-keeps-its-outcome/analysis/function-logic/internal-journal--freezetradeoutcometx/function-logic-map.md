# Function Logic Map: `freezeTradeOutcomeTx`

- Source: `internal/journal/trade_outcomes.go` (편집 전 `129`–`145`, base `233d5d9c`)
- AST evidence: `ast.json` (편집 전 `source_sha256` 0bd43abf96107b99…) — branches 1 · returns 2 · defers 0 · go 0
- Risk scan: `risk-pattern-report.md`
- 작성 시점: **편집 전**(tasks 0.5).

**역할.** CLOSED 체결 트랜잭션 안에서 성과 행을 계산해 INSERT 한다. 오류를 돌려주지 않는다(분석이 주문 경로를 깨지 못함 — 파일 머리 주석).
호출자(CodeGraph 1.6.0, `analysis/code-context/raw/`): 생산 `ApplyExitFill`(apply_hook.go:1023) 1 — 반환값을 버림.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `tx` | 살아 있는 ApplyTx | runApplyHooks | 닫힌 핸들이면 Query/Exec 가 ErrApplyTxClosed → false |
| `positionID` | CLOSED 포지션 | ApplyExitFill `where.PositionID` | 없는 id → compute false |
| `model` | `tx.costs` | ApplyHooks.Costs | 미구성 → compute false |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `:131` `!ok`(computeTradeOutcome 거절) | 없음 | `return false` `:132` | 거절 시험(TestAnAnalyticsFailureDoesNotRollBackTheClose · TestAnUnpricedFillLeavesNoOutcome · TestAPartlyEngineSoldAdoptedPositionFreezesNothing) |
| (fallthrough) | INSERT 실행 | trade_outcomes 한 행 | `return err == nil` `:144` — INSERT 오류도 false | TestTheOutcomeIsFrozenInTheClosingTransaction(성공); INSERT 실패 시험 없음 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `computeTradeOutcome` | 계산 | bool | AST |
| `tx.Exec` | INSERT | 오류 → false(삼킴) | AST |

Live config binding: 없음.

## State mutations and fallbacks

- trade_outcomes INSERT 한 줄(같은 트랜잭션). 실패는 삼킨다 — 반환 bool 로만 알림(그 bool 을 호출자가 버린다).

## a130 편집 계획

반환형 bool → 사유 문자열("" = 성공). B1 은 compute 의 사유를, INSERT 오류는 `insert_failed` 를 돌려준다. 행을 쓰는 조건은 그대로.

## Safety conclusion

- Safe edit boundary: 위 「a130 편집 계획」 만. 주문 · 손절 · 체결 반영의 동작 무변경.
- High-risk impact: yes(원장 / 엔진 기동) — 거절이던 성과가 행이 되는 방향, 거절 기록 추가.
