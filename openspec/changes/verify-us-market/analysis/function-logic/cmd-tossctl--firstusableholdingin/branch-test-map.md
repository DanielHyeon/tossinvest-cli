# Branch Test Map: `firstUsableHoldingIn`

- Source: `cmd/tossctl/verify.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if err != nil {` (cmd/tossctl/verify.go:421) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B2 | `for _, p := range positions {` (cmd/tossctl/verify.go:424) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B3 | `if p.Quantity < verifylive.MinQuantity {` (cmd/tossctl/verify.go:425) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B4 | `if !verifylive.SameMarket(verifylive.MarketOf(p.Symbol), market) {` (cmd/tossctl/verify.go:428) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
