# Branch Test Map: `PositionPolicyCommandService.verifyCapability`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | malformed tokens are invalid | `TestPositionPolicyCapabilityRejectsExpiryOtherEngineStaleAndConcurrentReplay` | yes | yes |
| B2 | all stored digests are scanned | concurrent replay test | yes | yes |
| B3 | matching digest is recorded without early return | concurrent replay test | yes | yes |
| B4 | absent/replayed token is invalid | concurrent replay test | yes | yes |
| B5 | instance/domain mismatch consumes authority | capability binding tests | yes | yes |
| B6 | clock rollback is stale and consumed | `TestReadoptCapabilityHasShortFreshnessAndRejectsClockRollback/clock_rollback_is_consumed` | yes | yes |
| B7 | exact 15s succeeds; 15s+1ns is stale and consumed | `TestReadoptCapabilityHasShortFreshnessAndRejectsClockRollback` | yes | yes |
| B8 | use before 3-second danger delay is rejected without consumption | `TestPositionPolicyDangerousCapabilityRequiresServerSideConfirmation` | yes | yes |
| B9 | expired grant is consumed | `TestPositionPolicyCapabilityRejectsExpiryOtherEngineStaleAndConcurrentReplay` | yes | yes |
