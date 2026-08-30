# Branch Test Map: `TestTheConditionalListUsesAnAllowedStatusFilter`

- Source: `internal/verifylive/us_market_test.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if _, err := h.run(Options{HoldingSymbol: "005930"}); err != nil {` (internal/verifylive/us_market_test.go:312) | 이 함수 자체가 테스트다 | yes | yes |
| B2 | `if !observationEquals(t, h.entries(), StepConditionalRegister, "conditional.list_by_status.ok", "true") {` (internal/verifylive/us_market_test.go:315) | 이 함수 자체가 테스트다 | yes | yes |
