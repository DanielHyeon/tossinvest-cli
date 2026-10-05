# Branch Test Map: `Artifact.terminal`

- Source: `internal/verifylive/record.go` (664-664)
- AST branches 0 — 행복 경로 한 행

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 분기 없는 단일 식. `Filled` 만 true → 종결(투영에서 빠짐); 실기록 전수에서 `terminal() == Cancelled` | `TestAFilledObjectIsNotOutstanding` (`record_filled_test.go:31`) · `TestTheRealRecordsAreJudgedExactlyAsBefore` (`record_replay_test.go:73`) | 해당 없음 — 편집 전 기준선(task 2.x RED 가 `ReconciledAbsent` 항을 추가) | 기존 스위트 통과(base `de147cc2`) |

## GREEN 로트 재추출 (a121 tasks 3.1·3.2, worktree a99a9059 + GREEN 편집)

- 편집: `|| a.ReconciledAbsent` 셋째 종결 추가(분기 0 유지). GREEN 관측: TestReconcileOnlyRemovesItsExactArtifactFromCleanupPlanning·TestReconcilePinsRedoSetBeforeAndAfter·TestReconcileLeavesTheReportAttributesAndVerdictsUnchanged PASS; 변이 V-terminal-third 원장 참조.
- AST 재추출: `go run ./tools/logic-map` — 분기 0→0, 편집 전후 분기 열을 difflib 로 정렬한 결과 equal 뿐(재번호 없음), 앵커는 줄 이동만 반영.
