# Branch Test Map: `TestAnExpiredNonceSendsNothing`

- Source: `internal/console/console_test.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if n := h.broker.mutationCount(); n != 0 {` (internal/console/console_test.go:460) | 이 함수 자체가 테스트다 | yes | yes |
| B2 | `if !final.Summary.Halted {` (internal/console/console_test.go:463) | 이 함수 자체가 테스트다 | yes | yes |
