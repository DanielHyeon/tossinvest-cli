# Branch Test Map: `anyObservation`

- Source: `internal/verifylive/fake_broker_test.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `for _, e := range entries {` (internal/verifylive/fake_broker_test.go:858) | 이 함수 자체가 테스트다 | yes | yes |
| B2 | `for _, o := range e.Observations {` (internal/verifylive/fake_broker_test.go:859) | 이 함수 자체가 테스트다 | yes | yes |
| B3 | `if o.Key == key {` (internal/verifylive/fake_broker_test.go:860) | 이 함수 자체가 테스트다 | yes | yes |
