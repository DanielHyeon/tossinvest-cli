# Function Logic Map: `TestListWatchlistRejectsInvalidCollections`

- Source: `internal/client/watchlist_test.go`
- AST evidence: `ast.json` (`cd9d991e892165066410a61d2046d8ebfa96d0827bc4202e04b1f4ea44fd87a4`)

## Intent

Local `httptest` fixture coverage for the a142 read contract. No live endpoint is contacted.

## Branches

| Branch | Assertion |
|---|---|
| B1 | local malformed and empty collection fixture table iterates |
| B2 | explicit empty collection cases selected |
| B3 | empty collection result is non-nil and successful |
| B4 | missing/null/wrong-type collection returns error and nil |

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
