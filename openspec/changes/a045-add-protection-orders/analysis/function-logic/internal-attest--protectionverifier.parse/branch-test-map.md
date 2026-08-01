# Branch Test Map: `ProtectionVerifier.parse`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | missing sealed configuration | zero verifier test | configurable raw path API | `ErrProtectionTrust` |
| B2 | owner metadata unsupported | Windows cross-compile and platform helper contract | Unix-only owner call | non-Unix build fails closed at runtime |
| B3 | policy file/path/permissions/canonical JSON invalid | policy malformed and parent TOCTOU tables | policy source absent | typed refusal |
| B4 | signed envelope/root is malformed, unsafe or unverifiable | signed adversarial tables | raw parse path existed | typed refusal without latching unverified generation |
| B5 | lower generation or same-generation changed digest after full cryptographic validation | `TestProtectionVerifierRejectsRollbackAcrossDirectVerifyCalls` | whole-set rollback passed across calls | typed trust refusal |
