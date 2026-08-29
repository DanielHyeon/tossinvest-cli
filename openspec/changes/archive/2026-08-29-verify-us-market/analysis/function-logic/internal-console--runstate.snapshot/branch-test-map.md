# Branch Test Map: `runState.snapshot`

- Source: `internal/console/run.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if r.partial != "" {` (internal/console/run.go:160) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B2 | `if r.err != nil {` (internal/console/run.go:163) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
