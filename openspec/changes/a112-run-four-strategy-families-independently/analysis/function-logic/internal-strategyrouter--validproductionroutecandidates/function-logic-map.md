# Function Logic Map: `validProductionRouteCandidates`

- Source: `internal/strategyrouter/production.go` (374-390)
- Function: `validProductionRouteCandidates` in package `strategyrouter`
- Signature: `validProductionRouteCandidates(params=2, results=1)`
- File SHA-256: `dbf4e5afdfefcc6210a870d5c5e1952d3531eb119181be452e704964759bbcd8`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 3.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Refuses a manifest whose candidate list is not exactly this market's table: right count, no unknown lane, no duplicate, matching horizon and version, valid states and identity digests.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode. package suite: `go test -tags tossos_testseams -covermode=count ./internal/strategyrouter/`; engine suite: `go test -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- Measured entry: the function body executed **46x** under the package suite.

Exact AST return positions: 377:3, 385:4, 389:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 376:2 | arm entered 2x (package suite); entered by `TestProductionRouteCandidatesRejectLegacyThreeFamilyAndPartialSets` |
| B2 | range | 380:2 | arm entered 164x (package suite); entered by `TestPairedProductionRouteAuthorityLoadsExactFourLanesIndependently`, `TestProductionRouteAuthorityBatchUsesEverySignedScopeInOneMarketSnapshot`, `TestProductionRouteAuthorityFailureIsMarketLocal` |
| B3 | if | 382:3 | arm entered 164x (package suite); entered by `TestPairedProductionRouteAuthorityLoadsExactFourLanesIndependently`, `TestProductionRouteAuthorityBatchUsesEverySignedScopeInOneMarketSnapshot`, `TestProductionRouteAuthorityFailureIsMarketLocal` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `productionRouteDescriptors` | 375:10 |
| `len` | 376:5 |
| `len` | 376:20 |
| `make` | 379:10 |
| `len` | 379:32 |
| `validDesiredState` | 383:5 |
| `validDesiredState` | 383:42 |
| `productionRouteIdentity` | 384:5 |
| `productionRouteIdentity` | 384:55 |
| `len` | 389:9 |
| `len` | 389:22 |

## State mutations and fallbacks

- AST assignments: 4. Defers: 0. Goroutine statements: 0.
- Builds one local `seen` map.

## Safety conclusion

- The count equality plus the duplicate check make 'exactly four families' a checked property. A three-family legacy manifest is no longer an activation authority.
