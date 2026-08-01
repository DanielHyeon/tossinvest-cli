# Branch Test Map: `optimizationPage.ExitPolicyWritable`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | lifecycle read failed or seam absent | `TestOptimizationLoadingModelIsServerBlockingAndReadFailureIsFailClosed` | existing coverage | pass |
| B2 | fields searched without inventing a target/draft | `TestOptimizationConfigurationErrorIsReadOnlyAndSuppressesPresetControls` | no error UI | pass |
| B3 | valid target enables exactly three owner preset previews; invalid target disables all | `TestOptimizationUsesOnePresetPreviewFlowWithoutClientDraft`; `TestOptimizationConfigurationErrorIsReadOnlyAndSuppressesPresetControls` | no descriptor-error gate | pass |
