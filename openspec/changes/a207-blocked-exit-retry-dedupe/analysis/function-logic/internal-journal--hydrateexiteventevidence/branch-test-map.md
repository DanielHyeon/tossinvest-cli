# Branch Test Map: `hydrateExitEventEvidence`

Source: `internal/journal/exit_snapshot.go` (397-426) — AST branches 7. RED/GREEN 은 a207 시험의 실제 실행(`analysis/tdd/`)으로만 yes 로 적고, 기존 시험은 이 change 에서
선택 실행으로 통과만 확인함(커버리지 count 미측정).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | legacy 행 | `TestLegacyEventRequiresEveryV10ColumnToBeNull` | no | no |
| B2 | 압축 행 · 표지만 있는 틀린 모양 | `TestA207ReadersReadACompactRowAsARepeat` · `TestA207AMarkedRowWithTheWrongShapeIsCorrupt` | yes | yes |
| B3 | 표지 없는 부분 증거 | `TestA207ReadersReadACompactRowAsARepeat` | no | yes |
| B4 | saved 손상 | `TestExitSnapshotCorruptionIsPerPositionAndNeverRecomputed` | no | no |
| B5 | recomputed 손상 | 기존(무변경) | no | no |
| B6 | effective 손상 | 기존(무변경) | no | no |
| B7 | 평탄화 불일치 | 기존(무변경) | no | no |
