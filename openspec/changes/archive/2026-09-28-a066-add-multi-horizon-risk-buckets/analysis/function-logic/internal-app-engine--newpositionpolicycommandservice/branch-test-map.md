# Branch Test Map: `NewPositionPolicyCommandService`

Post-edit, from `ast.json` (90–104, 2 branches).

| Branch | Line | Test | Kind |
|---|---|---|---|
| B1 | 91 | `TestPositionPolicyCommandServiceRequiresEngineOwnedJournal` (nil and empty Context, existing) | behavioural |
| B2 | 94 | existing service tests pass `clock.System()`; the nil-clock default is a declared fallback | declared |
| field `audit` | 102 | set: `TestA066EntryLockReleaseThroughTheEngineEndpoint` (engine audit log written); nil: `TestA066RelaxationRefusedWithoutAnEngineAuditLog`, `TestA066LatchReleaseRefusedWithoutAnEngineAuditLog` | behavioural |
