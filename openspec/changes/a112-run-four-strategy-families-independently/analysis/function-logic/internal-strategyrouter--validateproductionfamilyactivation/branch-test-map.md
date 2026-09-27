# Branch Test Map: `validateProductionFamilyActivation`

- Source: `internal/strategyrouter/production_family_activation.go` (471-532); HEAD blob SHA-256 `411d612f7444c31744585b53c6128ccb71d0406a3e708efe9b7aee6232ca67d7`. AST branch positions are authoritative.
- Measurement regime: 몸통 진입 count, strategyrouter 태그 스위트 per-test 67 (MISMATCH 0).

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 474:2 | arm entered 12x (strategyrouter tagged suite, pre-edit); `TestAnActivationWhoseBindingsDoNotMatchPromotesNothing`, `TestTheOtherMarketsWholeManifestInThisMarketsFilePromotesNothing` |
| B2 | if | 488:2 | arm entered 4x (strategyrouter tagged suite, pre-edit); `TestAnActivationOutsideItsApprovedLifetimePromotesNothing` |
| B3 | if | 492:2 | arm entered 1x (strategyrouter tagged suite, pre-edit); `TestAnActivationOutsideItsApprovedLifetimePromotesNothing` |
| B4 | range | 510:2 | arm entered 41x (strategyrouter tagged suite, pre-edit); `TestAVerifiedActivationCarriesTheProtectionFloorTheOrderPathMustBind`, `TestAVerifiedFourFamilyActivationPromotesExactlyTheLanesItNames`, `TestAnActivationPromotesOnlyTheFamiliesItTurnsOn`, `TestAnActivationWhoseDescriptorSetIsNotExactlyTheFourPromotesNothing`, `TestTheCommittedGoldenManifestMatchesItsPinAndPromotesTheFourLanes` |
| B5 | if | 512:3 | arm entered 5x (strategyrouter tagged suite, pre-edit); `TestAnActivationWhoseDescriptorSetIsNotExactlyTheFourPromotesNothing` |
| B6 | if | 519:3 | arm entered 1x (strategyrouter tagged suite, pre-edit); `TestAnActivationWhoseDescriptorSetIsNotExactlyTheFourPromotesNothing` |
| B7 | if | 523:3 | arm entered 2x (strategyrouter tagged suite, pre-edit); `TestAnActivationWhoseDescriptorSetIsNotExactlyTheFourPromotesNothing` |
| B8 | if | 528:2 | arm entered 1x (strategyrouter tagged suite, pre-edit); `TestAnActivationWhoseDescriptorSetIsNotExactlyTheFourPromotesNothing` |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
