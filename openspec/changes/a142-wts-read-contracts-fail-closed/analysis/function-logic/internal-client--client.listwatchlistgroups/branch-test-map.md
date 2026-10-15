# Branch Test Map: `Client.ListWatchlistGroups`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | no session rejects | existing client coverage | no | no |
| B2 | HTTP/wrong-type decode error propagates | planned local-server test | no | no |
| B3 | every group maps | planned dedicated fixture test | no | no |
| B4 | every item maps with code Symbol, empty currency and nil-price zero defaults | planned dedicated optional-field fixture test | no | no |
| planned validation | missing/null `watchlists` or any `items` errors | planned table test | no | no |
| planned sort | numeric ordering ascending and equal keys retain wire order | planned shuffled fixture test | no | no |
| isolation | mutation payloads and XSRF behavior unchanged | `TestWatchlistMutationContract`; `TestWatchlistMutationSendsXSRF` | no | no |

No test executed during pre-edit preparation.
