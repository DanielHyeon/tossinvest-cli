# Function Logic Map: `TestProductionEvaluateUsesRealRouterAndAllSixConcreteEvaluators`

- Source: `internal/strategyflow/production_integration_test.go` (15-137)
- Function: `TestProductionEvaluateUsesRealRouterAndAllSixConcreteEvaluators` in package `strategyflow`
- Signature: `TestProductionEvaluateUsesRealRouterAndAllSixConcreteEvaluators(params=1, results=0)`
- File SHA-256: `2f59dde328c3d720012c0c3dc1a259431d9b08deebc4c1106c9e1d35e323a282`
- Pinned revision: `base` — the AST and the SHA-256 above are `a8c3d067470fe9cd00523a7629ee93ee05de8e5c`'s file, because the checker requires this record at the frozen comparison base (the function moved or was renamed).
- AST evidence: `ast.json` — AST branches 19.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

The frozen-base record of the production integration table. It enumerates six lanes explicitly rather than counting descriptors, so it stayed green — but its name claimed to cover all concrete evaluators while two of the eight were absent. Renamed here to `TestProductionEvaluateUsesRealRouterAndTheSixLaneProductionFixtures`, which states what it actually covers; the breakout pair is covered instead by `TestBreakoutProposeIsCapFreeAndEvaluateAppliesTheFinalCap` and its siblings in `breakout_binding_test.go`.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: **none available**. `go test` does not instrument `_test.go` files, so no coverage profile can speak for this function. Each row below is classified from the arm's own source text instead, and the run that exercised the function is named.
- Measured entry: no measured profile entered this function body.

Exact AST return positions: 27:5, 32:5, 37:5, 42:5, 47:5, 52:5.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | range | 17:2 | no coverage block maps to this position |
| B2 | if | 58:4 | no coverage block maps to this position |
| B3 | if | 64:4 | no coverage block maps to this position |
| B4 | if | 68:4 | no coverage block maps to this position |
| B5 | if | 72:4 | no coverage block maps to this position |
| B6 | if | 75:4 | no coverage block maps to this position |
| B7 | range | 81:4 | no coverage block maps to this position |
| B8 | if | 82:5 | no coverage block maps to this position |
| B9 | if | 86:4 | no coverage block maps to this position |
| B10 | if | 89:4 | no coverage block maps to this position |
| B11 | if | 91:5 | no coverage block maps to this position |
| B12 | if | 95:4 | no coverage block maps to this position |
| B13 | if | 103:4 | no coverage block maps to this position |
| B14 | if | 106:4 | no coverage block maps to this position |
| B15 | if | 110:4 | no coverage block maps to this position |
| B16 | if | 116:4 | no coverage block maps to this position |
| B17 | if | 121:4 | no coverage block maps to this position |
| B18 | if | 125:4 | no coverage block maps to this position |
| B19 | if | 132:4 | no coverage block maps to this position |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `time.Date` | 16:17 |
| `continuationlane.StrategyflowKRFixture` | 26:21 |
| `ContinuationKR` | 27:12 |
| `continuationlane.StrategyflowUSFixture` | 31:21 |
| `ContinuationUS` | 32:12 |
| `reversallane.StrategyflowKRFixture` | 36:21 |
| `ReversalKR` | 37:12 |
| `reversallane.StrategyflowUSFixture` | 41:21 |
| `ReversalUS` | 42:12 |
| `weeklyvaluelane.StrategyflowKRFixture` | 46:21 |
| `WeeklyKR` | 47:12 |
| `weeklyvaluelane.StrategyflowUSFixture` | 51:21 |
| `WeeklyUS` | 52:12 |
| `t.Run` | 55:3 |
| `approvedFixture` | 56:16 |
| `strategyrouter.NewOwnerKey` | 57:16 |
| `approved.Symbol` | 57:64 |
| `t.Fatal` | 59:5 |
| `strategyrouter.StrategyflowRouteFixture` | 62:26 |
| `t.Fatal` | 65:5 |
| `test.input` | 67:22 |
| `approved.CandidateLifeID` | 67:33 |
| `t.Fatal` | 69:5 |
| `Evaluate` | 71:14 |
| `result.Lineage.Valid` | 72:90 |
| `t.Fatalf` | 73:5 |
| `result.ExecutionTerms.Valid` | 75:8 |
| `PriceMinor` | 75:41 |
| `result.ExecutionTerms.Entry` | 75:41 |
| `PriceMinor` | 76:5 |
| `result.ExecutionTerms.EffectiveStop` | 76:5 |
| `PriceMinor` | 76:72 |
| `result.ExecutionTerms.Target` | 76:72 |
| `result.ExecutionTerms.Quantity` | 77:5 |
| `result.ExecutionTerms.LineageIdentity` | 77:60 |
| `t.Fatalf` | 78:5 |
| `result.ExecutionTerms.Entry` | 81:49 |
| `result.ExecutionTerms.EffectiveStop` | 81:80 |
| `result.ExecutionTerms.Target` | 81:119 |
| `provenance.Source` | 82:8 |

(51 further call sites omitted; `ast.json` carries all 91.)

## State mutations and fallbacks

- AST assignments: 20. Defers: 0. Goroutine statements: 0.
- A test function mutates only its own fixtures; it opens no journal, issues no order and touches no shared state.

## Safety conclusion

- Test-only. It cannot change production behaviour; its value is the assertion it makes, and a green run means only that no guard arm fired.
