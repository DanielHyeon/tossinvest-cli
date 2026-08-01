# Branch Test Map: `Store.audit`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | full-ledger coverage validation fails before page query | `TestReadRejectsDeletedAuditRowsIncludingPartialMultiChangeDeletion` | deleted old rows could be hidden | PASS |
| B2 | bounded page query failure returns | DB error coverage | baseline | defensive branch reviewed |
| B3 | every returned row is inspected | apply/audit history tests | baseline | PASS |
| B4 | structural scan/timestamp error fails closed | `TestAuditTimestampCorruptionFailsClosed` | structural checks incomplete | PASS |
| B5 | empty/mismatched per-event digest fails closed | `TestAuditEventDigestRejectsValidLookingPersistedTampering` | valid-looking persisted edits passed | PASS |
