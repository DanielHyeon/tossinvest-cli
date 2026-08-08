# Branch Test Map: `consoleVerifyStarter`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | updater owns flock; refusal precedes broker build | `TestVerificationEntryPointsRefuseWhileSystemUpdateOwnsEngineExclusion/console_verify_starter` | failed: 1 broker build | pass: `ErrAlreadyRunning`, 0 builds |
| B2 | record path resolution fails | existing console verifier tests | baseline pass | full console/cmd package pass |
| B3 | record load fails | existing console verifier tests | baseline pass | full console/cmd package pass |
| B4 | fresh official broker/account construction fails | existing account tests | baseline pass | full console/cmd package pass |
| B5 | recorder open fails | existing recorder tests | baseline pass | full console/cmd package pass |
| B6 | guarded runner construction fails | existing verifylive option tests | baseline pass | full console/cmd package pass |
| B7 | cancellation/deadline is normalized after cleanup | existing shutdown tests | baseline pass | full console/cmd package pass |
