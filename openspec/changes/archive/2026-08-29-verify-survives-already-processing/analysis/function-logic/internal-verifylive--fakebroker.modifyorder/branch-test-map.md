# Branch Test Map: `fakeBroker.ModifyOrder`

- Source: `internal/verifylive/fake_broker_test.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if f.modifyAlreadyProcessing > 0 {` (internal/verifylive/fake_broker_test.go:447) | 이 함수 자체가 테스트 하네스다 | yes | yes |
| B2 | `if !ok {` (internal/verifylive/fake_broker_test.go:456) | 이 함수 자체가 테스트 하네스다 | yes | yes |
| B3 | `if intent.Price != nil {` (internal/verifylive/fake_broker_test.go:463) | 이 함수 자체가 테스트 하네스다 | yes | yes |
| B4 | `if intent.Quantity != nil {` (internal/verifylive/fake_broker_test.go:466) | 이 함수 자체가 테스트 하네스다 | yes | yes |
| B5 | `if f.amendIssuesNewID {` (internal/verifylive/fake_broker_test.go:472) | 이 함수 자체가 테스트 하네스다 | yes | yes |
