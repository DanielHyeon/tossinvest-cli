# Branch Test Map: `Client.ListPendingOrders`

| Branch | Scenario | Test / evidence | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | requireSession error | Static preserved branch; no isolated fixture receipt | not claimed | not isolated |
| B2 | GET/decode failure; object/scalar result | TestListPendingOrdersRejectsAbsentOrNullResult wrong-type cases | no (already rejected) | named test PASS |
| B3 | nil result rejected | TestListPendingOrdersRejectsAbsentOrNullResult; TestHasPendingOrderRejectsUnknownPendingCollection; TestFindOrderRejectsUnknownPendingCollectionWithoutCompletedFallback | retrospective baseline FAIL; pending guard mutant FAIL | named tests PASS |
| B4 | range parses populated result; [] skips loop | TestListPendingOrdersParsesStructuredFields; TestListPendingOrdersAcceptsExplicitEmptyResult | no (preserved behavior) | named tests PASS |

Named-test PASS is fixture evidence, not instrumented branch coverage. Historical RED terminal output was not retained; retrospective baseline failures and independent guard mutant logs are identified separately in analysis/verification-receipt.md. Static unchanged branches are explicitly not individually proven.
