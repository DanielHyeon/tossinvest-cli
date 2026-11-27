# Branch Test Map: `appendExitEventTx`

Source: `internal/journal/exit_state.go` (792-859) — AST branches 9. RED/GREEN 은 a207 시험의 실제 실행(`analysis/tdd/`)으로만 yes 로 적고, 기존 시험은 이 change 에서
선택 실행으로 통과만 확인함(커버리지 count 미측정).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 세대 조회 실패 | 기존(무변경) | no | no |
| B2 | 평가 있는 판정 행 | `TestA207AStopRefusedForNCyclesIsSubmittedOnTheFirstClearCycle` | yes | yes |
| B3 | saved 없는 첫 평가 | `TestExitSnapshotPersistsWholeRecoveryCandidates` | no | no |
| B4 | saved encode 실패 | 도달 시험 없음 | no | no |
| B5 | recomputed encode 실패 | 도달 시험 없음 | no | no |
| B6 | effective encode 실패 | 도달 시험 없음 | no | no |
| B7 | 무장 행만 술어를 물음 / 판정 전용 · legacy 행은 안 물음 | `TestA207ARepeatedRefusedArmIsWrittenCompact` · `TestA207OnlyArmRowsAskTheRepeatPredicate` | yes | yes |
| B8 | 반복이면 압축, 조회 오류면 전체 | `TestA207ARepeatedRefusedArmIsWrittenCompact` · `TestA207ALookupFaultWritesTheFullRowAndStillArms` | yes | yes |
| B9 | INSERT 실패 | 기존(무변경) | no | no |
