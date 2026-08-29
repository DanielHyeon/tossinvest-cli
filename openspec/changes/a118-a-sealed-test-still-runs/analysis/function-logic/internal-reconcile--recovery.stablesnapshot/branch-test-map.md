# Branch Test Map: `Recovery.stableSnapshot`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 루프가 MaxAttempts 로 돈다 | `TestRateLimitDoesNotConsumeAStabilisationAttempt` `a102_recovery_rate_limit_test.go:180` | no | yes |
| B2 | Collect 가 오류를 낸다 | `TestRateLimitedCollectDoesNotEndRecovery` `a102_recovery_rate_limit_test.go:150` | no | yes |
| B3 | rate limit 이 아닌 오류는 즉시 실패 | `TestNonRateLimitedCollectStillFailsImmediately` `a102_recovery_rate_limit_test.go:265` | no | yes |
| B4 | 예산이 소진되면 그 오류로 끝난다 | `TestRateLimitWaitBudgetExhaustionFailsClosed` `a102_recovery_rate_limit_test.go:302` | yes (뮤테이션 M1) | yes |
| B5 | 연속 일치가 나오면 스냅샷을 돌려준다 | `TestRecoveryReleasesTheLatchOnlyWhenItCompletes` `recovery_test.go:186` | no | yes |
| B6 | 안정되지 않으면 fail-closed | `TestRecoveryFailsClosedWhenTheAccountWillNotSettle` `recovery_test.go:302` | no | yes |
| B7 | 간격 대기 중 취소는 오류로 통과 | `TestRateLimitBackoffStopsOnContextCancel` `a102_recovery_rate_limit_test.go:345` | no | yes |
