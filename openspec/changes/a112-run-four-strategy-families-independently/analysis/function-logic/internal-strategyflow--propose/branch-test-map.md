# Branch Test Map: `Propose`

- Source: `internal/strategyflow/flow.go`; file SHA-256 `c4e9738af8202122e48460436ce5cf7717b8ec8af4495b1b581171114dfe06ce`. AST branch positions are authoritative.
- Rows carry measured counts. package suite: `go test -tags tossos_testseams -covermode=count ./internal/strategyflow/`; engine suite: `go test -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- Tests whose individual coverage profile entered at least one arm: `TestProductionEvaluateUsesRealRouterAndTheSixLaneProductionFixtures`.

| Branch | Anchor | Measured disposition |
|---|---|---|
| B1 | if at 21:2 | arm entered 6x (package suite); entered by `TestProductionEvaluateUsesRealRouterAndTheSixLaneProductionFixtures` |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
