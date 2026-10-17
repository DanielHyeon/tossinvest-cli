## Context

Pre-edit AST at baseline 6765858a identifies ListPendingOrders and ListCompletedOrdersRange collection boundaries and the legacy ListWatchlist route. Dedicated ListWatchlistGroups already provides a read path, but its current decoder lacks fields needed for flat-list compatibility. Copy and verify intake hashes before implementation.

## Goals / Non-Goals

**Goals:** Distinguish unavailable/malformed order data from authoritative empty collections; migrate watchlist reads without output-field loss.

**Non-Goals:** Order submission/cancellation changes, reconciliation policy redesign, watchlist mutations, new all-dates APIs, broad domain changes, exact order-ID normalization outside these collection guards.

## Decisions

1. Add minimal post-decode nil checks at pending result and every completed result.body boundary. Preserve normal wrong-type decode errors. Never return partial KR results when a required US response fails. Explicit [] is a successful empty result. Do not synthesize order absence from HTTP or schema failure.
2. Reuse the dedicated GET /api/v1/new-watchlists?includePrice=true&lazyLoad=false read path. Expand only the necessary read decoder fields to preserve group name, stock name/symbol, currency, base price and current price. Keep established optional-field defaults only where supported by existing/upstream fixtures; do not invent fallback financial values. Sort groups and items stably by ascending ordering; ties retain wire order. Preserve local flat Symbol=item.code (the old StockCode contract), without adding ProductCode or substituting ticker symbol. Copy currency verbatim; absent currency remains empty. Nullable base/close prices retain current zero defaults.
3. Required watchlist collection fields missing/null/wrong-type must produce an error; explicit empty arrays succeed. Validate bulk watchlists and each items array only. Do not import simple/groups route validation, itemCount checks or positive-ID mutation reconciliation rules. Do not alter any mutation payload or shared mutation semantics.
4. Test responses through local HTTP fixtures. Include existing consumers' error propagation in branch maps without changing their policy. Only existing token/session setup may be mocked; no real-account API calls.

5. Additional gstack review found the existing watchlist health probe still checks the retired dashboard POST. Align only that watchlist ProbeSpec method, URL/query, empty body and shape check with the dedicated read contract. Reject absent/null/wrong-type bulk watchlists and per-group items; accept explicit empty arrays and preserve status errors. Keep watchlist operation ID/handler, all other probes, monitoring schedules, configuration and live execution unchanged. Capture fresh pre-edit evidence for the containing registry function before editing; prove the probe contract through offline fixtures.

## Risks / Trade-offs

- Stricter errors can expose upstream drift → intentional fail-closed behavior, stable descriptive schema errors and explicit-empty tests.
- Shared watchlist decoder may affect mutation helpers → preserve mutation contracts and run associated existing tests.
- Unspecified ordering/optional values → compare upstream fixtures, retain current flat public contract, report unresolved differences before implementation.

## Migration Plan

No data/settings migration or deployment. Revert code commit if needed; never compensate by relaxing order-read failure handling.

## Open Questions

Resolved by independent upstream-fixture review: required bulk watchlists/items; optional currency and nullable prices retain the existing defaults; ordering is stable ascending; Symbol remains item.code. No user decisions pending.
