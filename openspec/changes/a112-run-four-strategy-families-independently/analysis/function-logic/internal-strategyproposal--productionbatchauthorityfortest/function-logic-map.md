# Function Logic Map: `ProductionBatchAuthorityForTest`

- Source: `internal/strategyproposal/production_testseam.go` (7-18)
- Function: `ProductionBatchAuthorityForTest` in package `strategyproposal`
- Signature: `ProductionBatchAuthorityForTest(params=2, results=1)`
- File SHA-256: `854454d6d04e8527260f0f6148ac72660dd5871ba1667917f4bf5048aff4156b`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 2.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Seals a batch from a symbol-keyed map of results, for engine tests. L3 changed it to build its keys with the same `batchKey` the production path uses, deriving the lane from `proposal.Lineage.LaneID`.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode. package suite: `go test -tags tossos_testseams -covermode=count ./internal/strategyproposal/`; engine suite: `go test -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- Measured entry: the function body executed **21x** under the engine suite.

Exact AST return positions: 11:4, 17:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | range | 9:2 | arm entered 21x (engine suite) |
| B2 | if | 10:3 | arm never entered: count 0 in every profile measured for this function |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `make` | 8:12 |
| `len` | 8:49 |
| `proposal.ValidProposal` | 10:7 |
| `batchKey` | 15:10 |

## State mutations and fallbacks

- AST assignments: 2. Defers: 0. Goroutine statements: 0.
- Builds and returns one sealed batch; no external state.

## Safety conclusion

- Before this fix the seam keyed by bare symbol while production keyed by (symbol, laneID), so `For` found nothing and seven engine tests reported `NO_ACCEPTED_PROPOSAL`. Sharing one key function makes that drift impossible by construction (decision 50).
