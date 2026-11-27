# Function Logic Map: `ExitObserver.endExitOrderCycle`

- Source: `internal/app/engine/exit_order_log.go` (235-259)
- Qualified: `ExitObserver.endExitOrderCycle` · revision: current(편집 뒤, worktree `~/tossos-work/a207` 브랜치 `feat/a207-blocked-exit-retry-dedupe`)
- AST evidence: `ast.json` (source_sha256 4de225c5bc26…) — AST branches 4 · returns 1
- 편집 전 증거: `analysis/ast-evidence/engine--ExitObserver.endExitOrderCycle.ast.json`(235-257, 4 갈래, sha 6bedc542611a)
- Risk scan: `risk-pattern-report.md` (파일 단위)

**역할.** ObserveOnce 의 defer — 주기 끝에 a197 단계 · a129 줄 · 체결 따라가기를 돎. a207 은 a197 단계 바로 뒤에 해소 알림 넘기기 호출 하나를 더함(분기 추가 0).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` | 관측 주기 ctx | ObserveOnce | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `:240` 전용 로거 없음 | — | `:241` return | `TestA207AStopRefusedForNCyclesIsSubmittedOnTheFirstClearCycle`(로거 없는 harness — 해소 알림은 그 앞에서 이미 넘겨짐) |
| B2 | `:245` 이 주기의 줄 쓰기 | writeExitOrder | — | `TestA129JARepeatedRefusalIsWrittenOnceUntilTheOutcomeChanges` |
| B3 | `:250` 래치 순회 | — | — | `TestA129JARepeatedRefusalIsWrittenOnceUntilTheOutcomeChanges` |
| B4 | `:251` 이 주기에 거절 없던 포지션 | 래치 풂 | — | `TestA129JARepeatedRefusalIsWrittenOnceUntilTheOutcomeChanges` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `o.runExternalBuyStage` | a197 외부 매수 치우기 | 패닉 회수 | AST `:237` |
| `o.flushUnblockNotices` | **a207** 해소 알림 넘기기 | 알림 부품 오류는 o.alert 가 로그로 남김 | AST `:239` |
| `o.writeExitOrder` · `o.flushClearCancelFailures` · `o.followExitFills` | a129 · a198 | — | AST |

## State mutations and fallbacks

- 메모리: 해소 대기열을 비움(a207). 알림: 일반 등급 사건(기록 전용 알림기에서는 relay Offer).

## Safety conclusion

- Safe edit boundary: B1 앞(로거 유무와 무관하게) a197 단계 뒤에 호출 하나 — 모든 판정 · 제출 뒤.
- High-risk impact: 주기 끝 시간이 다음 주기 시작을 그만큼 늦춤(design §5.4) — 해소 하나에 로그 한 줄 + Offer 하나.
