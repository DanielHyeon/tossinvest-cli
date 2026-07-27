# Branch Test Map: `resolveVerifyRecordFor`

- Source: `cmd/tossctl/verify.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if trimmed := strings.TrimSpace(override); trimmed != "" {` (cmd/tossctl/verify.go:518) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B2 | `if root != nil && strings.TrimSpace(root.configDir) != "" {` (cmd/tossctl/verify.go:522) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B3 | `if err != nil {` (cmd/tossctl/verify.go:526) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
