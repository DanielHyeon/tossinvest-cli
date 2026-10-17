# Function Logic Map: `Propose`

- Source: `internal/strategyflow/flow.go` (19-25)
- Function: `Propose` in package `strategyflow`
- Signature: `Propose(params=1, results=1)`
- File SHA-256: `c4e9738af8202122e48460436ce5cf7717b8ec8af4495b1b581171114dfe06ce`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 1.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

The cap-free q_candidate composition: same router, proposal registry, then a proposal seal on acceptance. Its one branch guards the seal.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode. package suite: `go test -tags tossos_testseams -covermode=count ./internal/strategyflow/`; engine suite: `go test -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- Measured entry: the function body executed **6x** under the package suite.

Exact AST return positions: 24:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 21:2 | arm entered 6x (package suite); entered by `TestProductionEvaluateUsesRealRouterAndTheSixLaneProductionFixtures` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `evaluateWith` | 20:12 |
| `proposalRegistry` | 20:59 |
| `sealProposalResult` | 22:12 |

## State mutations and fallbacks

- AST assignments: 2. Defers: 0. Goroutine statements: 0.
- Seals the result value it returns; no external state.

## Safety conclusion

- An accepted proposal is q_candidate authority only — not a Guardian decision, reservation, lease or order. The seal is what later proves the value was not edited after composition.
