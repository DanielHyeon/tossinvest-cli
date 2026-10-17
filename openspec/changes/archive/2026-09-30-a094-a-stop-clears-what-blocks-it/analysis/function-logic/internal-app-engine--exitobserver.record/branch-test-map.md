# Branch Test Map: `ExitObserver.record`

- Source: `internal/app/engine/exitloop.go`

> Test 열은 그 분기를 지나는 시험(현존 · 통과)이다. 「미진입」 분기의 인용은 그 함수를 도는 시험이며 그 갈래 자체의 증명이 아니다(진입 실측 열이 정본).
> RED 는 새 a094 시험에 한해 변이 원장(`analysis/implementation/mutation-ledger.md`)이 잰다.

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:1228` `if !o.quoteUsable(quote) {` | 아니오 | `TestA094TheParkCheckChangesNoOutcome` | n/a | yes |
| B2 | `:1247` `if judgement.ObservationSource == "" {` | 아니오 | `TestABaselineBreachProposesTheWholePosition` | n/a | yes |
| B3 | `:1248` `if quote.FetchedAt.IsZero() {` | 아니오 | `TestABaselineBreachProposesTheWholePosition` | n/a | yes |
| B4 | `:1250` `} else {` | 아니오 | `TestABaselineBreachProposesTheWholePosition` | n/a | yes |
| B5 | `:1271` `if orderable && (snapshot.CancelPendingFirst \|\| isFullExit(proposal)) {` | 예 | `TestAWorkingEntryIsCancelledBeforeTheLiquidation` | n/a | yes |
| B6 | `:1272` `if m.reJudge && !isProtective(proposal) {` | 예 | `TestCrossingTheFirstTakeProfitKeepsThePositionUnderJudgement` | n/a | yes |
| B7 | `:1294` `} else {` | 예 | `TestAWorkingEntryIsCancelledBeforeTheLiquidation` | n/a | yes |
| B8 | `:1296` `if err != nil {` | 예 | `TestA094AnotherIntentInFlightStopsTheArmReleaseLoop` | n/a | yes |
| B9 | `:1299` `if !cleared.cleared {` | 예 | `TestA094ConsecutiveClearFailuresRaiseAnEarlierAlert` · `TestA094AClosingRecordThatNeverComesHoldsAndAlerts` | n/a | yes |
| B10 | `:1306` `} else {` | 예 | `TestAWorkingEntryIsCancelledBeforeTheLiquidation` | n/a | yes |
| B11 | `:1314` `if orderable {` | 예 | `TestABaselineBreachProposesTheWholePosition` | n/a | yes |
| B12 | `:1316` `if intentID == "" {` | 아니오 | `TestABaselineBreachProposesTheWholePosition` | n/a | yes |
| B13 | `:1328` `if err != nil {` | 예 | `TestAnUnresolvedProposalSuppressesTheNextOne` | n/a | yes |
| B14 | `:1329` `if errors.Is(err, journal.ErrProposalPending) {` | 예 | `TestAnUnresolvedProposalSuppressesTheNextOne` | n/a | yes |
| B15 | `:1335` `if errors.Is(err, journal.ErrExitSnapshotQuarantined) {` | 예 | `TestAnUnresolvedProposalSuppressesTheNextOne` | n/a | yes |
| B16 | `:1348` `if recorded.ArmedProposal == nil \|\| recorded.ArmOutcome != journal.ExitArmArmed {` | 예 | `TestABaselineBreachProposesTheWholePosition` | n/a | yes |
