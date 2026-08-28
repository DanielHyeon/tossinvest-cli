# Branch Test Map: TestProjectAcceptedStrategyflowLineagePairsAllSixKRUSLanes (frozen base name; no longer in the tree)

- Source: `internal/journal/strategyflow_projection_test.go`; file SHA-256 `3cb2ab2ecea3135d897246827710eff819e5092c5e285000014e368341085721`. AST branch positions are authoritative.
- No coverage profile can cover a `_test.go` function. Each row states what the arm is, and the run that exercised it: `go test -count=1 -tags tossos_testseams -run '^TestProjectAcceptedStrategyflowLineagePairsAllEightKRUSLanes$' ./internal/journal/`.
- This function is itself the test. The run below is the evidence that it executes and passes.

| Branch | Anchor | Classification | Observed |
|---|---|---|---|
| B1 | if at 18:2 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B2 | range at 22:2 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B3 | if at 29:3 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B4 | if at 33:3 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B5 | if at 40:3 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B6 | if at 44:3 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B7 | if at 47:3 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B8 | if at 52:3 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B9 | if at 56:3 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B10 | if at 63:3 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B11 | if at 66:3 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B12 | if at 71:2 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
