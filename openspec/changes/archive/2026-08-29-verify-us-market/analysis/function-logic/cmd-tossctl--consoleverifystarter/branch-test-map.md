# Branch Test Map: `consoleVerifyStarter`

- Source: `cmd/tossctl/console.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if err != nil {` (cmd/tossctl/console.go:390) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B2 | `if err != nil {` (cmd/tossctl/console.go:394) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B3 | `if err != nil {` (cmd/tossctl/console.go:398) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B4 | `if err != nil {` (cmd/tossctl/console.go:403) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B5 | `if err != nil {` (cmd/tossctl/console.go:425) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B6 | `if runErr != nil && (errors.Is(runErr, context.Canceled) \|\| errors.Is(runErr, context.DeadlineExceeded)) {` (cmd/tossctl/console.go:434) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
