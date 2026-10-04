# Function Logic Map: `pairedStrategyFirstLegResults`

- Source: `internal/app/engine/strategy_first_leg_admission_test.go` (188-192)
- Function: `pairedStrategyFirstLegResults` in package `engine`
- Signature: `pairedStrategyFirstLegResults(params=1, results=1)`
- File SHA-256: `7b2eca4d50c6fb8aaec09b552e1576a74329ea353617b084bc259e914119ac41`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 0.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

The paired KR/US subset used by the dormant-path tests. It is in the required set because the rename of its sibling helper moved it inside a changed hunk; its own body is unchanged.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: **none available**. `go test` does not instrument `_test.go` files, so no coverage profile can speak for this function. Each row below is classified from the arm's own source text instead, and the run that exercised the function is named.
- Measured entry: no measured profile entered this function body.

Exact AST return positions: 191:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | body | 188:1 | branchless: the whole body is one path |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `t.Helper` | 189:2 |
| `allPairedStrategyFirstLegResults` | 190:9 |

## State mutations and fallbacks

- AST assignments: 1. Defers: 0. Goroutine statements: 0.
- A test function mutates only its own fixtures; it opens no journal, issues no order and touches no shared state.

## Safety conclusion

- Test-only. It cannot change production behaviour; its value is the assertion it makes, and a green run means only that no guard arm fired.
