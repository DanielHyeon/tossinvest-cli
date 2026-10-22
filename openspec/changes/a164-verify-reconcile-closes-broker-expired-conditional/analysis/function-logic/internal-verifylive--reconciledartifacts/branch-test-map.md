# Branch Test Map: `ReconciledArtifacts`

- Source: `internal/verifylive/projection_reconciled.go` (8-31)
- AST branches 6 — 시험 칸은 시험별 coverage 측정(본문 블록 실행)으로 채움

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | range 줄 11: `for _, e := range entries {` | `TestAFullRunProducesAReportWithNoIdempotencyGaps` · `TestProgressPointsAtTheRestartWhenOneIsPending` · `TestReconcileLeavesTheReportAttributesAndVerdictsUnchanged` · `TestReconcileTextLabelsReconciledAbsentInReportAndStatus` 외 4 | 해당 없음 — 편집 전 기준선 | 기준선 통과(측정, 아래 「측정」) |
| B2 | range 줄 12: `for _, a := range e.Artifacts {` | `TestAFullRunProducesAReportWithNoIdempotencyGaps` · `TestProgressPointsAtTheRestartWhenOneIsPending` · `TestReconcileLeavesTheReportAttributesAndVerdictsUnchanged` · `TestReconcileTextLabelsReconciledAbsentInReportAndStatus` 외 2 | 해당 없음 — 편집 전 기준선 | 기준선 통과(측정, 아래 「측정」) |
| B3 | if 줄 15: `if !seen {` | `TestAFullRunProducesAReportWithNoIdempotencyGaps` · `TestProgressPointsAtTheRestartWhenOneIsPending` · `TestReconcileLeavesTheReportAttributesAndVerdictsUnchanged` · `TestReconcileTextLabelsReconciledAbsentInReportAndStatus` 외 2 | 해당 없음 — 편집 전 기준선 | 기준선 통과(측정, 아래 「측정」) |
| B4 | if 줄 18: `if seen && prev.terminal() && !a.terminal() {` | `TestReconciledProjectionIsMonotoneLikeOutstanding` | 해당 없음 — 편집 전 기준선 | 기준선 통과(측정, 아래 「측정」) |
| B5 | range 줄 25: `for _, key := range order {` | `TestAFullRunProducesAReportWithNoIdempotencyGaps` · `TestProgressPointsAtTheRestartWhenOneIsPending` · `TestReconcileLeavesTheReportAttributesAndVerdictsUnchanged` · `TestReconcileTextLabelsReconciledAbsentInReportAndStatus` 외 2 | 해당 없음 — 편집 전 기준선 | 기준선 통과(측정, 아래 「측정」) |
| B6 | if 줄 26: `if a := latest[key]; a.ReconciledAbsent {` | `TestReconcileLeavesTheReportAttributesAndVerdictsUnchanged` · `TestReconcileTextLabelsReconciledAbsentInReportAndStatus` · `TestReconciledProjectionIsMonotoneLikeOutstanding` | 해당 없음 — 편집 전 기준선 | 기준선 통과(측정, 아래 「측정」) |

## 측정

- 기준선: 구현 base `aec9bde67` worktree, `go test -run '^<Test>$' -coverprofile` 을 시험 하나씩(verifylive 의
  `reconcile_*_test.go`·`record_filled_test.go`·`record_replay_test.go`·`report_test.go`, official 의 `reconcile_reads*_test.go`,
  cmd 의 `verify_reconcile*_test.go` 를 `-tags tossos_testseams` 로) 돌려 279 시험 프로필을 얻고, 분기 줄의 본문 `{` 블록 실행 수가
  0 보다 큰 시험만 적었다. 기준선 스위트는 verifylive 634 · official 455 · cmd(Reconcile 계열) 62 전부 통과.
