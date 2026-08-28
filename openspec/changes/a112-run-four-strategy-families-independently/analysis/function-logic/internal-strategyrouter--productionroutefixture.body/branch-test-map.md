# Branch Test Map: `productionRouteFixture.body`

- Source: `internal/strategyrouter/production_test.go`; file SHA-256 `4a6fe328016fbef89ac4b186f65b5561ef7ef89b9f379837a20f12911f2eca70`. AST branch positions are authoritative.
- No coverage profile can cover a `_test.go` function. Each row states what the arm is, and the run that exercised it: `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/`.
- This function is itself the test. The run below is the evidence that it executes and passes.

| Branch | Anchor | Classification | Observed |
|---|---|---|---|
| B1 | if at 251:2 | path arm — taken on the exercised path | exercised by the named run |
| B2 | else at 258:9 | path arm — taken on the exercised path | exercised by the named run |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
