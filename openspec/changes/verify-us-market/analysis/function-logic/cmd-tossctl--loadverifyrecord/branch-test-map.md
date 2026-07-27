# Branch Test Map: `loadVerifyRecord`

- Source: `cmd/tossctl/verify.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if err != nil {` (cmd/tossctl/verify.go:488) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B2 | `if err != nil {` (cmd/tossctl/verify.go:492) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
