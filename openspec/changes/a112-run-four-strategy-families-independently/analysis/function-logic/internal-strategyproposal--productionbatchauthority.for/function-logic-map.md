# Function Logic Map: `ProductionBatchAuthority.For`

- Source: `internal/strategyproposal/production.go` (100-106)
- Function: `ProductionBatchAuthority.For` in package `strategyproposal`
- Signature: `ProductionBatchAuthority.For(params=1, results=2)`
- File SHA-256: `e2285c5ef57e399bf3bf2ca3a0e91b7449b2c152dd9623d5a617454f934082ad`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 1.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Returns the single proposal for a symbol. Its one branch refuses whenever the symbol does not carry exactly one.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode. package suite: `go test -tags tossos_testseams -covermode=count ./internal/strategyproposal/`; engine suite: `go test -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- Measured entry: the function body executed **3x** under the package suite.

Exact AST return positions: 103:3, 105:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 102:2 | arm never entered: count 0 in every profile measured for this function |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `authority.LanesFor` | 101:12 |
| `len` | 102:5 |

## State mutations and fallbacks

- AST assignments: 1. Defers: 0. Goroutine statements: 0.
- None.

## Safety conclusion

- Refusing on two-or-more is deliberate: picking one here would reintroduce the pre-evaluation selection L3 exists to remove. Callers that want all families must use `LanesFor`.
