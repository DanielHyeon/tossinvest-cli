# Branch Test Map: `PositionPolicyCommandService.Apply`

RED was observed on 2026-08-01 before the capability/confirmation implementation; GREEN was observed with the focused engine suite after implementation.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | invalid, early, expired, cross-engine, or replayed grant is refused | `TestPositionPolicyCapabilityRequiresPreviewIsTimedAndOneTime`, `TestPositionPolicyCapabilityRejectsExpiryOtherEngineStaleAndConcurrentReplay` | yes | yes |
| B2 | RELEASE/READOPT without server-visible checkbox confirmation is refused without consuming the grant | `TestPositionPolicyDangerousCapabilityRequiresServerSideConfirmation` | yes | yes |
| B3 | authoritative current-state read fails after one-time consumption | `TestPositionPolicyCapabilityRejectsExpiryOtherEngineStaleAndConcurrentReplay/current_read_failure_consumes_fail_closed` | yes | yes |
| B4 | current state differs from the exact preview-bound before state | `TestPositionPolicyCapabilityRejectsExpiryOtherEngineStaleAndConcurrentReplay/stale_failure_consumes_fail_closed` | yes | yes |
| B5 | repository apply failure leaves the capability consumed | `TestPositionPolicyCapabilityRejectsExpiryOtherEngineStaleAndConcurrentReplay/repository_failure_consumes_fail_closed` | yes | yes |
| B6 | committed result differs from preview-bound after state | `TestPositionPolicyCapabilityRejectsExpiryOtherEngineStaleAndConcurrentReplay/result_mismatch_is_fail_closed` | yes | yes |
