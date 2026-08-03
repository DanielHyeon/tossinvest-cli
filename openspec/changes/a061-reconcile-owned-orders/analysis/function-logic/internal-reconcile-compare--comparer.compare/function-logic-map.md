# Function Logic Map: `Comparer.Compare`

Source: `internal/reconcile/compare.go`
Function: `Comparer.Compare`
Signature: `Comparer.Compare(params=2, results=1)`
Source SHA-256: `36ce21d173549fe4b957c6132a56993887fb62dfe3acaa7c9afd39a6e61154b2`
Revision: `current`

## Inputs and invariants

- Inputs are `Comparer.Compare(params=2, results=1)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/reconcile/compare.go:362 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | range | internal/reconcile/compare.go:368 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | range | internal/reconcile/compare.go:374 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | if | internal/reconcile/compare.go:375 | Preserve the condition, error propagation, and fail-closed behavior. |
| B5 | range | internal/reconcile/compare.go:380 | Preserve the condition, error propagation, and fail-closed behavior. |
| B6 | if | internal/reconcile/compare.go:382 | Preserve the condition, error propagation, and fail-closed behavior. |
| B7 | range | internal/reconcile/compare.go:389 | Preserve the condition, error propagation, and fail-closed behavior. |
| B8 | if | internal/reconcile/compare.go:394 | Preserve the condition, error propagation, and fail-closed behavior. |
| B9 | if | internal/reconcile/compare.go:397 | Preserve the condition, error propagation, and fail-closed behavior. |
| B10 | switch | internal/reconcile/compare.go:403 | Preserve the condition, error propagation, and fail-closed behavior. |
| B11 | case | internal/reconcile/compare.go:404 | Preserve the condition, error propagation, and fail-closed behavior. |
| B12 | case | internal/reconcile/compare.go:406 | Preserve the condition, error propagation, and fail-closed behavior. |
| B13 | case | internal/reconcile/compare.go:412 | Preserve the condition, error propagation, and fail-closed behavior. |
| B14 | case | internal/reconcile/compare.go:416 | Preserve the condition, error propagation, and fail-closed behavior. |
| B15 | range | internal/reconcile/compare.go:426 | Preserve the condition, error propagation, and fail-closed behavior. |
| B16 | range | internal/reconcile/compare.go:427 | Preserve the condition, error propagation, and fail-closed behavior. |
| B17 | if | internal/reconcile/compare.go:429 | Preserve the condition, error propagation, and fail-closed behavior. |
| B18 | range | internal/reconcile/compare.go:442 | Preserve the condition, error propagation, and fail-closed behavior. |
| B19 | if | internal/reconcile/compare.go:443 | Preserve the condition, error propagation, and fail-closed behavior. |
| B20 | if | internal/reconcile/compare.go:447 | Preserve the condition, error propagation, and fail-closed behavior. |
| B21 | range | internal/reconcile/compare.go:454 | Preserve the condition, error propagation, and fail-closed behavior. |
| B22 | if | internal/reconcile/compare.go:455 | Preserve the condition, error propagation, and fail-closed behavior. |
| B23 | range | internal/reconcile/compare.go:463 | Preserve the condition, error propagation, and fail-closed behavior. |
| B24 | if | internal/reconcile/compare.go:464 | Preserve the condition, error propagation, and fail-closed behavior. |
| B25 | if | internal/reconcile/compare.go:469 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `snap.AsOf.IsZero`: errors and state follow mapped branches.
- `Format`: errors and state follow mapped branches.
- `snap.AsOf.UTC`: errors and state follow mapped branches.
- `make`: errors and state follow mapped branches.
- `len`: errors and state follow mapped branches.
- `strings.ToUpper`: errors and state follow mapped branches.
- `strings.TrimSpace`: errors and state follow mapped branches.
- `append`: errors and state follow mapped branches.
- `sort.Strings`: errors and state follow mapped branches.
- `canonicalDecimal`: errors and state follow mapped branches.
- `isZeroDecimal`: errors and state follow mapped branches.
- `quantitiesAgree`: errors and state follow mapped branches.
- `localOrdersForComparison`: errors and state follow mapped branches.
- `brokerOrderIdentityForLocal`: errors and state follow mapped branches.
- `localOrder.Identity`: errors and state follow mapped branches.
- `identitiesCompatible`: errors and state follow mapped branches.
- `sort.Slice`: errors and state follow mapped branches.
- `less`: errors and state follow mapped branches.
- `missing.Identity`: errors and state follow mapped branches.
- `brokerOrderIdentity`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 36; return points: 3; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
