# Function Logic Map: `validProductionRouteCandidates`

- Source: `internal/strategyrouter/production.go` (548-571)
- Function: `validProductionRouteCandidates` in package `strategyrouter`
- Signature: `validProductionRouteCandidates(params=2, results=1)`
- File SHA-256: `617163a508030dea2768655db66006d52097f17be9c04bb7c9fc300d3adf8a21`
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

Exact AST return positions: 551:3, 566:4, 570:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 550:2 | arm entered 2x (untagged package suite); arm entered 2x (tagged package suite); entered by `TestProductionRouteCandidatesRejectLegacyThreeFamilyAndPartialSets` |
| B2 | range | 554:2 | arm entered 232x (untagged package suite); arm entered 232x (tagged package suite); entered by `TestPairedProductionRouteAuthorityLoadsExactFourLanesIndependently`, `TestProductionRouteAuthorityBatchUsesEverySignedScopeInOneMarketSnapshot`, `TestProductionRouteAuthorityCarriesThreeIndependentSeals`, `TestProductionRouteAuthorityFailureIsMarketLocal`, `TestProductionRouteAuthorityRestoresExactActiveOwner`, `TestProductionRouteAuthoritySelectsEverySignedSymbolScope`, `TestProductionRouteCandidatesCarryNoRawArbitrationScore`, `TestProductionRouteCandidatesRejectAScorePPMAboveTheApprovedRange`, `TestProductionRouteCandidatesRejectFamilyDriftAndPartialFamilyCoverage`, `TestProductionRouteCandidatesRejectLegacyThreeFamilyAndPartialSets` |
| B3 | if | 562:3 | arm not entered (untagged package suite); arm not entered (tagged package suite); no per-test profile in the attribution set entered it |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `productionRouteDescriptors` | 549:10 |
| `len` | 550:5 |
| `len` | 550:20 |
| `make` | 553:10 |
| `len` | 553:32 |
| `validDesiredState` | 564:5 |
| `validDesiredState` | 564:42 |
| `productionRouteIdentity` | 565:5 |
| `productionRouteIdentity` | 565:55 |
| `len` | 570:9 |
| `len` | 570:22 |

## State mutations and fallbacks

- AST assignments: 4. Defers: 0. Goroutine statements: 0.

## Safety conclusion

B1 (`len(values) != len(want)`) is the arm that kills a legacy three-family manifest; the tail return is always true once the loop completes and is not a refusal arm. The family comparison inside B3 is what stops four lanes from claiming one family. A `len(families)` count here was measured to be unfalsifiable and was deleted rather than kept as an unkillable defence; the table's own four-distinct-families property is asserted by `TestProductionRouteDescriptorsCoverFourFamiliesPerMarket` instead.

a112 6.1: 같은 파일의 다른 함수 편집으로 줄만 이동(본문 불변)

a112 게이트 준비(2026-10-04): a127 82080177(전략 권한 적재기가 현재 원장을 읽음)이 같은 파일을 편집해 이 번들이 낡았다 — 분기 · return · 호출 구조 동일
