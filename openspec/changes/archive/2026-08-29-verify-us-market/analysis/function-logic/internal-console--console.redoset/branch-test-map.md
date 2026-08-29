# Branch Test Map: `Console.redoSet`

- Source: `internal/console/data.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if err != nil {` (internal/console/data.go:325) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
