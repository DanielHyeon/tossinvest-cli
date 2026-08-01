# Branch Test Map: `RiskGuardian.LimitsDigest`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | branchless happy-path sentinel (AST has no conditional): hashes the constructor-frozen canonical limits JSON | direct strategy snapshot mismatch/success tests | missing binding | pass |
