# Branch Test Map: `optimizationFieldViews`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | for each complete field, project default/desired/effective/provenance and owner options; for an incomplete descriptor, carry its configuration error so the template suppresses all mutation controls | `TestOptimizationShowsExactlySixCategoriesAndThreeOwnerPolicies`; `TestOptimizationDOMHasNoArbitraryInputAndSubmitsOnlyOwnerOptions`; `TestOptimizationConfigurationErrorIsReadOnlyAndSuppressesPresetControls` | configuration-error UI absent before this edit | focused tests and `go test ./internal/console` pass |

Zero fields skip B1 and return an empty slice; `TestOptimizationLoadingModelIsServerBlockingAndReadFailureIsFailClosed` covers the resulting no-control page.
