# Branch Test Map: `TestAPlacementIsNeverRetried`

- Source: `internal/verifylive/us_market_test.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if _, err := h.run(Options{HoldingSymbol: "005930"}); err != nil {` (internal/verifylive/us_market_test.go:292) | 이 함수 자체가 테스트다 | yes | yes |
| B2 | `if !anyFailureMentioning(h.entries(), "already-processing") {` (internal/verifylive/us_market_test.go:295) | 이 함수 자체가 테스트다 | yes | yes |
