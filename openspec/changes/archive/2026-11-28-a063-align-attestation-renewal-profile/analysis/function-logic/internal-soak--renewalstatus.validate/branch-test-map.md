# Branch Test Map: `RenewalStatus.Validate`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | non-UTC attempt rejected | `TestRenewalStatusRejectsEmptyFailureReasonsAndNonUTCTimestamps` | n/a | yes |
| B2 | unknown field / bad schema rejected on load | `TestRenewalStatusRoundTripAndRejectsUnsafeInputs` | n/a | yes |
| B3 | schema bounds | `TestRenewalStatusRejectsFutureDuplicateAndOversizeSchemas` | n/a | yes |
| B4 | codes iterated | `TestRenewalStatusRejectsFutureDuplicateAndOversizeSchemas` | n/a | yes |
| B5 | duplicate code rejected | `TestRenewalStatusRejectsFutureDuplicateAndOversizeSchemas` | n/a | yes |
| B6 | issued + only `live_endpoint_uncovered` accepted; issued + other code rejected; issued without expiry rejected | `TestA063IssuedStatusMayCarryOnlyTheUncoveredScopeReason`, `TestRenewalStatusIssuedNeedsConsistentExpiry` | yes (`analysis/scope-absorb/red.txt`) | yes |
| B7 | failed without reasons rejected | `TestRenewalStatusRejectsEmptyFailureReasonsAndNonUTCTimestamps` | n/a | yes |
