# Branch Test Map: `ProtectionCapabilityMatrix.verifyScope`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | AST branch B1; fail-closed scenario specified in function logic map | focused adversarial table test for B1 | covered by package adversarial tests | pass |
| B2 | AST branch B2; fail-closed scenario specified in function logic map | focused adversarial table test for B2 | covered by package adversarial tests | pass |
| B3 | AST branch B3; fail-closed scenario specified in function logic map | focused adversarial table test for B3 | covered by package adversarial tests | pass |
| B4 | AST branch B4; fail-closed scenario specified in function logic map | focused adversarial table test for B4 | covered by package adversarial tests | pass |
| B5 | AST branch B5; fail-closed scenario specified in function logic map | focused adversarial table test for B5 | covered by package adversarial tests | pass |
| B6 | AST branch B6; fail-closed scenario specified in function logic map | focused adversarial table test for B6 | covered by package adversarial tests | pass |
| B7 | required verifier version/build differs from descriptor | exact tool build mismatch cases in scope suite | yes: new split verifier API was absent | pass |
| B1+ | malformed account/profile, no exact capability row, tool set not exactly two, or version/build mismatch fails closed. | every scope dimension mismatch, malformed account, tool version/build mismatch. | yes: new split verifier API was absent | pass |
