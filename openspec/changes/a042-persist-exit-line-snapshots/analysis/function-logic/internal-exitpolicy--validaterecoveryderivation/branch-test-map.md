# Branch Test Map: `ValidateRecoveryDerivation`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | malformed identity/definition | exact derivation and forgery tests | yes | yes |
| B2 | valid exact ratchet replay | persistence and `TestRecoveryRejectsForgedRatchetLevel` | yes | yes |
| B3 | valid exact ladder replay before first rung (`-1`) | `TestRecoveryAllowsLadderBeforeFirstRung` | yes | yes |
| B4 | exact evaluator rejects stored input | invalid input/policy tests | yes | yes |
| B5 | current protection differs | `TestRecoveryReevaluatesExactInputAndEveryExecutionField/current_protection` | yes | yes |
| B6 | remaining quantity differs despite same floor projection | `.../remaining_quantity_same_projection` | yes | yes |
| B7 | cancel-first bit differs | `.../cancel_pending_first` | yes | yes |
| B8 | changed bit differs | `.../changed` | yes | yes |
| B9 | ratchet level differs | `TestRecoveryRejectsForgedRatchetLevel` | yes | yes |
