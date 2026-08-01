# Branch Test Map: `Compare`

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
| B15 | trigger mismatch emits scoped discrepancy | trigger mismatch table case | yes: old unscoped signature | pass |
| B16 | iterate broker protections for orphan classification | orphan/terminal table cases | yes: old unscoped signature | pass |
| B17 | terminal broker protection is ignored for orphan classification | completed broker case | yes: old unscoped signature | pass |
| B18 | unmatched broker identity emits scoped orphan | orphan table case | yes: old unscoped signature | pass |
| B19 | count non-terminal broker protections | duplicate table case | yes: old unscoped signature | pass |
| B20 | terminal records do not add active order count | completed broker case | yes: old unscoped signature | pass |
| B21 | multiple active broker orders with local ownership emit duplicate | duplicate table case | yes: old unscoped signature | pass |
| B1+ | Exact typed scope is checked before classification; duplicate local/broker identity fails closed; classification stays pure. | mixed account/profile/market/symbol, duplicate broker ID, exact scoped missing/orphan/duplicate/mismatch, terminal ignored. | yes: compile failure against old unscoped signature and fields | pass |
