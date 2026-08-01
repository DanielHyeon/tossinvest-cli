# Branch Test Map: `BuildRegistry`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | all valid bindings/descriptors are copied | `TestRegistryRequiresExactlyOneMatchingOwnerForEveryField`, `TestRegistryReturnsCopiesInsteadOfMutableOwnerMetadata` | existing RED/GREEN baseline | PASS at `948e721` |
| B2 | nil provider fails closed | `TestRegistryRejectsNilEmptyAndFailingProviders/nil` | missing case at `948e721` | PASS |
| B3 | invalid/read-only category binding fails closed | `TestRegistryRejectsNilEmptyAndFailingProviders/invalid_category` | missing case at `948e721` | PASS |
| B4 | blank provider owner fails closed | `TestRegistryRejectsNilEmptyAndFailingProviders/blank_owner` | missing case at `948e721` | PASS |
| B5 | duplicate owner rejected | `TestRegistryRequiresExactlyOneMatchingOwnerForEveryField` | existing RED/GREEN baseline | PASS at `948e721` |
| B6 | provider descriptor error is rejected without partial output | `TestRegistryRejectsNilEmptyAndFailingProviders/provider_error` | missing case at `948e721` | PASS |
| B7 | empty descriptor set is rejected without partial output | `TestRegistryRejectsNilEmptyAndFailingProviders/empty_descriptors` | missing case at `948e721` | PASS |
| B8 | every returned descriptor is validated and copied | `TestRegistryReturnsCopiesInsteadOfMutableOwnerMetadata` | existing RED/GREEN baseline | PASS |
| B9 | missing descriptor key fails closed | `TestRegistryRejectsMissingKeysOwnerMismatchAndDuplicateKeys/missing_key` | existing RED/GREEN baseline | PASS |
| B10 | owner mismatch fails closed | `TestRegistryRejectsMissingKeysOwnerMismatchAndDuplicateKeys/owner_mismatch` | existing RED/GREEN baseline | PASS |
| B11 | duplicate descriptor key fails closed | `TestRegistryRejectsMissingKeysOwnerMismatchAndDuplicateKeys/duplicate_key` | existing RED/GREEN baseline | PASS |
| B12 | known owner/key with incomplete metadata is visible only as configuration-error read-only field; preview writes zero candidate rows | `TestIncompleteKnownDescriptorIsExposedReadOnlyWithConfigurationError`, `TestPreviewRejectsConfigurationErrorFieldWithoutPersistingCandidate`, `TestOptimizationConfigurationErrorIsReadOnlyAndSuppressesPresetControls` | BuildRegistry rejected the entire registry at `948e721` | PASS |
