# Branch Test Map: `decideFlatten`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | AST branch B1; fail-closed scenario specified in function logic map | focused adversarial table test for B1 | covered by package adversarial tests | pass |
| B2 | decision after deadline or before start/observations | decisionAt table | no decisionAt input | pass after remediation |
| B3 | stale observation before start or reversed cancel/sellable | observation order table | partial checks | pass after remediation |
| B4 | exact boundary returns opaque permit | happy path | only replayable enum returned | pass after remediation |
| B5 | copied permit, +1h, wrong scope/quantity | one-shot consume table | no sealed consumption | pass after remediation |
