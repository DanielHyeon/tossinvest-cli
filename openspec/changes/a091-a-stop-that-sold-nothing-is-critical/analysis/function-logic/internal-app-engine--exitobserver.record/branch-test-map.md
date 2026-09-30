# Branch Test Map: `ExitObserver.record`

- Source: `internal/app/engine/exitloop.go`

> Test 열은 그 함수를 지나는 현존 · 통과 시험이다. 「아니오」 분기의 인용은 그 갈래 자체의 증명이 아니다(진입 실측 열이 정본).
> a091 의 새 RED 는 구현 로트의 변이 원장(`analysis/implementation/mutation-ledger.md`)이 잰다 — 이 표는 편집 뒤 `3ec1efd2` 의 사실이다.

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:1232` `if !o.quoteUsable(quote) {` | 아니오 | `TestA094TheParkCheckChangesNoOutcome` | n/a | yes |
| B2 | `:1251` `if judgement.ObservationSource == "" {` | 아니오 | `TestABaselineBreachProposesTheWholePosition` | n/a | yes |
| B3 | `:1252` `if quote.FetchedAt.IsZero() {` | 아니오 | `TestABaselineBreachProposesTheWholePosition` | n/a | yes |
| B4 | `:1254` `} else {` | 아니오 | `TestABaselineBreachProposesTheWholePosition` | n/a | yes |
| B5 | `:1275` `if orderable && (snapshot.CancelPendingFirst \|\| isFullExit(proposal)) {` | 예 | `TestAWorkingEntryIsCancelledBeforeTheLiquidation` | n/a | yes |
| B6 | `:1276` `if m.reJudge && !isProtective(proposal) {` | 예 | `TestCrossingTheFirstTakeProfitKeepsThePositionUnderJudgement` | n/a | yes |
| B7 | `:1298` `} else {` | 예 | `TestAWorkingEntryIsCancelledBeforeTheLiquidation` | n/a | yes |
| B8 | `:1300` `if err != nil {` | 예 | `TestA094AnotherIntentInFlightStopsTheArmReleaseLoop` | n/a | yes |
| B9 | `:1303` `if !cleared.cleared {` | 예 | `TestA094ConsecutiveClearFailuresRaiseAnEarlierAlert` | n/a | yes |
| B10 | `:1310` `} else {` | 예 | `TestAWorkingEntryIsCancelledBeforeTheLiquidation` | n/a | yes |
| B11 | `:1318` `if orderable {` | 예 | `TestABaselineBreachProposesTheWholePosition` | n/a | yes |
| B12 | `:1320` `if intentID == "" {` | 아니오 | `TestABaselineBreachProposesTheWholePosition` | n/a | yes |
| B13 | `:1332` `if err != nil {` | 예 | `TestAnUnresolvedProposalSuppressesTheNextOne` | n/a | yes |
| B14 | `:1333` `if errors.Is(err, journal.ErrProposalPending) {` | 예 | `TestAnUnresolvedProposalSuppressesTheNextOne` | n/a | yes |
| B15 | `:1339` `if errors.Is(err, journal.ErrExitSnapshotQuarantined) {` | 예 | `TestAnUnresolvedProposalSuppressesTheNextOne` | n/a | yes |
| B16 | `:1352` `if recorded.ArmedProposal == nil \|\| recorded.ArmOutcome != journal.ExitArmArmed {` | 예 | `TestABaselineBreachProposesTheWholePosition` | n/a | yes |
