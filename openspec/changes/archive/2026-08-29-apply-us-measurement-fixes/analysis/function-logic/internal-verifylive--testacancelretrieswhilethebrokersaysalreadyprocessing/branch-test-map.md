# Branch Test Map: `TestACancelRetriesWhileTheBrokerSaysAlreadyProcessing`

- Source: `internal/verifylive/us_market_test.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if _, err := h.run(Options{HoldingSymbol: "005930"}); err != nil {` (internal/verifylive/us_market_test.go:240) | 이 함수 자체가 테스트다 | yes | yes |
| B2 | `if anyFailureMentioning(h.entries(), "already-processing") {` (internal/verifylive/us_market_test.go:246) | 이 함수 자체가 테스트다 | yes | yes |
| B3 | `for _, a := range Outstanding(h.entries()) {` (internal/verifylive/us_market_test.go:249) | 이 함수 자체가 테스트다 | yes | yes |
| B4 | `if !a.Deliberate {` (internal/verifylive/us_market_test.go:252) | 이 함수 자체가 테스트다 | yes | yes |
| B5 | `if !anyObservation(h.entries(), "order.cancel.retries") {` (internal/verifylive/us_market_test.go:257) | 이 함수 자체가 테스트다 | yes | yes |
