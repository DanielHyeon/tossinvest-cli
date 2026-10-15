# Function Logic Map: `Client.ListWatchlist`

- Source: `internal/client/watchlist.go`
- AST evidence: `ast.json` (`9da13ccf13a0fec5c4c019deeb2d19624bc7c5d18809dcbfb466980b5cb8020e`)
- Risk scan: `risk-pattern-report.md` (no configured match)

## Inputs and invariants

| Input/state | Valid range | Failure behavior |
|---|---|---|
| grouped read | delegated to `ListWatchlistGroups` | propagated error |
| flat output | group-order then item-order flattening | empty input yields non-nil empty slice |

## Control flow

This branchless wrapper reads grouped data through the existing dedicated bulk read, then flattens its mapped items. It performs no HTTP or mutation itself.

## Safety conclusion

The retired dashboard adapter and legacy POST are not reachable. Schema validation, stable ordering, code symbols, and optional-money defaults remain in the grouped reader.

## Branches and early returns
See the branch table below.

## Calls and live bindings
Calls are listed in `ast.json`; fixture tests use only local `httptest`.

## State mutations and fallbacks
No mutation or fallback is introduced by this evidence.
