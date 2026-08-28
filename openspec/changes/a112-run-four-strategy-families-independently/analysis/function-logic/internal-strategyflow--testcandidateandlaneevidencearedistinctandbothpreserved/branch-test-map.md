# Branch Test Map: `TestCandidateAndLaneEvidenceAreDistinctAndBothPreserved`

- Source: `internal/strategyflow/flow_test.go`; file SHA-256 `59776edda49cc64112b0a744fb25fdfefb39d484df7cd87ea8cf6171f25b656b`. AST branch positions are authoritative.
- No coverage profile can cover a `_test.go` function. Each row states what the arm is, and the run that exercised it: `go test -count=1 -tags tossos_testseams ./internal/strategyflow/`.
- This function is itself the test. The run below is the evidence that it executes and passes.

| Branch | Anchor | Classification | Observed |
|---|---|---|---|
| B1 | if at 212:2 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
