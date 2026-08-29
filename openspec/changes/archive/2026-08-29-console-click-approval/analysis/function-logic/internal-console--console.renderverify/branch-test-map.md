# Branch Test Map: `Console.renderVerify`

- Source: `internal/console/pages.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if run := c.currentRun(); run != nil {` (internal/console/pages.go:138) | TestTheApprovalScreenAsksForNoTypedString, TestTheApprovedFlowRunsExactlyTheApprovedBatch | yes | yes |
| B2 | `if v.Batch != nil {` (internal/console/pages.go:141) | TestTheApprovalScreenAsksForNoTypedString, TestTheApprovedFlowRunsExactlyTheApprovedBatch | yes | yes |
