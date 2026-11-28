# Branch Test Map: `ExitObserver.endExitOrderCycle`

Source: `internal/app/engine/exit_order_log.go` (235-259) — AST branches 4. RED/GREEN 은 a207 시험의 실제 실행(`analysis/tdd/`)으로만 yes 로 적고, 기존 시험은 이 change 에서
선택 실행으로 통과만 확인함(커버리지 count 미측정).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 로거 없음 | `TestA207AStopRefusedForNCyclesIsSubmittedOnTheFirstClearCycle` | no | yes |
| B2 | 줄 쓰기 | `TestA129JARepeatedRefusalIsWrittenOnceUntilTheOutcomeChanges` | no | no |
| B3 | 래치 순회 | `TestA129JARepeatedRefusalIsWrittenOnceUntilTheOutcomeChanges` | no | no |
| B4 | 래치 풂 | `TestA129JARepeatedRefusalIsWrittenOnceUntilTheOutcomeChanges` | no | no |
