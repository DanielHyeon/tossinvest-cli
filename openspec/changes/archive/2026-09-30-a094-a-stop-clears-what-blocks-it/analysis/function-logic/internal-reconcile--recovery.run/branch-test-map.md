# Branch Test Map: `Recovery.Run`

- Source: `internal/reconcile/recovery.go`

> Test 열은 그 분기를 지나는 시험(현존 · 통과)이다. 「미진입」 분기의 인용은 그 함수를 도는 시험이며 그 갈래 자체의 증명이 아니다(진입 실측 열이 정본).
> RED 는 새 a094 시험에 한해 변이 원장(`analysis/implementation/mutation-ledger.md`)이 잰다.

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:254` `if err != nil {` | 아니오 | `TestRecoveryReleasesTheLatchOnlyWhenItCompletes` | n/a | yes |
| B2 | `:267` `if err != nil {` | 아니오 | `TestRecoveryReleasesTheLatchOnlyWhenItCompletes` | n/a | yes |
| B3 | `:270` `for _, rec := range pending {` | 예 | `TestRecoveryReleasesTheLatchOnlyWhenItCompletes` | n/a | yes |
| B4 | `:271` `if rec.State != journal.StateInDoubt {` | 예 | `TestA094AnAckedPlaceIsConfirmedByItsRecordedNumber` · `TestA094AnUnconfirmedAckedPlaceStaysAndIsNamed` | n/a | yes |
| B5 | `:279` `if rec.State == journal.StateAcked && r.confirmAcked(ctx, rec) {` | 예 | `TestA094AnAckedPlaceIsConfirmedByItsRecordedNumber` | n/a | yes |
| B6 | `:284` `if berr != nil {` | 아니오 | `TestA094AnUnconfirmedAckedPlaceStaysAndIsNamed` · `TestA094AnAckedCancelOrNumberlessPlaceIsOnlyNamed` | n/a | yes |
| B7 | `:292` `if rerr != nil {` | 예 | `TestCrashMidDispatchBecomesInDoubtAndIsResolved` | n/a | yes |
| B8 | `:296` `if settled {` | 예 | `TestCrashMidDispatchBecomesInDoubtAndIsResolved` | n/a | yes |
| B9 | `:301` `if rerr != nil {` | 아니오 | `TestRecoveryReleasesTheLatchOnlyWhenItCompletes` | n/a | yes |
| B10 | `:306` `if res.State == journal.StateUnresolvedInDoubt {` | 아니오 | `TestRecoveryReleasesTheLatchOnlyWhenItCompletes` | n/a | yes |
| B11 | `:318` `if err != nil {` | 예 | `TestRecoveryReleasesTheLatchOnlyWhenItCompletes` | n/a | yes |
| B12 | `:325` `if err != nil {` | 아니오 | `TestRecoveryReleasesTheLatchOnlyWhenItCompletes` | n/a | yes |
| B13 | `:340` `if report.Diff.BlocksEntry() {` | 예 | `TestRecoveryReleasesTheLatchOnlyWhenItCompletes` | n/a | yes |
