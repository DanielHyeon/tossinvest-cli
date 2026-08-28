# Branch Test Map: `TestRecordQFinalCampaignFirstLegAcceptsAllStrategyflowLanesPaired`

- Source: `internal/journal/strategyflow_projection_test.go`; file SHA-256 `4897ee72290331ac04e2257d255a1922214b9849ad48877fa0bf8b10d2649156`. AST branch positions are authoritative.
- No coverage profile can cover a `_test.go` function. Each row states what the arm is, and the run that exercised it: `go test -count=1 -tags tossos_testseams -run '^TestRecordQFinalCampaignFirstLegAcceptsAllStrategyflowLanesPaired$' ./internal/journal/`.
- This function is itself the test. The run below is the evidence that it executes and passes.

| Branch | Anchor | Classification | Observed |
|---|---|---|---|
| B1 | range at 179:2 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B2 | if at 185:4 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B3 | if at 189:4 | path arm — taken on the exercised path | exercised by the named run |
| B4 | if at 195:4 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B5 | if at 198:4 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B6 | if at 201:4 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |
| B7 | if at 207:2 | guard arm — reached only when the assertion fails, so a passing run must not enter it | passing run: not entered |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
