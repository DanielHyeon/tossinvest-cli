# Branch Test Map: `SessionAdvisoryFor`

- Source: `internal/verifylive/hours.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if NormalizeMarket(market) != MarketUS {` (internal/verifylive/hours.go:111) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
