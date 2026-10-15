# Branch Test Map: `TestListWatchlistFromFixtures`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | client call returns no error | `TestListWatchlistFromFixtures` | no | yes |
| B2 | dedicated GET route and query asserted | `TestListWatchlistFromFixtures` | no | yes |
| B3 | group/item ordering and code symbols asserted | `TestListWatchlistFromFixtures` | no | yes |
| B4 | empty currency and nil prices map to zero | `TestListWatchlistFromFixtures` | no | yes |

Executed in `go test ./internal/client -count=1`.
