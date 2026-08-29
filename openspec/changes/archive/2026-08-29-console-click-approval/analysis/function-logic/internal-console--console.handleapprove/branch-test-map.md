# Branch Test Map: `Console.handleApprove`

- Source: `internal/console/pages.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if run == nil {` (internal/console/pages.go:214) | TestApprovalIsOneClickWithNothingTyped, TestAnExpiredApprovalSendsNothing, TestAWrongCSRFTokenSendsNothing, TestAMissingSessionOnTheApprovalSendsNothing, TestTheApprovalWindowIsJudgedByVerifylive | yes | yes |
| B2 | `if !view.Awaiting \|\| view.Batch == nil {` (internal/console/pages.go:219) | TestApprovalIsOneClickWithNothingTyped, TestAnExpiredApprovalSendsNothing, TestAWrongCSRFTokenSendsNothing, TestAMissingSessionOnTheApprovalSendsNothing, TestTheApprovalWindowIsJudgedByVerifylive | yes | yes |
| B3 | `if view.Batch.Expired(c.now()) {` (internal/console/pages.go:225) | TestApprovalIsOneClickWithNothingTyped, TestAnExpiredApprovalSendsNothing, TestAWrongCSRFTokenSendsNothing, TestAMissingSessionOnTheApprovalSendsNothing, TestTheApprovalWindowIsJudgedByVerifylive | yes | yes |
| B4 | `if !run.deliver(answer) {` (internal/console/pages.go:228) | TestApprovalIsOneClickWithNothingTyped, TestAnExpiredApprovalSendsNothing, TestAWrongCSRFTokenSendsNothing, TestAMissingSessionOnTheApprovalSendsNothing, TestTheApprovalWindowIsJudgedByVerifylive | yes | yes |
| B5 | `if answer != nil {` (internal/console/pages.go:232) | TestApprovalIsOneClickWithNothingTyped, TestAnExpiredApprovalSendsNothing, TestAWrongCSRFTokenSendsNothing, TestAMissingSessionOnTheApprovalSendsNothing, TestTheApprovalWindowIsJudgedByVerifylive | yes | yes |
