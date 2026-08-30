# Branch Test Map: `Runner.cancelOrder`

- Source: `internal/verifylive/mutate.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if err := r.gate(sr, request{` (internal/verifylive/mutate.go:297) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B2 | `for attempts = 1; ; attempts++ {` (internal/verifylive/mutate.go:309) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B3 | `if err == nil {` (internal/verifylive/mutate.go:313) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B4 | `if !transient \|\| attempts > CancelRetryAttempts {` (internal/verifylive/mutate.go:317) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B5 | `if sleepErr := r.sleep(ctx, wait); sleepErr != nil {` (internal/verifylive/mutate.go:322) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B6 | `if attempts > 1 {` (internal/verifylive/mutate.go:327) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B7 | `if err != nil {` (internal/verifylive/mutate.go:333) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B8 | `if id := strings.TrimSpace(res.CurrentOrderID); id != "" && id != orderID {` (internal/verifylive/mutate.go:338) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
