# Branch Test Map: `Client.ListPendingOrders`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | stored session absent | existing auth/client coverage | no | no |
| B2 | wrong-type `result` decode error propagates | `TestListPendingOrdersRejectsAbsentOrNullResult`; `TestHasPendingOrderRejectsUnknownPendingCollection`; `TestFindOrderRejectsUnknownPendingCollectionWithoutCompletedFallback` | yes | yes |
| B3 | absent/null result fails closed | `TestListPendingOrdersRejectsAbsentOrNullResult` | yes | yes |
| B4 | explicit `result:[]` is successful non-nil empty list; valid items remain parsed | `TestListPendingOrdersAcceptsExplicitEmptyResult`; `TestListPendingOrdersParsesStructuredFields` | yes | yes |

Focused local fixtures passed after implementation.
