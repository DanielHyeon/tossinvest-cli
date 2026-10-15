# Branch Test Map: `Client.ListCompletedOrdersRange`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1-B3 | session/metadata/market errors preserve behavior | existing client tests | no | no |
| B4/B6 | valid US order parses | `TestListCompletedOrdersParsesStructuredFields` | no | no |
| B5 | second required market fails after first succeeds | planned local-server test | no | no |
| planned guard | `result` absent/null | planned table test | no | no |
| planned guard | `result.body` absent/null | planned table test | no | no |
| B5/planned guard | body object/scalar | planned table test | no | no |
| B6 | body `[]` succeeds with non-nil empty list | planned table test | no | no |
| B7-B9 | current submitted-time sort stays unchanged | existing completed-order tests | no | no |

No test executed during pre-edit preparation.
