# Branch Test Map: `installAppendOnlyTriggers`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | every fixed immutable-table trigger is installed | `TestSchemaIsVersionedAndAppendOnly` | triggers absent before hardening | PASS |
| B2 | DDL failure aborts migration with trigger identity | migration rollback/integrity tests | triggers absent before hardening | PASS |
