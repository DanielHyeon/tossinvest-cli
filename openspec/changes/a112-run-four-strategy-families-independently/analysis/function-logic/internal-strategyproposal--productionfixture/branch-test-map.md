# Branch Test Map: `productionFixture`

- Source: `internal/strategyproposal/production_test.go`; file SHA-256 `28dc27e289908691099ee4c43139f1f8b0796bf186409849fd810c07209af0ee`. AST branch positions are authoritative.
- No coverage profile can cover a `_test.go` function. Each row states what the arm is, and the run that exercised it: `go test -count=1 -tags tossos_testseams ./internal/strategyproposal/`.
- This function is itself the test. The run below is the evidence that it executes and passes.

| Branch | Anchor | Classification | Observed |
|---|---|---|---|
| B1 | if at 107:2 | path arm — taken on the exercised path | exercised by the named run |
| B2 | if at 118:2 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B3 | if at 123:2 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B4 | if at 126:2 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B5 | if at 131:2 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B6 | if at 134:2 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B7 | if at 137:2 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B8 | if at 143:2 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B9 | if at 147:2 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B10 | if at 153:2 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B11 | if at 169:2 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B12 | if at 174:2 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B13 | if at 178:2 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B14 | if at 181:2 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B15 | if at 185:2 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
