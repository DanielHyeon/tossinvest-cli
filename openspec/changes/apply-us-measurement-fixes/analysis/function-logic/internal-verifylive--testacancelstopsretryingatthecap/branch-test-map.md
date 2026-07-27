# Branch Test Map: `TestACancelStopsRetryingAtTheCap`

- Source: `internal/verifylive/us_market_test.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if _, err := h.run(Options{HoldingSymbol: "005930"}); err != nil {` (internal/verifylive/us_market_test.go:269) | 이 함수 자체가 테스트다 | yes | yes |
| B2 | `if !anyFailureMentioning(h.entries(), "already-processing") {` (internal/verifylive/us_market_test.go:272) | 이 함수 자체가 테스트다 | yes | yes |
| B3 | `if n := broker.countRequests("POST /orders/"); n > 12 {` (internal/verifylive/us_market_test.go:275) | 이 함수 자체가 테스트다 | yes | yes |
| B4 | `if len(Outstanding(h.entries())) == 0 {` (internal/verifylive/us_market_test.go:278) | 이 함수 자체가 테스트다 | yes | yes |
