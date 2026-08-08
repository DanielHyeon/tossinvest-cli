# Branch Test Map: `TestEachInterlockClauseHasALine`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | all clause rows execute | `TestEachInterlockClauseHasALine` | baseline | pass |
| B2 | custom gate writer executes | same | baseline | pass |
| B3 | default gate writer executes | same | baseline | pass |
| B4 | ordinary rows receive matched Guardian | same | baseline | pass |
| B5 | no-Guardian row is identified | same | auto-construction masked refusal | pass |
| B6 | other rows use normal assembly | same | baseline | pass |
| B7 | successful assembly fails test | same | baseline | pass |
| B8 | expected operator line is present | same | baseline | pass |
