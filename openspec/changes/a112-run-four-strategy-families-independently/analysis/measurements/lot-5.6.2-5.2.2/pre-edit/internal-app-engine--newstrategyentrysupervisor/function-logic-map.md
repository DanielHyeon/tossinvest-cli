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

- Exact AST return nodes: `606:3, 613:3, 616:3, 624:3, 630:4, 633:4, 636:4, 640:4, 644:4, 650:4, 653:4, 656:4, 670:4, 674:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if| 602:2 | arm entered 46x (engine tagged suite); arm entered 46x (engine untagged suite); `TestARefreshOnlyWorkerSwallowsACentralIntegrityErrorToo`, `TestAnEffectiveMarketFaultLeavesItsPeerAndTheSupervisorAlone`, `TestBrokenSupervisorBookkeepingTakesTheSafetyLoopsDownWithIt`, `TestCentralIntegrityFailureEscapesOuterLoopAndDrainsSafety`, `TestContextIgnoringCycleWatchdogLatchesOnceAndLateResultHasNoAction`, `TestEveryWorkerCanHandOffItsFaultWithoutAnybodyDraining`, `TestExpiredAuthorityLatchesBeforeEvaluation`, `TestMarketFailureEmitsExactIrreversibleFaultAndKeepsPeerSafetyAlive`, `TestMarketPanicIsContainedAndCannotRecoverMemoryAuthority`, `TestMarketRestartAttemptAndDeadlineSaturateWithoutOverwritingFirstTypedRefusal`, `TestPairedMarketAbnormalReturnSchedulesOnlyLocalBoundedRestartAndKeepsEverySafetyLoopAlive`, `TestPairedMarketRestartHonorsPublishedAbsoluteDeadlineAfterHandoffRace`, `TestStrategyEntrySupervisorDefaultsPairedMarketsDormant`, `TestStrategyEntrySupervisorPollerKeepsCompletePeerRunningWhenOtherMarketDormant`, `TestStrategyEntrySupervisorPollsDormantRefreshWorkerWithoutOpeningPublicTrigger`, `TestStrategyEntrySupervisorPollsKRAndUSImmediatelyInTheSameWave`, `TestStrategyEntrySupervisorRejectsUnsafeProductionPollIntervals`, `TestStrategyEntrySupervisorStartsKRAndUSCyclesConcurrently`, `TestStrategyEntrySupervisorZeroPollIntervalKeepsExplicitTriggerSemantics`, `TestStrategySupervisorRejectsInvalidAssemblies`, `TestTheFaultStreamHoldsOneSlotForEveryWorkerThatCanLatch`, `TestTheFourEscalationsThatStopTheEngineAreExactlyTheSupervisorsOwnBrokenBookkeeping`, `TestTheMarketThatLeadsAWaveAlwaysPublishesIt`, `TestTheOnlyWorkerProductionActuallyRunsSwallowsEveryCycleError`, `TestUSFXReadFailureDoesNotCancelKRIdentityOrSafetyBudgets` |
| B2 | if| 605:2 | arm entered 1x (engine tagged suite); arm entered 1x (engine untagged suite); `TestStrategySupervisorRejectsInvalidAssemblies` |
| B3 | if| 609:2 | arm entered 34x (engine tagged suite); arm entered 34x (engine untagged suite); `TestCentralIntegrityFailureEscapesOuterLoopAndDrainsSafety`, `TestExpiredAuthorityLatchesBeforeEvaluation`, `TestMarketFailureEmitsExactIrreversibleFaultAndKeepsPeerSafetyAlive`, `TestMarketPanicIsContainedAndCannotRecoverMemoryAuthority`, `TestMarketQueueSaturationDoesNotConsumePeerQueue`, `TestMarketRestartAttemptAndDeadlineSaturateWithoutOverwritingFirstTypedRefusal`, `TestPairedMarketAbnormalReturnSchedulesOnlyLocalBoundedRestartAndKeepsEverySafetyLoopAlive`, `TestPairedMarketRestartHonorsPublishedAbsoluteDeadlineAfterHandoffRace`, `TestShutdownAndTriggerShareBarrierAndDrainBothQueues`, `TestStrategyEntrySupervisorDefaultsPairedMarketsDormant`, `TestStrategyEntrySupervisorPollerKeepsCompletePeerRunningWhenOtherMarketDormant`, `TestStrategyEntrySupervisorPollsDormantRefreshWorkerWithoutOpeningPublicTrigger`, `TestStrategyEntrySupervisorPollsKRAndUSImmediatelyInTheSameWave`, `TestStrategyEntrySupervisorRejectsUnsafeProductionPollIntervals`, `TestStrategyEntrySupervisorStartsKRAndUSCyclesConcurrently`, `TestStrategyEntrySupervisorZeroPollIntervalKeepsExplicitTriggerSemantics`, `TestStrategySupervisorRejectsInvalidAssemblies`, `TestUSFXReadFailureDoesNotCancelKRIdentityOrSafetyBudgets` |
| B4 | if| 612:2 | arm entered 1x (engine tagged suite); arm entered 1x (engine untagged suite); `TestStrategySupervisorRejectsInvalidAssemblies` |
| B5 | if| 615:2 | arm entered 1x (engine tagged suite); arm entered 1x (engine untagged suite); `TestStrategySupervisorRejectsInvalidAssemblies` |
| B6 | if| 619:2 | arm entered 23x (engine tagged suite); arm entered 23x (engine untagged suite); `TestCentralIntegrityFailureEscapesOuterLoopAndDrainsSafety`, `TestMarketFailureEmitsExactIrreversibleFaultAndKeepsPeerSafetyAlive`, `TestMarketPanicIsContainedAndCannotRecoverMemoryAuthority`, `TestMarketQueueSaturationDoesNotConsumePeerQueue`, `TestShutdownAndTriggerShareBarrierAndDrainBothQueues`, `TestStrategyEntrySupervisorDefaultsPairedMarketsDormant`, `TestStrategyEntrySupervisorPollerKeepsCompletePeerRunningWhenOtherMarketDormant`, `TestStrategyEntrySupervisorPollsDormantRefreshWorkerWithoutOpeningPublicTrigger`, `TestStrategyEntrySupervisorPollsKRAndUSImmediatelyInTheSameWave`, `TestStrategyEntrySupervisorRejectsUnsafeProductionPollIntervals`, `TestStrategyEntrySupervisorStartsKRAndUSCyclesConcurrently`, `TestStrategyEntrySupervisorZeroPollIntervalKeepsExplicitTriggerSemantics`, `TestStrategySupervisorRejectsInvalidAssemblies`, `TestUSFXReadFailureDoesNotCancelKRIdentityOrSafetyBudgets` |
| B7 | if| 623:2 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |
| B8 | range| 628:2 | arm entered 84x (engine tagged suite); arm entered 84x (engine untagged suite); `TestALatchedMarketSkipsTheTriggersAlreadySittingInItsQueue`, `TestARefreshOnlyWorkerSwallowsACentralIntegrityErrorToo`, `TestAnEffectiveMarketFaultLeavesItsPeerAndTheSupervisorAlone`, `TestBrokenSupervisorBookkeepingTakesTheSafetyLoopsDownWithIt`, `TestCentralIntegrityFailureEscapesOuterLoopAndDrainsSafety`, `TestContextIgnoringCycleWatchdogLatchesOnceAndLateResultHasNoAction`, `TestEveryWorkerCanHandOffItsFaultWithoutAnybodyDraining`, `TestExpiredAuthorityLatchesBeforeEvaluation`, `TestMarketFailureEmitsExactIrreversibleFaultAndKeepsPeerSafetyAlive`, `TestMarketPanicIsContainedAndCannotRecoverMemoryAuthority`, `TestMarketQueueSaturationDoesNotConsumePeerQueue`, `TestMarketRestartAttemptAndDeadlineSaturateWithoutOverwritingFirstTypedRefusal`, `TestPairedMarketAbnormalReturnSchedulesOnlyLocalBoundedRestartAndKeepsEverySafetyLoopAlive`, `TestPairedMarketRestartHonorsPublishedAbsoluteDeadlineAfterHandoffRace`, `TestShutdownAndTriggerShareBarrierAndDrainBothQueues`, `TestStrategyEntrySupervisorDefaultsPairedMarketsDormant`, `TestStrategyEntrySupervisorPollerKeepsCompletePeerRunningWhenOtherMarketDormant`, `TestStrategyEntrySupervisorPollsDormantRefreshWorkerWithoutOpeningPublicTrigger`, `TestStrategyEntrySupervisorPollsKRAndUSImmediatelyInTheSameWave`, `TestStrategyEntrySupervisorRejectsUnsafeProductionPollIntervals`, `TestStrategyEntrySupervisorStartsKRAndUSCyclesConcurrently`, `TestStrategyEntrySupervisorZeroPollIntervalKeepsExplicitTriggerSemantics`, `TestStrategySupervisorRejectsInvalidAssemblies`, `TestTheFaultStreamHoldsOneSlotForEveryWorkerThatCanLatch`, `TestTheFourEscalationsThatStopTheEngineAreExactlyTheSupervisorsOwnBrokenBookkeeping`, `TestTheMarketThatLeadsAWaveAlwaysPublishesIt`, `TestTheOnlyWorkerProductionActuallyRunsSwallowsEveryCycleError`, `TestUSFXReadFailureDoesNotCancelKRIdentityOrSafetyBudgets` |
| B9 | if| 629:3 | arm entered 1x (engine tagged suite); arm entered 1x (engine untagged suite); `TestStrategySupervisorRejectsInvalidAssemblies` |
| B10 | if| 632:3 | arm entered 1x (engine tagged suite); arm entered 1x (engine untagged suite); `TestStrategySupervisorRejectsInvalidAssemblies` |
| B11 | if| 635:3 | arm entered 1x (engine tagged suite); arm entered 1x (engine untagged suite); `TestStrategySupervisorRejectsInvalidAssemblies` |
| B12 | if| 638:3 | no coverage block for this arm (engine tagged suite); no coverage block for this arm (engine untagged suite); no per-test profile in the attribution set entered it |
| B13 | if| 642:3 | no coverage block for this arm (engine tagged suite); no coverage block for this arm (engine untagged suite); no per-test profile in the attribution set entered it |
| B14 | if| 646:3 | no coverage block for this arm (engine tagged suite); no coverage block for this arm (engine untagged suite); no per-test profile in the attribution set entered it |
| B15 | if| 652:3 | arm entered 1x (engine tagged suite); arm entered 1x (engine untagged suite); `TestStrategySupervisorRejectsInvalidAssemblies` |
| B16 | if| 655:3 | arm entered 1x (engine tagged suite); arm entered 1x (engine untagged suite); `TestStrategySupervisorRejectsInvalidAssemblies` |
| B17 | range| 668:2 | arm entered 72x (engine tagged suite); arm entered 72x (engine untagged suite); `TestALatchedMarketSkipsTheTriggersAlreadySittingInItsQueue`, `TestARefreshOnlyWorkerSwallowsACentralIntegrityErrorToo`, `TestAnEffectiveMarketFaultLeavesItsPeerAndTheSupervisorAlone`, `TestBrokenSupervisorBookkeepingTakesTheSafetyLoopsDownWithIt`, `TestCentralIntegrityFailureEscapesOuterLoopAndDrainsSafety`, `TestContextIgnoringCycleWatchdogLatchesOnceAndLateResultHasNoAction`, `TestEveryWorkerCanHandOffItsFaultWithoutAnybodyDraining`, `TestExpiredAuthorityLatchesBeforeEvaluation`, `TestMarketFailureEmitsExactIrreversibleFaultAndKeepsPeerSafetyAlive`, `TestMarketPanicIsContainedAndCannotRecoverMemoryAuthority`, `TestMarketQueueSaturationDoesNotConsumePeerQueue`, `TestMarketRestartAttemptAndDeadlineSaturateWithoutOverwritingFirstTypedRefusal`, `TestPairedMarketAbnormalReturnSchedulesOnlyLocalBoundedRestartAndKeepsEverySafetyLoopAlive`, `TestPairedMarketRestartHonorsPublishedAbsoluteDeadlineAfterHandoffRace`, `TestShutdownAndTriggerShareBarrierAndDrainBothQueues`, `TestStrategyEntrySupervisorDefaultsPairedMarketsDormant`, `TestStrategyEntrySupervisorPollerKeepsCompletePeerRunningWhenOtherMarketDormant`, `TestStrategyEntrySupervisorPollsDormantRefreshWorkerWithoutOpeningPublicTrigger`, `TestStrategyEntrySupervisorPollsKRAndUSImmediatelyInTheSameWave`, `TestStrategyEntrySupervisorStartsKRAndUSCyclesConcurrently`, `TestStrategyEntrySupervisorZeroPollIntervalKeepsExplicitTriggerSemantics`, `TestTheFaultStreamHoldsOneSlotForEveryWorkerThatCanLatch`, `TestTheFourEscalationsThatStopTheEngineAreExactlyTheSupervisorsOwnBrokenBookkeeping`, `TestTheMarketThatLeadsAWaveAlwaysPublishesIt`, `TestTheOnlyWorkerProductionActuallyRunsSwallowsEveryCycleError`, `TestUSFXReadFailureDoesNotCancelKRIdentityOrSafetyBudgets` |
| B18 | if| 669:3 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| fmt.Errorf | 606:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 613:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| len | 615:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| errors.New | 616:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| clock.System | 620:9 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| clk.Now | 622:9 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| now.IsZero | 623:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| errors.New | 624:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 627:13 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validStrategyMarket | 629:7 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 630:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 633:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 636:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.AuthorityExpiresAt.IsZero | 638:70 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| now.Before | 639:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validStrategyDigest | 639:51 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 640:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.RestartNotBefore.IsZero | 642:124 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validStrategyWorkerRefusal | 643:96 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.RestartNotBefore.IsZero | 643:151 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 644:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 650:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.AuthorityExpiresAt.IsZero | 652:72 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 653:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.RestartNotBefore.IsZero | 655:123 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 656:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 660:22 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 670:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 675:62 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 675:91 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
