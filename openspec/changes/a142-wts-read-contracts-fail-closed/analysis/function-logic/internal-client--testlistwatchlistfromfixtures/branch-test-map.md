# Branch Test Map: `TestListWatchlistFromFixtures`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | client call returns no error | `TestListWatchlistFromFixtures` | no | yes |
| B2 | dedicated GET route and query asserted | `TestListWatchlistFromFixtures` | no | yes |
| B3 | full retained fields, stable ties, ordering, and code symbols asserted | `TestListWatchlistFromFixtures` | no | yes |

Executed in `go test ./internal/client -count=1`.
