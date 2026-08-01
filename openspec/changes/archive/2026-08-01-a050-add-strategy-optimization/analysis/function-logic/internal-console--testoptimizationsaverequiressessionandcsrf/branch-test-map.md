# Branch Test Map: `TestOptimizationSaveRequiresSessionAndCSRF`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | missing CSRF is refused | self | lifecycle route initially absent | yes |
| B2 | refused request reached no seam | self | lifecycle route initially absent | yes |
| B3 | valid CSRF previews once and never legacy-saves | self | legacy handler saved directly | yes |
