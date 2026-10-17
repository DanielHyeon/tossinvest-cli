# Function Logic Map: `TestWrongMarketLaneInputAndForgedAcceptedLineageFailClosed`

- Source: `internal/strategyflow/flow_test.go` (134-158)
- Function: `TestWrongMarketLaneInputAndForgedAcceptedLineageFailClosed` in package `strategyflow`
- Signature: `TestWrongMarketLaneInputAndForgedAcceptedLineageFailClosed(params=1, results=0)`
- File SHA-256: `59776edda49cc64112b0a744fb25fdfefb39d484df7cd87ea8cf6171f25b656b`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 2.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Two fail-closed cases: a lane input from the wrong market, and a forged accepted lineage. Both must refuse.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: **none available**. `go test` does not instrument `_test.go` files, so no coverage profile can speak for this function. Each row below is classified from the arm's own source text instead, and the run that exercised the function is named.
- Measured entry: no measured profile entered this function body.

Exact AST return positions: 143:3, 146:3.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 150:2 | no coverage block maps to this position |
| B2 | if | 155:2 | no coverage block maps to this position |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `descriptorByID` | 135:16 |
| `approvedFixture` | 136:14 |
| `strategyrouter.NewOwnerKey` | 137:12 |
| `approved.Symbol` | 137:72 |
| `registryForTest` | 139:14 |
| `acceptedEvaluation` | 141:13 |
| `routeDecision` | 146:82 |
| `evaluateWith` | 149:16 |
| `ContinuationUS` | 149:110 |
| `t.Fatalf` | 151:3 |
| `evaluateWith` | 154:12 |
| `inputFor` | 154:106 |
| `t.Fatalf` | 156:3 |

## State mutations and fallbacks

- AST assignments: 11. Defers: 0. Goroutine statements: 0.
- A test function mutates only its own fixtures; it opens no journal, issues no order and touches no shared state.

## Safety conclusion

- Test-only. It cannot change production behaviour; its value is the assertion it makes, and a green run means only that no guard arm fired.
