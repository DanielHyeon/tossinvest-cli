# Branch Test Map: `digestSnapshot`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | happy path digest changes when any persisted immutable metadata field changes and verifies unchanged row | `TestSnapshotDigestCoversEveryPersistedImmutableMetadataField`, `TestSnapshotAndAuditCorruptionFailClosed` | prior digest omitted immutable metadata | PASS |
