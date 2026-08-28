# Branch Test Map: `TestProductionRouteAuthoritySelectsEverySignedSymbolScope`

- Source: `internal/strategyrouter/production_test.go`; file SHA-256 `4a6fe328016fbef89ac4b186f65b5561ef7ef89b9f379837a20f12911f2eca70`. AST branch positions are authoritative.
- No coverage profile can cover a `_test.go` function. Each row states what the arm is, and the run that exercised it: `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/`.
- This function is itself the test. The run below is the evidence that it executes and passes.

| Branch | Anchor | Classification | Observed |
|---|---|---|---|
| B1 | range at 70:2 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B2 | range at 80:3 | path arm — taken on the exercised path | exercised by the named run |
| B3 | if at 90:3 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B4 | if at 94:3 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
