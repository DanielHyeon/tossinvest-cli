# Function Logic Map: `TestProjectAcceptedStrategyflowLineagePairsAllSixKRUSLanes`

- Source: `internal/journal/strategyflow_projection_test.go` (16-74)
- Function: `TestProjectAcceptedStrategyflowLineagePairsAllSixKRUSLanes` in package `journal`
- Signature: `TestProjectAcceptedStrategyflowLineagePairsAllSixKRUSLanes(params=1, results=0)`
- File SHA-256: `3cb2ab2ecea3135d897246827710eff819e5092c5e285000014e368341085721`
- Pinned revision: `base` — the AST and the SHA-256 above are `a8c3d067470fe9cd00523a7629ee93ee05de8e5c`'s file, because the checker requires this record at the frozen comparison base (the function moved or was renamed).
- AST evidence: `ast.json` — AST branches 12.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

The frozen-base record of the journal projection matrix, renamed here to `TestProjectAcceptedStrategyflowLineagePairsAllEightKRUSLanes` and moved from six descriptors to eight (decision 50).

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: **none available**. `go test` does not instrument `_test.go` files, so no coverage profile can speak for this function. Each row below is classified from the arm's own source text instead, and the run that exercised the function is named.
- Measured entry: no measured profile entered this function body.

Exact AST return positions: none.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 18:2 | no coverage block maps to this position |
| B2 | range | 22:2 | no coverage block maps to this position |
| B3 | if | 29:3 | no coverage block maps to this position |
| B4 | if | 33:3 | no coverage block maps to this position |
| B5 | if | 40:3 | no coverage block maps to this position |
| B6 | if | 44:3 | no coverage block maps to this position |
| B7 | if | 47:3 | no coverage block maps to this position |
| B8 | if | 52:3 | no coverage block maps to this position |
| B9 | if | 56:3 | no coverage block maps to this position |
| B10 | if | 63:3 | no coverage block maps to this position |
| B11 | if | 66:3 | no coverage block maps to this position |
| B12 | if | 71:2 | no coverage block maps to this position |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `strategyflow.Descriptors` | 17:17 |
| `len` | 18:5 |
| `t.Fatalf` | 19:3 |
| `len` | 19:47 |
| `string` | 23:13 |
| `pairedProjectionPrices` | 27:77 |
| `strategyflow.AcceptedResultForJournalTest` | 28:18 |
| `t.Fatal` | 30:4 |
| `QFinalPolicyVersion` | 32:18 |
| `string` | 32:108 |
| `rune` | 32:115 |
| `t.Fatal` | 34:4 |
| `testIssued` | 38:83 |
| `ProjectAcceptedStrategyflowLineage` | 39:17 |
| `t.Fatalf` | 41:4 |
| `ProjectAcceptedStrategyflowLineage` | 43:18 |
| `t.Fatalf` | 45:4 |
| `t.Fatalf` | 49:4 |
| `decodeStrategyflowRiskBinding` | 51:17 |
| `t.Fatal` | 53:4 |
| `strategyflow.VerifyAcceptedProjection` | 55:17 |
| `string` | 55:55 |
| `Quantity` | 56:20 |
| `inner.ExecutionTerms` | 56:20 |
| `inner.Lineage` | 56:63 |
| `Identity` | 57:4 |
| `Policy` | 57:4 |
| `inner.ExecutionTerms` | 57:4 |
| `t.Fatalf` | 58:4 |
| `build` | 60:20 |
| `testIssued` | 62:62 |
| `Add` | 62:88 |
| `testIssued` | 62:88 |
| `t.Fatal` | 64:4 |
| `verifyStrategyRiskBinding` | 66:13 |
| `t.Fatalf` | 67:4 |
| `t.Fatalf` | 72:3 |

## State mutations and fallbacks

- AST assignments: 18. Defers: 0. Goroutine statements: 0.
- A test function mutates only its own fixtures; it opens no journal, issues no order and touches no shared state.

## Safety conclusion

- Test-only. It cannot change production behaviour; its value is the assertion it makes, and a green run means only that no guard arm fired.
