# Branch Test Map: `Runner.Run`

- Source: `internal/verifylive/runner.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if halt, err := r.approveBatch(ctx); err != nil \|\| halt != "" {` (internal/verifylive/runner.go:269) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B2 | `if outcome, err, stop := r.cleanup(ctx); outcome.Step != "" \|\| err != nil {` (internal/verifylive/runner.go:279) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B3 | `if outcome.Step != "" {` (internal/verifylive/runner.go:280) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B4 | `if stop {` (internal/verifylive/runner.go:283) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B5 | `if outcome.Reason == "" {` (internal/verifylive/runner.go:286) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B6 | `for _, step := range Steps() {` (internal/verifylive/runner.go:294) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B7 | `if err := ctx.Err(); err != nil {` (internal/verifylive/runner.go:295) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B8 | `if settled, verdict := r.settled(step.ID); settled {` (internal/verifylive/runner.go:302) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B9 | `if reason, skip := r.preflight(step); skip {` (internal/verifylive/runner.go:313) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B10 | `} else {` (internal/verifylive/runner.go:315) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B11 | `if err := r.recorder.Append(entry); err != nil {` (internal/verifylive/runner.go:321) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B12 | `if sr.verdict == VerdictAwaitingRestart {` (internal/verifylive/runner.go:330) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B13 | `if errors.Is(sr.abort, ErrNotATerminal) {` (internal/verifylive/runner.go:336) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B14 | `if errors.Is(sr.abort, ErrOutsidePlan) {` (internal/verifylive/runner.go:342) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B15 | `if sr.abort != nil && errors.Is(sr.abort, context.Canceled) {` (internal/verifylive/runner.go:351) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B16 | `if leftovers := undeliberate(summary.Outstanding); len(leftovers) > 0 {` (internal/verifylive/runner.go:360) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
