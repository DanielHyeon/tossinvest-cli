# Function Logic Map: `ApplyExitFill`

- Source: `internal/journal/apply_hook.go` (편집 전 `961`–`1026`, base `233d5d9c`)
- AST evidence: `ast.json` (편집 전 `source_sha256` 8daed316f087979f…) — branches 15 · returns 11 · defers 0 · go 0
- Risk scan: `risk-pattern-report.md`
- 작성 시점: **편집 전**(tasks 0.5).

**역할.** Exit apply hook — 체결 트랜잭션 안에서 taken 비율 이동 · 제안 해소 · 완료 표시 · 성과 동결을 한다. 판정 경로 값은 못 씀.
호출자(CodeGraph 1.6.0, `analysis/code-context/raw/`): 함수 값: `gateway.go:191` `Exit: journal.ApplyExitFill` → `runApplyHooks` B8 이 부름(CodeGraph 는 동적 호출을 못 봄 — reconciliation R1).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `tx` | 살아 있는 ApplyTx | runApplyHooks | — |
| `fill` | AppliedFill | RecordFill | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `:963` exitContextForFill 오류 | 없음 | err `:964` | 기존 exit apply 시험 |
| B2 | `:966` 대상 없음 | 없음 | nil `:969` | 외부 주문 시험 |
| B3 | `:973` exit 상태 없음 | 없음 | nil `:976` | TestAnExternalPositionHasNoOutcome |
| B4 | `:978` PendingState 오류 | 없음 | err | 없음 |
| B5 | `:982` SELL | — | — | 매도 시험 |
| B6 | `:985` TakenAfterFill 오류 | 없음 | err | 없음 |
| B7 | `:989` CompareDecimal 오류 | 없음 | err | 없음 |
| B8/B9 | `:992` cmp > 0 | — | — | 부분 익절 시험 |
| B10 | `:993` MoveTakenRatioTotal 오류 | — | err | 없음 |
| B11 | `:999` 종결 · 대기 제안 일치 | ResolvePending | — | 제안 해소 시험 |
| B12 | `:1001` ResolvePending 오류 | — | err | 없음 |
| B13 | `:1004` exit 이벤트 오류 | — | err | 없음 |
| B14 | `:1013` CLOSED ∧ 미완료 | markExitCompletedTx · **freezeTradeOutcomeTx(:1023, 반환 버림)** | — | TestTheOutcomeIsFrozenInTheClosingTransaction · TestAnAnalyticsFailureDoesNotRollBackTheClose |
| B15 | `:1014` markExitCompletedTx 오류 | — | err `:1015` | 없음 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `freezeTradeOutcomeTx` | 성과 동결 | 반환 bool 을 버림(:1023) — a130 의 결함 자리 | AST |
| `markExitCompletedTx` | 완료 | 오류 → B15 | AST |

Live config binding: 없음.

## State mutations and fallbacks

- exit_states taken · pending · completed, exit_events, trade_outcomes(성과).

## a130 편집 계획 (3판)

B14 안 :1023 한 자리만: 사유가 비어 있지 않으면 `tx.noteOutcomeRefusal(ctx, positionID, reason)`(새 메서드, 새 파일) — RecordFill 이 ctx 에 실은 목록에 쌓기만, 오류 · 반환 없음. 다른 분기 · 반환 무변경.

## Safety conclusion

- Safe edit boundary: 위 「a130 편집 계획」 만. 주문 · 손절 · 체결 반영의 동작 무변경.
- High-risk impact: yes(원장 / 엔진 기동) — 편입 뒤 증가 포지션은 묶음별 편입가로 새로 동결, 감소 · 재편입 · 관측 없음은 거절, 거절 기록 추가.
