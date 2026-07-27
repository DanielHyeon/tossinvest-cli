# Branch Test Map: `Runner.preflightStatic`

- Source: `internal/verifylive/runner.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if step.Deferred != "" {` (internal/verifylive/runner.go:512) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B2 | `if step.OptIn != "" && !r.optedIn(step) {` (internal/verifylive/runner.go:518) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B3 | `for _, dep := range step.DependsOn {` (internal/verifylive/runner.go:521) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B4 | `if !passed(dep) {` (internal/verifylive/runner.go:522) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B5 | `if step.NeedsHolding && r.holdingSymbol == "" {` (internal/verifylive/runner.go:526) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B6 | `if symbol := r.mutationSymbol(step); step.Mutates && !SameMarket(MarketOf(symbol), r.market) {` (internal/verifylive/runner.go:531) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
