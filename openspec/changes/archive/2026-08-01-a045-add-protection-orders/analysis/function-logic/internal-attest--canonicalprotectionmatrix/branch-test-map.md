# Branch Test Map: `canonicalProtectionMatrix`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | reversed evidence/capability arrays | canonical semantic form test | accepted as separately signed form | pass after remediation |
| B2 | non-UTC or alternate timestamp | timestamp table | nonzero offsets accepted | pass after remediation |
