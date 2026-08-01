# Branch Test Map: `ProtectionTrustRoot.validate`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | zero/rollback generation | generation tests | absent | pass after remediation |
| B2 | duplicate or unsorted keys | malformed root table | duplicate pass, order absent | pass after remediation |
| B3 | valid ACTIVE/REVOKED lifecycle | lifecycle table | pass | pass regression |
| B4 | iterate bounded key set | rotation/lifecycle table | pass | pass regression |
| B5 | key role/algorithm/bytes/time/status invalid | malformed key table | partial checks | rejected |
| B6 | duplicate ID/public key or non-increasing key ID | reordered/duplicate root table | order not enforced | rejected |
