# Branch Test Map: `Console.handleOptimization`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | non-GET/HEAD refused | route method/CSRF suite | legacy handler admitted POST | yes |
| B2 | empty category selects overview | UI contract tests | category IA absent | yes |
| B3 | unknown category warns without action | `TestOptimizationUnknownCategoryFallsBackToOverviewWithoutMutation` | category IA absent | yes |
| B4 | compatibility reader exists | optimization compatibility tests | lifecycle view absent | yes |
| B5 | compatibility load fails visibly | fail-closed UI test | error projection absent | yes |
| B6 | compatibility load succeeds read-only | optimization view tests | legacy state hidden | yes |
| B7 | lifecycle commander exists | optimization lifecycle tests | commander absent | yes |
| B8 | lifecycle read fails | `TestOptimizationReadFailureIsFailClosed` | controls could not be gated | yes |
| B9 | lifecycle read succeeds | `TestOptimizationShowsExactlySixCategoriesAndThreeOwnerPolicies` | lifecycle view absent | yes |
| B10 | each registered owner option is rendered | exact policy tests | owner cards absent | yes |
| B11 | desired missing uses legacy display fallback | optimization view tests | fallback unspecified | yes |
| B12 | protection commander is present | `TestExitProtectionCurrentRowUsesOnlyOpaqueActionAndCheckbox` | protection rows absent | yes |
| B13 | protection list fails | protection error projection test | error could be hidden | yes |
| B14 | protection list succeeds | `TestExitProtectionCurrentRowUsesOnlyOpaqueActionAndCheckbox` | opaque action absent | yes |
