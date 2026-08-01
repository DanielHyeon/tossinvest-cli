# Branch Test Map: `migrateAuditDigests`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | legacy audit read failure aborts | migration rollback coverage | v3 migration absent | PASS |
| B2 | every audit row is structurally scanned before writes | v2 multi-row migration test | audit digest absent | PASS |
| B3 | corrupt legacy structure aborts | `TestMigrationRejectsCorruptLegacyAuditRow` | migration could re-sign invalid row | PASS |
| B4 | row close failure aborts | DB fault review | v3 migration absent | defensive branch reviewed |
| B5 | row iteration failure aborts | DB fault review | v3 migration absent | defensive branch reviewed |
| B6 | every corroborated event receives digest | `TestMigrationFromV2BindsControlPointerAndLegacyAuditRows` | audit digest absent | PASS |
| B7 | unmatched snapshot/candidate event aborts | `TestMigrationRejectsCorruptLegacyAuditRow` | migration could re-sign plausible tamper | PASS |
| B8 | any digest update failure rolls back all | transaction fault review | v3 migration absent | defensive branch reviewed |
