# Branch Test Map: `normaliseAdoptionSettings`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 407 | `if value.IncludeSymbols == nil {` true/entered and complementary path | TestOptimizationKeepsDesiredAndUnavailableEffectiveAdoptionDistinct | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B2 | `if` line 410 | `if value.ExcludeSymbols == nil {` true/entered and complementary path | TestOptimizationKeepsDesiredAndUnavailableEffectiveAdoptionDistinct | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
