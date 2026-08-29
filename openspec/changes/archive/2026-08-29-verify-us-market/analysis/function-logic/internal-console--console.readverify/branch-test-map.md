# Branch Test Map: `Console.readVerify`

- Source: `internal/console/data.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if strings.TrimSpace(v.Record) == "" {` (internal/console/data.go:290) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B2 | `if err != nil {` (internal/console/data.go:294) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B3 | `for _, s := range progress.Steps {` (internal/console/data.go:308) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B4 | `if s.Verdict.Terminal() {` (internal/console/data.go:309) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
