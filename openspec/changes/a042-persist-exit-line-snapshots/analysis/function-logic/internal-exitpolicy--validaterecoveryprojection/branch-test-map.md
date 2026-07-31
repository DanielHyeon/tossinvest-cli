# Branch Test Map: `validateRecoveryProjection`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | unknown/orderable suppression | `TestRecoveryRefusesSemanticallyInvalidLadderOutputs` | yes | yes |
| B2 | ratio zero, over one, malformed | same table | yes | yes |
| B3 | fractional or numerically wrong projected quantity | same table | yes | yes |
| B4 | state-only/orderable disagreement | same table | yes | yes |
| B5 | valid zero-share state-only and ladder hold | snapshot package contract tests | yes | yes |
| B6 | projection recomputation error/mismatch | semantic output table | yes | yes |
| B7 | orderable/state-only flag relation | semantic output table | yes | yes |
| B8 | nonorderable proposal fields absent | ordinary no-action recovery | yes | yes |
| B9 | ladder hold state-only suppression rule | ladder pending/state-only tests | yes | yes |
| B10 | ordinary nonorderable state-only refused | `none_state_only` case | yes | yes |
