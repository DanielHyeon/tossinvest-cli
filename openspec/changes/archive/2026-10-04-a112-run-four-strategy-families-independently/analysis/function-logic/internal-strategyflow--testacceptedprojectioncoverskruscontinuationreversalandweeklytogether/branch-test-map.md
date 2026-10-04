# Branch Test Map: TestAcceptedProjectionCoversKRUSContinuationReversalAndWeeklyTogether (frozen base name; no longer in the tree)

- Source: `internal/strategyflow/canonical_projection_test.go`; file SHA-256 `ac31a33808c126bebe48a706d67aa69783547b85d2aa7d24ce184371a84d40b4`. AST branch positions are authoritative.
- No coverage profile can cover a `_test.go` function. Each row states what the arm is, and the run that exercised it: `go test -count=1 -tags tossos_testseams ./internal/strategyflow/`.
- This function is itself the test. The run below is the evidence that it executes and passes.

| Branch | Anchor | Classification | Observed |
|---|---|---|---|
| B1 | if at 25:2 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B2 | range at 29:2 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B3 | if at 32:3 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B4 | if at 36:3 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B5 | if at 40:3 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B6 | if at 44:3 | path arm — taken on the exercised path | exercised by the named run |
| B7 | if at 51:2 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
