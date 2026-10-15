# Branch Test Map: `Client.ListPendingOrders`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | stored session absent | existing auth/client coverage | no | no |
| B2 | wrong-type `result` decode error propagates | planned table test in `account_test.go` or `orders_parse_test.go` | no | no |
| B3 | valid order item remains parsed | `TestListPendingOrdersParsesStructuredFields` | no | no |
| planned guard | result absent | planned table test | no | no |
| planned guard | `result:null` | planned table test | no | no |
| planned guard | `result:[]` is successful non-nil empty list | planned table test | no | no |

No test executed during pre-edit preparation.
