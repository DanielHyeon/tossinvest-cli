# Branch Test Map: `Compare`

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
| B9 | AST branch B9; fail-closed scenario specified in function logic map | focused adversarial table test for B9 | required in follow-up RED | pending implementation |
| B10 | AST branch B10; fail-closed scenario specified in function logic map | focused adversarial table test for B10 | required in follow-up RED | pending implementation |
| B11 | AST branch B11; fail-closed scenario specified in function logic map | focused adversarial table test for B11 | required in follow-up RED | pending implementation |
| B12 | AST branch B12; fail-closed scenario specified in function logic map | focused adversarial table test for B12 | required in follow-up RED | pending implementation |
| B13 | AST branch B13; fail-closed scenario specified in function logic map | focused adversarial table test for B13 | required in follow-up RED | pending implementation |
| B14 | AST branch B14; fail-closed scenario specified in function logic map | focused adversarial table test for B14 | required in follow-up RED | pending implementation |
| B15 | trigger mismatch emits scoped discrepancy | trigger mismatch table case | yes: old unscoped signature | pass |
| B16 | iterate broker protections for orphan classification | orphan/terminal table cases | yes: old unscoped signature | pass |
| B17 | terminal broker protection is ignored for orphan classification | completed broker case | yes: old unscoped signature | pass |
| B18 | unmatched broker identity emits scoped orphan | orphan table case | yes: old unscoped signature | pass |
| B19 | count non-terminal broker protections | duplicate table case | yes: old unscoped signature | pass |
| B20 | terminal records do not add active order count | completed broker case | yes: old unscoped signature | pass |
| B21 | multiple active broker orders with local ownership emit duplicate | duplicate table case | yes: old unscoped signature | pass |
| B1+ | Exact typed scope is checked before classification; duplicate local/broker identity fails closed; classification stays pure. | mixed account/profile/market/symbol, duplicate broker ID, exact scoped missing/orphan/duplicate/mismatch, terminal ignored. | yes: compile failure against old unscoped signature and fields | pass |
