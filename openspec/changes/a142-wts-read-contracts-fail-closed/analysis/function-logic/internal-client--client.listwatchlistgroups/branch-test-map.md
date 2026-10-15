# Branch Test Map: `Client.ListWatchlistGroups`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | no-session error | existing client coverage | no | yes |
| B2 | bulk GET/decode failure | `TestListWatchlistRejectsInvalidCollections` | yes | yes |
| helper | required collections, stable ordering, code/currency/price mapping | `TestListWatchlistFromFixtures`; `TestListWatchlistRejectsInvalidCollections` | yes | yes |
| isolation | mutation request body and XSRF header preserved | `TestWatchlistMutationContract`; `TestWatchlistMutationSendsXSRF` | no | yes |

Focused local fixtures passed after implementation.
