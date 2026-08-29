# Function Logic Map: `LoadProductionRouteAuthorityBatch`

- Source: `internal/strategyrouter/production.go` (318-411)
- Function: `LoadProductionRouteAuthorityBatch` in package `strategyrouter`
- Signature: `LoadProductionRouteAuthorityBatch(params=3, results=2)`
- File SHA-256: `1175f67d72d78cc9f3ef65d505d97112382de26ea1eae89165314529dafb26d9`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 16.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Verifies one signed market manifest and reconstructs every requested symbol inside one read-only SQLite transaction. It never creates, migrates or writes either source. It now also derives the three seals (family, scoring, calibration) per symbol scope and deliberately leaves `Candidate.Score` unset, so no raw arbitration score reaches any consumer.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode.
- untagged package suite: `go test -count=1 -covermode=count ./internal/strategyrouter/`
- tagged package suite: `go test -count=1 -tags tossos_testseams -covermode=count ./internal/strategyrouter/`
- Measured entry: the function body was executed 25x (untagged package suite); executed 25x (tagged package suite).

Exact AST return positions: 320:3, 323:3, 334:3, 338:3, 342:3, 348:3, 352:3, 365:3, 370:4, 378:4, 382:4, 386:4, 408:3, 410:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 319:2 | arm not entered (untagged package suite); arm not entered (tagged package suite); no per-test profile in the attribution set entered it |
| B2 | if | 322:2 | arm not entered (untagged package suite); arm not entered (tagged package suite); no per-test profile in the attribution set entered it |
| B3 | if | 329:2 | arm not entered (untagged package suite); arm not entered (tagged package suite); no per-test profile in the attribution set entered it |
| B4 | if | 337:2 | arm entered 1x (untagged package suite); arm entered 1x (tagged package suite); entered by `TestProductionRouteAuthorityFailureIsMarketLocal` |
| B5 | if | 341:2 | arm not entered (untagged package suite); arm not entered (tagged package suite); no per-test profile in the attribution set entered it |
| B6 | if | 347:2 | arm entered 8x (untagged package suite); arm entered 8x (tagged package suite); entered by `TestProductionRouteManifestRefusesAMissingCalibrationSeal`, `TestProductionRouteManifestRefusesTheLegacySchemaVersion` |
| B7 | if | 351:2 | arm not entered (untagged package suite); arm not entered (tagged package suite); no per-test profile in the attribution set entered it |
| B8 | if | 364:2 | arm not entered (untagged package suite); arm not entered (tagged package suite); no per-test profile in the attribution set entered it |
| B9 | range | 368:2 | arm entered 19x (untagged package suite); arm entered 19x (tagged package suite); entered by `TestPairedProductionRouteAuthorityLoadsExactFourLanesIndependently`, `TestProductionRouteAuthorityBatchUsesEverySignedScopeInOneMarketSnapshot`, `TestProductionRouteAuthorityCarriesThreeIndependentSeals`, `TestProductionRouteAuthorityFailureIsMarketLocal`, `TestProductionRouteAuthorityRestoresExactActiveOwner`, `TestProductionRouteAuthoritySelectsEverySignedSymbolScope`, `TestProductionRouteCandidatesCarryNoRawArbitrationScore` |
| B10 | if | 369:3 | arm not entered (untagged package suite); arm not entered (tagged package suite); no per-test profile in the attribution set entered it |
| B11 | if | 373:3 | arm entered 2x (untagged package suite); arm entered 2x (tagged package suite); entered by `TestProductionRouteAuthorityBatchUsesEverySignedScopeInOneMarketSnapshot` |
| B12 | if | 377:3 | arm not entered (untagged package suite); arm not entered (tagged package suite); no per-test profile in the attribution set entered it |
| B13 | if | 381:3 | arm entered 1x (untagged package suite); arm entered 1x (tagged package suite); entered by `TestProductionRouteAuthorityRestoresExactActiveOwner` |
| B14 | if | 385:3 | arm not entered (untagged package suite); arm not entered (tagged package suite); no per-test profile in the attribution set entered it |
| B15 | range | 389:3 | arm entered 64x (untagged package suite); arm entered 64x (tagged package suite); entered by `TestPairedProductionRouteAuthorityLoadsExactFourLanesIndependently`, `TestProductionRouteAuthorityBatchUsesEverySignedScopeInOneMarketSnapshot`, `TestProductionRouteAuthorityCarriesThreeIndependentSeals`, `TestProductionRouteAuthorityFailureIsMarketLocal`, `TestProductionRouteAuthorityRestoresExactActiveOwner`, `TestProductionRouteAuthoritySelectsEverySignedSymbolScope`, `TestProductionRouteCandidatesCarryNoRawArbitrationScore` |
| B16 | if | 407:2 | arm not entered (untagged package suite); arm not entered (tagged package suite); no per-test profile in the attribution set entered it |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `config.ObservedAt.IsZero` | 319:19 |
| `ctx.Err` | 322:12 |
| `canonicalProductionRouteConfig` | 325:11 |
| `canonicalProductionRouteTargets` | 326:17 |
| `productionRouteOwnerUID` | 327:23 |
| `ProductionRouteFileName` | 328:10 |
| `filepath.IsAbs` | 329:32 |
| `filepath.IsAbs` | 329:69 |
| `validProductionRouteBaseConfig` | 330:11 |
| `productionRouteDigestValid` | 330:54 |
| `productionRouteIdentity` | 331:4 |
| `len` | 331:52 |
| `productionRouteIdentity` | 332:4 |
| `productionRouteIdentity` | 332:57 |
| `productionRouteIdentity` | 333:4 |
| `productionRouteIdentity` | 333:55 |
| `readProductionRouteFile` | 336:15 |
| `filepath.Join` | 336:39 |
| `productionRouteDigest` | 337:19 |
| `decodeProductionRouteManifest` | 340:19 |
| `len` | 341:19 |
| `verifyProductionRouteManifest` | 347:20 |
| `openProductionRouteSnapshot` | 350:17 |
| `fmt.Errorf` | 352:43 |
| `db.Close` | 354:8 |
| `tx.Rollback` | 355:8 |
| `productionRouteTime` | 356:17 |
| `productionRouteTime` | 357:14 |
| `productionRouteTime` | 358:26 |
| `newMarketRecord` | 359:17 |
| `string` | 360:68 |
| `strconv.FormatUint` | 360:98 |
| `EvaluateMarketLifecycle` | 364:19 |
| `make` | 367:12 |
| `len` | 367:54 |
| `ctx.Err` | 369:13 |
| `validProductionRouteScopes` | 372:19 |
| `NewOwnerKey` | 376:15 |
| `loadProductionRouteOwnersFrom` | 380:41 |
| `fmt.Errorf` | 382:44 |
| `newOwnerSnapshot` | 384:20 |
| `make` | 388:17 |
| `len` | 388:38 |
| `append` | 392:17 |
| `productionRouteFamilyScores` | 399:13 |
| `productionRouteFamilySeal` | 402:18 |
| `productionRouteScoringSeal` | 403:18 |
| `productionRouteCalibrationSeal` | 404:18 |
| `tx.Commit` | 407:12 |
| `fmt.Errorf` | 408:43 |

## State mutations and fallbacks

- AST assignments: 29. Defers: 2. Goroutine statements: 0.

## Safety conclusion

The function holds no writer, signer or transport handle; the SQLite handle is opened `mode=ro` with `query_only(1)` and is rolled back on every path. A manifest or journal integrity failure refuses the whole market snapshot; a symbol merely absent from the signed scope set is skipped without refusing the market.
