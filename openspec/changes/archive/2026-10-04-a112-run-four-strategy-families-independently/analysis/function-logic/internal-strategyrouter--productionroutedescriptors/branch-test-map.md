# Branch Test Map: `productionRouteDescriptors`

- Source: `internal/strategyrouter/production.go`; file SHA-256 `617163a508030dea2768655db66006d52097f17be9c04bb7c9fc300d3adf8a21`. AST branch positions are authoritative.
- Rows carry measured counts from Go coverage profiles, count mode.
- untagged package suite: `go test -count=1 -covermode=count ./internal/strategyrouter/`
- tagged package suite: `go test -count=1 -tags tossos_testseams -covermode=count ./internal/strategyrouter/`

Mutation receipts for this function (production source mutated, run, restored from a pristine copy taken before the mutation):

| # | mutation | result | killed by |
|---|---|---|---|
| M17 | bind the KR breakout lane to `FamilyContinuation` in the table | KILLED | `TestProductionRouteDescriptorsCoverFourFamiliesPerMarket` and six others |

| Branch | Anchor | Measured disposition |
|---|---|---|
| B1 | if at 587:2 | arm entered 41x (untagged package suite); arm entered 41x (tagged package suite); entered by `TestPairedProductionRouteAuthorityLoadsExactFourLanesIndependently`, `TestProductionRouteAuthorityBatchUsesEverySignedScopeInOneMarketSnapshot`, `TestProductionRouteAuthorityCarriesThreeIndependentSeals`, `TestProductionRouteAuthorityRestoresExactActiveOwner`, `TestProductionRouteAuthoritySelectsEverySignedSymbolScope`, `TestProductionRouteCandidatesCarryNoRawArbitrationScore`, `TestProductionRouteCandidatesRejectAScorePPMAboveTheApprovedRange`, `TestProductionRouteCandidatesRejectFamilyDriftAndPartialFamilyCoverage`, `TestProductionRouteCandidatesRejectLegacyThreeFamilyAndPartialSets`, `TestProductionRouteDescriptorsCoverFourFamiliesPerMarket` |
| B2 | if at 595:2 | arm entered 38x (untagged package suite); arm entered 38x (tagged package suite); entered by `TestPairedProductionRouteAuthorityLoadsExactFourLanesIndependently`, `TestProductionRouteAuthorityBatchUsesEverySignedScopeInOneMarketSnapshot`, `TestProductionRouteAuthorityCarriesThreeIndependentSeals`, `TestProductionRouteAuthorityFailureIsMarketLocal`, `TestProductionRouteAuthoritySelectsEverySignedSymbolScope`, `TestProductionRouteCandidatesCarryNoRawArbitrationScore`, `TestProductionRouteCandidatesRejectAScorePPMAboveTheApprovedRange`, `TestProductionRouteCandidatesRejectFamilyDriftAndPartialFamilyCoverage`, `TestProductionRouteCandidatesRejectLegacyThreeFamilyAndPartialSets`, `TestProductionRouteDescriptorsCoverFourFamiliesPerMarket` |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
