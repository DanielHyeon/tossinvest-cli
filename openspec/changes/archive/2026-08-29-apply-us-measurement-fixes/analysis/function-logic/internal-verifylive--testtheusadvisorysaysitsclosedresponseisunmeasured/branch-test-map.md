# Branch Test Map: `TestTheUSAdvisorySaysItsClosedResponseIsUnmeasured`

- Source: `internal/verifylive/us_market_test.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if strings.Contains(us.Detail, "order-hours-closed") {` (internal/verifylive/us_market_test.go:216) | 이 함수 자체가 테스트다 | yes | yes |
| B2 | `if !strings.Contains(us.Detail, "미측정") {` (internal/verifylive/us_market_test.go:219) | 이 함수 자체가 테스트다 | yes | yes |
| B3 | `if !strings.Contains(kr.Detail, "order-hours-closed") {` (internal/verifylive/us_market_test.go:223) | 이 함수 자체가 테스트다 | yes | yes |
