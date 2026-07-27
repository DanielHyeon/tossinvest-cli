# Branch Test Map: `TestASecondVerificationInTheSameProcessIsRefused`

- Source: `internal/console/console_test.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if !strings.Contains(page, "콘솔을 종료") {` (internal/console/console_test.go:551) | 이 함수 자체가 테스트다 | yes | yes |
| B2 | `if run := h.currentRun(); !run.finished() {` (internal/console/console_test.go:556) | 이 함수 자체가 테스트다 | yes | yes |
| B3 | `if h.broker.mutationCount() != before {` (internal/console/console_test.go:559) | 이 함수 자체가 테스트다 | yes | yes |
