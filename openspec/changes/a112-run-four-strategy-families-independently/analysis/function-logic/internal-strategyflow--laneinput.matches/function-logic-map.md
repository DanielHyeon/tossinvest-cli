# Function Logic Map: `LaneInput.matches`

- Source: `internal/strategyflow/registry.go` (122-134)
- Function: `LaneInput.matches` in package `strategyflow`
- Signature: `LaneInput.matches(params=1, results=1)`
- File SHA-256: `c7cfd15029a18c87f4de9ff2cb2730280cd1345a6d182b0eee687a11348cbdda`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 0.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Answers whether a tagged `LaneInput` belongs to the descriptor it is being bound to. L3 added the two breakout lane IDs to the table. Branchless: one map build and one comparison.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode. package suite: `go test -tags tossos_testseams -covermode=count ./internal/strategyflow/`; engine suite: `go test -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- Measured entry: the function body executed **61x** under the package suite.

Exact AST return positions: 133:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | body | 122:1 | branchless: the whole body is one path |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| — | — |

## State mutations and fallbacks

- AST assignments: 1. Defers: 0. Goroutine statements: 0.
- None.

## Safety conclusion

- Zero-value safe by construction — an unmapped lane ID yields `laneUnknown`, which never equals a real kind, so a forged or empty input cannot match any descriptor.
