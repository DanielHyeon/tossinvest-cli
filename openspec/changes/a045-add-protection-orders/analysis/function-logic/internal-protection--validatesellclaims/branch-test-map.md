# Branch Test Map: `ValidateSellClaims`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | AST branch B1; fail-closed scenario specified in function logic map | focused adversarial table test for B1 | covered by package adversarial tests | pass |
| B2 | AST branch B2; fail-closed scenario specified in function logic map | focused adversarial table test for B2 | covered by package adversarial tests | pass |
| B3 | each non-negative claim must fit the remaining holding before subtraction | `TestSellClaimsAreOverflowSafeAtInt64Boundary` | yes: overflowing addition wrapped and was accepted | pass |
| B1+ | Negative values fail first; each claim is then checked against remaining holding without addition. | MaxInt64 combinations, each individual >holding, exact full allocation, negatives. | yes: overflowing addition wrapped and was accepted | pass |
