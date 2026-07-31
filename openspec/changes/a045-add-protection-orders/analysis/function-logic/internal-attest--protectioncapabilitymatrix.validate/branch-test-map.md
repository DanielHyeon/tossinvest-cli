# Branch Test Map: `ProtectionCapabilityMatrix.validate`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | AST branch B1; fail-closed scenario specified in function logic map | focused adversarial table test for B1 | required in follow-up RED | pending implementation |
| B2 | AST branch B2; fail-closed scenario specified in function logic map | focused adversarial table test for B2 | required in follow-up RED | pending implementation |
| B3 | AST branch B3; fail-closed scenario specified in function logic map | focused adversarial table test for B3 | required in follow-up RED | pending implementation |
| B4 | AST branch B4; fail-closed scenario specified in function logic map | focused adversarial table test for B4 | required in follow-up RED | pending implementation |
| B5 | AST branch B5; fail-closed scenario specified in function logic map | focused adversarial table test for B5 | required in follow-up RED | pending implementation |
| B6 | AST branch B6; fail-closed scenario specified in function logic map | focused adversarial table test for B6 | required in follow-up RED | pending implementation |
| B7 | AST branch B7; fail-closed scenario specified in function logic map | focused adversarial table test for B7 | required in follow-up RED | pending implementation |
| B8 | AST branch B8; fail-closed scenario specified in function logic map | focused adversarial table test for B8 | required in follow-up RED | pending implementation |
| B9 | canonical matrix bytes do not match the declared digest | `TestProtectionMatrixDigestBindsCanonicalMatrix` | yes: capability mutation remained accepted | pass |
| B1+ | B1 version mismatch; B2 invalid time window; B3 evidence metadata invalid; B4 no rows; B5-B7 row validation/dedup; B8 marshal; B9 digest binding. | matrix digest missing/mismatch; row mutation after digest; duplicate row; expiry/version. | yes: digest-binding test failed before implementation | pass |
