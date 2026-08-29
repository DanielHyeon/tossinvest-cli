# Branch Test Map: `Console.verifyRecord`

- Source: `internal/console/data.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if verifylive.NormalizeMarket(market) == verifylive.MarketUS {` (internal/console/data.go:338) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
