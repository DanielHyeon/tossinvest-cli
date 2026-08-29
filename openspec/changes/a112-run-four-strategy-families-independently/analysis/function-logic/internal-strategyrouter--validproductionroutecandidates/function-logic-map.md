# Function Logic Map: `validProductionRouteCandidates`

- Source: `internal/strategyrouter/production.go` (532-555)
- Function: `validProductionRouteCandidates` in package `strategyrouter`
- Signature: `validProductionRouteCandidates(params=2, results=1)`
- File SHA-256: `1175f67d72d78cc9f3ef65d505d97112382de26ea1eae89165314529dafb26d9`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 3.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Refuses a manifest whose candidate list is not exactly this market's table: right count, no unknown lane, no duplicate lane, matching horizon and lane version, **the family the table binds to that lane**, a `score_ppm` inside the approved 0..1,000,000 range, valid states and identity digests.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode.
- untagged package suite: `go test -count=1 -covermode=count ./internal/strategyrouter/`
- tagged package suite: `go test -count=1 -tags tossos_testseams -covermode=count ./internal/strategyrouter/`
- Measured entry: the function body was executed 70x (untagged package suite); executed 70x (tagged package suite).

Exact AST return positions: 535:3, 550:4, 554:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 534:2 | arm entered 2x (untagged package suite); arm entered 2x (tagged package suite); entered by `TestProductionRouteCandidatesRejectLegacyThreeFamilyAndPartialSets` |
| B2 | range | 538:2 | arm entered 232x (untagged package suite); arm entered 232x (tagged package suite); entered by `TestPairedProductionRouteAuthorityLoadsExactFourLanesIndependently`, `TestProductionRouteAuthorityBatchUsesEverySignedScopeInOneMarketSnapshot`, `TestProductionRouteAuthorityCarriesThreeIndependentSeals`, `TestProductionRouteAuthorityFailureIsMarketLocal`, `TestProductionRouteAuthorityRestoresExactActiveOwner`, `TestProductionRouteAuthoritySelectsEverySignedSymbolScope`, `TestProductionRouteCandidatesCarryNoRawArbitrationScore`, `TestProductionRouteCandidatesRejectAScorePPMAboveTheApprovedRange`, `TestProductionRouteCandidatesRejectFamilyDriftAndPartialFamilyCoverage`, `TestProductionRouteCandidatesRejectLegacyThreeFamilyAndPartialSets` |
| B3 | if | 546:3 | arm not entered (untagged package suite); arm not entered (tagged package suite); no per-test profile in the attribution set entered it |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `productionRouteDescriptors` | 533:10 |
| `len` | 534:5 |
| `len` | 534:20 |
| `make` | 537:10 |
| `len` | 537:32 |
| `validDesiredState` | 548:5 |
| `validDesiredState` | 548:42 |
| `productionRouteIdentity` | 549:5 |
| `productionRouteIdentity` | 549:55 |
| `len` | 554:9 |
| `len` | 554:22 |

## State mutations and fallbacks

- AST assignments: 4. Defers: 0. Goroutine statements: 0.

## Safety conclusion

B1 (`len(values) != len(want)`) is the arm that kills a legacy three-family manifest; the tail return is always true once the loop completes and is not a refusal arm. The family comparison inside B3 is what stops four lanes from claiming one family. A `len(families)` count here was measured to be unfalsifiable and was deleted rather than kept as an unkillable defence; the table's own four-distinct-families property is asserted by `TestProductionRouteDescriptorsCoverFourFamiliesPerMarket` instead.
