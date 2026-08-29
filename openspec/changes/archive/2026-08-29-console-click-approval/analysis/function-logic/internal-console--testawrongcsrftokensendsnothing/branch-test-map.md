# Branch Test Map: `TestAWrongCSRFTokenSendsNothing`

- Source: `internal/console/console_test.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if resp.StatusCode != http.StatusForbidden {` (internal/console/console_test.go:371) | 이 함수 자체가 테스트다 | yes | yes |
| B2 | `if n := h.broker.mutationCount(); n != 0 {` (internal/console/console_test.go:374) | 이 함수 자체가 테스트다 | yes | yes |
| B3 | `if !h.currentRun().snapshot().Awaiting {` (internal/console/console_test.go:377) | 이 함수 자체가 테스트다 | yes | yes |
