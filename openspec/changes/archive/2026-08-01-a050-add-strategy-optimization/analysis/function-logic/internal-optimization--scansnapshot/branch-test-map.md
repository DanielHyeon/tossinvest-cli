# Branch Test Map: `scanSnapshot`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | scan failure | DB fault path | n/a | n/a |
| B2 | corrupt JSON/invariants | `TestSnapshotAndAuditCorruptionFailClosed` | yes | yes |
| B3 | corrupt timestamp | `TestAuditTimestampCorruptionFailsClosed` | yes | yes |
| B4 | corrupt digest | `TestSnapshotAndAuditCorruptionFailClosed` | yes | yes |
