# Branch Test Map: `Console.renderVerify`

- Source: `internal/console/pages.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if run := c.currentRun(); run != nil {` (internal/console/pages.go:153) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B2 | `if v.Batch != nil {` (internal/console/pages.go:156) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
