# Branch Test Map: `Runner.amendOrder`

- Source: `internal/verifylive/mutate.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if !sendQuantity {` (internal/verifylive/mutate.go:408) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B2 | `if err := r.gate(sr, request{` (internal/verifylive/mutate.go:418) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B3 | `if sendQuantity {` (internal/verifylive/mutate.go:426) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B4 | `for attempts = 1; ; attempts++ {` (internal/verifylive/mutate.go:439) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B5 | `if err == nil {` (internal/verifylive/mutate.go:443) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B6 | `if !transient \|\| attempts > TransientRetryAttempts {` (internal/verifylive/mutate.go:447) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B7 | `if sleepErr := r.sleep(ctx, wait); sleepErr != nil {` (internal/verifylive/mutate.go:452) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B8 | `if attempts > 1 {` (internal/verifylive/mutate.go:457) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B9 | `if err != nil {` (internal/verifylive/mutate.go:461) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B10 | `if current == "" {` (internal/verifylive/mutate.go:465) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B11 | `if current != orderID {` (internal/verifylive/mutate.go:468) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
