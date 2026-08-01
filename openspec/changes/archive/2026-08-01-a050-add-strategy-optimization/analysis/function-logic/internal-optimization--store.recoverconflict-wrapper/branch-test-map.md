# Branch Test Map: `Store.RecoverConflict`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | happy path delegates with immutable store actor | `TestCandidateCapabilityIsBoundToPreviewActorAcrossApplyRecoveryAndReplay`, `TestRollbackPreviewAndConflictRecoveryUseTheSameActor` | mutable actor binding existed before hardening | PASS |
