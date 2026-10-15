# Branch Test Map: `Client.ListCompletedOrdersRange`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | session error preserves behavior | existing client coverage | no | no |
| B2 | metadata error preserves behavior | existing client coverage | no | no |
| B3 | invalid market rejects | existing normalization coverage | no | no |
| B4 | each normalized market is requested | planned multi-market local-server test | no | no |
| B5 | GET/decode error after first market succeeds | planned local-server test | no | no |
| B6 | valid body items parse | `TestListCompletedOrdersParsesStructuredFields` | no | no |
| planned guard | `result` absent/null | planned table test | no | no |
| planned guard | `result.body` absent/null | planned table test | no | no |
| B7 | both submitted times nil preserve ID order | existing completed-order ordering coverage | no | no |
| B8 | left submitted time nil sorts after timestamped order | existing completed-order ordering coverage | no | no |
| B9 | right submitted time nil sorts after timestamped order | existing completed-order ordering coverage | no | no |
| planned guard | body object/scalar | planned table test | no | no |
| B6 | body `[]` succeeds with non-nil empty list | planned table test | no | no |

No test executed during pre-edit preparation.
