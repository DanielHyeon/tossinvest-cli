# Branch Test Map: `acceptedProjectionFixture`

- Source: `internal/strategyflow/canonical_projection_test.go`; file SHA-256 `a356cbfc5382c47714e44d515b44257fece366d7ba6d4a5582b2f7c0929a9da1`. AST branch positions are authoritative.
- No coverage profile can cover a `_test.go` function. Each row states what the arm is, and the run that exercised it: `go test -count=1 -tags tossos_testseams ./internal/strategyflow/`.
- This function is itself the test. The run below is the evidence that it executes and passes.

| Branch | Anchor | Classification | Observed |
|---|---|---|---|
| B1 | if at 14:2 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
