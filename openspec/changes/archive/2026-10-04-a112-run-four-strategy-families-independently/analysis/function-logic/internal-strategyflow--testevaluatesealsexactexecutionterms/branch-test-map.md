# Branch Test Map: `TestEvaluateSealsExactExecutionTerms`

- Source: `internal/strategyflow/execution_terms_test.go`; file SHA-256 `8f8745de1619d99ebd859d291d253ef47c56e2679d02a45aeb5250da702c8494`. AST branch positions are authoritative.
- No coverage profile can cover a `_test.go` function. Each row states what the arm is, and the run that exercised it: `go test -count=1 -tags tossos_testseams ./internal/strategyflow/`.
- This function is itself the test. The run below is the evidence that it executes and passes.

| Branch | Anchor | Classification | Observed |
|---|---|---|---|
| B1 | if at 13:2 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B2 | if at 26:2 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B3 | if at 29:2 | path arm — taken on the exercised path | exercised by the named run |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
