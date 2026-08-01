# Branch Test Map: `Store.validateAuditCoverage`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | aggregate SQL/schema failure refuses audit view | store corruption coverage | full-ledger check absent | PASS |
| B2 | deleted application, whole audit group, or one row from multi-change audit is detected regardless of page limit | `TestReadRejectsDeletedAuditRowsIncludingPartialMultiChangeDeletion`, control/audit hardening tests | append-only offline deletion could be hidden | PASS |
