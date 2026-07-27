# Branch Test Map: `Runner.amendOrder`

- Source: `internal/verifylive/mutate.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if !sendQuantity {` (internal/verifylive/mutate.go:317) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B2 | `if err := r.gate(sr, request{` (internal/verifylive/mutate.go:327) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B3 | `if sendQuantity {` (internal/verifylive/mutate.go:335) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B4 | `if err != nil {` (internal/verifylive/mutate.go:341) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B5 | `if current == "" {` (internal/verifylive/mutate.go:345) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B6 | `if current != orderID {` (internal/verifylive/mutate.go:348) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
