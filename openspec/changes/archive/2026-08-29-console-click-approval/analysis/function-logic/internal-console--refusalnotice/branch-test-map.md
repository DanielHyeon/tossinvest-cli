# Branch Test Map: `refusalNotice`

- Source: `internal/console/pages.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `switch {` (internal/console/pages.go:256) | TestAnExpiredApprovalSendsNothing | yes | yes |
| B2 | `case err == nil:` (internal/console/pages.go:257) | TestAnExpiredApprovalSendsNothing | yes | yes |
| B3 | `case strings.Contains(err.Error(), verifylive.ErrConfirmationExpired.Error()):` (internal/console/pages.go:259) | TestAnExpiredApprovalSendsNothing | yes | yes |
| B4 | `default:` (internal/console/pages.go:261) | TestAnExpiredApprovalSendsNothing | yes | yes |
