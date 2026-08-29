# Branch Test Map: `harness.observation`

- Source: `internal/verifylive/fake_broker_test.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `for _, e := range h.entries() {` (internal/verifylive/fake_broker_test.go:753) | 이 함수 자체가 테스트다 | yes | yes |
| B2 | `if e.StepID != id {` (internal/verifylive/fake_broker_test.go:754) | 이 함수 자체가 테스트다 | yes | yes |
| B3 | `for _, o := range e.Observations {` (internal/verifylive/fake_broker_test.go:757) | 이 함수 자체가 테스트다 | yes | yes |
| B4 | `if o.Key == key {` (internal/verifylive/fake_broker_test.go:758) | 이 함수 자체가 테스트다 | yes | yes |
