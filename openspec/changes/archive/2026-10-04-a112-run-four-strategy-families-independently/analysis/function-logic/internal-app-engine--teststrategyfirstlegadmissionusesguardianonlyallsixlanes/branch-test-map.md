# Branch Test Map: TestStrategyFirstLegAdmissionUsesGuardianOnlyAllSixLanes (frozen base name; no longer in the tree)

- Source: `internal/app/engine/strategy_first_leg_admission_test.go`; file SHA-256 `367a5e296f03b904d72489ab4dd505f8d0ce93cc74b5a80cd29971ac3434f60f`. AST branch positions are authoritative.
- No coverage profile can cover a `_test.go` function. Each row states what the arm is, and the run that exercised it: `go test -count=1 -tags tossos_testseams -run '^TestStrategyFirstLegAdmissionUsesGuardianOnlyAllEightLanes$' ./internal/app/engine/`.
- This function is itself the test. The run below is the evidence that it executes and passes.

| Branch | Anchor | Classification | Observed |
|---|---|---|---|
| B1 | range at 34:2 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B2 | if at 44:4 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B3 | if at 47:4 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B4 | if at 54:2 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
