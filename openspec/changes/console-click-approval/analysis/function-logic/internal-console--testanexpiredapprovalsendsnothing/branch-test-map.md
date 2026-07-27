# Branch Test Map: `TestAnExpiredApprovalSendsNothing`

- Source: `internal/console/console_test.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if n := h.broker.mutationCount(); n != 0 {` (internal/console/console_test.go:452) | 이 함수 자체가 테스트다 | yes | yes |
| B2 | `if !final.Summary.Halted {` (internal/console/console_test.go:455) | 이 함수 자체가 테스트다 | yes | yes |
