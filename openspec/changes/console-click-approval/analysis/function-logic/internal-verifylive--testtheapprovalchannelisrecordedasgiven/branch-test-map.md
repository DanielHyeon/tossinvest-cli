# Branch Test Map: `TestTheApprovalChannelIsRecordedAsGiven`

- Source: `internal/verifylive/plan_test.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if _, err := h.run(Options{HoldingSymbol: "005930", ApprovalChannel: channel}); err != nil {` (internal/verifylive/plan_test.go:808) | 이 함수 자체가 테스트다 | yes | yes |
| B2 | `for _, o := range approvalEntry(t, h.entries()).Observations {` (internal/verifylive/plan_test.go:811) | 이 함수 자체가 테스트다 | yes | yes |
| B3 | `if o.Key == "approval.model" {` (internal/verifylive/plan_test.go:812) | 이 함수 자체가 테스트다 | yes | yes |
| B4 | `if got := detailFor(t, ""); got != ApprovalChannelTyped {` (internal/verifylive/plan_test.go:820) | 이 함수 자체가 테스트다 | yes | yes |
| B5 | `if got := detailFor(t, ApprovalChannelConsoleClick); got != ApprovalChannelConsoleClick {` (internal/verifylive/plan_test.go:823) | 이 함수 자체가 테스트다 | yes | yes |
