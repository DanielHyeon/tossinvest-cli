# Branch Test Map: `Console.startRun`

- Source: `internal/console/run.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if c.spent {` (internal/console/run.go:305) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B2 | `if c.run != nil && !c.run.finished() {` (internal/console/run.go:309) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
