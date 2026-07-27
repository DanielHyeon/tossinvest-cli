# Branch Test Map: `Runner.stepConditionalRegister`

- Source: `internal/verifylive/steps.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if err != nil {` (internal/verifylive/steps.go:549) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B2 | `if sellable < MinQuantity {` (internal/verifylive/steps.go:552) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B3 | `if err != nil {` (internal/verifylive/steps.go:559) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B4 | `if err != nil {` (internal/verifylive/steps.go:563) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B5 | `if err != nil {` (internal/verifylive/steps.go:578) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B6 | `switch {` (internal/verifylive/steps.go:591) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B7 | `case isGateError(replayErr):` (internal/verifylive/steps.go:592) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B8 | `case replayErr != nil:` (internal/verifylive/steps.go:594) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B9 | `case replayID == id:` (internal/verifylive/steps.go:597) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B10 | `default:` (internal/verifylive/steps.go:599) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B11 | `if replayID != "" {` (internal/verifylive/steps.go:602) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B12 | `if err := r.cancelConditional(ctx, sr, replayID, symbol,` (internal/verifylive/steps.go:605) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B13 | `if co, err := r.readConditional(ctx, sr, id); err == nil {` (internal/verifylive/steps.go:613) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B14 | `} else {` (internal/verifylive/steps.go:620) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B15 | `if err != nil {` (internal/verifylive/steps.go:633) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B16 | `} else {` (internal/verifylive/steps.go:635) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
