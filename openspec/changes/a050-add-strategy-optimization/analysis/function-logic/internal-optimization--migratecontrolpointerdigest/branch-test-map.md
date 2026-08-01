# Branch Test Map: `migrateControlPointerDigest`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | unreadable legacy control aborts | migration rollback coverage | v3 migration absent | PASS |
| B2 | zero pointer binds empty digest; positive pointer validates snapshot | `TestMigrationFromV2BindsControlPointerAndLegacyAuditRows` | pointer unauthenticated | PASS |
| B3 | corrupt/missing referenced snapshot refuses migration | corrupt migration coverage | pointer unauthenticated | PASS |
| B4 | digest update failure rolls back | transaction fault review | v3 migration absent | defensive branch reviewed |
