# Branch Test Map: `exitFreshness`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `internal/console/protection_liveness.go:64`: map unwired, stopped and running liveness through the single shared 30-second operator verdict | `TestA111ConsoleConsumesTheSharedFreshnessVerdict` | intentional A111 RED before the corresponding production correction | focused A111 suite GREEN |
| B2 | `internal/console/protection_liveness.go:65`: map unwired, stopped and running liveness through the single shared 30-second operator verdict | `TestA111ConsoleConsumesTheSharedFreshnessVerdict` | intentional A111 RED before the corresponding production correction | focused A111 suite GREEN |
| B3 | `internal/console/protection_liveness.go:67`: map unwired, stopped and running liveness through the single shared 30-second operator verdict | `TestA111ConsoleConsumesTheSharedFreshnessVerdict` | intentional A111 RED before the corresponding production correction | focused A111 suite GREEN |
| B4 | `internal/console/protection_liveness.go:69`: map unwired, stopped and running liveness through the single shared 30-second operator verdict | `TestA111ConsoleConsumesTheSharedFreshnessVerdict` | intentional A111 RED before the corresponding production correction | focused A111 suite GREEN |
