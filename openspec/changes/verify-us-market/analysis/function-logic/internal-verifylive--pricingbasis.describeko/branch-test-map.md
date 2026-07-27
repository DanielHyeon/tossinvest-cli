# Branch Test Map: `PricingBasis.DescribeKO`

- Source: `internal/verifylive/plan.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `switch b {` (internal/verifylive/plan.go:164) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B2 | `case PriceFarBuy:` (internal/verifylive/plan.go:165) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B3 | `case PriceFarSell:` (internal/verifylive/plan.go:169) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B4 | `case PriceFarStop:` (internal/verifylive/plan.go:173) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B5 | `case PriceOneTickFurther:` (internal/verifylive/plan.go:176) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B6 | `case PriceIdenticalBody:` (internal/verifylive/plan.go:178) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B7 | `default:` (internal/verifylive/plan.go:181) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
