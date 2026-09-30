# Branch Test Map: `ExitObserver.submit`

- Source: `internal/app/engine/exitloop.go`

> Test 열은 그 분기를 지나는 시험(현존 · 통과)이다. 「미진입」 분기의 인용은 그 함수를 도는 시험이며 그 갈래 자체의 증명이 아니다(진입 실측 열이 정본).
> RED 는 새 a094 시험에 한해 변이 원장(`analysis/implementation/mutation-ledger.md`)이 잰다.

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:1394` `if err != nil {` | 아니오 | `TestAFloorThatCannotBeComputedSellsNothing` | n/a | yes |
| B2 | `:1397` `if isZeroQuantity(submitQuantity) {` | 예 | `TestAZeroFloorSubmitsNothingAndLeavesTheLevelProposable` | n/a | yes |
| B3 | `:1417` `if err != nil {` | 아니오 | `TestA094AnUnacceptedStopIsReleasedAndProposedAgain` | n/a | yes |
| B4 | `:1426` `if err := o.opts.Journal.AttachExitIntent(ctx, m.position.ID, intentID); err != nil {` | 예 | `TestABaselineBreachProposesTheWholePosition` | n/a | yes |
| B5 | `:1431` `if err != nil {` | 아니오 | `TestABaselineBreachProposesTheWholePosition` | n/a | yes |
| B6 | `:1441` `switch {` | — | `TestABaselineBreachProposesTheWholePosition` | n/a | yes |
| B7 | `:1442` `case out.State == journal.StateConfirmed:` | 예 | `TestABaselineBreachProposesTheWholePosition` | n/a | yes |
| B8 | `:1450` `case out.State == journal.StateInDoubt \|\| out.State == journal.StateUnresolvedInDoubt:` | 예 | `TestAnInDoubtSubmissionKeepsTheProposalArmed` | n/a | yes |
| B9 | `:1455` `case out.Reason == execgw.ReasonSymbolInFlight:` | 아니오 | `TestABaselineBreachProposesTheWholePosition` | n/a | yes |
| B10 | `:1458` `case out.AttemptID != "" && out.State != journal.StateNotDispatched && out.State != journal.StateFailedConfirmed:` | 예 | `TestA094AnUnrecordedOutcomeKeepsTheProposalArmed` | n/a | yes |
| B11 | `:1462` `if err == nil {` | 아니오 | `TestA094AnUnrecordedOutcomeKeepsTheProposalArmed` | n/a | yes |
| B12 | `:1467` `default:` | 예 | `TestA094AnUnacceptedStopIsReleasedAndProposedAgain` · `TestARefusedProposalReleasesTheLevelAndAlerts` | n/a | yes |
| B13 | `:1469` `if detail == "" && err != nil {` | 아니오 | `TestARefusedProposalReleasesTheLevelAndAlerts` | n/a | yes |
