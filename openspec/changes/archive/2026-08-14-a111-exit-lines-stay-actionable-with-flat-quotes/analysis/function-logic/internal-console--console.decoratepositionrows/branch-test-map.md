# Branch Test Map: `Console.decoratePositionRows`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Policy seam wired/unwired preserves unknown fallback | `TestPositionsShowRuntimeUnknownWhenCommanderUnavailableButDesiredIncludesUS` | preservation coverage | existing suite GREEN |
| B2 | A delayed successful policy cache miss is followed by a new marker-bound response authority | `TestA111HoldingsRoutesRecheckFreshnessAfterPolicyCacheMiss` | intentional A111 RED | focused A111 suite GREEN |
| B3 | Successful states are indexed once for both holdings routes | `TestA111ActualPositionsAndPositionManagementRoutesRenderTheSameFreshLine` | prior A111 convergence RED | focused A111 suite GREEN |
| B4 | Optional settings seam does not control effective freshness | `TestA111ActualPositionsAndPositionManagementRoutesRenderTheSameFreshLine` | prior A111 convergence RED | focused A111 suite GREEN |
| B5 | Settings failure withholds desired decoration | `TestThePositionsScreenRendersWithEitherSourceMissing` | preservation coverage | existing suite GREEN |
| B6 | One settings snapshot decorates every row | `TestExcludingASymbolFromThePositionsScreen` | preservation coverage | existing suite GREEN |
| B7 | Attempted policy read projects management only after the response authority is fixed | `TestA111HoldingsRoutesRecheckFreshnessAfterPolicyCacheMiss` | intentional A111 RED | focused A111 suite GREEN |
| B8 | Every row shares the same post-marker time and downgrade-only liveness | `TestA111HoldingsRoutesNeverResurrectStoppedMarkerAfterClockRollback` | intentional A111 RED | focused A111 suite GREEN |
| B9 | Journal rows require lifecycle proof | `TestA111ActualPositionsAndPositionManagementRoutesRenderTheSameFreshLine` | prior A111 convergence RED | focused A111 suite GREEN |
| B10 | Missing lifecycle state remains unknown | `TestPositionsShowRuntimeUnknownWithoutDesiredFallback` | preservation coverage | existing suite GREEN |
| B11 | Matching lifecycle state permits only same-generation projection | `TestA111ActualPositionsAndPositionManagementRoutesRenderTheSameFreshLine` | prior A111 convergence RED | focused A111 suite GREEN |
| B12 | Reconciliation block age uses post-policy/post-marker response time | `TestA111HoldingsRoutesRecheckFreshnessAfterPolicyCacheMiss` | intentional A111 RED | focused A111 suite GREEN |
