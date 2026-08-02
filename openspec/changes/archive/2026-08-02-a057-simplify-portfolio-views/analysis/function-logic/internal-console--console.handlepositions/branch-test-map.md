# Branch Test Map: `Console.handlePositions`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Branchless happy path builds the page, decorates request-local rows, and renders | positions rendering tests + `TestA057DashboardAndPositionsShareSimpleHoldingProjection` | yes — handler-only enrichment caused dashboard drift | yes — focused and full console suites |
