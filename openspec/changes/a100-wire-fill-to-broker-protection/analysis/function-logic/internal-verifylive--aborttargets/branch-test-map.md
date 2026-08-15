# Branch Test Map: `AbortTargets`

- Source: `internal/verifylive/abort.go:65-67`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `happy` at `internal/verifylive/abort.go:65` — `branchless` | `TestM0TriggeredWithoutChildIDIsManualReconcileOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
