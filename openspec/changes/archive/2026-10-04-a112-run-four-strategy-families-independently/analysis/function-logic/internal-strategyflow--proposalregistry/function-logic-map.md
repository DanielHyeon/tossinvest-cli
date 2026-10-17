# Function Logic Map: `proposalRegistry`

- Source: `internal/strategyflow/adapters.go` (26-37)
- Function: `proposalRegistry` in package `strategyflow`
- Signature: `proposalRegistry(params=0, results=1)`
- File SHA-256: `0f6b4e682e89e6d24c4c3686a5a1ad5ea1f0825e904236ea892b5905029065b6`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 0.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

The cap-free proposal twin of `defaultRegistry`, bound to `proposeBreakoutKR`/`proposeBreakoutUS` for the two new descriptors. Branchless.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode. package suite: `go test -tags tossos_testseams -covermode=count ./internal/strategyflow/`; engine suite: `go test -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- Measured entry: the function body executed **7x** under the package suite.

Exact AST return positions: 27:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | body | 26:1 | branchless: the whole body is one path |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `newRegistry` | 27:9 |

## State mutations and fallbacks

- AST assignments: 0. Defers: 0. Goroutine statements: 0.
- Constructs and returns a fresh registry; mutates nothing reachable from outside.

## Safety conclusion

- Proposal bindings read q_candidate and must not apply the manifest cap; the cap belongs to the evaluate side (decision 48).
