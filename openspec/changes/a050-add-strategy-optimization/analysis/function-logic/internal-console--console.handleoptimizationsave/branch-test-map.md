# Branch Test Map: `Console.handleOptimizationSave`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | unexpected/duplicate/action-invalid form refuses before commander | `TestOptimizationRejectsUnexpectedAndDuplicateFields` | `PostFormValue` silently ignored/selected values | yes |
| B2 | unwired lifecycle refuses legacy bypass | `TestOptimizationLifecycleUnwiredRefusesPOSTInsteadOfLegacyBypass` | legacy save existed | yes |
| B3 | dispatch exact action | lifecycle handler suite | action lifecycle absent | yes |
| B4 | apply action | capability apply tests | capability apply absent | yes |
| B5 | apply error maps explicitly | capability/error tests | mapping absent | yes |
| B6 | rollback-preview action | rollback lifecycle tests | rollback absent | yes |
| B7 | malformed rollback versions/category rejected | rollback validation tests | rollback absent | yes |
| B8 | rollback service error maps explicitly | rollback lifecycle tests | mapping absent | yes |
| B9 | default finite-option preview | `TestOptimizationPreviewRequiresSessionAndCSRFAndNeverUsesLegacySave` | preview absent | yes |
| B10 | malformed/invented option rejected | `TestOptimizationLifecycleRejectsAClientInventedPolicy` | arbitrary policy path existed | yes |
| B11 | preview service error maps explicitly | CAS/validation tests | mapping absent | yes |
| I1 | body over 4096 bytes is refused by middleware before handler/commander | `TestOptimizationRequestBodyIsBounded` | route had no body limit | yes |
