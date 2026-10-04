# Function Logic Map: `TestLaneRefusalPreservesFirstTypedCodeAndRouterLaneEvidence`

- Source: `internal/strategyflow/flow_test.go` (112-132)
- Function: `TestLaneRefusalPreservesFirstTypedCodeAndRouterLaneEvidence` in package `strategyflow`
- Signature: `TestLaneRefusalPreservesFirstTypedCodeAndRouterLaneEvidence(params=1, results=0)`
- File SHA-256: `59776edda49cc64112b0a744fb25fdfefb39d484df7cd87ea8cf6171f25b656b`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 3.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Pins first-wins refusal ordering — the property `evaluateWith`'s branch order depends on.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: **none available**. `go test` does not instrument `_test.go` files, so no coverage profile can speak for this function. Each row below is classified from the arm's own source text instead, and the run that exercised the function is named.
- Measured entry: no measured profile entered this function body.

Exact AST return positions: 117:3, 120:3.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 122:2 | no coverage block maps to this position |
| B2 | if | 125:2 | no coverage block maps to this position |
| B3 | if | 129:2 | no coverage block maps to this position |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `descriptorByID` | 113:16 |
| `approvedFixture` | 114:14 |
| `strategyrouter.NewOwnerKey` | 115:12 |
| `approved.Symbol` | 115:72 |
| `registryForTest` | 116:14 |
| `string` | 117:37 |
| `evaluateWith` | 119:12 |
| `inputFor` | 119:106 |
| `routeDecision` | 120:82 |
| `string` | 122:56 |
| `t.Fatalf` | 123:3 |
| `approved.EvidenceDigest` | 125:47 |
| `t.Fatalf` | 127:3 |
| `t.Fatalf` | 130:3 |

## State mutations and fallbacks

- AST assignments: 5. Defers: 0. Goroutine statements: 0.
- A test function mutates only its own fixtures; it opens no journal, issues no order and touches no shared state.

## Safety conclusion

- Test-only. It cannot change production behaviour; its value is the assertion it makes, and a green run means only that no guard arm fired.
