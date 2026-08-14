# Branch Test Map: `LeaseAnchor`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `internal/clock/clock.go:58`: use the optional monotonic leaseClock extension when present, otherwise preserve deterministic injected Clock.Now | `TestA111ObserverUsesClockLeaseHelpersForTheUseLease` | intentional A111 RED before the corresponding production correction | focused A111 suite GREEN |
