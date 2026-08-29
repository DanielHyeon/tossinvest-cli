# Branch Test Map: `TestARemeasureRunsOnlyTheFailedStepAndStillAsksForANewNonce`

- Source: `internal/console/remeasure_test.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if len(view.Redo) != 1 \|\| view.Redo[0] != verifylive.StepOrderCancel {` (internal/console/remeasure_test.go:139) | 이 함수 자체가 테스트다 | yes | yes |
| B2 | `if view.Batch == nil \|\| strings.TrimSpace(view.Batch.Nonce) == "" {` (internal/console/remeasure_test.go:142) | 이 함수 자체가 테스트다 | yes | yes |
| B3 | `if n := h.broker.mutationCount(); n != 0 {` (internal/console/remeasure_test.go:145) | 이 함수 자체가 테스트다 | yes | yes |
| B4 | `for _, m := range view.Batch.Plan.Mutations {` (internal/console/remeasure_test.go:149) | 이 함수 자체가 테스트다 | yes | yes |
| B5 | `if m.Step != verifylive.StepOrderCancel {` (internal/console/remeasure_test.go:150) | 이 함수 자체가 테스트다 | yes | yes |
| B6 | `if got := len(h.broker.placements()); got != 1 {` (internal/console/remeasure_test.go:158) | 이 함수 자체가 테스트다 | yes | yes |
| B7 | `for _, o := range final.Summary.Outcomes {` (internal/console/remeasure_test.go:163) | 이 함수 자체가 테스트다 | yes | yes |
| B8 | `if o.AlreadySettled {` (internal/console/remeasure_test.go:164) | 이 함수 자체가 테스트다 | yes | yes |
| B9 | `if !ran[verifylive.StepOrderCancel] {` (internal/console/remeasure_test.go:170) | 이 함수 자체가 테스트다 | yes | yes |
| B10 | `for _, step := range verifylive.Steps() {` (internal/console/remeasure_test.go:173) | 이 함수 자체가 테스트다 | yes | yes |
| B11 | `if step.ID == verifylive.StepOrderCancel {` (internal/console/remeasure_test.go:174) | 이 함수 자체가 테스트다 | yes | yes |
| B12 | `if !settled[step.ID] {` (internal/console/remeasure_test.go:177) | 이 함수 자체가 테스트다 | yes | yes |
