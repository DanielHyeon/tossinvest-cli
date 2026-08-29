# Branch Test Map: `RecordFileName`

- Source: `internal/verifylive/record.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if NormalizeMarket(market) == MarketUS {` (internal/verifylive/record.go:56) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
