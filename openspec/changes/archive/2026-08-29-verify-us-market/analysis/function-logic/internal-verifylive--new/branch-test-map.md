# Branch Test Map: `New`

- Source: `internal/verifylive/runner.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if o.Broker == nil {` (internal/verifylive/runner.go:160) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B2 | `if o.Recorder == nil {` (internal/verifylive/runner.go:163) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B3 | `if o.Confirm == nil {` (internal/verifylive/runner.go:166) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B4 | `if o.ConfirmBatch == nil {` (internal/verifylive/runner.go:169) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B5 | `if strings.TrimSpace(o.AccountRef) == "" {` (internal/verifylive/runner.go:173) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B6 | `if err != nil {` (internal/verifylive/runner.go:177) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B7 | `if r.approvalChannel == "" {` (internal/verifylive/runner.go:202) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B8 | `if r.out == nil {` (internal/verifylive/runner.go:205) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B9 | `if r.now == nil {` (internal/verifylive/runner.go:208) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B10 | `if r.sleep == nil {` (internal/verifylive/runner.go:211) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B11 | `if r.maxSellQuantity <= 0 {` (internal/verifylive/runner.go:214) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B12 | `if r.ttlWait <= 0 {` (internal/verifylive/runner.go:217) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B13 | `if r.process.InstanceID == "" {` (internal/verifylive/runner.go:220) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B14 | `for _, id := range o.Redo {` (internal/verifylive/runner.go:223) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
