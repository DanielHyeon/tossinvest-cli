# Branch Test Map: `ExitObserver.record`

- Source: `internal/app/engine/exitloop.go`

> Test 열은 그 분기를 지나는 시험(현존 · 통과)이다. 「미진입」 분기의 인용은 그 함수를 도는 시험이며 그 갈래 자체의 증명이 아니다(진입 실측 열이 정본).
> RED 는 새 a094 시험에 한해 변이 원장(`analysis/implementation/mutation-ledger.md`)이 잰다.

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:1224` `if !o.quoteUsable(quote) {` | 아니오 | `TestA094TheParkCheckChangesNoOutcome` | n/a | yes |
| B2 | `:1243` `if judgement.ObservationSource == "" {` | 아니오 | `TestABaselineBreachProposesTheWholePosition` | n/a | yes |
| B3 | `:1244` `if quote.FetchedAt.IsZero() {` | 아니오 | `TestABaselineBreachProposesTheWholePosition` | n/a | yes |
| B4 | `:1246` `} else {` | 아니오 | `TestABaselineBreachProposesTheWholePosition` | n/a | yes |
| B5 | `:1267` `if orderable && (snapshot.CancelPendingFirst \|\| isFullExit(proposal)) {` | 예 | `TestAWorkingEntryIsCancelledBeforeTheLiquidation` | n/a | yes |
| B6 | `:1268` `if m.reJudge && !isProtective(proposal) {` | 예 | `TestCrossingTheFirstTakeProfitKeepsThePositionUnderJudgement` | n/a | yes |
| B7 | `:1290` `} else {` | 예 | `TestAWorkingEntryIsCancelledBeforeTheLiquidation` | n/a | yes |
| B8 | `:1292` `if err != nil {` | 예 | `TestA094AnotherIntentInFlightStopsTheArmReleaseLoop` | n/a | yes |
| B9 | `:1295` `if !cleared.cleared {` | 예 | `TestA094ConsecutiveClearFailuresRaiseAnEarlierAlert` · `TestA094AClosingRecordThatNeverComesHoldsAndAlerts` | n/a | yes |
| B10 | `:1302` `} else {` | 예 | `TestAWorkingEntryIsCancelledBeforeTheLiquidation` | n/a | yes |
| B11 | `:1310` `if orderable {` | 예 | `TestABaselineBreachProposesTheWholePosition` | n/a | yes |
| B12 | `:1312` `if intentID == "" {` | 아니오 | `TestABaselineBreachProposesTheWholePosition` | n/a | yes |
| B13 | `:1324` `if err != nil {` | 예 | `TestAnUnresolvedProposalSuppressesTheNextOne` | n/a | yes |
| B14 | `:1325` `if errors.Is(err, journal.ErrProposalPending) {` | 예 | `TestAnUnresolvedProposalSuppressesTheNextOne` | n/a | yes |
| B15 | `:1331` `if errors.Is(err, journal.ErrExitSnapshotQuarantined) {` | 예 | `TestAnUnresolvedProposalSuppressesTheNextOne` | n/a | yes |
| B16 | `:1344` `if recorded.ArmedProposal == nil \|\| recorded.ArmOutcome != journal.ExitArmArmed {` | 예 | `TestABaselineBreachProposesTheWholePosition` | n/a | yes |
