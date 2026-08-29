# Branch Test Map: `TestDecliningSendsNothingAndStopsTheRun`

- Source: `internal/console/console_test.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if resp.StatusCode != http.StatusOK {` (internal/console/console_test.go:416) | 이 함수 자체가 테스트다 | yes | yes |
| B2 | `if n := h.broker.mutationCount(); n != 0 {` (internal/console/console_test.go:421) | 이 함수 자체가 테스트다 | yes | yes |
| B3 | `if !final.Summary.Halted {` (internal/console/console_test.go:424) | 이 함수 자체가 테스트다 | yes | yes |
| B4 | `if len(h.broker.placements()) != 0 {` (internal/console/console_test.go:427) | 이 함수 자체가 테스트다 | yes | yes |
| B5 | `if n := verifylive.StepCount(entries); n != 0 {` (internal/console/console_test.go:431) | 이 함수 자체가 테스트다 | yes | yes |
| B6 | `if len(entries) == 0 {` (internal/console/console_test.go:434) | 이 함수 자체가 테스트다 | yes | yes |
