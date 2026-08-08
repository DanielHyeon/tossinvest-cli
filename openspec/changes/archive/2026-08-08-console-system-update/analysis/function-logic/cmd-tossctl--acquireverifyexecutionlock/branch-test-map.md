# Branch Test Map: `acquireVerifyExecutionLock`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | isolated config resolves its own exclusion directory | existing config-dir tests | no | pass |
| B2 | updater already owns the same real flock | `TestVerificationEntryPointsRefuseWhileSystemUpdateOwnsEngineExclusion` | verification reached broker | pass: both entry points refuse before broker |
