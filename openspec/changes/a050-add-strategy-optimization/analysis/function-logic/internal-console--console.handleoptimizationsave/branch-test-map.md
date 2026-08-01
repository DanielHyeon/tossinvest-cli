# Branch Test Map: `Console.handleOptimizationSave`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | unwired lifecycle refuses legacy bypass | `TestOptimizationLifecycleUnwiredRefusesPOSTInsteadOfLegacyBypass` | legacy save existed | yes |
| B2 | dispatch action | lifecycle handler suite | action lifecycle absent | yes |
| B3 | apply action | capability apply tests | capability apply absent | yes |
| B4 | apply error maps explicitly | error mapping tests | mapping absent | yes |
| B5 | rollback preview action | rollback lifecycle tests | rollback absent | yes |
| B6 | malformed rollback versions/category rejected | rollback validation tests | rollback absent | yes |
| B7 | rollback preview service error maps explicitly | rollback lifecycle tests | mapping absent | yes |
| B8 | default finite-option preview | `TestOptimizationPreviewRequiresSessionAndCSRFAndNeverUsesLegacySave` | preview absent | yes |
| B9 | malformed/invented option rejected | `TestOptimizationLifecycleRejectsAClientInventedPolicy` | arbitrary policy path existed | yes |
| B10 | preview service error maps explicitly | CAS and validation tests | mapping absent | yes |
