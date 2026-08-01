# Branch Test Map: `migrateSnapshotDigests`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | legacy snapshot query error aborts migration | migration rollback coverage | v2 migration absent | PASS |
| B2 | every legacy snapshot is buffered for atomic rewrite | legacy multi-snapshot migration test | v2 migration absent | PASS |
| B3 | corrupt legacy structure aborts without partial update | snapshot corruption/migration rollback test | v2 migration absent | PASS |
| B4 | row close error aborts | DB error-order review | v2 migration absent | defensive branch reviewed |
| B5 | iterator error aborts | DB error-order review | v2 migration absent | defensive branch reviewed |
| B6 | every validated snapshot receives full v2 digest | migration digest test | metadata fields not fully bound | PASS |
| B7 | legacy digest must match the frozen v1 algorithm before it is replaced | legacy digest tamper migration test | corrupt legacy metadata could be re-signed during migration | PASS |
| B8 | any v2 digest update failure rolls back all digest changes | migration rollback test | v2 migration absent | PASS |
