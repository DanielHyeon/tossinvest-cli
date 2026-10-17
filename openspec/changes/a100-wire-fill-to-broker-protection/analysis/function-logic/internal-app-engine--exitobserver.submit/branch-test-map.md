# Branch Test Map: `ExitObserver.submit`

- Source: `internal/app/engine/exitloop.go`

> Test 열은 그 함수를 지나는 현존 · 통과 시험이다. 「아니오」 분기의 인용은 그 갈래 자체의 증명이 아니다(진입 실측 열이 정본).
> a091 의 새 RED 는 구현 로트의 변이 원장(`analysis/implementation/mutation-ledger.md`)이 잰다 — 이 표는 편집 뒤 `540aebe6` 의 사실이다.

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:1406` `if err != nil {` | 아니오 | `TestAFloorThatCannotBeComputedSellsNothing` | n/a | yes |
| B2 | `:1409` `if isZeroQuantity(submitQuantity) {` | 예 | `TestAZeroFloorSubmitsNothingAndLeavesTheLevelProposable` | n/a | yes |
| B3 | `:1429` `if err != nil {` | 아니오 | `TestA094AnUnacceptedStopIsReleasedAndProposedAgain` | n/a | yes |
| B4 | `:1438` `if err := o.opts.Journal.AttachExitIntent(ctx, m.position.ID, intentID); err != nil {` | 예 | `TestABaselineBreachProposesTheWholePosition` | n/a | yes |
| B5 | `:1443` `if err != nil {` | 아니오 | `TestABaselineBreachProposesTheWholePosition` | n/a | yes |
| B6 | `:1453` `switch {` | — | `TestABaselineBreachProposesTheWholePosition` | n/a | yes |
| B7 | `:1454` `case out.State == journal.StateConfirmed:` | 예 | `TestABaselineBreachProposesTheWholePosition` | n/a | yes |
| B8 | `:1462` `case out.State == journal.StateInDoubt \|\| out.State == journal.StateUnresolvedInDoubt:` | 예 | `TestAnInDoubtSubmissionKeepsTheProposalArmed` | n/a | yes |
| B9 | `:1467` `case out.Reason == execgw.ReasonSymbolInFlight:` | 아니오 | `TestABaselineBreachProposesTheWholePosition` | n/a | yes |
| B10 | `:1470` `case out.AttemptID != "" && out.State != journal.StateNotDispatched && out.State != journal.StateFailedConfirmed:` | 예 | `TestA094AnUnrecordedOutcomeKeepsTheProposalArmed` | n/a | yes |
| B11 | `:1474` `if err == nil {` | 아니오 | `TestA094AnUnrecordedOutcomeKeepsTheProposalArmed` | n/a | yes |
| B12 | `:1479` `default:` | 예 | `TestA094AnUnacceptedStopIsReleasedAndProposedAgain` · `TestARefusedProposalReleasesTheLevelAndAlerts` | n/a | yes |
| B13 | `:1481` `if detail == "" && err != nil {` | 아니오 | `TestARefusedProposalReleasesTheLevelAndAlerts` | n/a | yes |
