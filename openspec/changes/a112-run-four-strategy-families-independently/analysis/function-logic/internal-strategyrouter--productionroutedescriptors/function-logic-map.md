# Function Logic Map: `productionRouteDescriptors`

- Source: `internal/strategyrouter/production.go` (397-415)
- Function: `productionRouteDescriptors` in package `strategyrouter`
- Signature: `productionRouteDescriptors(params=1, results=1)`
- File SHA-256: `dbf4e5afdfefcc6210a870d5c5e1952d3531eb119181be452e704964759bbcd8`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 2.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

The per-market table of lanes a signed route manifest may contain. L3 added the KR and US breakout lanes, so each market's table is four entries.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode. package suite: `go test -tags tossos_testseams -covermode=count ./internal/strategyrouter/`; engine suite: `go test -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- Measured entry: the function body executed **52x** under the package suite.

Exact AST return positions: 399:3, 407:3, 414:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 398:2 | arm entered 26x (package suite); entered by `TestPairedProductionRouteAuthorityLoadsExactFourLanesIndependently`, `TestProductionRouteAuthorityBatchUsesEverySignedScopeInOneMarketSnapshot`, `TestProductionRouteAuthorityRestoresExactActiveOwner` |
| B2 | if | 406:2 | arm entered 25x (package suite); entered by `TestPairedProductionRouteAuthorityLoadsExactFourLanesIndependently`, `TestProductionRouteAuthorityBatchUsesEverySignedScopeInOneMarketSnapshot`, `TestProductionRouteAuthorityFailureIsMarketLocal` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| — | — |

## State mutations and fallbacks

- AST assignments: 0. Defers: 0. Goroutine statements: 0.
- None. Builds and returns a fresh map.

## Safety conclusion

- An unknown market returns nil, which `validProductionRouteCandidates` treats as a zero-length want and therefore refuses every candidate — the default is refusal.
