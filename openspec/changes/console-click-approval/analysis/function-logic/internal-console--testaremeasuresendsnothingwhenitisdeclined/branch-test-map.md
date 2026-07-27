# Branch Test Map: `TestARemeasureSendsNothingWhenItIsDeclined`

- Source: `internal/console/remeasure_test.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if n := h.broker.mutationCount(); n != 0 {` (internal/console/remeasure_test.go:198) | 이 함수 자체가 테스트다 | yes | yes |
| B2 | `if !final.Summary.Halted {` (internal/console/remeasure_test.go:201) | 이 함수 자체가 테스트다 | yes | yes |
| B3 | `if got := verifylive.RedoSet(loadRecord(t, h.record)); len(got) != 1 {` (internal/console/remeasure_test.go:205) | 이 함수 자체가 테스트다 | yes | yes |
