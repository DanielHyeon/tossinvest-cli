# Function Logic Map: `Journal.runApplyHooks`

- Source: `internal/journal/apply_hook.go` (편집 전 `264`–`296`, base `233d5d9c`)
- AST evidence: `ast.json` (편집 전 `source_sha256` 8daed316f087979f…) — branches 8 · returns 6 · defers 1 · go 0
- Risk scan: `risk-pattern-report.md`
- 작성 시점: **편집 전**(tasks 0.5).

**역할.** 체결 트랜잭션 안에서 Project → Campaign(+위험 소유자) → Exit 훅을 차례로 부른다. 핸들은 끝나면 무효화.
호출자(CodeGraph 1.6.0, `analysis/code-context/raw/`): 생산 `RecordFill`(fills.go) · `backfillConfirmedStrategyFillTx`(strategy_dispatch_runtime.go). 시험 3.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `tx` | 열린 *sql.Tx | 호출자 | — |
| `fill` | AppliedFill | 호출자 | — |
| `j.applyHooks` | SetApplyHooks 가 한 번 묶음 | applyMu | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `:268` 훅 셋 다 nil | 없음 | nil `:269` | 훅 없는 저널 시험 |
| B2 | `:275` Project | — | — | 전부 |
| B3 | `:276` Project 오류 | — | err `:277` | 프로젝션 실패 시험 |
| B4 | `:280` Campaign | — | — | 캠페인 시험 |
| B5 | `:281` Campaign 오류 | — | err | 있음 |
| B6 | `:285` 위험 소유자 바인딩 오류 | — | err | risk_bucket_owner_test |
| B7 | `:289` Exit | — | — | exit 시험 |
| B8 | `:290` Exit 오류 | — | err `:291` | 있음 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `hooks.Project/Campaign/Exit` | 주입 훅 | 오류 → 롤백 | AST |
| `applyRiskBucketOwnerBindingInTx` | 위험 소유자 | 오류 → 롤백 | AST |

Live config binding: 없음.

## State mutations and fallbacks

- 핸들 `&ApplyTx{tx, now, costs}` 생성(:273) · defer invalidate.

## a130 편집 계획

핸들 리터럴(:273)에 `outcomeRefused: hooks.OutcomeRefused` 필드 하나. 분기 · 순서 · 반환 무변경.

## Safety conclusion

- Safe edit boundary: 위 「a130 편집 계획」 만. 주문 · 손절 · 체결 반영의 동작 무변경.
- High-risk impact: yes(원장 / 엔진 기동) — 거절이던 성과가 행이 되는 방향, 거절 기록 추가.
