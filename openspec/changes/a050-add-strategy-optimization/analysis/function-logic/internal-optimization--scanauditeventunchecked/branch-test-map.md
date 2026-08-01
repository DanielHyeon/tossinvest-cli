# Branch Test Map: `scanAuditEventUnchecked`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | row scan mismatch/error refuses event | migration/read fault coverage | strict helper absent | PASS |
| B2 | invalid time/IDs/identity/reason refuses current and legacy event | `TestMigrationRejectsCorruptLegacyAuditRow`, audit corruption tests | structural audit validation absent | PASS |
