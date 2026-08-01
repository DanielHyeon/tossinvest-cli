# Branch Test Map: `validateJudgementSnapshot`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | orderable snapshot with missing proposal/reason | `TestOrderableSnapshotMustArmItsExactProposal/missing` | yes | yes |
| B2 | unknown suppression reason | `.../unknown_suppression` | yes | yes |
| B3 | mismatched action or level | `.../different_action`, `.../different_level` | yes | yes |
| B4 | working order not cleared | journal + engine typed arm-suppression tests | yes | yes |
| B5 | saved-monotone candidate clears proposal after validation | whole coherent snapshot test | yes | yes |
| B6 | typed reason exact-match gate | working-order and unknown-reason tests | yes | yes |
| B7 | proposal plus suppression is contradictory | proposal coherence table | yes | yes |
| B8 | proposal action/level exact equality | adversarial action/level cases | yes | yes |
