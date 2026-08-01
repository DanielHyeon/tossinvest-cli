# Branch Test Map: `Saga.Validate`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | AST branch B1; fail-closed scenario specified in function logic map | focused adversarial table test for B1 | covered by package adversarial tests | pass |
| B2 | AST branch B2; fail-closed scenario specified in function logic map | focused adversarial table test for B2 | covered by package adversarial tests | pass |
| B3 | AST branch B3; fail-closed scenario specified in function logic map | focused adversarial table test for B3 | covered by package adversarial tests | pass |
| B4 | AST branch B4; fail-closed scenario specified in function logic map | focused adversarial table test for B4 | covered by package adversarial tests | pass |
| B5 | AST branch B5; fail-closed scenario specified in function logic map | focused adversarial table test for B5 | covered by package adversarial tests | pass |
| B6 | AST branch B6; fail-closed scenario specified in function logic map | focused adversarial table test for B6 | covered by package adversarial tests | pass |
| B7 | AST branch B7; fail-closed scenario specified in function logic map | focused adversarial table test for B7 | covered by package adversarial tests | pass |
| B8 | state-specific field invariant rejects stale/missing mutation fields | `TestSagaValidateEnforcesStateSpecificFields` | yes: invalid combinations were accepted | pass |
| B1+ | Common identity/instrument/numeric/client/time/state checks are followed by state-specific field invariants. | table covering required and forbidden fields per state, revision and typed market/account scope. | yes: invalid combinations were accepted | pass |
