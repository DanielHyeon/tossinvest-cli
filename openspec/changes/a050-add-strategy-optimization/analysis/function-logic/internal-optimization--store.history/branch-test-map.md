# Branch Test Map: `Store.history`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | query failure | DB error tests | n/a | n/a |
| B2 | corrupt history row | `TestSnapshotAndAuditCorruptionFailClosed` | pending | pending |
| B3 | bounded newest-first history | `TestRollbackCreatesANewVersionAndNeverRewritesHistory` | existing | existing |
