# Branch Test Map: `Console.handlePositions`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Branch-free `/positions` orchestration delegates post-policy/post-marker freshness without another broker read; a stopped read stays stopped through rollback | `TestA111HoldingsRoutesRecheckFreshnessAfterPolicyCacheMiss`, `TestA111HoldingsRoutesNeverResurrectStoppedMarkerAfterClockRollback` | intentional A111 REDs before shared downgrade-only decoration | focused A111 suite GREEN |
