# Branch Test Map: `scanSnapshotUnchecked`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | scan/schema mismatch propagates | DB scan coverage | strict structural helper absent | PASS |
| B2 | malformed JSON, non-boolean ints, invalid version relation, or blank actor/audit rejects snapshot | `TestSnapshotAndAuditCorruptionFailClosed` and metadata tamper cases | structural checks incomplete | PASS |
| B3 | malformed/noncanonical created time rejects snapshot | snapshot timestamp corruption case | canonical time check absent | PASS |
