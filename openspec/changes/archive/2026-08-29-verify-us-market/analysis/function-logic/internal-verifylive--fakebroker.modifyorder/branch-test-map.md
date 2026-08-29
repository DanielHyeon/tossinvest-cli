# Branch Test Map: `fakeBroker.ModifyOrder`

- Source: `internal/verifylive/fake_broker_test.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if !ok {` (internal/verifylive/fake_broker_test.go:411) | 이 함수 자체가 테스트다 | yes | yes |
| B2 | `if intent.Price != nil {` (internal/verifylive/fake_broker_test.go:418) | 이 함수 자체가 테스트다 | yes | yes |
| B3 | `if intent.Quantity != nil {` (internal/verifylive/fake_broker_test.go:421) | 이 함수 자체가 테스트다 | yes | yes |
| B4 | `if f.amendIssuesNewID {` (internal/verifylive/fake_broker_test.go:427) | 이 함수 자체가 테스트다 | yes | yes |
