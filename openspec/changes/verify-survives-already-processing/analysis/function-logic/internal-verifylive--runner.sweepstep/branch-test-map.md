# Branch Test Map: `Runner.sweepStep`

- Source: `internal/verifylive/cleanup.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if sr.abort != nil \|\| ctx.Err() != nil {` (internal/verifylive/cleanup.go:209) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B2 | `for _, a := range Outstanding(mine) {` (internal/verifylive/cleanup.go:213) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B3 | `if a.Kind != KindOrder {` (internal/verifylive/cleanup.go:214) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
| B4 | `if err := r.cancelOrder(ctx, sr, a.ID, a.Symbol, "이 단계가 실패해 남긴 주문 — 다음 단계의 노출 상한을 비운다"); err != nil {` (internal/verifylive/cleanup.go:219) | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 | yes | yes |
