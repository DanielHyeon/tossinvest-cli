# Branch Test Map: `PendingCleanup`

- Source: `internal/verifylive/cleanup.go:119-121`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `happy` at `internal/verifylive/cleanup.go:119` — `branchless` | `TestM0TriggeredWithoutChildIDIsManualReconcileOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
