# Branch Test Map: `fakeBroker.ConditionalOrders`

- Source: `internal/verifylive/fake_broker_test.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if f.rejectBadConditionalStatus && status != "" && status != "OPEN" && status != "CLOSED" {` (internal/verifylive/fake_broker_test.go:329) | 이 함수 자체가 테스트다 | yes | yes |
| B2 | `for _, c := range f.conds {` (internal/verifylive/fake_broker_test.go:336) | 이 함수 자체가 테스트다 | yes | yes |
| B3 | `if status == "OPEN" && c.Status != "WATCHING" {` (internal/verifylive/fake_broker_test.go:339) | 이 함수 자체가 테스트다 | yes | yes |
| B4 | `if status != "" && status != "OPEN" && status != "CLOSED" && c.Status != status {` (internal/verifylive/fake_broker_test.go:342) | 이 함수 자체가 테스트다 | yes | yes |
| B5 | `if symbol != "" && c.Symbol != symbol {` (internal/verifylive/fake_broker_test.go:345) | 이 함수 자체가 테스트다 | yes | yes |
