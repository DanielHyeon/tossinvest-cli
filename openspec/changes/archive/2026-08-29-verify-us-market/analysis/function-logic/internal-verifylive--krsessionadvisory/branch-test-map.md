# Branch Test Map: `KRSessionAdvisory`

- Source: `internal/verifylive/hours.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `switch {` (internal/verifylive/hours.go:74) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B2 | `case weekend:` (internal/verifylive/hours.go:75) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B3 | `case hhmm >= krRegularOpen && hhmm < krRegularClose:` (internal/verifylive/hours.go:80) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B4 | `default:` (internal/verifylive/hours.go:86) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
