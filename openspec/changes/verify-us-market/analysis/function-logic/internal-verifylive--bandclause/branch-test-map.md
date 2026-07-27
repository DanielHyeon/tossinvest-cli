# Branch Test Map: `bandClause`

- Source: `internal/verifylive/plan.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if SameMarket(market, MarketKR) {` (internal/verifylive/plan.go:119) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
