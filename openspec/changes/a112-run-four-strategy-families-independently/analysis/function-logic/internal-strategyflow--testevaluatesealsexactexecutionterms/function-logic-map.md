# Function Logic Map: `TestEvaluateSealsExactExecutionTerms`

- Source: `internal/strategyflow/execution_terms_test.go` (9-34)
- Function: `TestEvaluateSealsExactExecutionTerms` in package `strategyflow`
- Signature: `TestEvaluateSealsExactExecutionTerms(params=1, results=0)`
- File SHA-256: `8f8745de1619d99ebd859d291d253ef47c56e2679d02a45aeb5250da702c8494`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 3.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Pins that an accepted evaluation seals exact execution terms bound to the same lineage identity. L3 touched it because the request now carries a RouteSet-derived decision.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: **none available**. `go test` does not instrument `_test.go` files, so no coverage profile can speak for this function. Each row below is classified from the arm's own source text instead, and the run that exercised the function is named.
- Measured entry: no measured profile entered this function body.

Exact AST return positions: 22:4, 24:64.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 13:2 | no coverage block maps to this position |
| B2 | if | 26:2 | no coverage block maps to this position |
| B3 | if | 29:2 | no coverage block maps to this position |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `descriptorByID` | 10:16 |
| `approvedFixture` | 11:14 |
| `strategyrouter.NewOwnerKey` | 12:14 |
| `approved.Symbol` | 12:74 |
| `t.Fatal` | 14:3 |
| `acceptedEvaluation` | 16:16 |
| `routeDecision` | 18:14 |
| `evaluateWith` | 20:12 |
| `inputFor` | 20:106 |
| `registryForTest` | 24:3 |
| `result.ExecutionTerms.Valid` | 26:36 |
| `t.Fatalf` | 27:3 |
| `PriceMinor` | 29:5 |
| `result.ExecutionTerms.Entry` | 29:5 |
| `PriceMinor` | 29:60 |
| `result.ExecutionTerms.EffectiveStop` | 29:60 |
| `PriceMinor` | 30:3 |
| `result.ExecutionTerms.Target` | 30:3 |
| `result.ExecutionTerms.Quantity` | 30:59 |
| `result.ExecutionTerms.LineageIdentity` | 31:3 |
| `t.Fatalf` | 32:3 |

## State mutations and fallbacks

- AST assignments: 7. Defers: 0. Goroutine statements: 0.
- A test function mutates only its own fixtures; it opens no journal, issues no order and touches no shared state.

## Safety conclusion

- Test-only. It cannot change production behaviour; its value is the assertion it makes, and a green run means only that no guard arm fired.
