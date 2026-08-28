# Function Logic Map: `TestEvaluateRejectsAcceptedLaneWithoutExactExecutionTerms`

- Source: `internal/strategyflow/execution_terms_test.go` (66-86)
- Function: `TestEvaluateRejectsAcceptedLaneWithoutExactExecutionTerms` in package `strategyflow`
- Signature: `TestEvaluateRejectsAcceptedLaneWithoutExactExecutionTerms(params=1, results=0)`
- File SHA-256: `8f8745de1619d99ebd859d291d253ef47c56e2679d02a45aeb5250da702c8494`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 2.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

The negative twin: an accepted lane whose execution terms are not exact must be refused, not repaired.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: **none available**. `go test` does not instrument `_test.go` files, so no coverage profile can speak for this function. Each row below is classified from the arm's own source text instead, and the run that exercised the function is named.
- Measured entry: no measured profile entered this function body.

Exact AST return positions: 79:4, 81:64.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 70:2 | no coverage block maps to this position |
| B2 | if | 83:2 | no coverage block maps to this position |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `descriptorByID` | 67:16 |
| `approvedFixture` | 68:14 |
| `strategyrouter.NewOwnerKey` | 69:14 |
| `approved.Symbol` | 69:74 |
| `t.Fatal` | 71:3 |
| `acceptedEvaluation` | 73:16 |
| `routeDecision` | 75:14 |
| `evaluateWith` | 77:12 |
| `inputFor` | 77:106 |
| `registryForTest` | 81:3 |
| `result.ExecutionTerms.Valid` | 83:103 |
| `t.Fatalf` | 84:3 |

## State mutations and fallbacks

- AST assignments: 7. Defers: 0. Goroutine statements: 0.
- A test function mutates only its own fixtures; it opens no journal, issues no order and touches no shared state.

## Safety conclusion

- Test-only. It cannot change production behaviour; its value is the assertion it makes, and a green run means only that no guard arm fired.
