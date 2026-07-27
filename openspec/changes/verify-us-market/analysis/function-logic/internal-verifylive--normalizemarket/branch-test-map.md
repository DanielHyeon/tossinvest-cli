# Branch Test Map: `NormalizeMarket`

- Source: `internal/verifylive/pricing.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `switch {` (internal/verifylive/pricing.go:82) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B2 | `case strings.EqualFold(strings.TrimSpace(market), MarketUS):` (internal/verifylive/pricing.go:83) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B3 | `default:` (internal/verifylive/pricing.go:85) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
