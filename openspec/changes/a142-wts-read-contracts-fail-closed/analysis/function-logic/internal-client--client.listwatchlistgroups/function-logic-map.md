# Function Logic Map: `Client.ListWatchlistGroups`

- Source: `internal/client/watchlist_manage.go`
- AST evidence: `ast.json` (`8513134797ea2a2d448e65bc6f8adbe195afd06ec055ff45049637510041b404`)
- Risk scan: `risk-pattern-report.md` (no configured match)

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| session | stored cookies exist | `requireSession` | return `ErrNoSession` |
| bulk envelope | dedicated endpoint full response | current route and a142 spec | decode/schema error propagated |
| watchlists/items | arrays; empty allowed | a142 spec | missing/null/wrong type error |
| optional fields | currency may be empty; base/close may be null | a142 decision 2 | empty currency; zero prices |
| ordering | ascending `ordering`, stable ties | a142 decision 2 | local stable sort only |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | session failure | none | error | existing session coverage |
| B2 | GET/decode failure | one read-only GET | propagated error | wrong-type envelope test |
| B3 | group range | local group mapping | append groups | multi-group fixture |
| B4 | item range | local item mapping | append items | code/currency/price fixture |
| planned validation | required collection absent/null | none | descriptive schema error | watchlists/items table tests |
| planned sort | unordered numeric values/equal ties | local stable sort | deterministic flat order | shuffled fixture |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `requireSession` | retain current authentication boundary | propagated | AST B1 |
| `getJSON` | dedicated bulk GET | status/transport/decode propagated | AST B2, CodeGraph |
| `coalesceMoney` | retain nil-price zero default | no currency/price invention | AST B4 |
| planned shared mapper | validate/sort/map read-only response for both flat/groups callers | propagated | a142 design |

## State mutations and fallbacks

- Existing function only reads/mapping data. The same file also contains mutation operations, but their bodies and behavior are excluded.
- No item-count, positive-ID, simple/groups-route, or mutation-reconciliation validation enters this change.

## Safety conclusion

- Safe edit boundary: envelope read fields plus a new read-only mapping helper used by bulk readers.
- High-risk impact: no direct trade mutation. Mutation contract isolation is mandatory because the file shares mutation methods.
