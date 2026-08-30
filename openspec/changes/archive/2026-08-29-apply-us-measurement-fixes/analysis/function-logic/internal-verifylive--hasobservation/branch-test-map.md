# Branch Test Map: `hasObservation`

- Source: `internal/verifylive/fake_broker_test.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if !ok {` (internal/verifylive/fake_broker_test.go:830) | 이 함수 자체가 테스트다 | yes | yes |
| B2 | `for _, o := range e.Observations {` (internal/verifylive/fake_broker_test.go:833) | 이 함수 자체가 테스트다 | yes | yes |
| B3 | `if o.Key == key {` (internal/verifylive/fake_broker_test.go:834) | 이 함수 자체가 테스트다 | yes | yes |
