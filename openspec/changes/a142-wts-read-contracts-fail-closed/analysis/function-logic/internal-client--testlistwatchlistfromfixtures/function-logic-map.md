# Function Logic Map: `TestListWatchlistFromFixtures`

- Source: `internal/client/watchlist_test.go`
- AST evidence: `ast.json` (`de3ec778a43b365513e8d9ae4070f3bd19a039c6ac5c5bf397d46af9288e7bf7`)

## Intent

Local `httptest` fixture coverage for the a142 read contract. No live endpoint is contacted.

## Branches

| Branch | Assertion |
|---|---|
| B1 | client call returns no error |
| B2 | dedicated GET route and query asserted |
| B3 | group/item ordering and code symbols asserted |
| B4 | empty currency and nil prices map to zero |

## Inputs and invariants
Evidence is local and source-bound.

## Branches and early returns
See the branch table below.

## Calls and live bindings
Calls are listed in `ast.json`; fixture tests use only local `httptest`.

## State mutations and fallbacks
No mutation or fallback is introduced by this evidence.

## Safety conclusion
Evidence records the current source and focused fixture result.
