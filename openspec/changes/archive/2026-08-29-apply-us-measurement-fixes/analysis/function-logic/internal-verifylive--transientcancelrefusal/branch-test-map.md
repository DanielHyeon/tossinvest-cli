# Branch Test Map: `transientCancelRefusal`

- Source: `internal/verifylive/mutate.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if !errors.As(err, &apiErr) \|\| apiErr.Code != http.StatusConflict {` (internal/verifylive/mutate.go:370) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B2 | `if jsonErr := json.Unmarshal([]byte(apiErr.Body), &body); jsonErr != nil {` (internal/verifylive/mutate.go:381) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B3 | `if body.Error.Code != "already-processing" {` (internal/verifylive/mutate.go:384) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B4 | `if wait <= 0 {` (internal/verifylive/mutate.go:388) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
| B5 | `if wait > CancelRetryMaxWait {` (internal/verifylive/mutate.go:391) | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 | yes | yes |
