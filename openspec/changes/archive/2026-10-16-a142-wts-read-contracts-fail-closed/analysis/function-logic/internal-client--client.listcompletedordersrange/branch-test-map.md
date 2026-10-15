# Branch Test Map: `Client.ListCompletedOrdersRange`

| Branch | Scenario | Test / evidence | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | session failure | Static preserved branch; no isolated fixture receipt | not claimed | not isolated |
| B2 | metadata failure | Static preserved branch; no isolated fixture receipt | not claimed | not isolated |
| B3 | market normalization error | Static preserved branch; no isolated fixture receipt | not claimed | not isolated |
| B4 | US then KR iteration | TestListCompletedOrdersRangeRejectsPartialMarketResult asserts both requests | test-level baseline FAIL on schema case | named test PASS |
| B5 | GET/decode failure returns nil aggregate | TestListCompletedOrdersRangeRejectsMissingCollections wrong-type; TestListCompletedOrdersRangeRejectsPartialMarketResult HTTP failure | no (preserved error path) | named tests PASS |
| B6 | nil body returns schema error and nil aggregate | TestListCompletedOrdersRangeRejectsMissingCollections; TestListCompletedOrdersRangeRejectsPartialMarketResult | retrospective baseline FAIL; body guard mutant FAIL | named tests PASS |
| B7 | range appends parsed order; [] skips loop | TestListCompletedOrdersParsesStructuredFields; TestListCompletedOrdersRangeAcceptsExplicitEmptyBody | no (preserved behavior) | named tests PASS |
| B8 | both timestamps nil in sort comparator | Static preserved branch; no isolated fixture receipt | not claimed | not isolated |
| B9 | left timestamp nil in comparator | Static preserved branch; no isolated fixture receipt | not claimed | not isolated |
| B10 | right timestamp nil in comparator | Static preserved branch; no isolated fixture receipt | not claimed | not isolated |

Named-test PASS is fixture evidence, not instrumented branch coverage. Historical RED terminal output was not retained; retrospective baseline failures and independent guard mutant logs are identified separately in analysis/verification-receipt.md. Static unchanged branches are explicitly not individually proven.
