# Branch Test Map: `armExitProposalTx`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Missing exit row cannot be armed | `TestGuardedExitColumnsAreWrittenOnlyByTheApplyHook` | structural preservation coverage | repository structural test GREEN |
| B2 | Proposal-state read error propagates without a write | `TestGuardedExitColumnsAreWrittenOnlyByTheApplyHook` | structural preservation coverage | repository structural test GREEN |
| B3 | A second proposal is refused while one is outstanding | `TestASecondProposalIsRefusedWhileOneIsOutstanding` | existing dedupe contract | journal suite GREEN |
| B4 | Failed guarded-tuple write rolls back the caller transaction | `TestAFailingExitHookRollsBackTheProjectionToo` | existing rollback contract | journal suite GREEN |
| Return | Successful arm writes the pending tuple only at apply-hook authority | `TestExitSnapshotDuplicateDecisionIsNotRearmed`, `TestGuardedExitColumnsAreWrittenOnlyByTheApplyHook` | retained structural gate RED | structural RED GREEN 20 repetitions |
