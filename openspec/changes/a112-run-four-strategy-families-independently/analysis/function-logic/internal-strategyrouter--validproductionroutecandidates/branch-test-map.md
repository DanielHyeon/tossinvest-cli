# Branch Test Map: `validProductionRouteCandidates`

- Source: `internal/strategyrouter/production.go`; file SHA-256 `1175f67d72d78cc9f3ef65d505d97112382de26ea1eae89165314529dafb26d9`. AST branch positions are authoritative.
- Rows carry measured counts from Go coverage profiles, count mode.
- untagged package suite: `go test -count=1 -covermode=count ./internal/strategyrouter/`
- tagged package suite: `go test -count=1 -tags tossos_testseams -covermode=count ./internal/strategyrouter/`

Mutation receipts for this function (production source mutated, run, restored from a pristine copy taken before the mutation):

| # | mutation | result | killed by |
|---|---|---|---|
| M6' | delete `descriptor.Family != value.Family` | KILLED | `TestProductionRouteCandidatesRejectFamilyDriftAndPartialFamilyCoverage` |
| M7 | delete the redundant `len(families) == len(want)` tail clause | **SURVIVED** — removed as dead code instead of kept | no test; no input can make it false |
| M8' | delete `value.ScorePPM > productionRouteScorePPMMax` | KILLED | `TestProductionRouteCandidatesRejectAScorePPMAboveTheApprovedRange` |

| Branch | Anchor | Measured disposition |
|---|---|---|
| B1 | if at 534:2 | arm entered 2x (untagged package suite); arm entered 2x (tagged package suite); entered by `TestProductionRouteCandidatesRejectLegacyThreeFamilyAndPartialSets` |
| B2 | range at 538:2 | arm entered 232x (untagged package suite); arm entered 232x (tagged package suite); entered by `TestPairedProductionRouteAuthorityLoadsExactFourLanesIndependently`, `TestProductionRouteAuthorityBatchUsesEverySignedScopeInOneMarketSnapshot`, `TestProductionRouteAuthorityCarriesThreeIndependentSeals`, `TestProductionRouteAuthorityFailureIsMarketLocal`, `TestProductionRouteAuthorityRestoresExactActiveOwner`, `TestProductionRouteAuthoritySelectsEverySignedSymbolScope`, `TestProductionRouteCandidatesCarryNoRawArbitrationScore`, `TestProductionRouteCandidatesRejectAScorePPMAboveTheApprovedRange`, `TestProductionRouteCandidatesRejectFamilyDriftAndPartialFamilyCoverage`, `TestProductionRouteCandidatesRejectLegacyThreeFamilyAndPartialSets` |
| B3 | if at 546:3 | arm not entered (untagged package suite); arm not entered (tagged package suite); no per-test profile in the attribution set entered it |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
