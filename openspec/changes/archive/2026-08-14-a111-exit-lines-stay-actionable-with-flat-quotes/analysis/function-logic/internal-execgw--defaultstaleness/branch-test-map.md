# Branch Test Map: `DefaultStaleness`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Branch-free happy path: bind QueryPrice entry-gate staleness to the same exported duration used by source/use quote validation | `TestA111FallbackSequenceRecoveryIsLazyAndPriceEvidenceUsesTheGateDuration` | intentional A111 RED before the corresponding production correction | focused A111 suite GREEN |
