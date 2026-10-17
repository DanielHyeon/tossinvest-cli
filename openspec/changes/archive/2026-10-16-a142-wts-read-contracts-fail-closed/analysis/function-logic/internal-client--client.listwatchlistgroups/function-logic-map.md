# Function Logic Map: `Client.ListWatchlistGroups`

- Source: `internal/client/watchlist_manage.go`
- AST evidence: `ast.json` (`de925d43fbf32a246efaf6c6df56fbf6a2f7158d00188818dcbc6dbb39dca4f1`)
- Risk scan: `risk-pattern-report.md` (no configured match)

## Inputs and invariants

| Input/state | Valid range | Failure behavior |
|---|---|---|
| session | existing session contract | propagated error |
| bulk envelope | `watchlists` and every `items` field must be arrays | absent/null errors; incompatible JSON errors |
| optional fields | currency may be empty; prices may be null | empty currency; zero monetary values |
| ordering | ascending numeric `ordering`, stable equal ties | local deterministic ordering |

## Branches and early returns

| Branch | Condition | Return/error | Test |
|---|---|---|---|
| B1 | session failure | error | static preserved branch; no isolated fixture receipt |
| B2 | bulk GET/decode failure | propagated error | `TestListWatchlistRejectsInvalidCollections` |

`mapBulkWatchlistGroups` validates the decoded collections, stable-sorts groups and items, and maps `Symbol` from `code`. It is called only after B2 succeeds.

## Safety conclusion

Only the bulk read decoder changed. Mutation operations in this file retain their existing request bodies and XSRF behavior.

## Calls and live bindings
Calls are listed in `ast.json`; fixture tests use only local `httptest`.

## State mutations and fallbacks
No mutation or fallback is introduced by this evidence.
