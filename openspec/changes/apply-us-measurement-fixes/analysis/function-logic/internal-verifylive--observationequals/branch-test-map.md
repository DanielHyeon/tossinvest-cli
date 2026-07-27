# Branch Test Map: `observationEquals`

- Source: `internal/verifylive/fake_broker_test.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if !ok {` (internal/verifylive/fake_broker_test.go:845) | 이 함수 자체가 테스트다 | yes | yes |
| B2 | `for _, o := range e.Observations {` (internal/verifylive/fake_broker_test.go:848) | 이 함수 자체가 테스트다 | yes | yes |
| B3 | `if o.Key == key {` (internal/verifylive/fake_broker_test.go:849) | 이 함수 자체가 테스트다 | yes | yes |
