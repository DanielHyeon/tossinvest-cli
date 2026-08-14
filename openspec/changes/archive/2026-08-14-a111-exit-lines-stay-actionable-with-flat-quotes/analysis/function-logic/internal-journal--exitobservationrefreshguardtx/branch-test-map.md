# Branch Test Map: `exitObservationRefreshGuardTx`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Guard read failure propagates and the refresh transaction writes nothing | `TestA111RefreshFailureRollsBackTheWholeTuple`, `TestGuardedExitColumnsAreWrittenOnlyByTheApplyHook` | A111 rollback RED plus repository structural RED | structural RED GREEN 20 repetitions; A111 suite GREEN |
| Return | Current pending proposal is reported only as a boolean and causes refresh conflict with complete tuple/event invariance | `TestA111RefreshRejectsARealCurrentPendingProposalAndPreservesItsEvidence` | intentional A111 real-pending RED | focused A111 journal suite GREEN |
| Lifecycle | The caller keeps the same transaction across helper read and latest managed-generation check | `TestA111RefreshRejectsARealReleasedLifecycleAndPreservesItsTuple` | intentional A111 real-lifecycle RED | focused A111 journal suite GREEN |
