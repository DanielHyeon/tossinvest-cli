# Branch Test Map: `fakeBroker.CancelOrder`

- Source: `internal/verifylive/fake_broker_test.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if f.cancelAlreadyProcessing > 0 {` (internal/verifylive/fake_broker_test.go:425) | 이 함수 자체가 테스트다 | yes | yes |
| B2 | `if !ok {` (internal/verifylive/fake_broker_test.go:430) | 이 함수 자체가 테스트다 | yes | yes |
