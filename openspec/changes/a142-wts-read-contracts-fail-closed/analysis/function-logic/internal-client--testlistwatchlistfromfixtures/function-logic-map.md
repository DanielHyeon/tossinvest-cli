# Function Logic Map: `TestListWatchlistFromFixtures`

- Source: `internal/client/watchlist_test.go`
- AST evidence: `ast.json` (`fb67a940c20d8bc5ce74f43f255d9e29944680d5e1804574bbfc77095cad1f42`)

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
