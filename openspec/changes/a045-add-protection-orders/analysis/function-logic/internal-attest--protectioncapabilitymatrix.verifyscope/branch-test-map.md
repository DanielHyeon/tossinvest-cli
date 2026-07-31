# Branch Test Map: `ProtectionCapabilityMatrix.verifyScope`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | AST branch B1; fail-closed scenario specified in function logic map | focused adversarial table test for B1 | required in follow-up RED | pending implementation |
| B2 | AST branch B2; fail-closed scenario specified in function logic map | focused adversarial table test for B2 | required in follow-up RED | pending implementation |
| B3 | AST branch B3; fail-closed scenario specified in function logic map | focused adversarial table test for B3 | required in follow-up RED | pending implementation |
| B4 | AST branch B4; fail-closed scenario specified in function logic map | focused adversarial table test for B4 | required in follow-up RED | pending implementation |
| B5 | AST branch B5; fail-closed scenario specified in function logic map | focused adversarial table test for B5 | required in follow-up RED | pending implementation |
| B6 | AST branch B6; fail-closed scenario specified in function logic map | focused adversarial table test for B6 | required in follow-up RED | pending implementation |
| B7 | required verifier version/build differs from descriptor | exact tool build mismatch cases in scope suite | yes: new split verifier API was absent | pass |
| B1+ | malformed account/profile, no exact capability row, tool set not exactly two, or version/build mismatch fails closed. | every scope dimension mismatch, malformed account, tool version/build mismatch. | yes: new split verifier API was absent | pass |
