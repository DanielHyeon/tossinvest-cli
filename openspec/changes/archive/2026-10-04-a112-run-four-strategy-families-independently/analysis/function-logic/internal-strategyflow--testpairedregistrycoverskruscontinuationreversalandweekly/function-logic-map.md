# Function Logic Map: `TestPairedRegistryCoversKRUSContinuationReversalAndWeekly`

- Source: `internal/strategyflow/flow_test.go` (19-46)
- Function: `TestPairedRegistryCoversKRUSContinuationReversalAndWeekly` in package `strategyflow`
- Signature: `TestPairedRegistryCoversKRUSContinuationReversalAndWeekly(params=1, results=0)`
- File SHA-256: `493e31e378b3aa9f7bf41e73cdb16db9f0cb5dc79342f8c5dbcacc1c657b4fe2`
- Pinned revision: `base` — the AST and the SHA-256 above are `a8c3d067470fe9cd00523a7629ee93ee05de8e5c`'s file, because the checker requires this record at the frozen comparison base (the function moved or was renamed).
- AST evidence: `ast.json` — AST branches 7.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

The frozen-base record of the registry coverage test, renamed in this worktree to `TestPairedRegistryCoversKRUSContinuationReversalWeeklyAndBreakout` and widened from six to eight descriptors.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: **none available**. `go test` does not instrument `_test.go` files, so no coverage profile can speak for this function. Each row below is classified from the arm's own source text instead, and the run that exercised the function is named.
- Measured entry: no measured profile entered this function body.

Exact AST return positions: none.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 21:2 | no coverage block maps to this position |
| B2 | if | 24:2 | no coverage block maps to this position |
| B3 | range | 32:2 | no coverage block maps to this position |
| B4 | if | 33:3 | no coverage block maps to this position |
| B5 | if | 36:3 | no coverage block maps to this position |
| B6 | range | 41:2 | no coverage block maps to this position |
| B7 | if | 42:3 | no coverage block maps to this position |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `Descriptors` | 20:17 |
| `ValidateDescriptors` | 21:12 |
| `t.Fatalf` | 22:3 |
| `len` | 24:5 |
| `t.Fatalf` | 25:3 |
| `len` | 25:56 |
| `t.Fatalf` | 34:4 |
| `t.Fatalf` | 37:4 |
| `t.Fatalf` | 43:4 |

## State mutations and fallbacks

- AST assignments: 5. Defers: 0. Goroutine statements: 0.
- A test function mutates only its own fixtures; it opens no journal, issues no order and touches no shared state.

## Safety conclusion

- Test-only. It cannot change production behaviour; its value is the assertion it makes, and a green run means only that no guard arm fired.
