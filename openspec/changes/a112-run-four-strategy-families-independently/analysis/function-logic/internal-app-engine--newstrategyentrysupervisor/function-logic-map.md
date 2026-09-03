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

- Exact AST return nodes: `580:3, 587:3, 590:3, 598:3, 604:4, 607:4, 610:4, 614:4, 618:4, 624:4, 627:4, 630:4, 644:4, 648:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if| 576:2 | arm entered 46x (engine tagged suite); arm entered 46x (engine untagged suite); `TestARefreshOnlyWorkerSwallowsACentralIntegrityErrorToo`, `TestAnEffectiveMarketFaultLeavesItsPeerAndTheSupervisorAlone`, `TestBrokenSupervisorBookkeepingTakesTheSafetyLoopsDownWithIt`, `TestCentralIntegrityFailureEscapesOuterLoopAndDrainsSafety`, `TestContextIgnoringCycleWatchdogLatchesOnceAndLateResultHasNoAction`, `TestEveryWorkerCanHandOffItsFaultWithoutAnybodyDraining`, `TestExpiredAuthorityLatchesBeforeEvaluation`, `TestMarketFailureEmitsExactIrreversibleFaultAndKeepsPeerSafetyAlive`, `TestMarketPanicIsContainedAndCannotRecoverMemoryAuthority`, `TestMarketRestartAttemptAndDeadlineSaturateWithoutOverwritingFirstTypedRefusal`, `TestPairedMarketAbnormalReturnSchedulesOnlyLocalBoundedRestartAndKeepsEverySafetyLoopAlive`, `TestPairedMarketRestartHonorsPublishedAbsoluteDeadlineAfterHandoffRace`, `TestStrategyEntrySupervisorDefaultsPairedMarketsDormant`, `TestStrategyEntrySupervisorPollerKeepsCompletePeerRunningWhenOtherMarketDormant`, `TestStrategyEntrySupervisorPollsDormantRefreshWorkerWithoutOpeningPublicTrigger`, `TestStrategyEntrySupervisorPollsKRAndUSImmediatelyInTheSameWave`, `TestStrategyEntrySupervisorRejectsUnsafeProductionPollIntervals`, `TestStrategyEntrySupervisorStartsKRAndUSCyclesConcurrently`, `TestStrategyEntrySupervisorZeroPollIntervalKeepsExplicitTriggerSemantics`, `TestStrategySupervisorRejectsInvalidAssemblies`, `TestTheFaultStreamHoldsOneSlotForEveryWorkerThatCanLatch`, `TestTheFourEscalationsThatStopTheEngineAreExactlyTheSupervisorsOwnBrokenBookkeeping`, `TestTheMarketThatLeadsAWaveAlwaysPublishesIt`, `TestTheOnlyWorkerProductionActuallyRunsSwallowsEveryCycleError`, `TestUSFXReadFailureDoesNotCancelKRIdentityOrSafetyBudgets` |
| B2 | if| 579:2 | arm entered 1x (engine tagged suite); arm entered 1x (engine untagged suite); `TestStrategySupervisorRejectsInvalidAssemblies` |
| B3 | if| 583:2 | arm entered 34x (engine tagged suite); arm entered 34x (engine untagged suite); `TestCentralIntegrityFailureEscapesOuterLoopAndDrainsSafety`, `TestExpiredAuthorityLatchesBeforeEvaluation`, `TestMarketFailureEmitsExactIrreversibleFaultAndKeepsPeerSafetyAlive`, `TestMarketPanicIsContainedAndCannotRecoverMemoryAuthority`, `TestMarketQueueSaturationDoesNotConsumePeerQueue`, `TestMarketRestartAttemptAndDeadlineSaturateWithoutOverwritingFirstTypedRefusal`, `TestPairedMarketAbnormalReturnSchedulesOnlyLocalBoundedRestartAndKeepsEverySafetyLoopAlive`, `TestPairedMarketRestartHonorsPublishedAbsoluteDeadlineAfterHandoffRace`, `TestShutdownAndTriggerShareBarrierAndDrainBothQueues`, `TestStrategyEntrySupervisorDefaultsPairedMarketsDormant`, `TestStrategyEntrySupervisorPollerKeepsCompletePeerRunningWhenOtherMarketDormant`, `TestStrategyEntrySupervisorPollsDormantRefreshWorkerWithoutOpeningPublicTrigger`, `TestStrategyEntrySupervisorPollsKRAndUSImmediatelyInTheSameWave`, `TestStrategyEntrySupervisorRejectsUnsafeProductionPollIntervals`, `TestStrategyEntrySupervisorStartsKRAndUSCyclesConcurrently`, `TestStrategyEntrySupervisorZeroPollIntervalKeepsExplicitTriggerSemantics`, `TestStrategySupervisorRejectsInvalidAssemblies`, `TestUSFXReadFailureDoesNotCancelKRIdentityOrSafetyBudgets` |
| B4 | if| 586:2 | arm entered 1x (engine tagged suite); arm entered 1x (engine untagged suite); `TestStrategySupervisorRejectsInvalidAssemblies` |
| B5 | if| 589:2 | arm entered 1x (engine tagged suite); arm entered 1x (engine untagged suite); `TestStrategySupervisorRejectsInvalidAssemblies` |
| B6 | if| 593:2 | arm entered 23x (engine tagged suite); arm entered 23x (engine untagged suite); `TestCentralIntegrityFailureEscapesOuterLoopAndDrainsSafety`, `TestMarketFailureEmitsExactIrreversibleFaultAndKeepsPeerSafetyAlive`, `TestMarketPanicIsContainedAndCannotRecoverMemoryAuthority`, `TestMarketQueueSaturationDoesNotConsumePeerQueue`, `TestShutdownAndTriggerShareBarrierAndDrainBothQueues`, `TestStrategyEntrySupervisorDefaultsPairedMarketsDormant`, `TestStrategyEntrySupervisorPollerKeepsCompletePeerRunningWhenOtherMarketDormant`, `TestStrategyEntrySupervisorPollsDormantRefreshWorkerWithoutOpeningPublicTrigger`, `TestStrategyEntrySupervisorPollsKRAndUSImmediatelyInTheSameWave`, `TestStrategyEntrySupervisorRejectsUnsafeProductionPollIntervals`, `TestStrategyEntrySupervisorStartsKRAndUSCyclesConcurrently`, `TestStrategyEntrySupervisorZeroPollIntervalKeepsExplicitTriggerSemantics`, `TestStrategySupervisorRejectsInvalidAssemblies`, `TestUSFXReadFailureDoesNotCancelKRIdentityOrSafetyBudgets` |
| B7 | if| 597:2 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |
| B8 | range| 602:2 | arm entered 84x (engine tagged suite); arm entered 84x (engine untagged suite); `TestALatchedMarketSkipsTheTriggersAlreadySittingInItsQueue`, `TestARefreshOnlyWorkerSwallowsACentralIntegrityErrorToo`, `TestAnEffectiveMarketFaultLeavesItsPeerAndTheSupervisorAlone`, `TestBrokenSupervisorBookkeepingTakesTheSafetyLoopsDownWithIt`, `TestCentralIntegrityFailureEscapesOuterLoopAndDrainsSafety`, `TestContextIgnoringCycleWatchdogLatchesOnceAndLateResultHasNoAction`, `TestEveryWorkerCanHandOffItsFaultWithoutAnybodyDraining`, `TestExpiredAuthorityLatchesBeforeEvaluation`, `TestMarketFailureEmitsExactIrreversibleFaultAndKeepsPeerSafetyAlive`, `TestMarketPanicIsContainedAndCannotRecoverMemoryAuthority`, `TestMarketQueueSaturationDoesNotConsumePeerQueue`, `TestMarketRestartAttemptAndDeadlineSaturateWithoutOverwritingFirstTypedRefusal`, `TestPairedMarketAbnormalReturnSchedulesOnlyLocalBoundedRestartAndKeepsEverySafetyLoopAlive`, `TestPairedMarketRestartHonorsPublishedAbsoluteDeadlineAfterHandoffRace`, `TestShutdownAndTriggerShareBarrierAndDrainBothQueues`, `TestStrategyEntrySupervisorDefaultsPairedMarketsDormant`, `TestStrategyEntrySupervisorPollerKeepsCompletePeerRunningWhenOtherMarketDormant`, `TestStrategyEntrySupervisorPollsDormantRefreshWorkerWithoutOpeningPublicTrigger`, `TestStrategyEntrySupervisorPollsKRAndUSImmediatelyInTheSameWave`, `TestStrategyEntrySupervisorRejectsUnsafeProductionPollIntervals`, `TestStrategyEntrySupervisorStartsKRAndUSCyclesConcurrently`, `TestStrategyEntrySupervisorZeroPollIntervalKeepsExplicitTriggerSemantics`, `TestStrategySupervisorRejectsInvalidAssemblies`, `TestTheFaultStreamHoldsOneSlotForEveryWorkerThatCanLatch`, `TestTheFourEscalationsThatStopTheEngineAreExactlyTheSupervisorsOwnBrokenBookkeeping`, `TestTheMarketThatLeadsAWaveAlwaysPublishesIt`, `TestTheOnlyWorkerProductionActuallyRunsSwallowsEveryCycleError`, `TestUSFXReadFailureDoesNotCancelKRIdentityOrSafetyBudgets` |
| B9 | if| 603:3 | arm entered 1x (engine tagged suite); arm entered 1x (engine untagged suite); `TestStrategySupervisorRejectsInvalidAssemblies` |
| B10 | if| 606:3 | arm entered 1x (engine tagged suite); arm entered 1x (engine untagged suite); `TestStrategySupervisorRejectsInvalidAssemblies` |
| B11 | if| 609:3 | arm entered 1x (engine tagged suite); arm entered 1x (engine untagged suite); `TestStrategySupervisorRejectsInvalidAssemblies` |
| B12 | if| 612:3 | no coverage block for this arm (engine tagged suite); no coverage block for this arm (engine untagged suite); no per-test profile in the attribution set entered it |
| B13 | if| 616:3 | no coverage block for this arm (engine tagged suite); no coverage block for this arm (engine untagged suite); no per-test profile in the attribution set entered it |
| B14 | if| 620:3 | no coverage block for this arm (engine tagged suite); no coverage block for this arm (engine untagged suite); no per-test profile in the attribution set entered it |
| B15 | if| 626:3 | arm entered 1x (engine tagged suite); arm entered 1x (engine untagged suite); `TestStrategySupervisorRejectsInvalidAssemblies` |
| B16 | if| 629:3 | arm entered 1x (engine tagged suite); arm entered 1x (engine untagged suite); `TestStrategySupervisorRejectsInvalidAssemblies` |
| B17 | range| 642:2 | arm entered 72x (engine tagged suite); arm entered 72x (engine untagged suite); `TestALatchedMarketSkipsTheTriggersAlreadySittingInItsQueue`, `TestARefreshOnlyWorkerSwallowsACentralIntegrityErrorToo`, `TestAnEffectiveMarketFaultLeavesItsPeerAndTheSupervisorAlone`, `TestBrokenSupervisorBookkeepingTakesTheSafetyLoopsDownWithIt`, `TestCentralIntegrityFailureEscapesOuterLoopAndDrainsSafety`, `TestContextIgnoringCycleWatchdogLatchesOnceAndLateResultHasNoAction`, `TestEveryWorkerCanHandOffItsFaultWithoutAnybodyDraining`, `TestExpiredAuthorityLatchesBeforeEvaluation`, `TestMarketFailureEmitsExactIrreversibleFaultAndKeepsPeerSafetyAlive`, `TestMarketPanicIsContainedAndCannotRecoverMemoryAuthority`, `TestMarketQueueSaturationDoesNotConsumePeerQueue`, `TestMarketRestartAttemptAndDeadlineSaturateWithoutOverwritingFirstTypedRefusal`, `TestPairedMarketAbnormalReturnSchedulesOnlyLocalBoundedRestartAndKeepsEverySafetyLoopAlive`, `TestPairedMarketRestartHonorsPublishedAbsoluteDeadlineAfterHandoffRace`, `TestShutdownAndTriggerShareBarrierAndDrainBothQueues`, `TestStrategyEntrySupervisorDefaultsPairedMarketsDormant`, `TestStrategyEntrySupervisorPollerKeepsCompletePeerRunningWhenOtherMarketDormant`, `TestStrategyEntrySupervisorPollsDormantRefreshWorkerWithoutOpeningPublicTrigger`, `TestStrategyEntrySupervisorPollsKRAndUSImmediatelyInTheSameWave`, `TestStrategyEntrySupervisorStartsKRAndUSCyclesConcurrently`, `TestStrategyEntrySupervisorZeroPollIntervalKeepsExplicitTriggerSemantics`, `TestTheFaultStreamHoldsOneSlotForEveryWorkerThatCanLatch`, `TestTheFourEscalationsThatStopTheEngineAreExactlyTheSupervisorsOwnBrokenBookkeeping`, `TestTheMarketThatLeadsAWaveAlwaysPublishesIt`, `TestTheOnlyWorkerProductionActuallyRunsSwallowsEveryCycleError`, `TestUSFXReadFailureDoesNotCancelKRIdentityOrSafetyBudgets` |
| B18 | if| 643:3 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| fmt.Errorf | 580:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 587:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| len | 589:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| errors.New | 590:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| clock.System | 594:9 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| clk.Now | 596:9 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| now.IsZero | 597:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| errors.New | 598:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 601:13 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validStrategyMarket | 603:7 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 604:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 607:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 610:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.AuthorityExpiresAt.IsZero | 612:70 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| now.Before | 613:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validStrategyDigest | 613:51 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 614:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.RestartNotBefore.IsZero | 616:124 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validStrategyWorkerRefusal | 617:96 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.RestartNotBefore.IsZero | 617:151 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 618:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 624:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.AuthorityExpiresAt.IsZero | 626:72 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 627:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.RestartNotBefore.IsZero | 629:123 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 630:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 634:22 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 644:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 649:62 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 649:91 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
