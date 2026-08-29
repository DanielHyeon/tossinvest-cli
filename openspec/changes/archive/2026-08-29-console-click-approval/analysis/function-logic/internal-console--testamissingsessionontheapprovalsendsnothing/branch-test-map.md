# Branch Test Map: `TestAMissingSessionOnTheApprovalSendsNothing`

- Source: `internal/console/console_test.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if err != nil {` (internal/console/console_test.go:392) | 이 함수 자체가 테스트다 | yes | yes |
| B2 | `if resp.StatusCode != http.StatusForbidden {` (internal/console/console_test.go:397) | 이 함수 자체가 테스트다 | yes | yes |
| B3 | `if n := h.broker.mutationCount(); n != 0 {` (internal/console/console_test.go:400) | 이 함수 자체가 테스트다 | yes | yes |
