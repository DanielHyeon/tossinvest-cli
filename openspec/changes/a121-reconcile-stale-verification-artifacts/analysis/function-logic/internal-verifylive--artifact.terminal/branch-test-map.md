# Branch Test Map: `Artifact.terminal`

- Source: `internal/verifylive/record.go` (575-575)
- AST branches 0 — 행복 경로 한 행

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 분기 없는 단일 식. `Filled` 만 true → 종결(투영에서 빠짐); 실기록 전수에서 `terminal() == Cancelled` | `TestAFilledObjectIsNotOutstanding` (`record_filled_test.go:31`) · `TestTheRealRecordsAreJudgedExactlyAsBefore` (`record_replay_test.go:73`) | 해당 없음 — 편집 전 기준선(task 2.x RED 가 `ReconciledAbsent` 항을 추가) | 기존 스위트 통과(base `de147cc2`) |
