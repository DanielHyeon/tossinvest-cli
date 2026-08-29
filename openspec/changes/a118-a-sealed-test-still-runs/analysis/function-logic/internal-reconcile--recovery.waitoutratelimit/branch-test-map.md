# Branch Test Map: `Recovery.waitOutRateLimit`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 예산 경계에서 정확히 멈춘다 | `TestRateLimitBudgetStopsExactlyAtTheBoundary` `a102_recovery_rate_limit_test.go:416` | yes (뮤테이션 M1) | yes |
| B2 | 예산이 backoff 하나도 못 덮으면 다른 사유를 말한다 | `TestBudgetTooSmallForOneBackoffSaysSo` `a102_recovery_rate_limit_test.go:486` | no | yes |
| B3 | 백오프 대기 중 취소는 즉시 통과한다 | `TestRateLimitBackoffStopsOnContextCancel` `a102_recovery_rate_limit_test.go:345` | no | yes |
