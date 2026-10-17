# Branch Test Map: `pairedStrategyFirstLegResults`

- Source: `internal/app/engine/strategy_first_leg_admission_test.go`; file SHA-256 `7b2eca4d50c6fb8aaec09b552e1576a74329ea353617b084bc259e914119ac41`. AST branch positions are authoritative.
- No coverage profile can cover a `_test.go` function. Each row states what the arm is, and the run that exercised it: `go test -count=1 -tags tossos_testseams -run '^TestStrategyFirstLegAdmissionUsesGuardianOnlyAllEightLanes$' ./internal/app/engine/`.
- This function is itself the test. The run below is the evidence that it executes and passes.

| Branch | Anchor | Classification | Observed |
|---|---|---|---|
| B1 | body at 188 | branchless | exercised by the named run |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
