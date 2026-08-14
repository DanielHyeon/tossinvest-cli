# Branch Test Map: `Console.accountPanelFrom`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Dashboard rows are decorated before absence/market aggregation | `TestA111HoldingsRoutesRecheckFreshnessAfterPolicyCacheMiss` | intentional A111 RED | focused A111 suite GREEN |
| B2 | Journal absence stays explicit | `TestAHoldingIsNotCalledUnmanagedWhenTheJournalCouldNotBeRead` | preservation coverage | existing suite GREEN |
| B3 | KR/US panels stay separated | `TestTheOverviewNeverAddsAcrossMarkets` | preservation coverage | existing suite GREEN |
| B4 | Unusable broker cache yields blocked rows | `TestAFailedBrokerReadIsNotTheSameAsAnEmptyCache` | preservation coverage | existing suite GREEN |
| B5 | Only matching broker rows contribute to a market | `TestTheOverviewNeverAddsAcrossMarkets` | preservation coverage | existing suite GREEN |
| B6 | Cross-market/nonbroker rows are skipped | `TestTheOverviewNeverAddsAcrossMarkets` | preservation coverage | existing suite GREEN |
| B7 | Unknown/managed/unmanaged classification is exclusive | `TestAHoldingIsNotCalledUnmanagedWhenTheJournalCouldNotBeRead` | preservation coverage | existing suite GREEN |
| B8 | Unknown management makes counts unmeasured | `TestAHoldingIsNotCalledUnmanagedWhenTheJournalCouldNotBeRead` | preservation coverage | existing suite GREEN |
| B9 | Managed state contributes to managed count | `TestAnAdoptedHoldingRendersAsManagedWithItsBasis` | preservation coverage | existing suite GREEN |
| B10 | Known unmanaged state contributes only to other count | `TestAnUnmanagedHoldingIsLabelledExactlyOnce` | preservation coverage | existing suite GREEN |
| B11 | Empty market is explicit | `TestTheOverviewNeverAddsAcrossMarkets` | preservation coverage | existing suite GREEN |
| B12 | Any unknown row withholds measured management counts | `TestAHoldingIsNotCalledUnmanagedWhenTheJournalCouldNotBeRead` | preservation coverage | existing suite GREEN |
| B13 | Fully known classification emits measured counts | `TestAnAdoptedHoldingRendersAsManagedWithItsBasis` | preservation coverage | existing suite GREEN |
| B14 | Other-market scan runs only for usable broker evidence | `TestAHoldingInNeitherMarketIsNamedRatherThanDropped` | preservation coverage | existing suite GREEN |
| B15 | Other-market rows are collected without dropping symbols | `TestAHoldingInNeitherMarketIsNamedRatherThanDropped` | preservation coverage | existing suite GREEN |
| B16 | KR/US rows are excluded from other-market output | `TestTwoMarketsAloneProduceNoOtherRow` | preservation coverage | existing suite GREEN |
| B17 | Unknown-currency holdings get a blocked explanatory row | `TestAHoldingInNeitherMarketIsNamedRatherThanDropped` | preservation coverage | existing suite GREEN |
| Timing | Dashboard policy delay is included before exit freshness projection | `TestA111HoldingsRoutesRecheckFreshnessAfterPolicyCacheMiss` | intentional A111 RED | focused A111 suite GREEN |
| Rollback | Dashboard never resurrects a stopped marker after response-clock rollback | `TestA111HoldingsRoutesNeverResurrectStoppedMarkerAfterClockRollback` | intentional A111 RED | focused A111 suite GREEN |
