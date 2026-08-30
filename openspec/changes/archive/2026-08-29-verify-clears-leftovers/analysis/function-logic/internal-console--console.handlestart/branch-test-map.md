# Branch Test Map: `Console.handleStart`

- Source: `internal/console/pages.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if r.PostFormValue("mode") == startModeRedo {` (internal/console/pages.go:197) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B2 | `} else if snap := c.readVerify(market); snap.Present && len(snap.Pending) == 0 &&` (internal/console/pages.go:209) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B3 | `if err != nil {` (internal/console/pages.go:199) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B4 | `if len(set) == 0 {` (internal/console/pages.go:203) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B5 | `} else if snap := c.readVerify(market); snap.Present && len(snap.Pending) == 0 &&` (internal/console/pages.go:209) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B6 | `if len(snap.Redo) > 0 {` (internal/console/pages.go:222) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B7 | `if _, err := c.startRun(market, redo); err != nil {` (internal/console/pages.go:228) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
