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

- Exact AST return nodes: `599:3, 606:3, 609:3, 617:3, 623:4, 626:4, 629:4, 633:4, 637:4, 643:4, 646:4, 649:4, 663:4, 667:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if| 595:2 | arm entered 46x (engine tagged suite); arm entered 46x (engine untagged suite); `TestARefreshOnlyWorkerSwallowsACentralIntegrityErrorToo`, `TestAnEffectiveMarketFaultLeavesItsPeerAndTheSupervisorAlone`, `TestBrokenSupervisorBookkeepingTakesTheSafetyLoopsDownWithIt`, `TestCentralIntegrityFailureEscapesOuterLoopAndDrainsSafety`, `TestContextIgnoringCycleWatchdogLatchesOnceAndLateResultHasNoAction`, `TestEveryWorkerCanHandOffItsFaultWithoutAnybodyDraining`, `TestExpiredAuthorityLatchesBeforeEvaluation`, `TestMarketFailureEmitsExactIrreversibleFaultAndKeepsPeerSafetyAlive`, `TestMarketPanicIsContainedAndCannotRecoverMemoryAuthority`, `TestMarketRestartAttemptAndDeadlineSaturateWithoutOverwritingFirstTypedRefusal`, `TestPairedMarketAbnormalReturnSchedulesOnlyLocalBoundedRestartAndKeepsEverySafetyLoopAlive`, `TestPairedMarketRestartHonorsPublishedAbsoluteDeadlineAfterHandoffRace`, `TestStrategyEntrySupervisorDefaultsPairedMarketsDormant`, `TestStrategyEntrySupervisorPollerKeepsCompletePeerRunningWhenOtherMarketDormant`, `TestStrategyEntrySupervisorPollsDormantRefreshWorkerWithoutOpeningPublicTrigger`, `TestStrategyEntrySupervisorPollsKRAndUSImmediatelyInTheSameWave`, `TestStrategyEntrySupervisorRejectsUnsafeProductionPollIntervals`, `TestStrategyEntrySupervisorStartsKRAndUSCyclesConcurrently`, `TestStrategyEntrySupervisorZeroPollIntervalKeepsExplicitTriggerSemantics`, `TestStrategySupervisorRejectsInvalidAssemblies`, `TestTheFaultStreamHoldsOneSlotForEveryWorkerThatCanLatch`, `TestTheFourEscalationsThatStopTheEngineAreExactlyTheSupervisorsOwnBrokenBookkeeping`, `TestTheMarketThatLeadsAWaveAlwaysPublishesIt`, `TestTheOnlyWorkerProductionActuallyRunsSwallowsEveryCycleError`, `TestUSFXReadFailureDoesNotCancelKRIdentityOrSafetyBudgets` |
| B2 | if| 598:2 | arm entered 1x (engine tagged suite); arm entered 1x (engine untagged suite); `TestStrategySupervisorRejectsInvalidAssemblies` |
| B3 | if| 602:2 | arm entered 34x (engine tagged suite); arm entered 34x (engine untagged suite); `TestCentralIntegrityFailureEscapesOuterLoopAndDrainsSafety`, `TestExpiredAuthorityLatchesBeforeEvaluation`, `TestMarketFailureEmitsExactIrreversibleFaultAndKeepsPeerSafetyAlive`, `TestMarketPanicIsContainedAndCannotRecoverMemoryAuthority`, `TestMarketQueueSaturationDoesNotConsumePeerQueue`, `TestMarketRestartAttemptAndDeadlineSaturateWithoutOverwritingFirstTypedRefusal`, `TestPairedMarketAbnormalReturnSchedulesOnlyLocalBoundedRestartAndKeepsEverySafetyLoopAlive`, `TestPairedMarketRestartHonorsPublishedAbsoluteDeadlineAfterHandoffRace`, `TestShutdownAndTriggerShareBarrierAndDrainBothQueues`, `TestStrategyEntrySupervisorDefaultsPairedMarketsDormant`, `TestStrategyEntrySupervisorPollerKeepsCompletePeerRunningWhenOtherMarketDormant`, `TestStrategyEntrySupervisorPollsDormantRefreshWorkerWithoutOpeningPublicTrigger`, `TestStrategyEntrySupervisorPollsKRAndUSImmediatelyInTheSameWave`, `TestStrategyEntrySupervisorRejectsUnsafeProductionPollIntervals`, `TestStrategyEntrySupervisorStartsKRAndUSCyclesConcurrently`, `TestStrategyEntrySupervisorZeroPollIntervalKeepsExplicitTriggerSemantics`, `TestStrategySupervisorRejectsInvalidAssemblies`, `TestUSFXReadFailureDoesNotCancelKRIdentityOrSafetyBudgets` |
| B4 | if| 605:2 | arm entered 1x (engine tagged suite); arm entered 1x (engine untagged suite); `TestStrategySupervisorRejectsInvalidAssemblies` |
| B5 | if| 608:2 | arm entered 1x (engine tagged suite); arm entered 1x (engine untagged suite); `TestStrategySupervisorRejectsInvalidAssemblies` |
| B6 | if| 612:2 | arm entered 23x (engine tagged suite); arm entered 23x (engine untagged suite); `TestCentralIntegrityFailureEscapesOuterLoopAndDrainsSafety`, `TestMarketFailureEmitsExactIrreversibleFaultAndKeepsPeerSafetyAlive`, `TestMarketPanicIsContainedAndCannotRecoverMemoryAuthority`, `TestMarketQueueSaturationDoesNotConsumePeerQueue`, `TestShutdownAndTriggerShareBarrierAndDrainBothQueues`, `TestStrategyEntrySupervisorDefaultsPairedMarketsDormant`, `TestStrategyEntrySupervisorPollerKeepsCompletePeerRunningWhenOtherMarketDormant`, `TestStrategyEntrySupervisorPollsDormantRefreshWorkerWithoutOpeningPublicTrigger`, `TestStrategyEntrySupervisorPollsKRAndUSImmediatelyInTheSameWave`, `TestStrategyEntrySupervisorRejectsUnsafeProductionPollIntervals`, `TestStrategyEntrySupervisorStartsKRAndUSCyclesConcurrently`, `TestStrategyEntrySupervisorZeroPollIntervalKeepsExplicitTriggerSemantics`, `TestStrategySupervisorRejectsInvalidAssemblies`, `TestUSFXReadFailureDoesNotCancelKRIdentityOrSafetyBudgets` |
| B7 | if| 616:2 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |
| B8 | range| 621:2 | arm entered 84x (engine tagged suite); arm entered 84x (engine untagged suite); `TestALatchedMarketSkipsTheTriggersAlreadySittingInItsQueue`, `TestARefreshOnlyWorkerSwallowsACentralIntegrityErrorToo`, `TestAnEffectiveMarketFaultLeavesItsPeerAndTheSupervisorAlone`, `TestBrokenSupervisorBookkeepingTakesTheSafetyLoopsDownWithIt`, `TestCentralIntegrityFailureEscapesOuterLoopAndDrainsSafety`, `TestContextIgnoringCycleWatchdogLatchesOnceAndLateResultHasNoAction`, `TestEveryWorkerCanHandOffItsFaultWithoutAnybodyDraining`, `TestExpiredAuthorityLatchesBeforeEvaluation`, `TestMarketFailureEmitsExactIrreversibleFaultAndKeepsPeerSafetyAlive`, `TestMarketPanicIsContainedAndCannotRecoverMemoryAuthority`, `TestMarketQueueSaturationDoesNotConsumePeerQueue`, `TestMarketRestartAttemptAndDeadlineSaturateWithoutOverwritingFirstTypedRefusal`, `TestPairedMarketAbnormalReturnSchedulesOnlyLocalBoundedRestartAndKeepsEverySafetyLoopAlive`, `TestPairedMarketRestartHonorsPublishedAbsoluteDeadlineAfterHandoffRace`, `TestShutdownAndTriggerShareBarrierAndDrainBothQueues`, `TestStrategyEntrySupervisorDefaultsPairedMarketsDormant`, `TestStrategyEntrySupervisorPollerKeepsCompletePeerRunningWhenOtherMarketDormant`, `TestStrategyEntrySupervisorPollsDormantRefreshWorkerWithoutOpeningPublicTrigger`, `TestStrategyEntrySupervisorPollsKRAndUSImmediatelyInTheSameWave`, `TestStrategyEntrySupervisorRejectsUnsafeProductionPollIntervals`, `TestStrategyEntrySupervisorStartsKRAndUSCyclesConcurrently`, `TestStrategyEntrySupervisorZeroPollIntervalKeepsExplicitTriggerSemantics`, `TestStrategySupervisorRejectsInvalidAssemblies`, `TestTheFaultStreamHoldsOneSlotForEveryWorkerThatCanLatch`, `TestTheFourEscalationsThatStopTheEngineAreExactlyTheSupervisorsOwnBrokenBookkeeping`, `TestTheMarketThatLeadsAWaveAlwaysPublishesIt`, `TestTheOnlyWorkerProductionActuallyRunsSwallowsEveryCycleError`, `TestUSFXReadFailureDoesNotCancelKRIdentityOrSafetyBudgets` |
| B9 | if| 622:3 | arm entered 1x (engine tagged suite); arm entered 1x (engine untagged suite); `TestStrategySupervisorRejectsInvalidAssemblies` |
| B10 | if| 625:3 | arm entered 1x (engine tagged suite); arm entered 1x (engine untagged suite); `TestStrategySupervisorRejectsInvalidAssemblies` |
| B11 | if| 628:3 | arm entered 1x (engine tagged suite); arm entered 1x (engine untagged suite); `TestStrategySupervisorRejectsInvalidAssemblies` |
| B12 | if| 631:3 | no coverage block for this arm (engine tagged suite); no coverage block for this arm (engine untagged suite); no per-test profile in the attribution set entered it |
| B13 | if| 635:3 | no coverage block for this arm (engine tagged suite); no coverage block for this arm (engine untagged suite); no per-test profile in the attribution set entered it |
| B14 | if| 639:3 | no coverage block for this arm (engine tagged suite); no coverage block for this arm (engine untagged suite); no per-test profile in the attribution set entered it |
| B15 | if| 645:3 | arm entered 1x (engine tagged suite); arm entered 1x (engine untagged suite); `TestStrategySupervisorRejectsInvalidAssemblies` |
| B16 | if| 648:3 | arm entered 1x (engine tagged suite); arm entered 1x (engine untagged suite); `TestStrategySupervisorRejectsInvalidAssemblies` |
| B17 | range| 661:2 | arm entered 72x (engine tagged suite); arm entered 72x (engine untagged suite); `TestALatchedMarketSkipsTheTriggersAlreadySittingInItsQueue`, `TestARefreshOnlyWorkerSwallowsACentralIntegrityErrorToo`, `TestAnEffectiveMarketFaultLeavesItsPeerAndTheSupervisorAlone`, `TestBrokenSupervisorBookkeepingTakesTheSafetyLoopsDownWithIt`, `TestCentralIntegrityFailureEscapesOuterLoopAndDrainsSafety`, `TestContextIgnoringCycleWatchdogLatchesOnceAndLateResultHasNoAction`, `TestEveryWorkerCanHandOffItsFaultWithoutAnybodyDraining`, `TestExpiredAuthorityLatchesBeforeEvaluation`, `TestMarketFailureEmitsExactIrreversibleFaultAndKeepsPeerSafetyAlive`, `TestMarketPanicIsContainedAndCannotRecoverMemoryAuthority`, `TestMarketQueueSaturationDoesNotConsumePeerQueue`, `TestMarketRestartAttemptAndDeadlineSaturateWithoutOverwritingFirstTypedRefusal`, `TestPairedMarketAbnormalReturnSchedulesOnlyLocalBoundedRestartAndKeepsEverySafetyLoopAlive`, `TestPairedMarketRestartHonorsPublishedAbsoluteDeadlineAfterHandoffRace`, `TestShutdownAndTriggerShareBarrierAndDrainBothQueues`, `TestStrategyEntrySupervisorDefaultsPairedMarketsDormant`, `TestStrategyEntrySupervisorPollerKeepsCompletePeerRunningWhenOtherMarketDormant`, `TestStrategyEntrySupervisorPollsDormantRefreshWorkerWithoutOpeningPublicTrigger`, `TestStrategyEntrySupervisorPollsKRAndUSImmediatelyInTheSameWave`, `TestStrategyEntrySupervisorStartsKRAndUSCyclesConcurrently`, `TestStrategyEntrySupervisorZeroPollIntervalKeepsExplicitTriggerSemantics`, `TestTheFaultStreamHoldsOneSlotForEveryWorkerThatCanLatch`, `TestTheFourEscalationsThatStopTheEngineAreExactlyTheSupervisorsOwnBrokenBookkeeping`, `TestTheMarketThatLeadsAWaveAlwaysPublishesIt`, `TestTheOnlyWorkerProductionActuallyRunsSwallowsEveryCycleError`, `TestUSFXReadFailureDoesNotCancelKRIdentityOrSafetyBudgets` |
| B18 | if| 662:3 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| fmt.Errorf | 599:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 606:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| len | 608:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| errors.New | 609:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| clock.System | 613:9 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| clk.Now | 615:9 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| now.IsZero | 616:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| errors.New | 617:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 620:13 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validStrategyMarket | 622:7 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 623:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 626:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 629:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.AuthorityExpiresAt.IsZero | 631:70 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| now.Before | 632:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validStrategyDigest | 632:51 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 633:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.RestartNotBefore.IsZero | 635:124 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validStrategyWorkerRefusal | 636:96 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.RestartNotBefore.IsZero | 636:151 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 637:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 643:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.AuthorityExpiresAt.IsZero | 645:72 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 646:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.RestartNotBefore.IsZero | 648:123 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 649:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 653:22 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 663:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 668:62 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 668:91 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
