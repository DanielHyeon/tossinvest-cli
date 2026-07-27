# Branch Test Map: `Console.drive`

- Source: `internal/console/run.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if steps > 0 {` (internal/console/run.go:341) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
