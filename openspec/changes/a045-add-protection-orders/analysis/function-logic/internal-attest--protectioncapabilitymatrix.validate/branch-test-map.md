# Branch Test Map: `ProtectionCapabilityMatrix.validate`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | AST branch B1; fail-closed scenario specified in function logic map | focused adversarial table test for B1 | covered by package adversarial tests | pass |
| B2 | AST branch B2; fail-closed scenario specified in function logic map | focused adversarial table test for B2 | covered by package adversarial tests | pass |
| B3 | AST branch B3; fail-closed scenario specified in function logic map | focused adversarial table test for B3 | covered by package adversarial tests | pass |
| B4 | AST branch B4; fail-closed scenario specified in function logic map | focused adversarial table test for B4 | covered by package adversarial tests | pass |
| B5 | AST branch B5; fail-closed scenario specified in function logic map | focused adversarial table test for B5 | covered by package adversarial tests | pass |
| B6 | AST branch B6; fail-closed scenario specified in function logic map | focused adversarial table test for B6 | covered by package adversarial tests | pass |
| B7 | AST branch B7; fail-closed scenario specified in function logic map | focused adversarial table test for B7 | covered by package adversarial tests | pass |
| B8 | AST branch B8; fail-closed scenario specified in function logic map | focused adversarial table test for B8 | covered by package adversarial tests | pass |
| B9 | canonical matrix bytes do not match the declared digest | `TestProtectionMatrixDigestBindsCanonicalMatrix` | yes: capability mutation remained accepted | pass |
| B10 | declared digest differs from canonical semantic matrix | digest-binding and reordered-row tables | aliases could claim distinct signed forms | rejected |
| B1+ | version/window/exact-UTC/evidence/rows/sorting/canonical marshal/digest aggregate | matrix adversarial tables | incomplete semantic canonicalization | pass |
