# Function Logic Map: `TestCandidateAndLaneEvidenceAreDistinctAndBothPreserved`

- Source: `internal/strategyflow/flow_test.go` (201-216)
- Function: `TestCandidateAndLaneEvidenceAreDistinctAndBothPreserved` in package `strategyflow`
- Signature: `TestCandidateAndLaneEvidenceAreDistinctAndBothPreserved(params=1, results=0)`
- File SHA-256: `59776edda49cc64112b0a744fb25fdfefb39d484df7cd87ea8cf6171f25b656b`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 1.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Pins that candidate evidence and lane evidence stay distinct and both survive composition.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: **none available**. `go test` does not instrument `_test.go` files, so no coverage profile can speak for this function. Each row below is classified from the arm's own source text instead, and the run that exercised the function is named.
- Measured entry: no measured profile entered this function body.

Exact AST return positions: 210:3, 211:66.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 212:2 | no coverage block maps to this position |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `descriptorByID` | 202:16 |
| `approvedFixture` | 203:14 |
| `strategyrouter.NewOwnerKey` | 204:12 |
| `approved.Symbol` | 204:72 |
| `acceptedEvaluation` | 205:16 |
| `evaluateWith` | 207:12 |
| `inputFor` | 207:106 |
| `routeDecision` | 208:15 |
| `registryForTest` | 211:5 |
| `approved.EvidenceDigest` | 212:77 |
| `t.Fatalf` | 214:3 |

## State mutations and fallbacks

- AST assignments: 8. Defers: 0. Goroutine statements: 0.
- A test function mutates only its own fixtures; it opens no journal, issues no order and touches no shared state.

## Safety conclusion

- Test-only. It cannot change production behaviour; its value is the assertion it makes, and a green run means only that no guard arm fired.
