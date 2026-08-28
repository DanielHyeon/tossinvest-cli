# Function Logic Map: `TestExistingOwnerRoutePinsCampaignAndAcceptsLaneConfigLineage`

- Source: `internal/strategyflow/flow_test.go` (176-199)
- Function: `TestExistingOwnerRoutePinsCampaignAndAcceptsLaneConfigLineage` in package `strategyflow`
- Signature: `TestExistingOwnerRoutePinsCampaignAndAcceptsLaneConfigLineage(params=1, results=0)`
- File SHA-256: `59776edda49cc64112b0a744fb25fdfefb39d484df7cd87ea8cf6171f25b656b`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 2.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Pins that an existing campaign owner constrains the route and that lane config lineage is carried through.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: **none available**. `go test` does not instrument `_test.go` files, so no coverage profile can speak for this function. Each row below is classified from the arm's own source text instead, and the run that exercised the function is named.
- Measured entry: no measured profile entered this function body.

Exact AST return positions: 183:3, 185:66, 193:3, 195:66.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 186:2 | no coverage block maps to this position |
| B2 | if | 196:2 | no coverage block maps to this position |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `descriptorByID` | 177:16 |
| `approvedFixture` | 178:14 |
| `strategyrouter.NewOwnerKey` | 179:12 |
| `approved.Symbol` | 179:72 |
| `acceptedEvaluation` | 180:16 |
| `evaluateWith` | 182:12 |
| `inputFor` | 182:106 |
| `registryForTest` | 185:5 |
| `t.Fatalf` | 188:3 |
| `evaluateWith` | 192:13 |
| `inputFor` | 192:107 |
| `registryForTest` | 195:5 |
| `t.Fatalf` | 197:3 |

## State mutations and fallbacks

- AST assignments: 8. Defers: 0. Goroutine statements: 0.
- A test function mutates only its own fixtures; it opens no journal, issues no order and touches no shared state.

## Safety conclusion

- Test-only. It cannot change production behaviour; its value is the assertion it makes, and a green run means only that no guard arm fired.
