# Function Logic Map: `ProductionBatchAuthority.Len`

- Source: `internal/strategyproposal/production.go` (80-80)
- Function: `ProductionBatchAuthority.Len` in package `strategyproposal`
- Signature: `ProductionBatchAuthority.Len(params=0, results=1)`
- File SHA-256: `6cc7474d631e24c1daee677743fdbcc942787e9ae6874ed318cd3550326803b3`
- Pinned revision: `base` — the AST and the SHA-256 above are `a8c3d067470fe9cd00523a7629ee93ee05de8e5c`'s file, because the checker requires this record at the frozen comparison base (the function moved or was renamed).
- AST evidence: `ast.json` — AST branches 0.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Returns the number of sealed proposals. Branchless, one expression. Pinned at the base revision because the function moved without its body changing.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode. package suite: `go test -tags tossos_testseams -covermode=count ./internal/strategyproposal/`; engine suite: `go test -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- Measured entry: no measured profile entered this function body.

Exact AST return positions: 80:69.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | body | 80:1 | branchless: the whole body is one path |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `len` | 80:76 |

## State mutations and fallbacks

- AST assignments: 0. Defers: 0. Goroutine statements: 0.
- None.

## Safety conclusion

- With the (symbol, laneID) key this now counts proposals, not symbols. Any caller that read it as a symbol count would be wrong — measured: no production caller does.
