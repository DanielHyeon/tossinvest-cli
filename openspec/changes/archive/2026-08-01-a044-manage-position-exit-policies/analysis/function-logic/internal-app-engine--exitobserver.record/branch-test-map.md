# Branch Test Map: `ExitObserver.record`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | missing fetched-at uses observation fallback | `exitloop observation suite` | yes | yes |
| B2 | fresh quote preserves broker timestamp | `TestASuccessfulObservationStampsThePriceFreshness` | yes | yes |
| B3 | full exit or cancel-first clears conflicting orders | `TestABreachDisplacesAnOutstandingTakeProfit` | yes | yes |
| B4 | clear failure returns before journal/order | `exitloop conflict suite` | yes | yes |
| B5 | uncleared conflict suppresses order | `TestAnUncancellableEntryWithholdsTheLiquidationAndAlertsPastTheBound` | yes | yes |
| B6 | cleared conflict permits proposal | `TestABreachDisplacesAnOutstandingTakeProfit` | yes | yes |
| B7 | orderable snapshot builds proposal | `TestABaselineBreachProposesTheWholePosition` | yes | yes |
| B8 | intent-id derivation failure refuses | `exitloop identity suite` | yes | yes |
| B9 | journal stale/error returns before submit | `TestLateOldGenerationJudgementIsQuarantined` | yes | yes |
| B10 | pending proposal is surfaced without duplicate | `TestAnUnresolvedProposalSuppressesTheNextOne` | yes | yes |
| B11 | only durably armed proposal reaches submit | `TestAnArmedProposalIsNotProposedTwiceAfterARestart` | yes | yes |
