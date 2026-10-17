# Function Logic Map: `TestCompleteLineageSealDetectsMutation`

- Source: `internal/strategyflow/flow_test.go` (160-174)
- Function: `TestCompleteLineageSealDetectsMutation` in package `strategyflow`
- Signature: `TestCompleteLineageSealDetectsMutation(params=1, results=0)`
- File SHA-256: `59776edda49cc64112b0a744fb25fdfefb39d484df7cd87ea8cf6171f25b656b`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 2.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Mutates a sealed lineage and requires the seal to notice. This is the test that makes the seal evidence rather than decoration.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: **none available**. `go test` does not instrument `_test.go` files, so no coverage profile can speak for this function. Each row below is classified from the arm's own source text instead, and the run that exercised the function is named.
- Measured entry: no measured profile entered this function body.

Exact AST return positions: 165:3, 166:66.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 167:2 | no coverage block maps to this position |
| B2 | if | 171:2 | no coverage block maps to this position |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `descriptorByID` | 161:16 |
| `approvedFixture` | 162:14 |
| `strategyrouter.NewOwnerKey` | 163:12 |
| `approved.Symbol` | 163:72 |
| `evaluateWith` | 164:12 |
| `inputFor` | 164:106 |
| `routeDecision` | 165:82 |
| `registryForTest` | 166:5 |
| `acceptedEvaluation` | 166:73 |
| `result.Lineage.Valid` | 167:6 |
| `t.Fatalf` | 168:3 |
| `result.Lineage.Valid` | 171:5 |
| `t.Fatal` | 172:3 |

## State mutations and fallbacks

- AST assignments: 5. Defers: 0. Goroutine statements: 0.
- A test function mutates only its own fixtures; it opens no journal, issues no order and touches no shared state.

## Safety conclusion

- Test-only. It cannot change production behaviour; its value is the assertion it makes, and a green run means only that no guard arm fired.
