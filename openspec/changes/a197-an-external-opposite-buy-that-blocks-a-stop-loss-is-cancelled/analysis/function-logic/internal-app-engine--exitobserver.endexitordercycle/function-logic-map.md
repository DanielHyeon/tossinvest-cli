# Function Logic Map: `ExitObserver.endExitOrderCycle`

- Source: `internal/app/engine/exit_order_log.go` (**편집 전** `235`–`255`, base `b93ed133a`)
- Qualified: `ExitObserver.endExitOrderCycle`
- AST evidence: `ast.json` — branches 4 · returns 1 · defers 1
- Risk scan: `risk-pattern-report.md`
- 작성 시점: **편집 전**(a197 재개 개정 2026-11-25, Manager 결정 3 ① — 외부 치우기 단계를 관측 주기 끝에서 실행). 분기 주장은 `ast.json` 열거에서 왔다.

**역할.** `ObserveOnce` 의 첫 defer(`exitloop.go` `defer@482`)다. 모든 return 경로에서, 그 주기의 `judge`(call `:542`) 뒤에 돈다. a129 의 청산 주문 줄 대기열을 쓰고, a198 의 길 치우기 취소 실패 줄을 비우며(`flushClearCancelFailures`), 거절 래치를 정리하고, 체결 따라가기를 시한(`exitFillReadBudget`) 안에서 돈다. 호출자는 `ObserveOnce` 하나(defer).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` | 관측 주기의 ctx | `ObserveOnce` | 취소되면 하위 읽기가 실패 — 이 함수는 오류를 돌려주지 않음 |
| `o.opts.ExitOrderLog` | nil 가능(전용 로거 없는 조립·시험) | 운영 배선 `cmd/tossctl/engine.go`(a198 design D4 인용) | nil 이면 B1 에서 **전부 건너뜀** |
| `o.exitOrderQueue` · `o.exitOrderRefused(Now)` | 이 주기의 메모리 대기열 | `noteExitOrder` | — |
| 불변식 | 판정 · 제출 · 해제 뒤에만 쓴다(a129 · a198 D4 — 정본 exit-policy 「…줄은 그 주기의 모든 판정·제출·해제 뒤」) | a198 핀 `a198_clear_cancel_log_order_test.go:140` | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `:236` `if o.opts.ExitOrderLog == nil` | 없음 | `return` `:237` — **뒤의 모든 일을 건너뜀** | `TestA129FWithoutTheDedicatedLoggerNothingChanges`, `TestA198WithoutTheDedicatedLoggerTheCancelQueueStaysEmpty` |
| B2 | `:241` `range queue` | `writeExitOrder` 줄 쓰기 | — | `TestA129AConfirmedStopAndItsFillsEachLeaveALine` 등 |
| — | (분기 밖) `flushClearCancelFailures` `:244` | a198 실패 줄 쓰기 | — | `TestA198TheCancelFailureQueueIsFlushedByTheCycleEnd`(구조) |
| B3 | `:246` `range o.exitOrderRefused` | 이번 주기에 거절 없던 포지션의 래치 삭제 | — | `TestA129MTheRefusalLatchDoesNotOutliveItsEpisode` |
| B4 | `:247` `if !o.exitOrderRefusedNow[id]` | `delete` | — | 같음 |
| — | `context.WithTimeout(ctx, exitFillReadBudget)` `:252` · `defer cancel` `:253` · `followExitFills` `:254` | 원장 읽기만 | — | `TestA129BTheFillFollowRunsAfterTheJudgementOfTheSameCycle`, `TestA129CTheFillFollowRunsOnAYieldedCycle` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `o.writeExitOrder` | a129 줄 | 패닉 격리는 그 함수 안(a129 G) | AST `:242` |
| `o.flushClearCancelFailures` | a198 D4 | 줄 하나 단위 recover | AST `:244` |
| `context.WithTimeout` · `o.followExitFills` | 체결 따라가기 | 시한 `exitFillReadBudget` | AST `:252` `:254` |

Live config binding: 없음.

## State mutations and fallbacks

- 메모리만: `exitOrderQueue` 비움, `exitOrderRefused` 정리, a198 대기열 비움. 원장 쓰기 · 브로커 호출 0(편집 전).

## Safety conclusion

- Safe edit boundary (a197 재개 개정 D5): 외부 치우기 단계 호출 **한 줄을 B1 앞**(함수 첫 문장)에 둔다. B1 뒤에 두면 전용 로거가 없는 조립에서 단계가 조용히 돌지 않는다(B1 `return@237`). 기존 줄 · 분기 순서 무변경. a198 핀(`flushClearCancelFailures` 호출 존재 · `noteClearCancelFailure` 의 `Warn` 0)은 이 편집과 무관.
- 귀결: 단계가 B2 의 줄 쓰기보다 **먼저** 돈다 — 줄 쓰기는 판정이 아니고 a198 D4 의 조건(판정 · 제출 · 해제 뒤)은 그대로 성립. 단계 시간만큼 그 주기의 줄 쓰기와 다음 주기 시작(`Run` `Sleep@424`)이 늦어진다(design D4 공시).
- 이 함수는 a163 렌더 해시 · a111 핀 · a166 census 대상이 아니다(rg). 파일 `exit_order_log.go` 는 a087 번들의 파일이나 그 해시는 a198 로 이미 어긋남(a198 design §6).
- High-risk impact: yes — 손절 경로의 관측 주기 안에서 브로커 조회 · 취소가 돌게 된다.
