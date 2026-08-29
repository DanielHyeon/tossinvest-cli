# Branch Test Map: `marketOf`

- Source: `internal/console/pages.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if v := strings.TrimSpace(r.URL.Query().Get("market")); v != "" {` (internal/console/pages.go:127) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
