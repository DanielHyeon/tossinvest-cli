# Branch Test Map: `StepLabel`

- Source: `internal/verifylive/korean.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if id == StepCleanup {` (internal/verifylive/korean.go:66) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B2 | `if label, ok := stepLabels[id]; ok {` (internal/verifylive/korean.go:69) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
