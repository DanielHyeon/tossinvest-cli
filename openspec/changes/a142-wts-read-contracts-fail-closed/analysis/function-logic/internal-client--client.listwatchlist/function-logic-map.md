# Function Logic Map: `Client.ListWatchlist`

- Source: `internal/client/watchlist.go`
- AST evidence: `ast.json` (`fc5440b66ddc096a9daeb2a8a125cf132d69bb69b7c191aaab864a44eb74f33b`)
- Risk scan: `risk-pattern-report.md` (no configured match)

## Inputs and invariants

| Input/state | Valid range | Failure behavior |
|---|---|---|
| grouped read | delegated to `ListWatchlistGroups` | propagated error |
| flat output | group-order then item-order flattening | empty input yields non-nil empty slice |

## Control flow

The wrapper calls ListWatchlistGroups, returns its error at B1 (line 11), then iterates groups at B2 (line 15) and appends their items. Empty groups yield a non-nil empty output. It performs no direct HTTP or mutation itself.

## Safety conclusion

The retired dashboard adapter and legacy POST are not reachable. Schema validation, stable ordering, code symbols, and optional-money defaults remain in the grouped reader.

## Branches and early returns
| Branch | Condition | Result | Evidence |
|---|---|---|---|
| B1 | grouped read returned error | nil output and propagated error | TestListWatchlistRejectsInvalidCollections |
| B2 | range over groups | append items in mapped order | TestListWatchlistFromFixtures |

## Calls and live bindings
Calls are listed in `ast.json`; fixture tests use only local `httptest`.

## State mutations and fallbacks
No mutation or fallback is introduced by this evidence.
