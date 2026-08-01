# Branch Test Map: `canonicalProtectionAccount`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 8 and 14 digits | account table | pass | pass regression |
| B2 | one or multiple hyphens and arbitrary characters | account table | single/multiple hyphens accepted | pass after remediation |
| B3 | short/long/padded | account table | pass | pass regression |
| B4 | any non-ASCII-digit byte, including hyphen | account table | separators normalized | rejected |
