# Branch Test Map: TestQFinalCampaignFirstLegProjectsGuardianOwnedLineageAllSixLanes (frozen base name; no longer in the tree)

- Source: `internal/execgw/riskguardian_account_base_testseam_test.go`; file SHA-256 `57f4899aaf068d61c4456252f9a5e2d2372d0c515b2c395914e6609e86f32739`. AST branch positions are authoritative.
- No coverage profile can cover a `_test.go` function. Each row states what the arm is, and the run that exercised it: `go test -count=1 -tags tossos_testseams -run '^TestQFinalCampaignFirstLegProjectsGuardianOwnedLineageAllEightLanes$' ./internal/execgw/`.
- This function is itself the test. The run below is the evidence that it executes and passes.

| Branch | Anchor | Classification | Observed |
|---|---|---|---|
| B1 | if at 100:2 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B2 | range at 104:2 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B3 | if at 120:4 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B4 | if at 127:4 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B5 | if at 140:4 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B6 | if at 143:4 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B7 | if at 147:4 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B8 | if at 150:4 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B9 | if at 155:4 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B10 | if at 159:4 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B11 | if at 163:4 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B12 | if at 167:4 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B13 | if at 171:4 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B14 | if at 177:2 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
