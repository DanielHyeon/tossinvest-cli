# Branch Test Map: `activeEntryLossLock`

Post-edit, from `ast.json` (121–141, 3 branches).

| Branch | Line | Test | Kind |
|---|---|---|---|
| B1 | 129 | `TestA066EntryLossLockReleaseIsOperatorApprovedAuditedAndOpensEntry` | behavioural |
| B2 | 132 | `TestRefuseEntryUnderLossLockFailsClosedOnUnknownScopeAndReadError`, `TestA066StorageErrorExitsFailClosed` | behavioural + structural |
| B3 | 136 | `TestA066StorageErrorExitsFailClosed` | structural |
| found | 140 | `TestA066EntryLossLockReleaseRefusals` (lock still refuses after every refused release), `TestMigrationV34ToV35KeepsExistingLocksInForceAndSwapsTheTrigger` | behavioural |
