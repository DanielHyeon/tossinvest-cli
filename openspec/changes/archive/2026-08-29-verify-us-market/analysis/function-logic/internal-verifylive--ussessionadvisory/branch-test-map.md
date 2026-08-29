# Branch Test Map: `usSessionAdvisory`

- Source: `internal/verifylive/hours.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if err != nil {` (internal/verifylive/hours.go:119) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B2 | `if err != nil {` (internal/verifylive/hours.go:131) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B3 | `if inside {` (internal/verifylive/hours.go:139) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
