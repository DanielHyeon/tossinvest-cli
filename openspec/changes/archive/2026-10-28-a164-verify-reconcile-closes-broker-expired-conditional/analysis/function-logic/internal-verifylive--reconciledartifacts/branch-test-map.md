# Branch Test Map: `ReconciledArtifacts`

- Source: `internal/verifylive/projection_reconciled.go` (9-32)
- AST branches 6 — 시험 칸은 시험별 coverage 측정(본문 블록 실행)으로 채움

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | range 줄 12: `for _, e := range entries {` | `TestAFullRunProducesAReportWithNoIdempotencyGaps` · `TestExpiryReconcileEndsTheArtifactForEveryOutstandingReader` · `TestExpiryReconcileTextLabelsReconciledExpired` · `TestProgressPointsAtTheRestartWhenOneIsPending` 외 7 | 해당 없음 — a164 GREEN 재추출(리뷰 수리 f4e71ad1d 포함) — 새 갈래 RED 는 analysis/red-green/red-2026-10-22.md | 기준선 통과(측정, 아래 「측정」) |
| B2 | range 줄 13: `for _, a := range e.Artifacts {` | `TestAFullRunProducesAReportWithNoIdempotencyGaps` · `TestExpiryReconcileEndsTheArtifactForEveryOutstandingReader` · `TestExpiryReconcileTextLabelsReconciledExpired` · `TestProgressPointsAtTheRestartWhenOneIsPending` 외 5 | 해당 없음 — a164 GREEN 재추출(리뷰 수리 f4e71ad1d 포함) — 새 갈래 RED 는 analysis/red-green/red-2026-10-22.md | 기준선 통과(측정, 아래 「측정」) |
| B3 | if 줄 16: `if !seen {` | `TestAFullRunProducesAReportWithNoIdempotencyGaps` · `TestExpiryReconcileEndsTheArtifactForEveryOutstandingReader` · `TestExpiryReconcileTextLabelsReconciledExpired` · `TestProgressPointsAtTheRestartWhenOneIsPending` 외 5 | 해당 없음 — a164 GREEN 재추출(리뷰 수리 f4e71ad1d 포함) — 새 갈래 RED 는 analysis/red-green/red-2026-10-22.md | 기준선 통과(측정, 아래 「측정」) |
| B4 | if 줄 19: `if seen && prev.terminal() && !a.terminal() {` | `TestReconciledProjectionIsMonotoneLikeOutstanding` | 해당 없음 — a164 GREEN 재추출(리뷰 수리 f4e71ad1d 포함) — 새 갈래 RED 는 analysis/red-green/red-2026-10-22.md | 기준선 통과(측정, 아래 「측정」) |
| B5 | range 줄 26: `for _, key := range order {` | `TestAFullRunProducesAReportWithNoIdempotencyGaps` · `TestExpiryReconcileEndsTheArtifactForEveryOutstandingReader` · `TestExpiryReconcileTextLabelsReconciledExpired` · `TestProgressPointsAtTheRestartWhenOneIsPending` 외 5 | 해당 없음 — a164 GREEN 재추출(리뷰 수리 f4e71ad1d 포함) — 새 갈래 RED 는 analysis/red-green/red-2026-10-22.md | 기준선 통과(측정, 아래 「측정」) |
| B6 | if 줄 27: `if a := latest[key]; a.ReconciledAbsent \|\| a.ReconciledExpired {` | `TestExpiryReconcileEndsTheArtifactForEveryOutstandingReader` · `TestExpiryReconcileTextLabelsReconciledExpired` · `TestReconcileAbsenceKeepsItsA121Wording` · `TestReconcileLeavesTheReportAttributesAndVerdictsUnchanged` 외 2 | 해당 없음 — a164 GREEN 재추출(리뷰 수리 f4e71ad1d 포함) — 새 갈래 RED 는 analysis/red-green/red-2026-10-22.md | 기준선 통과(측정, 아래 「측정」) |

## 측정

- 기준선: 구현 base `aec9bde67` worktree, `go test -run '^<Test>$' -coverprofile` 을 시험 하나씩(verifylive 의
  `reconcile_*_test.go`·`record_filled_test.go`·`record_replay_test.go`·`report_test.go`, official 의 `reconcile_reads*_test.go`,
  cmd 의 `verify_reconcile*_test.go` 를 `-tags tossos_testseams` 로) 돌려 279 시험 프로필을 얻고, 분기 줄의 본문 `{` 블록 실행 수가
  0 보다 큰 시험만 적었다. 기준선 스위트는 verifylive 634 · official 455 · cmd(Reconcile 계열) 62 전부 통과.

## a164 GREEN 재추출

- 편집 전(구현 base `aec9bde67`) 분기 열 길이 6 에서 편집 뒤 6 로. `source_sha256` 55b48e9c512e… → e7ff440063e7….
- 옛/새 분기 열(kind + 원문 줄)을 difflib 으로 정렬한 결과: equal 옛 B1..B5 = 새 B1..B5 · replace 옛 B6..B6(1) → 새 B6..B6(1).
- 위 표의 B 번호는 **새** AST 기준이고, 시험 칸은 GREEN 코드에서 시험별 프로필을 다시 떠 측정한 값이다(손 재번호 없음).
