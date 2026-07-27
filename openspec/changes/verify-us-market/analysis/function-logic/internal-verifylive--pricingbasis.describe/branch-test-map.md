# Branch Test Map: `PricingBasis.Describe`

- Source: `internal/verifylive/plan.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `switch b {` (internal/verifylive/plan.go:134) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B2 | `case PriceFarBuy:` (internal/verifylive/plan.go:135) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B3 | `case PriceFarSell:` (internal/verifylive/plan.go:139) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B4 | `case PriceFarStop:` (internal/verifylive/plan.go:143) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B5 | `case PriceOneTickFurther:` (internal/verifylive/plan.go:147) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B6 | `case PriceIdenticalBody:` (internal/verifylive/plan.go:150) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B7 | `default:` (internal/verifylive/plan.go:153) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
