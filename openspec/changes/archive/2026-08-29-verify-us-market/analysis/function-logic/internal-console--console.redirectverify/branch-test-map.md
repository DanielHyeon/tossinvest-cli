# Branch Test Map: `Console.redirectVerify`

- Source: `internal/console/pages.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if strings.TrimSpace(notice) != "" {` (internal/console/pages.go:284) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
