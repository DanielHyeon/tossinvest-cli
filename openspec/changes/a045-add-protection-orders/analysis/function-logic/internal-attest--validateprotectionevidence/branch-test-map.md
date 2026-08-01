# Branch Test Map: `validateProtectionEvidence`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | AST branch B1; fail-closed scenario specified in function logic map | focused adversarial table test for B1 | covered by package adversarial tests | pass |
| B2 | AST branch B2; fail-closed scenario specified in function logic map | focused adversarial table test for B2 | covered by package adversarial tests | pass |
| B3 | AST branch B3; fail-closed scenario specified in function logic map | focused adversarial table test for B3 | covered by package adversarial tests | pass |
| B4 | AST branch B4; fail-closed scenario specified in function logic map | focused adversarial table test for B4 | covered by package adversarial tests | pass |
| B5 | AST branch B5; fail-closed scenario specified in function logic map | focused adversarial table test for B5 | covered by package adversarial tests | pass |
| B6 | a required verifier descriptor is absent | missing-tool case in `TestProtectionMatrixRejectsMissingSafetyClaimsAndMetadata` | yes: new descriptor binding fields were absent | pass |
| B1+ | unknown/duplicate tool, blank version, malformed digest, wrong basename, capability binding mismatch, or absent tool fails closed. | missing/duplicate/unknown tool, bad basename, bad digest/build, matrix binding mismatch, evidence bytes missing/mismatch. | yes: descriptor/source/binding fields did not exist | pass |
