# CodeGraph baseline

- Query date: 2026-10-15
- CodeGraph: `1.6.0`
- Worktree HEAD: `6765858af1beb472268cb699efa1964addb84d60`
- Index refresh: not run. Parent reserved global/single-writer refresh for a141; queries used the existing local read-only index.

## Definitions

| Symbol | Source | Lines | CodeGraph result |
|---|---|---:|---|
| `Client.ListPendingOrders` | `internal/client/account.go` | 154-170 | exact definition |
| `Client.ListCompletedOrdersRange` | `internal/client/completed_orders.go` | 27-75 | exact definition |
| `Client.ListWatchlist` | `internal/client/watchlist.go` | 26-69 | exact definition |
| `Client.ListWatchlistGroups` | `internal/client/watchlist_manage.go` | 41-62 | exact definition |

## Call relationships

| Target | Key callers / affected consumers | Key callees |
|---|---|---|
| pending | `FindOrderWithAliases`; `HasPendingOrder`; `lookupPendingOrder`; `findMatchingPendingOrder`; WTS operations; account/order parser tests | `requireSession`, `getJSON`, `parsePendingOrder` |
| completed range | `ListCompletedOrders`; `FindOrderWithAliases`; `FindCompletedOrderFromLineageHint`; `findMatchingCompletedOrder`; WTS operations; range test | `requireSession`, `ensureTradingMetadata`, `normalizeHistoryMarkets`, `getJSON`, `parseCompletedOrder` |
| watchlist | WTS operations; watchlist tests; `NewCatalog` transitively | `requireSession`, legacy `postJSON`, `coalesceMoney` |
| watchlist groups | WTS operations | `requireSession`, `getJSON`, `coalesceMoney` |

`impact` reported 17 nodes/16 edges for pending, 7/7 for completed range, and 5/4 for watchlist at depth 2. The existing order-read paths reach reconciliation/cancel/amend consumers; no codegraph result authorizes their modification.

## Direct source cross-check

SHA-256 matches `/tmp/tossos-a142-intake` for `account.go`, `completed_orders.go`, and `watchlist.go`; `watchlist_manage.go` is `8513134797ea2a2d448e65bc6f8adbe195afd06ec055ff45049637510041b404`. Source lines match all four queried definitions.
