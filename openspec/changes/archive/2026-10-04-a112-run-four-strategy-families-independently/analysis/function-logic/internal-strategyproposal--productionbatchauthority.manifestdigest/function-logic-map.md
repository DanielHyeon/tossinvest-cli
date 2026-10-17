# Function Logic Map: `ProductionBatchAuthority.ManifestDigest`

- Source: `internal/strategyproposal/production.go` (81-81)
- Function: `ProductionBatchAuthority.ManifestDigest` in package `strategyproposal`
- Signature: `ProductionBatchAuthority.ManifestDigest(params=0, results=1)`
- File SHA-256: `6cc7474d631e24c1daee677743fdbcc942787e9ae6874ed318cd3550326803b3`
- Pinned revision: `base` — the AST and the SHA-256 above are `a8c3d067470fe9cd00523a7629ee93ee05de8e5c`'s file, because the checker requires this record at the frozen comparison base (the function moved or was renamed).
- AST evidence: `ast.json` — AST branches 0.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Returns the manifest digest the batch was sealed with. Branchless. Pinned at the base revision because the function moved without its body changing.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode. package suite: `go test -tags tossos_testseams -covermode=count ./internal/strategyproposal/`; engine suite: `go test -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- Measured entry: no measured profile entered this function body.

Exact AST return positions: 81:69.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | body | 81:1 | branchless: the whole body is one path |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| (no call expressions in this function) | — |

## State mutations and fallbacks

- AST assignments: 0. Defers: 0. Goroutine statements: 0.
- None.

## Safety conclusion

- The digest is the batch's provenance; it must stay the value passed at construction and is never recomputed here.
