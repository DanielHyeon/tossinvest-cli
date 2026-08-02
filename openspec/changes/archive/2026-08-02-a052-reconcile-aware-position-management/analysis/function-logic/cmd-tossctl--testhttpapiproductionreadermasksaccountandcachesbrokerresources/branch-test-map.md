# Branch Test Map: `TestHTTPAPIProductionReaderMasksAccountAndCachesBrokerResources`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `range` line 48 | `for range 2 {` entered and complementary path | TestHTTPAPIProductionReaderMasksAccountAndCachesBrokerResources | pre-existing regression at frozen base | verified by current package suite |
| B2 | `if` line 50 | `if err != nil {` entered and complementary path | TestHTTPAPIProductionReaderMasksAccountAndCachesBrokerResources | pre-existing regression at frozen base | verified by current package suite |
| B3 | `if` line 53 | `if len(positions.Items) != 1 \|\| positions.Items[0].Symbol != "005930" \|\|` entered and complementary path | TestHTTPAPIProductionReaderMasksAccountAndCachesBrokerResources | pre-existing regression at frozen base | verified by current package suite |
| B4 | `if` line 58 | `if err != nil {` entered and complementary path | TestHTTPAPIProductionReaderMasksAccountAndCachesBrokerResources | pre-existing regression at frozen base | verified by current package suite |
| B5 | `if` line 61 | `if len(projectedOrders.Items) != 1 \|\| projectedOrders.Items[0].ID != "order-1" \|\|` entered and complementary path | TestHTTPAPIProductionReaderMasksAccountAndCachesBrokerResources | pre-existing regression at frozen base | verified by current package suite |
| B6 | `if` line 66 | `if holdings.calls != 1 \|\| orders.calls != 1 {` entered and complementary path | TestHTTPAPIProductionReaderMasksAccountAndCachesBrokerResources | pre-existing regression at frozen base | verified by current package suite |
