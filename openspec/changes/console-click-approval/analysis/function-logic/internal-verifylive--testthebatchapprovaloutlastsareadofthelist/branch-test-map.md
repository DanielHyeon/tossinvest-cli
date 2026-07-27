# Branch Test Map: `TestTheBatchApprovalOutlastsAReadOfTheList`

- Source: `internal/verifylive/confirm_test.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if BatchApprovalTTL <= ConfirmationTTL {` (internal/verifylive/confirm_test.go:203) | 이 함수 자체가 테스트다 | yes | yes |
| B2 | `if BatchApprovalTTL > 15*time.Minute {` (internal/verifylive/confirm_test.go:207) | 이 함수 자체가 테스트다 | yes | yes |
