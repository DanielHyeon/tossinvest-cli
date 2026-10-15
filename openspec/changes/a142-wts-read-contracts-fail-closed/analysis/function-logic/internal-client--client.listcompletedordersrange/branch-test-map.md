# Branch Test Map: `Client.ListCompletedOrdersRange`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | session error preserves behavior | existing client coverage | no | yes |
| B2 | metadata error preserves behavior | existing client coverage | no | yes |
| B3 | invalid market rejects | existing normalization coverage | no | yes |
| B4 | each normalized market is requested | `TestListCompletedOrdersRangeRejectsPartialMarketResult` | yes | yes |
| B5 | GET/decode failure discards aggregate | `TestListCompletedOrdersRangeRejectsMissingCollections` | yes | yes |
| B6 | missing/null body fails closed | `TestListCompletedOrdersRangeRejectsMissingCollections`; `TestListCompletedOrdersRangeRejectsPartialMarketResult` | yes | yes |
| B7 | valid body parses; `body:[]` is non-nil empty | `TestListCompletedOrdersParsesStructuredFields`; `TestListCompletedOrdersRangeAcceptsExplicitEmptyBody` | yes | yes |
| B8 | both sort timestamps nil | existing ordering coverage | no | yes |
| B9 | left timestamp nil | existing ordering coverage | no | yes |
| B10 | right timestamp nil | existing ordering coverage | no | yes |

Focused local fixtures passed after implementation.
