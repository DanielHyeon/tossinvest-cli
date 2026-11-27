# Branch Test Map: `validReasonCode`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | dispatch | `TestA063IssuedStatusMayCarryOnlyTheUncoveredScopeReason` | yes | yes |
| B2 | new closed code accepted; unknown code still rejected via `Validate` | `TestA063IssuedStatusMayCarryOnlyTheUncoveredScopeReason`, `TestRenewalStatusRejectsFutureDuplicateAndOversizeSchemas` | yes (`analysis/scope-absorb/red.txt`) | yes |
