# Function Logic Map: `laneMatchesMarket`

- Source: `internal/strategyproposal/production.go` (546-554)
- Function: `laneMatchesMarket` in package `strategyproposal`
- Signature: `laneMatchesMarket(params=4, results=1)`
- File SHA-256: `e2285c5ef57e399bf3bf2ca3a0e91b7449b2c152dd9623d5a617454f934082ad`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 2.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Answers whether a (lane, version, horizon) triple is one this market may run. L3 added the KR and US breakout lanes at `LaneVersionV1`/`HorizonShort`.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode. package suite: `go test -tags tossos_testseams -covermode=count ./internal/strategyproposal/`; engine suite: `go test -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- Measured entry: the function body executed **24x** under the package suite.

Exact AST return positions: 548:3, 551:3, 553:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 547:2 | arm entered 15x (package suite); entered by `TestLaneMatchesMarketCoversTheBreakoutFamily`, `TestProductionProposalAuthorityLoadsPairedSignedKRUSSnapshots`, `TestProductionProposalAuthorityRecognizesExactPairedSixLaneMatrix` |
| B2 | if | 550:2 | arm entered 9x (package suite); entered by `TestLaneMatchesMarketCoversTheBreakoutFamily`, `TestProductionProposalAuthorityFailureIsMarketLocal`, `TestProductionProposalAuthorityLoadsPairedSignedKRUSSnapshots` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| — | — |

## State mutations and fallbacks

- AST assignments: 0. Defers: 0. Goroutine statements: 0.
- None.

## Safety conclusion

- An unknown market falls through to `false`, so the default is refusal. This is the only place the market/lane pairing is asserted on the proposal path.
