# Branch Test Map: `Runner.runCleanup`

- Source: `internal/verifylive/cleanup.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `for _, a := range targets {` (internal/verifylive/cleanup.go:163) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B2 | `switch a.Kind {` (internal/verifylive/cleanup.go:165) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B3 | `case KindOrder:` (internal/verifylive/cleanup.go:166) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B4 | `case KindConditional:` (internal/verifylive/cleanup.go:168) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B5 | `default:` (internal/verifylive/cleanup.go:170) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B6 | `if err == nil {` (internal/verifylive/cleanup.go:173) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B7 | `if first == nil {` (internal/verifylive/cleanup.go:176) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B8 | `if errors.Is(err, ErrOutsidePlan) \|\| errors.Is(err, context.Canceled) \|\|` (internal/verifylive/cleanup.go:179) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
