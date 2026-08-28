# Function Logic Map: `TestRecordQFinalCampaignFirstLegAcceptsAllStrategyflowLanesPaired`

- Source: `internal/journal/strategyflow_projection_test.go` (176-210)
- Function: `TestRecordQFinalCampaignFirstLegAcceptsAllStrategyflowLanesPaired` in package `journal`
- Signature: `TestRecordQFinalCampaignFirstLegAcceptsAllStrategyflowLanesPaired(params=1, results=0)`
- File SHA-256: `4897ee72290331ac04e2257d255a1922214b9849ad48877fa0bf8b10d2649156`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 7.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Records a q_final first leg for every strategyflow lane. It is in the required set because the count fix above sits inside its diff hunk; its own body is unchanged and it now exercises eight lanes instead of six.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: **none available**. `go test` does not instrument `_test.go` files, so no coverage profile can speak for this function. Each row below is classified from the arm's own source text instead, and the run that exercised the function is named.
- Measured entry: no measured profile entered this function body.

Exact AST return positions: none.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | range | 179:2 | no coverage block maps to this position |
| B2 | if | 185:4 | no coverage block maps to this position |
| B3 | if | 189:4 | no coverage block maps to this position |
| B4 | if | 195:4 | no coverage block maps to this position |
| B5 | if | 198:4 | no coverage block maps to this position |
| B6 | if | 201:4 | no coverage block maps to this position |
| B7 | if | 207:2 | no coverage block maps to this position |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `strategyflow.Descriptors` | 177:17 |
| `t.Run` | 181:3 |
| `openTestJournal` | 182:9 |
| `strategyflowFirstLegFixture` | 183:15 |
| `fmt.Sprintf` | 183:61 |
| `decodeStrategyflowRiskBinding` | 184:18 |
| `t.Fatal` | 186:5 |
| `strategyflow.VerifyAcceptedProjection` | 188:18 |
| `string` | 188:56 |
| `Quantity` | 189:21 |
| `inner.ExecutionTerms` | 189:21 |
| `inner.Lineage` | 190:5 |
| `strings.HasPrefix` | 191:5 |
| `Identity` | 191:63 |
| `Policy` | 191:63 |
| `inner.ExecutionTerms` | 191:63 |
| `t.Fatalf` | 192:5 |
| `j.RecordQFinalCampaignFirstLeg` | 194:20 |
| `context.Background` | 194:51 |
| `t.Fatal` | 196:5 |
| `string` | 198:25 |
| `t.Fatalf` | 199:5 |
| `countRiskBucketRows` | 201:14 |
| `t.Fatalf` | 202:5 |
| `string` | 204:12 |
| `t.Fatalf` | 208:3 |

## State mutations and fallbacks

- AST assignments: 10. Defers: 0. Goroutine statements: 0.
- A test function mutates only its own fixtures; it opens no journal, issues no order and touches no shared state.

## Safety conclusion

- Test-only. It cannot change production behaviour; its value is the assertion it makes, and a green run means only that no guard arm fired.
