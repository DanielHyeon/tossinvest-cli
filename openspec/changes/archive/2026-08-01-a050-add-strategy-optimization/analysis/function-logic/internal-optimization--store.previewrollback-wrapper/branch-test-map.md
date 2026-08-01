# Branch Test Map: `Store.PreviewRollback`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | happy path delegates with the immutable store actor | `TestRollbackPreviewAndConflictRecoveryUseTheSameActor`, `TestDirectStoreCandidatesRemainBoundToTheStoreActor` | mutable actor binding existed before hardening | PASS |
