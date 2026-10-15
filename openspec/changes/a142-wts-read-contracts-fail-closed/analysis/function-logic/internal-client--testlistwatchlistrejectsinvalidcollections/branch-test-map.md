# Branch Test Map: `TestListWatchlistRejectsInvalidCollections`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | local malformed and empty collection fixture table iterates | `TestListWatchlistRejectsInvalidCollections` | no | yes |
| B2 | explicit empty collection cases selected | `TestListWatchlistRejectsInvalidCollections` | no | yes |
| B3 | empty collection result is non-nil and successful | `TestListWatchlistRejectsInvalidCollections` | no | yes |
| B4 | missing/null/wrong-type collection returns error and nil | `TestListWatchlistRejectsInvalidCollections` | no | yes |

Executed in `go test ./internal/client -count=1`.
