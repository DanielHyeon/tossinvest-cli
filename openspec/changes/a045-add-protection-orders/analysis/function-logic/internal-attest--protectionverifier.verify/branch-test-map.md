# Branch Test Map: `ProtectionVerifier.Verify`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | parse fails or final authorization refuses | `TestProtectionVerifierIsSealedAndRechecksHardRevocation` plus signed malformed tables | raw caller-controlled verifier APIs existed | zero verifier and all invalid artifacts fail closed |
