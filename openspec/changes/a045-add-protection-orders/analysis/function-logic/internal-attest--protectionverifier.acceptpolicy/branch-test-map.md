# Branch Test Map: `ProtectionVerifier.acceptPolicy`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | rollback or digest reuse | `TestProtectionVerifierRejectsRollbackAcrossDirectVerifyCalls` | not latched across calls | typed refusal |
| B2 | first/newer canonical policy | signed happy path and generation rotation | no observation state | accepted and latched |
