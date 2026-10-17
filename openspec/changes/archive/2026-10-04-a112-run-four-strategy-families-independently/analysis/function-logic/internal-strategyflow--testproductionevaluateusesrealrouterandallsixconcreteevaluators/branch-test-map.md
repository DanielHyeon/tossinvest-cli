# Branch Test Map: TestProductionEvaluateUsesRealRouterAndAllSixConcreteEvaluators (frozen base name; no longer in the tree)

- Source: `internal/strategyflow/production_integration_test.go`; file SHA-256 `2f59dde328c3d720012c0c3dc1a259431d9b08deebc4c1106c9e1d35e323a282`. AST branch positions are authoritative.
- No coverage profile can cover a `_test.go` function. Each row states what the arm is, and the run that exercised it: `go test -count=1 -tags tossos_testseams -run '^TestProductionEvaluateUsesRealRouterAndTheSixLaneProductionFixtures$' ./internal/strategyflow/`.
- This function is itself the test. The run below is the evidence that it executes and passes.

| Branch | Anchor | Classification | Observed |
|---|---|---|---|
| B1 | range at 17:2 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B2 | if at 58:4 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B3 | if at 64:4 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B4 | if at 68:4 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B5 | if at 72:4 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B6 | if at 75:4 | path arm — taken on the exercised path | exercised by the named run |
| B7 | range at 81:4 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B8 | if at 82:5 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B9 | if at 86:4 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B10 | if at 89:4 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B11 | if at 91:5 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B12 | if at 95:4 | path arm — taken on the exercised path | exercised by the named run |
| B13 | if at 103:4 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B14 | if at 106:4 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B15 | if at 110:4 | path arm — taken on the exercised path | exercised by the named run |
| B16 | if at 116:4 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B17 | if at 121:4 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B18 | if at 125:4 | path arm — taken on the exercised path | exercised by the named run |
| B19 | if at 132:4 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
