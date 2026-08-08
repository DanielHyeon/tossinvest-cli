# Branch Test Map: `runVerifyRun`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `--list` remains read-only and lock-free | existing verify list tests | baseline pass | full command package pass |
| B2 | nil Cobra context gets background fallback | existing verification tests | baseline pass | full command package pass |
| B3 | updater owns flock; refusal precedes broker build | `TestVerificationEntryPointsRefuseWhileSystemUpdateOwnsEngineExclusion/standalone_verify_run` | failed: 1 broker build | pass: `ErrAlreadyRunning`, 0 builds |
| B4 | record path resolution fails | existing record-path tests | baseline pass | full command package pass |
| B5 | record load fails | existing record tests | baseline pass | full command package pass |
| B6 | prior settled steps lack resume/redo | existing resume tests | baseline pass | full command package pass |
| B7 | broker/account construction fails | existing account tests | baseline pass | full command package pass |
| B8 | holding symbol is absent | existing holding-selection tests | baseline pass | full command package pass |
| B9 | US run receives non-US probe symbol | existing market normalization tests | baseline pass | full command package pass |
| B10 | recorder open fails | existing recorder tests | baseline pass | full command package pass |
| B11 | guarded runner construction fails | existing verifylive option tests | baseline pass | full command package pass |
| B12 | runner cancellation/deadline preserves evidence | existing interrupt tests | baseline pass | full command package pass |
