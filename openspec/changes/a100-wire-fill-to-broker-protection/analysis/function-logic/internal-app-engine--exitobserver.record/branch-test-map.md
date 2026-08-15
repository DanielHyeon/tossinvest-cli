# Branch Test Map: `ExitObserver.record`

Source: `internal/app/engine/exitloop.go` (1177-1303).

| Branch | Scenario | Test |
|---|---|---|
| B1 | invalid quote no-op | `TestA111InvalidQuoteNeverRefreshesAnEvaluatedSnapshot` |
| B2 | missing source fallback | `TestA111FallbackSequenceRecoveryIsLazyAndPriceEvidenceUsesTheGateDuration` |
| B3 | missing fetched time cycle fallback | `TestStableObservationIDReusesOneFallbackWithinCycle` |
| B4 | fetched time retained | `TestA111QuoteEvidenceUsesOnePostBatchClockAndNeverFallsBackFromBadOfficialTime` |
| B5 | clear before arm | `TestAWorkingEntryIsCancelledBeforeTheLiquidation` |
| B6 | rejudge take-profit withheld | `TestA111SupersededRejudgementStillReleasesWithoutNonprotectiveOrderSideEffects` |
| B7 | normal clear path | `TestAWorkingEntryIsCancelledBeforeTheLiquidation` |
| B8 | clearing error | `TestAnUncancellableEntryWithholdsTheLiquidationAndAlertsPastTheBound` |
| B9 | uncleared delay | `TestAnUncancellableEntryWithholdsTheLiquidationAndAlertsPastTheBound` |
| B10 | delay clears | `TestAWorkingEntryIsCancelledBeforeTheLiquidation` |
| B11 | proposal constructed | `TestABaselineBreachProposesTheWholePosition` |
| B12 | fresh intent ID | `TestTheWholeExitPathEndToEnd` |
| B13 | record error | `TestAnUnresolvedProposalSuppressesTheNextOne` |
| B14 | pending no-op | `TestAnUnresolvedProposalSuppressesTheNextOne` |
| B15 | quarantine announcement | `TestUnknownLegacyPolicyIdentityIsDurablyGenerationQuarantined` |
| B16 | unarmed no submit | `TestARefusedProposalReleasesTheLevelAndAlerts` |
