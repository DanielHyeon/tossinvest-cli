# Branch Test Map: `checkReconcileAbsence`

- Source: `internal/verifylive/reconcile_check.go` (371-396)
- AST branches 7 — 시험 칸은 시험별 coverage 측정(본문 블록 실행)으로 채움

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if 줄 372: `if len(r.condOpen) > 0 {` | `TestReconcileRefusesALiveSuccessorInTheOpenConditionalGroup` | 해당 없음 — 편집 전 기준선 | 기준선 통과(측정, 아래 「측정」) |
| B2 | if 줄 376: `if len(r.plainOpen) > 0 {` | `TestReconcileRefusesAnyOpenPlainOrderOnTheSymbol` | 해당 없음 — 편집 전 기준선 | 기준선 통과(측정, 아래 「측정」) |
| B3 | range 줄 380: `for _, row := range r.closed {` | `TestReconcileAcceptsAnExpiredRowOfAnotherIdentifierOnceTheAllowlistHasIt` · `TestReconcileRefusesAnOpenFamilyClosedStatusEvenIfTheAllowlistListsIt` · `TestReconcileRefusesAnotherIdentifierThatFiredInClosed` · `TestReconcileRefusesClosedStatusOutsideTheAllowlist` 외 2 | 해당 없음 — 편집 전 기준선 | 기준선 통과(측정, 아래 「측정」) |
| B4 | if 줄 381: `if row.ID != target.ID {` | `TestReconcileAcceptsAnExpiredRowOfAnotherIdentifierOnceTheAllowlistHasIt` · `TestReconcileRefusesAnOpenFamilyClosedStatusEvenIfTheAllowlistListsIt` · `TestReconcileRefusesAnotherIdentifierThatFiredInClosed` · `TestReconcileRefusesClosedStatusOutsideTheAllowlist` | 해당 없음 — 편집 전 기준선 | 기준선 통과(측정, 아래 「측정」) |
| B5 | if 줄 384: `if row.Status == "EXPIRED" {` | `TestReconcileRefusesTheTargetExpiredInClosedWithThePermanentMessage` | 해당 없음 — 편집 전 기준선 | 기준선 통과(측정, 아래 「측정」) |
| B6 | range 줄 390: `for _, row := range r.closed {` | `TestReconcileAcceptsAnExpiredRowOfAnotherIdentifierOnceTheAllowlistHasIt` · `TestReconcileRefusesAnOpenFamilyClosedStatusEvenIfTheAllowlistListsIt` · `TestReconcileRefusesAnotherIdentifierThatFiredInClosed` · `TestReconcileRefusesClosedStatusOutsideTheAllowlist` | 해당 없음 — 편집 전 기준선 | 기준선 통과(측정, 아래 「측정」) |
| B7 | if 줄 391: `if row.TriggeredOrderID != "" \|\| row.Status == "COMPLETED" {` | `TestReconcileRefusesAnotherIdentifierThatFiredInClosed` | 해당 없음 — 편집 전 기준선 | 기준선 통과(측정, 아래 「측정」) |

## 측정

- 기준선: 구현 base `aec9bde67` worktree, `go test -run '^<Test>$' -coverprofile` 을 시험 하나씩(verifylive 의
  `reconcile_*_test.go`·`record_filled_test.go`·`record_replay_test.go`·`report_test.go`, official 의 `reconcile_reads*_test.go`,
  cmd 의 `verify_reconcile*_test.go` 를 `-tags tossos_testseams` 로) 돌려 279 시험 프로필을 얻고, 분기 줄의 본문 `{` 블록 실행 수가
  0 보다 큰 시험만 적었다. 기준선 스위트는 verifylive 634 · official 455 · cmd(Reconcile 계열) 62 전부 통과.
