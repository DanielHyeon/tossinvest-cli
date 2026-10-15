# Branch Test Map: `TestListWatchlistRejectsInvalidCollections`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | local malformed and empty collection fixture table iterates | `TestListWatchlistRejectsInvalidCollections` | no | yes |
| B2 | explicit empty collection cases selected | `TestListWatchlistRejectsInvalidCollections` | no | yes |
| B3 | empty collection result is non-nil and successful | `TestListWatchlistRejectsInvalidCollections` | no | yes |
| B4 | missing/null/wrong-type collection returns error and nil | `TestListWatchlistRejectsInvalidCollections` | no | yes |

Named test PASS is confirmed by the final focused receipt. Assertion conditions are evaluated; passing does not execute their fatal branches. No instrumented branch-hit claim is made.
