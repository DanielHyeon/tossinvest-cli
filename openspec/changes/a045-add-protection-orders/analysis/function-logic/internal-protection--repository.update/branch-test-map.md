# Branch Test Map: `Repository.Update`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | AST branch B1; fail-closed scenario specified in function logic map | focused adversarial table test for B1 | required in follow-up RED | pending implementation |
| B2 | AST branch B2; fail-closed scenario specified in function logic map | focused adversarial table test for B2 | required in follow-up RED | pending implementation |
| B3 | AST branch B3; fail-closed scenario specified in function logic map | focused adversarial table test for B3 | required in follow-up RED | pending implementation |
| B4 | AST branch B4; fail-closed scenario specified in function logic map | focused adversarial table test for B4 | required in follow-up RED | pending implementation |
| B5 | AST branch B5; fail-closed scenario specified in function logic map | focused adversarial table test for B5 | required in follow-up RED | pending implementation |
| B6 | stored row lookup fails | missing/stale repository path | yes: old update never loaded old row | pass |
| B7 | missing stored row is classified concurrent | stale update test | yes: old update only detected after write | pass |
| B8 | expected revision differs from stored revision | stale update test | yes: no pre-write stored comparison | pass |
| B9 | immutable identity or persisted transition guard fails | identity mutation/state jump test | yes: both were accepted before guard | pass |
| B10 | SQL update fails | repository error path | covered by fail-closed implementation | pass |
| B11 | rows-affected lookup fails | repository result error path | covered by fail-closed implementation | pass |
| B12 | transaction commit fails | repository commit error path | covered by fail-closed implementation | pass |
| B1+ | Validate, load current row in transaction, compare identity/transition/revision, CAS update, and commit. | identity mutation refusal, state jump refusal, valid adjacent update, stale revision. | yes: identity mutation and state jump were accepted | pass |
