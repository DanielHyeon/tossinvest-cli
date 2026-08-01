# Branch Test Map: `Store.snapshot`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | query/scan failure | existing invalid-history cases | existing | existing |
| B2 | malformed JSON maps | corruption test | yes | yes |
| B3 | invalid row invariants | corruption test | yes | yes |
| B4 | invalid stored timestamp | `TestAuditTimestampCorruptionFailsClosed` | yes | yes |
| B5 | invalid settings digest | `TestSnapshotAndAuditCorruptionFailClosed` | yes | yes |
