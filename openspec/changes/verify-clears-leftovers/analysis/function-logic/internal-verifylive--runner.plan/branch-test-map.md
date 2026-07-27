# Branch Test Map: `Runner.Plan`

- Source: `internal/verifylive/plan.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `for _, line := range r.planCleanup() {` (internal/verifylive/plan.go:537) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B2 | `for _, step := range Steps() {` (internal/verifylive/plan.go:542) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B3 | `if settled, verdict := r.settled(step.ID); settled {` (internal/verifylive/plan.go:543) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B4 | `if step.Mutates {` (internal/verifylive/plan.go:544) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B5 | `if reason, skip := r.preflightStatic(step, passed); skip {` (internal/verifylive/plan.go:553) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B6 | `if step.Mutates {` (internal/verifylive/plan.go:554) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B7 | `if !step.Mutates {` (internal/verifylive/plan.go:560) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B8 | `if !ok {` (internal/verifylive/plan.go:566) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B9 | `for _, line := range lines {` (internal/verifylive/plan.go:575) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
