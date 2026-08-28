# Function Logic Map: `acceptedProjectionFixture`

- Source: `internal/strategyflow/canonical_projection_test.go` (10-21)
- Function: `acceptedProjectionFixture` in package `strategyflow`
- Signature: `acceptedProjectionFixture(params=2, results=1)`
- File SHA-256: `a356cbfc5382c47714e44d515b44257fece366d7ba6d4a5582b2f7c0929a9da1`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 1.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Builds one accepted result per descriptor for the projection test. L3 widened it to the eight-descriptor matrix; its single branch is the error guard on fixture construction.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: **none available**. `go test` does not instrument `_test.go` files, so no coverage profile can speak for this function. Each row below is classified from the arm's own source text instead, and the run that exercised the function is named.
- Measured entry: no measured profile entered this function body.

Exact AST return positions: 18:2, 19:3, 20:66.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 14:2 | no coverage block maps to this position |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `t.Helper` | 11:2 |
| `approvedFixture` | 12:14 |
| `strategyrouter.NewOwnerKey` | 13:14 |
| `strings.ToLower` | 13:49 |
| `string` | 13:65 |
| `approved.Symbol` | 13:112 |
| `t.Fatal` | 15:3 |
| `acceptedEvaluation` | 17:16 |
| `evaluateWith` | 18:9 |
| `inputFor` | 18:103 |
| `routeDecision` | 19:82 |
| `registryForTest` | 20:5 |

## State mutations and fallbacks

- AST assignments: 3. Defers: 0. Goroutine statements: 0.
- A test function mutates only its own fixtures; it opens no journal, issues no order and touches no shared state.

## Safety conclusion

- Test-only. It cannot change production behaviour; its value is the assertion it makes, and a green run means only that no guard arm fired.
