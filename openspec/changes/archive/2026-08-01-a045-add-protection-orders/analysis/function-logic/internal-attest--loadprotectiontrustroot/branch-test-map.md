# Branch Test Map: `loadProtectionTrustRoot`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | absent/unsafe/replaced root | trust filesystem table | pass | pass regression |
| B2 | generation rollback or same-generation digest reuse | policy generation table | missing | pass after remediation |
| B3 | ACTIVE key becomes REVOKED after parse | parse-before-revoke | copied key bypasses | pass after remediation |
| B4 | decoded root violates version/generation/key contract | malformed root table | partial root checks | rejected |
| B5 | root generation differs from policy generation | generation mismatch test | generation absent | rejected |
