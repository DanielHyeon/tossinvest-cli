# Branch Test Map: `Runner.writeBanner`

- Source: `internal/verifylive/runner.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if r.confirmEach {` (internal/verifylive/runner.go:347) | TestTheApprovalScreenAsksForNoTypedString(console), TestApprovalIsOneClickWithNothingTyped(console) | yes | yes |
| B2 | `if r.approvalChannel == ApprovalChannelConsoleClick {` (internal/verifylive/runner.go:357) | TestTheApprovalScreenAsksForNoTypedString(console), TestApprovalIsOneClickWithNothingTyped(console) | yes | yes |
