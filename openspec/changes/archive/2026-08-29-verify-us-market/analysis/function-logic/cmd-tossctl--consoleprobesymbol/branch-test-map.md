# Branch Test Map: `consoleProbeSymbol`

- Source: `cmd/tossctl/console.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if verifylive.NormalizeMarket(market) != verifylive.MarketUS {` (cmd/tossctl/console.go:81) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
