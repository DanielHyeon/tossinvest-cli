# Branch Test Map: `Evaluate`

Source: `internal/risk/chain.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `preflight(in)` refused | `TestPreflightAndReductionReportDistinctSteps` | pre-existing | yes |
| B2 | `in.Intent.Side == SideSell` | `TestPreflightAndReductionReportDistinctSteps` | pre-existing | yes |
| B3 | iterating the 12 `entryChain` rungs in order | `TestEveryRungReportsItsOwnName` | pre-existing | yes |
| B4 | a rung refused | `TestEveryRungReportsItsOwnName`, `TestFirstFailureStopsTheChain` | pre-existing | yes |
