# Branch Test Map: `Transition`

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
| B9 | AST branch B9; fail-closed scenario specified in function logic map | focused adversarial table test for B9 | covered by package adversarial tests | pass |
| B10 | AST branch B10; fail-closed scenario specified in function logic map | focused adversarial table test for B10 | covered by package adversarial tests | pass |
| B11 | AST branch B11; fail-closed scenario specified in function logic map | focused adversarial table test for B11 | covered by package adversarial tests | pass |
| B12 | AST branch B12; fail-closed scenario specified in function logic map | focused adversarial table test for B12 | covered by package adversarial tests | pass |
| B13 | AST branch B13; fail-closed scenario specified in function logic map | focused adversarial table test for B13 | covered by package adversarial tests | pass |
| B14 | AST branch B14; fail-closed scenario specified in function logic map | focused adversarial table test for B14 | covered by package adversarial tests | pass |
| B15 | AST branch B15; fail-closed scenario specified in function logic map | focused adversarial table test for B15 | covered by package adversarial tests | pass |
| B16 | AST branch B16; fail-closed scenario specified in function logic map | focused adversarial table test for B16 | covered by package adversarial tests | pass |
| B17 | AST branch B17; fail-closed scenario specified in function logic map | focused adversarial table test for B17 | covered by package adversarial tests | pass |
| B18 | AST branch B18; fail-closed scenario specified in function logic map | focused adversarial table test for B18 | covered by package adversarial tests | pass |
| B19 | AST branch B19; fail-closed scenario specified in function logic map | focused adversarial table test for B19 | covered by package adversarial tests | pass |
| B20 | AST branch B20; fail-closed scenario specified in function logic map | focused adversarial table test for B20 | covered by package adversarial tests | pass |
| B21 | AST branch B21; fail-closed scenario specified in function logic map | focused adversarial table test for B21 | covered by package adversarial tests | pass |
| B22 | AST branch B22; fail-closed scenario specified in function logic map | focused adversarial table test for B22 | covered by package adversarial tests | pass |
| B23 | AST branch B23; fail-closed scenario specified in function logic map | focused adversarial table test for B23 | covered by package adversarial tests | pass |
| B24 | generated output fails state-specific validation | `TestTransitionValidatesItsOutput` and transition suite | yes: trigger during REPLACING returned an invalid TRIGGERED saga | pass |
| B1+ | Input/time/event guards and monotonic replace are followed by output invariant validation. | every allowed transition, forbidden transition, state-specific stale fields, output validation, replace/trigger crash windows. | yes: invalid output was returned | pass |
| B25 | registration/replace result carries a different attempt ID | lineage table | mismatch not checked | pass after remediation |
| B26 | trigger/close observation carries a different broker ID | lineage table | mismatch ignored | pass after remediation |
