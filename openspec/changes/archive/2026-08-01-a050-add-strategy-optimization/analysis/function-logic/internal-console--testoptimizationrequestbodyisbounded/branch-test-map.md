# Branch Test Map: `TestOptimizationRequestBodyIsBounded`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | request construction | this test | unbounded | yes |
| B2 | request transport | this test | unbounded | yes |
| B3 | 413 and no commander calls | this test | handler reached | yes |
