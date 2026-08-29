# Branch Test Map: `TestTheApprovedFlowRunsExactlyTheApprovedBatch`

- Source: `internal/console/console_test.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if len(view.Batch.Plan.Mutations) == 0 {` (internal/console/console_test.go:470) | 이 함수 자체가 테스트다 | yes | yes |
| B2 | `if strings.Contains(page, view.Batch.Nonce) {` (internal/console/console_test.go:474) | 이 함수 자체가 테스트다 | yes | yes |
| B3 | `if !strings.Contains(shown, flatten(view.Batch.Summary())) {` (internal/console/console_test.go:482) | 이 함수 자체가 테스트다 | yes | yes |
| B4 | `for _, m := range view.Batch.Plan.Mutations {` (internal/console/console_test.go:488) | 이 함수 자체가 테스트다 | yes | yes |
| B5 | `if !strings.Contains(shown, flatten(m.HeadlineKO())) {` (internal/console/console_test.go:489) | 이 함수 자체가 테스트다 | yes | yes |
| B6 | `if !strings.Contains(shown, flatten(m.EndsKO)) {` (internal/console/console_test.go:492) | 이 함수 자체가 테스트다 | yes | yes |
| B7 | `if n := h.broker.mutationCount(); n == 0 {` (internal/console/console_test.go:500) | 이 함수 자체가 테스트다 | yes | yes |
| B8 | `for _, p := range h.broker.placements() {` (internal/console/console_test.go:503) | 이 함수 자체가 테스트다 | yes | yes |
| B9 | `if p.Symbol != "005930" \|\| !strings.EqualFold(p.Side, "buy") \|\| p.Quantity != 1 {` (internal/console/console_test.go:504) | 이 함수 자체가 테스트다 | yes | yes |
| B10 | `if final.Err != "" && strings.Contains(final.Err, verifylive.ErrOutsidePlan.Error()) {` (internal/console/console_test.go:508) | 이 함수 자체가 테스트다 | yes | yes |
| B11 | `if approval.Verdict != verifylive.VerdictPass {` (internal/console/console_test.go:513) | 이 함수 자체가 테스트다 | yes | yes |
| B12 | `if got, want := observation(approval, "approval.plan_digest"), view.Batch.Plan.Digest(); got != want {` (internal/console/console_test.go:516) | 이 함수 자체가 테스트다 | yes | yes |
