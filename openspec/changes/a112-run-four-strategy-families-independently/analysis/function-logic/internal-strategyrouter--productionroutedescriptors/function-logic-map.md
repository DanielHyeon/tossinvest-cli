# Function Logic Map: `productionRouteDescriptors`

- Source: `internal/strategyrouter/production.go` (570-588)
- Function: `productionRouteDescriptors` in package `strategyrouter`
- Signature: `productionRouteDescriptors(params=1, results=1)`
- File SHA-256: `7d60a867a87576cab66d25591d3211b50f1d532efa29166bd71bb5b623cc09f4`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 2.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Returns the exact four-family lane table for one market, or nil for any other market. Each row now also names the family that lane belongs to, which makes the table the single place the family/lane binding is declared.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode.
- untagged package suite: `go test -count=1 -covermode=count ./internal/strategyrouter/`
- tagged package suite: `go test -count=1 -tags tossos_testseams -covermode=count ./internal/strategyrouter/`
- Measured entry: the function body was executed 80x (untagged package suite); executed 80x (tagged package suite).

Exact AST return positions: 572:3, 580:3, 587:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 571:2 | arm entered 41x (untagged package suite); arm entered 41x (tagged package suite); entered by `TestPairedProductionRouteAuthorityLoadsExactFourLanesIndependently`, `TestProductionRouteAuthorityBatchUsesEverySignedScopeInOneMarketSnapshot`, `TestProductionRouteAuthorityCarriesThreeIndependentSeals`, `TestProductionRouteAuthorityRestoresExactActiveOwner`, `TestProductionRouteAuthoritySelectsEverySignedSymbolScope`, `TestProductionRouteCandidatesCarryNoRawArbitrationScore`, `TestProductionRouteCandidatesRejectAScorePPMAboveTheApprovedRange`, `TestProductionRouteCandidatesRejectFamilyDriftAndPartialFamilyCoverage`, `TestProductionRouteCandidatesRejectLegacyThreeFamilyAndPartialSets`, `TestProductionRouteDescriptorsCoverFourFamiliesPerMarket` |
| B2 | if | 579:2 | arm entered 38x (untagged package suite); arm entered 38x (tagged package suite); entered by `TestPairedProductionRouteAuthorityLoadsExactFourLanesIndependently`, `TestProductionRouteAuthorityBatchUsesEverySignedScopeInOneMarketSnapshot`, `TestProductionRouteAuthorityCarriesThreeIndependentSeals`, `TestProductionRouteAuthorityFailureIsMarketLocal`, `TestProductionRouteAuthoritySelectsEverySignedSymbolScope`, `TestProductionRouteCandidatesCarryNoRawArbitrationScore`, `TestProductionRouteCandidatesRejectAScorePPMAboveTheApprovedRange`, `TestProductionRouteCandidatesRejectFamilyDriftAndPartialFamilyCoverage`, `TestProductionRouteCandidatesRejectLegacyThreeFamilyAndPartialSets`, `TestProductionRouteDescriptorsCoverFourFamiliesPerMarket` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| (no call expressions in this function) | — |

## State mutations and fallbacks

- AST assignments: 0. Defers: 0. Goroutine statements: 0.

## Safety conclusion

The table is the only declaration of the family/lane binding, so a mistake here would propagate silently into candidate validation. The mutation receipt shows the untagged table test fails on a duplicated family.

a112 6.1: 같은 파일의 다른 함수 편집으로 줄만 이동(본문 불변)
