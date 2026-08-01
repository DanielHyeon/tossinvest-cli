# Branch Test Map: `digestControlPointer`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | happy path binds empty/versioned pointer and detects otherwise valid-looking rollback | `TestControlPointerRollbackTamperFailsReadPreviewAndApply`, `TestMigrationFromV2BindsControlPointerAndLegacyAuditRows` | control version was unauthenticated | PASS |
