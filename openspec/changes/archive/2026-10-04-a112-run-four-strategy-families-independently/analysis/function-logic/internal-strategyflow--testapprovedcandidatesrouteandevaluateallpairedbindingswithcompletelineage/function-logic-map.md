# Function Logic Map: `TestApprovedCandidatesRouteAndEvaluateAllPairedBindingsWithCompleteLineage`

- Source: `internal/strategyflow/flow_test.go` (50-86)
- Function: `TestApprovedCandidatesRouteAndEvaluateAllPairedBindingsWithCompleteLineage` in package `strategyflow`
- Signature: `TestApprovedCandidatesRouteAndEvaluateAllPairedBindingsWithCompleteLineage(params=1, results=0)`
- File SHA-256: `59776edda49cc64112b0a744fb25fdfefb39d484df7cd87ea8cf6171f25b656b`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 6.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Drives every paired binding end to end and asserts complete lineage on each. Widened to the 8-set.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: **none available**. `go test` does not instrument `_test.go` files, so no coverage profile can speak for this function. Each row below is classified from the arm's own source text instead, and the run that exercised the function is named.
- Measured entry: no measured profile entered this function body.

Exact AST return positions: 62:5, 66:5.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | range | 51:2 | no coverage block maps to this position |
| B2 | if | 56:4 | no coverage block maps to this position |
| B3 | if | 68:4 | no coverage block maps to this position |
| B4 | if | 71:4 | no coverage block maps to this position |
| B5 | if | 74:4 | no coverage block maps to this position |
| B6 | if | 81:4 | no coverage block maps to this position |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `Descriptors` | 51:29 |
| `t.Run` | 53:3 |
| `approvedFixture` | 54:16 |
| `strategyrouter.NewOwnerKey` | 55:16 |
| `approved.Symbol` | 55:70 |
| `t.Fatal` | 57:5 |
| `registryForTest` | 60:16 |
| `acceptedEvaluation` | 62:12 |
| `evaluateWith` | 64:14 |
| `inputFor` | 64:108 |
| `routeDecision` | 66:84 |
| `result.Lineage.Valid` | 68:62 |
| `t.Fatalf` | 69:5 |
| `t.Fatalf` | 72:5 |
| `approved.CandidateLifeID` | 74:87 |
| `t.Fatalf` | 79:5 |
| `t.Fatalf` | 82:5 |

## State mutations and fallbacks

- AST assignments: 8. Defers: 0. Goroutine statements: 0.
- A test function mutates only its own fixtures; it opens no journal, issues no order and touches no shared state.

## Safety conclusion

- Test-only. It cannot change production behaviour; its value is the assertion it makes, and a green run means only that no guard arm fired.
