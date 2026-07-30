# Branch Test Map: `transientRefusal`

- Source: `internal/verifylive/mutate.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if !errors.As(err, &apiErr) \|\| apiErr.Code != http.StatusConflict {` (internal/verifylive/mutate.go:373) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B2 | `if jsonErr := json.Unmarshal([]byte(apiErr.Body), &body); jsonErr != nil {` (internal/verifylive/mutate.go:384) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B3 | `if body.Error.Code != "already-processing" {` (internal/verifylive/mutate.go:387) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B4 | `if wait <= 0 {` (internal/verifylive/mutate.go:391) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B5 | `if wait > TransientRetryMaxWait {` (internal/verifylive/mutate.go:394) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
