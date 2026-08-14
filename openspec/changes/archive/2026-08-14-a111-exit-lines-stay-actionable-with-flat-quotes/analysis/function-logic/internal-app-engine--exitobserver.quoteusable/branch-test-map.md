# Branch Test Map: `ExitObserver.quoteUsable`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `internal/app/engine/exitloop.go:1073`: reject wall evidence from the future and require clock.LeaseElapsed to remain at the inclusive price-evidence bound | `TestA111QuoteUseLeaseCannotBeExtendedByWallClockRollback` | intentional A111 RED before the corresponding production correction | focused A111 suite GREEN |
