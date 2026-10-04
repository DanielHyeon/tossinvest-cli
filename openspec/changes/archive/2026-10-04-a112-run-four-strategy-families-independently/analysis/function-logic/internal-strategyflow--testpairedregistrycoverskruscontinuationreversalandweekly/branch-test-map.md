# Branch Test Map: TestPairedRegistryCoversKRUSContinuationReversalAndWeekly (frozen base name; no longer in the tree)

- Source: `internal/strategyflow/flow_test.go`; file SHA-256 `493e31e378b3aa9f7bf41e73cdb16db9f0cb5dc79342f8c5dbcacc1c657b4fe2`. AST branch positions are authoritative.
- No coverage profile can cover a `_test.go` function. Each row states what the arm is, and the run that exercised it: `go test -count=1 -tags tossos_testseams ./internal/strategyflow/`.
- This function is itself the test. The run below is the evidence that it executes and passes.

| Branch | Anchor | Classification | Observed |
|---|---|---|---|
| B1 | if at 21:2 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B2 | if at 24:2 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B3 | range at 32:2 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B4 | if at 33:3 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B5 | if at 36:3 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B6 | range at 41:2 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B7 | if at 42:3 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
