# Branch Test Map: `Console.handlePositions`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | no settings seam still renders read-only position exit lines | `TestTradingViewsAreInputFreeAndRejectPOST` | no | no |
| B2 | settings seam can annotate state but renders no form/control | `TestTradingViewsAreInputFreeAndRejectPOST` | no | no |
| B3 | settings read failure does not hide broker/journal rows | existing `TestPositions...` seam failure coverage | existing | existing |
| B4 | complete/stale/unknown/one-share snapshots map without recomputation | `TestPositionsRenderCanonicalExitLineFixtures` | no | no |
