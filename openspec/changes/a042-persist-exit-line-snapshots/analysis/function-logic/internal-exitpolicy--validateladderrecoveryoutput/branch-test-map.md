# Branch Test Map: `validateLadderRecoveryOutput`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `ActiveRung=-1` valid | `TestRecoveryAllowsLadderBeforeFirstRung` | yes | yes |
| B2 | `ActiveRung=-2` or `999` invalid | `TestRecoveryRefusesInvalidLadderRungBounds` | yes | yes |
| B3 | foreign policy action | semantic output table `foreign_action` | yes | yes |
| B4 | level/rung mismatch | semantic output table `wrong_level` | yes | yes |
| B5 | orderable vs nonorderable level path | ladder snapshot/recovery suites | yes | yes |
| B6 | nonnumeric or unequal orderable level | semantic output `wrong_level` case | yes | yes |
