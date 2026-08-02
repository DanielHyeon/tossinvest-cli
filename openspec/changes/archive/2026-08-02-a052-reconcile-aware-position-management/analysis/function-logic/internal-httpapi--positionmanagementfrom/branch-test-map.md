# Branch Test Map: `positionManagementFrom`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 379 | `if actual.EffectiveKnown && actual.Effective != nil {` true/entered and complementary path | TestOptimizationKeepsDesiredAndUnavailableEffectiveAdoptionDistinct; TestOptimizationUsesCanonicalCategoryOrderAndOwnerDescriptors | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B2 | `range` line 385 | `for _, option := range value.StopOptions {` true/entered and complementary path | TestOptimizationKeepsDesiredAndUnavailableEffectiveAdoptionDistinct; TestOptimizationUsesCanonicalCategoryOrderAndOwnerDescriptors | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B3 | `if` line 388 | `if out.IncludeDefault == nil {` true/entered and complementary path | TestOptimizationKeepsDesiredAndUnavailableEffectiveAdoptionDistinct; TestOptimizationUsesCanonicalCategoryOrderAndOwnerDescriptors | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B4 | `if` line 391 | `if out.ExcludeDefault == nil {` true/entered and complementary path | TestOptimizationKeepsDesiredAndUnavailableEffectiveAdoptionDistinct; TestOptimizationUsesCanonicalCategoryOrderAndOwnerDescriptors | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
