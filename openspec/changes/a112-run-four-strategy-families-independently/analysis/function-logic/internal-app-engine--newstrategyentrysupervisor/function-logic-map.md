# Function Logic Map: `NewStrategyEntrySupervisor`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- Current-base source SHA-256: `627c647d087032586c4b63ca315a30fd9fad6b51af329fa4e8bf4fecd7104e08`
- Signature: `NewStrategyEntrySupervisor(params=1, results=2)`
- Source range: `545:1`–`622:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `582:3, 589:3, 592:3, 600:3, 606:4, 609:4, 612:4, 616:4, 620:4, 626:4, 629:4, 632:4, 646:4, 650:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if| 578:2 | arm entered 46x (engine tagged suite); arm entered 46x (engine untagged suite); `TestARefreshOnlyWorkerSwallowsACentralIntegrityErrorToo`, `TestAnEffectiveMarketFaultLeavesItsPeerAndTheSupervisorAlone`, `TestBrokenSupervisorBookkeepingTakesTheSafetyLoopsDownWithIt`, `TestCentralIntegrityFailureEscapesOuterLoopAndDrainsSafety`, `TestContextIgnoringCycleWatchdogLatchesOnceAndLateResultHasNoAction`, `TestEveryWorkerCanHandOffItsFaultWithoutAnybodyDraining`, `TestExpiredAuthorityLatchesBeforeEvaluation`, `TestMarketFailureEmitsExactIrreversibleFaultAndKeepsPeerSafetyAlive`, `TestMarketPanicIsContainedAndCannotRecoverMemoryAuthority`, `TestMarketRestartAttemptAndDeadlineSaturateWithoutOverwritingFirstTypedRefusal`, `TestPairedMarketAbnormalReturnSchedulesOnlyLocalBoundedRestartAndKeepsEverySafetyLoopAlive`, `TestPairedMarketRestartHonorsPublishedAbsoluteDeadlineAfterHandoffRace`, `TestStrategyEntrySupervisorDefaultsPairedMarketsDormant`, `TestStrategyEntrySupervisorPollerKeepsCompletePeerRunningWhenOtherMarketDormant`, `TestStrategyEntrySupervisorPollsDormantRefreshWorkerWithoutOpeningPublicTrigger`, `TestStrategyEntrySupervisorPollsKRAndUSImmediatelyInTheSameWave`, `TestStrategyEntrySupervisorRejectsUnsafeProductionPollIntervals`, `TestStrategyEntrySupervisorStartsKRAndUSCyclesConcurrently`, `TestStrategyEntrySupervisorZeroPollIntervalKeepsExplicitTriggerSemantics`, `TestStrategySupervisorRejectsInvalidAssemblies`, `TestTheFaultStreamHoldsOneSlotForEveryWorkerThatCanLatch`, `TestTheFourEscalationsThatStopTheEngineAreExactlyTheSupervisorsOwnBrokenBookkeeping`, `TestTheMarketThatLeadsAWaveAlwaysPublishesIt`, `TestTheOnlyWorkerProductionActuallyRunsSwallowsEveryCycleError`, `TestUSFXReadFailureDoesNotCancelKRIdentityOrSafetyBudgets` |
| B2 | if| 581:2 | arm entered 1x (engine tagged suite); arm entered 1x (engine untagged suite); `TestStrategySupervisorRejectsInvalidAssemblies` |
| B3 | if| 585:2 | arm entered 34x (engine tagged suite); arm entered 34x (engine untagged suite); `TestCentralIntegrityFailureEscapesOuterLoopAndDrainsSafety`, `TestExpiredAuthorityLatchesBeforeEvaluation`, `TestMarketFailureEmitsExactIrreversibleFaultAndKeepsPeerSafetyAlive`, `TestMarketPanicIsContainedAndCannotRecoverMemoryAuthority`, `TestMarketQueueSaturationDoesNotConsumePeerQueue`, `TestMarketRestartAttemptAndDeadlineSaturateWithoutOverwritingFirstTypedRefusal`, `TestPairedMarketAbnormalReturnSchedulesOnlyLocalBoundedRestartAndKeepsEverySafetyLoopAlive`, `TestPairedMarketRestartHonorsPublishedAbsoluteDeadlineAfterHandoffRace`, `TestShutdownAndTriggerShareBarrierAndDrainBothQueues`, `TestStrategyEntrySupervisorDefaultsPairedMarketsDormant`, `TestStrategyEntrySupervisorPollerKeepsCompletePeerRunningWhenOtherMarketDormant`, `TestStrategyEntrySupervisorPollsDormantRefreshWorkerWithoutOpeningPublicTrigger`, `TestStrategyEntrySupervisorPollsKRAndUSImmediatelyInTheSameWave`, `TestStrategyEntrySupervisorRejectsUnsafeProductionPollIntervals`, `TestStrategyEntrySupervisorStartsKRAndUSCyclesConcurrently`, `TestStrategyEntrySupervisorZeroPollIntervalKeepsExplicitTriggerSemantics`, `TestStrategySupervisorRejectsInvalidAssemblies`, `TestUSFXReadFailureDoesNotCancelKRIdentityOrSafetyBudgets` |
| B4 | if| 588:2 | arm entered 1x (engine tagged suite); arm entered 1x (engine untagged suite); `TestStrategySupervisorRejectsInvalidAssemblies` |
| B5 | if| 591:2 | arm entered 1x (engine tagged suite); arm entered 1x (engine untagged suite); `TestStrategySupervisorRejectsInvalidAssemblies` |
| B6 | if| 595:2 | arm entered 23x (engine tagged suite); arm entered 23x (engine untagged suite); `TestCentralIntegrityFailureEscapesOuterLoopAndDrainsSafety`, `TestMarketFailureEmitsExactIrreversibleFaultAndKeepsPeerSafetyAlive`, `TestMarketPanicIsContainedAndCannotRecoverMemoryAuthority`, `TestMarketQueueSaturationDoesNotConsumePeerQueue`, `TestShutdownAndTriggerShareBarrierAndDrainBothQueues`, `TestStrategyEntrySupervisorDefaultsPairedMarketsDormant`, `TestStrategyEntrySupervisorPollerKeepsCompletePeerRunningWhenOtherMarketDormant`, `TestStrategyEntrySupervisorPollsDormantRefreshWorkerWithoutOpeningPublicTrigger`, `TestStrategyEntrySupervisorPollsKRAndUSImmediatelyInTheSameWave`, `TestStrategyEntrySupervisorRejectsUnsafeProductionPollIntervals`, `TestStrategyEntrySupervisorStartsKRAndUSCyclesConcurrently`, `TestStrategyEntrySupervisorZeroPollIntervalKeepsExplicitTriggerSemantics`, `TestStrategySupervisorRejectsInvalidAssemblies`, `TestUSFXReadFailureDoesNotCancelKRIdentityOrSafetyBudgets` |
| B7 | if| 599:2 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |
| B8 | range| 604:2 | arm entered 84x (engine tagged suite); arm entered 84x (engine untagged suite); `TestALatchedMarketSkipsTheTriggersAlreadySittingInItsQueue`, `TestARefreshOnlyWorkerSwallowsACentralIntegrityErrorToo`, `TestAnEffectiveMarketFaultLeavesItsPeerAndTheSupervisorAlone`, `TestBrokenSupervisorBookkeepingTakesTheSafetyLoopsDownWithIt`, `TestCentralIntegrityFailureEscapesOuterLoopAndDrainsSafety`, `TestContextIgnoringCycleWatchdogLatchesOnceAndLateResultHasNoAction`, `TestEveryWorkerCanHandOffItsFaultWithoutAnybodyDraining`, `TestExpiredAuthorityLatchesBeforeEvaluation`, `TestMarketFailureEmitsExactIrreversibleFaultAndKeepsPeerSafetyAlive`, `TestMarketPanicIsContainedAndCannotRecoverMemoryAuthority`, `TestMarketQueueSaturationDoesNotConsumePeerQueue`, `TestMarketRestartAttemptAndDeadlineSaturateWithoutOverwritingFirstTypedRefusal`, `TestPairedMarketAbnormalReturnSchedulesOnlyLocalBoundedRestartAndKeepsEverySafetyLoopAlive`, `TestPairedMarketRestartHonorsPublishedAbsoluteDeadlineAfterHandoffRace`, `TestShutdownAndTriggerShareBarrierAndDrainBothQueues`, `TestStrategyEntrySupervisorDefaultsPairedMarketsDormant`, `TestStrategyEntrySupervisorPollerKeepsCompletePeerRunningWhenOtherMarketDormant`, `TestStrategyEntrySupervisorPollsDormantRefreshWorkerWithoutOpeningPublicTrigger`, `TestStrategyEntrySupervisorPollsKRAndUSImmediatelyInTheSameWave`, `TestStrategyEntrySupervisorRejectsUnsafeProductionPollIntervals`, `TestStrategyEntrySupervisorStartsKRAndUSCyclesConcurrently`, `TestStrategyEntrySupervisorZeroPollIntervalKeepsExplicitTriggerSemantics`, `TestStrategySupervisorRejectsInvalidAssemblies`, `TestTheFaultStreamHoldsOneSlotForEveryWorkerThatCanLatch`, `TestTheFourEscalationsThatStopTheEngineAreExactlyTheSupervisorsOwnBrokenBookkeeping`, `TestTheMarketThatLeadsAWaveAlwaysPublishesIt`, `TestTheOnlyWorkerProductionActuallyRunsSwallowsEveryCycleError`, `TestUSFXReadFailureDoesNotCancelKRIdentityOrSafetyBudgets` |
| B9 | if| 605:3 | arm entered 1x (engine tagged suite); arm entered 1x (engine untagged suite); `TestStrategySupervisorRejectsInvalidAssemblies` |
| B10 | if| 608:3 | arm entered 1x (engine tagged suite); arm entered 1x (engine untagged suite); `TestStrategySupervisorRejectsInvalidAssemblies` |
| B11 | if| 611:3 | arm entered 1x (engine tagged suite); arm entered 1x (engine untagged suite); `TestStrategySupervisorRejectsInvalidAssemblies` |
| B12 | if| 614:3 | no coverage block for this arm (engine tagged suite); no coverage block for this arm (engine untagged suite); no per-test profile in the attribution set entered it |
| B13 | if| 618:3 | no coverage block for this arm (engine tagged suite); no coverage block for this arm (engine untagged suite); no per-test profile in the attribution set entered it |
| B14 | if| 622:3 | no coverage block for this arm (engine tagged suite); no coverage block for this arm (engine untagged suite); no per-test profile in the attribution set entered it |
| B15 | if| 628:3 | arm entered 1x (engine tagged suite); arm entered 1x (engine untagged suite); `TestStrategySupervisorRejectsInvalidAssemblies` |
| B16 | if| 631:3 | arm entered 1x (engine tagged suite); arm entered 1x (engine untagged suite); `TestStrategySupervisorRejectsInvalidAssemblies` |
| B17 | range| 644:2 | arm entered 72x (engine tagged suite); arm entered 72x (engine untagged suite); `TestALatchedMarketSkipsTheTriggersAlreadySittingInItsQueue`, `TestARefreshOnlyWorkerSwallowsACentralIntegrityErrorToo`, `TestAnEffectiveMarketFaultLeavesItsPeerAndTheSupervisorAlone`, `TestBrokenSupervisorBookkeepingTakesTheSafetyLoopsDownWithIt`, `TestCentralIntegrityFailureEscapesOuterLoopAndDrainsSafety`, `TestContextIgnoringCycleWatchdogLatchesOnceAndLateResultHasNoAction`, `TestEveryWorkerCanHandOffItsFaultWithoutAnybodyDraining`, `TestExpiredAuthorityLatchesBeforeEvaluation`, `TestMarketFailureEmitsExactIrreversibleFaultAndKeepsPeerSafetyAlive`, `TestMarketPanicIsContainedAndCannotRecoverMemoryAuthority`, `TestMarketQueueSaturationDoesNotConsumePeerQueue`, `TestMarketRestartAttemptAndDeadlineSaturateWithoutOverwritingFirstTypedRefusal`, `TestPairedMarketAbnormalReturnSchedulesOnlyLocalBoundedRestartAndKeepsEverySafetyLoopAlive`, `TestPairedMarketRestartHonorsPublishedAbsoluteDeadlineAfterHandoffRace`, `TestShutdownAndTriggerShareBarrierAndDrainBothQueues`, `TestStrategyEntrySupervisorDefaultsPairedMarketsDormant`, `TestStrategyEntrySupervisorPollerKeepsCompletePeerRunningWhenOtherMarketDormant`, `TestStrategyEntrySupervisorPollsDormantRefreshWorkerWithoutOpeningPublicTrigger`, `TestStrategyEntrySupervisorPollsKRAndUSImmediatelyInTheSameWave`, `TestStrategyEntrySupervisorStartsKRAndUSCyclesConcurrently`, `TestStrategyEntrySupervisorZeroPollIntervalKeepsExplicitTriggerSemantics`, `TestTheFaultStreamHoldsOneSlotForEveryWorkerThatCanLatch`, `TestTheFourEscalationsThatStopTheEngineAreExactlyTheSupervisorsOwnBrokenBookkeeping`, `TestTheMarketThatLeadsAWaveAlwaysPublishesIt`, `TestTheOnlyWorkerProductionActuallyRunsSwallowsEveryCycleError`, `TestUSFXReadFailureDoesNotCancelKRIdentityOrSafetyBudgets` |
| B18 | if| 645:3 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| fmt.Errorf | 582:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 589:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| len | 591:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| errors.New | 592:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| clock.System | 596:9 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| clk.Now | 598:9 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| now.IsZero | 599:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| errors.New | 600:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 603:13 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validStrategyMarket | 605:7 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 606:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 609:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 612:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.AuthorityExpiresAt.IsZero | 614:70 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| now.Before | 615:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validStrategyDigest | 615:51 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 616:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.RestartNotBefore.IsZero | 618:124 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validStrategyWorkerRefusal | 619:96 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.RestartNotBefore.IsZero | 619:151 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 620:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 626:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.AuthorityExpiresAt.IsZero | 628:72 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 629:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.RestartNotBefore.IsZero | 631:123 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 632:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 636:22 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 646:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 651:62 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 651:91 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
