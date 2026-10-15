# Branch Test Map: `Client.ListWatchlist`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | grouped-read error propagates | `TestListWatchlistRejectsInvalidCollections` | yes | yes |
| B2 | groups flatten in mapped stable order | `TestListWatchlistFromFixtures` | yes | yes |

Focused local fixtures passed after implementation.
