# Function Logic Map: `Client.ListWatchlist`

- Source: `internal/client/watchlist.go`
- AST evidence: `ast.json` (`9992281d395fc5ba71ea8033e10e9e819062d614d04761c55aef80ac66028c73`)
- Risk scan: `risk-pattern-report.md` (no configured match)

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| session | stored cookies exist | `requireSession` | return `ErrNoSession` |
| flat output | group, `Symbol=item.code`, name, currency, base, last | a142 spec | preserve current public shape |
| read route | bulk GET `/api/v1/new-watchlists?includePrice=true&lazyLoad=false` | a142 spec | no legacy dashboard POST |
| required collections | `result.watchlists` and every group `items` are arrays | a142 spec | missing/null/wrong type error; explicit empty valid |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | session failure | none | error | existing session coverage |
| B2 | legacy POST/decode failure | legacy POST only at baseline | error | replaced by dedicated GET test |
| B3-B5 | find/decode WATCHLIST section | no mutation | section or error | removed with legacy adapter |
| B6-B8 | flatten legacy groups/items | local construction only | flat slice | replaced by shared dedicated decoder test |
| planned wrapper | dedicated bulk helper returns | one read-only GET indirectly | fields/order/error preserved | multi-group contract test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `requireSession` | preserve session gate | propagated | AST B1 |
| baseline `postJSON`, `json.Unmarshal` | legacy dashboard adapter | propagated; retired from this function | AST B2-B5 |
| planned read-only helper | shared dedicated endpoint mapping | propagated; no fallback to legacy route | a142 design |
| `coalesceMoney` | nullable money to current zero default | no inference | AST B8 / shared decoder |

## State mutations and fallbacks

- Read-only route replacement. No watchlist mutation method, payload, operation ID, domain field, token, or setting change.
- No fallback to the legacy endpoint after dedicated collection-schema failure.

## Safety conclusion

- Safe edit boundary: replace this adapter with a narrow call to the dedicated bulk read/flatten helper.
- High-risk impact: no direct trade mutation. Compatibility-sensitive output and shared decoder require mutation regression coverage.
