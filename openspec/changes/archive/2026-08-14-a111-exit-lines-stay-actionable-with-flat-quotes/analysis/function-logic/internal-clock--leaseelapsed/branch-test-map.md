# Branch Test Map: `LeaseElapsed`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `internal/clock/clock.go:66`: use the matching monotonic leaseClock elapsed path or deterministic injected Clock.Since fallback | `TestA111ObserverUsesClockLeaseHelpersForTheUseLease` | intentional A111 RED before the corresponding production correction | focused A111 suite GREEN |
