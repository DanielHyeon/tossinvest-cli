# Function Logic Map: `Evaluate`

- Source: `internal/strategyflow/flow.go` (12-14)
- Function: `Evaluate` in package `strategyflow`
- Signature: `Evaluate(params=1, results=1)`
- File SHA-256: `c4e9738af8202122e48460436ce5cf7717b8ec8af4495b1b581171114dfe06ce`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 0.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

The fixed production composition. Branchless: it delegates to `evaluateWith` with the real router and the real evaluate registry. L3 changed its router argument from `strategyrouter.Route` to `strategyrouter.RouteSet`.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode. package suite: `go test -tags tossos_testseams -covermode=count ./internal/strategyflow/`; engine suite: `go test -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- Measured entry: the function body executed **6x** under the package suite.

Exact AST return positions: 13:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | body | 12:1 | branchless: the whole body is one path |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `evaluateWith` | 13:9 |
| `defaultRegistry` | 13:56 |

## State mutations and fallbacks

- AST assignments: 0. Defers: 0. Goroutine statements: 0.
- None of its own.

## Safety conclusion

- Substitution stays package-private so no caller can forge router or lane acceptance. The change of router function is the whole of L3's edit here.
