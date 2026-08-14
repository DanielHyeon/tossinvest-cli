# Branch Test Map: `Collector.holdings`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | adapter exposes the lossless raw reader | `TestA110CollectorPreservesUnreadableRawHoldingQuantity` | yes | yes |
| B2 | raw reader error remains an error | existing collector adapter-error tests | preserve | yes |
| B3 | blank raw quantity remains blank and fails closed; valid `04.00` becomes exact `4` and remains nonblocking external | `TestA110CollectorPreservesUnreadableRawHoldingQuantity` | yes (M19 also kills helper) | yes |
| B4 | legacy reader error remains an error | existing collector adapter-error tests | preserve | yes |
| B5 | legacy float holdings retain previous normalization | existing collector snapshot tests | preserve | yes |
