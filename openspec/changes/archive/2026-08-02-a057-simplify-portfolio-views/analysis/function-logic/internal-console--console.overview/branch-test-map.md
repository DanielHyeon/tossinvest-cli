# Branch Test Map: `Console.overview`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | branchless assembly keeps one ledger open, zero dashboard broker calls, and adds shared rows | existing ledger/rate-budget tests plus `TestA057DashboardAndPositionsShareSimpleHoldingProjection` | yes — dashboard previously had no holding rows | yes — focused and full console suites |
