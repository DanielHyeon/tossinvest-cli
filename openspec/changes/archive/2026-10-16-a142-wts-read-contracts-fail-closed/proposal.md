## Why

Current AST-backed intake shows that omitted/null pending and completed order arrays can be returned as successful empty lists. Such schema drift must not become evidence that an order is absent. The legacy watchlist dashboard route also needs replacement with the existing dedicated read endpoint.

## What Changes

- Reject missing/null order collections; preserve explicit empty-array success and existing parsing errors.
- Read watchlists through the dedicated new-watchlists endpoint and preserve flat output fields.
- Align the existing watchlist health-probe definition with the dedicated GET route and required collection contract; retain operation ID and schedules.
- Add offline negative/compatibility tests and independent review.

## Capabilities

### New Capabilities

- `wts-read-integrity`: Fail-closed order collections and compatible dedicated watchlist reads.

### Modified Capabilities

None.

## Impact

STORY-TOS-a142 / FEAT-TOS-014. Owned files: internal/client/account.go, completed_orders.go, watchlist.go, narrowly scoped read decoder in watchlist_manage.go, associated tests; the watchlist ProbeSpec in internal/ops/wts_operations.go and its offline fixture tests only. The read methods feed order lookup/recovery: treat order-array changes as high risk. Preserve mutation bodies, journal, Guardian, auth/token sharing, operation IDs, runtime settings and live state. Pre-edit AST evidence: /tmp/tossos-a142-intake/ast (copy into change evidence before implementation). Depends on reviewed a141 landing to establish an unambiguous baseline.
