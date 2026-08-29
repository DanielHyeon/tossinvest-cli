# Branch Test Map: `LoadProductionRouteAuthorityBatch`

- Source: `internal/strategyrouter/production.go`; file SHA-256 `1175f67d72d78cc9f3ef65d505d97112382de26ea1eae89165314529dafb26d9`. AST branch positions are authoritative.
- Rows carry measured counts from Go coverage profiles, count mode.
- untagged package suite: `go test -count=1 -covermode=count ./internal/strategyrouter/`
- tagged package suite: `go test -count=1 -tags tossos_testseams -covermode=count ./internal/strategyrouter/`

Mutation receipts for this function (production source mutated, run, restored from a pristine copy taken before the mutation):

| # | mutation | result | killed by |
|---|---|---|---|
| M10 | restore a raw score with `Score: int64(value.ScorePPM)` | KILLED | `TestProductionRouteCandidatesCarryNoRawArbitrationScore` |
| M12 | make the scoring seal ignore `score_ppm` | KILLED | `TestProductionRouteAuthorityCarriesThreeIndependentSeals` |

| Branch | Anchor | Measured disposition |
|---|---|---|
| B1 | if at 319:2 | arm not entered (untagged package suite); arm not entered (tagged package suite); no per-test profile in the attribution set entered it |
| B2 | if at 322:2 | arm not entered (untagged package suite); arm not entered (tagged package suite); no per-test profile in the attribution set entered it |
| B3 | if at 329:2 | arm not entered (untagged package suite); arm not entered (tagged package suite); no per-test profile in the attribution set entered it |
| B4 | if at 337:2 | arm entered 1x (untagged package suite); arm entered 1x (tagged package suite); entered by `TestProductionRouteAuthorityFailureIsMarketLocal` |
| B5 | if at 341:2 | arm not entered (untagged package suite); arm not entered (tagged package suite); no per-test profile in the attribution set entered it |
| B6 | if at 347:2 | arm entered 8x (untagged package suite); arm entered 8x (tagged package suite); entered by `TestProductionRouteManifestRefusesAMissingCalibrationSeal`, `TestProductionRouteManifestRefusesTheLegacySchemaVersion` |
| B7 | if at 351:2 | arm not entered (untagged package suite); arm not entered (tagged package suite); no per-test profile in the attribution set entered it |
| B8 | if at 364:2 | arm not entered (untagged package suite); arm not entered (tagged package suite); no per-test profile in the attribution set entered it |
| B9 | range at 368:2 | arm entered 19x (untagged package suite); arm entered 19x (tagged package suite); entered by `TestPairedProductionRouteAuthorityLoadsExactFourLanesIndependently`, `TestProductionRouteAuthorityBatchUsesEverySignedScopeInOneMarketSnapshot`, `TestProductionRouteAuthorityCarriesThreeIndependentSeals`, `TestProductionRouteAuthorityFailureIsMarketLocal`, `TestProductionRouteAuthorityRestoresExactActiveOwner`, `TestProductionRouteAuthoritySelectsEverySignedSymbolScope`, `TestProductionRouteCandidatesCarryNoRawArbitrationScore` |
| B10 | if at 369:3 | arm not entered (untagged package suite); arm not entered (tagged package suite); no per-test profile in the attribution set entered it |
| B11 | if at 373:3 | arm entered 2x (untagged package suite); arm entered 2x (tagged package suite); entered by `TestProductionRouteAuthorityBatchUsesEverySignedScopeInOneMarketSnapshot` |
| B12 | if at 377:3 | arm not entered (untagged package suite); arm not entered (tagged package suite); no per-test profile in the attribution set entered it |
| B13 | if at 381:3 | arm entered 1x (untagged package suite); arm entered 1x (tagged package suite); entered by `TestProductionRouteAuthorityRestoresExactActiveOwner` |
| B14 | if at 385:3 | arm not entered (untagged package suite); arm not entered (tagged package suite); no per-test profile in the attribution set entered it |
| B15 | range at 389:3 | arm entered 64x (untagged package suite); arm entered 64x (tagged package suite); entered by `TestPairedProductionRouteAuthorityLoadsExactFourLanesIndependently`, `TestProductionRouteAuthorityBatchUsesEverySignedScopeInOneMarketSnapshot`, `TestProductionRouteAuthorityCarriesThreeIndependentSeals`, `TestProductionRouteAuthorityFailureIsMarketLocal`, `TestProductionRouteAuthorityRestoresExactActiveOwner`, `TestProductionRouteAuthoritySelectsEverySignedSymbolScope`, `TestProductionRouteCandidatesCarryNoRawArbitrationScore` |
| B16 | if at 407:2 | arm not entered (untagged package suite); arm not entered (tagged package suite); no per-test profile in the attribution set entered it |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
