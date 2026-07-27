# Branch Test Map: `runVerifyRun`

- Source: `cmd/tossctl/verify.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if opts.list {` (cmd/tossctl/verify.go:257) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B2 | `if ctx == nil {` (cmd/tossctl/verify.go:263) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B3 | `if err != nil {` (cmd/tossctl/verify.go:273) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B4 | `if err != nil {` (cmd/tossctl/verify.go:277) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B5 | `if steps := verifylive.StepCount(prior); steps > 0 && !opts.resume && len(opts.redo) == 0 {` (cmd/tossctl/verify.go:283) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B6 | `if err != nil {` (cmd/tossctl/verify.go:292) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B7 | `if holding == "" {` (cmd/tossctl/verify.go:297) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B8 | `if market == verifylive.MarketUS && !verifylive.SameMarket(verifylive.MarketOf(symbol), market) {` (cmd/tossctl/verify.go:301) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B9 | `if err != nil {` (cmd/tossctl/verify.go:309) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B10 | `if err != nil {` (cmd/tossctl/verify.go:332) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B11 | `if runErr != nil && (errors.Is(runErr, context.Canceled) \|\| errors.Is(runErr, context.DeadlineExceeded)) {` (cmd/tossctl/verify.go:343) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
