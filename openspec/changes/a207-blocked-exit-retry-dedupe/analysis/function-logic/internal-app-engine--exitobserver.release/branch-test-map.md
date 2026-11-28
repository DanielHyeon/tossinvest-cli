# Branch Test Map: `ExitObserver.release`

Source: `internal/app/engine/exitloop.go` (1651-1661) — AST branches 2. RED/GREEN 은 a207 시험의 실제 실행(`analysis/tdd/`)으로만 yes 로 적고, 기존 시험은 이 change 에서
선택 실행으로 통과만 확인함(커버리지 count 미측정).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 원장 해제 실패 | `TestA129OAFailedReleaseStillSurfaces` | no | no |
| B2 | 원장이 풂 / 원장이 거절 | `TestA207AStopRefusedForNCyclesIsSubmittedOnTheFirstClearCycle` · `TestA207AReleaseTheLedgerRefusesIsNotCounted` | no | yes |
