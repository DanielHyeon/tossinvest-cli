# Branch Test Map: `Console.handleReportJSON`

- Source: `internal/console/pages.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if err != nil {` (internal/console/pages.go:316) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B2 | `if err != nil {` (internal/console/pages.go:321) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
