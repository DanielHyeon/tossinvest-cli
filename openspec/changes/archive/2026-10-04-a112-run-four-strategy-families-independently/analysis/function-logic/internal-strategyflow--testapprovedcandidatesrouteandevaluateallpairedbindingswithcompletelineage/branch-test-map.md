# Branch Test Map: `TestApprovedCandidatesRouteAndEvaluateAllPairedBindingsWithCompleteLineage`

- Source: `internal/strategyflow/flow_test.go`; file SHA-256 `59776edda49cc64112b0a744fb25fdfefb39d484df7cd87ea8cf6171f25b656b`. AST branch positions are authoritative.
- No coverage profile can cover a `_test.go` function. Each row states what the arm is, and the run that exercised it: `go test -count=1 -tags tossos_testseams ./internal/strategyflow/`.
- This function is itself the test. The run below is the evidence that it executes and passes.

| Branch | Anchor | Classification | Observed |
|---|---|---|---|
| B1 | range at 51:2 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B2 | if at 56:4 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B3 | if at 68:4 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B4 | if at 71:4 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B5 | if at 74:4 | path arm — taken on the exercised path | exercised by the named run |
| B6 | if at 81:4 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
