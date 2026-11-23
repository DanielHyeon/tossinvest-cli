# Function Logic Map: `ExitObserver.endExitOrderCycle`

- Source: `internal/app/engine/exit_order_log.go:235-254`
- AST evidence: `ast.json` (**pre-edit**, HEAD `1ee5ab86a` — source_sha256 `419352e9…`, 분기 4). 편집 뒤 재생성.
- Risk scan: `risk-pattern-report.md` (적중 0)
- 편집 범위(design D4 재결정, Manager 2026-11-24 — gstack r2 G1): a129 대기열 쓰기(B2) 뒤에 한 줄 — `o.flushClearCancelFailures()`(새 파일 a198_clear_cancel_log.go 의 새 함수)로 길 치우기 취소 실패 줄 대기열을 비우며 씀. B1(로거 nil 이면 즉시 반환)·래치 정리(B3·B4)·체결 따라가기 무편집.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `o.opts.ExitOrderLog` | nil(운영 외) · 전용 로거 | 조립(cmd/tossctl/engine.go:680) | nil → 즉시 반환(B1) |
| `o.exitOrderQueue` | 이 주기에 submit 이 넣은 a129 줄 | 관측기 메모리 | — |
| `o.exitOrderRefused`·`exitOrderRefusedNow` | 거절 래치 | 관측기 메모리 | — |
| 불변식 | 줄 쓰기와 체결 기록 읽기는 그 주기의 모든 판정·제출·해제 뒤(정본 exit-policy:689) · 쓰기 패닉이 루프를 바꾸지 않음(writeExitOrder 의 recover) | 정본 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 L236 | 전용 로거 없음 | 없음 | return | TestA129FWithoutTheDedicatedLoggerNothingChanges |
| B2 L241 | a129 대기열 순회 | 줄 쓰기 | — | TestA129AConfirmedStopAndItsFillsEachLeaveALine · TestA129DEveryOtherOutcomeLeavesOneLineAfterItsRelease |
| B3 L245 | 래치 순회 | — | — | TestA129MTheRefusalLatchDoesNotOutliveItsEpisode |
| B4 L246 | 이 주기에 거절 없던 포지션 | 래치 삭제 | — | TestA129MTheRefusalLatchDoesNotOutliveItsEpisode |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `writeExitOrder` | a129 줄 하나 쓰기 | 패닉은 그 쓰기 하나만 recover | AST |
| `context.WithTimeout` · `followExitFills` | 체결 따라가기(원장 읽기, 예산 `exitFillReadBudget`) | 시한 | AST |

## State mutations and fallbacks

- 대기열 비움, 래치 정리. 원장 쓰기·브로커 호출 0.
- 호출 자리: `ObserveOnce` 의 첫 문장 `defer o.endExitOrderCycle(ctx)`(exitloop.go:481) — 모든 return 경로와 패닉 언와인딩에서 돎.
